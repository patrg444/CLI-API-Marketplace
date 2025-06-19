# Variables for API Fargate Module

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

# Networking Configuration
variable "vpc_id" {
  description = "VPC ID where resources will be created"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for ECS tasks"
  type        = list(string)
  validation {
    condition     = length(var.private_subnet_ids) >= 2
    error_message = "At least 2 private subnets are required for high availability."
  }
}

variable "ecs_security_group_id" {
  description = "Security group ID for ECS tasks"
  type        = string
}

variable "alb_listener_arn" {
  description = "ARN of the ALB listener to attach the target group"
  type        = string
}

# Container Configuration
variable "container_image" {
  description = "Container image to deploy (if empty, uses ECR repository)"
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
  description = "CPU units for the task (256, 512, 1024, 2048, 4096)"
  type        = number
  default     = 256
  validation {
    condition     = contains([256, 512, 1024, 2048, 4096], var.cpu)
    error_message = "CPU must be one of: 256, 512, 1024, 2048, 4096."
  }
}

variable "memory" {
  description = "Memory in MB for the task"
  type        = number
  default     = 512
  validation {
    condition     = var.memory >= 512 && var.memory <= 30720
    error_message = "Memory must be between 512 and 30720 MB."
  }
}

# Database Configuration
variable "database_host" {
  description = "Database hostname"
  type        = string
}

variable "database_port" {
  description = "Database port"
  type        = number
  default     = 5432
}

variable "database_name" {
  description = "Database name"
  type        = string
}

variable "database_user" {
  description = "Database username"
  type        = string
}

variable "db_password_secret_arn" {
  description = "ARN of the secret containing database password"
  type        = string
}

# Environment Variables
variable "environment_variables" {
  description = "Additional environment variables for the container"
  type = list(object({
    name  = string
    value = string
  }))
  default = []
}

# Service Configuration
variable "desired_count" {
  description = "Desired number of tasks"
  type        = number
  default     = 1
  validation {
    condition     = var.desired_count >= 0 && var.desired_count <= 100
    error_message = "Desired count must be between 0 and 100."
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
    condition     = var.min_capacity >= 0 && var.min_capacity <= 100
    error_message = "Min capacity must be between 0 and 100."
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
  description = "Host headers for ALB listener rule"
  type        = list(string)
  default     = ["*"]
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
  description = "Health check command for the container"
  type        = string
  default     = "curl -f http://localhost:8000/health || exit 1"
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

variable "health_check_retries" {
  description = "Number of health check retries"
  type        = number
  default     = 3
  validation {
    condition     = var.health_check_retries >= 1 && var.health_check_retries <= 10
    error_message = "Health check retries must be between 1 and 10."
  }
}

variable "health_check_start_period" {
  description = "Health check start period in seconds"
  type        = number
  default     = 60
  validation {
    condition     = var.health_check_start_period >= 0 && var.health_check_start_period <= 300
    error_message = "Health check start period must be between 0 and 300 seconds."
  }
}

variable "health_check_path" {
  description = "Health check path for ALB target group"
  type        = string
  default     = "/health"
}

variable "health_check_matcher" {
  description = "Health check response codes"
  type        = string
  default     = "200"
}

variable "health_check_healthy_threshold" {
  description = "Number of consecutive successful health checks"
  type        = number
  default     = 2
  validation {
    condition     = var.health_check_healthy_threshold >= 2 && var.health_check_healthy_threshold <= 10
    error_message = "Health check healthy threshold must be between 2 and 10."
  }
}

variable "health_check_unhealthy_threshold" {
  description = "Number of consecutive failed health checks"
  type        = number
  default     = 3
  validation {
    condition     = var.health_check_unhealthy_threshold >= 2 && var.health_check_unhealthy_threshold <= 10
    error_message = "Health check unhealthy threshold must be between 2 and 10."
  }
}

# Logging Configuration
variable "log_retention_days" {
  description = "CloudWatch log retention period in days"
  type        = number
  default     = 7
  validation {
    condition = contains([
      1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653
    ], var.log_retention_days)
    error_message = "Log retention days must be a valid CloudWatch retention period."
  }
}

# Monitoring Configuration
variable "enable_container_insights" {
  description = "Enable Container Insights for the ECS cluster"
  type        = bool
  default     = true
}

# Service Discovery Configuration
variable "enable_service_discovery" {
  description = "Enable service discovery for the ECS service"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
