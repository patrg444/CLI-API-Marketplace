# API Fargate Module for API-Direct User Deployments
# Creates ECS Fargate service for running API containers

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# Data sources
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# ECR Repository for the API container
resource "aws_ecr_repository" "api" {
  name                 = var.project_name
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  tags = {
    Name        = "${var.project_name}-ecr"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# ECR Repository Policy
resource "aws_ecr_repository_policy" "api" {
  repository = aws_ecr_repository.api.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowPushPull"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action = [
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:BatchCheckLayerAvailability",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload"
        ]
      }
    ]
  })
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "${var.project_name}-cluster"

  configuration {
    execute_command_configuration {
      logging = "OVERRIDE"
      log_configuration {
        cloud_watch_log_group_name = aws_cloudwatch_log_group.ecs_exec.name
      }
    }
  }

  setting {
    name  = "containerInsights"
    value = var.enable_container_insights ? "enabled" : "disabled"
  }

  tags = {
    Name        = "${var.project_name}-cluster"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "api" {
  name              = "/ecs/${var.project_name}"
  retention_in_days = var.log_retention_days

  tags = {
    Name        = "${var.project_name}-logs"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

resource "aws_cloudwatch_log_group" "ecs_exec" {
  name              = "/ecs/${var.project_name}/exec"
  retention_in_days = var.log_retention_days

  tags = {
    Name        = "${var.project_name}-exec-logs"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# ECS Task Execution Role
resource "aws_iam_role" "ecs_task_execution" {
  name = "${var.project_name}-ecs-task-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-ecs-task-execution"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# ECS Task Execution Role Policy Attachments
resource "aws_iam_role_policy_attachment" "ecs_task_execution" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# Additional policy for Secrets Manager access
resource "aws_iam_role_policy" "ecs_task_execution_secrets" {
  name = "${var.project_name}-ecs-secrets"
  role = aws_iam_role.ecs_task_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = [
          var.db_password_secret_arn,
          "arn:aws:secretsmanager:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:secret:${var.project_name}/*"
        ]
      }
    ]
  })
}

# ECS Task Role (for the application)
resource "aws_iam_role" "ecs_task" {
  name = "${var.project_name}-ecs-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name        = "${var.project_name}-ecs-task"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# ECS Task Role Policy for application permissions
resource "aws_iam_role_policy" "ecs_task_policy" {
  name = "${var.project_name}-ecs-task-policy"
  role = aws_iam_role.ecs_task.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = [
          var.db_password_secret_arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = [
          "${aws_cloudwatch_log_group.api.arn}:*"
        ]
      }
    ]
  })
}

# ECS Task Definition
resource "aws_ecs_task_definition" "api" {
  family                   = var.project_name
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn           = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name  = "api"
      image = var.container_image != "" ? var.container_image : "${aws_ecr_repository.api.repository_url}:latest"
      
      portMappings = [
        {
          containerPort = var.container_port
          protocol      = "tcp"
        }
      ]

      environment = concat([
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

      secrets = [
        {
          name      = "DATABASE_PASSWORD"
          valueFrom = var.db_password_secret_arn
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.api.name
          "awslogs-region"        = data.aws_region.current.name
          "awslogs-stream-prefix" = "ecs"
        }
      }

      healthCheck = {
        command = [
          "CMD-SHELL",
          var.health_check_command
        ]
        interval    = var.health_check_interval
        timeout     = var.health_check_timeout
        retries     = var.health_check_retries
        startPeriod = var.health_check_start_period
      }

      essential = true
    }
  ])

  tags = {
    Name        = "${var.project_name}-task"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# Application Load Balancer Target Group
resource "aws_lb_target_group" "api" {
  name        = "${var.project_name}-tg"
  port        = var.container_port
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = var.health_check_healthy_threshold
    interval            = var.health_check_interval
    matcher             = var.health_check_matcher
    path                = var.health_check_path
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = var.health_check_timeout
    unhealthy_threshold = var.health_check_unhealthy_threshold
  }

  tags = {
    Name        = "${var.project_name}-tg"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# ALB Listener Rule
resource "aws_lb_listener_rule" "api" {
  listener_arn = var.alb_listener_arn
  priority     = var.listener_rule_priority

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api.arn
  }

  condition {
    host_header {
      values = var.host_headers
    }
  }

  tags = {
    Name        = "${var.project_name}-listener-rule"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# ECS Service
resource "aws_ecs_service" "api" {
  name            = var.project_name
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    security_groups  = [var.ecs_security_group_id]
    subnets          = var.private_subnet_ids
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "api"
    container_port   = var.container_port
  }

  # Auto Scaling
  dynamic "service_registries" {
    for_each = var.enable_service_discovery ? [1] : []
    content {
      registry_arn = aws_service_discovery_service.api[0].arn
    }
  }

  depends_on = [aws_lb_target_group.api]

  tags = {
    Name        = "${var.project_name}-service"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

# Auto Scaling Target
resource "aws_appautoscaling_target" "ecs_target" {
  count = var.enable_auto_scaling ? 1 : 0

  max_capacity       = var.max_capacity
  min_capacity       = var.min_capacity
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.api.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

# Auto Scaling Policy - CPU
resource "aws_appautoscaling_policy" "ecs_cpu_policy" {
  count = var.enable_auto_scaling ? 1 : 0

  name               = "${var.project_name}-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target[0].resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target[0].service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value = var.cpu_target_value
  }
}

# Auto Scaling Policy - Memory
resource "aws_appautoscaling_policy" "ecs_memory_policy" {
  count = var.enable_auto_scaling ? 1 : 0

  name               = "${var.project_name}-memory-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target[0].resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target[0].service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
    target_value = var.memory_target_value
  }
}

# Service Discovery (optional)
resource "aws_service_discovery_private_dns_namespace" "main" {
  count = var.enable_service_discovery ? 1 : 0

  name        = "${var.project_name}.local"
  description = "Service discovery namespace for ${var.project_name}"
  vpc         = var.vpc_id

  tags = {
    Name        = "${var.project_name}-namespace"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}

resource "aws_service_discovery_service" "api" {
  count = var.enable_service_discovery ? 1 : 0

  name = "api"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.main[0].id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_grace_period_seconds = 30

  tags = {
    Name        = "${var.project_name}-service-discovery"
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "api-direct"
  }
}
