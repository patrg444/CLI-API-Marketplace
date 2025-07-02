import React, { useState, useEffect } from 'react'
import { useRouter } from 'next/router'
import { useQuery, useMutation } from 'react-query'
import {
  Elements,
  CardElement,
  useStripe,
  useElements,
} from '@stripe/react-stripe-js'
import Layout from '@/components/Layout'
import apiService from '@/services/api'
import { getStripe, formatCurrency } from '@/utils/stripe'
import { API, PricingPlan } from '@/types/api'
import { mockAuthUtils } from '@/utils/mockAuth'

const CheckoutForm: React.FC<{
  api: API
  selectedPlan: PricingPlan
  onSuccess: (subscriptionId: string) => void
}> = ({ api, selectedPlan, onSuccess }) => {
  const stripe = useStripe()
  const elements = useElements()
  const [error, setError] = useState<string | null>(null)
  const [processing, setProcessing] = useState(false)
  const [succeeded, setSucceeded] = useState(false)

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()

    if (!stripe || !elements) {
      return
    }

    setProcessing(true)
    setError(null)

    const card = elements.getElement(CardElement)
    if (!card) {
      setError('Card element not found')
      setProcessing(false)
      return
    }

    try {
      // Create payment method
      const { error: pmError, paymentMethod } = await stripe.createPaymentMethod({
        type: 'card',
        card: card,
      })

      if (pmError) {
        setError(pmError.message || 'Payment method creation failed')
        setProcessing(false)
        return
      }

      // Register consumer if needed
      await apiService.registerConsumer()

      // Create subscription
      const subscription = await apiService.createSubscription({
        api_id: api.id,
        pricing_plan_id: selectedPlan.id,
        payment_method_id: paymentMethod.id,
      })

      // Handle 3D Secure if required
      if (subscription.status === 'incomplete') {
        const { error: confirmError } = await stripe.confirmCardPayment(
          subscription.stripe_subscription_id || '',
        )
        if (confirmError) {
          setError(confirmError.message || 'Payment confirmation failed')
          setProcessing(false)
          return
        }
      }

      setSucceeded(true)
      onSuccess(subscription.id)
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Subscription creation failed')
      setProcessing(false)
    }
  }

  const cardStyle = {
    style: {
      base: {
        color: '#32325d',
        fontFamily: 'Arial, sans-serif',
        fontSmoothing: 'antialiased',
        fontSize: '16px',
        '::placeholder': {
          color: '#aab7c4',
        },
      },
      invalid: {
        color: '#fa755a',
        iconColor: '#fa755a',
      },
    },
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Card details
        </label>
        <div className="border border-gray-300 rounded-md p-3">
          <CardElement options={cardStyle} />
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <button
        type="submit"
        disabled={!stripe || processing || succeeded}
        className={`w-full py-3 px-4 rounded-md font-medium text-white transition-colors ${
          processing || succeeded
            ? 'bg-gray-400 cursor-not-allowed'
            : 'bg-indigo-600 hover:bg-indigo-700'
        }`}
      >
        {processing ? (
          <span className="flex items-center justify-center">
            <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Processing...
          </span>
        ) : succeeded ? (
          'Success!'
        ) : (
          `Subscribe for ${formatCurrency(selectedPlan.monthly_price || 0)}/month`
        )}
      </button>

      <p className="text-xs text-gray-500 text-center">
        Your subscription will renew automatically each month. You can cancel anytime from your dashboard.
      </p>
    </form>
  )
}

const SubscribePage: React.FC = () => {
  const router = useRouter()
  const { apiId } = router.query
  const [selectedPlanId, setSelectedPlanId] = useState<string | null>(null)
  const [showApiKey, setShowApiKey] = useState(false)
  const [newApiKey, setNewApiKey] = useState<string | null>(null)

  // Check authentication
  const { data: user, isLoading: userLoading } = useQuery('currentUser', async () => {
    try {
      const currentUser = await mockAuthUtils.getCurrentUser()
      return currentUser
    } catch {
      router.push(`/auth/login?redirect=/subscribe/${apiId}`)
      return null
    }
  })

  // Fetch API details
  const { data: api, isLoading: apiLoading } = useQuery(
    ['api', apiId],
    () => apiService.getAPI(apiId as string),
    {
      enabled: !!apiId && !!user,
    }
  )

  // Get selected plan from query params or default to first non-free plan
  useEffect(() => {
    const planId = router.query.plan as string
    if (planId) {
      setSelectedPlanId(planId)
    } else if (api?.pricing_plans) {
      const nonFreePlan = api.pricing_plans.find(p => p.type !== 'free' && p.is_active)
      if (nonFreePlan) {
        setSelectedPlanId(nonFreePlan.id)
      }
    }
  }, [router.query.plan, api])

  const selectedPlan = api?.pricing_plans.find(p => p.id === selectedPlanId)

  const handleSubscriptionSuccess = async (subscriptionId: string) => {
    try {
      // Generate API key for the new subscription
      const keyResult = await apiService.createAPIKey(subscriptionId, `${api?.name} API Key`)
      setNewApiKey(keyResult.key)
      setShowApiKey(true)
    } catch (error) {
      console.error('Error generating API key:', error)
      // Still redirect to dashboard even if key generation fails
      router.push('/dashboard')
    }
  }

  const handleContinueToDashboard = () => {
    router.push('/dashboard')
  }

  if (userLoading || apiLoading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-screen">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
        </div>
      </Layout>
    )
  }

  if (!user || !api || !selectedPlan) {
    return null
  }

  if (showApiKey && newApiKey) {
    return (
      <Layout>
        <div className="max-w-2xl mx-auto px-4 py-16">
          <div className="bg-white shadow-lg rounded-lg p-8">
            <div className="text-center mb-8">
              <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4">
                <svg className="h-6 w-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <h2 className="text-2xl font-bold text-gray-900">Subscription Successful!</h2>
              <p className="mt-2 text-gray-600">Your API key has been generated</p>
            </div>

            <div className="bg-gray-50 rounded-lg p-6 mb-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Your API Key</h3>
              <div className="bg-white rounded border border-gray-200 p-4 font-mono text-sm break-all">
                {newApiKey}
              </div>
              <p className="mt-4 text-sm text-amber-600">
                <strong>Important:</strong> This is the only time you{'\''}ll see this key. Please copy and store it securely.
              </p>
            </div>

            <div className="space-y-4">
              <button
                onClick={() => navigator.clipboard.writeText(newApiKey)}
                className="w-full py-2 px-4 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 transition-colors"
              >
                Copy API Key
              </button>
              <button
                onClick={handleContinueToDashboard}
                className="w-full py-2 px-4 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 transition-colors"
              >
                Continue to Dashboard
              </button>
            </div>
          </div>
        </div>
      </Layout>
    )
  }

  return (
    <Layout>
      <div className="max-w-4xl mx-auto px-4 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Order Summary */}
          <div className="order-2 lg:order-1">
            <h2 className="text-2xl font-bold text-gray-900 mb-6">Order Summary</h2>
            
            <div className="bg-white shadow rounded-lg p-6 mb-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">{api.name}</h3>
              <p className="text-gray-600 mb-4">{api.description}</p>
              
              <div className="border-t pt-4">
                <h4 className="font-medium text-gray-900 mb-2">Selected Plan: {selectedPlan.name}</h4>
                <div className="space-y-2 text-sm text-gray-600">
                  {selectedPlan.call_limit && (
                    <p>• {selectedPlan.call_limit.toLocaleString()} API calls/month</p>
                  )}
                  {selectedPlan.rate_limit_per_minute && (
                    <p>• {selectedPlan.rate_limit_per_minute} requests/minute</p>
                  )}
                  {selectedPlan.features && Object.entries(selectedPlan.features).map(([key, value]) => (
                    <p key={key}>• {key}: {String(value)}</p>
                  ))}
                </div>
              </div>

              <div className="border-t mt-4 pt-4">
                <div className="flex justify-between text-lg font-medium">
                  <span>Monthly Total</span>
                  <span>{formatCurrency(selectedPlan.monthly_price || 0)}</span>
                </div>
              </div>
            </div>

            {/* Plan Selection */}
            {api.pricing_plans.filter(p => p.type !== 'free' && p.is_active).length > 1 && (
              <div className="bg-white shadow rounded-lg p-6">
                <h4 className="font-medium text-gray-900 mb-4">Change Plan</h4>
                <div className="space-y-2">
                  {api.pricing_plans
                    .filter(p => p.type !== 'free' && p.is_active)
                    .map(plan => (
                      <label
                        key={plan.id}
                        className="flex items-center p-3 border rounded-lg cursor-pointer hover:bg-gray-50"
                      >
                        <input
                          type="radio"
                          name="plan"
                          value={plan.id}
                          checked={selectedPlanId === plan.id}
                          onChange={() => setSelectedPlanId(plan.id)}
                          className="mr-3"
                        />
                        <div className="flex-1">
                          <span className="font-medium">{plan.name}</span>
                          <span className="ml-2 text-gray-600">
                            {formatCurrency(plan.monthly_price || 0)}/month
                          </span>
                        </div>
                      </label>
                    ))}
                </div>
              </div>
            )}
          </div>

          {/* Payment Form */}
          <div className="order-1 lg:order-2">
            <h2 className="text-2xl font-bold text-gray-900 mb-6">Payment Details</h2>
            
            <div className="bg-white shadow rounded-lg p-6">
              <Elements stripe={getStripe()}>
                <CheckoutForm
                  api={api}
                  selectedPlan={selectedPlan}
                  onSuccess={handleSubscriptionSuccess}
                />
              </Elements>
            </div>

            <div className="mt-6 flex items-center justify-center text-sm text-gray-500">
              <svg className="h-5 w-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
              Secured by Stripe
            </div>
          </div>
        </div>
      </div>
    </Layout>
  )
}

export default SubscribePage
