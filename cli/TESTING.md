# CLI Marketplace Commands Testing Guide

## Overview

This document describes the testing approach for the new marketplace commands added to the API Direct CLI.

## Test Structure

### Unit Tests
Each command has its own test file:
- `analytics_test.go` - Tests for analytics commands
- `earnings_test.go` - Tests for earnings and payout commands  
- `subscriptions_test.go` - Tests for subscription management
- `review_test.go` - Tests for review system commands
- `search_test.go` - Tests for marketplace search and browse

### Test Utilities
- `test_utils.go` - Shared test utilities and mocks
- Mock HTTP client for simulating API responses
- Input mocking for interactive command flows

## Running Tests

### All Tests
```bash
./test.sh
```

### Specific Command Tests
```bash
# Analytics tests only
go test ./cmd -run TestAnalytics -v

# Earnings tests only  
go test ./cmd -run TestEarnings -v

# Subscriptions tests only
go test ./cmd -run TestSubscriptions -v

# Review tests only
go test ./cmd -run TestReview -v

# Search tests only
go test ./cmd -run TestSearch -v
```

### With Coverage
```bash
go test ./cmd -v -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Race Condition Detection
```bash
go test ./cmd -race
```

## Test Cases

### Analytics Command Tests
- ✅ Usage analytics for all APIs
- ✅ Usage analytics for specific API
- ✅ Revenue analytics with breakdown
- ✅ Consumer analytics
- ✅ Performance metrics
- ✅ Different output formats (table, JSON, CSV)
- ✅ Custom time periods
- ✅ Error handling

### Earnings Command Tests
- ✅ Earnings summary display
- ✅ Detailed earnings breakdown
- ✅ Payout requests (full and partial)
- ✅ Payout history
- ✅ Payout method setup
- ✅ Validation (minimum amounts, balance checks)
- ✅ Interactive confirmation flows
- ✅ Multiple output formats

### Subscriptions Command Tests
- ✅ List all subscriptions
- ✅ Filter by status
- ✅ Show subscription details
- ✅ Cancel subscription flow
- ✅ Usage tracking
- ✅ API key management
- ✅ Key regeneration with confirmation
- ✅ Usage examples display

### Review Command Tests
- ✅ Submit review (with flags and interactive)
- ✅ List reviews with filtering
- ✅ View my reviews
- ✅ Respond to reviews (creators)
- ✅ Report inappropriate reviews
- ✅ Review statistics
- ✅ Rating validation
- ✅ Star rating display

### Search Command Tests
- ✅ Search with query
- ✅ Filter by category, tags, price
- ✅ Different sort orders
- ✅ Browse categories
- ✅ Trending APIs
- ✅ Featured APIs
- ✅ Grid and table output formats
- ✅ Pagination support

## Mock Patterns

### HTTP Client Mock
```go
type mockHTTPClient struct {
    responses map[string]mockResponse
}

type mockResponse struct {
    statusCode int
    body       interface{}
    err        error
}
```

### Interactive Input Mock
```go
// Mock user input for confirmations
stdin = strings.NewReader("y\n")
defer func() { stdin = oldStdin }()
```

## Test Data

### Sample API Response
```json
{
  "period": {
    "start": "2024-01-01",
    "end": "2024-01-31"
  },
  "total_calls": 15000,
  "unique_consumers": 250,
  "apis": [
    {
      "name": "weather-api",
      "calls": 10000,
      "consumers": 150,
      "error_rate": 0.02
    }
  ]
}
```

## Coverage Goals

- Target: 70%+ code coverage
- Critical paths: 90%+ coverage
- Error handling: 100% coverage

## Integration Testing

While these tests focus on unit testing with mocks, integration tests should be run against a test API server:

```bash
# Set test environment
export APIDIRECT_API_ENDPOINT=https://test.api-direct.io
export APIDIRECT_TEST_MODE=true

# Run integration tests
go test ./cmd -tags=integration
```

## Continuous Integration

The test suite is designed to run in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run CLI Tests
  run: |
    cd cli
    ./test.sh
```

## Known Limitations

1. Mock HTTP client doesn't test actual network conditions
2. Interactive flows are simplified in tests
3. File system operations are not fully mocked
4. Some edge cases around concurrent operations

## Future Improvements

1. Add integration tests with real API
2. Implement property-based testing for complex inputs
3. Add performance benchmarks
4. Create end-to-end test scenarios
5. Add mutation testing