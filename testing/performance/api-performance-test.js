import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const apiDuration = new Trend('api_duration');
const searchDuration = new Trend('search_duration');
const categoryFilterDuration = new Trend('category_filter_duration');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up to 10 users
    { duration: '1m', target: 50 },    // Ramp up to 50 users
    { duration: '2m', target: 100 },   // Stay at 100 users
    { duration: '1m', target: 200 },   // Spike to 200 users
    { duration: '2m', target: 100 },   // Back to 100 users
    { duration: '30s', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% of requests under 500ms
    http_req_failed: ['rate<0.05'],                  // Error rate under 5%
    errors: ['rate<0.05'],                           // Custom error rate under 5%
    api_duration: ['p(95)<300'],                     // API calls 95th percentile under 300ms
    search_duration: ['p(95)<400'],                  // Search 95th percentile under 400ms
  },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:3000/api';

// Helper function to make requests with timing
function timedRequest(url, metric) {
  const start = new Date();
  const res = http.get(url);
  const duration = new Date() - start;
  
  metric.add(duration);
  
  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'response has data': (r) => r.json('data') !== undefined,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  
  errorRate.add(!success);
  
  return res;
}

export default function () {
  // Test scenarios based on user behavior
  const scenario = Math.random();
  
  if (scenario < 0.3) {
    // 30% - Browse categories and APIs
    browseScenario();
  } else if (scenario < 0.6) {
    // 30% - Search for specific APIs
    searchScenario();
  } else if (scenario < 0.8) {
    // 20% - Filter by category
    categoryFilterScenario();
  } else if (scenario < 0.95) {
    // 15% - View specific API details
    apiDetailsScenario();
  } else {
    // 5% - Heavy user scenario (multiple operations)
    heavyUserScenario();
  }
  
  sleep(Math.random() * 3 + 1); // Random think time 1-4 seconds
}

function browseScenario() {
  // Get categories
  const categoriesRes = timedRequest(`${BASE_URL}/categories`, apiDuration);
  
  check(categoriesRes, {
    'categories loaded': (r) => r.json('data.length') > 0,
  });
  
  // Get featured APIs
  timedRequest(`${BASE_URL}/apis/featured`, apiDuration);
  
  // Get trending APIs
  timedRequest(`${BASE_URL}/apis/trending`, apiDuration);
  
  // Browse first page of APIs
  timedRequest(`${BASE_URL}/apis?page=1&limit=10`, apiDuration);
}

function searchScenario() {
  const searchTerms = ['weather', 'ai', 'gpt', 'crypto', 'translation', 'stock'];
  const searchTerm = searchTerms[Math.floor(Math.random() * searchTerms.length)];
  
  // Search for APIs
  const searchRes = timedRequest(
    `${BASE_URL}/apis?search=${searchTerm}`,
    searchDuration
  );
  
  check(searchRes, {
    'search returns results': (r) => r.json('data') !== null,
    'search metadata present': (r) => r.json('meta.total') !== undefined,
  });
  
  // If results found, view one
  const results = searchRes.json('data');
  if (results && results.length > 0) {
    const randomApi = results[Math.floor(Math.random() * results.length)];
    timedRequest(`${BASE_URL}/apis/${randomApi.id}`, apiDuration);
  }
}

function categoryFilterScenario() {
  const categories = ['ai-ml', 'data', 'finance', 'weather'];
  const category = categories[Math.floor(Math.random() * categories.length)];
  
  // Filter by category
  const filterRes = timedRequest(
    `${BASE_URL}/apis?category=${category}&limit=20`,
    categoryFilterDuration
  );
  
  check(filterRes, {
    'category filter works': (r) => {
      const data = r.json('data');
      return data && data.every(api => api.category === category);
    },
  });
  
  // Sort filtered results
  const sortOptions = ['rating', 'popular', 'price-low', 'newest'];
  const sort = sortOptions[Math.floor(Math.random() * sortOptions.length)];
  
  timedRequest(
    `${BASE_URL}/apis?category=${category}&sort=${sort}`,
    categoryFilterDuration
  );
}

function apiDetailsScenario() {
  const apiIds = [
    'sentiment-analyzer-pro',
    'global-weather-api',
    'gpt-4-turbo-wrapper',
    'stock-market-predictor',
    'crypto-analytics-api'
  ];
  
  const apiId = apiIds[Math.floor(Math.random() * apiIds.length)];
  
  // Get API details
  const detailsRes = timedRequest(`${BASE_URL}/apis/${apiId}`, apiDuration);
  
  check(detailsRes, {
    'API details loaded': (r) => r.json('data.id') === apiId,
    'API has required fields': (r) => {
      const data = r.json('data');
      return data && data.name && data.description && data.pricing;
    },
  });
}

function heavyUserScenario() {
  // Simulate a power user making multiple requests
  
  // Get all categories
  timedRequest(`${BASE_URL}/categories`, apiDuration);
  
  // Search multiple times
  const searches = ['ai', 'weather', 'crypto'];
  searches.forEach(term => {
    timedRequest(`${BASE_URL}/apis?search=${term}&limit=5`, searchDuration);
    sleep(0.5);
  });
  
  // Browse with different filters
  timedRequest(`${BASE_URL}/apis?sort=rating&limit=10`, apiDuration);
  timedRequest(`${BASE_URL}/apis?sort=price-low&maxPrice=0.005`, apiDuration);
  
  // Paginate through results
  for (let page = 1; page <= 3; page++) {
    timedRequest(`${BASE_URL}/apis?page=${page}&limit=10`, apiDuration);
    sleep(0.3);
  }
  
  // View multiple API details
  const detailIds = ['global-weather-api', 'gpt-4-turbo-wrapper'];
  detailIds.forEach(id => {
    timedRequest(`${BASE_URL}/apis/${id}`, apiDuration);
    sleep(0.5);
  });
}

// Handle test summary
export function handleSummary(data) {
  return {
    'performance-report.html': htmlReport(data),
    'performance-summary.json': JSON.stringify(data, null, 2),
  };
}

function htmlReport(data) {
  const metrics = data.metrics;
  
  return `
<!DOCTYPE html>
<html>
<head>
    <title>API Performance Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .metric { margin: 20px 0; padding: 15px; background: #f5f5f5; border-radius: 5px; }
        .pass { color: green; }
        .fail { color: red; }
        .chart { margin: 20px 0; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
    </style>
</head>
<body>
    <h1>API Performance Test Report</h1>
    <p>Test Duration: ${new Date(data.state.testRunDurationMs).toISOString()}</p>
    
    <h2>Summary</h2>
    <div class="metric">
        <h3>HTTP Request Duration</h3>
        <p>95th percentile: ${metrics.http_req_duration.values['p(95)'].toFixed(2)}ms</p>
        <p>99th percentile: ${metrics.http_req_duration.values['p(99)'].toFixed(2)}ms</p>
        <p>Average: ${metrics.http_req_duration.values.avg.toFixed(2)}ms</p>
    </div>
    
    <div class="metric">
        <h3>Error Rate</h3>
        <p class="${metrics.http_req_failed.values.rate < 0.05 ? 'pass' : 'fail'}">
            ${(metrics.http_req_failed.values.rate * 100).toFixed(2)}%
        </p>
    </div>
    
    <div class="metric">
        <h3>API Endpoint Performance</h3>
        <table>
            <tr>
                <th>Metric</th>
                <th>P95 (ms)</th>
                <th>Average (ms)</th>
            </tr>
            <tr>
                <td>General API Calls</td>
                <td>${metrics.api_duration.values['p(95)'].toFixed(2)}</td>
                <td>${metrics.api_duration.values.avg.toFixed(2)}</td>
            </tr>
            <tr>
                <td>Search Operations</td>
                <td>${metrics.search_duration.values['p(95)'].toFixed(2)}</td>
                <td>${metrics.search_duration.values.avg.toFixed(2)}</td>
            </tr>
            <tr>
                <td>Category Filtering</td>
                <td>${metrics.category_filter_duration.values['p(95)'].toFixed(2)}</td>
                <td>${metrics.category_filter_duration.values.avg.toFixed(2)}</td>
            </tr>
        </table>
    </div>
    
    <div class="metric">
        <h3>Throughput</h3>
        <p>Total Requests: ${metrics.http_reqs.values.count}</p>
        <p>Requests/sec: ${metrics.http_reqs.values.rate.toFixed(2)}</p>
    </div>
    
    <h2>Recommendations</h2>
    <ul>
        ${generateRecommendations(metrics)}
    </ul>
</body>
</html>
  `;
}

function generateRecommendations(metrics) {
  const recommendations = [];
  
  if (metrics.http_req_duration.values['p(95)'] > 500) {
    recommendations.push('<li>Consider implementing caching for frequently accessed endpoints</li>');
  }
  
  if (metrics.search_duration.values['p(95)'] > 400) {
    recommendations.push('<li>Optimize search queries or implement search result caching</li>');
  }
  
  if (metrics.http_req_failed.values.rate > 0.01) {
    recommendations.push('<li>Investigate error sources and improve error handling</li>');
  }
  
  if (recommendations.length === 0) {
    recommendations.push('<li class="pass">Performance meets all targets!</li>');
  }
  
  return recommendations.join('\n');
}