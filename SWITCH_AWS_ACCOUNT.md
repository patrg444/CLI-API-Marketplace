# Switching AWS Accounts

## Option 1: Reconfigure Default Profile (Recommended for single account)

```bash
aws configure
```

This will prompt you for:
- AWS Access Key ID: [Enter your new account's access key]
- AWS Secret Access Key: [Enter your new account's secret key]
- Default region name: us-east-1 (or your preferred region)
- Default output format: json

## Option 2: Use Named Profiles (Recommended for multiple accounts)

### Create a new profile for your API marketplace account:

```bash
aws configure --profile apidirect
```

Enter your credentials when prompted.

### To use this profile:

```bash
# Set it as default for this session
export AWS_PROFILE=apidirect

# Or use it for specific commands
aws s3 ls --profile apidirect
```

## Option 3: Set Environment Variables (Temporary)

```bash
export AWS_ACCESS_KEY_ID="your-new-access-key"
export AWS_SECRET_ACCESS_KEY="your-new-secret-key"
export AWS_DEFAULT_REGION="us-east-1"
```

## Verify You're Using the Correct Account

After configuring, verify with:

```bash
aws sts get-caller-identity
```

This should show:
- Account: [Your correct account ID]
- Arn: arn:aws:iam::[account-id]:user/[your-username]

## Get Your AWS Credentials

1. Log in to AWS Console: https://console.aws.amazon.com/
2. Go to IAM → Users → Your username
3. Click "Security credentials" tab
4. Click "Create access key"
5. Select "Command Line Interface (CLI)"
6. Download or copy the credentials

## Update Our Setup Script to Use Profile

If using profiles, update the setup script:

```bash
# Add profile to AWS commands
AWS_PROFILE=apidirect ./setup-aws.sh
```

Or modify ~/.aws/config:

```
[default]
region = us-east-1

[profile apidirect]
region = us-east-1
```

## Security Best Practices

1. **Never commit credentials** to version control
2. **Use IAM roles** when possible (for EC2/ECS)
3. **Rotate access keys** regularly
4. **Use MFA** for your AWS account
5. **Create specific IAM users** with limited permissions

## Ready to Continue?

Once you've configured the correct account, run:

```bash
# Verify you're on the right account
aws sts get-caller-identity

# If correct, proceed with setup
./setup-aws.sh
```