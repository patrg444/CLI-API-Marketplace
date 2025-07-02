#!/bin/bash
# Script to build and publish Docker images for API Direct CLI

set -e

# Configuration
DOCKER_REGISTRY=${DOCKER_REGISTRY:-"docker.io"}
DOCKER_ORG=${DOCKER_ORG:-"apidirect"}
IMAGE_NAME="cli"
VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "latest")}

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🐳 Building Docker images for API Direct CLI v${VERSION}${NC}"

# Build multi-arch images
echo -e "${YELLOW}Building production image...${NC}"
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag "${DOCKER_REGISTRY}/${DOCKER_ORG}/${IMAGE_NAME}:${VERSION}" \
    --tag "${DOCKER_REGISTRY}/${DOCKER_ORG}/${IMAGE_NAME}:latest" \
    --file docker/Dockerfile \
    --push \
    .

echo -e "${YELLOW}Building development image...${NC}"
docker buildx build \
    --platform linux/amd64,linux/arm64 \
    --tag "${DOCKER_REGISTRY}/${DOCKER_ORG}/${IMAGE_NAME}:dev" \
    --file docker/Dockerfile.dev \
    --push \
    .

# Create and push manifest
echo -e "${YELLOW}Creating multi-arch manifest...${NC}"
docker manifest create \
    "${DOCKER_REGISTRY}/${DOCKER_ORG}/${IMAGE_NAME}:${VERSION}" \
    "${DOCKER_REGISTRY}/${DOCKER_ORG}/${IMAGE_NAME}:${VERSION}-amd64" \
    "${DOCKER_REGISTRY}/${DOCKER_ORG}/${IMAGE_NAME}:${VERSION}-arm64"

docker manifest push "${DOCKER_REGISTRY}/${DOCKER_ORG}/${IMAGE_NAME}:${VERSION}"

echo -e "${GREEN}✅ Docker images published successfully!${NC}"
echo -e "Production: ${DOCKER_REGISTRY}/${DOCKER_ORG}/${IMAGE_NAME}:${VERSION}"
echo -e "Development: ${DOCKER_REGISTRY}/${DOCKER_ORG}/${IMAGE_NAME}:dev"