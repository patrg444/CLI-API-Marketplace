# API-Direct Testing Suite

Comprehensive testing infrastructure for the API-Direct platform covering E2E tests, performance testing, security audits, and test data generation.

## Overview

This testing suite is designed for the final Testing & Polish Sprint before launch. It includes:

- 🧪 **End-to-End Testing** with Playwright
- 🚀 **Performance Testing** with k6
- 🔒 **Security Testing** scripts
- 📊 **Test Data Generation** with faker.js
- ✅ **Manual Testing Checklists**

## Prerequisites

### Required Software
- Node.js 18+ and npm
- Docker and Docker Compose
- k6 (for load testing)
- PostgreSQL client tools
- Chrome, Firefox, Safari browsers

### Installation

1. **Install E2E Testing Dependencies**
```bash
cd testing/e2e
npm install
npx playwright install
```

2. **Install k6 for Load Testing**
```bash
# macOS
brew install k6

# Linux
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Windows (using Chocolatey)
choco install k6
```

3. **Install Test Data Generator Dependencies**
```bash
cd testing/data-generators
npm install
```

## Quick Start

### 1. Start All Services
```bash
# From project root
docker-compose up -d
```

### 2. Generate Test Data
```bash
cd testing/data-generators
node generate-test-data.js
```

### 3. Run E2E Tests
```bash
cd testing/e2e

# Run all tests
npm test

# Run specific test suites
npm run test:search
npm run test:reviews
npm run test:consumer
npm run test:creator

# Run with UI mode
npm run test:ui

# Run in debug mode
npm run test:debug
```

### 4. Run Performance Tests
```bash
cd testing/performance

# Run load test
k6 run k6-load-test.js

# Run with custom parameters
k6 run -u 500 -d 10m k6-load-test.js

# Run with environment variables
k6 run -e BASE_URL=https://api-direct.com k6-load-test.js
```

## Test Structure

```
testing/
├── e2e/                    # End-to-end tests
│   ├── tests/
│   │   ├── search/         # Search functionality tests
│   │   ├── reviews/        # Review system tests
│   │   ├── consumer-flows/ # Consumer journey tests
│   │   └── creator-flows/  # Creator journey tests
│   ├── playwright.config.ts
│   └── package.json
├── performance/            # Performance tests
│   ├── k6-load-test.js    # Main load test script
│   └── scenarios/         # Additional test scenarios
├── security/              # Security testing
│   ├── api-security.sh    # API security tests
│   └── owasp-zap.sh      # OWASP ZAP scanner
├── data-generators/       # Test data generation
│   └── generate-test-data.js
├── checklists/           # Manual testing checklists
│   └── MASTER_CHECKLIST.md
└── reports/              # Test results and reports
```

## Test Scenarios

### E2E Test Coverage

#### Search & Discovery
- Basic text search
- Fuzzy search with typos
- Autocomplete suggestions
- Category filtering
- Price range filtering
- Rating filtering
- Tag-based search
- Multi-filter combinations
- Sorting options
- Pagination
- URL persistence

#### Review System
- Review display and statistics
- Review submission (authenticated)
- Character limits and validation
- Helpful/not helpful voting
- Review sorting
- Creator responses
- Verified purchase badges

#### Consumer Journey
- Sign up and email verification
- Browse marketplace
- Search and filter APIs
- View API details
- Subscribe to API
- Get and manage API keys
- Check usage statistics
- Submit reviews
- Manage subscriptions

#### Creator Journey
- Sign up as creator
- Create and deploy API
- Publish to marketplace
- Set pricing plans
- Upload documentation
- Complete Stripe Connect
- View earnings
- Respond to reviews

### Performance Targets

| Metric | Target | Critical |
|--------|--------|----------|
| Search Response Time | < 200ms (p95) | < 500ms |
| API Details Load | < 150ms (p95) | < 300ms |
| Review Submission | < 500ms (p95) | < 1s |
| Dashboard Load | < 300ms (p95) | < 600ms |
| Concurrent Users | 1000 | 500 |
| Error Rate | < 1% | < 5% |

## Test Data

The test data generator creates:
- 50 creators with verified accounts
- 200 consumers with subscriptions
- 100 published APIs across all categories
- 500 reviews with ratings
- API keys and usage data
- Realistic pricing plans

### Using Test Data

1. **Load into Database**
```bash
# Import test data
psql -h localhost -U apidirect -d apidirect < test-data/seed.sql
```

2. **Index in Elasticsearch**
```bash
# Run indexer
curl -X POST http://localhost:8086/api/v1/marketplace/admin/reindex
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: E2E Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - name: Install dependencies
        run: |
          cd testing/e2e
          npm ci
          npx playwright install
      - name: Run tests
        run: |
          cd testing/e2e
          npm test
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: playwright-report
          path: testing/e2e/playwright-report/
```

## Troubleshooting

### Common Issues

1. **Services not starting**
   - Check Docker logs: `docker-compose logs [service-name]`
   - Ensure ports are not in use
   - Verify environment variables

2. **E2E tests failing**
   - Check service health: `docker-compose ps`
   - Review screenshots in `test-results/`
   - Run in debug mode: `npm run test:debug`

3. **Performance test errors**
   - Verify service URLs in test scripts
   - Check network connectivity
   - Monitor service resources during tests

### Debug Commands

```bash
# Check service health
curl http://localhost:8086/health

# View Elasticsearch indices
curl http://localhost:9200/_cat/indices

# Check Redis
redis-cli ping

# Database connection
psql -h localhost -U apidirect -d apidirect -c "SELECT 1"
```

## Reporting

### E2E Test Reports
- HTML report: `testing/e2e/playwright-report/index.html`
- JUnit XML: `testing/e2e/test-results/junit.xml`
- JSON results: `testing/e2e/test-results/results.json`

### Performance Test Reports
- Console output with summary statistics
- HTML report: `testing/performance/summary.html`
- JSON metrics: `testing/performance/summary.json`

### Manual Testing
Use the checklists in `testing/checklists/` to track manual testing progress.

## Best Practices

1. **Run tests in isolation**: Each test should be independent
2. **Use test data**: Don't rely on production data
3. **Clean up**: Reset state between test runs
4. **Monitor resources**: Watch CPU/memory during load tests
5. **Document failures**: Include screenshots and logs
6. **Test incrementally**: Run quick tests before full suite

## Contributing

1. Add new tests to appropriate directories
2. Follow existing naming conventions
3. Include data-testid attributes for E2E tests
4. Document new test scenarios
5. Update checklists for manual tests

## Resources

- [Playwright Documentation](https://playwright.dev/docs/intro)
- [k6 Documentation](https://k6.io/docs/)
- [OWASP Testing Guide](https://owasp.org/www-project-web-security-testing-guide/)
- [API-Direct Testing Plan](./TESTING_PLAN.md)
