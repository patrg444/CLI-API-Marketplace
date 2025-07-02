#!/bin/bash

# Infrastructure validation test runner
# This script runs all infrastructure tests to ensure configurations are valid

set -e

echo "🚀 Starting Infrastructure Validation Tests"
echo "========================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run a test suite
run_test_suite() {
    local test_name=$1
    local test_command=$2
    
    echo -e "\n${YELLOW}Running $test_name...${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if eval "$test_command"; then
        echo -e "${GREEN}✓ $test_name passed${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ $test_name failed${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# 1. Terraform Validation Tests
echo -e "\n📋 Terraform Validation"
echo "----------------------"

# Check if Terraform is installed
if ! command -v terraform &> /dev/null; then
    echo -e "${RED}Terraform is not installed. Please install Terraform first.${NC}"
    exit 1
fi

# Validate Terraform syntax
for dir in ../terraform/aws/* ../terraform/modules/*; do
    if [ -d "$dir" ]; then
        module_name=$(basename "$dir")
        run_test_suite "Terraform syntax - $module_name" "terraform -chdir=$dir init -backend=false && terraform -chdir=$dir validate"
    fi
done

# Run Terraform unit tests
if command -v go &> /dev/null; then
    run_test_suite "Terraform Go tests" "go test -v ./terraform_test.go -timeout 30m"
else
    echo -e "${YELLOW}Go is not installed. Skipping Terraform Go tests.${NC}"
fi

# 2. Docker Validation Tests
echo -e "\n🐳 Docker Validation"
echo "-------------------"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Lint Dockerfiles
for dockerfile in ../docker/*/Dockerfile; do
    if [ -f "$dockerfile" ]; then
        dir_name=$(basename $(dirname "$dockerfile"))
        run_test_suite "Dockerfile lint - $dir_name" "docker run --rm -i hadolint/hadolint < $dockerfile"
    fi
done

# Validate docker-compose files
for compose_file in ../docker-compose*.yml; do
    if [ -f "$compose_file" ]; then
        file_name=$(basename "$compose_file")
        run_test_suite "Docker Compose validation - $file_name" "docker-compose -f $compose_file config > /dev/null"
    fi
done

# Run Docker Go tests
if command -v go &> /dev/null; then
    run_test_suite "Docker Go tests" "go test -v ./docker_test.go -timeout 20m"
fi

# 3. Kubernetes Validation Tests
echo -e "\n☸️  Kubernetes Validation"
echo "-----------------------"

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo -e "${YELLOW}kubectl is not installed. Skipping Kubernetes validation.${NC}"
else
    # Validate Kubernetes manifests
    for manifest in ../k8s/**/*.yaml; do
        if [ -f "$manifest" ]; then
            manifest_name=$(basename "$manifest")
            run_test_suite "K8s manifest validation - $manifest_name" "kubectl --dry-run=client apply -f $manifest"
        fi
    done
    
    # Run Kubernetes Go tests if cluster is available
    if kubectl cluster-info &> /dev/null; then
        if command -v go &> /dev/null; then
            run_test_suite "Kubernetes Go tests" "go test -v ./kubernetes_test.go -timeout 30m"
        fi
    else
        echo -e "${YELLOW}No Kubernetes cluster available. Skipping cluster tests.${NC}"
    fi
fi

# 4. Security Validation
echo -e "\n🔒 Security Validation"
echo "---------------------"

# Check for secrets in code
run_test_suite "Secret scanning" "! grep -r --include='*.tf' --include='*.yaml' --include='*.yml' 'password\\|secret\\|key' ../terraform ../k8s ../docker | grep -v '# ' | grep -v 'variable' | grep -v 'resource'"

# Check for hardcoded IPs
run_test_suite "Hardcoded IP check" "! grep -r --include='*.tf' --include='*.yaml' '\\b(?:[0-9]{1,3}\\.){3}[0-9]{1,3}\\b' ../terraform ../k8s | grep -v '0.0.0.0' | grep -v '127.0.0.1'"

# 5. Best Practices Validation
echo -e "\n⭐ Best Practices Validation"
echo "---------------------------"

# Check for TODO comments
run_test_suite "TODO comment check" "! grep -r 'TODO\\|FIXME\\|XXX' ../terraform ../k8s ../docker --include='*.tf' --include='*.yaml' --include='*.yml'"

# Check for resource tagging in Terraform
run_test_suite "Resource tagging" "grep -r 'tags\\s*=' ../terraform --include='*.tf' | wc -l | xargs test 10 -lt"

# 6. Performance Tests
echo -e "\n⚡ Performance Validation"
echo "------------------------"

# Test Terraform plan performance
run_test_suite "Terraform plan performance" "time terraform -chdir=../terraform/aws/ecs plan -input=false -out=/dev/null 2>&1 | grep real | awk '{print \$2}' | sed 's/[ms]//g' | awk '{if (\$1 < 30) exit 0; else exit 1}'"

# Summary
echo -e "\n========================================="
echo "📊 Test Summary"
echo "========================================="
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}✅ All infrastructure validation tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Please fix the issues before proceeding.${NC}"
    exit 1
fi