# Outputs for Database Module

output "db_instance_id" {
  description = "RDS instance identifier"
  value       = aws_db_instance.main.identifier
}

output "db_instance_arn" {
  description = "RDS instance ARN"
  value       = aws_db_instance.main.arn
}

output "db_instance_endpoint" {
  description = "RDS instance endpoint"
  value       = aws_db_instance.main.endpoint
}

output "db_instance_address" {
  description = "RDS instance hostname"
  value       = aws_db_instance.main.address
}

output "db_instance_port" {
  description = "RDS instance port"
  value       = aws_db_instance.main.port
}

output "db_name" {
  description = "Database name"
  value       = aws_db_instance.main.db_name
}

output "db_username" {
  description = "Database username"
  value       = aws_db_instance.main.username
  sensitive   = true
}

output "db_password_secret_arn" {
  description = "ARN of the secret containing database password"
  value       = aws_secretsmanager_secret.db_password.arn
}

output "db_password_secret_name" {
  description = "Name of the secret containing database password"
  value       = aws_secretsmanager_secret.db_password.name
}

output "db_subnet_group_name" {
  description = "Database subnet group name"
  value       = aws_db_subnet_group.main.name
}

output "db_parameter_group_name" {
  description = "Database parameter group name"
  value       = aws_db_parameter_group.main.name
}

# Read Replica outputs (conditional)
output "db_read_replica_id" {
  description = "Read replica instance identifier"
  value       = var.create_read_replica ? aws_db_instance.read_replica[0].identifier : null
}

output "db_read_replica_endpoint" {
  description = "Read replica instance endpoint"
  value       = var.create_read_replica ? aws_db_instance.read_replica[0].endpoint : null
}

output "db_read_replica_address" {
  description = "Read replica instance hostname"
  value       = var.create_read_replica ? aws_db_instance.read_replica[0].address : null
}

# Connection information for applications
output "database_url" {
  description = "Database connection URL (without password)"
  value       = "postgresql://${aws_db_instance.main.username}@${aws_db_instance.main.endpoint}/${aws_db_instance.main.db_name}"
  sensitive   = true
}

output "database_connection_info" {
  description = "Database connection information for applications"
  value = {
    host     = aws_db_instance.main.address
    port     = aws_db_instance.main.port
    database = aws_db_instance.main.db_name
    username = aws_db_instance.main.username
    secret_arn = aws_secretsmanager_secret.db_password.arn
  }
  sensitive = true
}

# Read replica connection info (conditional)
output "read_replica_connection_info" {
  description = "Read replica connection information"
  value = var.create_read_replica ? {
    host     = aws_db_instance.read_replica[0].address
    port     = aws_db_instance.read_replica[0].port
    database = aws_db_instance.read_replica[0].db_name
    username = aws_db_instance.read_replica[0].username
    secret_arn = aws_secretsmanager_secret.db_password.arn
  } : null
  sensitive = true
}

# Monitoring and logging
output "cloudwatch_log_group_name" {
  description = "CloudWatch log group name for PostgreSQL logs"
  value       = aws_cloudwatch_log_group.postgresql.name
}

output "enhanced_monitoring_role_arn" {
  description = "Enhanced monitoring IAM role ARN"
  value       = var.monitoring_interval > 0 ? aws_iam_role.rds_enhanced_monitoring[0].arn : null
}

# Backup information
output "backup_retention_period" {
  description = "Backup retention period in days"
  value       = aws_db_instance.main.backup_retention_period
}

output "backup_window" {
  description = "Backup window"
  value       = aws_db_instance.main.backup_window
}

output "maintenance_window" {
  description = "Maintenance window"
  value       = aws_db_instance.main.maintenance_window
}

# Performance insights
output "performance_insights_enabled" {
  description = "Whether Performance Insights is enabled"
  value       = aws_db_instance.main.performance_insights_enabled
}

output "performance_insights_kms_key_id" {
  description = "Performance Insights KMS key ID"
  value       = aws_db_instance.main.performance_insights_kms_key_id
}

# Storage information
output "allocated_storage" {
  description = "Allocated storage in GB"
  value       = aws_db_instance.main.allocated_storage
}

output "max_allocated_storage" {
  description = "Maximum allocated storage in GB"
  value       = aws_db_instance.main.max_allocated_storage
}

output "storage_type" {
  description = "Storage type"
  value       = aws_db_instance.main.storage_type
}

output "storage_encrypted" {
  description = "Whether storage is encrypted"
  value       = aws_db_instance.main.storage_encrypted
}

# Engine information
output "engine" {
  description = "Database engine"
  value       = aws_db_instance.main.engine
}

output "engine_version" {
  description = "Database engine version"
  value       = aws_db_instance.main.engine_version
}

output "instance_class" {
  description = "Database instance class"
  value       = aws_db_instance.main.instance_class
}
