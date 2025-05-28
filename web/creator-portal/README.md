# API-Direct Creator Portal

Web interface for API creators to manage their deployed APIs, view analytics, and configure marketplace listings.

## Features

- **Dashboard**: Overview of API performance, revenue, and activity
- **API Management**: View, start/stop, and manage deployed APIs
- **Marketplace Integration**: Configure pricing and publish APIs to marketplace
- **Analytics**: Detailed usage statistics and revenue tracking
- **Billing**: Track earnings and manage payouts

## Setup

1. **Install dependencies**:
```bash
npm install
```

2. **Configure environment**:
Create a `.env` file with:
```env
REACT_APP_AWS_REGION=us-east-1
REACT_APP_USER_POOL_ID=<from terraform output>
REACT_APP_USER_POOL_CLIENT_ID=<from terraform output>
REACT_APP_AUTH_DOMAIN=<from terraform output>
REACT_APP_API_ENDPOINT=<ALB endpoint>
```

3. **Run development server**:
```bash
npm start
```

4. **Build for production**:
```bash
npm run build
```

## Deployment

The portal can be deployed to:
- AWS S3 + CloudFront
- AWS Amplify
- Any static hosting service

## Architecture

- **Frontend**: React with Material-UI
- **Authentication**: AWS Cognito (same pool as CLI)
- **State Management**: React Context API
- **API Communication**: Axios with JWT tokens

## Development

### Project Structure
```
src/
├── components/      # Reusable UI components
├── pages/          # Route pages
├── services/       # API service layer
├── utils/          # Helper functions
└── App.js          # Main app component
```

### Adding New Features

1. Create page component in `src/pages/`
2. Add route in `App.js`
3. Add navigation item in `Layout.js`
4. Create API service in `src/services/`

## API Integration

The portal communicates with the backend services:
- Storage Service: Code version management
- Deployment Service: API lifecycle management
- Gateway Service: Usage metrics and logs
- Billing Service: Revenue and payout tracking
