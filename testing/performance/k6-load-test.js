import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const searchSuccessRate = new Rate('search_success');
const apiCallSuccessRate = new Rate('api_call_success');

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp up to 200 users
    { duration: '5m', target: 200 },  // Stay at 200 users
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'], // 95% of requests must complete below 200ms
    errors: ['rate<0.01'],            // Error rate must be below 1%
    search_success: ['rate>0.95'],    // Search success rate must be above 95%
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3001';
const API_GATEWAY_URL = __ENV.API_GATEWAY_URL || 'http://localhost:8082';

// Test data
const searchQueries = [
  'payment processing',
  'machine learning',
  'authentication',
  'data analytics',
  'image recognition',
  'stripe api',
  'weather data',
  'email service',
  'sms gateway',
  'translation api'
];

const categories = [
  'Financial Services',
  'AI/ML',
  'Communication',
  'Data & Analytics',
  'Developer Tools',
  'Security',
  'Media',
  'IoT',
  'Healthcare',
  'Education'
];

const priceRanges = ['free', 'low', 'medium', 'high'];

export default function () {
  // Randomly choose test scenario
  const scenario = Math.random();
  
  if (scenario < 0.4) {
    // 40% - Search functionality
    testSearch();
  } else if (scenario < 0.7) {
    // 30% - Browse with filters
    testBrowseWithFilters();
  } else if (scenario < 0.9) {
    // 20% - API details page
    testAPIDetails();
  } else {
    // 10% - API Gateway calls
    testAPIGateway();
  }
  
  sleep(1);
}

function testSearch() {
  const query = searchQueries[Math.floor(Math.random() * searchQueries.length)];
  
  const searchParams = new URLSearchParams({
    q: query,
    page: '1',
    limit: '20'
  });
  
  const response = http.post(
    `${BASE_URL}/api/v1/marketplace/search?${searchParams}`,
    JSON.stringify({
      query: query,
      filters: {}
    }),
    {
      headers: { 'Content-Type': 'application/json' },
    }
  );
  
  const success = check(response, {
    'search status is 200': (r) => r.status === 200,
    'search returns results': (r) => {
      const body = JSON.parse(r.body);
      return body.results && Array.isArray(body.results);
    },
    'search response time < 200ms': (r) => r.timings.duration < 200,
  });
  
  errorRate.add(!success);
  searchSuccessRate.add(success);
}

function testBrowseWithFilters() {
  const category = categories[Math.floor(Math.random() * categories.length)];
  const priceRange = priceRanges[Math.floor(Math.random() * priceRanges.length)];
  const minRating = Math.floor(Math.random() * 3) + 2; // 2-4
  
  const searchParams = new URLSearchParams({
    category: category,
    price_range: priceRange,
    min_rating: minRating.toString(),
    sort_by: 'rating',
    page: '1',
    limit: '20'
  });
  
  const response = http.get(`${BASE_URL}/api/v1/marketplace/apis?${searchParams}`);
  
  const success = check(response, {
    'browse status is 200': (r) => r.status === 200,
    'browse returns filtered results': (r) => {
      const body = JSON.parse(r.body);
      return body.apis && Array.isArray(body.apis);
    },
    'browse response time < 300ms': (r) => r.timings.duration < 300,
  });
  
  errorRate.add(!success);
}

function testAPIDetails() {
  // Simulate getting a random API ID (in real test, use actual IDs)
  const apiId = `api-${Math.floor(Math.random() * 100) + 1}`;
  
  const response = http.get(`${BASE_URL}/api/v1/marketplace/apis/${apiId}`);
  
  const success = check(response, {
    'API details status is 200': (r) => r.status === 200,
    'API details contains required fields': (r) => {
      if (r.status !== 200) return false;
      const body = JSON.parse(r.body);
      return body.id && body.name && body.description && body.pricing;
    },
    'API details response time < 150ms': (r) => r.timings.duration < 150,
  });
  
  errorRate.add(!success);
  
  // If successful, also load reviews
  if (success && response.status === 200) {
    testReviews(apiId);
  }
}

function testReviews(apiId) {
  const reviewParams = new URLSearchParams({
    sort_by: 'recent',
    page: '1',
    limit: '10'
  });
  
  const response = http.get(`${BASE_URL}/api/v1/marketplace/apis/${apiId}/reviews?${reviewParams}`);
  
  check(response, {
    'reviews status is 200': (r) => r.status === 200,
    'reviews response contains data': (r) => {
      const body = JSON.parse(r.body);
      return body.reviews && body.stats;
    },
    'reviews response time < 200ms': (r) => r.timings.duration < 200,
  });
}

function testAPIGateway() {
  // Simulate API key (in real test, use test API keys)
  const apiKey = 'test_api_key_' + Math.random().toString(36).substring(7);
  
  const response = http.post(
    `${API_GATEWAY_URL}/api/test-payment-api/v1/process`,
    JSON.stringify({
      amount: 100,
      currency: 'USD'
    }),
    {
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': apiKey,
      },
    }
  );
  
  const success = check(response, {
    'API gateway responds': (r) => r.status < 500,
    'API gateway response time < 100ms overhead': (r) => r.timings.duration < 200,
  });
  
  apiCallSuccessRate.add(success);
  errorRate.add(!success);
}

// Lifecycle hooks
export function setup() {
  console.log('Starting load test...');
  console.log(`Base URL: ${BASE_URL}`);
  console.log(`API Gateway URL: ${API_GATEWAY_URL}`);
}

export function teardown(data) {
  console.log('Load test completed');
}

// Custom summary
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'summary.json': JSON.stringify(data),
    'summary.html': htmlReport(data),
  };
}

function textSummary(data, options) {
  // Simple text summary
  return `
Load Test Results
=================
Total Requests: ${data.metrics.http_reqs.values.count}
Success Rate: ${(100 - data.metrics.errors.values.rate * 100).toFixed(2)}%
Average Response Time: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms
95th Percentile: ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms
Search Success Rate: ${(data.metrics.search_success.values.rate * 100).toFixed(2)}%
`;
}

function htmlReport(data) {
  // Generate HTML report
  return `
<!DOCTYPE html>
<html>
<head>
  <title>Load Test Results</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 40px; }
    .metric { margin: 20px 0; padding: 20px; background: #f5f5f5; }
    .success { color: green; }
    .fail { color: red; }
  </style>
</head>
<body>
  <h1>API-Direct Load Test Results</h1>
  <div class="metric">
    <h2>Overall Performance</h2>
    <p>Total Requests: ${data.metrics.http_reqs.values.count}</p>
    <p>Success Rate: <span class="${data.metrics.errors.values.rate < 0.01 ? 'success' : 'fail'}">${(100 - data.metrics.errors.values.rate * 100).toFixed(2)}%</span></p>
    <p>Average Response Time: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms</p>
    <p>95th Percentile: <span class="${data.metrics.http_req_duration.values['p(95)'] < 200 ? 'success' : 'fail'}">${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms</span></p>
  </div>
  <div class="metric">
    <h2>Search Performance</h2>
    <p>Success Rate: <span class="${data.metrics.search_success.values.rate > 0.95 ? 'success' : 'fail'}">${(data.metrics.search_success.values.rate * 100).toFixed(2)}%</span></p>
  </div>
</body>
</html>
`;
}
