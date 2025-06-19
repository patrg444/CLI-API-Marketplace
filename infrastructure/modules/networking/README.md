# Networking Module

This Terraform module creates the foundational networking infrastructure for API-Direct user deployments, including VPC, subnets, load balancer, and security groups.

## Features

- **VPC with public and private subnets** across multiple availability zones
- **Application Load Balancer** with HTTP/HTTPS listeners
- **NAT Gateways** for private subnet internet access
- **Security Groups** for ALB, ECS tasks, and RDS
- **Configurable SSL certificate** support
- **Comprehensive tagging** for resource management

## Usage

```hcl
module "networking" {
  source = "./modules/networking"

  project_name = "my-api"
  environment  = "prod"
  vpc_cidr     = "10.0.0.0/16"
  az_count     = 2

  # Optional: SSL certificate for HTTPS
  ssl_certificate_arn = "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"

  # Optional: Enable deletion protection for load balancer
  enable_deletion_protection = false

  tags = {
    Owner = "api-direct"
    Cost  = "user-deployment"
  }
}
```

## Architecture

```
Internet Gateway
       |
   Public Subnets (ALB)
       |
   Private Subnets (ECS Tasks, RDS)
       |
   NAT Gateways
```

### Network Layout

- **Public Subnets**: Host the Application Load Balancer
- **Private Subnets**: Host ECS tasks and RDS database
- **NAT Gateways**: Provide internet access for private subnets
- **Security Groups**: Control traffic between components

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| project_name | Name of the project (used for resource naming) | `string` | n/a | yes |
| environment | Environment name (dev, staging, prod) | `string` | `"prod"` | no |
| vpc_cidr | CIDR block for the VPC | `string` | `"10.0.0.0/16"` | no |
| az_count | Number of Availability Zones to use | `number` | `2` | no |
| enable_deletion_protection | Enable deletion protection for the load balancer | `bool` | `false` | no |
| ssl_certificate_arn | ARN of the SSL certificate for HTTPS listener | `string` | `""` | no |
| tags | Additional tags to apply to all resources | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| vpc_id | ID of the VPC |
| vpc_cidr_block | CIDR block of the VPC |
| public_subnet_ids | IDs of the public subnets |
| private_subnet_ids | IDs of the private subnets |
| alb_security_group_id | ID of the Application Load Balancer security group |
| ecs_tasks_security_group_id | ID of the ECS tasks security group |
| rds_security_group_id | ID of the RDS security group |
| alb_arn | ARN of the Application Load Balancer |
| alb_dns_name | DNS name of the Application Load Balancer |
| alb_listener_http_arn | ARN of the HTTP listener |
| alb_listener_https_arn | ARN of the HTTPS listener (if SSL certificate provided) |
| api_url | URL for the API (HTTP or HTTPS based on SSL certificate) |

## Security Groups

### ALB Security Group
- **Inbound**: HTTP (80) and HTTPS (443) from anywhere
- **Outbound**: All traffic

### ECS Tasks Security Group
- **Inbound**: HTTP (80-65535) from ALB security group
- **Outbound**: All traffic

### RDS Security Group
- **Inbound**: PostgreSQL (5432) from ECS tasks security group
- **Outbound**: None

## Cost Considerations

This module creates the following billable resources:
- **NAT Gateways**: ~$45/month per gateway
- **Elastic IPs**: ~$3.65/month per unused IP
- **Application Load Balancer**: ~$16/month + data processing charges

For cost optimization:
- Use fewer availability zones (minimum 2 for high availability)
- Consider using NAT instances instead of NAT gateways for dev environments

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| aws | ~> 5.0 |

## Providers

| Name | Version |
|------|---------|
| aws | ~> 5.0 |

## Resources Created

- 1 VPC
- 2+ Public subnets (based on az_count)
- 2+ Private subnets (based on az_count)
- 1 Internet Gateway
- 2+ NAT Gateways (based on az_count)
- 2+ Elastic IPs (based on az_count)
- Route tables and associations
- 3 Security groups (ALB, ECS, RDS)
- 1 Application Load Balancer
- 1 Target group
- 1-2 Load balancer listeners (HTTP, optionally HTTPS)

## Example: Complete Setup

```hcl
module "networking" {
  source = "./modules/networking"

  project_name = "weather-api"
  environment  = "prod"
  vpc_cidr     = "10.0.0.0/16"
  az_count     = 3

  ssl_certificate_arn        = aws_acm_certificate.api.arn
  enable_deletion_protection = true

  tags = {
    Project     = "weather-api"
    Environment = "prod"
    ManagedBy   = "api-direct"
    Owner       = "john@example.com"
  }
}

# Use outputs in other modules
module "database" {
  source = "./modules/database"

  vpc_id                = module.networking.vpc_id
  private_subnet_ids    = module.networking.private_subnet_ids
  rds_security_group_id = module.networking.rds_security_group_id
  
  # ... other variables
}
