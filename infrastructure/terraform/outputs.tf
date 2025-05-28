# Consolidated outputs for API-Direct Platform infrastructure

# VPC Outputs
output "vpc_id" {
  value       = module.vpc.vpc_id
  description = "The ID of the VPC"
}

output "vpc_cidr" {
  value       = module.vpc.vpc_cidr_block
  description = "The CIDR block of the VPC"
}

output "private_subnet_ids" {
  value       = module.vpc.private_subnets
  description = "List of IDs of private subnets"
}

output "public_subnet_ids" {
  value       = module.vpc.public_subnets
  description = "List of IDs of public subnets"
}

# Essential Service Endpoints
output "api_gateway_endpoint" {
  value       = "http://${aws_lb.main.dns_name}"
  description = "API Gateway endpoint URL (use HTTPS when certificate is validated)"
}

output "cognito_auth_domain" {
  value       = "https://${aws_cognito_user_pool_domain.main.domain}.auth.${var.aws_region}.amazoncognito.com"
  description = "Cognito authentication domain URL"
}

# Configuration for CLI and services
output "platform_config" {
  value = {
    # AWS Resources
    region     = var.aws_region
    account_id = data.aws_caller_identity.current.account_id
    
    # Authentication
    cognito = {
      user_pool_id     = aws_cognito_user_pool.main.id
      cli_client_id    = aws_cognito_user_pool_client.cli.id
      web_client_id    = aws_cognito_user_pool_client.web.id
      auth_domain      = aws_cognito_user_pool_domain.main.domain
      auth_url         = "https://${aws_cognito_user_pool_domain.main.domain}.auth.${var.aws_region}.amazoncognito.com"
    }
    
    # Database
    database = {
      endpoint          = aws_db_instance.main.endpoint
      address           = aws_db_instance.main.address
      port              = aws_db_instance.main.port
      name              = aws_db_instance.main.db_name
      credentials_secret = aws_secretsmanager_secret.db_credentials.arn
    }
    
    # Storage
    storage = {
      code_bucket = aws_s3_bucket.code_storage.id
      ecr_registry = split("/", aws_ecr_repository.user_functions.repository_url)[0]
      ecr_functions_repo = aws_ecr_repository.user_functions.repository_url
    }
    
    # Kubernetes
    eks = {
      cluster_name     = aws_eks_cluster.main.id
      cluster_endpoint = aws_eks_cluster.main.endpoint
      oidc_issuer_url  = aws_eks_cluster.main.identity[0].oidc[0].issuer
    }
    
    # Load Balancer
    alb = {
      dns_name = aws_lb.main.dns_name
      zone_id  = aws_lb.main.zone_id
    }
  }
  description = "Platform configuration for services and CLI"
  sensitive   = false
}

# Kubernetes Configuration
output "eks_kubeconfig_command" {
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${aws_eks_cluster.main.id}"
  description = "Command to update local kubeconfig for EKS cluster access"
}

# Service Discovery
output "service_endpoints" {
  value = {
    api_gateway = "http://${aws_lb.main.dns_name}/api"
    marketplace = "http://${aws_lb.main.dns_name}"
    # Add HTTPS URLs when certificate is validated:
    # api_gateway_https = "https://api.${var.environment == "prod" ? "" : "${var.environment}."}api-direct.io"
    # marketplace_https = "https://${var.environment == "prod" ? "" : "${var.environment}."}api-direct.io"
  }
  description = "Service endpoint URLs"
}
