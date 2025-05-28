# AWS Cognito User Pool for API-Direct Platform
resource "aws_cognito_user_pool" "main" {
  name = "${var.project_name}-${var.environment}-user-pool"

  # Username configuration
  username_attributes      = ["email"]
  auto_verified_attributes = ["email"]
  
  # Password policy
  password_policy {
    minimum_length                   = 12
    require_lowercase                = true
    require_uppercase                = true
    require_numbers                  = true
    require_symbols                  = true
    temporary_password_validity_days = 7
  }

  # Email configuration
  email_configuration {
    email_sending_account = "COGNITO_DEFAULT"
  }

  # Account recovery
  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  # User attribute schema
  schema {
    name                     = "email"
    attribute_data_type      = "String"
    required                 = true
    mutable                  = true
    developer_only_attribute = false
    
    string_attribute_constraints {
      min_length = 0
      max_length = 2048
    }
  }

  schema {
    name                     = "name"
    attribute_data_type      = "String"
    required                 = false
    mutable                  = true
    developer_only_attribute = false
    
    string_attribute_constraints {
      min_length = 0
      max_length = 2048
    }
  }

  schema {
    name                     = "organization"
    attribute_data_type      = "String"
    required                 = false
    mutable                  = true
    developer_only_attribute = false
    
    string_attribute_constraints {
      min_length = 0
      max_length = 256
    }
  }

  # MFA configuration
  mfa_configuration = "OPTIONAL"
  
  software_token_mfa_configuration {
    enabled = true
  }

  # Device tracking
  device_configuration {
    challenge_required_on_new_device      = true
    device_only_remembered_on_user_prompt = true
  }

  # Lambda triggers (placeholders for future enhancement)
  # lambda_config {
  #   post_confirmation = aws_lambda_function.post_confirmation.arn
  # }

  tags = local.common_tags
}

# Cognito User Pool Domain
resource "aws_cognito_user_pool_domain" "main" {
  domain       = "${var.project_name}-${var.environment}-auth"
  user_pool_id = aws_cognito_user_pool.main.id
}

# CLI App Client
resource "aws_cognito_user_pool_client" "cli" {
  name                                 = "${var.project_name}-cli"
  user_pool_id                         = aws_cognito_user_pool.main.id
  
  # OAuth configuration for CLI
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_scopes                 = ["email", "openid", "profile"]
  callback_urls                        = ["http://localhost:8080/callback"]
  logout_urls                          = ["http://localhost:8080/logout"]
  supported_identity_providers         = ["COGNITO"]
  
  # Token configuration
  access_token_validity  = 60    # 60 minutes
  id_token_validity      = 60    # 60 minutes
  refresh_token_validity = 30    # 30 days
  
  token_validity_units {
    access_token  = "minutes"
    id_token      = "minutes"
    refresh_token = "days"
  }

  # Prevent client secret generation (for public CLI client)
  generate_secret = false
  
  # Enable token revocation
  enable_token_revocation = true
  
  # Explicit auth flows
  explicit_auth_flows = [
    "ALLOW_REFRESH_TOKEN_AUTH",
    "ALLOW_USER_SRP_AUTH"
  ]

  # Read attributes
  read_attributes = [
    "email",
    "email_verified",
    "name",
    "custom:organization",
    "sub"
  ]

  # Write attributes
  write_attributes = [
    "email",
    "name",
    "custom:organization"
  ]
}

# Web App Client (for future marketplace frontend)
resource "aws_cognito_user_pool_client" "web" {
  name                                 = "${var.project_name}-web"
  user_pool_id                         = aws_cognito_user_pool.main.id
  
  # OAuth configuration for web
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code"]
  allowed_oauth_scopes                 = ["email", "openid", "profile"]
  callback_urls                        = ["https://marketplace.api-direct.io/callback", "http://localhost:3000/callback"]
  logout_urls                          = ["https://marketplace.api-direct.io/logout", "http://localhost:3000/logout"]
  supported_identity_providers         = ["COGNITO"]
  
  # Token configuration
  access_token_validity  = 60    # 60 minutes
  id_token_validity      = 60    # 60 minutes
  refresh_token_validity = 7     # 7 days (shorter for web)
  
  token_validity_units {
    access_token  = "minutes"
    id_token      = "minutes"
    refresh_token = "days"
  }

  # Generate client secret for web app
  generate_secret = true
  
  # Enable token revocation
  enable_token_revocation = true
  
  # Explicit auth flows
  explicit_auth_flows = [
    "ALLOW_REFRESH_TOKEN_AUTH",
    "ALLOW_USER_SRP_AUTH"
  ]

  # Read/Write attributes (same as CLI)
  read_attributes = [
    "email",
    "email_verified",
    "name",
    "custom:organization",
    "sub"
  ]

  write_attributes = [
    "email",
    "name",
    "custom:organization"
  ]
}

# Outputs for other services
output "cognito_user_pool_id" {
  value       = aws_cognito_user_pool.main.id
  description = "The ID of the Cognito User Pool"
}

output "cognito_user_pool_arn" {
  value       = aws_cognito_user_pool.main.arn
  description = "The ARN of the Cognito User Pool"
}

output "cognito_user_pool_endpoint" {
  value       = aws_cognito_user_pool.main.endpoint
  description = "The endpoint of the Cognito User Pool"
}

output "cognito_cli_client_id" {
  value       = aws_cognito_user_pool_client.cli.id
  description = "The ID of the CLI Cognito App Client"
}

output "cognito_web_client_id" {
  value       = aws_cognito_user_pool_client.web.id
  description = "The ID of the Web Cognito App Client"
  sensitive   = false
}

output "cognito_web_client_secret" {
  value       = aws_cognito_user_pool_client.web.client_secret
  description = "The secret of the Web Cognito App Client"
  sensitive   = true
}

output "cognito_domain" {
  value       = aws_cognito_user_pool_domain.main.domain
  description = "The Cognito User Pool domain"
}
