import axios, { AxiosInstance } from 'axios'
import { Auth } from 'aws-amplify'
import { API, PricingPlan, Subscription, APIKey, UsageSummary, PaymentMethod, Invoice } from '@/types/api'

class APIService {
  private client: AxiosInstance
  private apiKeyClient: AxiosInstance
  private billingClient: AxiosInstance
  private meteringClient: AxiosInstance
  private marketplaceClient: AxiosInstance

  constructor() {
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
        const session = await Auth.currentSession()
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
    const response = await this.marketplaceClient.get('/api/v1/marketplace/apis', { params })
    return response.data
  }

  async getAPI(id: string): Promise<API> {
    const response = await this.marketplaceClient.get(`/api/v1/marketplace/apis/${id}`)
    return response.data
  }

  async getAPIDocumentation(apiId: string) {
    const response = await this.marketplaceClient.get(`/api/v1/marketplace/apis/${apiId}/documentation`)
    return response.data
  }

  // Advanced Search
  async searchAPIs(params: {
    query?: string
    category?: string
    tags?: string[]
    price_range?: 'free' | 'low' | 'medium' | 'high'
    min_rating?: number
    has_free_tier?: boolean
    sort_by?: 'relevance' | 'rating' | 'subscriptions' | 'newest'
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
    const response = await this.marketplaceClient.post('/api/v1/marketplace/search', params)
    return response.data
  }

  async getSearchSuggestions(query: string): Promise<string[]> {
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
