# Application Load Balancer for API-Direct Platform

# ALB Security Group
resource "aws_security_group" "alb" {
  name_prefix = "${var.project_name}-${var.environment}-alb"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP from anywhere"
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS from anywhere"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound traffic"
  }

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-alb-sg"
  })
}

# Application Load Balancer
resource "aws_lb" "main" {
  name               = "${var.project_name}-${var.environment}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = module.vpc.public_subnets

  enable_deletion_protection = var.environment == "prod"
  enable_http2              = true
  enable_cross_zone_load_balancing = true

  access_logs {
    bucket  = aws_s3_bucket.alb_logs.bucket
    prefix  = "alb"
    enabled = true
  }

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-alb"
  })
}

# S3 Bucket for ALB Logs
resource "aws_s3_bucket" "alb_logs" {
  bucket = "${var.project_name}-${var.environment}-alb-logs-${data.aws_caller_identity.current.account_id}"
  
  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-alb-logs"
  })
}

# S3 Bucket Policy for ALB Logs
resource "aws_s3_bucket_policy" "alb_logs" {
  bucket = aws_s3_bucket.alb_logs.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_elb_service_account.main.id}:root"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.alb_logs.arn}/*"
      }
    ]
  })
}

# Data source for ELB service account
data "aws_elb_service_account" "main" {}

# S3 Bucket Lifecycle for ALB Logs
resource "aws_s3_bucket_lifecycle_configuration" "alb_logs" {
  bucket = aws_s3_bucket.alb_logs.id

  rule {
    id     = "expire-old-logs"
    status = "Enabled"

    expiration {
      days = 30
    }
  }
}

# S3 Bucket Public Access Block for ALB Logs
resource "aws_s3_bucket_public_access_block" "alb_logs" {
  bucket = aws_s3_bucket.alb_logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ALB Target Group for API Gateway Service
resource "aws_lb_target_group" "api_gateway" {
  name                 = "${var.project_name}-${var.environment}-api-gateway"
  port                 = 8080
  protocol             = "HTTP"
  vpc_id               = module.vpc.vpc_id
  target_type          = "ip"
  deregistration_delay = 30

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  stickiness {
    type            = "lb_cookie"
    cookie_duration = 86400
    enabled         = false
  }

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-api-gateway-tg"
  })
}

# ALB Target Group for Marketplace Frontend
resource "aws_lb_target_group" "marketplace_frontend" {
  name                 = "${var.project_name}-${var.environment}-marketplace"
  port                 = 3000
  protocol             = "HTTP"
  vpc_id               = module.vpc.vpc_id
  target_type          = "ip"
  deregistration_delay = 30

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-marketplace-tg"
  })
}

# ALB Listener (HTTP)
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

# ACM Certificate (placeholder - requires domain validation)
resource "aws_acm_certificate" "main" {
  domain_name       = var.environment == "prod" ? "api-direct.io" : "${var.environment}.api-direct.io"
  validation_method = "DNS"

  subject_alternative_names = [
    var.environment == "prod" ? "*.api-direct.io" : "*.${var.environment}.api-direct.io",
    var.environment == "prod" ? "api.api-direct.io" : "api.${var.environment}.api-direct.io",
    var.environment == "prod" ? "marketplace.api-direct.io" : "marketplace.${var.environment}.api-direct.io"
  ]

  lifecycle {
    create_before_destroy = true
  }

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-${var.environment}-cert"
  })
}

# ALB Listener (HTTPS) - Commented out until certificate is validated
# resource "aws_lb_listener" "https" {
#   load_balancer_arn = aws_lb.main.arn
#   port              = "443"
#   protocol          = "HTTPS"
#   ssl_policy        = "ELBSecurityPolicy-TLS-1-2-2017-01"
#   certificate_arn   = aws_acm_certificate.main.arn
#
#   default_action {
#     type             = "forward"
#     target_group_arn = aws_lb_target_group.api_gateway.arn
#   }
# }

# ALB Listener Rules for routing
# resource "aws_lb_listener_rule" "api" {
#   listener_arn = aws_lb_listener.https.arn
#   priority     = 100
#
#   action {
#     type             = "forward"
#     target_group_arn = aws_lb_target_group.api_gateway.arn
#   }
#
#   condition {
#     host_header {
#       values = [
#         var.environment == "prod" ? "api.api-direct.io" : "api.${var.environment}.api-direct.io"
#       ]
#     }
#   }
# }

# resource "aws_lb_listener_rule" "marketplace" {
#   listener_arn = aws_lb_listener.https.arn
#   priority     = 200
#
#   action {
#     type             = "forward"
#     target_group_arn = aws_lb_target_group.marketplace_frontend.arn
#   }
#
#   condition {
#     host_header {
#       values = [
#         var.environment == "prod" ? "marketplace.api-direct.io" : "marketplace.${var.environment}.api-direct.io",
#         var.environment == "prod" ? "api-direct.io" : "${var.environment}.api-direct.io"
#       ]
#     }
#   }
# }

# Security Group Rule to allow ALB to communicate with EKS nodes
resource "aws_security_group_rule" "eks_nodes_from_alb" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 65535
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.alb.id
  security_group_id        = aws_security_group.eks_nodes.id
  description              = "Allow ALB to communicate with EKS nodes"
}

# Outputs
output "alb_dns_name" {
  value       = aws_lb.main.dns_name
  description = "The DNS name of the load balancer"
}

output "alb_zone_id" {
  value       = aws_lb.main.zone_id
  description = "The zone ID of the load balancer"
}

output "alb_arn" {
  value       = aws_lb.main.arn
  description = "The ARN of the load balancer"
}

output "api_gateway_target_group_arn" {
  value       = aws_lb_target_group.api_gateway.arn
  description = "The ARN of the API Gateway target group"
}

output "marketplace_target_group_arn" {
  value       = aws_lb_target_group.marketplace_frontend.arn
  description = "The ARN of the marketplace frontend target group"
}
