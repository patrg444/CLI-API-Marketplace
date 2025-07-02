# 🧪 CLI Test Coverage Plan

## Current Status
- **Commands**: 28 total
- **Tested**: 8 (28.6%)
- **Untested**: 20 (71.4%)

## Priority Testing Needs

### 🔴 Critical Priority (Core User Journey)

#### 1. **Authentication** (`auth.go`)
```bash
apidirect login
apidirect logout
```
- [ ] Mock browser-based authentication flow
- [ ] Token storage and retrieval
- [ ] Credential persistence
- [ ] Error handling (network issues, invalid tokens)
- [ ] Multi-environment support

#### 2. **Project Initialization** (`init.go`)
```bash
apidirect init
apidirect init --template fastapi
```
- [ ] Interactive wizard flow
- [ ] Template selection (FastAPI, Express, Go, Rails)
- [ ] Project structure creation
- [ ] Git initialization
- [ ] Dependency file generation

#### 3. **Local Development** (`run.go`)
```bash
apidirect run
apidirect run --port 3000
```
- [ ] Runtime detection (Python, Node.js, Go, Ruby)
- [ ] Process management
- [ ] Live reload
- [ ] Environment variable loading
- [ ] Error handling (port conflicts, missing deps)

#### 4. **API Import** (`import.go`)
```bash
apidirect import
apidirect import ./my-api
```
- [ ] Framework detection
- [ ] Manifest generation
- [ ] Endpoint discovery
- [ ] Environment variable detection
- [ ] Validation of imported projects

#### 5. **Configuration Validation** (`validate.go`)
```bash
apidirect validate
```
- [ ] Manifest validation
- [ ] Schema checking
- [ ] Dependency verification
- [ ] Port availability
- [ ] Error reporting

### 🟡 High Priority (Deployment & Publishing)

#### 6. **Deployment** (`deploy_v2.go`)
Already partially tested in E2E, but needs unit tests:
- [ ] Manifest loading
- [ ] Build context creation
- [ ] Error scenarios
- [ ] Rollback handling
- [ ] Multi-environment deployment

#### 7. **Status Checking** (`status.go`)
```bash
apidirect status my-api
apidirect status --watch
```
- [ ] Status retrieval for both modes
- [ ] Watch mode
- [ ] JSON output
- [ ] Error handling

#### 8. **Logs** (`logs_v2.go`)
```bash
apidirect logs my-api
apidirect logs my-api --tail 100
```
- [ ] Log streaming
- [ ] Filtering options
- [ ] Both deployment modes
- [ ] Error handling

#### 9. **Publishing** (`publish.go`)
```bash
apidirect publish my-api
apidirect unpublish my-api
```
- [ ] Publishing workflow
- [ ] Validation checks
- [ ] Pricing configuration
- [ ] Unpublish functionality

#### 10. **Destroy/Cleanup** (`destroy.go`)
```bash
apidirect destroy my-api
```
- [ ] Safe destruction workflow
- [ ] Confirmation prompts
- [ ] BYOA infrastructure cleanup
- [ ] Hosted deployment removal

### 🟢 Medium Priority (Management & Tools)

#### 11. **Environment Management** (`env.go`)
```bash
apidirect env set KEY=value
apidirect env list
```
- [ ] Environment variable management
- [ ] Secret handling
- [ ] Multi-environment support

#### 12. **Scaling** (`scale.go`)
```bash
apidirect scale my-api --replicas 3
```
- [ ] Scaling operations
- [ ] Auto-scaling configuration
- [ ] Both deployment modes

#### 13. **Marketplace Operations** (`marketplace.go`)
```bash
apidirect marketplace list
apidirect marketplace featured
```
- [ ] Listing APIs
- [ ] Filtering and search
- [ ] Category browsing

#### 14. **Subscriptions** (`subscribe.go`)
```bash
apidirect subscribe api-name
```
- [ ] Subscription workflow
- [ ] Payment integration
- [ ] API key generation

### 🔵 Low Priority (Supporting Features)

#### 15. **API Info** (`info.go`)
- [ ] Information display
- [ ] Formatting options

#### 16. **Pricing Management** (`pricing.go`)
- [ ] Pricing configuration
- [ ] Update workflows

#### 17. **Self Update** (`self_update.go`)
- [ ] Version checking
- [ ] Update process
- [ ] Rollback capability

## Test Implementation Strategy

### 1. **Unit Tests** (Priority 1)
For each command, create `<command>_test.go` with:
- Command parsing tests
- Flag validation tests
- Error handling tests
- Mock service interactions

### 2. **Integration Tests** (Priority 2)
Test command workflows:
- Init → Import → Validate → Deploy
- Login → Deploy → Status → Logs
- Deploy → Publish → Subscribe

### 3. **E2E Tests** (Priority 3)
Complete user journeys:
- New user onboarding
- API development workflow
- Marketplace publishing
- Subscription management

## Test Structure

```go
// Example: auth_test.go
func TestAuthCommand(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        wantErr  bool
        setup    func()
        validate func(t *testing.T)
    }{
        {
            name: "successful login",
            args: []string{"login"},
            setup: mockBrowserAuth,
            validate: checkTokenStored,
        },
        {
            name: "login with invalid token",
            args: []string{"login"},
            wantErr: true,
            setup: mockInvalidAuth,
        },
    }
}
```

## Coverage Goals

### Phase 1 (Immediate)
- Critical commands: 80% coverage
- Core user journey: E2E tests
- Error scenarios: Comprehensive

### Phase 2 (Short-term)
- All commands: 60% coverage
- Integration tests: Key workflows
- Mock services: Complete

### Phase 3 (Long-term)
- Overall: 80%+ coverage
- Performance tests
- Load testing for concurrent operations

## Next Steps

1. **Start with auth_test.go** - Foundation for all operations
2. **Create test helpers** - Mock services, test fixtures
3. **Implement init_test.go** - Critical entry point
4. **Add run_test.go** - Development workflow
5. **Complete deployment tests** - Build on existing E2E

## Test Execution

```bash
# Run all CLI tests
go test ./cmd/...

# Run specific command tests
go test ./cmd -run TestAuth

# Run with coverage
go test ./cmd/... -cover

# Generate coverage report
go test ./cmd/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```