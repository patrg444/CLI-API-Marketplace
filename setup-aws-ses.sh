#!/bin/bash

# AWS SES Setup Script
# This will help automate some of the SES setup

echo "🚀 Setting up AWS SES for API-Direct"
echo "===================================="

# Load AWS credentials from .env
source .env

# Export AWS credentials
export AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID
export AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY
export AWS_DEFAULT_REGION=$AWS_REGION

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo "❌ AWS CLI is not installed. Please install it first:"
    echo "brew install awscli"
    exit 1
fi

echo "✅ AWS CLI found"

# Verify identity (email first, easier for sandbox)
echo ""
echo "📧 Setting up email verification..."
read -p "Enter the email address to verify (e.g., admin@apidirect.dev): " EMAIL

# Verify email address
aws ses verify-email-identity --email-address $EMAIL

echo "✅ Verification email sent to $EMAIL"
echo "⚠️  Please check your email and click the verification link!"
echo ""
read -p "Press Enter after you've verified the email..."

# Get SMTP password from AWS secret key
echo ""
echo "🔐 Generating SMTP credentials..."

# This is the algorithm to convert AWS Secret Key to SES SMTP password
# For us-east-1 region
python3 << EOF
import hmac
import hashlib
import base64

secret_key = "$AWS_SECRET_ACCESS_KEY"
message = "SendRawEmail"
version = b'\x04'

signature = hmac.new(
    ('AWS4' + secret_key).encode('utf-8'),
    message.encode('utf-8'),
    hashlib.sha256
).digest()

smtp_password = base64.b64encode(version + signature).decode('utf-8')
print(f"SMTP Password: {smtp_password}")
EOF

echo ""
echo "📝 Add these to your .env:"
echo "====================================="
echo "SMTP_HOST=email-smtp.$AWS_REGION.amazonaws.com"
echo "SMTP_PORT=587"
echo "SMTP_USER=$AWS_ACCESS_KEY_ID"
echo "# SMTP_PASSWORD= (see above)"
echo "EMAIL_FROM=$EMAIL"
echo ""

# Check sending statistics
echo "📊 Checking SES sending quota..."
aws ses describe-configuration-set --configuration-set-name default 2>/dev/null || echo "No configuration set found"

echo ""
echo "⚠️  Important Notes:"
echo "1. SES starts in sandbox mode - you can only send to verified emails"
echo "2. To send to any email, request production access in AWS Console"
echo "3. Make sure to verify your domain for better deliverability"
echo ""
echo "✅ Basic SES setup complete!"