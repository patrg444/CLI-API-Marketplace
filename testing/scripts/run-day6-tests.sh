#!/bin/bash

echo "======================================="
echo "Day 6 Testing: Documentation & Polish"
echo "======================================="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Report directory
REPORT_DIR="testing/reports/day6"
mkdir -p $REPORT_DIR

# Initialize report
REPORT_FILE="$REPORT_DIR/documentation-polish-report.md"
echo "# Day 6: Documentation & Polish Report" > $REPORT_FILE
echo "Date: $(date)" >> $REPORT_FILE
echo "" >> $REPORT_FILE

echo "📚 Checking Documentation..."
echo "=============================="
echo "" >> $REPORT_FILE
echo "## Documentation Review" >> $REPORT_FILE
echo "" >> $REPORT_FILE

# Check main documentation files
DOCS=("README.md" "DEPLOYMENT.md" "PHASE2_IMPLEMENTATION.md")
for doc in "${DOCS[@]}"; do
    if [ -f "../$doc" ]; then
        echo -e "${GREEN}✓${NC} Found: $doc"
        echo "- ✅ Found: $doc" >> $REPORT_FILE
        
        # Check for TODO items
        TODO_COUNT=$(grep -ci "TODO" "../$doc" || true)
        if [ $TODO_COUNT -gt 0 ]; then
            echo -e "  ${YELLOW}⚠${NC} Contains $TODO_COUNT TODO items"
            echo "  - ⚠️ Contains $TODO_COUNT TODO items" >> $REPORT_FILE
        fi
    else
        echo -e "${RED}✗${NC} Missing: $doc"
        echo "- ❌ Missing: $doc" >> $REPORT_FILE
    fi
done

echo ""
echo "🎨 Checking UI/UX Consistency..."
echo "================================="
echo "" >> $REPORT_FILE
echo "## UI/UX Consistency" >> $REPORT_FILE
echo "" >> $REPORT_FILE

# Check for loading states in React components
echo "Checking loading states..."
LOADING_COUNT=$(find ../web -name "*.js" -o -name "*.jsx" -o -name "*.tsx" | xargs grep -l "loading" | wc -l || echo 0)
echo -e "${GREEN}✓${NC} Found loading states in $LOADING_COUNT components"
echo "- ✅ Loading states found in $LOADING_COUNT components" >> $REPORT_FILE

# Check for error handling
echo "Checking error handling..."
ERROR_COUNT=$(find ../web -name "*.js" -o -name "*.jsx" -o -name "*.tsx" | xargs grep -l "error" | wc -l || echo 0)
echo -e "${GREEN}✓${NC} Found error handling in $ERROR_COUNT components"
echo "- ✅ Error handling found in $ERROR_COUNT components" >> $REPORT_FILE

echo ""
echo "♿ Checking Accessibility..."
echo "============================"
echo "" >> $REPORT_FILE
echo "## Accessibility Audit" >> $REPORT_FILE
echo "" >> $REPORT_FILE

# Check for alt text on images
echo "Checking alt attributes..."
NO_ALT=$(find ../web -name "*.js" -o -name "*.jsx" -o -name "*.tsx" | xargs grep -E '<img[^>]+>' | grep -v 'alt=' | wc -l || true)
if [ $NO_ALT -eq 0 ]; then
    echo -e "${GREEN}✓${NC} All images have alt attributes"
    echo "- ✅ All images have alt attributes" >> $REPORT_FILE
else
    echo -e "${RED}✗${NC} Found $NO_ALT images without alt attributes"
    echo "- ❌ Found $NO_ALT images without alt attributes" >> $REPORT_FILE
fi

# Check for ARIA labels
echo "Checking ARIA labels..."
ARIA_COUNT=$(find ../web -name "*.js" -o -name "*.jsx" -o -name "*.tsx" | xargs grep -E 'aria-[a-z]+' | wc -l || true)
echo -e "${GREEN}✓${NC} Found $ARIA_COUNT ARIA attributes"
echo "- ✅ Found $ARIA_COUNT ARIA attributes" >> $REPORT_FILE

# Check for semantic HTML
echo "Checking semantic HTML..."
SEMANTIC_COUNT=$(find ../web -name "*.js" -o -name "*.jsx" -o -name "*.tsx" | xargs grep -E '<(nav|header|main|footer|article|section)' | wc -l || true)
echo -e "${GREEN}✓${NC} Found $SEMANTIC_COUNT semantic HTML elements"
echo "- ✅ Found $SEMANTIC_COUNT semantic HTML elements" >> $REPORT_FILE

echo ""
echo "🔍 Checking Code Quality..."
echo "==========================="
echo "" >> $REPORT_FILE
echo "## Code Quality" >> $REPORT_FILE
echo "" >> $REPORT_FILE

# Check for console.log statements
echo "Checking for console.log statements..."
CONSOLE_COUNT=$(find ../services ../web -name "*.js" -o -name "*.go" | xargs grep -n "console\.log\|fmt\.Println" | wc -l || true)
if [ $CONSOLE_COUNT -eq 0 ]; then
    echo -e "${GREEN}✓${NC} No console.log statements found"
    echo "- ✅ No console.log statements found" >> $REPORT_FILE
else
    echo -e "${YELLOW}⚠${NC} Found $CONSOLE_COUNT console.log/debug print statements"
    echo "- ⚠️ Found $CONSOLE_COUNT console.log/debug print statements" >> $REPORT_FILE
fi

# Check for TODO comments
echo "Checking for TODO comments..."
TODO_CODE_COUNT=$(find ../services ../web -name "*.js" -o -name "*.go" -o -name "*.jsx" -o -name "*.tsx" | xargs grep -n "TODO\|FIXME\|XXX" | wc -l || true)
if [ $TODO_CODE_COUNT -eq 0 ]; then
    echo -e "${GREEN}✓${NC} No TODO comments found"
    echo "- ✅ No TODO comments found" >> $REPORT_FILE
else
    echo -e "${YELLOW}⚠${NC} Found $TODO_CODE_COUNT TODO/FIXME comments"
    echo "- ⚠️ Found $TODO_CODE_COUNT TODO/FIXME comments" >> $REPORT_FILE
fi

echo ""
echo "📝 Checking Error Messages..."
echo "============================="
echo "" >> $REPORT_FILE
echo "## Error Messages" >> $REPORT_FILE
echo "" >> $REPORT_FILE

# Sample error messages from different services
echo "Sampling error messages from services..."
ERROR_SAMPLES=$(find ../services -name "*.go" | xargs grep -h "errors\.\|fmt\.Errorf" | head -5 || echo "No error samples found")
echo "Sample error messages found:" >> $REPORT_FILE
echo "\`\`\`" >> $REPORT_FILE
echo "$ERROR_SAMPLES" >> $REPORT_FILE
echo "\`\`\`" >> $REPORT_FILE

echo ""
echo "📦 Checking Package Dependencies..."
echo "==================================="
echo "" >> $REPORT_FILE
echo "## Package Dependencies" >> $REPORT_FILE
echo "" >> $REPORT_FILE

# Check for outdated packages in web apps
for app in ../web/marketplace ../web/creator-portal; do
    if [ -f "$app/package.json" ]; then
        echo "Checking $app dependencies..."
        echo "### $app" >> $REPORT_FILE
        
        # Count total dependencies
        DEP_COUNT=$(cd $app && npm list --depth=0 2>/dev/null | wc -l || echo "0")
        echo -e "${GREEN}✓${NC} $app has approximately $DEP_COUNT dependencies"
        echo "- Total dependencies: ~$DEP_COUNT" >> $REPORT_FILE
    fi
done

echo ""
echo "🌐 Checking API Documentation..."
echo "================================"
echo "" >> $REPORT_FILE
echo "## API Documentation" >> $REPORT_FILE
echo "" >> $REPORT_FILE

# Check for OpenAPI/Swagger files
OPENAPI_COUNT=$(find .. -name "*.yaml" -o -name "*.yml" -o -name "*.json" | xargs grep -l "openapi\|swagger" | wc -l || true)
echo -e "${GREEN}✓${NC} Found $OPENAPI_COUNT OpenAPI specification files"
echo "- ✅ Found $OPENAPI_COUNT OpenAPI specification files" >> $REPORT_FILE

echo ""
echo "✅ Generating Summary..."
echo "======================="
echo "" >> $REPORT_FILE
echo "## Summary" >> $REPORT_FILE
echo "" >> $REPORT_FILE

# Generate summary
cat >> $REPORT_FILE << EOF

### Completed Checks:
1. Documentation files presence and TODO items
2. UI/UX consistency (loading states, error handling)
3. Accessibility (alt text, ARIA labels, semantic HTML)
4. Code quality (console logs, TODO comments)
5. Error message sampling
6. Package dependency overview
7. API documentation presence

### Key Findings:
- Loading states: $LOADING_COUNT components
- Error handling: $ERROR_COUNT components
- ARIA attributes: $ARIA_COUNT occurrences
- Semantic HTML: $SEMANTIC_COUNT elements
- Debug statements: $CONSOLE_COUNT found
- TODO comments: $TODO_CODE_COUNT found
- OpenAPI specs: $OPENAPI_COUNT files

### Recommendations:
1. Review and resolve TODO comments
2. Ensure all images have descriptive alt text
3. Remove debug console.log statements before production
4. Verify error messages are user-friendly
5. Keep dependencies up to date
EOF

echo ""
echo "======================================="
echo "Day 6 Testing Complete!"
echo "======================================="
echo ""
echo "📊 Report saved to: $REPORT_FILE"
echo ""

# Display summary
if [ $NO_ALT -gt 0 ] || [ $CONSOLE_COUNT -gt 0 ]; then
    echo -e "${YELLOW}⚠️  Some issues found that need attention${NC}"
else
    echo -e "${GREEN}✅ Documentation and polish checks passed!${NC}"
fi
