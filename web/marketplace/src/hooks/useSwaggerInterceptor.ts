import { useEffect, useState } from 'react'
import { APIKey, Subscription } from '@/types/api'
import apiService from '@/services/api'

interface UseSwaggerInterceptorProps {
  subscriptionId?: string
  apiId: string
}

export const useSwaggerInterceptor = ({ subscriptionId, apiId }: UseSwaggerInterceptorProps) => {
  const [apiKey, setApiKey] = useState<APIKey | null>(null)
  const [isLoadingKey, setIsLoadingKey] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!subscriptionId) {
      setApiKey(null)
      return
    }

    const fetchApiKey = async () => {
      setIsLoadingKey(true)
      setError(null)
      
      try {
        // First, get all API keys
        const keys = await apiService.listAPIKeys()
        
        // Find the key associated with this subscription
        const subscriptionKey = keys.find(key => key.subscription_id === subscriptionId)
        
        if (subscriptionKey) {
          setApiKey(subscriptionKey)
        } else {
          // If no key exists for this subscription, we might need to create one
          // This depends on your business logic - you might want to create a key automatically
          // or prompt the user to create one
          setError('No API key found for this subscription')
        }
      } catch (err) {
        console.error('Error fetching API key:', err)
        setError('Failed to fetch API key')
      } finally {
        setIsLoadingKey(false)
      }
    }

    fetchApiKey()
  }, [subscriptionId])

  return {
    apiKey,
    isLoadingKey,
    error
  }
}

// Helper hook to get the API base URL based on the API ID
export const useAPIBaseUrl = (apiId: string): string => {
  // This should be configured based on your API gateway URL structure
  // For now, we'll use the gateway service URL from environment
  const gatewayUrl = process.env.NEXT_PUBLIC_GATEWAY_URL || 'http://localhost:8082'
  return `${gatewayUrl}/api/v1/apis/${apiId}`
}
