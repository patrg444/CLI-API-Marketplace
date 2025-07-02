#!/bin/bash

# End-to-End Test Runner for API-Direct CLI
# This script provides various options for running E2E tests

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
TEST_MODE="all"
VERBOSE=false
TIMEOUT="30m"
SKIP_PREREQ=false

# Function to print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check prerequisites
check_prerequisites() {
    print_color "$BLUE" "Checking prerequisites..."
    
    local missing_deps=()
    
    # Check AWS CLI
    if ! command -v aws &> /dev/null; then
        missing_deps+=("aws-cli")
    fi
    
    # Check Terraform
    if ! command -v terraform &> /dev/null; then
        missing_deps+=("terraform")
    fi
    
    # Check Go
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi
    
    # Check apidirect CLI
    if ! command -v apidirect &> /dev/null; then
        print_color "$YELLOW" "Warning: apidirect CLI not found in PATH"
        print_color "$YELLOW" "Building CLI..."
        (cd ../.. && go build -o apidirect)
        export PATH=$PATH:$(pwd)/../..
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        print_color "$RED" "Missing dependencies: ${missing_deps[*]}"
        print_color "$RED" "Please install missing dependencies before running tests"
        exit 1
    fi
    
    # Check AWS credentials
    if aws sts get-caller-identity &> /dev/null; then
        print_color "$GREEN" "✓ AWS credentials configured"
    else
        print_color "$YELLOW" "⚠ AWS credentials not configured (some tests will be skipped)"
    fi
    
    print_color "$GREEN" "✓ All prerequisites satisfied"
}

# Function to run tests
run_tests() {
    local test_filter=""
    local env_vars=""
    
    case $TEST_MODE in
        "all")
            print_color "$BLUE" "Running all E2E tests..."
            test_filter="."
            ;;
        "byoa")
            print_color "$BLUE" "Running BYOA deployment tests..."
            test_filter="TestBYOA"
            ;;
        "hosted")
            print_color "$BLUE" "Running hosted deployment tests..."
            test_filter="TestHosted"
            ;;
        "modes")
            print_color "$BLUE" "Running deployment mode comparison tests..."
            test_filter="TestBothDeploymentModes|TestDeploymentMode"
            ;;
        "mock")
            print_color "$BLUE" "Running mock AWS tests..."
            test_filter="Mock"
            env_vars="MOCK_AWS=true"
            ;;
        "integration")
            print_color "$BLUE" "Running integration tests..."
            test_filter="Integration"
            env_vars="RUN_INTEGRATION_TESTS=true"
            ;;
        "quick")
            print_color "$BLUE" "Running quick tests (no AWS)..."
            test_filter="."
            env_vars="SKIP_E2E_TESTS=true"
            ;;
        *)
            print_color "$RED" "Unknown test mode: $TEST_MODE"
            exit 1
            ;;
    esac
    
    # Build test command
    local cmd="go test"
    
    if [ "$VERBOSE" = true ]; then
        cmd="$cmd -v"
    fi
    
    cmd="$cmd -timeout $TIMEOUT"
    cmd="$cmd -run $test_filter"
    cmd="$cmd ./..."
    
    # Run tests
    if [ -n "$env_vars" ]; then
        env $env_vars $cmd
    else
        $cmd
    fi
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Options:
    -m, --mode MODE       Test mode: all, byoa, mock, integration, quick (default: all)
    -v, --verbose         Enable verbose output
    -t, --timeout TIME    Test timeout (default: 30m)
    -s, --skip-prereq     Skip prerequisite checks
    -h, --help           Show this help message

Test Modes:
    all          Run all E2E tests
    byoa         Run BYOA deployment tests only
    hosted       Run hosted deployment tests only
    modes        Run deployment mode comparison tests
    mock         Run tests with mock AWS services
    integration  Run integration tests (requires RUN_INTEGRATION_TESTS=true)
    quick        Run quick tests without AWS dependencies

Examples:
    $0                    # Run all tests
    $0 -m mock           # Run mock tests only
    $0 -m byoa -v        # Run BYOA tests with verbose output
    $0 -m quick -s       # Run quick tests, skip prerequisites

Environment Variables:
    SKIP_E2E_TESTS=true              Skip E2E tests
    RUN_INTEGRATION_TESTS=true       Enable integration tests
    MOCK_AWS=true                    Use mock AWS services
    APIDIRECT_TEST_MODE=true         Run in test mode (no real deployments)
    AWS_ENDPOINT_URL=<url>           Custom AWS endpoint for testing

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--mode)
            TEST_MODE="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -s|--skip-prereq)
            SKIP_PREREQ=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            print_color "$RED" "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Main execution
print_color "$BLUE" "=== API-Direct E2E Test Runner ==="
print_color "$BLUE" "Test mode: $TEST_MODE"
print_color "$BLUE" "Timeout: $TIMEOUT"
echo

# Check prerequisites unless skipped
if [ "$SKIP_PREREQ" = false ]; then
    check_prerequisites
fi

# Change to test directory
cd "$(dirname "$0")"

# Run tests
if run_tests; then
    print_color "$GREEN" "✓ All tests passed!"
    exit 0
else
    print_color "$RED" "✗ Tests failed!"
    exit 1
fi