// Marketplace Backend API for Vercel
// This provides real data for the marketplace frontend

const rateLimiter = require('./rate-limiter');

// Helper function to check rate limits
const checkRateLimit = (res, clientIp, endpoint = 'default') => {
  const limitCheck = rateLimiter.check(clientIp, endpoint);
  
  // Always set rate limit headers
  res.setHeader('X-RateLimit-Limit', limitCheck.limit);
  res.setHeader('X-RateLimit-Remaining', limitCheck.remaining);
  res.setHeader('X-RateLimit-Reset', new Date(limitCheck.resetAt).toISOString());
  
  if (!limitCheck.allowed) {
    res.setHeader('Retry-After', limitCheck.retryAfter);
    res.status(429).json({
      success: false,
      error: 'Too many requests. Please try again later.',
      retryAfter: limitCheck.retryAfter
    });
    return false;
  }
  
  return true;
};

const marketplaceData = {
  categories: [
    { id: 'ai-ml', name: 'AI/ML', icon: 'brain', count: 23 },
    { id: 'data', name: 'Data', icon: 'database', count: 18 },
    { id: 'finance', name: 'Finance', icon: 'credit-card', count: 12 },
    { id: 'weather', name: 'Weather', icon: 'cloud', count: 7 }
  ],
  
  apis: [
    {
      id: 'sentiment-analyzer-pro',
      name: 'sentiment-analyzer-pro',
      author: '@aidevlabs',
      description: 'Advanced emotion detection with support for 50+ languages and real-time processing',
      category: 'ai-ml',
      icon: 'brain',
      color: 'purple',
      rating: 4.8,
      reviews: 142,
      calls: 47000,
      pricing: {
        type: 'freemium',
        freeCalls: 1000,
        pricePerCall: 0.005,
        currency: 'USD'
      },
      tags: ['sentiment', 'emotion', 'nlp', 'ai', 'machine-learning'],
      featured: true,
      trending: true,
      growth: 342
    },
    {
      id: 'global-weather-api',
      name: 'global-weather-api',
      author: '@weathertech',
      description: 'Real-time weather data for 200K+ cities worldwide with 14-day forecasts',
      category: 'weather',
      icon: 'cloud-sun',
      color: 'blue',
      rating: 5.0,
      reviews: 89,
      calls: 125000,
      pricing: {
        type: 'freemium',
        freeCalls: 1000,
        pricePerCall: 0.001,
        currency: 'USD'
      },
      tags: ['weather', 'forecast', 'climate', 'real-time'],
      featured: true,
      trending: false
    },
    {
      id: 'gpt-4-turbo-wrapper',
      name: 'gpt-4-turbo-wrapper',
      author: '@openai-devs',
      description: 'Latest GPT-4 Turbo with vision capabilities and function calling',
      category: 'ai-ml',
      icon: 'robot',
      color: 'green',
      rating: 4.9,
      reviews: 256,
      calls: 89000,
      pricing: {
        type: 'freemium',
        freeCalls: 100,
        pricePerCall: 0.01,
        currency: 'USD'
      },
      tags: ['gpt', 'openai', 'ai', 'chatbot', 'vision'],
      featured: true,
      trending: true,
      growth: 218
    },
    {
      id: 'stock-market-predictor',
      name: 'stock-market-predictor',
      author: '@fintech-labs',
      description: 'AI-powered stock analysis and predictions with 85% accuracy',
      category: 'finance',
      icon: 'chart-line',
      color: 'green',
      rating: 4.6,
      reviews: 78,
      calls: 34000,
      pricing: {
        type: 'subscription',
        monthlyPrice: 99,
        calls: 10000,
        currency: 'USD'
      },
      tags: ['stocks', 'finance', 'prediction', 'ai', 'trading'],
      featured: false,
      trending: true,
      growth: 156
    },
    {
      id: 'translation-api',
      name: 'translation-api',
      author: '@linguatech',
      description: 'Real-time translation API supporting 100+ languages with context awareness',
      category: 'data',
      icon: 'language',
      color: 'green',
      rating: 4.7,
      reviews: 89,
      calls: 23000,
      pricing: {
        type: 'freemium',
        freeCalls: 500,
        pricePerCall: 0.002,
        currency: 'USD'
      },
      tags: ['translation', 'language', 'nlp', 'international'],
      featured: false,
      trending: false
    },
    {
      id: 'crypto-analytics-api',
      name: 'crypto-analytics-api',
      author: '@blockchain-insights',
      description: 'Real-time cryptocurrency market data and advanced analytics',
      category: 'finance',
      icon: 'bitcoin',
      color: 'yellow',
      rating: 4.9,
      reviews: 167,
      calls: 67000,
      pricing: {
        type: 'freemium',
        freeCalls: 100,
        pricePerCall: 0.008,
        currency: 'USD'
      },
      tags: ['crypto', 'blockchain', 'bitcoin', 'ethereum', 'defi'],
      featured: true,
      trending: true,
      growth: 425
    }
  ]
};

// CORS headers for Vercel
const headers = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
  'Access-Control-Allow-Headers': 'Content-Type',
  'Content-Type': 'application/json'
};

export default function handler(req, res) {
  try {
    // Handle CORS preflight
    if (req.method === 'OPTIONS') {
      res.status(200).setHeaders(headers).end();
      return;
    }

    // Set CORS headers
    Object.entries(headers).forEach(([key, value]) => {
      res.setHeader(key, value);
    });
    
    // Get client IP
    const clientIp = req.headers['x-forwarded-for'] || 
                    req.headers['x-real-ip'] || 
                    req.connection?.remoteAddress || 
                    'unknown';
    
    // Only allow GET requests for now
    if (req.method !== 'GET') {
      return res.status(405).json({
        success: false,
        error: 'Method not allowed'
      });
    }

    const { pathname, searchParams } = new URL(req.url, `http://${req.headers.host}`);
    // Remove /api prefix and trailing slashes
    const path = pathname.replace('/api', '').replace(/\/$/, '');

  // Routes
  switch (path) {
    case '/categories':
      // Apply rate limiting
      if (!checkRateLimit(res, clientIp, 'default')) {
        return;
      }
      
      return res.status(200).json({
        success: true,
        data: marketplaceData.categories
      });

    case '/apis':
      // Apply rate limiting for search endpoint
      const endpoint = searchParams.get('search') ? 'search' : 'default';
      if (!checkRateLimit(res, clientIp, endpoint)) {
        return;
      }
      
      let apis = [...marketplaceData.apis];
      
      // Filter by category
      const category = searchParams.get('category');
      if (category && category !== 'all') {
        apis = apis.filter(api => api.category === category);
      }

      // Filter by search query with sanitization
      const search = searchParams.get('search');
      if (search) {
        // Sanitize search query - remove special regex characters
        const query = search.toLowerCase().replace(/[.*+?^${}()|[\]\\]/g, '\\$&').trim();
        
        // Limit search query length
        if (query.length > 100) {
          return res.status(400).json({
            success: false,
            error: 'Search query too long (max 100 characters)'
          });
        }
        
        apis = apis.filter(api => 
          api.name.toLowerCase().includes(query) ||
          api.description.toLowerCase().includes(query) ||
          api.tags.some(tag => tag.toLowerCase().includes(query))
        );
      }

      // Filter by price
      const maxPrice = searchParams.get('maxPrice');
      if (maxPrice) {
        const price = parseFloat(maxPrice);
        apis = apis.filter(api => {
          if (api.pricing.type === 'freemium') {
            return api.pricing.pricePerCall <= price;
          }
          return true;
        });
      }

      // Sort
      const sort = searchParams.get('sort') || 'popular';
      switch (sort) {
        case 'newest':
          apis.reverse(); // Simple simulation
          break;
        case 'rating':
          apis.sort((a, b) => b.rating - a.rating);
          break;
        case 'price-low':
          apis.sort((a, b) => {
            const priceA = a.pricing.pricePerCall || a.pricing.monthlyPrice || 0;
            const priceB = b.pricing.pricePerCall || b.pricing.monthlyPrice || 0;
            return priceA - priceB;
          });
          break;
        case 'popular':
        default:
          apis.sort((a, b) => b.calls - a.calls);
      }

      // Pagination with validation
      let page = parseInt(searchParams.get('page') || '1');
      let limit = parseInt(searchParams.get('limit') || '10');
      
      // Validate and set defaults for invalid values
      if (isNaN(page) || page < 1) {
        page = 1;
      }
      if (isNaN(limit) || limit < 1 || limit > 100) {
        limit = 10;
      }
      const startIndex = (page - 1) * limit;
      const endIndex = startIndex + limit;
      const paginatedApis = apis.slice(startIndex, endIndex);

      return res.status(200).json({
        success: true,
        data: paginatedApis,
        meta: {
          total: apis.length,
          page,
          limit,
          totalPages: Math.ceil(apis.length / limit)
        }
      });

    case '/apis/featured':
      // Apply rate limiting
      if (!checkRateLimit(res, clientIp, 'default')) {
        return;
      }
      
      const featuredApis = marketplaceData.apis.filter(api => api.featured);
      return res.status(200).json({
        success: true,
        data: featuredApis
      });

    case '/apis/trending':
      // Apply rate limiting
      if (!checkRateLimit(res, clientIp, 'default')) {
        return;
      }
      
      const trendingApis = marketplaceData.apis
        .filter(api => api.trending)
        .sort((a, b) => (b.growth || 0) - (a.growth || 0));
      return res.status(200).json({
        success: true,
        data: trendingApis
      });

    default:
      // Check if it's a specific API request
      if (path.startsWith('/apis/')) {
        // Apply rate limiting for API details endpoint
        if (!checkRateLimit(res, clientIp, 'apiDetails')) {
          return;
        }
        
        const apiId = path.replace('/apis/', '');
        
        // Prevent path traversal first
        if (apiId.includes('..') || apiId.includes('/') || apiId.includes('\\')) {
          return res.status(400).json({
            success: false,
            error: 'Invalid API ID'
          });
        }
        
        // Then validate API ID format (alphanumeric with hyphens only)
        if (!/^[a-zA-Z0-9-]+$/.test(apiId)) {
          return res.status(400).json({
            success: false,
            error: 'Invalid API ID format'
          });
        }
        
        const api = marketplaceData.apis.find(a => a.id === apiId);
        
        if (api) {
          return res.status(200).json({
            success: true,
            data: api
          });
        }
        
        return res.status(404).json({
          success: false,
          error: 'API not found'
        });
      }

      return res.status(404).json({
        success: false,
        error: 'Endpoint not found'
      });
  }
  } catch (error) {
    console.error('API Error:', error);
    return res.status(500).json({
      success: false,
      error: 'Internal server error'
    });
  }
}