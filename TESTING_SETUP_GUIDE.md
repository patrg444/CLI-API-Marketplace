# API-Direct Testing Setup Guide

This guide provides step-by-step instructions for setting up and running the complete end-to-end testing suite for the API-Direct platform.

## Prerequisites

### Required Software
- **Docker** and **Docker Compose** (for backend services)
- **Node.js 18+** and **npm** (for frontend and tests)
- **Go 1.21+** (for fixing backend service dependencies)

### Installation Commands

```bash
# macOS (using Homebrew)
brew install docker node go

# Ubuntu/Debian
sudo apt update
sudo apt install docker.io docker-compose nodejs npm golang-go

# Windows (using Chocolatey)
choco install docker-desktop nodejs golang
```

## Quick Start (Automated)

The fastest way to run tests is using our automated script:

```bash
# From project root
./scripts/run-e2e-tests.sh
```

This script will:
1. ✅ Check all prerequisites
2. 🐳 Start backend services (PostgreSQL, Redis, Elasticsearch)
3. 🌐 Start the marketplace application on port 3001
4. 🧪 Install and run the complete Playwright test suite
5. 📊 Generate HTML, JUnit, and JSON test reports
6. 🧹 Clean up all services when complete

## Manual Setup (Step by Step)

If you prefer to set up the environment manually or need to debug issues:

### Step 1: Fix Go Dependencies (One-time setup)

```bash
# Install Go dependencies for all services
./scripts/fix-go-dependencies.sh
```

### Step 2: Configure Environment Variables

```bash
# Copy the environment template
cp web/marketplace/.env.template web/marketplace/.env.local

# Edit the file with your actual values (optional for testing)
# The template includes placeholder values that work for local development
```

### Step 3: Start Backend Services

```bash
# Start test database and services
docker-compose -f docker-compose.test.yml up -d

# Wait for services to be ready (about 30 seconds)
docker-compose -f docker-compose.test.yml ps
```

### Step 4: Start Marketplace Application

```bash
# Install dependencies (first time only)
cd web/marketplace
npm install

# Start the marketplace on port 3001
npm run dev -- --port 3001
```

### Step 5: Run Tests

```bash
# In a new terminal, install test dependencies (first time only)
cd testing/e2e
npm install
npx playwright install

# Run the full test suite
npm test

# Or run specific test categories
npm run test:consumer    # Consumer journey tests
npm run test:creator     # Creator workflow tests  
npm run test:search      # Search functionality tests
npm run test:reviews     # Review system tests
```

## Test Results and Reports

After running tests, you'll find results in:

```
testing/e2e/
├── playwright-report/           # Interactive HTML report
│   └── index.html              # Open this in your browser
├── test-results/               # Raw test artifacts
│   ├── junit.xml              # JUnit format for CI/CD
│   ├── results.json           # JSON format for automation
│   └── [test-folders]/        # Screenshots and videos for failed tests
```

### Viewing the HTML Report

```bash
# From the testing/e2e directory
npx playwright show-report
```

## Troubleshooting

### Common Issues and Solutions

#### 1. "Go is not installed" Error
```bash
# Install Go using your package manager
# macOS: brew install go
# Ubuntu: sudo apt install golang-go
# Then run: ./scripts/fix-go-dependencies.sh
```

#### 2. "Port 3001 already in use" Error
```bash
# Find and kill the process using port 3001
lsof -ti:3001 | xargs kill -9
```

#### 3. "Docker services not ready" Error
```bash
# Check service status
docker-compose -f docker-compose.test.yml ps

# View service logs
docker-compose -f docker-compose.test.yml logs [service-name]

# Restart services
docker-compose -f docker-compose.test.yml down
docker-compose -f docker-compose.test.yml up -d
```

#### 4. "Marketplace won't start" Error
```bash
# Check if .env.local exists
ls -la web/marketplace/.env.local

# View marketplace logs
tail -f logs/marketplace.log

# Check for missing environment variables
cd web/marketplace && npm run dev -- --port 3001
```

#### 5. Test Failures Due to Timing
```bash
# Run tests with longer timeout
cd testing/e2e
npx playwright test --timeout=90000

# Run tests in headed mode to see what's happening
npm run test:headed
```

### Debug Mode

To debug failing tests interactively:

```bash
cd testing/e2e

# Run in debug mode (opens browser with debugger)
npm run test:debug

# Run specific test file
npx playwright test tests/consumer-flows/subscription-journey.spec.ts --debug

# Run with UI mode for test exploration
npm run test:ui
```

## Environment Variables Reference

### Required for Full Functionality
```bash
# AWS Cognito (for user authentication)
REACT_APP_COGNITO_USER_POOL_ID=us-east-1_XXXXXXXXX
REACT_APP_COGNITO_CLIENT_ID=your_client_id
REACT_APP_COGNITO_REGION=us-east-1

# Stripe (for payment processing)
REACT_APP_STRIPE_PUBLISHABLE_KEY=pk_test_XXXXXXXXX
```

### Optional (have defaults)
```bash
# API Service URLs (default to localhost)
REACT_APP_API_URL=http://localhost:8082
REACT_APP_APIKEY_SERVICE_URL=http://localhost:8083
```

## CI/CD Integration

To run these tests in your CI/CD pipeline:

```yaml
# GitHub Actions example
name: E2E Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - name: Run E2E Tests
        run: ./scripts/run-e2e-tests.sh
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: test-results
          path: testing/e2e/test-results/
```

## Performance Considerations

### Test Execution Time
- **Full suite**: ~15-20 minutes (444 tests across 6 browsers)
- **Single browser**: ~3-5 minutes
- **Specific test category**: ~1-2 minutes

### Resource Requirements
- **RAM**: 4GB minimum, 8GB recommended
- **CPU**: 2 cores minimum, 4 cores recommended  
- **Disk**: 2GB free space for Docker images and test artifacts

### Optimization Tips
```bash
# Run tests in parallel (default)
npm test

# Run on single browser for faster feedback
npx playwright test --project=chromium

# Run without video recording for speed
npx playwright test --config=playwright.config.minimal.ts
```

## Support

If you encounter issues not covered in this guide:

1. Check the [main README](./README.md) for general setup
2. Review test logs in `testing/e2e/test-results/`
3. Check service logs in `logs/` directory
4. Open an issue with full error output and system information

## Contributing

When adding new tests:

1. Follow existing test patterns in `testing/e2e/tests/`
2. Use `data-testid` attributes for reliable element selection
3. Add test documentation to this guide
4. Ensure tests clean up after themselves