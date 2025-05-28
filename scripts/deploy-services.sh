#!/bin/bash

# Deploy services to Kubernetes
# This script builds and deploys all microservices to the EKS cluster

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
AWS_REGION=${AWS_REGION:-us-east-1}
ECR_REGISTRY=$(aws ecr describe-repositories --repository-names api-direct-services --region $AWS_REGION --query 'repositories[0].repositoryUri' --output text | cut -d'/' -f1)
CLUSTER_NAME="api-direct-${ENVIRONMENT:-dev}"

echo -e "${GREEN}Starting deployment of API-Direct services...${NC}"

# Ensure we're logged into ECR
echo -e "${YELLOW}Logging into ECR...${NC}"
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $ECR_REGISTRY

# Update kubeconfig
echo -e "${YELLOW}Updating kubeconfig for EKS cluster...${NC}"
aws eks update-kubeconfig --region $AWS_REGION --name $CLUSTER_NAME

# Build and push services
services=(
  "storage:8080"
  "deployment:8081"
  "gateway:8082"
  "apikey:8083"
  "metering:8084"
  "billing:8085"
)

for service_info in "${services[@]}"; do
  IFS=':' read -r service port <<< "$service_info"
  
  echo -e "${YELLOW}Building $service service...${NC}"
  
  # Skip if service directory doesn't exist yet
  if [ ! -d "services/$service" ]; then
    echo -e "${YELLOW}Service $service not implemented yet, skipping...${NC}"
    continue
  fi
  
  # Build Docker image
  docker build -t api-direct-$service:latest services/$service/
  
  # Tag for ECR
  docker tag api-direct-$service:latest $ECR_REGISTRY/api-direct-$service:latest
  
  # Push to ECR
  echo -e "${YELLOW}Pushing $service to ECR...${NC}"
  docker push $ECR_REGISTRY/api-direct-$service:latest
done

# Apply Kubernetes configurations
echo -e "${YELLOW}Applying Kubernetes configurations...${NC}"

# Create namespace if it doesn't exist
kubectl create namespace api-direct --dry-run=client -o yaml | kubectl apply -f -

# Apply infrastructure components
kubectl apply -f infrastructure/k8s/namespace.yaml
kubectl apply -f infrastructure/k8s/redis-service.yaml

# Update ConfigMap with actual values from Terraform
echo -e "${YELLOW}Updating platform configuration...${NC}"
COGNITO_USER_POOL_ID=$(aws cognito-idp list-user-pools --max-results 10 --region $AWS_REGION --query "UserPools[?Name=='api-direct-${ENVIRONMENT:-dev}'].Id" --output text)
kubectl create configmap platform-config \
  --from-literal=aws_region=$AWS_REGION \
  --from-literal=cognito_user_pool_id=$COGNITO_USER_POOL_ID \
  --namespace api-direct \
  --dry-run=client -o yaml | kubectl apply -f -

# Apply service deployments
for service_info in "${services[@]}"; do
  IFS=':' read -r service port <<< "$service_info"
  
  if [ -f "infrastructure/k8s/${service}-service.yaml" ]; then
    echo -e "${YELLOW}Deploying $service service...${NC}"
    # Update image in deployment to use ECR
    sed "s|api-direct-${service}:latest|${ECR_REGISTRY}/api-direct-${service}:latest|g" \
      infrastructure/k8s/${service}-service.yaml | kubectl apply -f -
  fi
done

# Apply ingress rules
echo -e "${YELLOW}Configuring ingress...${NC}"
kubectl apply -f infrastructure/k8s/ingress.yaml

# Wait for deployments to be ready
echo -e "${YELLOW}Waiting for services to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment --all -n api-direct

# Get load balancer URL
LB_URL=$(kubectl get ingress -n api-direct api-direct-ingress -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

echo -e "${GREEN}Deployment complete!${NC}"
echo -e "${GREEN}Services are available at:${NC}"
echo -e "  API Gateway: http://$LB_URL/api"
echo -e "  Storage Service: http://$LB_URL/storage"
echo -e "  Deployment Service: http://$LB_URL/deployment"
echo -e "  API Key Service: http://$LB_URL/apikey"

# Output database migration reminder
echo -e "${YELLOW}Don't forget to run database migrations:${NC}"
echo -e "  psql \$DATABASE_URL < infrastructure/database/migrations/002_marketplace_schema.sql"
