#!/bin/bash

# Comprehensive test runner for new test suites
# This script runs all newly created tests with proper reporting

set -e

echo "🧪 Starting Comprehensive Test Suite..."
echo "====================================="

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create test results directory
TEST_RESULTS_DIR="test-results/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$TEST_RESULTS_DIR"

# Test counters
TOTAL_SUITES=0
PASSED_SUITES=0
FAILED_SUITES=0

# Function to run a test suite
run_test_suite() {
    local suite_name=$1
    local test_command=$2
    local test_dir=$3
    
    echo -e "\n${BLUE}📋 Running: ${suite_name}${NC}"
    echo "----------------------------------------"
    
    TOTAL_SUITES=$((TOTAL_SUITES + 1))
    
    # Change to test directory if specified
    if [ -n "$test_dir" ]; then
        cd "$test_dir"
    fi
    
    # Run the test and capture result
    if eval "$test_command" > "$TEST_RESULTS_DIR/${suite_name}.log" 2>&1; then
        echo -e "${GREEN}✅ PASSED: ${suite_name}${NC}"
        PASSED_SUITES=$((PASSED_SUITES + 1))
        
        # Extract key metrics if available
        if grep -q "Success Rate:" "$TEST_RESULTS_DIR/${suite_name}.log"; then
            grep "Success Rate:" "$TEST_RESULTS_DIR/${suite_name}.log"
        fi
    else
        echo -e "${RED}❌ FAILED: ${suite_name}${NC}"
        FAILED_SUITES=$((FAILED_SUITES + 1))
        
        # Show last few lines of error
        echo -e "${RED}Error details:${NC}"
        tail -n 10 "$TEST_RESULTS_DIR/${suite_name}.log"
    fi
    
    # Return to original directory
    if [ -n "$test_dir" ]; then
        cd - > /dev/null
    fi
}

# 1. API Integration Tests
echo -e "${YELLOW}🔌 API Integration Tests${NC}"
run_test_suite "api-integration" \
    "node web/marketplace/tests/api-integration.test.js" \
    ""

# 2. Security Tests
echo -e "\n${YELLOW}🔒 Security Tests${NC}"
run_test_suite "auth-security" \
    "npx playwright test tests/security/auth-security.spec.ts --reporter=json" \
    "testing/e2e"

# 3. Performance Tests
echo -e "\n${YELLOW}⚡ Performance Tests${NC}"
# Check if k6 is installed
if command -v k6 &> /dev/null; then
    run_test_suite "api-performance" \
        "k6 run --out json=$TEST_RESULTS_DIR/performance-metrics.json testing/performance/api-performance-test.js" \
        ""
else
    echo -e "${YELLOW}⚠️  k6 not installed. Skipping performance tests.${NC}"
    echo "Install k6: brew install k6"
fi

# 4. Console Integration Tests
echo -e "\n${YELLOW}🖥️  Console Integration Tests${NC}"
run_test_suite "console-integration" \
    "npx playwright test tests/console/console-integration.spec.ts --reporter=json" \
    "testing/e2e"

# 5. Run existing E2E tests
echo -e "\n${YELLOW}🔄 Running Existing E2E Tests${NC}"
run_test_suite "existing-e2e" \
    "npx playwright test --grep-invert 'security|console' --reporter=json" \
    "testing/e2e"

# 6. Generate combined test report
echo -e "\n${BLUE}📊 Generating Combined Test Report...${NC}"

cat > "$TEST_RESULTS_DIR/summary.json" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "total_suites": $TOTAL_SUITES,
  "passed_suites": $PASSED_SUITES,
  "failed_suites": $FAILED_SUITES,
  "success_rate": $(echo "scale=2; $PASSED_SUITES * 100 / $TOTAL_SUITES" | bc)%,
  "test_results_directory": "$TEST_RESULTS_DIR"
}
EOF

# Generate HTML report
cat > "$TEST_RESULTS_DIR/report.html" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Test Results - $(date +%Y-%m-%d)</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
        .summary { display: flex; justify-content: space-around; margin: 20px 0; }
        .metric { text-align: center; padding: 20px; background: #ecf0f1; border-radius: 5px; }
        .metric h3 { margin: 0; color: #34495e; }
        .metric .value { font-size: 2em; font-weight: bold; margin: 10px 0; }
        .passed { color: #27ae60; }
        .failed { color: #e74c3c; }
        .suite-results { margin: 20px 0; }
        .suite { padding: 10px; margin: 5px 0; border-radius: 5px; }
        .suite.passed { background: #d4edda; border: 1px solid #c3e6cb; }
        .suite.failed { background: #f8d7da; border: 1px solid #f5c6cb; }
        .recommendations { background: #fff3cd; padding: 15px; border-radius: 5px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Comprehensive Test Results</h1>
        <p>Generated: $(date)</p>
    </div>
    
    <div class="summary">
        <div class="metric">
            <h3>Total Suites</h3>
            <div class="value">$TOTAL_SUITES</div>
        </div>
        <div class="metric">
            <h3>Passed</h3>
            <div class="value passed">$PASSED_SUITES</div>
        </div>
        <div class="metric">
            <h3>Failed</h3>
            <div class="value failed">$FAILED_SUITES</div>
        </div>
        <div class="metric">
            <h3>Success Rate</h3>
            <div class="value">$(echo "scale=1; $PASSED_SUITES * 100 / $TOTAL_SUITES" | bc)%</div>
        </div>
    </div>
    
    <div class="suite-results">
        <h2>Test Suite Results</h2>
EOF

# Add individual suite results to HTML
for log_file in "$TEST_RESULTS_DIR"/*.log; do
    if [ -f "$log_file" ]; then
        suite_name=$(basename "$log_file" .log)
        if grep -q "PASSED\|Success\|✅" "$log_file"; then
            echo "<div class='suite passed'>✅ $suite_name</div>" >> "$TEST_RESULTS_DIR/report.html"
        else
            echo "<div class='suite failed'>❌ $suite_name</div>" >> "$TEST_RESULTS_DIR/report.html"
        fi
    fi
done

cat >> "$TEST_RESULTS_DIR/report.html" << EOF
    </div>
    
    <div class="recommendations">
        <h2>Recommendations</h2>
        <ul>
EOF

# Add recommendations based on results
if [ $FAILED_SUITES -gt 0 ]; then
    echo "<li>Review failed test logs in: $TEST_RESULTS_DIR</li>" >> "$TEST_RESULTS_DIR/report.html"
    echo "<li>Fix failing tests before deployment</li>" >> "$TEST_RESULTS_DIR/report.html"
fi

if [ $PASSED_SUITES -eq $TOTAL_SUITES ]; then
    echo "<li class='passed'>All tests passing! Ready for deployment.</li>" >> "$TEST_RESULTS_DIR/report.html"
fi

cat >> "$TEST_RESULTS_DIR/report.html" << EOF
            <li>Run performance tests under production-like load</li>
            <li>Monitor API response times in production</li>
            <li>Review security test results regularly</li>
        </ul>
    </div>
</body>
</html>
EOF

# Display summary
echo -e "\n${BLUE}======================================${NC}"
echo -e "${BLUE}📊 TEST EXECUTION SUMMARY${NC}"
echo -e "${BLUE}======================================${NC}"
echo -e "Total Test Suites: ${TOTAL_SUITES}"
echo -e "${GREEN}Passed: ${PASSED_SUITES}${NC}"
echo -e "${RED}Failed: ${FAILED_SUITES}${NC}"
echo -e "Success Rate: $(echo "scale=1; $PASSED_SUITES * 100 / $TOTAL_SUITES" | bc)%"
echo -e "\nTest results saved to: ${TEST_RESULTS_DIR}"
echo -e "HTML Report: ${TEST_RESULTS_DIR}/report.html"

# Open HTML report if on macOS
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo -e "\n${BLUE}Opening test report...${NC}"
    open "$TEST_RESULTS_DIR/report.html"
fi

# Set exit code based on test results
if [ $FAILED_SUITES -gt 0 ]; then
    echo -e "\n${RED}❌ Some tests failed. Please review the results.${NC}"
    exit 1
else
    echo -e "\n${GREEN}✅ All tests passed successfully!${NC}"
    exit 0
fi