import React, { useState } from 'react';
import Layout from '../../components/Layout';
import Link from 'next/link';

const ApiReference: React.FC = () => {
  const [activeSection, setActiveSection] = useState('overview');

  const sections = [
    { id: 'overview', title: 'Overview' },
    { id: 'apis', title: 'APIs Endpoints' },
    { id: 'subscriptions', title: 'Subscriptions' },
    { id: 'analytics', title: 'Analytics' },
    { id: 'webhooks', title: 'Webhooks' },
    { id: 'errors', title: 'Error Codes' },
  ];

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <Link href="/docs" className="text-blue-600 hover:text-blue-500 font-medium">
            ← Back to Documentation
          </Link>
        </div>

        <div className="mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">API Reference</h1>
          <p className="text-xl text-gray-600">
            Complete reference documentation for all available endpoints and parameters.
          </p>
        </div>

        <div className="flex flex-col lg:flex-row gap-8">
          {/* Sidebar Navigation */}
          <nav className="lg:w-64 flex-shrink-0">
            <div className="bg-white border border-gray-200 rounded-lg p-4 sticky top-8">
              <h3 className="font-semibold text-gray-900 mb-4">Sections</h3>
              <ul className="space-y-2">
                {sections.map((section) => (
                  <li key={section.id}>
                    <button
                      onClick={() => setActiveSection(section.id)}
                      className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${
                        activeSection === section.id
                          ? 'bg-blue-100 text-blue-700 font-medium'
                          : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                      }`}
                    >
                      {section.title}
                    </button>
                  </li>
                ))}
              </ul>
            </div>
          </nav>

          {/* Main Content */}
          <main className="flex-1">
            {activeSection === 'overview' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">API Overview</h2>
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Base URL</h3>
                    <code className="bg-gray-100 px-3 py-2 rounded text-sm">https://api.marketplace.com/v1</code>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Authentication</h3>
                    <p className="text-gray-600 mb-3">All API requests require authentication using Bearer tokens:</p>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`Authorization: Bearer YOUR_API_KEY`}
                    </pre>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Response Format</h3>
                    <p className="text-gray-600 mb-3">All responses are in JSON format:</p>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`{
  "status": "success",
  "data": { ... },
  "meta": {
    "timestamp": "2024-01-01T12:00:00Z",
    "rate_limit_remaining": 999
  }
}`}
                    </pre>
                  </div>
                </section>
              </div>
            )}

            {activeSection === 'apis' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">APIs Endpoints</h2>
                  
                  {/* List APIs */}
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">GET /apis</h3>
                    <p className="text-gray-600 mb-4">Retrieve a list of available APIs with filtering and pagination.</p>
                    
                    <h4 className="font-medium text-gray-900 mb-2">Query Parameters</h4>
                    <div className="overflow-x-auto">
                      <table className="min-w-full divide-y divide-gray-200 text-sm">
                        <thead className="bg-gray-50">
                          <tr>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Parameter</th>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Type</th>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Description</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200">
                          <tr>
                            <td className="px-4 py-2 font-mono">category</td>
                            <td className="px-4 py-2 text-gray-600">string</td>
                            <td className="px-4 py-2 text-gray-600">Filter by API category</td>
                          </tr>
                          <tr>
                            <td className="px-4 py-2 font-mono">page</td>
                            <td className="px-4 py-2 text-gray-600">integer</td>
                            <td className="px-4 py-2 text-gray-600">Page number (default: 1)</td>
                          </tr>
                          <tr>
                            <td className="px-4 py-2 font-mono">limit</td>
                            <td className="px-4 py-2 text-gray-600">integer</td>
                            <td className="px-4 py-2 text-gray-600">Items per page (max: 100)</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>

                    <h4 className="font-medium text-gray-900 mb-2 mt-4">Example Request</h4>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`curl -X GET "https://api.marketplace.com/v1/apis?category=AI/ML&page=1&limit=10" \\
  -H "Authorization: Bearer YOUR_API_KEY"`}
                    </pre>

                    <h4 className="font-medium text-gray-900 mb-2 mt-4">Example Response</h4>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`{
  "status": "success",
  "data": {
    "apis": [
      {
        "id": "api_12345",
        "name": "Image Recognition API",
        "description": "Advanced image analysis and object detection",
        "category": "AI/ML",
        "pricing": {
          "free_tier": true,
          "starting_price": 0.01
        },
        "rating": 4.8,
        "total_subscriptions": 1250
      }
    ],
    "total": 45,
    "page": 1,
    "limit": 10
  }
}`}
                    </pre>
                  </div>

                  {/* Get API Details */}
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">GET /apis/{"{id}"}</h3>
                    <p className="text-gray-600 mb-4">Get detailed information about a specific API.</p>
                    
                    <h4 className="font-medium text-gray-900 mb-2">Path Parameters</h4>
                    <div className="overflow-x-auto">
                      <table className="min-w-full divide-y divide-gray-200 text-sm">
                        <thead className="bg-gray-50">
                          <tr>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Parameter</th>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Type</th>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Description</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr>
                            <td className="px-4 py-2 font-mono">id</td>
                            <td className="px-4 py-2 text-gray-600">string</td>
                            <td className="px-4 py-2 text-gray-600">API identifier</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </div>

                  {/* Search APIs */}
                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">POST /apis/search</h3>
                    <p className="text-gray-600 mb-4">Advanced search with multiple filters and sorting options.</p>
                    
                    <h4 className="font-medium text-gray-900 mb-2">Request Body</h4>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`{
  "q": "search term",
  "category": "AI/ML",
  "tags": ["computer-vision", "ml"],
  "min_rating": 4.0,
  "has_free_tier": true,
  "sort_by": "rating",
  "page": 1,
  "limit": 20
}`}
                    </pre>
                  </div>
                </section>
              </div>
            )}

            {activeSection === 'subscriptions' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Subscriptions</h2>
                  
                  {/* List Subscriptions */}
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">GET /subscriptions</h3>
                    <p className="text-gray-600 mb-4">Get all active subscriptions for the authenticated user.</p>
                    
                    <h4 className="font-medium text-gray-900 mb-2">Example Response</h4>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`{
  "status": "success",
  "data": {
    "subscriptions": [
      {
        "id": "sub_12345",
        "api_id": "api_12345",
        "api_name": "Image Recognition API",
        "plan": "pro",
        "status": "active",
        "created_at": "2024-01-01T12:00:00Z",
        "usage": {
          "current_period_requests": 5420,
          "limit": 10000
        }
      }
    ]
  }
}`}
                    </pre>
                  </div>

                  {/* Create Subscription */}
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">POST /subscriptions</h3>
                    <p className="text-gray-600 mb-4">Subscribe to an API plan.</p>
                    
                    <h4 className="font-medium text-gray-900 mb-2">Request Body</h4>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`{
  "api_id": "api_12345",
  "plan": "pro"
}`}
                    </pre>
                  </div>

                  {/* Cancel Subscription */}
                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">DELETE /subscriptions/{"{id}"}</h3>
                    <p className="text-gray-600 mb-4">Cancel an active subscription.</p>
                    
                    <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4">
                      <p className="text-sm text-yellow-700">
                        <strong>Note:</strong> Cancellation takes effect at the end of the current billing period.
                      </p>
                    </div>
                  </div>
                </section>
              </div>
            )}

            {activeSection === 'analytics' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Analytics</h2>
                  
                  {/* Usage Stats */}
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">GET /analytics/usage</h3>
                    <p className="text-gray-600 mb-4">Get usage statistics for your API subscriptions.</p>
                    
                    <h4 className="font-medium text-gray-900 mb-2">Query Parameters</h4>
                    <div className="overflow-x-auto">
                      <table className="min-w-full divide-y divide-gray-200 text-sm">
                        <thead className="bg-gray-50">
                          <tr>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Parameter</th>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Type</th>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Description</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200">
                          <tr>
                            <td className="px-4 py-2 font-mono">start_date</td>
                            <td className="px-4 py-2 text-gray-600">string</td>
                            <td className="px-4 py-2 text-gray-600">Start date (ISO 8601)</td>
                          </tr>
                          <tr>
                            <td className="px-4 py-2 font-mono">end_date</td>
                            <td className="px-4 py-2 text-gray-600">string</td>
                            <td className="px-4 py-2 text-gray-600">End date (ISO 8601)</td>
                          </tr>
                          <tr>
                            <td className="px-4 py-2 font-mono">api_id</td>
                            <td className="px-4 py-2 text-gray-600">string</td>
                            <td className="px-4 py-2 text-gray-600">Filter by specific API</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </div>

                  {/* Billing Info */}
                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">GET /analytics/billing</h3>
                    <p className="text-gray-600 mb-4">Get billing information and cost breakdown.</p>
                    
                    <h4 className="font-medium text-gray-900 mb-2">Example Response</h4>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`{
  "status": "success",
  "data": {
    "current_period": {
      "start_date": "2024-01-01",
      "end_date": "2024-01-31",
      "total_cost": 125.50,
      "breakdown": [
        {
          "api_name": "Image Recognition API",
          "requests": 5420,
          "cost": 54.20
        }
      ]
    }
  }
}`}
                    </pre>
                  </div>
                </section>
              </div>
            )}

            {activeSection === 'webhooks' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Webhooks</h2>
                  
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Overview</h3>
                    <p className="text-gray-600 mb-4">
                      Webhooks allow you to receive real-time notifications about events in your account.
                    </p>
                    
                    <h4 className="font-medium text-gray-900 mb-2">Supported Events</h4>
                    <ul className="list-disc list-inside space-y-1 text-gray-600 text-sm">
                      <li><code>subscription.created</code> - New subscription created</li>
                      <li><code>subscription.cancelled</code> - Subscription cancelled</li>
                      <li><code>usage.limit_reached</code> - Usage limit reached</li>
                      <li><code>payment.successful</code> - Payment processed successfully</li>
                      <li><code>payment.failed</code> - Payment failed</li>
                    </ul>
                  </div>

                  {/* Create Webhook */}
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">POST /webhooks</h3>
                    <p className="text-gray-600 mb-4">Create a new webhook endpoint.</p>
                    
                    <h4 className="font-medium text-gray-900 mb-2">Request Body</h4>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`{
  "url": "https://your-app.com/webhooks",
  "events": ["subscription.created", "usage.limit_reached"],
  "secret": "your-webhook-secret"
}`}
                    </pre>
                  </div>

                  {/* Webhook Payload */}
                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Webhook Payload Format</h3>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`{
  "id": "evt_12345",
  "type": "subscription.created",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "subscription": {
      "id": "sub_12345",
      "api_id": "api_12345",
      "plan": "pro",
      "status": "active"
    }
  }
}`}
                    </pre>
                  </div>
                </section>
              </div>
            )}

            {activeSection === 'errors' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Error Codes</h2>
                  
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">HTTP Status Codes</h3>
                    <div className="overflow-x-auto">
                      <table className="min-w-full divide-y divide-gray-200 text-sm">
                        <thead className="bg-gray-50">
                          <tr>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Code</th>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Status</th>
                            <th className="px-4 py-2 text-left font-medium text-gray-500">Description</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200">
                          <tr>
                            <td className="px-4 py-2 font-mono">200</td>
                            <td className="px-4 py-2 text-green-600">OK</td>
                            <td className="px-4 py-2 text-gray-600">Request successful</td>
                          </tr>
                          <tr>
                            <td className="px-4 py-2 font-mono">400</td>
                            <td className="px-4 py-2 text-red-600">Bad Request</td>
                            <td className="px-4 py-2 text-gray-600">Invalid request parameters</td>
                          </tr>
                          <tr>
                            <td className="px-4 py-2 font-mono">401</td>
                            <td className="px-4 py-2 text-red-600">Unauthorized</td>
                            <td className="px-4 py-2 text-gray-600">Invalid or missing API key</td>
                          </tr>
                          <tr>
                            <td className="px-4 py-2 font-mono">403</td>
                            <td className="px-4 py-2 text-red-600">Forbidden</td>
                            <td className="px-4 py-2 text-gray-600">Access denied</td>
                          </tr>
                          <tr>
                            <td className="px-4 py-2 font-mono">404</td>
                            <td className="px-4 py-2 text-red-600">Not Found</td>
                            <td className="px-4 py-2 text-gray-600">Resource not found</td>
                          </tr>
                          <tr>
                            <td className="px-4 py-2 font-mono">429</td>
                            <td className="px-4 py-2 text-yellow-600">Too Many Requests</td>
                            <td className="px-4 py-2 text-gray-600">Rate limit exceeded</td>
                          </tr>
                          <tr>
                            <td className="px-4 py-2 font-mono">500</td>
                            <td className="px-4 py-2 text-red-600">Internal Server Error</td>
                            <td className="px-4 py-2 text-gray-600">Server error</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Error Response Format</h3>
                    <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`{
  "status": "error",
  "error": {
    "code": "INVALID_API_KEY",
    "message": "The provided API key is invalid or expired",
    "details": {
      "field": "authorization",
      "help_url": "https://docs.marketplace.com/authentication"
    }
  },
  "meta": {
    "timestamp": "2024-01-01T12:00:00Z",
    "request_id": "req_12345"
  }
}`}
                    </pre>
                  </div>
                </section>
              </div>
            )}
          </main>
        </div>

        {/* CTA */}
        <div className="mt-16 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl p-8 text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Start Building with Our APIs</h2>
          <p className="text-gray-600 mb-6">
            Ready to integrate? Create your account and get your first API key.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link 
              href="/auth/signup"
              className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 transition-colors"
            >
              Get Started Free
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

export default ApiReference;