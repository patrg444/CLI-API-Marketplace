# API-Direct Testing Guide

This document provides comprehensive information about testing the API-Direct platform.

## Overview

The API-Direct platform uses different testing frameworks and approaches for different components:

- **Backend**: Python with pytest and asyncio
- **Frontend**: JavaScript with Jest and React Testing Library
- **End-to-End**: Integration tests spanning multiple components

## Backend Testing

### Setup

```bash
cd backend
pip install -r requirements.txt
pip install pytest pytest-asyncio pytest-cov
```

### Running Tests

```bash
# Run all backend tests
python -m pytest

# Run specific test file
python -m pytest api/tests/test_auth.py

# Run with coverage
python -m pytest --cov=api --cov-report=html

# Run with verbose output
python -m pytest -v
```

### Test Structure

Backend tests are organized by functionality:

```
backend/api/tests/
├── test_auth.py           # Authentication tests
├── test_auth_mocked.py    # Auth tests with mocked database
├── test_api_keys.py       # API key management tests
├── test_websocket.py      # WebSocket functionality tests
├── test_websocket_simple.py # Simplified WebSocket tests
├── test_database.py       # Database operations tests
└── test_endpoints.py      # API endpoint tests
```

### Key Test Categories

#### 1. Authentication Tests (`test_auth.py`, `test_auth_mocked.py`)

Tests user registration, login, JWT tokens, and password reset functionality.

```python
# Example test
def test_register_new_user(self, mock_db):
    response = client.post("/auth/register", json={
        "email": "test@example.com",
        "password": "SecurePassword123!",
        "name": "Test User",
        "company": "Test Company"
    })
    assert response.status_code == 200
    assert 'access_token' in response.json()
```

#### 2. API Key Management Tests (`test_api_keys.py`)

Tests CLI authentication via API keys.

```python
# Create API key
response = client.post("/api/keys", json={
    "name": "Test CLI Key",
    "scopes": ["read", "write"],
    "expires_in_days": 30
})

# Use API key
response = client.get("/api/keys/test", headers={"X-API-Key": api_key})
```

#### 3. WebSocket Tests (`test_websocket.py`, `test_websocket_simple.py`)

Tests real-time communication features.

```python
async def test_connect_user(self, manager, mock_websocket):
    await manager.connect(mock_websocket, user_id)
    assert user_id in manager.active_connections
```

#### 4. Database Tests (`test_database.py`)

Tests database operations and models.

```python
async def test_create_user(self, db_manager):
    user = await db_manager.create_user(
        email="test@example.com",
        password="password123",
        name="Test User"
    )
    assert user.email == "test@example.com"
```

### Mocking Strategy

The backend uses mocking extensively to avoid database dependencies during testing:

```python
# Mock database pool
with patch('asyncpg.create_pool', new_callable=AsyncMock) as mock_pool:
    mock_conn = AsyncMock()
    mock_pool.return_value.acquire.return_value.__aenter__.return_value = mock_conn
```

## Frontend Testing

### Setup

```bash
cd web/marketplace
npm install
```

### Running Tests

```bash
# Run all frontend tests
npm test

# Run with coverage
npm test -- --coverage

# Run in watch mode
npm test -- --watch

# Run specific test file
npm test APICard.test.js
```

### Test Structure

Frontend tests follow the component structure:

```
web/marketplace/src/
├── components/
│   ├── __tests__/
│   │   ├── APICard.test.js
│   │   ├── Navigation.test.js
│   │   └── Dashboard.test.js
│   └── ...
├── services/
│   └── __tests__/
│       └── api.test.js
└── utils/
    └── __tests__/
        └── formatters.test.js
```

### Key Test Categories

#### 1. Component Tests

Test React components in isolation:

```javascript
describe('APICard', () => {
  it('renders API information correctly', () => {
    const api = {
      name: 'Weather API',
      description: 'Get weather data',
      endpoint: 'https://api.example.com/weather'
    };
    
    render(<APICard api={api} />);
    expect(screen.getByText('Weather API')).toBeInTheDocument();
  });
});
```

#### 2. Integration Tests

Test component interactions:

```javascript
it('navigates to API details on click', async () => {
  const { getByRole } = render(
    <BrowserRouter>
      <APICard api={mockApi} />
    </BrowserRouter>
  );
  
  fireEvent.click(getByRole('link'));
  expect(mockNavigate).toHaveBeenCalledWith('/api/123');
});
```

#### 3. Service Tests

Test API calls and data fetching:

```javascript
describe('API Service', () => {
  it('fetches user APIs', async () => {
    const apis = await apiService.getUserAPIs();
    expect(apis).toHaveLength(2);
    expect(apis[0]).toHaveProperty('name');
  });
});
```

## Database Testing

### Database Setup

For testing, we support both PostgreSQL and SQLite:

```python
# PostgreSQL (production-like)
DATABASE_URL = "postgresql://user:pass@localhost/apidirect_test"

# SQLite (fast, in-memory)
DATABASE_URL = "sqlite+aiosqlite:///:memory:"
```

### Schema Initialization

```bash
# Initialize test database
python database/init_db.py

# Run with test data
ENVIRONMENT=development python database/init_db.py
```

## Test Data

### Backend Test Data

Located in `backend/api/tests/fixtures/`:

- `test_users.json` - Sample user data
- `test_apis.json` - Sample API configurations
- `test_transactions.json` - Sample billing data

### Frontend Test Data

Mock data is defined inline in test files:

```javascript
const mockApi = {
  id: '123',
  name: 'Test API',
  status: 'running',
  calls: 1000
};
```

## Continuous Integration

### GitHub Actions Workflow

```yaml
name: Tests
on: [push, pull_request]

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Python
        uses: actions/setup-python@v2
      - name: Install dependencies
        run: |
          cd backend
          pip install -r requirements.txt
          pip install pytest pytest-asyncio
      - name: Run tests
        run: |
          cd backend
          python -m pytest

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup Node.js
        uses: actions/setup-node@v2
      - name: Install dependencies
        run: |
          cd web/marketplace
          npm install
      - name: Run tests
        run: |
          cd web/marketplace
          npm test
```

## Best Practices

### 1. Test Isolation

Each test should be independent and not rely on other tests:

```python
def setup_method(self):
    """Reset state before each test"""
    self.test_user_id = str(uuid.uuid4())
    self.access_token = create_access_token(self.test_user_id)
```

### 2. Use Fixtures

Reuse common test setup:

```python
@pytest.fixture
async def db_connection():
    conn = await create_test_connection()
    yield conn
    await conn.close()
```

### 3. Test Edge Cases

Always test error conditions:

```python
def test_login_invalid_credentials(self):
    response = client.post("/auth/login", json={
        "email": "wrong@example.com",
        "password": "wrongpassword"
    })
    assert response.status_code == 401
```

### 4. Mock External Services

Never call real external services in tests:

```python
@patch('stripe.Charge.create')
def test_process_payment(self, mock_charge):
    mock_charge.return_value = {'id': 'ch_test123'}
    # Test payment processing
```

## Troubleshooting

### Common Issues

1. **Import Errors**
   ```python
   # Add parent directory to path
   sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
   ```

2. **Async Test Issues**
   ```python
   # Use pytest.mark.asyncio
   @pytest.mark.asyncio
   async def test_async_function():
       result = await async_operation()
   ```

3. **Database Connection Issues**
   ```python
   # Use mocked connections for unit tests
   mock_conn = AsyncMock()
   mock_db.acquire.return_value.__aenter__.return_value = mock_conn
   ```

### Debug Mode

Run tests with debugging output:

```bash
# Python
python -m pytest -v -s --log-cli-level=DEBUG

# JavaScript
DEBUG=* npm test
```

## Coverage Goals

We aim for the following test coverage:

- **Backend API**: >80% coverage
- **Frontend Components**: >70% coverage
- **Critical Paths**: 100% coverage (auth, payments, deployments)

Check current coverage:

```bash
# Backend
python -m pytest --cov=api --cov-report=term-missing

# Frontend
npm test -- --coverage --watchAll=false
```

## Test Environment Variables

Create a `.env.test` file for test-specific configuration:

```bash
# Backend
DATABASE_URL=postgresql://localhost/apidirect_test
REDIS_URL=redis://localhost:6379/1
JWT_SECRET=test-secret-key
STRIPE_SECRET_KEY=sk_test_...

# Frontend
REACT_APP_API_URL=http://localhost:8000
REACT_APP_WS_URL=ws://localhost:8000
```

## Contributing

When adding new features:

1. Write tests first (TDD approach)
2. Ensure all tests pass
3. Add integration tests for complex features
4. Update this documentation

## Resources

- [pytest documentation](https://docs.pytest.org/)
- [Jest documentation](https://jestjs.io/)
- [React Testing Library](https://testing-library.com/react)
- [FastAPI testing guide](https://fastapi.tiangolo.com/tutorial/testing/)