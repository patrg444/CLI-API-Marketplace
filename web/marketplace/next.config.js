/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  swcMinify: true,
  output: 'standalone', // Enable for Docker deployment
  poweredByHeader: false, // Remove X-Powered-By header for security
  
  // Performance optimizations
  experimental: {
    // optimizeCss: true,  // Removed due to critters dependency issue
    // gzipSize: true,
  },
  
  // Image optimization
  images: {
    domains: ['localhost', 'api-direct.com', 'api-marketplace.com'],
    formats: ['image/webp', 'image/avif'],
  },
  
  // Security headers
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
          {
            key: 'X-XSS-Protection',
            value: '1; mode=block',
          },
        ],
      },
    ];
  },
  
  // Environment variables
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8082',
    NEXT_PUBLIC_APIKEY_SERVICE_URL: process.env.NEXT_PUBLIC_APIKEY_SERVICE_URL || 'http://localhost:8083',
    NEXT_PUBLIC_BILLING_SERVICE_URL: process.env.NEXT_PUBLIC_BILLING_SERVICE_URL || 'http://localhost:8085',
    NEXT_PUBLIC_METERING_SERVICE_URL: process.env.NEXT_PUBLIC_METERING_SERVICE_URL || 'http://localhost:8084',
    NEXT_PUBLIC_MARKETPLACE_SERVICE_URL: process.env.NEXT_PUBLIC_MARKETPLACE_SERVICE_URL || 'http://localhost:8086',
    NEXT_PUBLIC_GATEWAY_URL: process.env.NEXT_PUBLIC_GATEWAY_URL || 'http://localhost:8082',
    NEXT_PUBLIC_AWS_REGION: process.env.NEXT_PUBLIC_AWS_REGION || 'us-east-1',
    NEXT_PUBLIC_AWS_USER_POOL_ID: process.env.NEXT_PUBLIC_AWS_USER_POOL_ID || '',
    NEXT_PUBLIC_AWS_USER_POOL_WEB_CLIENT_ID: process.env.NEXT_PUBLIC_AWS_USER_POOL_WEB_CLIENT_ID || '',
    NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY: process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY || '',
    NEXT_PUBLIC_GA_MEASUREMENT_ID: process.env.NEXT_PUBLIC_GA_MEASUREMENT_ID || '',
    NEXT_PUBLIC_MIXPANEL_TOKEN: process.env.NEXT_PUBLIC_MIXPANEL_TOKEN || '',
  },
}

module.exports = nextConfig
