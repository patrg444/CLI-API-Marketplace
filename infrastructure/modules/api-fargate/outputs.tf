# Outputs for API Fargate Module

# ECR Repository
output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = aws_ecr_repository.api.repository_url
}

output "ecr_repository_arn" {
  description = "ARN of the ECR repository"
  value       = aws_ecr_repository.api.arn
}

output "ecr_repository_name" {
  description = "Name of the ECR repository"
  value       = aws_ecr_repository.api.name
}

# ECS Cluster
output "ecs_cluster_id" {
  description = "ID of the ECS cluster"
  value       = aws_ecs_cluster.main.id
}

output "ecs_cluster_arn" {
  description = "ARN of the ECS cluster"
  value       = aws_ecs_cluster.main.arn
}

output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = aws_ecs_cluster.main.name
}

# ECS Service
output "ecs_service_id" {
  description = "ID of the ECS service"
  value       = aws_ecs_service.api.id
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.api.name
}

output "ecs_service_arn" {
  description = "ARN of the ECS service"
  value       = aws_ecs_service.api.id
}

# ECS Task Definition
output "ecs_task_definition_arn" {
  description = "ARN of the ECS task definition"
  value       = aws_ecs_task_definition.api.arn
}

output "ecs_task_definition_family" {
  description = "Family of the ECS task definition"
  value       = aws_ecs_task_definition.api.family
}

output "ecs_task_definition_revision" {
  description = "Revision of the ECS task definition"
  value       = aws_ecs_task_definition.api.revision
}

# IAM Roles
output "ecs_task_execution_role_arn" {
  description = "ARN of the ECS task execution role"
  value       = aws_iam_role.ecs_task_execution.arn
}

output "ecs_task_role_arn" {
  description = "ARN of the ECS task role"
  value       = aws_iam_role.ecs_task.arn
}

# Load Balancer
output "target_group_arn" {
  description = "ARN of the target group"
  value       = aws_lb_target_group.api.arn
}

output "target_group_name" {
  description = "Name of the target group"
  value       = aws_lb_target_group.api.name
}

output "listener_rule_arn" {
  description = "ARN of the ALB listener rule"
  value       = aws_lb_listener_rule.api.arn
}

# CloudWatch Logs
output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.api.name
}

output "cloudwatch_log_group_arn" {
  description = "ARN of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.api.arn
}

output "cloudwatch_exec_log_group_name" {
  description = "Name of the CloudWatch exec log group"
  value       = aws_cloudwatch_log_group.ecs_exec.name
}

# Auto Scaling
output "autoscaling_target_resource_id" {
  description = "Resource ID of the autoscaling target"
  value       = var.enable_auto_scaling ? aws_appautoscaling_target.ecs_target[0].resource_id : null
}

output "autoscaling_cpu_policy_arn" {
  description = "ARN of the CPU autoscaling policy"
  value       = var.enable_auto_scaling ? aws_appautoscaling_policy.ecs_cpu_policy[0].arn : null
}

output "autoscaling_memory_policy_arn" {
  description = "ARN of the memory autoscaling policy"
  value       = var.enable_auto_scaling ? aws_appautoscaling_policy.ecs_memory_policy[0].arn : null
}

# Service Discovery (conditional)
output "service_discovery_namespace_id" {
  description = "ID of the service discovery namespace"
  value       = var.enable_service_discovery ? aws_service_discovery_private_dns_namespace.main[0].id : null
}

output "service_discovery_service_id" {
  description = "ID of the service discovery service"
  value       = var.enable_service_discovery ? aws_service_discovery_service.api[0].id : null
}

output "service_discovery_service_arn" {
  description = "ARN of the service discovery service"
  value       = var.enable_service_discovery ? aws_service_discovery_service.api[0].arn : null
}

# Container Configuration
output "container_port" {
  description = "Port that the container exposes"
  value       = var.container_port
}

output "container_image" {
  description = "Container image being used"
  value       = var.container_image != "" ? var.container_image : "${aws_ecr_repository.api.repository_url}:latest"
}

# Resource Configuration
output "cpu" {
  description = "CPU units allocated to the task"
  value       = var.cpu
}

output "memory" {
  description = "Memory in MB allocated to the task"
  value       = var.memory
}

output "desired_count" {
  description = "Desired number of tasks"
  value       = var.desired_count
}

# Health Check Configuration
output "health_check_path" {
  description = "Health check path"
  value       = var.health_check_path
}

output "health_check_matcher" {
  description = "Health check response codes"
  value       = var.health_check_matcher
}

# Environment Information
output "environment" {
  description = "Environment name"
  value       = var.environment
}

output "project_name" {
  description = "Project name"
  value       = var.project_name
}

# Deployment Information
output "deployment_info" {
  description = "Complete deployment information"
  value = {
    project_name    = var.project_name
    environment     = var.environment
    cluster_name    = aws_ecs_cluster.main.name
    service_name    = aws_ecs_service.api.name
    task_definition = aws_ecs_task_definition.api.arn
    target_group    = aws_lb_target_group.api.arn
    ecr_repository  = aws_ecr_repository.api.repository_url
    log_group       = aws_cloudwatch_log_group.api.name
    container_port  = var.container_port
    desired_count   = var.desired_count
    cpu             = var.cpu
    memory          = var.memory
  }
  sensitive = false
}

# Database Connection Information (for reference)
output "database_connection_info" {
  description = "Database connection information used by the service"
  value = {
    host     = var.database_host
    port     = var.database_port
    database = var.database_name
    username = var.database_user
  }
  sensitive = true
}

# Container Environment Variables (non-sensitive)
output "container_environment_variables" {
  description = "Environment variables set in the container"
  value = concat([
    {
      name  = "PORT"
      value = tostring(var.container_port)
    },
    {
      name  = "ENVIRONMENT"
      value = var.environment
    },
    {
      name  = "PROJECT_NAME"
      value = var.project_name
    },
    {
      name  = "DATABASE_HOST"
      value = var.database_host
    },
    {
      name  = "DATABASE_PORT"
      value = tostring(var.database_port)
    },
    {
      name  = "DATABASE_NAME"
      value = var.database_name
    },
    {
      name  = "DATABASE_USER"
      value = var.database_user
    }
  ], var.environment_variables)
  sensitive = false
}
