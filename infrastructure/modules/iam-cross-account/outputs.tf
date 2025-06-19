# Outputs for IAM Cross-Account Module

# IAM Role Information
output "deployment_role_arn" {
  description = "ARN of the deployment role for API-Direct to assume"
  value       = aws_iam_role.api_direct_deployment.arn
}

output "deployment_role_name" {
  description = "Name of the deployment role"
  value       = aws_iam_role.api_direct_deployment.name
}

output "deployment_role_unique_id" {
  description = "Unique ID of the deployment role"
  value       = aws_iam_role.api_direct_deployment.unique_id
}

# External ID Information
output "external_id" {
  description = "External ID for cross-account role assumption"
  value       = random_uuid.external_id.result
  sensitive   = true
}

output "external_id_secret_arn" {
  description = "ARN of the secret containing the external ID"
  value       = aws_secretsmanager_secret.external_id.arn
}

output "external_id_secret_name" {
  description = "Name of the secret containing the external ID"
  value       = aws_secretsmanager_secret.external_id.name
}

# Account Information
output "account_id" {
  description = "AWS Account ID where the role is created"
  value       = data.aws_caller_identity.current.account_id
}

output "region" {
  description = "AWS Region where the role is created"
  value       = data.aws_region.current.name
}

# Policy ARNs
output "ecs_management_policy_arn" {
  description = "ARN of the ECS management policy"
  value       = aws_iam_policy.ecs_management.arn
}

output "alb_management_policy_arn" {
  description = "ARN of the ALB management policy"
  value       = aws_iam_policy.alb_management.arn
}

output "iam_management_policy_arn" {
  description = "ARN of the IAM management policy"
  value       = aws_iam_policy.iam_management.arn
}

output "logs_management_policy_arn" {
  description = "ARN of the CloudWatch Logs management policy"
  value       = aws_iam_policy.logs_management.arn
}

output "autoscaling_management_policy_arn" {
  description = "ARN of the Auto Scaling management policy"
  value       = aws_iam_policy.autoscaling_management.arn
}

output "secrets_access_policy_arn" {
  description = "ARN of the Secrets Manager access policy"
  value       = aws_iam_policy.secrets_access.arn
}

output "service_discovery_management_policy_arn" {
  description = "ARN of the Service Discovery management policy"
  value       = var.enable_service_discovery ? aws_iam_policy.service_discovery_management[0].arn : null
}

output "cloudformation_management_policy_arn" {
  description = "ARN of the CloudFormation management policy"
  value       = var.enable_cloudformation_access ? aws_iam_policy.cloudformation_management[0].arn : null
}

# Role Assumption Information
output "role_assumption_command" {
  description = "AWS CLI command to assume the role"
  value = "aws sts assume-role --role-arn ${aws_iam_role.api_direct_deployment.arn} --role-session-name ${var.session_name_prefix}-session --external-id ${random_uuid.external_id.result}"
  sensitive = true
}

output "role_assumption_info" {
  description = "Complete information needed for role assumption"
  value = {
    role_arn           = aws_iam_role.api_direct_deployment.arn
    external_id        = random_uuid.external_id.result
    session_name       = "${var.session_name_prefix}-session"
    max_duration       = var.max_session_duration
    account_id         = data.aws_caller_identity.current.account_id
    region             = data.aws_region.current.name
    api_direct_account = var.api_direct_account_id
  }
  sensitive = true
}

# Security Configuration
output "security_configuration" {
  description = "Security configuration for the cross-account setup"
  value = {
    require_mfa                = var.require_mfa
    max_session_duration       = var.max_session_duration
    external_id_rotation_days  = var.external_id_rotation_days
    allowed_user_patterns      = var.allowed_user_patterns
    allowed_source_ips         = var.allowed_source_ips
    compliance_framework       = var.compliance_framework
    data_residency            = var.data_residency_requirements
  }
  sensitive = false
}

# Feature Flags
output "enabled_features" {
  description = "List of enabled features for the deployment role"
  value = {
    service_discovery      = var.enable_service_discovery
    cloudformation_access  = var.enable_cloudformation_access
    vpc_management        = var.enable_vpc_management
    rds_management        = var.enable_rds_management
    cost_controls         = var.enable_cost_controls
    cloudtrail_logging    = var.enable_cloudtrail_logging
    temporary_credentials = var.enable_temporary_credentials
    backup_policies       = var.enable_backup_policies
  }
}

# Integration Information
output "integration_config" {
  description = "Configuration for API-Direct integration"
  value = {
    webhook_url           = var.webhook_url
    notification_topic    = var.notification_topic_arn
    service_endpoints     = var.api_direct_service_endpoints
    max_monthly_spend     = var.max_monthly_spend
    backup_retention_days = var.backup_retention_days
  }
  sensitive = false
}

# Resource Scope
output "resource_scope" {
  description = "Scope of resources that can be managed"
  value = {
    allowed_regions      = var.allowed_regions
    required_tags        = var.resource_tags_required
    project_name         = var.project_name
    environment          = var.environment
  }
}

# Deployment Instructions
output "deployment_instructions" {
  description = "Instructions for API-Direct to use this cross-account setup"
  value = {
    step_1 = "Store the external_id and role_arn in API-Direct's secure configuration"
    step_2 = "Configure API-Direct deployment service to assume the role using the external_id"
    step_3 = "Use the role to deploy infrastructure modules (networking, database, api-fargate)"
    step_4 = "Monitor role usage through CloudTrail and configured notifications"
    step_5 = "Rotate external_id every ${var.external_id_rotation_days} days for security"
  }
}

# Terraform State Information
output "terraform_backend_config" {
  description = "Recommended Terraform backend configuration for this deployment"
  value = {
    backend = "s3"
    config = {
      bucket         = "${var.project_name}-terraform-state-${data.aws_caller_identity.current.account_id}"
      key            = "${var.project_name}/terraform.tfstate"
      region         = data.aws_region.current.name
      encrypt        = true
      dynamodb_table = "${var.project_name}-terraform-locks"
    }
  }
}

# Cost Estimation
output "estimated_monthly_cost" {
  description = "Estimated monthly cost for the IAM resources (USD)"
  value = {
    iam_roles_and_policies = 0.00
    secrets_manager        = 0.40
    cloudtrail_logging     = var.enable_cloudtrail_logging ? 2.00 : 0.00
    total_estimated        = var.enable_cloudtrail_logging ? 2.40 : 0.40
    note                   = "Costs are estimates and may vary based on usage"
  }
}

# Compliance Information
output "compliance_status" {
  description = "Compliance status and requirements"
  value = {
    framework              = var.compliance_framework
    data_residency        = var.data_residency_requirements
    encryption_at_rest    = true
    encryption_in_transit = true
    audit_logging         = var.enable_cloudtrail_logging
    access_controls       = "least-privilege"
    mfa_required          = var.require_mfa
  }
}

# Monitoring and Alerting
output "monitoring_setup" {
  description = "Monitoring and alerting configuration"
  value = {
    cloudtrail_enabled    = var.enable_cloudtrail_logging
    notification_topic    = var.notification_topic_arn
    webhook_url          = var.webhook_url
    cost_monitoring      = var.enable_cost_controls
    max_monthly_spend    = var.max_monthly_spend
  }
}
