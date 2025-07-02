const express = require('express');
const cors = require('cors');
const app = express();

app.use(cors());
app.use(express.json());

// Mock data
const mockAPIs = [
  {
    id: '1',
    name: 'Weather API',
    description: 'Get real-time weather data for any location worldwide',
    category: 'Weather',
    pricing: { model: 'usage', price: 0.001, unit: 'request' },
    rating: 4.8,
    reviews: 156,
    creator: { name: 'WeatherTech Inc', verified: true },
    endpoints: 12,
    monthlyUsers: 5420,
    icon: '🌤️'
  },
  {
    id: '2',
    name: 'Translation API',
    description: 'Translate text between 100+ languages with AI',
    category: 'AI/ML',
    pricing: { model: 'usage', price: 0.002, unit: 'character' },
    rating: 4.9,
    reviews: 89,
    creator: { name: 'LinguaAI', verified: true },
    endpoints: 8,
    monthlyUsers: 3200,
    icon: '🌐'
  },
  {
    id: '3',
    name: 'Stock Market Data',
    description: 'Real-time and historical stock market data',
    category: 'Finance',
    pricing: { model: 'monthly', price: 49.99, unit: 'month' },
    rating: 4.7,
    reviews: 234,
    creator: { name: 'FinanceHub', verified: true },
    endpoints: 24,
    monthlyUsers: 8900,
    icon: '📈'
  },
  {
    id: '4',
    name: 'Email Validation API',
    description: 'Validate and verify email addresses in real-time',
    category: 'Utilities',
    pricing: { model: 'usage', price: 0.0005, unit: 'validation' },
    rating: 4.6,
    reviews: 412,
    creator: { name: 'EmailGuard', verified: false },
    endpoints: 4,
    monthlyUsers: 12300,
    icon: '✉️'
  },
  {
    id: '5',
    name: 'Image Recognition API',
    description: 'Identify objects, faces, and text in images using AI',
    category: 'AI/ML',
    pricing: { model: 'usage', price: 0.005, unit: 'image' },
    rating: 4.9,
    reviews: 178,
    creator: { name: 'VisionAI Labs', verified: true },
    endpoints: 10,
    monthlyUsers: 6700,
    icon: '🖼️'
  },
  {
    id: '6',
    name: 'SMS Gateway API',
    description: 'Send SMS messages worldwide with delivery tracking',
    category: 'Communication',
    pricing: { model: 'usage', price: 0.05, unit: 'message' },
    rating: 4.5,
    reviews: 567,
    creator: { name: 'GlobalSMS', verified: true },
    endpoints: 6,
    monthlyUsers: 15600,
    icon: '📱'
  }
];

const mockUser = {
  id: 'user123',
  name: 'Demo User',
  email: 'demo@apidirect.com',
  role: 'consumer',
  apiKeys: [
    { id: '1', name: 'Production Key', key: 'api_key_demo_12345', created: '2024-01-15' },
    { id: '2', name: 'Test Key', key: 'api_key_test_67890', created: '2024-02-20' }
  ],
  subscriptions: [mockAPIs[0], mockAPIs[2]]
};

// Routes
app.get('/api/health', (req, res) => {
  res.json({ status: 'healthy', timestamp: new Date() });
});

app.get('/api/marketplace/apis', (req, res) => {
  const { category, search, sort } = req.query;
  let filtered = [...mockAPIs];
  
  if (category && category !== 'all') {
    filtered = filtered.filter(api => api.category.toLowerCase() === category.toLowerCase());
  }
  
  if (search) {
    filtered = filtered.filter(api => 
      api.name.toLowerCase().includes(search.toLowerCase()) ||
      api.description.toLowerCase().includes(search.toLowerCase())
    );
  }
  
  if (sort === 'rating') {
    filtered.sort((a, b) => b.rating - a.rating);
  } else if (sort === 'popular') {
    filtered.sort((a, b) => b.monthlyUsers - a.monthlyUsers);
  }
  
  res.json({ apis: filtered, total: filtered.length });
});

app.get('/api/marketplace/apis/:id', (req, res) => {
  const api = mockAPIs.find(a => a.id === req.params.id);
  if (api) {
    res.json({
      ...api,
      documentation: {
        baseUrl: `https://api.${api.creator.name.toLowerCase().replace(' ', '')}.com/v1`,
        authentication: 'API Key',
        rateLimit: '1000 requests/hour',
        sdks: ['Python', 'JavaScript', 'Java', 'Go', 'Ruby']
      }
    });
  } else {
    res.status(404).json({ error: 'API not found' });
  }
});

app.get('/api/marketplace/categories', (req, res) => {
  const categories = [...new Set(mockAPIs.map(api => api.category))];
  res.json({ categories });
});

app.post('/api/auth/login', (req, res) => {
  // Mock login - accept any credentials
  res.json({
    user: mockUser,
    token: 'mock_jwt_token_' + Date.now()
  });
});

app.post('/api/auth/register', (req, res) => {
  res.json({
    user: { ...mockUser, email: req.body.email, name: req.body.name },
    token: 'mock_jwt_token_' + Date.now()
  });
});

app.get('/api/user/profile', (req, res) => {
  res.json({ user: mockUser });
});

app.get('/api/user/subscriptions', (req, res) => {
  res.json({ subscriptions: mockUser.subscriptions });
});

app.get('/api/user/usage', (req, res) => {
  res.json({
    usage: [
      { date: '2024-06-01', calls: 1250, cost: 1.25 },
      { date: '2024-06-02', calls: 980, cost: 0.98 },
      { date: '2024-06-03', calls: 1450, cost: 1.45 },
      { date: '2024-06-04', calls: 1100, cost: 1.10 },
      { date: '2024-06-05', calls: 1320, cost: 1.32 }
    ],
    total: { calls: 6100, cost: 6.10 }
  });
});

// Start server
const PORT = 8000;
app.listen(PORT, () => {
  console.log(`Mock backend server running on http://localhost:${PORT}`);
  console.log('Available endpoints:');
  console.log('- GET  /api/health');
  console.log('- GET  /api/marketplace/apis');
  console.log('- GET  /api/marketplace/apis/:id');
  console.log('- GET  /api/marketplace/categories');
  console.log('- POST /api/auth/login');
  console.log('- POST /api/auth/register');
  console.log('- GET  /api/user/profile');
  console.log('- GET  /api/user/subscriptions');
  console.log('- GET  /api/user/usage');
});