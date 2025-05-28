#!/bin/bash

# API-Direct Services Deployment Script
# This script builds and deploys backend services to EKS

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}=== API-Direct Services Deployment ===${NC}"

# Check prerequisites
echo -e "\n${YELLOW}Checking prerequisites...${NC}"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is not installed. Please install it first.${NC}"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl is not installed. Please install it first.${NC}"
    exit 1
fi

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    echo -e "${RED}AWS CLI is not installed. Please install it first.${NC}"
    exit 1
fi

# Check if connected to EKS cluster
if ! kubectl get nodes &> /dev/null; then
    echo -e "${RED}Not connected to EKS cluster. Run: aws eks update-kubeconfig --region <region> --name <cluster-name>${NC}"
    exit 1
fi

echo -e "${GREEN}✓ All prerequisites met${NC}"

# Get ECR registry URL from environment or Terraform output
if [ -z "$ECR_REGISTRY" ]; then
    echo -e "${YELLOW}Getting ECR registry URL from Terraform outputs...${NC}"
    cd infrastructure/terraform
    ECR_REGISTRY=$(terraform output -raw ecr_registry_url 2>/dev/null || echo "")
    cd ../..
    
    if [ -z "$ECR_REGISTRY" ]; then
        echo -e "${RED}ECR_REGISTRY not found. Please set it or ensure Terraform has been applied.${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}ECR Registry: ${ECR_REGISTRY}${NC}"

# Login to ECR
echo -e "\n${YELLOW}Logging into ECR...${NC}"
aws ecr get-login-password --region ${AWS_REGION:-us-east-1} | docker login --username AWS --password-stdin $ECR_REGISTRY

# Build and push services
SERVICES=("storage" "deployment")

for SERVICE in "${SERVICES[@]}"; do
    echo -e "\n${YELLOW}Building ${SERVICE} service...${NC}"
    
    # Build Docker image
    docker build -t ${SERVICE}-service:latest ./services/${SERVICE}
    
    # Tag for ECR
    docker tag ${SERVICE}-service:latest ${ECR_REGISTRY}/api-direct/${SERVICE}-service:latest
    
    # Push to ECR
    echo -e "${YELLOW}Pushing ${SERVICE} service to ECR...${NC}"
    docker push ${ECR_REGISTRY}/api-direct/${SERVICE}-service:latest
done

# Deploy Kubernetes resources
echo -e "\n${YELLOW}Deploying Kubernetes resources...${NC}"

# Apply namespaces first
kubectl apply -f infrastructure/k8s/namespace.yaml

# Get IAM role ARNs from Terraform outputs
cd infrastructure/terraform
STORAGE_ROLE_ARN=$(terraform output -raw storage_service_role_arn 2>/dev/null || echo "")
DEPLOYMENT_ROLE_ARN=$(terraform output -raw deployment_service_role_arn 2>/dev/null || echo "")
CODE_BUCKET=$(terraform output -raw code_storage_bucket_name 2>/dev/null || echo "")
cd ../..

# Create secrets
echo -e "${YELLOW}Creating Kubernetes secrets...${NC}"
kubectl create secret generic platform-secrets \
    --from-literal=code-storage-bucket=${CODE_BUCKET} \
    --namespace=api-direct \
    --dry-run=client -o yaml | kubectl apply -f -

# Apply service manifests with substitutions
for MANIFEST in infrastructure/k8s/*.yaml; do
    if [[ $MANIFEST == *"namespace.yaml"* ]]; then
        continue  # Already applied
    fi
    
    echo -e "${YELLOW}Applying $(basename $MANIFEST)...${NC}"
    
    # Substitute environment variables
    envsubst < $MANIFEST | kubectl apply -f -
done

# Wait for deployments to be ready
echo -e "\n${YELLOW}Waiting for deployments to be ready...${NC}"

kubectl wait --for=condition=available --timeout=300s \
    deployment/storage-service deployment/deployment-service \
    -n api-direct

# Get ingress endpoint
echo -e "\n${YELLOW}Getting ALB endpoint...${NC}"
ALB_ENDPOINT=""
for i in {1..60}; do
    ALB_ENDPOINT=$(kubectl get ingress api-direct-ingress -n api-direct -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
    if [ -n "$ALB_ENDPOINT" ]; then
        break
    fi
    echo -n "."
    sleep 5
done

if [ -z "$ALB_ENDPOINT" ]; then
    echo -e "\n${YELLOW}Warning: Could not get ALB endpoint. Check ingress status manually.${NC}"
else
    echo -e "\n${GREEN}✓ ALB Endpoint: ${ALB_ENDPOINT}${NC}"
fi

echo -e "\n${GREEN}=== Deployment complete! ===${NC}"
echo -e "\n${YELLOW}Service endpoints:${NC}"
echo -e "  Storage Service:    http://${ALB_ENDPOINT}/storage/health"
echo -e "  Deployment Service: http://${ALB_ENDPOINT}/deployment/health"
echo -e "\n${YELLOW}Next steps:${NC}"
echo -e "1. Update CLI configuration with the ALB endpoint"
echo -e "2. Test the services with: curl http://${ALB_ENDPOINT}/storage/health"
echo -e "3. Deploy a test API using the CLI"

# Save endpoint for CLI configuration
if [ -n "$ALB_ENDPOINT" ]; then
    echo "export APIDIRECT_API_ENDPOINT=http://${ALB_ENDPOINT}" >> cli-env.sh
    echo -e "\n${GREEN}API endpoint added to cli-env.sh${NC}"
fi
