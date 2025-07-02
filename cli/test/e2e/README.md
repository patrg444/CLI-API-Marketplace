# End-to-End Testing for API-Direct CLI

This directory contains comprehensive end-to-end tests for the API-Direct CLI, covering both **Hosted** and **BYOA** (Bring Your Own AWS) deployment modes.

## Test Structure

### Core Test Files

#### 1. **hosted_deployment_test.go** 🆕
Complete hosted deployment mode testing:
- Full hosted deployment lifecycle
- Mock backend services for API-Direct platform
- Concurrent deployment testing
- Mode switching scenarios

#### 2. **byoa_test.go**
Complete BYOA deployment lifecycle testing:
- Full deployment flow (create → deploy → status → destroy)
- Deployment validation
- Error handling
- AWS credential verification

#### 3. **deployment_modes_test.go** 🆕
Comprehensive comparison of both deployment modes:
- Tests both modes with identical scenarios
- Mode transition testing
- Feature validation for each mode
- Documentation accuracy verification

#### 4. **mock_aws_test.go**
Mock AWS services for testing without real AWS resources:
- Mock STS for identity verification
- Mock ECS for service status
- Mock CloudFormation for stack management
- Terraform execution mocking

#### 5. **integration_test.go**
Integration testing for CLI commands:
- Command sequencing
- Error scenarios
- Configuration management
- Prerequisites validation

#### 6. **test_fixtures.go**
Sample projects and manifests for testing:
- FastAPI (simple and complex)
- Express.js
- Go Gin
- Ruby on Rails
- Various manifest configurations

## 📊 Test Coverage

For detailed test coverage information, see [DEPLOYMENT_MODES_TEST_COVERAGE.md](./DEPLOYMENT_MODES_TEST_COVERAGE.md).

| Feature | Hosted | BYOA | Coverage |
|---------|--------|------|----------|
| Deployment | ✅ | ✅ | 100% |
| Status | ✅ | ✅ | 100% |
| Logs | ✅ | ✅ | 100% |
| Scaling | ✅ | ✅ | 100% |
| Updates | ✅ | ✅ | 100% |
| Mode Switch | ✅ | ✅ | 100% |

## Running Tests

### Prerequisites
```bash
# Install required tools
brew install awscli terraform

# Install Go dependencies
cd cli
go mod download
```

### Run All Tests
```bash
# Run all E2E tests
go test ./test/e2e/... -v

# Run with real AWS (requires credentials)
RUN_INTEGRATION_TESTS=true go test ./test/e2e/... -v

# Run only mock tests
go test ./test/e2e/... -run Mock -v

# Skip E2E tests
SKIP_E2E_TESTS=true go test ./test/e2e/...
```

### Test Specific Functionality
```bash
# Test BYOA deployment
go test ./test/e2e/... -run TestBYOADeploymentFlow -v

# Test error handling
go test ./test/e2e/... -run TestBYOADeploymentValidation -v

# Test with mock AWS
go test ./test/e2e/... -run TestBYOAWithMockAWS -v
```

## Test Configuration

### Environment Variables
```bash
# Skip E2E tests
export SKIP_E2E_TESTS=true

# Enable integration tests
export RUN_INTEGRATION_TESTS=true

# Use mock AWS services
export MOCK_AWS=true

# Test mode (no actual deployments)
export APIDIRECT_TEST_MODE=true

# Custom AWS endpoint (for mocking)
export AWS_ENDPOINT_URL=http://localhost:4566
```

### AWS Credentials
Tests can run with:
1. **Real AWS credentials** - Full deployment testing
2. **Mock credentials** - Using mock AWS services
3. **No credentials** - Skip AWS-dependent tests

## Test Scenarios

### 1. **Complete Deployment Flow**
```
1. Create API project
2. Import and validate
3. Deploy to AWS
4. Check deployment status
5. Test API endpoints
6. Destroy deployment
```

### 2. **Error Scenarios**
- Missing manifest
- Invalid manifest
- AWS credential errors
- Non-existent deployments
- Wrong AWS account

### 3. **Integration Tests**
- CLI command availability
- Configuration persistence
- Multi-language support
- Scaling configurations

## Adding New Tests

### 1. Create New Test File
```go
package e2e

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestNewFeature(t *testing.T) {
    // Test implementation
}
```

### 2. Add Test Fixtures
```go
func getNewFixture() *APIProject {
    return &APIProject{
        Name:    "new-api",
        Runtime: "python3.9",
        // ... fixture details
    }
}
```

### 3. Mock AWS Services
```go
func (m *MockAWSServices) handleNewService(w http.ResponseWriter, r *http.Request) {
    // Mock implementation
}
```

## Debugging Tests

### Enable Verbose Output
```bash
go test ./test/e2e/... -v -run TestName
```

### Check Test Logs
```go
t.Logf("Debug info: %v", variable)
```

### Skip Slow Tests
```go
if testing.Short() {
    t.Skip("Skipping slow test")
}
```

## CI/CD Integration

### GitHub Actions
```yaml
- name: Run E2E Tests
  env:
    SKIP_E2E_TESTS: ${{ secrets.SKIP_E2E_TESTS }}
    AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
    AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
  run: |
    cd cli
    go test ./test/e2e/... -v -timeout 30m
```

### Local Testing
```bash
# Quick tests only
go test ./test/e2e/... -short

# Full test suite
make test-e2e
```

## Test Coverage

Current test coverage includes:
- ✅ BYOA deployment lifecycle
- ✅ Error handling and validation
- ✅ Mock AWS services
- ✅ Multi-language project support
- ✅ Configuration management
- ✅ CLI command integration

## Known Issues

1. **AWS Credentials**: Tests may fail if AWS credentials are expired
2. **Terraform State**: Mock tests don't fully simulate Terraform state
3. **Network Dependencies**: Some tests require internet connectivity

## Contributing

When adding new E2E tests:
1. Follow existing test patterns
2. Add appropriate fixtures
3. Document test scenarios
4. Handle both real and mock AWS
5. Ensure tests are idempotent