/**
 * Integration tests for Dashboard functionality
 */

// Mock API client
jest.mock('../api-client-updated.js', () => {
  return class MockAPIClient {
    constructor() {
      this.token = 'mock-token';
    }
    
    getToken() {
      return this.token;
    }
    
    async getMe() {
      return {
        id: '123',
        name: 'Test User',
        email: 'test@example.com'
      };
    }
    
    async getDashboardStats() {
      return {
        metrics: {
          total_revenue_30d: 2500.75,
          total_calls_30d: 15000,
          active_deployments: 3,
          hosted_deployments: 2,
          byoa_deployments: 1
        },
        revenue_trend: [
          { date: '2024-01-01', revenue: 100 },
          { date: '2024-01-02', revenue: 150 },
          { date: '2024-01-03', revenue: 200 }
        ]
      };
    }
    
    async getMyAPIs() {
      return [
        {
          id: 'api-1',
          name: 'Weather API',
          status: 'running',
          deployment_type: 'hosted',
          endpoint_url: 'https://api.example.com/weather',
          calls_today: 500,
          updated_at: new Date().toISOString()
        },
        {
          id: 'api-2',
          name: 'Translation API',
          status: 'deploying',
          deployment_type: 'byoa',
          endpoint_url: null,
          calls_today: 0,
          updated_at: new Date().toISOString()
        }
      ];
    }
    
    connectWebSocket() {
      return {
        onmessage: null,
        close: jest.fn()
      };
    }
  };
});

describe('Dashboard Integration Tests', () => {
  let dashboard;
  
  beforeEach(() => {
    // Set up DOM
    document.body.innerHTML = `
      <div class="user-name"></div>
      <div class="user-email"></div>
      
      <div data-metric="revenue">
        <div class="text-3xl">$0</div>
        <div class="text-sm text-green-600">Loading...</div>
      </div>
      
      <div data-metric="calls">
        <div class="text-3xl">0</div>
        <div class="text-sm text-blue-600">Loading...</div>
      </div>
      
      <div data-metric="deployments">
        <div class="text-3xl">0</div>
        <div class="text-sm text-gray-500">Loading...</div>
      </div>
      
      <div id="recent-deployments"></div>
      <canvas id="revenueChart"></canvas>
    `;
    
    // Clear module cache
    jest.resetModules();
  });
  
  test('should load user information on page load', async () => {
    // Import dashboard module
    require('../dashboard-updated.js');
    
    // Wait for async operations
    await new Promise(resolve => setTimeout(resolve, 100));
    
    expect(document.querySelector('.user-name').textContent).toBe('Test User');
    expect(document.querySelector('.user-email').textContent).toBe('test@example.com');
  });
  
  test('should display dashboard metrics correctly', async () => {
    require('../dashboard-updated.js');
    
    // Wait for async operations
    await new Promise(resolve => setTimeout(resolve, 100));
    
    // Check revenue metric
    const revenueElement = document.querySelector('[data-metric="revenue"] .text-3xl');
    expect(revenueElement.textContent).toBe('$2500.75');
    
    // Check calls metric
    const callsElement = document.querySelector('[data-metric="calls"] .text-3xl');
    expect(callsElement.textContent).toBe('15,000');
    
    // Check deployments metric
    const deploymentsElement = document.querySelector('[data-metric="deployments"] .text-3xl');
    expect(deploymentsElement.textContent).toBe('3');
    
    // Check deployment details
    const deploymentDetails = document.querySelector('[data-metric="deployments"] .text-sm');
    expect(deploymentDetails.textContent).toContain('2 hosted, 1 BYOA');
  });
  
  test('should display recent deployments', async () => {
    require('../dashboard-updated.js');
    
    // Wait for async operations
    await new Promise(resolve => setTimeout(resolve, 100));
    
    const deploymentsList = document.getElementById('recent-deployments');
    const deployments = deploymentsList.querySelectorAll('.bg-gray-50');
    
    expect(deployments.length).toBe(2);
    
    // Check first deployment
    expect(deployments[0].textContent).toContain('Weather API');
    expect(deployments[0].textContent).toContain('Hosted');
    expect(deployments[0].textContent).toContain('500 calls today');
    
    // Check second deployment
    expect(deployments[1].textContent).toContain('Translation API');
    expect(deployments[1].textContent).toContain('BYOA');
    expect(deployments[1].textContent).toContain('Setting up...');
  });
  
  test('should show correct status indicators', async () => {
    require('../dashboard-updated.js');
    
    await new Promise(resolve => setTimeout(resolve, 100));
    
    const deployments = document.querySelectorAll('.bg-gray-50');
    
    // First API should have green status (running)
    const firstStatus = deployments[0].querySelector('.rounded-full');
    expect(firstStatus.classList.contains('bg-green-500')).toBe(true);
    
    // Second API should have yellow pulsing status (deploying)
    const secondStatus = deployments[1].querySelector('.rounded-full');
    expect(secondStatus.classList.contains('bg-yellow-500')).toBe(true);
    expect(secondStatus.classList.contains('animate-pulse')).toBe(true);
  });
  
  test('should handle empty deployments state', async () => {
    // Mock empty response
    const MockAPIClient = require('../api-client-updated.js');
    MockAPIClient.prototype.getMyAPIs = async () => [];
    
    require('../dashboard-updated.js');
    
    await new Promise(resolve => setTimeout(resolve, 100));
    
    const deploymentsList = document.getElementById('recent-deployments');
    expect(deploymentsList.textContent).toContain('No deployments yet');
    expect(deploymentsList.textContent).toContain('Deploy your first API');
    expect(deploymentsList.innerHTML).toContain('/deploy.html');
  });
});

describe('Real-time Updates', () => {
  let mockWebSocket;
  
  beforeEach(() => {
    document.body.innerHTML = `
      <div data-metric="revenue"><div class="text-3xl">$100</div></div>
      <div data-metric="calls"><div class="text-3xl">1,000</div></div>
      <div id="recent-deployments"></div>
    `;
    
    mockWebSocket = {
      onmessage: null,
      close: jest.fn()
    };
    
    // Mock WebSocket
    const MockAPIClient = require('../api-client-updated.js');
    MockAPIClient.prototype.connectWebSocket = () => mockWebSocket;
  });
  
  test('should update metrics via WebSocket', async () => {
    require('../dashboard-updated.js');
    
    await new Promise(resolve => setTimeout(resolve, 100));
    
    // Simulate WebSocket message
    mockWebSocket.onmessage({
      data: JSON.stringify({
        type: 'metrics_update',
        metric: 'revenue',
        value: 2600.50
      })
    });
    
    const revenueElement = document.querySelector('[data-metric="revenue"] .text-3xl');
    expect(revenueElement.textContent).toBe('$2600.50');
  });
  
  test('should increment calls counter on API call', async () => {
    require('../dashboard-updated.js');
    
    await new Promise(resolve => setTimeout(resolve, 100));
    
    // Set initial value
    const callsElement = document.querySelector('[data-metric="calls"] .text-3xl');
    callsElement.textContent = '1,000';
    
    // Simulate API call WebSocket message
    mockWebSocket.onmessage({
      data: JSON.stringify({
        type: 'api_call',
        api_id: 'test-api'
      })
    });
    
    // Should increment by 1
    expect(callsElement.textContent).toBe('1,001');
  });
  
  test('should show deployment notifications', async () => {
    // Mock notification function
    global.showNotification = jest.fn();
    
    require('../dashboard-updated.js');
    
    await new Promise(resolve => setTimeout(resolve, 100));
    
    // Simulate deployment update
    mockWebSocket.onmessage({
      data: JSON.stringify({
        type: 'deployment_update',
        api_name: 'New API',
        status: 'running'
      })
    });
    
    // Check if notification was shown
    expect(document.body.textContent).toContain('API New API is now running');
  });
});

// Export for Jest
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {};
}