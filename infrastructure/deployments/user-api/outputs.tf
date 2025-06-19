# Outputs for API-Direct BYOA User API Deployment

# Cross-Account Access Information
output "cross_account_role_arn" {
  description = "ARN of the cross-account role for API-Direct to assume"
  value       = module.iam_cross_account.deployment_role_arn
}

output "external_id" {
  description = "External ID for cross-account role assumption"
  value       = module.iam_cross_account.external_id
  sensitive   = true
}

output "external_id_secret_arn" {
  description = "ARN of the secret containing the external ID"
  value       = module.iam_cross_account.external_id_secret_arn
}

# Networking Information
output "vpc_id" {
  description = "ID of the VPC"
  value       = module.networking.vpc_id
}

output "vpc_cidr" {
  description = "CIDR block of the VPC"
  value       = module.networking.vpc_cidr
}

output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = module.networking.public_subnet_ids
}

output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = module.networking.private_subnet_ids
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = module.networking.alb_dns_name
}

output "alb_zone_id" {
  description = "Zone ID of the Application Load Balancer"
  value       = module.networking.alb_zone_id
}

output "alb_arn" {
  description = "ARN of the Application Load Balancer"
  value       = module.networking.alb_arn
}

# Database Information
output "database_endpoint" {
  description = "RDS instance endpoint"
  value       = module.database.db_instance_address
}

output "database_port" {
  description = "RDS instance port"
  value       = module.database.db_instance_port
}

output "database_name" {
  description = "Database name"
  value       = module.database.db_name
}

output "database_username" {
  description = "Database username"
  value       = module.database.db_username
}

output "database_password_secret_arn" {
  description = "ARN of the secret containing the database password"
  value       = module.database.db_password_secret_arn
}

output "database_instance_id" {
  description = "RDS instance identifier"
  value       = module.database.db_instance_id
}

output "database_read_replica_endpoint" {
  description = "RDS read replica endpoint (if created)"
  value       = module.database.db_read_replica_address
}

# Container Service Information
output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = module.api_fargate.ecs_cluster_name
}

output "ecs_cluster_arn" {
  description = "ARN of the ECS cluster"
  value       = module.api_fargate.ecs_cluster_arn
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = module.api_fargate.ecs_service_name
}

output "ecs_service_arn" {
  description = "ARN of the ECS service"
  value       = module.api_fargate.ecs_service_arn
}

output "ecs_task_definition_arn" {
  description = "ARN of the ECS task definition"
  value       = module.api_fargate.ecs_task_definition_arn
}

output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = module.api_fargate.ecr_repository_url
}

output "ecr_repository_name" {
  description = "Name of the ECR repository"
  value       = module.api_fargate.ecr_repository_name
}

# Load Balancer Target Group
output "target_group_arn" {
  description = "ARN of the target group"
  value       = module.api_fargate.target_group_arn
}

output "target_group_name" {
  description = "Name of the target group"
  value       = module.api_fargate.target_group_name
}

# CloudWatch Logs
output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = module.api_fargate.cloudwatch_log_group_name
}

output "cloudwatch_log_group_arn" {
  description = "ARN of the CloudWatch log group"
  value       = module.api_fargate.cloudwatch_log_group_arn
}

# Auto Scaling Information
output "autoscaling_target_resource_id" {
  description = "Resource ID of the autoscaling target"
  value       = module.api_fargate.autoscaling_target_resource_id
}

output "autoscaling_cpu_policy_arn" {
  description = "ARN of the CPU autoscaling policy"
  value       = module.api_fargate.autoscaling_cpu_policy_arn
}

output "autoscaling_memory_policy_arn" {
  description = "ARN of the memory autoscaling policy"
  value       = module.api_fargate.autoscaling_memory_policy_arn
}

# Monitoring and Alerting
output "cloudwatch_dashboard_url" {
  description = "URL of the CloudWatch dashboard"
  value = var.create_monitoring_dashboard ? "https://${data.aws_region.current.name}.console.aws.amazon.com/cloudwatch/home?region=${data.aws_region.current.name}#dashboards:name=${var.project_name}-api-dashboard" : null
}

output "sns_topic_arn" {
  description = "ARN of the SNS topic for alerts"
  value       = var.create_alerting ? aws_sns_topic.alerts[0].arn : null
}

# API Access Information
output "api_url" {
  description = "URL to access the API"
  value       = "http://${module.networking.alb_dns_name}"
}

output "api_https_url" {
  description = "HTTPS URL to access the API (if SSL certificate is configured)"
  value       = var.ssl_certificate_arn != "" ? "https://${module.networking.alb_dns_name}" : null
}

# Deployment Information
output "deployment_info" {
  description = "Complete deployment information"
  value = {
    project_name    = var.project_name
    environment     = var.environment
    aws_region      = var.aws_region
    aws_account_id  = data.aws_caller_identity.current.account_id
    deployed_at     = timestamp()
    
    # Infrastructure
    vpc_id              = module.networking.vpc_id
    alb_dns_name        = module.networking.alb_dns_name
    database_endpoint   = module.database.db_instance_address
    ecs_cluster_name    = module.api_fargate.ecs_cluster_name
    ecs_service_name    = module.api_fargate.ecs_service_name
    ecr_repository_url  = module.api_fargate.ecr_repository_url
    
    # Access
    api_url             = "http://${module.networking.alb_dns_name}"
    api_https_url       = var.ssl_certificate_arn != "" ? "https://${module.networking.alb_dns_name}" : null
    
    # Monitoring
    log_group_name      = module.api_fargate.cloudwatch_log_group_name
    dashboard_url       = var.create_monitoring_dashboard ? "https://${data.aws_region.current.name}.console.aws.amazon.com/cloudwatch/home?region=${data.aws_region.current.name}#dashboards:name=${var.project_name}-api-dashboard" : null
    
    # Configuration
    container_port      = var.container_port
    desired_count       = var.desired_count
    cpu                 = var.cpu
    memory              = var.memory
    auto_scaling_enabled = var.enable_auto_scaling
  }
  sensitive = false
}

# Security Configuration
output "security_configuration" {
  description = "Security configuration summary"
  value = {
    cross_account_role_arn    = module.iam_cross_account.deployment_role_arn
    require_mfa              = var.require_mfa
    compliance_framework     = var.compliance_framework
    data_residency          = var.data_residency_requirements
    deletion_protection     = var.enable_deletion_protection
    database_encrypted      = true
    logs_encrypted          = true
    secrets_manager_enabled = true
  }
}

# Cost Information
output "estimated_monthly_cost" {
  description = "Estimated monthly cost breakdown (USD)"
  value = {
    # ECS Fargate costs (estimated based on configuration)
    ecs_fargate = {
      cpu_cost    = (var.cpu / 1024) * var.desired_count * 24 * 30 * 0.04048
      memory_cost = (var.memory / 1024) * var.desired_count * 24 * 30 * 0.004445
      total       = ((var.cpu / 1024) * var.desired_count * 24 * 30 * 0.04048) + ((var.memory / 1024) * var.desired_count * 24 * 30 * 0.004445)
    }
    
    # RDS costs (estimated based on instance class)
    rds = {
      instance_cost = var.db_instance_class == "db.t3.micro" ? 15.33 : 
                     var.db_instance_class == "db.t3.small" ? 30.66 :
                     var.db_instance_class == "db.t3.medium" ? 61.32 : 100.00
      storage_cost  = var.db_allocated_storage * 0.115
      backup_cost   = var.db_backup_retention_period > 0 ? var.db_allocated_storage * 0.095 : 0
      total         = (var.db_instance_class == "db.t3.micro" ? 15.33 : 
                      var.db_instance_class == "db.t3.small" ? 30.66 :
                      var.db_instance_class == "db.t3.medium" ? 61.32 : 100.00) + 
                     (var.db_allocated_storage * 0.115) + 
                     (var.db_backup_retention_period > 0 ? var.db_allocated_storage * 0.095 : 0)
    }
    
    # ALB costs
    alb = {
      fixed_cost = 16.20
      lcu_cost   = 5.00  # Estimated based on typical usage
      total      = 21.20
    }
    
    # VPC costs (NAT Gateway)
    vpc = {
      nat_gateway_cost = var.az_count * 32.40
      data_transfer    = 10.00  # Estimated
      total           = (var.az_count * 32.40) + 10.00
    }
    
    # CloudWatch costs
    cloudwatch = {
      logs_cost       = 5.00   # Estimated based on typical API usage
      metrics_cost    = 2.00   # Custom metrics
      dashboard_cost  = var.create_monitoring_dashboard ? 3.00 : 0
      alarms_cost     = var.create_alerting ? 1.00 : 0
      total          = 5.00 + 2.00 + (var.create_monitoring_dashboard ? 3.00 : 0) + (var.create_alerting ? 1.00 : 0)
    }
    
    # Other AWS services
    other = {
      secrets_manager = 0.40
      ecr_storage    = 1.00   # Estimated for container images
      sns            = var.create_alerting ? 0.50 : 0
      total          = 0.40 + 1.00 + (var.create_alerting ? 0.50 : 0)
    }
    
    # Total estimated cost
    total_estimated = ((var.cpu / 1024) * var.desired_count * 24 * 30 * 0.04048) + 
                     ((var.memory / 1024) * var.desired_count * 24 * 30 * 0.004445) +
                     (var.db_instance_class == "db.t3.micro" ? 15.33 : 
                      var.db_instance_class == "db.t3.small" ? 30.66 :
                      var.db_instance_class == "db.t3.medium" ? 61.32 : 100.00) + 
                     (var.db_allocated_storage * 0.115) + 
                     (var.db_backup_retention_period > 0 ? var.db_allocated_storage * 0.095 : 0) +
                     21.20 + (var.az_count * 32.40) + 10.00 + 8.00 + 1.90
    
    note = "Costs are estimates based on us-east-1 pricing and may vary based on actual usage, region, and AWS pricing changes"
  }
}

# Terraform State Information
output "terraform_state_info" {
  description = "Information about Terraform state management"
  value = {
    backend_bucket     = "${var.project_name}-terraform-state-${data.aws_caller_identity.current.account_id}"
    backend_key        = "${var.project_name}/terraform.tfstate"
    backend_region     = data.aws_region.current.name
    dynamodb_table     = "${var.project_name}-terraform-locks"
    state_encryption   = true
  }
}

# Next Steps
output "next_steps" {
  description = "Next steps for completing the deployment"
  value = {
    step_1 = "Configure your application container image and push to ECR: ${module.api_fargate.ecr_repository_url}"
    step_2 = "Update ECS service to deploy your container image"
    step_3 = "Configure DNS to point to the load balancer: ${module.networking.alb_dns_name}"
    step_4 = "Set up SSL certificate if HTTPS is required"
    step_5 = "Configure monitoring alerts and notifications"
    step_6 = "Test your API endpoints and health checks"
    step_7 = "Set up CI/CD pipeline for automated deployments"
  }
}

# API-Direct Integration Information
output "api_direct_integration" {
  description = "Information needed for API-Direct platform integration"
  value = {
    # Cross-account access
    role_arn           = module.iam_cross_account.deployment_role_arn
    external_id_secret = module.iam_cross_account.external_id_secret_arn
    
    # Infrastructure endpoints
    api_endpoint       = "http://${module.networking.alb_dns_name}"
    api_https_endpoint = var.ssl_certificate_arn != "" ? "https://${module.networking.alb_dns_name}" : null
    
    # Container registry
    ecr_repository     = module.api_fargate.ecr_repository_url
    
    # Monitoring
    log_group          = module.api_fargate.cloudwatch_log_group_name
    dashboard_url      = var.create_monitoring_dashboard ? "https://${data.aws_region.current.name}.console.aws.amazon.com/cloudwatch/home?region=${data.aws_region.current.name}#dashboards:name=${var.project_name}-api-dashboard" : null
    
    # Deployment targets
    ecs_cluster        = module.api_fargate.ecs_cluster_name
    ecs_service        = module.api_fargate.ecs_service_name
    target_group       = module.api_fargate.target_group_arn
    
    # Database
    database_endpoint  = module.database.db_instance_address
    database_secret    = module.database.db_password_secret_arn
    
    # Configuration
    aws_region         = data.aws_region.current.name
    aws_account_id     = data.aws_caller_identity.current.account_id
    project_name       = var.project_name
    environment        = var.environment
  }
  sensitive = false
}
