/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  images: {
    domains: ['localhost', 'api-direct.com'],
  },
  env: {
    NEXT_PUBLIC_API_URL: process.env.REACT_APP_API_URL || 'http://localhost:8082',
    NEXT_PUBLIC_APIKEY_SERVICE_URL: process.env.REACT_APP_APIKEY_SERVICE_URL || 'http://localhost:8083',
    NEXT_PUBLIC_COGNITO_USER_POOL_ID: process.env.REACT_APP_COGNITO_USER_POOL_ID,
    NEXT_PUBLIC_COGNITO_CLIENT_ID: process.env.REACT_APP_COGNITO_CLIENT_ID,
    NEXT_PUBLIC_COGNITO_REGION: process.env.REACT_APP_COGNITO_REGION || 'us-east-1',
    NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY: process.env.REACT_APP_STRIPE_PUBLISHABLE_KEY,
  },
}

module.exports = nextConfig
