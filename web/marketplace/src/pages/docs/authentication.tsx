import React from 'react';
import Layout from '../../components/Layout';
import Link from 'next/link';

const Authentication: React.FC = () => {
  return (
    <Layout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <Link href="/docs" className="text-blue-600 hover:text-blue-500 font-medium">
            ← Back to Documentation
          </Link>
        </div>

        <div className="mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">Authentication</h1>
          <p className="text-xl text-gray-600">
            Secure your API calls with proper authentication methods and API key management.
          </p>
        </div>

        {/* Table of Contents */}
        <div className="bg-gray-50 rounded-lg p-6 mb-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">In this guide</h2>
          <ul className="space-y-2">
            <li><a href="#api-keys" className="text-blue-600 hover:text-blue-500">1. API Key Authentication</a></li>
            <li><a href="#bearer-tokens" className="text-blue-600 hover:text-blue-500">2. Bearer Token Authentication</a></li>
            <li><a href="#oauth" className="text-blue-600 hover:text-blue-500">3. OAuth 2.0</a></li>
            <li><a href="#security" className="text-blue-600 hover:text-blue-500">4. Security Best Practices</a></li>
            <li><a href="#rate-limiting" className="text-blue-600 hover:text-blue-500">5. Rate Limiting</a></li>
            <li><a href="#troubleshooting" className="text-blue-600 hover:text-blue-500">6. Troubleshooting</a></li>
          </ul>
        </div>

        {/* API Keys */}
        <section id="api-keys" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">1. API Key Authentication</h2>
          <p className="text-gray-600 mb-6">
            API keys are the most common authentication method. Each API key is tied to your account and specific API subscriptions.
          </p>

          <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Generating API Keys</h3>
            <ol className="list-decimal list-inside space-y-2 text-gray-600 mb-4">
              <li>Navigate to your <Link href="/dashboard" className="text-blue-600 hover:text-blue-500">Dashboard</Link></li>
              <li>Go to the &quot;API Keys&quot; section</li>
              <li>Click &quot;Generate New Key&quot;</li>
              <li>Select the API(s) you want to access</li>
              <li>Copy and securely store your key</li>
            </ol>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Using API Keys</h3>
            <p className="text-gray-600 mb-4">Include your API key in the Authorization header:</p>
            
            <div className="space-y-4">
              <div>
                <h4 className="font-medium text-gray-900 mb-2">cURL Example</h4>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm">
{`curl -X GET "https://api.marketplace.com/v1/data" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json"`}
                </pre>
              </div>

              <div>
                <h4 className="font-medium text-gray-900 mb-2">JavaScript Example</h4>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm">
{`const response = await fetch('https://api.marketplace.com/v1/data', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  }
});

const data = await response.json();`}
                </pre>
              </div>

              <div>
                <h4 className="font-medium text-gray-900 mb-2">Python Example</h4>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm">
{`import requests

headers = {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
}

response = requests.get('https://api.marketplace.com/v1/data', headers=headers)
data = response.json()`}
                </pre>
              </div>
            </div>
          </div>
        </section>

        {/* Bearer Tokens */}
        <section id="bearer-tokens" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">2. Bearer Token Authentication</h2>
          <p className="text-gray-600 mb-6">
            Some APIs use temporary bearer tokens that expire after a certain period for enhanced security.
          </p>

          <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Token Lifecycle</h3>
            <div className="space-y-4">
              <div className="flex items-start">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center mr-3">
                  <span className="text-sm font-semibold text-blue-600">1</span>
                </div>
                <div>
                  <h4 className="font-medium text-gray-900">Request Token</h4>
                  <p className="text-gray-600 text-sm">Exchange your API key for a bearer token</p>
                </div>
              </div>
              <div className="flex items-start">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center mr-3">
                  <span className="text-sm font-semibold text-blue-600">2</span>
                </div>
                <div>
                  <h4 className="font-medium text-gray-900">Use Token</h4>
                  <p className="text-gray-600 text-sm">Include token in Authorization header for API calls</p>
                </div>
              </div>
              <div className="flex items-start">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center mr-3">
                  <span className="text-sm font-semibold text-blue-600">3</span>
                </div>
                <div>
                  <h4 className="font-medium text-gray-900">Refresh Token</h4>
                  <p className="text-gray-600 text-sm">Request new token before expiration</p>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Token Request Example</h3>
            <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm">
{`POST /v1/auth/token
Content-Type: application/json

{
  &quot;api_key&quot;: &quot;YOUR_API_KEY&quot;,
  &quot;grant_type&quot;: &quot;api_key&quot;
}

Response:
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}`}
            </pre>
          </div>
        </section>

        {/* OAuth */}
        <section id="oauth" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">3. OAuth 2.0</h2>
          <p className="text-gray-600 mb-6">
            For applications that need to access user data on their behalf, OAuth 2.0 provides secure authorization.
          </p>

          <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">OAuth Flow</h3>
            <div className="space-y-4">
              <div className="bg-blue-50 p-4 rounded-lg">
                <h4 className="font-medium text-blue-900 mb-2">Authorization Code Flow</h4>
                <ol className="list-decimal list-inside space-y-1 text-blue-800 text-sm">
                  <li>Direct user to authorization URL</li>
                  <li>User grants permission</li>
                  <li>Receive authorization code</li>
                  <li>Exchange code for access token</li>
                  <li>Use access token for API calls</li>
                </ol>
              </div>
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">OAuth Configuration</h3>
            <div className="grid md:grid-cols-2 gap-4 text-sm">
              <div>
                <p className="font-medium text-gray-900">Authorization URL:</p>
                <code className="text-blue-600">https://auth.marketplace.com/oauth/authorize</code>
              </div>
              <div>
                <p className="font-medium text-gray-900">Token URL:</p>
                <code className="text-blue-600">https://auth.marketplace.com/oauth/token</code>
              </div>
              <div>
                <p className="font-medium text-gray-900">Scopes:</p>
                <code className="text-blue-600">read, write, admin</code>
              </div>
              <div>
                <p className="font-medium text-gray-900">Response Type:</p>
                <code className="text-blue-600">code</code>
              </div>
            </div>
          </div>
        </section>

        {/* Security */}
        <section id="security" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">4. Security Best Practices</h2>
          
          <div className="grid md:grid-cols-2 gap-6 mb-6">
            <div className="bg-green-50 border border-green-200 rounded-lg p-6">
              <h3 className="text-lg font-semibold text-green-800 mb-3">✓ Do</h3>
              <ul className="space-y-2 text-green-700 text-sm">
                <li>Store API keys in environment variables</li>
                <li>Use HTTPS for all API calls</li>
                <li>Rotate keys regularly</li>
                <li>Implement proper error handling</li>
                <li>Monitor API usage</li>
                <li>Use least privilege access</li>
              </ul>
            </div>

            <div className="bg-red-50 border border-red-200 rounded-lg p-6">
              <h3 className="text-lg font-semibold text-red-800 mb-3">✗ Don&apos;t</h3>
              <ul className="space-y-2 text-red-700 text-sm">
                <li>Hardcode keys in source code</li>
                <li>Commit keys to version control</li>
                <li>Share keys publicly</li>
                <li>Use keys in client-side code</li>
                <li>Ignore authentication errors</li>
                <li>Use keys across environments</li>
              </ul>
            </div>
          </div>

          <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-yellow-700">
                  <strong>Security Alert:</strong> If you suspect your API key has been compromised, immediately revoke it from your dashboard and generate a new one.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Rate Limiting */}
        <section id="rate-limiting" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">5. Rate Limiting</h2>
          <p className="text-gray-600 mb-6">
            All APIs implement rate limiting to ensure fair usage and system stability.
          </p>

          <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Rate Limit Headers</h3>
            <p className="text-gray-600 mb-4">Every API response includes rate limit information:</p>
            <pre className="bg-gray-50 text-gray-800 p-4 rounded-lg text-sm">
{`X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200`}
            </pre>
          </div>

          <div className="grid md:grid-cols-3 gap-4">
            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <h4 className="font-semibold text-gray-900 mb-2">Free Tier</h4>
              <p className="text-2xl font-bold text-blue-600 mb-1">1,000</p>
              <p className="text-sm text-gray-600">requests/hour</p>
            </div>
            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <h4 className="font-semibold text-gray-900 mb-2">Pro Plan</h4>
              <p className="text-2xl font-bold text-blue-600 mb-1">10,000</p>
              <p className="text-sm text-gray-600">requests/hour</p>
            </div>
            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <h4 className="font-semibold text-gray-900 mb-2">Enterprise</h4>
              <p className="text-2xl font-bold text-blue-600 mb-1">Custom</p>
              <p className="text-sm text-gray-600">rate limits</p>
            </div>
          </div>
        </section>

        {/* Troubleshooting */}
        <section id="troubleshooting" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">6. Troubleshooting</h2>
          
          <div className="space-y-6">
            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-3">Common Authentication Errors</h3>
              <div className="space-y-4">
                <div>
                  <h4 className="font-medium text-red-600 mb-2">401 Unauthorized</h4>
                  <ul className="text-sm text-gray-600 space-y-1">
                    <li>• Check that your API key is correct</li>
                    <li>• Ensure you&apos;re using the Bearer token format</li>
                    <li>• Verify the API key hasn&apos;t expired</li>
                  </ul>
                </div>
                <div>
                  <h4 className="font-medium text-red-600 mb-2">403 Forbidden</h4>
                  <ul className="text-sm text-gray-600 space-y-1">
                    <li>• Check your subscription status</li>
                    <li>• Verify you have access to the endpoint</li>
                    <li>• Ensure your plan supports the feature</li>
                  </ul>
                </div>
                <div>
                  <h4 className="font-medium text-red-600 mb-2">429 Too Many Requests</h4>
                  <ul className="text-sm text-gray-600 space-y-1">
                    <li>• You&apos;ve exceeded your rate limit</li>
                    <li>• Wait before making more requests</li>
                    <li>• Consider upgrading your plan</li>
                  </ul>
                </div>
              </div>
            </div>

            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-3">Testing Authentication</h3>
              <p className="text-gray-600 mb-4">Use this endpoint to test your authentication:</p>
              <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`curl -X GET "https://api.marketplace.com/v1/auth/verify" \\
  -H "Authorization: Bearer YOUR_API_KEY"`}
              </pre>
              <p className="text-sm text-gray-600 mt-2">
                This will return your account information and confirm your authentication is working.
              </p>
            </div>
          </div>
        </section>

        {/* CTA */}
        <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl p-8 text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Need Help with Authentication?</h2>
          <p className="text-gray-600 mb-6">
            Our support team is here to help you implement secure authentication.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link 
              href="/docs/support"
              className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 transition-colors"
            >
              Contact Support
            </Link>
            <Link 
              href="/docs/examples"
              className="inline-flex items-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors"
            >
              View Examples
            </Link>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default Authentication;