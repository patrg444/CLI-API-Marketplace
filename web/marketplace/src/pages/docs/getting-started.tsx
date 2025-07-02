import React from 'react';
import Layout from '../../components/Layout';
import Link from 'next/link';

const GettingStarted: React.FC = () => {
  return (
    <Layout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <Link href="/docs" className="text-blue-600 hover:text-blue-500 font-medium">
            ← Back to Documentation
          </Link>
        </div>

        <div className="mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">Getting Started</h1>
          <p className="text-xl text-gray-600">
            Learn the basics of using our API marketplace and how to get your first API key.
          </p>
        </div>

        {/* Table of Contents */}
        <div className="bg-gray-50 rounded-lg p-6 mb-8">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">In this guide</h2>
          <ul className="space-y-2">
            <li><a href="#account-setup" className="text-blue-600 hover:text-blue-500">1. Account Setup</a></li>
            <li><a href="#api-discovery" className="text-blue-600 hover:text-blue-500">2. API Discovery</a></li>
            <li><a href="#subscription" className="text-blue-600 hover:text-blue-500">3. API Subscription</a></li>
            <li><a href="#api-keys" className="text-blue-600 hover:text-blue-500">4. API Keys</a></li>
            <li><a href="#first-request" className="text-blue-600 hover:text-blue-500">5. Making Your First Request</a></li>
            <li><a href="#next-steps" className="text-blue-600 hover:text-blue-500">6. Next Steps</a></li>
          </ul>
        </div>

        {/* Account Setup */}
        <section id="account-setup" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">1. Account Setup</h2>
          <p className="text-gray-600 mb-6">
            Getting started with our API marketplace is simple. Follow these steps to create your developer account.
          </p>
          
          <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Sign Up Process</h3>
            <ol className="list-decimal list-inside space-y-3 text-gray-600">
              <li>Visit the <Link href="/auth/signup" className="text-blue-600 hover:text-blue-500">sign-up page</Link></li>
              <li>Enter your email address and create a secure password</li>
              <li>Verify your email address through the confirmation link</li>
              <li>Complete your developer profile with your organization details</li>
              <li>Choose your preferred subscription plan (free tier available)</li>
            </ol>
          </div>

          <div className="bg-blue-50 border-l-4 border-blue-400 p-4 mb-6">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-blue-700">
                  <strong>Free Tier:</strong> Start with our generous free tier that includes access to most APIs with reasonable rate limits.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* API Discovery */}
        <section id="api-discovery" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">2. API Discovery</h2>
          <p className="text-gray-600 mb-6">
            Explore our comprehensive catalog of APIs to find the perfect tools for your project.
          </p>

          <div className="grid md:grid-cols-2 gap-6 mb-6">
            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-3">Browse by Category</h3>
              <p className="text-gray-600 mb-4">
                Use our category filters to quickly find APIs in specific domains like AI/ML, Finance, or E-commerce.
              </p>
              <Link href="/#search-section" className="text-blue-600 hover:text-blue-500 font-medium">
                Browse Categories →
              </Link>
            </div>

            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-3">Search & Filter</h3>
              <p className="text-gray-600 mb-4">
                Use advanced search filters to find APIs by features, pricing, rating, and more.
              </p>
              <Link href="/" className="text-blue-600 hover:text-blue-500 font-medium">
                Start Searching →
              </Link>
            </div>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Popular API Categories</h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div className="text-gray-600">• AI/ML</div>
              <div className="text-gray-600">• Analytics</div>
              <div className="text-gray-600">• Authentication</div>
              <div className="text-gray-600">• Communication</div>
              <div className="text-gray-600">• Data</div>
              <div className="text-gray-600">• E-commerce</div>
              <div className="text-gray-600">• Finance</div>
              <div className="text-gray-600">• Maps & Location</div>
            </div>
          </div>
        </section>

        {/* Subscription */}
        <section id="subscription" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">3. API Subscription</h2>
          <p className="text-gray-600 mb-6">
            Subscribe to APIs that match your needs and budget.
          </p>

          <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Subscription Process</h3>
            <ol className="list-decimal list-inside space-y-3 text-gray-600">
              <li>Click on an API card to view detailed information</li>
              <li>Review pricing plans, features, and documentation</li>
              <li>Select the plan that best fits your usage needs</li>
              <li>Complete the subscription process</li>
              <li>Access your subscribed APIs from your dashboard</li>
            </ol>
          </div>

          <div className="grid md:grid-cols-3 gap-4">
            <div className="bg-green-50 border border-green-200 rounded-lg p-4">
              <h4 className="font-semibold text-green-800 mb-2">Free Tier</h4>
              <p className="text-sm text-green-700">Perfect for development and testing</p>
            </div>
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h4 className="font-semibold text-blue-800 mb-2">Pro Plans</h4>
              <p className="text-sm text-blue-700">Higher limits for production use</p>
            </div>
            <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
              <h4 className="font-semibold text-purple-800 mb-2">Enterprise</h4>
              <p className="text-sm text-purple-700">Custom solutions for scale</p>
            </div>
          </div>
        </section>

        {/* API Keys */}
        <section id="api-keys" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">4. API Keys</h2>
          <p className="text-gray-600 mb-6">
            Generate and manage your API keys for secure access to subscribed services.
          </p>

          <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Managing API Keys</h3>
            <ol className="list-decimal list-inside space-y-3 text-gray-600 mb-4">
              <li>Navigate to your <Link href="/dashboard" className="text-blue-600 hover:text-blue-500">Dashboard</Link></li>
              <li>Go to the {"\"API Keys\""} section</li>
              <li>Click {"\"Generate New Key\""} for each subscribed API</li>
              <li>Copy and securely store your API keys</li>
              <li>Use keys in your application&apos;s authorization headers</li>
            </ol>
          </div>

          <div className="bg-red-50 border-l-4 border-red-400 p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-red-700">
                  <strong>Security:</strong> Never share your API keys publicly or commit them to version control. Use environment variables instead.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* First Request */}
        <section id="first-request" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">5. Making Your First Request</h2>
          <p className="text-gray-600 mb-6">
            Test your API integration with a simple request.
          </p>

          <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Example API Call</h3>
            <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm">
{`curl -X GET "https://api.marketplace.com/v1/example" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json"`}
            </pre>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-3">Response Format</h3>
            <pre className="bg-gray-50 text-gray-800 p-4 rounded-lg overflow-x-auto text-sm">
{`{
  "status": "success",
  "data": {
    "message": "Hello from the API!",
    "timestamp": "2024-01-01T12:00:00Z"
  },
  "meta": {
    "rate_limit_remaining": 999
  }
}`}
            </pre>
          </div>
        </section>

        {/* Next Steps */}
        <section id="next-steps" className="mb-12">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">6. Next Steps</h2>
          <p className="text-gray-600 mb-6">
            Now that you&apos;ve made your first API call, here&apos;s what to explore next.
          </p>

          <div className="grid md:grid-cols-2 gap-6">
            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-3">Learn More</h3>
              <ul className="space-y-2 text-gray-600">
                <li><Link href="/docs/authentication" className="text-blue-600 hover:text-blue-500">Authentication Methods</Link></li>
                <li><Link href="/docs/api-reference" className="text-blue-600 hover:text-blue-500">API Reference</Link></li>
                <li><Link href="/docs/examples" className="text-blue-600 hover:text-blue-500">Code Examples</Link></li>
                <li><Link href="/docs/sdks" className="text-blue-600 hover:text-blue-500">SDKs & Libraries</Link></li>
              </ul>
            </div>

            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-3">Get Help</h3>
              <ul className="space-y-2 text-gray-600">
                <li><Link href="/docs/support" className="text-blue-600 hover:text-blue-500">Support Center</Link></li>
                <li><Link href="/community" className="text-blue-600 hover:text-blue-500">Developer Community</Link></li>
                <li><Link href="/contact" className="text-blue-600 hover:text-blue-500">Contact Support</Link></li>
              </ul>
            </div>
          </div>
        </section>

        {/* CTA */}
        <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl p-8 text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Ready to Start Building?</h2>
          <p className="text-gray-600 mb-6">
            Join thousands of developers building amazing applications with our APIs.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link 
              href="/auth/signup"
              className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 transition-colors"
            >
              Create Free Account
            </Link>
            <Link 
              href="/"
              className="inline-flex items-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors"
            >
              Browse APIs
            </Link>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default GettingStarted;