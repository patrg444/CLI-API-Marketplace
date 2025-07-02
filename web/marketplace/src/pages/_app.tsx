import '@/styles/globals.css'
import type { AppProps } from 'next/app'
import { QueryClient, QueryClientProvider } from 'react-query'
import { loadStripe } from '@stripe/stripe-js'
import { Elements } from '@stripe/react-stripe-js'

// Mock authentication configuration
// In a real implementation, you would initialize your auth provider here
// For testing, we'll use localStorage to simulate user authentication

// Initialize Stripe (only if key is provided)
const stripePromise = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY 
  ? loadStripe(process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY)
  : null

// Initialize React Query
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
})

export default function App({ Component, pageProps }: AppProps) {
  const AppContent = () => <Component {...pageProps} />

  return (
    <QueryClientProvider client={queryClient}>
      {stripePromise ? (
        <Elements stripe={stripePromise}>
          <AppContent />
        </Elements>
      ) : (
        <AppContent />
      )}
    </QueryClientProvider>
  )
}
