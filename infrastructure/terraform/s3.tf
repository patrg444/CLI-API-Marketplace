# S3 Buckets for API-Direct Platform

# S3 Bucket for Code Storage
resource "aws_s3_bucket" "code_storage" {
  bucket = "${var.project_name}-${var.environment}-code-storage-${data.aws_caller_identity.current.account_id}"
  
  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-code-storage"
  })
}

# S3 Bucket Versioning
resource "aws_s3_bucket_versioning" "code_storage" {
  bucket = aws_s3_bucket.code_storage.id
  
  versioning_configuration {
    status = "Enabled"
  }
}

# S3 Bucket Encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "code_storage" {
  bucket = aws_s3_bucket.code_storage.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.s3.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

# S3 Bucket Public Access Block
resource "aws_s3_bucket_public_access_block" "code_storage" {
  bucket = aws_s3_bucket.code_storage.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3 Bucket Lifecycle Policy
resource "aws_s3_bucket_lifecycle_configuration" "code_storage" {
  bucket = aws_s3_bucket.code_storage.id

  rule {
    id     = "delete-old-versions"
    status = "Enabled"

    noncurrent_version_expiration {
      noncurrent_days = 90
    }
  }

  rule {
    id     = "transition-old-versions"
    status = "Enabled"

    noncurrent_version_transition {
      noncurrent_days = 30
      storage_class   = "STANDARD_IA"
    }
  }
}

# KMS Key for S3 Encryption
resource "aws_kms_key" "s3" {
  description             = "KMS key for S3 encryption - ${var.project_name}-${var.environment}"
  deletion_window_in_days = 30

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-s3-kms"
  })
}

resource "aws_kms_alias" "s3" {
  name          = "alias/${var.project_name}-${var.environment}-s3"
  target_key_id = aws_kms_key.s3.key_id
}

# S3 Bucket for Terraform State (if not already exists)
resource "aws_s3_bucket" "terraform_state" {
  bucket = "api-direct-terraform-state"
  
  tags = merge(local.common_tags, {
    Name = "api-direct-terraform-state"
    Purpose = "Terraform state storage"
  })

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id
  
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# DynamoDB Table for Terraform State Locking
resource "aws_dynamodb_table" "terraform_locks" {
  name         = "api-direct-terraform-locks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags = merge(local.common_tags, {
    Name = "api-direct-terraform-locks"
    Purpose = "Terraform state locking"
  })
}

# Data source for current AWS account ID
data "aws_caller_identity" "current" {}

# Outputs
output "code_storage_bucket_name" {
  value       = aws_s3_bucket.code_storage.id
  description = "The name of the S3 bucket for code storage"
}

output "code_storage_bucket_arn" {
  value       = aws_s3_bucket.code_storage.arn
  description = "The ARN of the S3 bucket for code storage"
}

output "terraform_state_bucket_name" {
  value       = aws_s3_bucket.terraform_state.id
  description = "The name of the S3 bucket for Terraform state"
}
