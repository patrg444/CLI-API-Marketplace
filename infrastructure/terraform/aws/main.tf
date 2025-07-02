terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_caller_identity" "current" {}

# VPC Module
module "vpc" {
  source = "./modules/vpc"
  
  project_name = var.project_name
  environment  = var.environment
  vpc_cidr     = var.vpc_cidr
  
  availability_zones = data.aws_availability_zones.available.names
  public_subnets     = var.public_subnets
  private_subnets    = var.private_subnets
  database_subnets   = var.database_subnets
}

# Security Groups
module "security_groups" {
  source = "./modules/security"
  
  project_name = var.project_name
  environment  = var.environment
  vpc_id       = module.vpc.vpc_id
  vpc_cidr     = var.vpc_cidr
}

# RDS PostgreSQL
module "rds" {
  source = "./modules/rds"
  
  project_name = var.project_name
  environment  = var.environment
  
  subnet_ids         = module.vpc.database_subnet_ids
  security_group_id  = module.security_groups.rds_security_group_id
  
  db_name     = var.db_name
  db_username = var.db_username
  db_password = var.db_password
  
  instance_class    = var.rds_instance_class
  allocated_storage = var.rds_allocated_storage
  
  backup_retention_period = var.environment == "production" ? 7 : 1
  multi_az               = var.environment == "production"
}

# ElastiCache Redis
module "redis" {
  source = "./modules/elasticache"
  
  project_name = var.project_name
  environment  = var.environment
  
  subnet_ids        = module.vpc.private_subnet_ids
  security_group_id = module.security_groups.redis_security_group_id
  
  node_type = var.redis_node_type
  auth_token = var.redis_auth_token
}

# S3 Buckets
module "s3" {
  source = "./modules/s3"
  
  project_name = var.project_name
  environment  = var.environment
  account_id   = data.aws_caller_identity.current.account_id
}

# Cognito User Pool
module "cognito" {
  source = "./modules/cognito"
  
  project_name = var.project_name
  environment  = var.environment
  
  email_from_address = var.email_from_address
  domain_name        = var.domain_name
}

# ECS Cluster
module "ecs" {
  source = "./modules/ecs"
  
  project_name = var.project_name
  environment  = var.environment
}

# Application Load Balancer
module "alb" {
  source = "./modules/alb"
  
  project_name = var.project_name
  environment  = var.environment
  
  vpc_id          = module.vpc.vpc_id
  public_subnets  = module.vpc.public_subnet_ids
  security_group_id = module.security_groups.alb_security_group_id
  
  certificate_arn = var.certificate_arn
  domain_name     = var.domain_name
}

# API Gateway
module "api_gateway" {
  source = "./modules/api_gateway"
  
  project_name = var.project_name
  environment  = var.environment
  
  cognito_user_pool_arn = module.cognito.user_pool_arn
  alb_dns_name          = module.alb.alb_dns_name
}

# ECR Repositories
module "ecr" {
  source = "./modules/ecr"
  
  project_name = var.project_name
  environment  = var.environment
  
  services = [
    "marketplace",
    "apikey",
    "billing",
    "gateway",
    "metering",
    "deployment",
    "payout"
  ]
}

# IAM Roles
module "iam" {
  source = "./modules/iam"
  
  project_name = var.project_name
  environment  = var.environment
  account_id   = data.aws_caller_identity.current.account_id
  
  s3_bucket_arns = [
    module.s3.assets_bucket_arn,
    module.s3.backups_bucket_arn
  ]
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "ecs_logs" {
  for_each = toset([
    "marketplace",
    "apikey",
    "billing",
    "gateway",
    "metering",
    "deployment",
    "payout"
  ])
  
  name              = "/ecs/${var.project_name}-${var.environment}/${each.key}"
  retention_in_days = var.environment == "production" ? 30 : 7
}

# Outputs
output "vpc_id" {
  value = module.vpc.vpc_id
}

output "alb_dns_name" {
  value = module.alb.alb_dns_name
}

output "cognito_user_pool_id" {
  value = module.cognito.user_pool_id
}

output "cognito_client_id" {
  value = module.cognito.client_id
}

output "rds_endpoint" {
  value = module.rds.endpoint
}

output "redis_endpoint" {
  value = module.redis.endpoint
}

output "api_gateway_url" {
  value = module.api_gateway.api_url
}

output "s3_assets_bucket" {
  value = module.s3.assets_bucket_name
}

output "ecr_repositories" {
  value = module.ecr.repository_urls
}