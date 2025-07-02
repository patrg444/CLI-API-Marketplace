import React, { useState } from 'react';
import Layout from '../../components/Layout';
import Link from 'next/link';

const Examples: React.FC = () => {
  const [activeExample, setActiveExample] = useState('authentication');

  const examples = [
    { id: 'authentication', title: 'Authentication', icon: '🔐' },
    { id: 'search', title: 'Search APIs', icon: '🔍' },
    { id: 'subscription', title: 'Manage Subscriptions', icon: '📋' },
    { id: 'webhooks', title: 'Handle Webhooks', icon: '🔗' },
    { id: 'analytics', title: 'Usage Analytics', icon: '📊' },
    { id: 'error-handling', title: 'Error Handling', icon: '⚠️' },
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
          <h1 className="text-4xl font-bold text-gray-900 mb-4">Code Examples</h1>
          <p className="text-xl text-gray-600">
            Practical examples and tutorials to help you implement APIs quickly.
          </p>
        </div>

        <div className="flex flex-col lg:flex-row gap-8">
          {/* Sidebar Navigation */}
          <nav className="lg:w-64 flex-shrink-0">
            <div className="bg-white border border-gray-200 rounded-lg p-4 sticky top-8">
              <h3 className="font-semibold text-gray-900 mb-4">Examples</h3>
              <ul className="space-y-2">
                {examples.map((example) => (
                  <li key={example.id}>
                    <button
                      onClick={() => setActiveExample(example.id)}
                      className={`w-full text-left flex items-center px-3 py-2 rounded-lg text-sm transition-colors ${
                        activeExample === example.id
                          ? 'bg-blue-100 text-blue-700 font-medium'
                          : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                      }`}
                    >
                      <span className="mr-3">{example.icon}</span>
                      {example.title}
                    </button>
                  </li>
                ))}
              </ul>
            </div>
          </nav>

          {/* Main Content */}
          <main className="flex-1">
            {activeExample === 'authentication' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Authentication Examples</h2>
                  
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Basic API Key Authentication</h3>
                    <div className="space-y-4">
                      <div>
                        <h4 className="font-medium text-gray-900 mb-2">JavaScript/Node.js</h4>
                        <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`const fetch = require('node-fetch');

async function authenticatedRequest() {
  const response = await fetch('https://api.marketplace.com/v1/apis', {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer ' + process.env.MARKETPLACE_API_KEY,
      'Content-Type': 'application/json',
      'User-Agent': 'MyApp/1.0.0'
    }
  });

  if (!response.ok) {
    throw new Error(\`HTTP error! status: \${response.status}\`);
  }

  const data = await response.json();
  return data;
}

// Usage
authenticatedRequest()
  .then(data => console.log('APIs:', data.data.apis))
  .catch(error => console.error('Error:', error));`}
                        </pre>
                      </div>

                      <div>
                        <h4 className="font-medium text-gray-900 mb-2">Python</h4>
                        <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`import requests
import os

def authenticated_request():
    headers = {
        'Authorization': f'Bearer {os.getenv("MARKETPLACE_API_KEY")}',
        'Content-Type': 'application/json',
        'User-Agent': 'MyApp/1.0.0'
    }
    
    response = requests.get('https://api.marketplace.com/v1/apis', headers=headers)
    response.raise_for_status()  # Raises an HTTPError for bad responses
    
    return response.json()

# Usage
try:
    data = authenticated_request()
    print(f"Found {len(data['data']['apis'])} APIs")
except requests.exceptions.RequestException as e:
    print(f"Error: {e}")`}
                        </pre>
                      </div>
                    </div>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Environment Variables Setup</h3>
                    <p className="text-gray-600 mb-4">Always store your API keys securely using environment variables:</p>
                    
                    <div className="space-y-4">
                      <div>
                        <h4 className="font-medium text-gray-900 mb-2">.env file</h4>
                        <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`MARKETPLACE_API_KEY=your_api_key_here
MARKETPLACE_ENVIRONMENT=production`}
                        </pre>
                      </div>

                      <div>
                        <h4 className="font-medium text-gray-900 mb-2">Loading in Node.js</h4>
                        <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`require('dotenv').config();

const apiKey = process.env.MARKETPLACE_API_KEY;
if (!apiKey) {
  throw new Error('MARKETPLACE_API_KEY environment variable is required');
}`}
                        </pre>
                      </div>
                    </div>
                  </div>
                </section>
              </div>
            )}

            {activeExample === 'search' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Search APIs Examples</h2>
                  
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Basic Search</h3>
                    <div className="space-y-4">
                      <div>
                        <h4 className="font-medium text-gray-900 mb-2">JavaScript</h4>
                        <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`async function searchAPIs(query, filters = {}) {
  const searchParams = new URLSearchParams({
    q: query,
    ...filters,
    limit: '20'
  });

  const response = await fetch(\`https://api.marketplace.com/v1/apis?\${searchParams}\`, {
    headers: {
      'Authorization': 'Bearer ' + process.env.MARKETPLACE_API_KEY,
      'Content-Type': 'application/json'
    }
  });

  const data = await response.json();
  return data;
}

// Usage examples
searchAPIs('image recognition')
  .then(data => console.log(\`Found \${data.total} APIs\`));

searchAPIs('payment', { category: 'Finance', has_free_tier: 'true' })
  .then(data => console.log('Payment APIs with free tier:', data.data.apis));`}
                        </pre>
                      </div>

                      <div>
                        <h4 className="font-medium text-gray-900 mb-2">Python</h4>
                        <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`import requests
from urllib.parse import urlencode

def search_apis(query, **filters):
    params = {
        'q': query,
        'limit': 20,
        **filters
    }
    
    url = f"https://api.marketplace.com/v1/apis?{urlencode(params)}"
    headers = {
        'Authorization': f'Bearer {os.getenv("MARKETPLACE_API_KEY")}',
        'Content-Type': 'application/json'
    }
    
    response = requests.get(url, headers=headers)
    response.raise_for_status()
    
    return response.json()

# Usage examples
data = search_apis('machine learning')
print(f"Found {data['total']} ML APIs")

# Search with filters
payment_apis = search_apis(
    'payment', 
    category='Finance', 
    has_free_tier='true',
    min_rating='4'
)
print(f"High-rated payment APIs: {len(payment_apis['data']['apis'])}")`}
                        </pre>
                      </div>
                    </div>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Advanced Search with Facets</h3>
                    <div>
                      <h4 className="font-medium text-gray-900 mb-2">POST Request for Complex Queries</h4>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`async function advancedSearch(searchQuery) {
  const response = await fetch('https://api.marketplace.com/v1/apis/search', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer ' + process.env.MARKETPLACE_API_KEY,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      q: searchQuery.term,
      category: searchQuery.category,
      tags: searchQuery.tags,
      min_rating: searchQuery.minRating,
      has_free_tier: searchQuery.hasFreeTier,
      sort_by: searchQuery.sortBy || 'rating',
      page: searchQuery.page || 1,
      limit: searchQuery.limit || 20
    })
  });

  return await response.json();
}

// Usage
const searchQuery = {
  term: 'computer vision',
  category: 'AI/ML',
  tags: ['image-processing', 'deep-learning'],
  minRating: 4.0,
  hasFreeTier: true,
  sortBy: 'popularity'
};

advancedSearch(searchQuery)
  .then(data => {
    console.log(\`Found \${data.total} APIs\`);
    console.log('Available facets:', data.facets);
  });`}
                      </pre>
                    </div>
                  </div>
                </section>
              </div>
            )}

            {activeExample === 'subscription' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Subscription Management</h2>
                  
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Subscribe to an API</h3>
                    <div>
                      <h4 className="font-medium text-gray-900 mb-2">JavaScript</h4>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`async function subscribeToAPI(apiId, plan = 'free') {
  try {
    const response = await fetch('https://api.marketplace.com/v1/subscriptions', {
      method: 'POST',
      headers: {
        'Authorization': 'Bearer ' + process.env.MARKETPLACE_API_KEY,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        api_id: apiId,
        plan: plan
      })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(\`Subscription failed: \${error.error.message}\`);
    }

    const subscription = await response.json();
    console.log('Successfully subscribed:', subscription.data);
    
    return subscription.data;
  } catch (error) {
    console.error('Subscription error:', error.message);
    throw error;
  }
}

// Usage
subscribeToAPI('api_12345', 'pro')
  .then(subscription => {
    console.log(\`Subscribed to \${subscription.api_name} - \${subscription.plan} plan\`);
  })
  .catch(error => console.error('Failed to subscribe:', error.message));`}
                      </pre>
                    </div>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">List Active Subscriptions</h3>
                    <div>
                      <h4 className="font-medium text-gray-900 mb-2">Python</h4>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`import requests
import os

def get_subscriptions():
    headers = {
        'Authorization': f'Bearer {os.getenv("MARKETPLACE_API_KEY")}',
        'Content-Type': 'application/json'
    }
    
    response = requests.get('https://api.marketplace.com/v1/subscriptions', headers=headers)
    response.raise_for_status()
    
    data = response.json()
    return data['data']['subscriptions']

def print_subscription_summary():
    try:
        subscriptions = get_subscriptions()
        
        print(f"Active Subscriptions: {len(subscriptions)}")
        print("-" * 40)
        
        for sub in subscriptions:
            usage = sub.get('usage', {})
            usage_percent = (usage.get('current_period_requests', 0) / 
                           usage.get('limit', 1)) * 100
            
            print(f"API: {sub['api_name']}")
            print(f"Plan: {sub['plan']}")
            print(f"Status: {sub['status']}")
            print(f"Usage: {usage.get('current_period_requests', 0):,} / {usage.get('limit', 0):,} ({usage_percent:.1f}%)")
            print("-" * 40)
            
    except requests.exceptions.RequestException as e:
        print(f"Error fetching subscriptions: {e}")

# Usage
print_subscription_summary()`}
                      </pre>
                    </div>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Cancel Subscription</h3>
                    <div>
                      <h4 className="font-medium text-gray-900 mb-2">JavaScript</h4>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`async function cancelSubscription(subscriptionId) {
  try {
    const response = await fetch(\`https://api.marketplace.com/v1/subscriptions/\${subscriptionId}\`, {
      method: 'DELETE',
      headers: {
        'Authorization': 'Bearer ' + process.env.MARKETPLACE_API_KEY,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(\`Cancellation failed: \${error.error.message}\`);
    }

    const result = await response.json();
    console.log('Subscription cancelled successfully');
    
    return result;
  } catch (error) {
    console.error('Cancellation error:', error.message);
    throw error;
  }
}

// Usage with confirmation
async function cancelWithConfirmation(subscriptionId, apiName) {
  const confirm = await new Promise(resolve => {
    // In a real app, you'd use a proper dialog
    const userConfirm = prompt(\`Are you sure you want to cancel subscription to \${apiName}? Type 'yes' to confirm:\`);
    resolve(userConfirm === 'yes');
  });

  if (confirm) {
    await cancelSubscription(subscriptionId);
    console.log(\`Subscription to \${apiName} has been cancelled\`);
  } else {
    console.log('Cancellation aborted');
  }
}`}
                      </pre>
                    </div>
                  </div>
                </section>
              </div>
            )}

            {activeExample === 'webhooks' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Webhook Handling</h2>
                  
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Create Webhook Endpoint</h3>
                    <div>
                      <h4 className="font-medium text-gray-900 mb-2">Express.js Server</h4>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`const express = require('express');
const crypto = require('crypto');

const app = express();
app.use(express.raw({ type: 'application/json' }));

// Webhook secret from your dashboard
const WEBHOOK_SECRET = process.env.MARKETPLACE_WEBHOOK_SECRET;

function verifyWebhookSignature(payload, signature, secret) {
  const hmac = crypto.createHmac('sha256', secret);
  const digest = 'sha256=' + hmac.update(payload).digest('hex');
  
  return crypto.timingSafeEqual(
    Buffer.from(digest),
    Buffer.from(signature)
  );
}

app.post('/webhooks/marketplace', (req, res) => {
  const signature = req.headers['x-marketplace-signature'];
  const payload = req.body;

  // Verify webhook signature
  if (!verifyWebhookSignature(payload, signature, WEBHOOK_SECRET)) {
    console.error('Invalid webhook signature');
    return res.status(401).send('Unauthorized');
  }

  const event = JSON.parse(payload);
  console.log('Received webhook:', event.type);

  // Handle different event types
  switch (event.type) {
    case 'subscription.created':
      handleSubscriptionCreated(event.data.subscription);
      break;
    case 'subscription.cancelled':
      handleSubscriptionCancelled(event.data.subscription);
      break;
    case 'usage.limit_reached':
      handleUsageLimitReached(event.data);
      break;
    case 'payment.successful':
      handlePaymentSuccessful(event.data.payment);
      break;
    case 'payment.failed':
      handlePaymentFailed(event.data.payment);
      break;
    default:
      console.log('Unhandled event type:', event.type);
  }

  res.status(200).send('OK');
});

function handleSubscriptionCreated(subscription) {
  console.log(\`New subscription: \${subscription.api_name} - \${subscription.plan}\`);
  // Send welcome email, update user permissions, etc.
}

function handleSubscriptionCancelled(subscription) {
  console.log(\`Subscription cancelled: \${subscription.api_name}\`);
  // Update user permissions, send cancellation email, etc.
}

function handleUsageLimitReached(data) {
  console.log(\`Usage limit reached for \${data.api_name}\`);
  // Send notification to user, suggest plan upgrade, etc.
}

function handlePaymentSuccessful(payment) {
  console.log(\`Payment successful: $\${payment.amount}\`);
  // Update billing records, send receipt, etc.
}

function handlePaymentFailed(payment) {
  console.log(\`Payment failed: $\${payment.amount}\`);
  // Send payment failure notification, retry payment, etc.
}

app.listen(3000, () => {
  console.log('Webhook server listening on port 3000');
});`}
                      </pre>
                    </div>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Register Webhook</h3>
                    <div>
                      <h4 className="font-medium text-gray-900 mb-2">JavaScript</h4>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`async function registerWebhook(url, events, secret) {
  const response = await fetch('https://api.marketplace.com/v1/webhooks', {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer ' + process.env.MARKETPLACE_API_KEY,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      url: url,
      events: events,
      secret: secret
    })
  });

  const webhook = await response.json();
  return webhook.data;
}

// Usage
const webhookEvents = [
  'subscription.created',
  'subscription.cancelled',
  'usage.limit_reached',
  'payment.successful',
  'payment.failed'
];

registerWebhook(
  'https://your-app.com/webhooks/marketplace',
  webhookEvents,
  'your-webhook-secret-key'
).then(webhook => {
  console.log('Webhook registered:', webhook.id);
}).catch(error => {
  console.error('Failed to register webhook:', error);
});`}
                      </pre>
                    </div>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Python Flask Example</h3>
                    <div>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`from flask import Flask, request, jsonify
import hashlib
import hmac
import json
import os

app = Flask(__name__)

WEBHOOK_SECRET = os.getenv('MARKETPLACE_WEBHOOK_SECRET')

def verify_signature(payload, signature, secret):
    expected_signature = 'sha256=' + hmac.new(
        secret.encode('utf-8'),
        payload,
        hashlib.sha256
    ).hexdigest()
    
    return hmac.compare_digest(expected_signature, signature)

@app.route('/webhooks/marketplace', methods=['POST'])
def handle_webhook():
    signature = request.headers.get('X-Marketplace-Signature')
    payload = request.get_data()
    
    if not verify_signature(payload, signature, WEBHOOK_SECRET):
        return jsonify({'error': 'Invalid signature'}), 401
    
    event = json.loads(payload)
    event_type = event['type']
    
    print(f"Received webhook: {event_type}")
    
    handlers = {
        'subscription.created': handle_subscription_created,
        'subscription.cancelled': handle_subscription_cancelled,
        'usage.limit_reached': handle_usage_limit_reached,
        'payment.successful': handle_payment_successful,
        'payment.failed': handle_payment_failed,
    }
    
    handler = handlers.get(event_type)
    if handler:
        handler(event['data'])
    else:
        print(f"Unhandled event type: {event_type}")
    
    return jsonify({'status': 'success'}), 200

def handle_subscription_created(data):
    subscription = data['subscription']
    print(f"New subscription: {subscription['api_name']} - {subscription['plan']}")
    # Your business logic here

def handle_subscription_cancelled(data):
    subscription = data['subscription']
    print(f"Subscription cancelled: {subscription['api_name']}")
    # Your business logic here

def handle_usage_limit_reached(data):
    print(f"Usage limit reached for {data['api_name']}")
    # Your business logic here

def handle_payment_successful(data):
    payment = data['payment']
    print(f"Payment successful: {payment['amount']}")
    # Your business logic here

def handle_payment_failed(data):
    payment = data['payment']
    print(f"Payment failed: {payment['amount']}")
    # Your business logic here

if __name__ == '__main__':
    app.run(debug=True, port=3000)`}
                      </pre>
                    </div>
                  </div>
                </section>
              </div>
            )}

            {activeExample === 'analytics' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Usage Analytics</h2>
                  
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Get Usage Statistics</h3>
                    <div>
                      <h4 className="font-medium text-gray-900 mb-2">JavaScript</h4>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`async function getUsageAnalytics(startDate, endDate, apiId = null) {
  const params = new URLSearchParams({
    start_date: startDate,
    end_date: endDate
  });
  
  if (apiId) {
    params.append('api_id', apiId);
  }

  const response = await fetch(\`https://api.marketplace.com/v1/analytics/usage?\${params}\`, {
    headers: {
      'Authorization': 'Bearer ' + process.env.MARKETPLACE_API_KEY,
      'Content-Type': 'application/json'
    }
  });

  const data = await response.json();
  return data.data;
}

async function displayUsageReport() {
  try {
    // Get usage for the last 30 days
    const endDate = new Date().toISOString().split('T')[0];
    const startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    
    const usage = await getUsageAnalytics(startDate, endDate);
    
    console.log('=== Usage Report (Last 30 Days) ===');
    console.log(\`Total Requests: \${usage.total_requests.toLocaleString()}\`);
    console.log(\`Total Cost: $\${usage.total_cost.toFixed(2)}\`);
    console.log(\`Average Daily Requests: \${Math.round(usage.total_requests / 30).toLocaleString()}\`);
    
    console.log('\\n=== API Breakdown ===');
    usage.by_api.forEach(api => {
      console.log(\`\${api.api_name}: \${api.requests.toLocaleString()} requests ($\${api.cost.toFixed(2)})\`);
    });
    
    if (usage.daily_breakdown) {
      console.log('\\n=== Daily Breakdown (Last 7 Days) ===');
      usage.daily_breakdown.slice(-7).forEach(day => {
        console.log(\`\${day.date}: \${day.requests.toLocaleString()} requests\`);
      });
    }
    
  } catch (error) {
    console.error('Error fetching usage analytics:', error);
  }
}

// Usage
displayUsageReport();`}
                      </pre>
                    </div>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Billing Information</h3>
                    <div>
                      <h4 className="font-medium text-gray-900 mb-2">Python</h4>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`import requests
from datetime import datetime, timedelta
import os

def get_billing_info():
    headers = {
        'Authorization': f'Bearer {os.getenv("MARKETPLACE_API_KEY")}',
        'Content-Type': 'application/json'
    }
    
    response = requests.get('https://api.marketplace.com/v1/analytics/billing', headers=headers)
    response.raise_for_status()
    
    return response.json()['data']

def format_billing_report():
    try:
        billing = get_billing_info()
        current_period = billing['current_period']
        
        print("=== Current Billing Period ===")
        print(f"Period: {current_period['start_date']} to {current_period['end_date']}")
        print(f"Total Cost: $\{current_period['total_cost']:.2f}")
        
        print("\\n=== Cost Breakdown ===")
        for item in current_period['breakdown']:
            print(f"{item['api_name']}: {item['requests']:,} requests = $\{item['cost']:.2f}")
        
        # Check if approaching billing limits
        if 'projected_cost' in billing:
            projected = billing['projected_cost']
            print(f"\\nProjected month-end cost: $\{projected:.2f}")
            
            if projected > current_period['total_cost'] * 1.5:
                print("⚠️  Warning: Usage is trending higher than usual")
        
        # Show cost per request for each API
        print("\\n=== Cost Efficiency ===")
        for item in current_period['breakdown']:
            if item['requests'] > 0:
                cost_per_request = item['cost'] / item['requests']
                print(f"{item['api_name']}: $\{cost_per_request:.4f} per request")
                
    except requests.exceptions.RequestException as e:
        print(f"Error fetching billing info: {e}")

# Usage
format_billing_report()`}
                      </pre>
                    </div>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Usage Monitoring & Alerts</h3>
                    <div>
                      <h4 className="font-medium text-gray-900 mb-2">Python Script for Monitoring</h4>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`import requests
import smtplib
from email.mime.text import MIMEText
from datetime import datetime
import os

def check_usage_limits():
    headers = {
        'Authorization': f'Bearer {os.getenv("MARKETPLACE_API_KEY")}',
        'Content-Type': 'application/json'
    }
    
    # Get current subscriptions and usage
    response = requests.get('https://api.marketplace.com/v1/subscriptions', headers=headers)
    subscriptions = response.json()['data']['subscriptions']
    
    alerts = []
    
    for sub in subscriptions:
        usage = sub.get('usage', {})
        current_requests = usage.get('current_period_requests', 0)
        limit = usage.get('limit', float('inf'))
        
        if limit > 0:
            usage_percent = (current_requests / limit) * 100
            
            # Check different threshold levels
            if usage_percent >= 90:
                alerts.append({
                    'level': 'critical',
                    'api': sub['api_name'],
                    'usage_percent': usage_percent,
                    'current': current_requests,
                    'limit': limit
                })
            elif usage_percent >= 75:
                alerts.append({
                    'level': 'warning',
                    'api': sub['api_name'],
                    'usage_percent': usage_percent,
                    'current': current_requests,
                    'limit': limit
                })
    
    return alerts

def send_alert_email(alerts):
    if not alerts:
        return
    
    smtp_server = os.getenv('SMTP_SERVER')
    smtp_port = int(os.getenv('SMTP_PORT', '587'))
    email_user = os.getenv('EMAIL_USER')
    email_pass = os.getenv('EMAIL_PASS')
    to_email = os.getenv('ALERT_EMAIL')
    
    if not all([smtp_server, email_user, email_pass, to_email]):
        print("Email configuration missing, printing alerts to console:")
        for alert in alerts:
            print(f"ALERT: {alert['api']} at {alert['usage_percent']:.1f}% usage")
        return
    
    subject = f"API Usage Alert - {len(alerts)} API(s) need attention"
    
    body = "API Usage Alert\\n\\n"
    for alert in alerts:
        body += f"API: {alert['api']}\\n"
        body += f"Usage: {alert['current']:,} / {alert['limit']:,} ({alert['usage_percent']:.1f}%)\\n"
        body += f"Level: {alert['level'].upper()}\\n\\n"
    
    msg = MIMEText(body)
    msg['Subject'] = subject
    msg['From'] = email_user
    msg['To'] = to_email
    
    try:
        server = smtplib.SMTP(smtp_server, smtp_port)
        server.starttls()
        server.login(email_user, email_pass)
        server.send_message(msg)
        server.quit()
        print(f"Alert email sent for {len(alerts)} APIs")
    except Exception as e:
        print(f"Failed to send email: {e}")

def main():
    print(f"Checking usage limits at {datetime.now()}")
    alerts = check_usage_limits()
    
    if alerts:
        print(f"Found {len(alerts)} usage alerts")
        send_alert_email(alerts)
    else:
        print("All APIs within normal usage limits")

if __name__ == "__main__":
    main()`}
                      </pre>
                    </div>
                  </div>
                </section>
              </div>
            )}

            {activeExample === 'error-handling' && (
              <div className="space-y-8">
                <section>
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Error Handling Best Practices</h2>
                  
                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Comprehensive Error Handling</h3>
                    <div>
                      <h4 className="font-medium text-gray-900 mb-2">JavaScript</h4>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`class MarketplaceAPIError extends Error {
  constructor(message, code, statusCode, details = null) {
    super(message);
    this.name = 'MarketplaceAPIError';
    this.code = code;
    this.statusCode = statusCode;
    this.details = details;
  }
}

class MarketplaceClient {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.baseURL = 'https://api.marketplace.com/v1';
  }

  async makeRequest(endpoint, options = {}) {
    const url = \`\${this.baseURL}\${endpoint}\`;
    const config = {
      headers: {
        'Authorization': \`Bearer \${this.apiKey}\`,
        'Content-Type': 'application/json',
        ...options.headers
      },
      ...options
    };

    try {
      const response = await fetch(url, config);
      
      // Handle rate limiting
      if (response.status === 429) {
        const retryAfter = response.headers.get('Retry-After');
        const resetTime = response.headers.get('X-RateLimit-Reset');
        
        throw new MarketplaceAPIError(
          'Rate limit exceeded',
          'RATE_LIMIT_EXCEEDED',
          429,
          { retryAfter, resetTime }
        );
      }
      
      // Handle authentication errors
      if (response.status === 401) {
        throw new MarketplaceAPIError(
          'Invalid API key or token expired',
          'UNAUTHORIZED',
          401
        );
      }
      
      // Handle forbidden access
      if (response.status === 403) {
        throw new MarketplaceAPIError(
          'Access forbidden - check your subscription',
          'FORBIDDEN',
          403
        );
      }
      
      // Parse response
      const data = await response.json();
      
      // Handle API-level errors
      if (!response.ok) {
        throw new MarketplaceAPIError(
          data.error?.message || 'API request failed',
          data.error?.code || 'UNKNOWN_ERROR',
          response.status,
          data.error?.details
        );
      }
      
      return data;
      
    } catch (error) {
      // Network or parsing errors
      if (!(error instanceof MarketplaceAPIError)) {
        throw new MarketplaceAPIError(
          'Network error or invalid response',
          'NETWORK_ERROR',
          0,
          { originalError: error.message }
        );
      }
      throw error;
    }
  }

  async searchAPIsWithRetry(query, maxRetries = 3, backoffMs = 1000) {
    let lastError;
    
    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        return await this.makeRequest('/apis', {
          method: 'GET',
          headers: {},
          // Add query params here
        });
        
      } catch (error) {
        lastError = error;
        
        // Don't retry on certain errors
        if (error.code === 'UNAUTHORIZED' || error.code === 'FORBIDDEN') {
          throw error;
        }
        
        // Exponential backoff for retryable errors
        if (attempt < maxRetries) {
          const delay = backoffMs * Math.pow(2, attempt - 1);
          console.log(\`Request failed (attempt \${attempt}/\${maxRetries}), retrying in \${delay}ms...\`);
          await new Promise(resolve => setTimeout(resolve, delay));
        }
      }
    }
    
    throw lastError;
  }
}

// Usage with proper error handling
async function searchAPIsExample() {
  const client = new MarketplaceClient(process.env.MARKETPLACE_API_KEY);
  
  try {
    const results = await client.searchAPIsWithRetry('machine learning');
    console.log(\`Found \${results.total} APIs\`);
    
  } catch (error) {
    if (error instanceof MarketplaceAPIError) {
      switch (error.code) {
        case 'RATE_LIMIT_EXCEEDED':
          console.error(\`Rate limited. Retry after: \${error.details.retryAfter} seconds\`);
          break;
        case 'UNAUTHORIZED':
          console.error('API key is invalid. Please check your credentials.');
          break;
        case 'FORBIDDEN':
          console.error('Access denied. Check your subscription status.');
          break;
        case 'NETWORK_ERROR':
          console.error('Network error. Please check your connection.');
          break;
        default:
          console.error(\`API Error (\${error.code}): \${error.message}\`);
      }
    } else {
      console.error('Unexpected error:', error);
    }
  }
}`}
                      </pre>
                    </div>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Python Error Handling</h3>
                    <div>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`import requests
import time
import logging
from typing import Optional, Dict, Any

class MarketplaceAPIError(Exception):
    def __init__(self, message: str, code: str, status_code: int, details: Optional[Dict] = None):
        super().__init__(message)
        self.code = code
        self.status_code = status_code
        self.details = details or {}

class MarketplaceClient:
    def __init__(self, api_key: str):
        self.api_key = api_key
        self.base_url = 'https://api.marketplace.com/v1'
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {api_key}',
            'Content-Type': 'application/json'
        })

    def make_request(self, endpoint: str, method: str = 'GET', **kwargs) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        
        try:
            response = self.session.request(method, url, **kwargs)
            
            # Handle rate limiting
            if response.status_code == 429:
                retry_after = response.headers.get('Retry-After')
                reset_time = response.headers.get('X-RateLimit-Reset')
                
                raise MarketplaceAPIError(
                    'Rate limit exceeded',
                    'RATE_LIMIT_EXCEEDED',
                    429,
                    {'retry_after': retry_after, 'reset_time': reset_time}
                )
            
            # Handle authentication errors
            if response.status_code == 401:
                raise MarketplaceAPIError(
                    'Invalid API key or token expired',
                    'UNAUTHORIZED',
                    401
                )
            
            # Handle forbidden access
            if response.status_code == 403:
                raise MarketplaceAPIError(
                    'Access forbidden - check your subscription',
                    'FORBIDDEN',
                    403
                )
            
            # Parse JSON response
            try:
                data = response.json()
            except ValueError as e:
                raise MarketplaceAPIError(
                    'Invalid JSON response',
                    'INVALID_RESPONSE',
                    response.status_code,
                    {'original_error': str(e)}
                )
            
            # Handle API-level errors
            if not response.ok:
                error_info = data.get('error', {})
                raise MarketplaceAPIError(
                    error_info.get('message', 'API request failed'),
                    error_info.get('code', 'UNKNOWN_ERROR'),
                    response.status_code,
                    error_info.get('details')
                )
            
            return data
            
        except requests.exceptions.RequestException as e:
            # Network errors
            raise MarketplaceAPIError(
                'Network error',
                'NETWORK_ERROR',
                0,
                {'original_error': str(e)}
            )

    def search_apis_with_retry(self, query: str, max_retries: int = 3, 
                              backoff_seconds: float = 1.0) -> Dict[str, Any]:
        last_error = None
        
        for attempt in range(1, max_retries + 1):
            try:
                return self.make_request('/apis', params={'q': query})
                
            except MarketplaceAPIError as error:
                last_error = error
                
                # Don't retry on certain errors
                if error.code in ['UNAUTHORIZED', 'FORBIDDEN']:
                    raise error
                
                # Exponential backoff for retryable errors
                if attempt < max_retries:
                    delay = backoff_seconds * (2 ** (attempt - 1))
                    logging.warning(f"Request failed (attempt {attempt}/{max_retries}), "
                                  f"retrying in {delay}s...")
                    time.sleep(delay)
        
        raise last_error

# Usage example with logging
def search_apis_example():
    logging.basicConfig(level=logging.INFO)
    client = MarketplaceClient(os.getenv('MARKETPLACE_API_KEY'))
    
    try:
        results = client.search_apis_with_retry('machine learning')
        logging.info(f"Found {results['total']} APIs")
        return results
        
    except MarketplaceAPIError as error:
        if error.code == 'RATE_LIMIT_EXCEEDED':
            retry_after = error.details.get('retry_after', 'unknown')
            logging.error(f"Rate limited. Retry after: {retry_after} seconds")
        elif error.code == 'UNAUTHORIZED':
            logging.error('API key is invalid. Please check your credentials.')
        elif error.code == 'FORBIDDEN':
            logging.error('Access denied. Check your subscription status.')
        elif error.code == 'NETWORK_ERROR':
            logging.error('Network error. Please check your connection.')
        else:
            logging.error(f"API Error ({error.code}): {error}")
            
    except Exception as error:
        logging.error(f"Unexpected error: {error}")

# Usage
search_apis_example()`}
                      </pre>
                    </div>
                  </div>

                  <div className="bg-white border border-gray-200 rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-3">Circuit Breaker Pattern</h3>
                    <p className="text-gray-600 mb-4">Implement circuit breaker to prevent cascading failures:</p>
                    <div>
                      <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm overflow-x-auto">
{`class CircuitBreaker {
  constructor(threshold = 5, timeout = 60000) {
    this.threshold = threshold; // Number of failures before opening
    this.timeout = timeout; // Time to wait before trying again (ms)
    this.failureCount = 0;
    this.state = 'CLOSED'; // CLOSED, OPEN, HALF_OPEN
    this.lastFailureTime = null;
  }

  async call(fn) {
    if (this.state === 'OPEN') {
      if (Date.now() - this.lastFailureTime > this.timeout) {
        this.state = 'HALF_OPEN';
      } else {
        throw new Error('Circuit breaker is OPEN');
      }
    }

    try {
      const result = await fn();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      throw error;
    }
  }

  onSuccess() {
    this.failureCount = 0;
    this.state = 'CLOSED';
  }

  onFailure() {
    this.failureCount++;
    this.lastFailureTime = Date.now();
    
    if (this.failureCount >= this.threshold) {
      this.state = 'OPEN';
    }
  }
}

// Usage with circuit breaker
const circuitBreaker = new CircuitBreaker(3, 30000); // 3 failures, 30s timeout

async function searchWithCircuitBreaker(query) {
  try {
    return await circuitBreaker.call(async () => {
      return await client.makeRequest('/apis', {
        method: 'GET',
        // Add query params
      });
    });
  } catch (error) {
    console.error('Circuit breaker prevented request or API failed:', error.message);
    throw error;
  }
}`}
                      </pre>
                    </div>
                  </div>
                </section>
              </div>
            )}
          </main>
        </div>

        {/* CTA */}
        <div className="mt-16 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl p-8 text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Ready to Implement?</h2>
          <p className="text-gray-600 mb-6">
            Use these examples as a starting point for your API integration.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link 
              href="/auth/signup"
              className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 transition-colors"
            >
              Get Your API Key
            </Link>
            <Link 
              href="/docs/sdks"
              className="inline-flex items-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors"
            >
              Download SDKs
            </Link>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default Examples;