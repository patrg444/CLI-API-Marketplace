# AWS Cognito Setup Guide for API-Direct

## Overview

API-Direct uses AWS Cognito for production authentication. This guide explains how to set up Cognito and configure the platform to use it.

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌────────────┐
│   CLI Tool  │────▶│ AWS Cognito  │◀────│  Console   │
└─────────────┘     └──────────────┘     └────────────┘
                            │
                            ▼
                    ┌──────────────┐
                    │ Backend API  │
                    └──────────────┘
```

## Current Setup Status

### ✅ What's Ready:
1. **CLI Tool** - Built with Cognito authentication support
2. **Backend API** - Supports both Cognito and mock auth
3. **Console/Marketplace** - Frontend ready to integrate
4. **Hybrid Auth** - Can switch between mock (local) and Cognito (production)

### 🔧 What Needs AWS Setup:
1. Cognito User Pool creation
2. Cognito App Client configuration
3. Environment variables configuration
4. Domain setup for hosted UI (optional)

## Step 1: Create Cognito User Pool

When you're ready to set up AWS Cognito:

```bash
# Using AWS CLI (when available)
aws cognito-idp create-user-pool \
  --pool-name "apidirect-users" \
  --auto-verified-attributes email \
  --username-attributes email \
  --mfa-configuration "OFF" \
  --email-configuration SourceArn=arn:aws:ses:us-east-1:YOUR_ACCOUNT:identity/noreply@apidirect.dev
```

## Step 2: Create App Client

```bash
aws cognito-idp create-user-pool-client \
  --user-pool-id YOUR_USER_POOL_ID \
  --client-name "apidirect-cli" \
  --generate-secret \
  --explicit-auth-flows ALLOW_USER_PASSWORD_AUTH ALLOW_REFRESH_TOKEN_AUTH \
  --supported-identity-providers COGNITO
```

## Step 3: Configure Environment Variables

Once you have your Cognito resources, update your environment:

```bash
# Production (.env)
COGNITO_USER_POOL_ID=us-east-1_xxxxxxxxx
COGNITO_CLIENT_ID=xxxxxxxxxxxxxxxxxxxxxxxxxx
COGNITO_REGION=us-east-1
USE_MOCK_AUTH=false

# For CLI users
export APIDIRECT_COGNITO_POOL="us-east-1_xxxxxxxxx"
export APIDIRECT_COGNITO_CLIENT="xxxxxxxxxxxxxxxxxxxxxxxxxx"
export APIDIRECT_AUTH_DOMAIN="https://your-domain.auth.us-east-1.amazoncognito.com"
```

## Step 4: Test Authentication Flow

### With CLI:
```bash
# Configure CLI with Cognito
export APIDIRECT_COGNITO_POOL="your-pool-id"
export APIDIRECT_COGNITO_CLIENT="your-client-id"

# Login
apidirect login

# Verify authentication
apidirect whoami
```

### With Backend API:
```bash
# Start backend with Cognito
export USE_MOCK_AUTH=false
export COGNITO_USER_POOL_ID="your-pool-id"
export COGNITO_CLIENT_ID="your-client-id"
./start-local-backend.sh
```

## Local Development (Current State)

For local development without AWS access:

```bash
# Use mock authentication
export USE_MOCK_AUTH=true
./start-local-backend.sh

# Console will work with mock auth
# Visit https://console.apidirect.dev
# Login with: demo@apidirect.dev / secret
```

## Integration Points

### 1. CLI Authentication Flow
- CLI uses Cognito hosted UI for browser-based login
- Stores tokens in `~/.apidirect/config.yaml`
- Automatically refreshes tokens

### 2. Backend API
- Validates Cognito JWT tokens
- Falls back to mock auth if `USE_MOCK_AUTH=true`
- Extracts user info from token claims

### 3. Console/Marketplace
- Uses same Cognito tokens
- Shares authentication with CLI
- WebSocket connections authenticated via JWT

## Security Best Practices

1. **Never commit Cognito credentials**
   - Use environment variables
   - Use AWS Secrets Manager in production

2. **Token Validation**
   - Backend validates JWT signatures
   - Checks token expiration
   - Verifies issuer and audience

3. **CORS Configuration**
   - Restrict to known domains
   - Update for production domains

## Migration Path

When ready to switch from mock to Cognito auth:

1. Create Cognito resources in AWS
2. Update environment variables
3. Set `USE_MOCK_AUTH=false`
4. Restart services
5. Users authenticate via `apidirect login`

## Troubleshooting

### "Missing Cognito configuration"
- Ensure all COGNITO_* environment variables are set
- Check AWS credentials have Cognito access

### "Token validation failed"
- Verify Cognito User Pool ID matches
- Check token hasn't expired
- Ensure backend can reach Cognito JWKS endpoint

### "CORS errors"
- Add your domain to CORS_ORIGINS
- Check both backend and Cognito CORS settings

## Next Steps

1. **When AWS is available:**
   - Create Cognito User Pool
   - Configure App Clients
   - Set up user groups (creators, consumers, admins)

2. **Frontend Integration:**
   - Add Cognito login to Console
   - Implement token refresh
   - Add user profile management

3. **Enhanced Features:**
   - Multi-factor authentication
   - Social login providers
   - Custom user attributes