# Local Development Setup Guide

## 🚀 Quick Start

### Prerequisites
- Docker Desktop installed and running
- Python 3.11+
- Node.js 18+ (for frontend development)

### 1. Start Backend Services

```bash
# Make the script executable
chmod +x start-local-backend.sh

# Start PostgreSQL, Redis, and FastAPI backend
./start-local-backend.sh
```

This will:
- Start PostgreSQL on port 5432
- Start Redis on port 6379
- Start FastAPI backend on http://localhost:8000
- API Documentation available at http://localhost:8000/docs

### 2. Access the Applications

#### Backend API
- **URL**: http://localhost:8000
- **API Docs**: http://localhost:8000/docs
- **Mock Login Credentials**:
  - Email: `demo@apidirect.dev`
  - Password: `secret`

#### Console Dashboard
- **URL**: https://console.apidirect.dev
- Use the same login credentials above
- The dashboard will connect to your local backend

#### Marketplace
- **URL**: https://marketplace.apidirect.dev
- Currently shows static data (needs backend integration)

### 3. Alternative: Run Everything with Docker

```bash
# Start all services including backend
docker-compose -f docker-compose.local.yml up

# Or run in background
docker-compose -f docker-compose.local.yml up -d
```

### 4. Development Workflow

1. **Backend Development**:
   - Edit files in `/backend/api/`
   - FastAPI will auto-reload on changes
   - Check logs: `docker-compose -f docker-compose.local.yml logs -f backend`

2. **Frontend Development**:
   - Console: Edit files in `/web/console/`
   - Changes pushed to GitHub auto-deploy via Vercel
   - Or run locally: `cd web/console && python -m http.server 3000`

3. **Database Access**:
   ```bash
   # Connect to PostgreSQL
   psql -h localhost -U apidirect -d apidirect
   # Password: localpassword
   ```

### 5. Testing the Integration

1. Open https://console.apidirect.dev/login.html
2. Login with demo credentials
3. You should see:
   - Dashboard stats loading from local backend
   - Real-time updates via WebSocket
   - Mock data for APIs and billing

### 6. Troubleshooting

**Backend won't start?**
```bash
# Check if ports are in use
lsof -i :8000
lsof -i :5432
lsof -i :6379

# Stop services and restart
docker-compose -f docker-compose.local.yml down
./start-local-backend.sh
```

**CORS errors?**
- Make sure your backend has the correct CORS origins set
- Check that USE_MOCK_AUTH=true is set

**Can't login?**
- Verify backend is running: http://localhost:8000/health
- Check browser console for errors
- Try incognito mode to avoid cache issues

## 🔧 Configuration

### Environment Variables
Create a `.env` file in the backend directory:

```env
DATABASE_URL=postgresql://apidirect:localpassword@localhost:5432/apidirect
REDIS_URL=redis://localhost:6379
JWT_SECRET=local-development-secret
USE_MOCK_AUTH=true
CORS_ORIGINS=http://localhost:3000,https://console.apidirect.dev,https://marketplace.apidirect.dev
```

### Mock Authentication
The local setup uses mock authentication to avoid AWS dependencies:
- Pre-configured demo user
- JWT tokens for session management
- No external service dependencies

## 📝 Next Steps

1. **Connect Marketplace to Backend**
   - Update marketplace to use the API client
   - Implement search and filtering
   - Add subscription flow

2. **Enhance Console Features**
   - Add create/edit API functionality
   - Implement real analytics charts
   - Build out billing pages

3. **CLI Integration**
   - Update CLI to work with local backend
   - Test deployment commands
   - Implement template system

## 🎯 Development Tips

- Use the API docs at http://localhost:8000/docs to explore endpoints
- Check the browser console for API requests/responses
- Monitor the backend logs for errors
- The WebSocket connection provides real-time updates every 5 seconds

Happy coding! 🚀