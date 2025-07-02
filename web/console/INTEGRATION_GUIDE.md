# Console Dashboard Integration Guide

## Overview
The console dashboard provides a visual interface for API creators to:
- View earnings and API metrics
- Manage API deployments
- Configure API settings
- Track usage analytics
- Handle payouts

## Current Status
✅ **Frontend**: Complete HTML/CSS/JS dashboard
✅ **API Client**: Updated to match backend endpoints
❌ **Backend Connection**: Needs to be connected to running backend

## Quick Start

### 1. Start the Backend
```bash
cd ../../backend/api
python main.py
```

### 2. Start the Console
```bash
cd web/console
python app.py
# OR
./run-console.sh
```

### 3. Access Console
Open http://localhost:5000 in your browser

## File Structure
```
console/
├── api-client-updated.js   # API client matching backend endpoints
├── dashboard-updated.js    # Dashboard functionality with real API calls
├── app.py                  # Flask server for console
├── pages/                  # HTML pages
│   ├── dashboard.html      # Main dashboard
│   ├── apis.html          # API management
│   ├── analytics.html     # Analytics view
│   ├── earnings.html      # Earnings & payouts
│   └── marketplace.html   # Marketplace browser
└── login.html             # Authentication page
```

## Integration Points

### Authentication
- Uses JWT tokens from backend `/api/auth/login`
- Token stored in localStorage
- Auto-redirect to login if not authenticated

### Real-time Updates
- WebSocket connection at `ws://localhost:8000/ws`
- Receives deployment updates, metrics, API calls
- Auto-reconnects on disconnect

### API Endpoints Used
```javascript
// Auth
POST   /api/auth/login
POST   /api/auth/register
GET    /api/auth/me

// Dashboard
GET    /api/dashboard/overview
GET    /api/dashboard/recent-deployments

// APIs
GET    /api/my-apis
POST   /api/deploy
PUT    /api/my-apis/{id}
DELETE /api/my-apis/{id}

// Analytics
GET    /api/analytics/usage-by-consumer
GET    /api/analytics/geographic
GET    /api/analytics/errors
GET    /api/analytics/revenue

// API Keys
GET    /api/keys
POST   /api/keys
DELETE /api/keys/{id}

// Marketplace
GET    /api/marketplace/listings
POST   /api/marketplace/publish

// Trials
POST   /api/trials/start
GET    /api/trials/{id}/status

// Payments
GET    /api/subscription/plans
POST   /api/subscription
POST   /api/payouts/request
```

## Next Steps for Full Integration

1. **Update HTML Pages**: Add script includes for api-client-updated.js
2. **Environment Config**: Set API_BASE_URL based on environment
3. **Error Handling**: Add user-friendly error messages
4. **Loading States**: Show spinners during API calls
5. **Data Refresh**: Add pull-to-refresh or auto-refresh
6. **Mobile Responsive**: Test and fix mobile layouts

## Testing
1. Create test account via `/api/auth/register`
2. Deploy test API via CLI or `/api/deploy`
3. Generate some API calls to see metrics
4. Check all dashboard pages load correctly

## Production Deployment
1. Build static assets (minify JS/CSS)
2. Configure HTTPS/SSL
3. Set production API endpoints
4. Enable CORS for production domain
5. Set up CDN for static assets