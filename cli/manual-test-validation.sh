#!/bin/bash

echo "🧪 Manual Test Validation for Marketplace Commands"
echo "=================================================="

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track results
TOTAL_TESTS=0
PASSED_TESTS=0

# Function to check if file contains pattern
check_pattern() {
    local file=$1
    local pattern=$2
    local description=$3
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if grep -q "$pattern" "$file"; then
        echo -e "${GREEN}✅ $description${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        echo -e "${RED}❌ $description${NC}"
        return 1
    fi
}

# Function to validate a command file
validate_command() {
    local file=$1
    local cmd_name=$2
    
    echo -e "\n${YELLOW}Validating $file...${NC}"
    
    if [ ! -f "$file" ]; then
        echo -e "${RED}❌ File not found!${NC}"
        return
    fi
    
    # Check basic structure
    check_pattern "$file" "package cmd" "Package declaration"
    check_pattern "$file" "github.com/spf13/cobra" "Cobra import"
    check_pattern "$file" "github.com/fatih/color" "Color import"
    check_pattern "$file" "${cmd_name}Cmd.*=.*&cobra.Command" "Command definition"
    check_pattern "$file" "func init()" "Init function"
    check_pattern "$file" "rootCmd.AddCommand" "Command registration"
}

echo "Starting validation..."

# Validate each command
validate_command "cmd/analytics.go" "analytics"
echo "  Subcommands:"
check_pattern "cmd/analytics.go" "analyticsUsageCmd" "  - usage subcommand"
check_pattern "cmd/analytics.go" "analyticsRevenueCmd" "  - revenue subcommand"
check_pattern "cmd/analytics.go" "analyticsConsumersCmd" "  - consumers subcommand"
check_pattern "cmd/analytics.go" "analyticsPerformanceCmd" "  - performance subcommand"

validate_command "cmd/earnings.go" "earnings"
echo "  Subcommands:"
check_pattern "cmd/earnings.go" "earningsSummaryCmd" "  - summary subcommand"
check_pattern "cmd/earnings.go" "earningsDetailsCmd" "  - details subcommand"
check_pattern "cmd/earnings.go" "earningsPayoutCmd" "  - payout subcommand"
check_pattern "cmd/earnings.go" "earningsHistoryCmd" "  - history subcommand"
check_pattern "cmd/earnings.go" "earningsSetupCmd" "  - setup subcommand"

validate_command "cmd/subscriptions.go" "subscriptions"
echo "  Subcommands:"
check_pattern "cmd/subscriptions.go" "subscriptionsListCmd" "  - list subcommand"
check_pattern "cmd/subscriptions.go" "subscriptionsShowCmd" "  - show subcommand"
check_pattern "cmd/subscriptions.go" "subscriptionsCancelCmd" "  - cancel subcommand"
check_pattern "cmd/subscriptions.go" "subscriptionsUsageCmd" "  - usage subcommand"
check_pattern "cmd/subscriptions.go" "subscriptionsKeysCmd" "  - keys subcommand"

validate_command "cmd/review.go" "review"
echo "  Subcommands:"
check_pattern "cmd/review.go" "reviewSubmitCmd" "  - submit subcommand"
check_pattern "cmd/review.go" "reviewListCmd" "  - list subcommand"
check_pattern "cmd/review.go" "reviewMyCmd" "  - my subcommand"
check_pattern "cmd/review.go" "reviewResponseCmd" "  - respond subcommand"
check_pattern "cmd/review.go" "reviewReportCmd" "  - report subcommand"
check_pattern "cmd/review.go" "reviewStatsCmd" "  - stats subcommand"

validate_command "cmd/search.go" "search"
echo "  Related commands:"
check_pattern "cmd/search.go" "browseCmd" "  - browse command"
check_pattern "cmd/search.go" "trendingCmd" "  - trending command"
check_pattern "cmd/search.go" "featuredCmd" "  - featured command"

# Check test files
echo -e "\n${YELLOW}Checking test files...${NC}"
for test_file in cmd/*_test.go; do
    if [ -f "$test_file" ]; then
        base_name=$(basename "$test_file")
        check_pattern "$test_file" "func Test" "Test functions in $base_name"
        check_pattern "$test_file" "github.com/stretchr/testify/assert" "Assert library in $base_name"
    fi
done

# Check test utilities
echo -e "\n${YELLOW}Checking test utilities...${NC}"
check_pattern "cmd/test_utils.go" "type HTTPClient interface" "HTTP client interface"
check_pattern "cmd/test_utils.go" "makeAuthenticatedRequest" "Auth request helper"
check_pattern "cmd/test_utils.go" "confirmAction" "Confirm action helper"

# Summary
echo -e "\n${YELLOW}========================================${NC}"
echo -e "${YELLOW}Test Summary:${NC}"
echo -e "Total checks: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$((TOTAL_TESTS - PASSED_TESTS))${NC}"

if [ $PASSED_TESTS -eq $TOTAL_TESTS ]; then
    echo -e "\n${GREEN}✅ All validation checks passed!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some validation checks failed.${NC}"
    exit 1
fi