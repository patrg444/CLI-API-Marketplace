import React, { useState } from 'react'
import { useRouter } from 'next/router'
import { useQuery, useMutation } from 'react-query'
import Layout from '@/components/Layout'
import apiService from '@/services/api'
import { formatCurrency, getCardBrandIcon } from '@/utils/stripe'
import { getCurrentUser } from 'aws-amplify/auth'
import { Subscription, APIKey, Invoice, PaymentMethod } from '@/types/api'

const Dashboard: React.FC = () => {
  const router = useRouter()
  const [selectedApiKey, setSelectedApiKey] = useState<APIKey | null>(null)
  const [showApiKeyModal, setShowApiKeyModal] = useState(false)
  const [newApiKeyName, setNewApiKeyName] = useState('')
  
  // Check authentication
  const { data: user, isLoading: userLoading } = useQuery('currentUser', async () => {
    try {
      const currentUser = await getCurrentUser()
      return currentUser
    } catch {
      router.push('/auth/login')
      return null
    }
  })

  // Fetch subscriptions
  const { data: subscriptions, refetch: refetchSubscriptions } = useQuery(
    'subscriptions',
    () => apiService.listMySubscriptions(),
    { enabled: !!user }
  )

  // Fetch API keys
  const { data: apiKeys, refetch: refetchApiKeys } = useQuery(
    'apiKeys',
    () => apiService.listAPIKeys(),
    { enabled: !!user }
  )

  // Fetch invoices
  const { data: invoices } = useQuery(
    'invoices',
    () => apiService.listInvoices(),
    { enabled: !!user }
  )

  // Fetch usage
  const { data: usage } = useQuery(
    'usage',
    () => apiService.getMyUsage(),
    { enabled: !!user }
  )

  // Fetch payment methods
  const { data: paymentMethods } = useQuery(
    'paymentMethods',
    () => apiService.listPaymentMethods(),
    { enabled: !!user }
  )

  // Cancel subscription mutation
  const cancelSubscriptionMutation = useMutation(
    (subscriptionId: string) => apiService.cancelSubscription(subscriptionId),
    {
      onSuccess: () => {
        refetchSubscriptions()
      }
    }
  )

  // Revoke API key mutation
  const revokeApiKeyMutation = useMutation(
    (keyId: string) => apiService.revokeAPIKey(keyId),
    {
      onSuccess: () => {
        refetchApiKeys()
      }
    }
  )

  // Update API key name mutation
  const updateApiKeyNameMutation = useMutation(
    ({ keyId, name }: { keyId: string; name: string }) => 
      apiService.updateAPIKeyName(keyId, name),
    {
      onSuccess: () => {
        refetchApiKeys()
        setShowApiKeyModal(false)
        setSelectedApiKey(null)
        setNewApiKeyName('')
      }
    }
  )

  const handleCancelSubscription = (subscriptionId: string) => {
    if (confirm('Are you sure you want to cancel this subscription?')) {
      cancelSubscriptionMutation.mutate(subscriptionId)
    }
  }

  const handleRevokeApiKey = (keyId: string) => {
    if (confirm('Are you sure you want to revoke this API key? This action cannot be undone.')) {
      revokeApiKeyMutation.mutate(keyId)
    }
  }

  const handleEditApiKey = (apiKey: APIKey) => {
    setSelectedApiKey(apiKey)
    setNewApiKeyName(apiKey.name)
    setShowApiKeyModal(true)
  }

  const handleUpdateApiKeyName = () => {
    if (selectedApiKey && newApiKeyName.trim()) {
      updateApiKeyNameMutation.mutate({
        keyId: selectedApiKey.id,
        name: newApiKeyName.trim()
      })
    }
  }

  if (userLoading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-screen">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
        </div>
      </Layout>
    )
  }

  if (!user) {
    return null
  }

  const activeSubscriptions = subscriptions?.filter(s => s.status === 'active') || []
  const totalCalls = usage?.total_calls || 0
  const monthlyUsage = usage?.current_month_cost || 0

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">My Dashboard</h1>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* My Subscriptions */}
          <div className="lg:col-span-2">
            <div className="bg-white shadow rounded-lg">
              <div className="px-6 py-4 border-b border-gray-200">
                <h2 className="text-lg font-medium text-gray-900">My Subscriptions</h2>
              </div>
              <div className="p-6">
                {activeSubscriptions.length > 0 ? (
                  <div className="space-y-4">
                    {activeSubscriptions.map((subscription) => (
                      <div key={subscription.id} className="border rounded-lg p-4">
                        <div className="flex justify-between items-start">
                          <div>
                            <h3 className="font-medium text-gray-900">
                              {subscription.api?.name || 'Unknown API'}
                            </h3>
                            <p className="text-sm text-gray-500 mt-1">
                              Plan: {subscription.pricing_plan?.name || 'Unknown Plan'}
                            </p>
                            <p className="text-sm text-gray-500">
                              Status: <span className="text-green-600">{subscription.status}</span>
                            </p>
                            {subscription.current_period_end && (
                              <p className="text-sm text-gray-500">
                                Renews: {new Date(subscription.current_period_end).toLocaleDateString()}
                              </p>
                            )}
                          </div>
                          <div className="flex space-x-2">
                            <button
                              onClick={() => router.push(`/api/${subscription.api_id}`)}
                              className="text-sm text-indigo-600 hover:text-indigo-500"
                            >
                              View API
                            </button>
                            <button
                              onClick={() => handleCancelSubscription(subscription.id)}
                              className="text-sm text-red-600 hover:text-red-500"
                            >
                              Cancel
                            </button>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-12">
                    <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                    </svg>
                    <h3 className="mt-2 text-sm font-medium text-gray-900">No subscriptions</h3>
                    <p className="mt-1 text-sm text-gray-500">Get started by subscribing to an API.</p>
                    <div className="mt-6">
                      <button
                        onClick={() => router.push('/')}
                        className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
                      >
                        Browse APIs
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* My API Keys */}
            <div className="mt-6 bg-white shadow rounded-lg">
              <div className="px-6 py-4 border-b border-gray-200">
                <h2 className="text-lg font-medium text-gray-900">My API Keys</h2>
              </div>
              <div className="p-6">
                {apiKeys && apiKeys.length > 0 ? (
                  <div className="space-y-3">
                    {apiKeys.map((apiKey) => (
                      <div key={apiKey.id} className="flex items-center justify-between p-3 border rounded-lg">
                        <div className="flex-1">
                          <p className="font-medium text-gray-900">{apiKey.name}</p>
                          <p className="text-sm text-gray-500">
                            {apiKey.key_prefix}••• • Created: {new Date(apiKey.created_at).toLocaleDateString()}
                          </p>
                          {apiKey.last_used_at && (
                            <p className="text-xs text-gray-400">
                              Last used: {new Date(apiKey.last_used_at).toLocaleDateString()}
                            </p>
                          )}
                        </div>
                        <div className="flex space-x-2">
                          <button
                            onClick={() => handleEditApiKey(apiKey)}
                            className="text-sm text-indigo-600 hover:text-indigo-500"
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => handleRevokeApiKey(apiKey.id)}
                            className="text-sm text-red-600 hover:text-red-500"
                          >
                            Revoke
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                    </svg>
                    <h3 className="mt-2 text-sm font-medium text-gray-900">No API keys</h3>
                    <p className="mt-1 text-sm text-gray-500">API keys will appear here after subscribing.</p>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Quick Stats */}
            <div className="bg-white shadow rounded-lg p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Quick Stats</h3>
              <dl className="space-y-3">
                <div className="flex justify-between">
                  <dt className="text-sm text-gray-500">Active Subscriptions</dt>
                  <dd className="text-sm font-medium text-gray-900">{activeSubscriptions.length}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-sm text-gray-500">Total API Calls</dt>
                  <dd className="text-sm font-medium text-gray-900">{totalCalls.toLocaleString()}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="text-sm text-gray-500">This Month's Usage</dt>
                  <dd className="text-sm font-medium text-gray-900">{formatCurrency(monthlyUsage)}</dd>
                </div>
              </dl>
            </div>

            {/* Billing History */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-6 py-4 border-b border-gray-200">
                <h3 className="text-lg font-medium text-gray-900">Billing History</h3>
              </div>
              <div className="p-6">
                {invoices && invoices.length > 0 ? (
                  <div className="space-y-3">
                    {invoices.slice(0, 5).map((invoice) => (
                      <div key={invoice.id} className="text-sm">
                        <div className="flex justify-between">
                          <span className="text-gray-900">
                            {new Date(invoice.created_at).toLocaleDateString()}
                          </span>
                          <span className="font-medium text-gray-900">
                            {formatCurrency(invoice.amount_paid)}
                          </span>
                        </div>
                        <div className="flex justify-between mt-1">
                          <span className="text-gray-500">
                            {invoice.status === 'paid' ? '✓ Paid' : invoice.status}
                          </span>
                          {invoice.invoice_pdf && (
                            <a
                              href={invoice.invoice_pdf}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-indigo-600 hover:text-indigo-500"
                            >
                              Download
                            </a>
                          )}
                        </div>
                      </div>
                    ))}
                    {invoices.length > 5 && (
                      <button className="text-sm text-indigo-600 hover:text-indigo-500 mt-2">
                        View all invoices
                      </button>
                    )}
                  </div>
                ) : (
                  <p className="text-sm text-gray-500">No billing history available.</p>
                )}
              </div>
            </div>

            {/* Payment Methods */}
            <div className="bg-white shadow rounded-lg">
              <div className="px-6 py-4 border-b border-gray-200">
                <h3 className="text-lg font-medium text-gray-900">Payment Methods</h3>
              </div>
              <div className="p-6">
                {paymentMethods && paymentMethods.length > 0 ? (
                  <div className="space-y-3">
                    {paymentMethods.map((pm) => (
                      <div key={pm.id} className="flex items-center text-sm">
                        <span className="mr-2">{getCardBrandIcon(pm.card?.brand || '')}</span>
                        <span className="text-gray-900">
                          •••• {pm.card?.last4}
                        </span>
                        {pm.is_default && (
                          <span className="ml-2 text-xs text-gray-500">(default)</span>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-gray-500">No payment methods on file.</p>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Usage Overview */}
        <div className="mt-6 bg-white shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">Usage Overview</h2>
          </div>
          <div className="p-6">
            {usage && usage.subscriptions.length > 0 ? (
              <div className="space-y-4">
                {usage.subscriptions.map((sub) => (
                  <div key={sub.subscription_id} className="border rounded-lg p-4">
                    <h4 className="font-medium text-gray-900">{sub.api_name}</h4>
                    <div className="mt-2 grid grid-cols-3 gap-4 text-sm">
                      <div>
                        <p className="text-gray-500">Total Calls</p>
                        <p className="font-medium">{sub.total_calls.toLocaleString()}</p>
                      </div>
                      <div>
                        <p className="text-gray-500">Success Rate</p>
                        <p className="font-medium">
                          {sub.total_calls > 0 
                            ? `${((sub.successful_calls / sub.total_calls) * 100).toFixed(1)}%`
                            : 'N/A'}
                        </p>
                      </div>
                      <div>
                        <p className="text-gray-500">Failed Calls</p>
                        <p className="font-medium">{sub.failed_calls.toLocaleString()}</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-12">
                <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
                <h3 className="mt-2 text-sm font-medium text-gray-900">No usage data</h3>
                <p className="mt-1 text-sm text-gray-500">Usage data will appear here once you start making API calls.</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Edit API Key Modal */}
      {showApiKeyModal && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Edit API Key Name</h3>
            <input
              type="text"
              value={newApiKeyName}
              onChange={(e) => setNewApiKeyName(e.target.value)}
              className="w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Enter new name"
            />
            <div className="mt-4 flex justify-end space-x-3">
              <button
                onClick={() => {
                  setShowApiKeyModal(false)
                  setSelectedApiKey(null)
                  setNewApiKeyName('')
                }}
                className="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-500"
              >
                Cancel
              </button>
              <button
                onClick={handleUpdateApiKeyName}
                className="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-md hover:bg-indigo-700"
              >
                Update
              </button>
            </div>
          </div>
        </div>
      )}
    </Layout>
  )
}

export default Dashboard
