#!/bin/bash

# Script to re-run only the failed tests from Day 1
# This will help verify that the bug fixes work correctly

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "==================================="
echo "Re-running Failed Tests from Day 1"
echo "==================================="
echo ""

# Navigate to E2E test directory
cd "$PROJECT_ROOT/testing/e2e"

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo "⚠️  Dependencies not installed. Run setup-test-env.sh first."
    exit 1
fi

echo "📋 Tests to re-run:"
echo "  1. Price range filtering (Search)"
echo "  2. Set API pricing (API Management)"
echo ""

# Create a temporary test report directory
REPORT_DIR="$PROJECT_ROOT/testing/reports/bug-fixes"
mkdir -p "$REPORT_DIR"

# Re-run only the failed tests with grep pattern
echo "🔄 Re-running failed tests..."
echo ""

# Run tests with specific grep pattern for failed tests
npx playwright test \
    --grep "Price range filtering|Set API pricing" \
    --reporter=list \
    --reporter=html \
    --output=$REPORT_DIR 2>&1 | tee "$REPORT_DIR/rerun-output.log"

# Capture test result
TEST_EXIT_CODE=${PIPESTATUS[0]}

# Generate summary report
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
SUMMARY_FILE="$REPORT_DIR/rerun-summary.md"

cat > "$SUMMARY_FILE" << EOF
# Bug Fix Verification Report

**Date**: $TIMESTAMP
**Purpose**: Verify fixes for Day 1 failed tests

## Fixed Issues

### 1. Price Filter Bug 🔧
- **File**: \`services/marketplace/store/api.go\`
- **Fix**: Updated PriceRange() method to properly track minimum prices
- **Details**: Now correctly categorizes APIs based on their lowest available price

### 2. API Pricing Validation 🔧
- **File**: \`web/creator-portal/src/pages/MarketplaceSettings.js\`
- **Fix**: Added validation to prevent negative pricing values
- **Details**: onChange handlers now check if value >= 0 before updating

## Test Results

EOF

if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo "✅ All tests PASSED!" | tee -a "$SUMMARY_FILE"
    echo "" | tee -a "$SUMMARY_FILE"
    echo "### Summary" | tee -a "$SUMMARY_FILE"
    echo "- ✅ Price range filtering: PASSED" | tee -a "$SUMMARY_FILE"
    echo "- ✅ Set API pricing validation: PASSED" | tee -a "$SUMMARY_FILE"
    echo "" | tee -a "$SUMMARY_FILE"
    echo "**Result**: Both bug fixes have been verified successfully! 🎉" | tee -a "$SUMMARY_FILE"
else
    echo "❌ Some tests still failing" | tee -a "$SUMMARY_FILE"
    echo "" | tee -a "$SUMMARY_FILE"
    echo "Please check the detailed test output in:" | tee -a "$SUMMARY_FILE"
    echo "- Log: \`$REPORT_DIR/rerun-output.log\`" | tee -a "$SUMMARY_FILE"
    echo "- HTML Report: \`$REPORT_DIR/html-report/index.html\`" | tee -a "$SUMMARY_FILE"
fi

echo ""
echo "📊 Reports generated in: $REPORT_DIR"
echo "   - Summary: rerun-summary.md"
echo "   - Full log: rerun-output.log"
echo "   - HTML report: html-report/index.html"
echo ""

# Update main test tracking if all tests pass
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo "🎯 Updating test tracking status..."
    
    # Create updated status file
    cat > "$PROJECT_ROOT/testing/reports/day1/bug-fixes-complete.json" << EOF
{
  "timestamp": "$TIMESTAMP",
  "bugs_fixed": 2,
  "tests_rerun": 2,
  "all_passing": true,
  "fixes": [
    {
      "issue": "Price filter bug",
      "file": "services/marketplace/store/api.go",
      "status": "fixed",
      "test": "Price range filtering",
      "result": "passed"
    },
    {
      "issue": "API pricing validation",
      "file": "web/creator-portal/src/pages/MarketplaceSettings.js", 
      "status": "fixed",
      "test": "Set API pricing",
      "result": "passed"
    }
  ]
}
EOF
    echo "✅ Test tracking updated!"
fi

exit $TEST_EXIT_CODE
