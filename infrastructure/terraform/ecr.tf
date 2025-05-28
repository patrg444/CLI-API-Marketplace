# ECR (Elastic Container Registry) for API-Direct Platform

# ECR Repository for base runtime images
resource "aws_ecr_repository" "runtime_images" {
  for_each = toset(["python", "nodejs", "go"])
  
  name                 = "${var.project_name}/${var.environment}/runtime/${each.key}"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "KMS"
    kms_key        = aws_kms_key.ecr.arn
  }

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-runtime-${each.key}"
    Type = "runtime"
  })
}

# ECR Repository for user function images
resource "aws_ecr_repository" "user_functions" {
  name                 = "${var.project_name}/${var.environment}/functions"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "KMS"
    kms_key        = aws_kms_key.ecr.arn
  }

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-functions"
    Type = "user-functions"
  })
}

# ECR Lifecycle Policy for runtime images
resource "aws_ecr_lifecycle_policy" "runtime_images" {
  for_each   = aws_ecr_repository.runtime_images
  repository = each.value.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# ECR Lifecycle Policy for user function images
resource "aws_ecr_lifecycle_policy" "user_functions" {
  repository = aws_ecr_repository.user_functions.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Expire untagged images after 7 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 7
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 2
        description  = "Keep last 50 tagged images per API"
        selection = {
          tagStatus   = "tagged"
          tagPrefixList = ["api-"]
          countType   = "imageCountMoreThan"
          countNumber = 50
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# KMS Key for ECR Encryption
resource "aws_kms_key" "ecr" {
  description             = "KMS key for ECR encryption - ${var.project_name}-${var.environment}"
  deletion_window_in_days = 30

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-ecr-kms"
  })
}

resource "aws_kms_alias" "ecr" {
  name          = "alias/${var.project_name}-${var.environment}-ecr"
  target_key_id = aws_kms_key.ecr.key_id
}

# IAM Policy for ECR Access (to be attached to EKS nodes and build systems)
resource "aws_iam_policy" "ecr_access" {
  name        = "${var.project_name}-${var.environment}-ecr-access"
  description = "Policy for accessing ECR repositories"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload"
        ]
        Resource = "*"
      }
    ]
  })

  tags = local.common_tags
}

# Outputs
output "ecr_runtime_repository_urls" {
  value = {
    for k, v in aws_ecr_repository.runtime_images : k => v.repository_url
  }
  description = "URLs of the ECR repositories for runtime images"
}

output "ecr_functions_repository_url" {
  value       = aws_ecr_repository.user_functions.repository_url
  description = "URL of the ECR repository for user function images"
}

output "ecr_registry_url" {
  value       = split("/", aws_ecr_repository.user_functions.repository_url)[0]
  description = "ECR registry URL"
}
