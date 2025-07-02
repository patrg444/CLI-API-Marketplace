import axios, { AxiosInstance } from 'axios'
import { mockAuthUtils } from '@/utils/mockAuth'
import { API, PricingPlan, Subscription, APIKey, UsageSummary, PaymentMethod, Invoice } from '@/types/api'

// Mock data for development/testing
const MOCK_APIS: API[] = [
  {
    id: '1',
    creator_id: 'creator-1',
    name: 'Payment Processing API',
    description: 'Secure payment processing with multiple payment methods',
    category: 'Financial Services',
    is_published: true,
    icon_url: '',
    tags: ['payments', 'stripe', 'security'],
    average_rating: 4.5,
    total_reviews: 42,
    total_subscriptions: 1250,
    created_at: '2023-01-15T10:00:00Z',
    updated_at: '2023-12-01T15:30:00Z',
    pricing_plans: [
      {
        id: 'free-1',
        api_id: '1',
        name: 'Free Tier',
        type: 'free',
        monthly_price: 0,
        call_limit: 1000,
        rate_limit_per_minute: 10,
        features: {},
        is_active: true
      },
      {
        id: 'pro-1',
        api_id: '1',
        name: 'Professional',
        type: 'subscription',
        monthly_price: 99,
        call_limit: 50000,
        rate_limit_per_minute: 100,
        features: { 'priority_support': true },
        is_active: true
      }
    ]
  },
  {
    id: '2',
    creator_id: 'creator-2',
    name: 'Weather Data API',
    description: 'Real-time weather data and forecasts',
    category: 'Data',
    is_published: true,
    icon_url: '',
    tags: ['weather', 'forecast', 'data'],
    average_rating: 4.2,
    total_reviews: 28,
    total_subscriptions: 890,
    created_at: '2023-02-20T08:00:00Z',
    updated_at: '2023-11-15T12:45:00Z',
    pricing_plans: [
      {
        id: 'basic-2',
        api_id: '2',
        name: 'Basic',
        type: 'subscription',
        monthly_price: 29,
        call_limit: 10000,
        rate_limit_per_minute: 60,
        features: {},
        is_active: true
      }
    ]
  },
  {
    id: '3',
    creator_id: 'creator-3',
    name: 'Machine Learning API',
    description: 'Advanced machine learning models and predictions',
    category: 'AI/ML',
    is_published: true,
    icon_url: '',
    tags: ['ai', 'ml', 'predictions'],
    average_rating: 4.8,
    total_reviews: 156,
    total_subscriptions: 2100,
    created_at: '2023-03-10T14:20:00Z',
    updated_at: '2023-12-10T10:15:00Z',
    pricing_plans: [
      {
        id: 'free-3',
        api_id: '3',
        name: 'Free Tier',
        type: 'free',
        monthly_price: 0,
        call_limit: 500,
        rate_limit_per_minute: 5,
        features: {},
        is_active: true
      },
      {
        id: 'standard-3',
        api_id: '3',
        name: 'Standard',
        type: 'subscription',
        monthly_price: 149,
        call_limit: 100000,
        rate_limit_per_minute: 200,
        features: { 'premium_models': true },
        is_active: true
      }
    ]
  },
  {
    id: '4',
    creator_id: 'creator-4',
    name: 'Email Service API',
    description: 'Reliable email delivery and marketing automation',
    category: 'Communication',
    is_published: true,
    icon_url: '',
    tags: ['email', 'marketing', 'notifications'],
    average_rating: 4.3,
    total_reviews: 89,
    total_subscriptions: 1800,
    created_at: '2023-04-05T11:30:00Z',
    updated_at: '2023-12-05T16:45:00Z',
    pricing_plans: [
      {
        id: 'basic-4',
        api_id: '4',
        name: 'Basic',
        type: 'subscription',
        monthly_price: 19,
        call_limit: 5000,
        rate_limit_per_minute: 30,
        features: {},
        is_active: true
      }
    ]
  },
  {
    id: '5',
    creator_id: 'creator-5',
    name: 'Analytics API',
    description: 'Advanced analytics and data visualization tools',
    category: 'Analytics',
    is_published: true,
    icon_url: '',
    tags: ['analytics', 'data', 'visualization'],
    average_rating: 4.6,
    total_reviews: 134,
    total_subscriptions: 950,
    created_at: '2023-05-12T09:15:00Z',
    updated_at: '2023-12-08T14:20:00Z',
    pricing_plans: [
      {
        id: 'free-5',
        api_id: '5',
        name: 'Free Tier',
        type: 'free',
        monthly_price: 0,
        call_limit: 1000,
        rate_limit_per_minute: 10,
        features: {},
        is_active: true
      },
      {
        id: 'pro-5',
        api_id: '5',
        name: 'Professional',
        type: 'subscription',
        monthly_price: 79,
        call_limit: 25000,
        rate_limit_per_minute: 100,
        features: { 'advanced_charts': true },
        is_active: true
      }
    ]
  },
  {
    id: '6',
    creator_id: 'creator-6',
    name: 'Maps & Location API',
    description: 'Geolocation, mapping, and routing services',
    category: 'Maps & Location',
    is_published: true,
    icon_url: '',
    tags: ['maps', 'location', 'routing'],
    average_rating: 4.4,
    total_reviews: 67,
    total_subscriptions: 1200,
    created_at: '2023-06-18T13:40:00Z',
    updated_at: '2023-12-12T10:30:00Z',
    pricing_plans: [
      {
        id: 'standard-6',
        api_id: '6',
        name: 'Standard',
        type: 'subscription',
        monthly_price: 49,
        call_limit: 15000,
        rate_limit_per_minute: 80,
        features: {},
        is_active: true
      }
    ]
  },
  {
    id: '7',
    creator_id: 'creator-7',
    name: 'Authentication API',
    description: 'Secure user authentication and authorization',
    category: 'Authentication',
    is_published: true,
    icon_url: '',
    tags: ['auth', 'security', 'oauth'],
    average_rating: 4.7,
    total_reviews: 203,
    total_subscriptions: 2800,
    created_at: '2023-07-22T15:20:00Z',
    updated_at: '2023-12-15T11:10:00Z',
    pricing_plans: [
      {
        id: 'free-7',
        api_id: '7',
        name: 'Free Tier',
        type: 'free',
        monthly_price: 0,
        call_limit: 2000,
        rate_limit_per_minute: 20,
        features: {},
        is_active: true
      },
      {
        id: 'enterprise-7',
        api_id: '7',
        name: 'Enterprise',
        type: 'subscription',
        monthly_price: 299,
        call_limit: 500000,
        rate_limit_per_minute: 1000,
        features: { 'sso': true, 'audit_logs': true },
        is_active: true
      }
    ]
  },
  {
    id: '8',
    creator_id: 'creator-8',
    name: 'Storage API',
    description: 'Cloud storage and file management services',
    category: 'Storage',
    is_published: true,
    icon_url: '',
    tags: ['storage', 'files', 'cloud'],
    average_rating: 4.1,
    total_reviews: 156,
    total_subscriptions: 1500,
    created_at: '2023-08-30T12:00:00Z',
    updated_at: '2023-12-18T09:45:00Z',
    pricing_plans: [
      {
        id: 'basic-8',
        api_id: '8',
        name: 'Basic',
        type: 'subscription',
        monthly_price: 25,
        call_limit: 8000,
        rate_limit_per_minute: 50,
        features: {},
        is_active: true
      }
    ]
  },
  {
    id: '9',
    creator_id: 'creator-9',
    name: 'Social Media API',
    description: 'Social media integration and content management',
    category: 'Social',
    is_published: true,
    icon_url: '',
    tags: ['social', 'content', 'integration'],
    average_rating: 4.0,
    total_reviews: 98,
    total_subscriptions: 680,
    created_at: '2023-09-15T14:30:00Z',
    updated_at: '2023-12-20T16:15:00Z',
    pricing_plans: [
      {
        id: 'starter-9',
        api_id: '9',
        name: 'Starter',
        type: 'subscription',
        monthly_price: 39,
        call_limit: 10000,
        rate_limit_per_minute: 60,
        features: {},
        is_active: true
      }
    ]
  },
  {
    id: '10',
    creator_id: 'creator-10',
    name: 'E-commerce API',
    description: 'Complete e-commerce platform integration',
    category: 'E-commerce',
    is_published: true,
    icon_url: '',
    tags: ['ecommerce', 'shopping', 'orders'],
    average_rating: 4.5,
    total_reviews: 167,
    total_subscriptions: 2200,
    created_at: '2023-10-08T10:45:00Z',
    updated_at: '2023-12-22T13:30:00Z',
    pricing_plans: [
      {
        id: 'free-10',
        api_id: '10',
        name: 'Free Tier',
        type: 'free',
        monthly_price: 0,
        call_limit: 500,
        rate_limit_per_minute: 5,
        features: {},
        is_active: true
      },
      {
        id: 'business-10',
        api_id: '10',
        name: 'Business',
        type: 'subscription',
        monthly_price: 199,
        call_limit: 100000,
        rate_limit_per_minute: 200,
        features: { 'webhooks': true, 'analytics': true },
        is_active: true
      }
    ]
  },
  {
    id: '11',
    creator_id: 'creator-11',
    name: 'Media Processing API',
    description: 'Image and video processing and optimization',
    category: 'Media',
    is_published: true,
    icon_url: '',
    tags: ['media', 'images', 'video'],
    average_rating: 4.2,
    total_reviews: 78,
    total_subscriptions: 1100,
    created_at: '2023-11-12T16:20:00Z',
    updated_at: '2023-12-25T12:00:00Z',
    pricing_plans: [
      {
        id: 'standard-11',
        api_id: '11',
        name: 'Standard',
        type: 'subscription',
        monthly_price: 69,
        call_limit: 20000,
        rate_limit_per_minute: 90,
        features: {},
        is_active: true
      }
    ]
  },
  {
    id: '12',
    creator_id: 'creator-12',
    name: 'Tools API',
    description: 'Developer tools and utilities for productivity',
    category: 'Tools',
    is_published: true,
    icon_url: '',
    tags: ['tools', 'utilities', 'dev'],
    average_rating: 4.3,
    total_reviews: 145,
    total_subscriptions: 890,
    created_at: '2023-12-01T11:15:00Z',
    updated_at: '2023-12-28T14:45:00Z',
    pricing_plans: [
      {
        id: 'free-12',
        api_id: '12',
        name: 'Free Tier',
        type: 'free',
        monthly_price: 0,
        call_limit: 1500,
        rate_limit_per_minute: 15,
        features: {},
        is_active: true
      },
      {
        id: 'pro-12',
        api_id: '12',
        name: 'Pro',
        type: 'subscription',
        monthly_price: 59,
        call_limit: 30000,
        rate_limit_per_minute: 120,
        features: { 'priority_support': true },
        is_active: true
      }
    ]
  },
  {
    id: '13',
    creator_id: 'creator-13',
    name: 'Finance API',
    description: 'Financial data and market information',
    category: 'Financial Services',
    is_published: true,
    icon_url: '',
    tags: ['finance', 'market', 'trading'],
    average_rating: 4.6,
    total_reviews: 234,
    total_subscriptions: 1600,
    created_at: '2023-12-10T09:30:00Z',
    updated_at: '2023-12-30T10:20:00Z',
    pricing_plans: [
      {
        id: 'premium-13',
        api_id: '13',
        name: 'Premium',
        type: 'subscription',
        monthly_price: 149,
        call_limit: 50000,
        rate_limit_per_minute: 150,
        features: { 'real_time_data': true },
        is_active: true
      }
    ]
  }
]

const MOCK_SUBSCRIPTIONS: Subscription[] = [
  {
    id: 'sub-1',
    consumer_id: 'test-user-123',
    api_id: '1',
    pricing_plan_id: 'free-1',
    status: 'active',
    started_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
    current_period_end: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
    api: MOCK_APIS[0],
    pricing_plan: MOCK_APIS[0].pricing_plans![0]
  }
]

const MOCK_API_KEYS: APIKey[] = [
  {
    id: 'key-1',
    consumer_id: 'test-user-123',
    name: 'Development Key',
    key_prefix: 'ak_test_',
    is_active: true,
    created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
    last_used_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString()
  }
]

class APIService {
  private client: AxiosInstance
  private apiKeyClient: AxiosInstance
  private billingClient: AxiosInstance
  private meteringClient: AxiosInstance
  private marketplaceClient: AxiosInstance
  private isDevelopment: boolean

  constructor() {
    this.isDevelopment = process.env.NODE_ENV === 'development' || !process.env.REACT_APP_API_URL;
    this.client = axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_URL,
    })

    this.apiKeyClient = axios.create({
      baseURL: process.env.NEXT_PUBLIC_APIKEY_SERVICE_URL || 'http://localhost:8083',
    })

    this.billingClient = axios.create({
      baseURL: process.env.NEXT_PUBLIC_BILLING_SERVICE_URL || 'http://localhost:8085',
    })

    this.meteringClient = axios.create({
      baseURL: process.env.NEXT_PUBLIC_METERING_SERVICE_URL || 'http://localhost:8084',
    })

    this.marketplaceClient = axios.create({
      baseURL: process.env.NEXT_PUBLIC_MARKETPLACE_SERVICE_URL || 'http://localhost:8086',
    })

    // Add auth interceptor
    const authInterceptor = async (config: any) => {
      try {
        const session = await mockAuthUtils.getCurrentSession()
        const token = session.getIdToken().getJwtToken()
        config.headers.Authorization = `Bearer ${token}`
      } catch (error) {
        console.error('Auth error:', error)
      }
      return config
    }

    this.client.interceptors.request.use(authInterceptor)
    this.apiKeyClient.interceptors.request.use(authInterceptor)
    this.billingClient.interceptors.request.use(authInterceptor)
    this.meteringClient.interceptors.request.use(authInterceptor)
    this.marketplaceClient.interceptors.request.use(authInterceptor)
  }

  // API Discovery
  async listAPIs(params?: {
    category?: string
    search?: string
    page?: number
    limit?: number
  }): Promise<{ apis: API[]; total: number }> {
    if (this.isDevelopment) {
      // Return mock data in development
      let filteredApis = [...MOCK_APIS]
      
      if (params?.category && params.category !== 'All Categories') {
        filteredApis = filteredApis.filter(api => api.category === params.category)
      }
      
      if (params?.search) {
        const searchLower = params.search.toLowerCase()
        filteredApis = filteredApis.filter(api => 
          api.name.toLowerCase().includes(searchLower) ||
          api.description.toLowerCase().includes(searchLower) ||
          api.tags?.some(tag => tag.toLowerCase().includes(searchLower))
        )
      }
      
      return {
        apis: filteredApis,
        total: filteredApis.length
      }
    }
    
    const response = await this.marketplaceClient.get('/api/v1/marketplace/apis', { params })
    return response.data
  }

  async getAPI(id: string): Promise<API> {
    if (this.isDevelopment) {
      const api = MOCK_APIS.find(a => a.id === id)
      if (!api) throw new Error('API not found')
      return api
    }
    
    const response = await this.marketplaceClient.get(`/api/v1/marketplace/apis/${id}`)
    return response.data
  }

  async getAPIDocumentation(apiId: string) {
    const response = await this.marketplaceClient.get(`/api/v1/marketplace/apis/${apiId}/documentation`)
    return response.data
  }

  // Advanced Search
  async searchAPIs(params: {
    q?: string
    category?: string
    tags?: string[]
    price_range?: 'free' | 'low' | 'medium' | 'high'
    min_rating?: number
    has_free_tier?: boolean
    sort_by?: 'relevance' | 'rating' | 'popularity' | 'newest'
    page?: number
    limit?: number
  }): Promise<{ 
    apis: API[]
    total: number
    facets: {
      categories: { [key: string]: number }
      tags: { [key: string]: number }
      price_ranges: { [key: string]: number }
      ratings: { [key: string]: number }
    }
  }> {
    if (this.isDevelopment) {
      // Mock search implementation
      let filteredApis = [...MOCK_APIS]
      
      // Filter by query (with fuzzy matching)
      if (params.q) {
        const queryLower = params.q.toLowerCase()
        const queryWords = queryLower.split(/\s+/).filter(word => word.length > 0)
        
        filteredApis = filteredApis.filter(api => {
          const searchText = `${api.name} ${api.description} ${api.tags?.join(' ') || ''}`.toLowerCase()
          
          // Exact match
          if (searchText.includes(queryLower)) {
            return true
          }
          
          // Fuzzy matching: check if most query words appear in the text
          const matchingWords = queryWords.filter(word => {
            // Allow for small typos by checking if the word is similar
            return searchText.includes(word) || 
                   searchText.split(/\s+/).some(textWord => {
                     // Simple fuzzy matching: allow 1-2 character differences for words > 4 chars
                     if (word.length > 4 && textWord.length > 4) {
                       const maxDiff = Math.min(2, Math.floor(word.length / 3))
                       let diff = 0
                       const minLen = Math.min(word.length, textWord.length)
                       for (let i = 0; i < minLen; i++) {
                         if (word[i] !== textWord[i]) diff++
                         if (diff > maxDiff) break
                       }
                       return diff <= maxDiff && Math.abs(word.length - textWord.length) <= 1
                     }
                     return false
                   })
          })
          
          // Return true if at least half of the query words match
          return matchingWords.length >= Math.ceil(queryWords.length / 2)
        })
      }
      
      // Filter by category
      if (params.category && params.category !== 'All Categories') {
        filteredApis = filteredApis.filter(api => api.category === params.category)
      }
      
      // Filter by rating
      if (params.min_rating) {
        filteredApis = filteredApis.filter(api => (api.average_rating || 0) >= params.min_rating!)
      }
      
      // Filter by free tier
      if (params.has_free_tier) {
        filteredApis = filteredApis.filter(api => 
          api.pricing_plans?.some(plan => plan.type === 'free')
        )
      }
      
      // Filter by price range
      if (params.price_range) {
        filteredApis = filteredApis.filter(api => {
          const paidPlans = api.pricing_plans?.filter(plan => plan.type !== 'free') || []
          const minPrice = paidPlans.length > 0 ? Math.min(...paidPlans.map(p => p.monthly_price || 0)) : 0
          
          switch (params.price_range) {
            case 'free':
              return api.pricing_plans?.some(plan => plan.type === 'free')
            case 'low':
              return minPrice > 0 && minPrice <= 50
            case 'medium':
              return minPrice > 50 && minPrice <= 200
            case 'high':
              return minPrice > 200
            default:
              return true
          }
        })
      }
      
      // Sort results
      if (params.sort_by) {
        switch (params.sort_by) {
          case 'rating':
            filteredApis.sort((a, b) => (b.average_rating || 0) - (a.average_rating || 0))
            break
          case 'popularity':
            filteredApis.sort((a, b) => (b.total_subscriptions || 0) - (a.total_subscriptions || 0))
            break
          case 'newest':
            filteredApis.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
            break
          default: // relevance
            // Keep current order for relevance
            break
        }
      }
      
      return {
        apis: filteredApis,
        total: filteredApis.length,
        facets: {
          categories: {
            'Financial Services': MOCK_APIS.filter(api => api.category === 'Financial Services').length,
            'Data': MOCK_APIS.filter(api => api.category === 'Data').length,
            'AI/ML': MOCK_APIS.filter(api => api.category === 'AI/ML').length,
            'Communication': MOCK_APIS.filter(api => api.category === 'Communication').length,
            'Analytics': MOCK_APIS.filter(api => api.category === 'Analytics').length,
            'Authentication': MOCK_APIS.filter(api => api.category === 'Authentication').length,
            'E-commerce': MOCK_APIS.filter(api => api.category === 'E-commerce').length,
            'Storage': MOCK_APIS.filter(api => api.category === 'Storage').length,
            'Social': MOCK_APIS.filter(api => api.category === 'Social').length,
            'Media': MOCK_APIS.filter(api => api.category === 'Media').length,
            'Tools': MOCK_APIS.filter(api => api.category === 'Tools').length,
            'Maps & Location': MOCK_APIS.filter(api => api.category === 'Maps & Location').length,
          },
          tags: {
            'payments': MOCK_APIS.filter(api => api.tags?.includes('payments')).length,
            'stripe': MOCK_APIS.filter(api => api.tags?.includes('stripe')).length,
            'weather': MOCK_APIS.filter(api => api.tags?.includes('weather')).length,
            'ai': MOCK_APIS.filter(api => api.tags?.includes('ai')).length,
            'ml': MOCK_APIS.filter(api => api.tags?.includes('ml')).length,
            'email': MOCK_APIS.filter(api => api.tags?.includes('email')).length,
            'auth': MOCK_APIS.filter(api => api.tags?.includes('auth')).length,
            'security': MOCK_APIS.filter(api => api.tags?.includes('security')).length,
            'data': MOCK_APIS.filter(api => api.tags?.includes('data')).length,
            'analytics': MOCK_APIS.filter(api => api.tags?.includes('analytics')).length,
          },
          price_ranges: {
            'free': MOCK_APIS.filter(api => api.pricing_plans?.some(plan => plan.type === 'free')).length,
            'low': MOCK_APIS.filter(api => {
              const paidPlans = api.pricing_plans?.filter(plan => plan.type !== 'free') || []
              const minPrice = paidPlans.length > 0 ? Math.min(...paidPlans.map(p => p.monthly_price || 0)) : 0
              return minPrice > 0 && minPrice <= 50
            }).length,
            'medium': MOCK_APIS.filter(api => {
              const paidPlans = api.pricing_plans?.filter(plan => plan.type !== 'free') || []
              const minPrice = paidPlans.length > 0 ? Math.min(...paidPlans.map(p => p.monthly_price || 0)) : 0
              return minPrice > 50 && minPrice <= 200
            }).length,
          },
          ratings: {
            '4+': MOCK_APIS.filter(api => (api.average_rating || 0) >= 4).length,
            '3+': MOCK_APIS.filter(api => (api.average_rating || 0) >= 3).length,
          }
        }
      }
    }
    
    const response = await this.marketplaceClient.post('/api/v1/marketplace/search', params)
    return response.data
  }

  async getSearchSuggestions(query: string): Promise<string[]> {
    if (this.isDevelopment) {
      // Mock suggestions based on available APIs and common search terms
      const suggestions = [
        'Payment Processing API',
        'Weather Data API',
        'payment processing',
        'weather forecast',
        'stripe integration',
        'data analytics',
        'authentication',
        'email service'
      ]
      
      const queryLower = query.toLowerCase()
      return suggestions.filter(suggestion => 
        suggestion.toLowerCase().includes(queryLower)
      ).slice(0, 5)
    }
    
    const response = await this.marketplaceClient.get('/api/v1/marketplace/search/suggestions', {
      params: { q: query }
    })
    return response.data.suggestions || []
  }

  // Consumer Management
  async registerConsumer(): Promise<{ consumer_id: string; stripe_customer_id: string }> {
    const response = await this.billingClient.post('/api/v1/consumers/register')
    return response.data
  }

  // Subscriptions
  async createSubscription(data: {
    api_id: string
    pricing_plan_id: string
    payment_method_id?: string
  }): Promise<Subscription> {
    const response = await this.billingClient.post('/api/v1/subscriptions', data)
    return response.data
  }

  async listMySubscriptions(): Promise<Subscription[]> {
    if (this.isDevelopment) {
      return MOCK_SUBSCRIPTIONS
    }
    
    const response = await this.billingClient.get('/api/v1/subscriptions')
    return response.data.subscriptions || []
  }

  async getSubscription(subscriptionId: string): Promise<Subscription> {
    const response = await this.billingClient.get(`/api/v1/subscriptions/${subscriptionId}`)
    return response.data
  }

  async cancelSubscription(subscriptionId: string): Promise<void> {
    await this.billingClient.put(`/api/v1/subscriptions/${subscriptionId}/cancel`)
  }

  async updateSubscription(subscriptionId: string, pricingPlanId: string): Promise<Subscription> {
    const response = await this.billingClient.put(`/api/v1/subscriptions/${subscriptionId}/upgrade`, {
      new_pricing_plan_id: pricingPlanId
    })
    return response.data
  }

  // Payment Methods
  async addPaymentMethod(paymentMethodId: string): Promise<PaymentMethod> {
    const response = await this.billingClient.post('/api/v1/payment-methods', {
      payment_method_id: paymentMethodId
    })
    return response.data
  }

  async listPaymentMethods(): Promise<PaymentMethod[]> {
    const response = await this.billingClient.get('/api/v1/payment-methods')
    return response.data.payment_methods || []
  }

  async setDefaultPaymentMethod(paymentMethodId: string): Promise<void> {
    await this.billingClient.put(`/api/v1/payment-methods/${paymentMethodId}/default`)
  }

  async removePaymentMethod(paymentMethodId: string): Promise<void> {
    await this.billingClient.delete(`/api/v1/payment-methods/${paymentMethodId}`)
  }

  // Invoices
  async listInvoices(): Promise<Invoice[]> {
    const response = await this.billingClient.get('/api/v1/invoices')
    return response.data.invoices || []
  }

  async getInvoice(invoiceId: string): Promise<Invoice> {
    const response = await this.billingClient.get(`/api/v1/invoices/${invoiceId}`)
    return response.data
  }

  async downloadInvoice(invoiceId: string): Promise<string> {
    const response = await this.billingClient.get(`/api/v1/invoices/${invoiceId}/download`)
    return response.data.download_url
  }

  // API Keys
  async listAPIKeys(): Promise<APIKey[]> {
    if (this.isDevelopment) {
      return MOCK_API_KEYS
    }
    
    const response = await this.apiKeyClient.get('/api/v1/keys')
    return response.data.keys || []
  }

  async createAPIKey(subscriptionId: string, name: string): Promise<{ key: string; key_data: APIKey }> {
    const response = await this.apiKeyClient.post('/api/v1/keys', {
      subscription_id: subscriptionId,
      name,
    })
    return response.data
  }

  async revokeAPIKey(keyId: string): Promise<void> {
    await this.apiKeyClient.delete(`/api/v1/keys/${keyId}`)
  }

  async updateAPIKeyName(keyId: string, name: string): Promise<void> {
    await this.apiKeyClient.put(`/api/v1/keys/${keyId}`, { name })
  }

  // Usage & Billing
  async getUsageSummary(subscriptionId: string, start?: string, end?: string): Promise<UsageSummary> {
    const response = await this.meteringClient.get('/api/v1/usage/summary', {
      params: { subscription_id: subscriptionId, start_date: start, end_date: end },
    })
    return response.data
  }

  async getMyUsage(): Promise<{
    consumer_id: string
    total_calls: number
    successful_calls: number
    failed_calls: number
    current_month_cost: number
    subscriptions: Array<{
      subscription_id: string
      api_id: string
      api_name: string
      total_calls: number
      successful_calls: number
      failed_calls: number
    }>
  }> {
    if (this.isDevelopment) {
      return {
        consumer_id: 'test-user',
        total_calls: 15420,
        successful_calls: 15380,
        failed_calls: 40,
        current_month_cost: 24.99,
        subscriptions: [
          {
            subscription_id: 'sub-1',
            api_id: '1',
            api_name: 'Payment Processing API',
            total_calls: 15420,
            successful_calls: 15380,
            failed_calls: 40
          }
        ]
      }
    }
    
    const response = await this.meteringClient.get('/api/v1/usage/me')
    return response.data
  }

  // Reviews
  async submitReview(apiId: string, data: {
    rating: number
    title: string
    comment: string
  }): Promise<void> {
    await this.marketplaceClient.post(`/api/v1/marketplace/apis/${apiId}/reviews`, data)
  }

  async getAPIReviews(apiId: string, params?: {
    page?: number
    limit?: number
    sort?: 'newest' | 'oldest' | 'highest' | 'lowest' | 'most_helpful'
  }) {
    const response = await this.marketplaceClient.get(`/api/v1/marketplace/apis/${apiId}/reviews`, {
      params: params || { page: 1, limit: 10 }
    })
    return response.data
  }

  async getReviewStats(apiId: string) {
    const response = await this.marketplaceClient.get(`/api/v1/marketplace/apis/${apiId}/reviews/stats`)
    return response.data
  }

  async voteOnReview(reviewId: string, helpful: boolean): Promise<void> {
    await this.marketplaceClient.post(`/api/v1/marketplace/reviews/${reviewId}/vote`, {
      helpful
    })
  }

  async respondToReview(reviewId: string, response: string): Promise<void> {
    await this.marketplaceClient.post(`/api/v1/marketplace/reviews/${reviewId}/response`, {
      response
    })
  }
}

export default new APIService()
