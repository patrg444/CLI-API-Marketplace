# Variables for API-Direct BYOA User API Deployment

# Core Configuration
variable "project_name" {
  description = "Name of the project (used for resource naming)"
  type        = string
  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.project_name))
    error_message = "Project name must contain only lowercase letters, numbers, and hyphens."
  }
}

variable "environment" {
  description = "Environment name (e.g., dev, staging, prod)"
  type        = string
  default     = "prod"
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod."
  }
}

variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-east-1"
  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.aws_region))
    error_message = "AWS region must be a valid region name."
  }
}

variable "owner_email" {
  description = "Email address of the API owner"
  type        = string
  validation {
    condition     = can(regex("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", var.owner_email))
    error_message = "Owner email must be a valid email address."
  }
}

variable "api_direct_account_id" {
  description = "AWS Account ID of the API-Direct platform"
  type        = string
  validation {
    condition     = can(regex("^[0-9]{12}$", var.api_direct_account_id))
    error_message = "API-Direct account ID must be a 12-digit AWS account ID."
  }
}

# Networking Configuration
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
  validation {
    condition     = can(cidrhost(var.vpc_cidr, 0))
    error_message = "VPC CIDR must be a valid CIDR block."
  }
}

variable "az_count" {
  description = "Number of Availability Zones to use"
  type        = number
  default     = 2
  validation {
    condition     = var.az_count >= 2 && var.az_count <= 6
    error_message = "AZ count must be between 2 and 6."
  }
}

variable "ssl_certificate_arn" {
  description = "ARN of the SSL certificate for HTTPS"
  type        = string
  default     = ""
}

variable "enable_deletion_protection" {
  description = "Enable deletion protection for the load balancer"
  type        = bool
  default     = true
}

# Database Configuration
variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
  validation {
    condition     = can(regex("^db\\.", var.db_instance_class))
    error_message = "DB instance class must start with 'db.'."
  }
}

variable "db_allocated_storage" {
  description = "Initial allocated storage for RDS instance (GB)"
  type        = number
  default     = 20
  validation {
    condition     = var.db_allocated_storage >= 20 && var.db_allocated_storage <= 65536
    error_message = "DB allocated storage must be between 20 and 65536 GB."
  }
}

variable "db_max_allocated_storage" {
  description = "Maximum allocated storage for RDS instance (GB)"
  type        = number
  default     = 100
  validation {
    condition     = var.db_max_allocated_storage >= 20 && var.db_max_allocated_storage <= 65536
    error_message = "DB max allocated storage must be between 20 and 65536 GB."
  }
}

variable "db_storage_type" {
  description = "Storage type for RDS instance"
  type        = string
  default     = "gp3"
  validation {
    condition     = contains(["gp2", "gp3", "io1", "io2"], var.db_storage_type)
    error_message = "DB storage type must be one of: gp2, gp3, io1, io2."
  }
}

variable "db_backup_retention_period" {
  description = "Backup retention period in days"
  type        = number
  default     = 7
  validation {
    condition     = var.db_backup_retention_period >= 0 && var.db_backup_retention_period <= 35
    error_message = "Backup retention period must be between 0 and 35 days."
  }
}

variable "db_backup_window" {
  description = "Preferred backup window"
  type        = string
  default     = "03:00-04:00"
}

variable "db_maintenance_window" {
  description = "Preferred maintenance window"
  type        = string
  default     = "sun:04:00-sun:05:00"
}

variable "db_monitoring_interval" {
  description = "Enhanced monitoring interval in seconds"
  type        = number
  default     = 60
  validation {
    condition     = contains([0, 1, 5, 10, 15, 30, 60], var.db_monitoring_interval)
    error_message = "Monitoring interval must be one of: 0, 1, 5, 10, 15, 30, 60."
  }
}

variable "db_performance_insights_enabled" {
  description = "Enable Performance Insights"
  type        = bool
  default     = true
}

variable "db_performance_insights_retention_period" {
  description = "Performance Insights retention period in days"
  type        = number
  default     = 7
  validation {
    condition     = contains([7, 731], var.db_performance_insights_retention_period)
    error_message = "Performance Insights retention period must be 7 or 731 days."
  }
}

variable "db_create_read_replica" {
  description = "Create a read replica for the database"
  type        = bool
  default     = false
}

variable "db_replica_instance_class" {
  description = "Instance class for the read replica"
  type        = string
  default     = "db.t3.micro"
}

variable "db_deletion_protection" {
  description = "Enable deletion protection for the database"
  type        = bool
  default     = true
}

variable "db_auto_minor_version_upgrade" {
  description = "Enable automatic minor version upgrades"
  type        = bool
  default     = true
}

# Container Configuration
variable "container_image" {
  description = "Container image to deploy (will use ECR if not specified)"
  type        = string
  default     = ""
}

variable "container_port" {
  description = "Port that the container exposes"
  type        = number
  default     = 8000
  validation {
    condition     = var.container_port > 0 && var.container_port <= 65535
    error_message = "Container port must be between 1 and 65535."
  }
}

variable "cpu" {
  description = "CPU units for the container (256, 512, 1024, 2048, 4096)"
  type        = number
  default     = 256
  validation {
    condition     = contains([256, 512, 1024, 2048, 4096], var.cpu)
    error_message = "CPU must be one of: 256, 512, 1024, 2048, 4096."
  }
}

variable "memory" {
  description = "Memory in MB for the container"
  type        = number
  default     = 512
  validation {
    condition     = var.memory >= 512 && var.memory <= 30720
    error_message = "Memory must be between 512 and 30720 MB."
  }
}

variable "desired_count" {
  description = "Desired number of tasks"
  type        = number
  default     = 2
  validation {
    condition     = var.desired_count >= 1 && var.desired_count <= 100
    error_message = "Desired count must be between 1 and 100."
  }
}

# Auto Scaling Configuration
variable "enable_auto_scaling" {
  description = "Enable auto scaling for the ECS service"
  type        = bool
  default     = true
}

variable "min_capacity" {
  description = "Minimum number of tasks"
  type        = number
  default     = 1
  validation {
    condition     = var.min_capacity >= 1 && var.min_capacity <= 100
    error_message = "Min capacity must be between 1 and 100."
  }
}

variable "max_capacity" {
  description = "Maximum number of tasks"
  type        = number
  default     = 10
  validation {
    condition     = var.max_capacity >= 1 && var.max_capacity <= 100
    error_message = "Max capacity must be between 1 and 100."
  }
}

variable "cpu_target_value" {
  description = "Target CPU utilization percentage for auto scaling"
  type        = number
  default     = 70
  validation {
    condition     = var.cpu_target_value >= 10 && var.cpu_target_value <= 90
    error_message = "CPU target value must be between 10 and 90."
  }
}

variable "memory_target_value" {
  description = "Target memory utilization percentage for auto scaling"
  type        = number
  default     = 80
  validation {
    condition     = var.memory_target_value >= 10 && var.memory_target_value <= 90
    error_message = "Memory target value must be between 10 and 90."
  }
}

# Load Balancer Configuration
variable "host_headers" {
  description = "Host headers for routing traffic to this API"
  type        = list(string)
  default     = []
}

variable "listener_rule_priority" {
  description = "Priority for the ALB listener rule"
  type        = number
  default     = 100
  validation {
    condition     = var.listener_rule_priority >= 1 && var.listener_rule_priority <= 50000
    error_message = "Listener rule priority must be between 1 and 50000."
  }
}

# Health Check Configuration
variable "health_check_command" {
  description = "Command to run for container health checks"
  type        = list(string)
  default     = ["CMD-SHELL", "curl -f http://localhost:8000/health || exit 1"]
}

variable "health_check_path" {
  description = "Path for ALB health checks"
  type        = string
  default     = "/health"
}

variable "health_check_matcher" {
  description = "HTTP response codes for successful health checks"
  type        = string
  default     = "200"
}

variable "health_check_interval" {
  description = "Health check interval in seconds"
  type        = number
  default     = 30
  validation {
    condition     = var.health_check_interval >= 5 && var.health_check_interval <= 300
    error_message = "Health check interval must be between 5 and 300 seconds."
  }
}

variable "health_check_timeout" {
  description = "Health check timeout in seconds"
  type        = number
  default     = 5
  validation {
    condition     = var.health_check_timeout >= 2 && var.health_check_timeout <= 120
    error_message = "Health check timeout must be between 2 and 120 seconds."
  }
}

variable "health_check_healthy_threshold" {
  description = "Number of consecutive successful health checks"
  type        = number
  default     = 2
  validation {
    condition     = var.health_check_healthy_threshold >= 2 && var.health_check_healthy_threshold <= 10
    error_message = "Healthy threshold must be between 2 and 10."
  }
}

variable "health_check_unhealthy_threshold" {
  description = "Number of consecutive failed health checks"
  type        = number
  default     = 2
  validation {
    condition     = var.health_check_unhealthy_threshold >= 2 && var.health_check_unhealthy_threshold <= 10
    error_message = "Unhealthy threshold must be between 2 and 10."
  }
}

# Environment Variables
variable "environment_variables" {
  description = "Environment variables for the container"
  type = list(object({
    name  = string
    value = string
  }))
  default = []
}

# Monitoring Configuration
variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 14
  validation {
    condition = contains([
      1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653
    ], var.log_retention_days)
    error_message = "Log retention days must be a valid CloudWatch retention period."
  }
}

variable "enable_container_insights" {
  description = "Enable Container Insights for the ECS cluster"
  type        = bool
  default     = true
}

variable "enable_service_discovery" {
  description = "Enable service discovery for the ECS service"
  type        = bool
  default     = false
}

variable "create_monitoring_dashboard" {
  description = "Create a CloudWatch dashboard for monitoring"
  type        = bool
  default     = true
}

variable "create_alerting" {
  description = "Create CloudWatch alarms and SNS notifications"
  type        = bool
  default     = true
}

variable "alert_email" {
  description = "Email address for alerts"
  type        = string
  default     = ""
}

# Security Configuration
variable "allowed_user_patterns" {
  description = "List of user ID patterns allowed to assume the cross-account role"
  type        = list(string)
  default     = ["*"]
}

variable "max_session_duration" {
  description = "Maximum session duration in seconds for the cross-account role"
  type        = number
  default     = 3600
  validation {
    condition     = var.max_session_duration >= 900 && var.max_session_duration <= 43200
    error_message = "Max session duration must be between 900 and 43200 seconds."
  }
}

variable "require_mfa" {
  description = "Require MFA for cross-account role assumption"
  type        = bool
  default     = true
}

variable "allowed_source_ips" {
  description = "List of IP addresses/CIDR blocks allowed to assume the role"
  type        = list(string)
  default     = []
}

# Feature Flags
variable "enable_cloudformation_access" {
  description = "Enable CloudFormation access for the cross-account role"
  type        = bool
  default     = false
}

variable "enable_cost_controls" {
  description = "Enable cost control policies"
  type        = bool
  default     = true
}

# Compliance Configuration
variable "compliance_framework" {
  description = "Compliance framework to adhere to"
  type        = string
  default     = "SOC2"
  validation {
    condition     = contains(["SOC2", "HIPAA", "PCI-DSS", "ISO27001", "GDPR", "NONE"], var.compliance_framework)
    error_message = "Compliance framework must be one of: SOC2, HIPAA, PCI-DSS, ISO27001, GDPR, NONE."
  }
}

variable "data_residency_requirements" {
  description = "Data residency requirements"
  type        = string
  default     = "US"
  validation {
    condition     = contains(["US", "EU", "APAC", "GLOBAL"], var.data_residency_requirements)
    error_message = "Data residency must be one of: US, EU, APAC, GLOBAL."
  }
}

# Integration Configuration
variable "webhook_url" {
  description = "Webhook URL for deployment notifications"
  type        = string
  default     = ""
}

variable "api_direct_service_endpoints" {
  description = "API-Direct service endpoints for integration"
  type = object({
    deployment_service = string
    monitoring_service = string
    billing_service    = string
  })
  default = {
    deployment_service = ""
    monitoring_service = ""
    billing_service    = ""
  }
}

# Additional Tags
variable "additional_tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
