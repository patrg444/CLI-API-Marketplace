# Variables for IAM Cross-Account Module

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

variable "api_direct_account_id" {
  description = "AWS Account ID of the API-Direct platform"
  type        = string
  validation {
    condition     = can(regex("^[0-9]{12}$", var.api_direct_account_id))
    error_message = "API-Direct account ID must be a 12-digit AWS account ID."
  }
}

variable "allowed_user_patterns" {
  description = "List of user ID patterns allowed to assume the role"
  type        = list(string)
  default     = ["*"]
  validation {
    condition     = length(var.allowed_user_patterns) > 0
    error_message = "At least one user pattern must be specified."
  }
}

variable "max_session_duration" {
  description = "Maximum session duration in seconds for the assumed role"
  type        = number
  default     = 3600
  validation {
    condition     = var.max_session_duration >= 900 && var.max_session_duration <= 43200
    error_message = "Max session duration must be between 900 (15 minutes) and 43200 (12 hours) seconds."
  }
}

# Feature Flags
variable "enable_service_discovery" {
  description = "Enable Service Discovery management permissions"
  type        = bool
  default     = false
}

variable "enable_cloudformation_access" {
  description = "Enable CloudFormation stack management permissions"
  type        = bool
  default     = false
}

variable "enable_vpc_management" {
  description = "Enable VPC and networking resource management permissions"
  type        = bool
  default     = false
}

variable "enable_rds_management" {
  description = "Enable RDS database management permissions"
  type        = bool
  default     = false
}

# Security Configuration
variable "require_mfa" {
  description = "Require MFA for role assumption"
  type        = bool
  default     = true
}

variable "allowed_source_ips" {
  description = "List of IP addresses/CIDR blocks allowed to assume the role"
  type        = list(string)
  default     = []
  validation {
    condition = alltrue([
      for ip in var.allowed_source_ips : can(cidrhost(ip, 0))
    ])
    error_message = "All IP addresses must be valid CIDR blocks."
  }
}

variable "session_name_prefix" {
  description = "Prefix for role session names"
  type        = string
  default     = "api-direct-deployment"
  validation {
    condition     = can(regex("^[a-zA-Z0-9+=,.@_-]+$", var.session_name_prefix))
    error_message = "Session name prefix must contain only alphanumeric characters and +=,.@_-"
  }
}

# Resource Scope Configuration
variable "allowed_regions" {
  description = "List of AWS regions where resources can be managed"
  type        = list(string)
  default     = []
  validation {
    condition = alltrue([
      for region in var.allowed_regions : can(regex("^[a-z0-9-]+$", region))
    ])
    error_message = "All regions must be valid AWS region names."
  }
}

variable "resource_tags_required" {
  description = "Required tags that must be present on all managed resources"
  type        = map(string)
  default = {
    ManagedBy = "api-direct"
  }
}

# Cost Control
variable "enable_cost_controls" {
  description = "Enable cost control policies"
  type        = bool
  default     = true
}

variable "max_monthly_spend" {
  description = "Maximum monthly spend limit in USD (0 = no limit)"
  type        = number
  default     = 0
  validation {
    condition     = var.max_monthly_spend >= 0
    error_message = "Max monthly spend must be non-negative."
  }
}

# Monitoring and Logging
variable "enable_cloudtrail_logging" {
  description = "Enable CloudTrail logging for role usage"
  type        = bool
  default     = true
}

variable "notification_topic_arn" {
  description = "SNS topic ARN for notifications about role usage"
  type        = string
  default     = ""
}

# Advanced Security
variable "external_id_rotation_days" {
  description = "Number of days after which external ID should be rotated"
  type        = number
  default     = 90
  validation {
    condition     = var.external_id_rotation_days >= 1 && var.external_id_rotation_days <= 365
    error_message = "External ID rotation days must be between 1 and 365."
  }
}

variable "enable_temporary_credentials" {
  description = "Enable temporary credential generation"
  type        = bool
  default     = true
}

variable "credential_duration_hours" {
  description = "Duration in hours for temporary credentials"
  type        = number
  default     = 1
  validation {
    condition     = var.credential_duration_hours >= 1 && var.credential_duration_hours <= 12
    error_message = "Credential duration must be between 1 and 12 hours."
  }
}

# Compliance and Governance
variable "compliance_framework" {
  description = "Compliance framework to adhere to (SOC2, HIPAA, PCI-DSS, etc.)"
  type        = string
  default     = "SOC2"
  validation {
    condition     = contains(["SOC2", "HIPAA", "PCI-DSS", "ISO27001", "GDPR", "NONE"], var.compliance_framework)
    error_message = "Compliance framework must be one of: SOC2, HIPAA, PCI-DSS, ISO27001, GDPR, NONE."
  }
}

variable "data_residency_requirements" {
  description = "Data residency requirements for the deployment"
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
  validation {
    condition = var.webhook_url == "" || can(regex("^https://", var.webhook_url))
    error_message = "Webhook URL must be empty or start with https://"
  }
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

# Backup and Recovery
variable "enable_backup_policies" {
  description = "Enable automatic backup policies for managed resources"
  type        = bool
  default     = true
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 30
  validation {
    condition     = var.backup_retention_days >= 1 && var.backup_retention_days <= 365
    error_message = "Backup retention days must be between 1 and 365."
  }
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
