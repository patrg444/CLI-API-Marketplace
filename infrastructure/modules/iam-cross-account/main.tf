# IAM Cross-Account Role Module for API-Direct BYOA
# Creates IAM roles and policies for secure cross-account access

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
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# Generate a unique external ID for this deployment
resource "random_uuid" "external_id" {}

# Store the external ID in AWS Secrets Manager for API-Direct to retrieve
resource "aws_secretsmanager_secret" "external_id" {
  name                    = "${var.project_name}-api-direct-external-id"
  description             = "External ID for API-Direct cross-account access"
  recovery_window_in_days = var.environment == "prod" ? 30 : 0

  tags = {
    Name        = "${var.project_name}-external-id"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
    Purpose     = "cross-account-access"
  }
}

resource "aws_secretsmanager_secret_version" "external_id" {
  secret_id = aws_secretsmanager_secret.external_id.id
  secret_string = jsonencode({
    external_id    = random_uuid.external_id.result
    account_id     = data.aws_caller_identity.current.account_id
    region         = data.aws_region.current.name
    project_name   = var.project_name
    created_at     = timestamp()
    api_direct_account = var.api_direct_account_id
  })
}

# IAM Role for API-Direct to assume
resource "aws_iam_role" "api_direct_deployment" {
  name = "${var.project_name}-api-direct-deployment-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${var.api_direct_account_id}:root"
        }
        Action = "sts:AssumeRole"
        Condition = {
          StringEquals = {
            "sts:ExternalId" = random_uuid.external_id.result
          }
          StringLike = {
            "aws:userid" = var.allowed_user_patterns
          }
        }
      }
    ]
  })

  max_session_duration = var.max_session_duration

  tags = {
    Name        = "${var.project_name}-api-direct-deployment-role"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
    Purpose     = "cross-account-deployment"
  }
}

# Policy for ECS and ECR management
resource "aws_iam_policy" "ecs_management" {
  name        = "${var.project_name}-ecs-management"
  description = "Policy for managing ECS resources for ${var.project_name}"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ecs:CreateCluster",
          "ecs:DeleteCluster",
          "ecs:DescribeClusters",
          "ecs:UpdateCluster",
          "ecs:CreateService",
          "ecs:DeleteService",
          "ecs:DescribeServices",
          "ecs:UpdateService",
          "ecs:RegisterTaskDefinition",
          "ecs:DeregisterTaskDefinition",
          "ecs:DescribeTaskDefinition",
          "ecs:ListTaskDefinitions",
          "ecs:RunTask",
          "ecs:StopTask",
          "ecs:DescribeTasks",
          "ecs:ListTasks",
          "ecs:TagResource",
          "ecs:UntagResource",
          "ecs:ListTagsForResource"
        ]
        Resource = [
          "arn:aws:ecs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:cluster/${var.project_name}*",
          "arn:aws:ecs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:service/${var.project_name}*",
          "arn:aws:ecs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:task-definition/${var.project_name}*",
          "arn:aws:ecs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:task/${var.project_name}*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:CreateRepository",
          "ecr:DeleteRepository",
          "ecr:DescribeRepositories",
          "ecr:GetRepositoryPolicy",
          "ecr:SetRepositoryPolicy",
          "ecr:DeleteRepositoryPolicy",
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:TagResource",
          "ecr:UntagResource",
          "ecr:ListTagsForResource"
        ]
        Resource = [
          "arn:aws:ecr:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:repository/${var.project_name}*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken"
        ]
        Resource = "*"
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-ecs-management"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# Policy for Load Balancer management
resource "aws_iam_policy" "alb_management" {
  name        = "${var.project_name}-alb-management"
  description = "Policy for managing ALB resources for ${var.project_name}"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "elasticloadbalancing:CreateTargetGroup",
          "elasticloadbalancing:DeleteTargetGroup",
          "elasticloadbalancing:DescribeTargetGroups",
          "elasticloadbalancing:ModifyTargetGroup",
          "elasticloadbalancing:CreateRule",
          "elasticloadbalancing:DeleteRule",
          "elasticloadbalancing:DescribeRules",
          "elasticloadbalancing:ModifyRule",
          "elasticloadbalancing:RegisterTargets",
          "elasticloadbalancing:DeregisterTargets",
          "elasticloadbalancing:DescribeTargetHealth",
          "elasticloadbalancing:AddTags",
          "elasticloadbalancing:RemoveTags",
          "elasticloadbalancing:DescribeTags"
        ]
        Resource = [
          "arn:aws:elasticloadbalancing:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:targetgroup/${var.project_name}*",
          "arn:aws:elasticloadbalancing:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:listener-rule/app/${var.project_name}*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "elasticloadbalancing:DescribeLoadBalancers",
          "elasticloadbalancing:DescribeListeners"
        ]
        Resource = "*"
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-alb-management"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# Policy for IAM role management (for ECS tasks)
resource "aws_iam_policy" "iam_management" {
  name        = "${var.project_name}-iam-management"
  description = "Policy for managing IAM resources for ${var.project_name}"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "iam:CreateRole",
          "iam:DeleteRole",
          "iam:GetRole",
          "iam:UpdateRole",
          "iam:AttachRolePolicy",
          "iam:DetachRolePolicy",
          "iam:CreatePolicy",
          "iam:DeletePolicy",
          "iam:GetPolicy",
          "iam:GetPolicyVersion",
          "iam:ListPolicyVersions",
          "iam:TagRole",
          "iam:UntagRole",
          "iam:TagPolicy",
          "iam:UntagPolicy",
          "iam:PassRole"
        ]
        Resource = [
          "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.project_name}*",
          "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/${var.project_name}*"
        ]
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-iam-management"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# Policy for CloudWatch Logs management
resource "aws_iam_policy" "logs_management" {
  name        = "${var.project_name}-logs-management"
  description = "Policy for managing CloudWatch Logs for ${var.project_name}"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:DeleteLogGroup",
          "logs:DescribeLogGroups",
          "logs:PutRetentionPolicy",
          "logs:DeleteRetentionPolicy",
          "logs:TagLogGroup",
          "logs:UntagLogGroup",
          "logs:ListTagsLogGroup"
        ]
        Resource = [
          "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/ecs/${var.project_name}*"
        ]
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-logs-management"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# Policy for Auto Scaling management
resource "aws_iam_policy" "autoscaling_management" {
  name        = "${var.project_name}-autoscaling-management"
  description = "Policy for managing Auto Scaling for ${var.project_name}"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "application-autoscaling:RegisterScalableTarget",
          "application-autoscaling:DeregisterScalableTarget",
          "application-autoscaling:DescribeScalableTargets",
          "application-autoscaling:PutScalingPolicy",
          "application-autoscaling:DeleteScalingPolicy",
          "application-autoscaling:DescribeScalingPolicies",
          "application-autoscaling:TagResource",
          "application-autoscaling:UntagResource",
          "application-autoscaling:ListTagsForResource"
        ]
        Resource = [
          "arn:aws:application-autoscaling:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:scalable-target/ecs/service/${var.project_name}*"
        ]
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-autoscaling-management"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# Policy for Service Discovery management (optional)
resource "aws_iam_policy" "service_discovery_management" {
  count = var.enable_service_discovery ? 1 : 0

  name        = "${var.project_name}-service-discovery-management"
  description = "Policy for managing Service Discovery for ${var.project_name}"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "servicediscovery:CreatePrivateDnsNamespace",
          "servicediscovery:DeleteNamespace",
          "servicediscovery:GetNamespace",
          "servicediscovery:CreateService",
          "servicediscovery:DeleteService",
          "servicediscovery:GetService",
          "servicediscovery:TagResource",
          "servicediscovery:UntagResource",
          "servicediscovery:ListTagsForResource"
        ]
        Resource = [
          "arn:aws:servicediscovery:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:namespace/*",
          "arn:aws:servicediscovery:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:service/*"
        ]
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-service-discovery-management"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# Policy for Secrets Manager access (read-only for database credentials)
resource "aws_iam_policy" "secrets_access" {
  name        = "${var.project_name}-secrets-access"
  description = "Policy for accessing secrets for ${var.project_name}"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = [
          "arn:aws:secretsmanager:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:secret:${var.project_name}*"
        ]
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-secrets-access"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# Attach policies to the role
resource "aws_iam_role_policy_attachment" "ecs_management" {
  role       = aws_iam_role.api_direct_deployment.name
  policy_arn = aws_iam_policy.ecs_management.arn
}

resource "aws_iam_role_policy_attachment" "alb_management" {
  role       = aws_iam_role.api_direct_deployment.name
  policy_arn = aws_iam_policy.alb_management.arn
}

resource "aws_iam_role_policy_attachment" "iam_management" {
  role       = aws_iam_role.api_direct_deployment.name
  policy_arn = aws_iam_policy.iam_management.arn
}

resource "aws_iam_role_policy_attachment" "logs_management" {
  role       = aws_iam_role.api_direct_deployment.name
  policy_arn = aws_iam_policy.logs_management.arn
}

resource "aws_iam_role_policy_attachment" "autoscaling_management" {
  role       = aws_iam_role.api_direct_deployment.name
  policy_arn = aws_iam_policy.autoscaling_management.arn
}

resource "aws_iam_role_policy_attachment" "service_discovery_management" {
  count = var.enable_service_discovery ? 1 : 0

  role       = aws_iam_role.api_direct_deployment.name
  policy_arn = aws_iam_policy.service_discovery_management[0].arn
}

resource "aws_iam_role_policy_attachment" "secrets_access" {
  role       = aws_iam_role.api_direct_deployment.name
  policy_arn = aws_iam_policy.secrets_access.arn
}

# Optional: CloudFormation stack management for advanced deployments
resource "aws_iam_policy" "cloudformation_management" {
  count = var.enable_cloudformation_access ? 1 : 0

  name        = "${var.project_name}-cloudformation-management"
  description = "Policy for managing CloudFormation stacks for ${var.project_name}"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "cloudformation:CreateStack",
          "cloudformation:UpdateStack",
          "cloudformation:DeleteStack",
          "cloudformation:DescribeStacks",
          "cloudformation:DescribeStackEvents",
          "cloudformation:DescribeStackResources",
          "cloudformation:GetTemplate",
          "cloudformation:ValidateTemplate",
          "cloudformation:TagResource",
          "cloudformation:UntagResource",
          "cloudformation:ListTagsForResource"
        ]
        Resource = [
          "arn:aws:cloudformation:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:stack/${var.project_name}*"
        ]
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-cloudformation-management"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

resource "aws_iam_role_policy_attachment" "cloudformation_management" {
  count = var.enable_cloudformation_access ? 1 : 0

  role       = aws_iam_role.api_direct_deployment.name
  policy_arn = aws_iam_policy.cloudformation_management[0].arn
}
