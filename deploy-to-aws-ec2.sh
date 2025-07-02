#!/bin/bash
# AWS EC2 Deployment Script for API Direct Backend

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🚀 API Direct AWS EC2 Deployment${NC}"
echo "=================================="

# Configuration
AWS_REGION="us-east-1"
INSTANCE_TYPE="t3.medium"  # 2 vCPU, 4GB RAM
KEY_NAME="api-direct-key"
SECURITY_GROUP_NAME="api-direct-sg"
INSTANCE_NAME="api-direct-backend"

# Set AWS credentials
export AWS_ACCESS_KEY_ID=REPLACE_WITH_NEW_KEY
export AWS_SECRET_ACCESS_KEY=REPLACE_WITH_NEW_SECRET
export AWS_DEFAULT_REGION=$AWS_REGION

# Function to check if resource exists
resource_exists() {
    local resource_type=$1
    local query=$2
    local name=$3
    
    if aws $resource_type $query --query "$name" --output text 2>/dev/null | grep -q "None\|^$"; then
        return 1
    else
        return 0
    fi
}

echo -e "${YELLOW}1. Creating SSH Key Pair...${NC}"
if ! aws ec2 describe-key-pairs --key-names $KEY_NAME 2>/dev/null; then
    aws ec2 create-key-pair --key-name $KEY_NAME --query 'KeyMaterial' --output text > ~/.ssh/${KEY_NAME}.pem
    chmod 400 ~/.ssh/${KEY_NAME}.pem
    echo -e "${GREEN}✅ Created key pair: $KEY_NAME${NC}"
else
    echo -e "${GREEN}✅ Key pair already exists: $KEY_NAME${NC}"
fi

echo -e "${YELLOW}2. Creating Security Group...${NC}"
SG_ID=$(aws ec2 describe-security-groups --group-names $SECURITY_GROUP_NAME --query 'SecurityGroups[0].GroupId' --output text 2>/dev/null || echo "")

if [ -z "$SG_ID" ] || [ "$SG_ID" == "None" ]; then
    SG_ID=$(aws ec2 create-security-group \
        --group-name $SECURITY_GROUP_NAME \
        --description "Security group for API Direct backend" \
        --query 'GroupId' --output text)
    
    # Add rules
    aws ec2 authorize-security-group-ingress --group-id $SG_ID --protocol tcp --port 22 --cidr 0.0.0.0/0
    aws ec2 authorize-security-group-ingress --group-id $SG_ID --protocol tcp --port 80 --cidr 0.0.0.0/0
    aws ec2 authorize-security-group-ingress --group-id $SG_ID --protocol tcp --port 443 --cidr 0.0.0.0/0
    aws ec2 authorize-security-group-ingress --group-id $SG_ID --protocol tcp --port 8000 --cidr 0.0.0.0/0
    
    echo -e "${GREEN}✅ Created security group: $SG_ID${NC}"
else
    echo -e "${GREEN}✅ Security group already exists: $SG_ID${NC}"
fi

echo -e "${YELLOW}3. Finding Ubuntu 22.04 AMI...${NC}"
AMI_ID=$(aws ec2 describe-images \
    --owners 099720109477 \
    --filters "Name=name,Values=ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*" \
    --query 'sort_by(Images, &CreationDate)[-1].ImageId' \
    --output text)
echo -e "${GREEN}✅ Using AMI: $AMI_ID${NC}"

echo -e "${YELLOW}4. Launching EC2 Instance...${NC}"
# Check if instance already exists
INSTANCE_ID=$(aws ec2 describe-instances \
    --filters "Name=tag:Name,Values=$INSTANCE_NAME" "Name=instance-state-name,Values=running,stopped" \
    --query 'Reservations[0].Instances[0].InstanceId' \
    --output text 2>/dev/null || echo "None")

if [ "$INSTANCE_ID" == "None" ] || [ -z "$INSTANCE_ID" ]; then
    # Create user data script
    cat > /tmp/user-data.sh << 'EOF'
#!/bin/bash
apt-get update
apt-get install -y docker.io docker-compose git nginx certbot python3-certbot-nginx
systemctl enable docker
systemctl start docker
usermod -aG docker ubuntu

# Clone the repository
cd /home/ubuntu
git clone https://github.com/yourusername/CLI-API-Marketplace.git
chown -R ubuntu:ubuntu CLI-API-Marketplace

# Install Docker Compose v2
curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
EOF

    INSTANCE_ID=$(aws ec2 run-instances \
        --image-id $AMI_ID \
        --instance-type $INSTANCE_TYPE \
        --key-name $KEY_NAME \
        --security-group-ids $SG_ID \
        --user-data file:///tmp/user-data.sh \
        --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=$INSTANCE_NAME}]" \
        --block-device-mappings "DeviceName=/dev/sda1,Ebs={VolumeSize=50,VolumeType=gp3}" \
        --query 'Instances[0].InstanceId' \
        --output text)
    
    echo -e "${GREEN}✅ Launched instance: $INSTANCE_ID${NC}"
    echo "Waiting for instance to be running..."
    aws ec2 wait instance-running --instance-ids $INSTANCE_ID
else
    echo -e "${GREEN}✅ Instance already exists: $INSTANCE_ID${NC}"
fi

echo -e "${YELLOW}5. Getting Instance Details...${NC}"
PUBLIC_IP=$(aws ec2 describe-instances \
    --instance-ids $INSTANCE_ID \
    --query 'Reservations[0].Instances[0].PublicIpAddress' \
    --output text)

echo -e "${GREEN}✅ Instance Public IP: $PUBLIC_IP${NC}"

echo -e "${YELLOW}6. Creating Elastic IP...${NC}"
# Check if EIP already allocated
EIP_ALLOC=$(aws ec2 describe-addresses \
    --filters "Name=tag:Name,Values=api-direct-eip" \
    --query 'Addresses[0].AllocationId' \
    --output text 2>/dev/null || echo "None")

if [ "$EIP_ALLOC" == "None" ] || [ -z "$EIP_ALLOC" ]; then
    EIP_ALLOC=$(aws ec2 allocate-address \
        --domain vpc \
        --tag-specifications "ResourceType=elastic-ip,Tags=[{Key=Name,Value=api-direct-eip}]" \
        --query 'AllocationId' \
        --output text)
    echo -e "${GREEN}✅ Allocated Elastic IP${NC}"
else
    echo -e "${GREEN}✅ Elastic IP already allocated${NC}"
fi

# Associate EIP with instance
aws ec2 associate-address --instance-id $INSTANCE_ID --allocation-id $EIP_ALLOC 2>/dev/null || true

ELASTIC_IP=$(aws ec2 describe-addresses \
    --allocation-ids $EIP_ALLOC \
    --query 'Addresses[0].PublicIp' \
    --output text)

echo ""
echo -e "${GREEN}🎉 EC2 Instance Ready!${NC}"
echo "======================"
echo -e "Instance ID: ${YELLOW}$INSTANCE_ID${NC}"
echo -e "Elastic IP: ${YELLOW}$ELASTIC_IP${NC}"
echo -e "SSH Key: ${YELLOW}~/.ssh/${KEY_NAME}.pem${NC}"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Update DNS A record for api.apidirect.dev to: $ELASTIC_IP"
echo "2. SSH into the instance:"
echo "   ssh -i ~/.ssh/${KEY_NAME}.pem ubuntu@$ELASTIC_IP"
echo "3. Deploy the application:"
echo "   cd CLI-API-Marketplace"
echo "   sudo ./deploy-production.sh"
echo ""
echo -e "${YELLOW}To check instance status:${NC}"
echo "aws ec2 describe-instance-status --instance-ids $INSTANCE_ID"

# Save deployment info
cat > deployment-info.txt << EOF
API Direct AWS Deployment Info
==============================
Date: $(date)
Instance ID: $INSTANCE_ID
Elastic IP: $ELASTIC_IP
Region: $AWS_REGION
Security Group: $SG_ID
Key Name: $KEY_NAME
EOF

echo ""
echo -e "${GREEN}Deployment info saved to: deployment-info.txt${NC}"