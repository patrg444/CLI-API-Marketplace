export interface API {
  id: string
  creator_id: string
  name: string
  description: string
  category: string
  tags: string[]
  icon_url?: string
  banner_url?: string
  is_published: boolean
  published_at?: string
  created_at: string
  updated_at: string
  pricing_plans: PricingPlan[]
  documentation?: APIDocumentation
  reviews?: APIReview[]
  average_rating?: number
  total_reviews?: number
  total_subscriptions?: number
}

export interface PricingPlan {
  id: string
  api_id: string
  name: string
  type: 'free' | 'pay_per_use' | 'subscription'
  price_per_call?: number
  monthly_price?: number
  call_limit?: number
  rate_limit_per_minute?: number
  rate_limit_per_day?: number
  rate_limit_per_month?: number
  features?: Record<string, any>
  is_active: boolean
}

export interface APIDocumentation {
  id: string
  api_id: string
  openapi_spec?: any
  markdown_content?: string
  has_openapi: boolean
}

export interface APIReview {
  id: string
  api_id: string
  consumer_id: string
  consumer_email?: string
  rating: number
  comment: string
  created_at: string
}

export interface Subscription {
  id: string
  consumer_id: string
  api_id: string
  api?: API
  pricing_plan_id: string
  pricing_plan?: PricingPlan
  api_key_id?: string
  stripe_subscription_id?: string
  status: 'active' | 'cancelled' | 'past_due' | 'trial' | 'incomplete' | 'incomplete_expired'
  started_at: string
  cancelled_at?: string
  expires_at?: string
  current_period_start?: string
  current_period_end?: string
}

export interface APIKey {
  id: string
  key_prefix: string
  consumer_id: string
  subscription_id?: string
  name: string
  is_active: boolean
  created_at: string
  last_used_at?: string
}

export interface UsageSummary {
  subscription_id: string
  consumer_id: string
  api_id: string
  period_start: string
  period_end: string
  total_calls: number
  successful_calls: number
  failed_calls: number
  total_response_time_ms: number
  total_request_size_bytes: number
  total_response_size_bytes: number
  endpoint_usage: Record<string, number>
}

export interface PaymentMethod {
  id: string
  type: 'card'
  card?: {
    brand: string
    last4: string
    exp_month: number
    exp_year: number
  }
  is_default: boolean
  created_at: string
}

export interface Invoice {
  id: string
  stripe_invoice_id: string
  consumer_id: string
  subscription_id?: string
  amount_paid: number
  amount_due: number
  currency: string
  status: 'draft' | 'open' | 'paid' | 'uncollectible' | 'void'
  period_start: string
  period_end: string
  created_at: string
  paid_at?: string
  invoice_pdf?: string
  hosted_invoice_url?: string
}

export const API_CATEGORIES = [
  'AI/ML',
  'Data',
  'Communication',
  'Finance',
  'Media',
  'Security',
  'Analytics',
  'Productivity',
  'Developer Tools',
  'Other'
] as const

export type APICategory = typeof API_CATEGORIES[number]
