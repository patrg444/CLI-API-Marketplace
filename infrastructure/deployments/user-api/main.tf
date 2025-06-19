# Main Terraform Configuration for API-Direct BYOA User API Deployment
# This orchestrates all infrastructure modules for a complete API deployment

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1"
    }
  }

  # Backend configuration will be dynamically generated
  # backend "s3" {
  #   bucket         = "PROJECT_NAME-terraform-state-ACCOUNT_ID"
  #   key            = "PROJECT_NAME/terraform.tfstate"
  #   region         = "REGION"
  #   encrypt        = true
  #   dynamodb_table = "PROJECT_NAME-terraform-locks"
  # }
}

# Configure the AWS Provider
provider "aws" {
  region = var.aws_region

  # Default tags applied to all resources
  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "api-direct"
      DeployedAt  = timestamp()
      Owner       = var.owner_email
    }
  }
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Local values for computed configurations
locals {
  # Common tags
  common_tags = merge(var.additional_tags, {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
    DeployedAt  = timestamp()
    Owner       = var.owner_email
  })

  # API-Direct account ID (this would be configured per environment)
  api_direct_account_id = var.api_direct_account_id

  # Database configuration
  db_name     = replace(var.project_name, "-", "_")
  db_username = "apiuser"

  # Container configuration
  container_port = var.container_port
  health_check_path = var.health_check_path
}

# 1. IAM Cross-Account Role Module
# This must be deployed first to establish secure access
module "iam_cross_account" {
  source = "../../modules/iam-cross-account"

  project_name          = var.project_name
  environment          = var.environment
  api_direct_account_id = local.api_direct_account_id

  # Security configuration
  allowed_user_patterns     = var.allowed_user_patterns
  max_session_duration     = var.max_session_duration
  require_mfa              = var.require_mfa
  allowed_source_ips       = var.allowed_source_ips

  # Feature flags
  enable_service_discovery     = var.enable_service_discovery
  enable_cloudformation_access = var.enable_cloudformation_access
  enable_cost_controls        = var.enable_cost_controls

  # Compliance and governance
  compliance_framework        = var.compliance_framework
  data_residency_requirements = var.data_residency_requirements

  # Integration
  webhook_url                = var.webhook_url
  api_direct_service_endpoints = var.api_direct_service_endpoints

  tags = local.common_tags
}

# 2. Networking Module
# Creates VPC, subnets, load balancer, and security groups
module "networking" {
  source = "../../modules/networking"

  project_name = var.project_name
  environment  = var.environment

  # Network configuration
  vpc_cidr    = var.vpc_cidr
  az_count    = var.az_count

  # Load balancer configuration
  ssl_certificate_arn        = var.ssl_certificate_arn
  enable_deletion_protection = var.enable_deletion_protection

  tags = local.common_tags
}

# 3. Database Module
# Creates RDS PostgreSQL instance with security and monitoring
module "database" {
  source = "../../modules/database"

  project_name           = var.project_name
  environment           = var.environment
  private_subnet_ids    = module.networking.private_subnet_ids
  rds_security_group_id = module.networking.rds_security_group_id

  # Database configuration
  db_name     = local.db_name
  db_username = local.db_username

  # Instance configuration
  instance_class        = var.db_instance_class
  allocated_storage     = var.db_allocated_storage
  max_allocated_storage = var.db_max_allocated_storage
  storage_type         = var.db_storage_type

  # Backup and maintenance
  backup_retention_period = var.db_backup_retention_period
  backup_window          = var.db_backup_window
  maintenance_window     = var.db_maintenance_window

  # Monitoring and performance
  monitoring_interval                    = var.db_monitoring_interval
  performance_insights_enabled          = var.db_performance_insights_enabled
  performance_insights_retention_period = var.db_performance_insights_retention_period

  # High availability
  create_read_replica    = var.db_create_read_replica
  replica_instance_class = var.db_replica_instance_class

  # Security
  deletion_protection        = var.db_deletion_protection
  auto_minor_version_upgrade = var.db_auto_minor_version_upgrade

  tags = local.common_tags

  depends_on = [module.networking]
}

# 4. API Fargate Module
# Creates ECS Fargate service for running the API container
module "api_fargate" {
  source = "../../modules/api-fargate"

  project_name           = var.project_name
  environment           = var.environment
  vpc_id                = module.networking.vpc_id
  private_subnet_ids    = module.networking.private_subnet_ids
  ecs_security_group_id = module.networking.ecs_tasks_security_group_id
  alb_listener_arn      = module.networking.alb_listener_http_arn

  # Container configuration
  container_image = var.container_image
  container_port  = local.container_port
  cpu            = var.cpu
  memory         = var.memory

  # Database connection
  database_host             = module.database.db_instance_address
  database_port             = module.database.db_instance_port
  database_name             = module.database.db_name
  database_user             = module.database.db_username
  db_password_secret_arn    = module.database.db_password_secret_arn

  # Environment variables
  environment_variables = var.environment_variables

  # Service configuration
  desired_count = var.desired_count

  # Auto scaling
  enable_auto_scaling = var.enable_auto_scaling
  min_capacity       = var.min_capacity
  max_capacity       = var.max_capacity
  cpu_target_value   = var.cpu_target_value
  memory_target_value = var.memory_target_value

  # Load balancer configuration
  host_headers            = var.host_headers
  listener_rule_priority  = var.listener_rule_priority

  # Health checks
  health_check_command              = var.health_check_command
  health_check_path                = local.health_check_path
  health_check_matcher             = var.health_check_matcher
  health_check_interval            = var.health_check_interval
  health_check_timeout             = var.health_check_timeout
  health_check_healthy_threshold   = var.health_check_healthy_threshold
  health_check_unhealthy_threshold = var.health_check_unhealthy_threshold

  # Monitoring
  log_retention_days         = var.log_retention_days
  enable_container_insights  = var.enable_container_insights
  enable_service_discovery   = var.enable_service_discovery

  tags = local.common_tags

  depends_on = [module.networking, module.database]
}

# Optional: CloudWatch Dashboard for monitoring
resource "aws_cloudwatch_dashboard" "api_dashboard" {
  count = var.create_monitoring_dashboard ? 1 : 0

  dashboard_name = "${var.project_name}-api-dashboard"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6

        properties = {
          metrics = [
            ["AWS/ECS", "CPUUtilization", "ServiceName", module.api_fargate.ecs_service_name, "ClusterName", module.api_fargate.ecs_cluster_name],
            [".", "MemoryUtilization", ".", ".", ".", "."],
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "ECS Service Metrics"
          period  = 300
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 12
        height = 6

        properties = {
          metrics = [
            ["AWS/ApplicationELB", "RequestCount", "LoadBalancer", module.networking.alb_dns_name],
            [".", "TargetResponseTime", ".", "."],
            [".", "HTTPCode_Target_2XX_Count", ".", "."],
            [".", "HTTPCode_Target_4XX_Count", ".", "."],
            [".", "HTTPCode_Target_5XX_Count", ".", "."],
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Load Balancer Metrics"
          period  = 300
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 12
        width  = 12
        height = 6

        properties = {
          metrics = [
            ["AWS/RDS", "CPUUtilization", "DBInstanceIdentifier", module.database.db_instance_id],
            [".", "DatabaseConnections", ".", "."],
            [".", "ReadLatency", ".", "."],
            [".", "WriteLatency", ".", "."],
          ]
          view    = "timeSeries"
          stacked = false
          region  = data.aws_region.current.name
          title   = "Database Metrics"
          period  = 300
        }
      }
    ]
  })

  tags = local.common_tags
}

# Optional: SNS Topic for alerts
resource "aws_sns_topic" "alerts" {
  count = var.create_alerting ? 1 : 0

  name = "${var.project_name}-alerts"

  tags = local.common_tags
}

resource "aws_sns_topic_subscription" "email_alerts" {
  count = var.create_alerting && var.alert_email != "" ? 1 : 0

  topic_arn = aws_sns_topic.alerts[0].arn
  protocol  = "email"
  endpoint  = var.alert_email
}

# CloudWatch Alarms
resource "aws_cloudwatch_metric_alarm" "high_cpu" {
  count = var.create_alerting ? 1 : 0

  alarm_name          = "${var.project_name}-high-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ECS"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "This metric monitors ECS CPU utilization"
  alarm_actions       = [aws_sns_topic.alerts[0].arn]

  dimensions = {
    ServiceName = module.api_fargate.ecs_service_name
    ClusterName = module.api_fargate.ecs_cluster_name
  }

  tags = local.common_tags
}

resource "aws_cloudwatch_metric_alarm" "high_memory" {
  count = var.create_alerting ? 1 : 0

  alarm_name          = "${var.project_name}-high-memory"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "MemoryUtilization"
  namespace           = "AWS/ECS"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "This metric monitors ECS memory utilization"
  alarm_actions       = [aws_sns_topic.alerts[0].arn]

  dimensions = {
    ServiceName = module.api_fargate.ecs_service_name
    ClusterName = module.api_fargate.ecs_cluster_name
  }

  tags = local.common_tags
}
