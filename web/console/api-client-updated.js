// Updated API Client for Console Dashboard - Matches actual backend endpoints
const API_BASE_URL = window.location.hostname === 'localhost' 
  ? 'http://localhost:8000' 
  : 'https://api.api-direct.io';

class APIClient {
  constructor() {
    this.baseURL = API_BASE_URL;
    this.token = null;
  }

  // Set authentication token
  setToken(token) {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('api_token', token);
    }
  }

  // Get token from storage
  getToken() {
    if (this.token) return this.token;
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('api_token');
    }
    return this.token;
  }

  // Clear token
  clearToken() {
    this.token = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('api_token');
    }
  }

  // Make authenticated request
  async request(endpoint, options = {}) {
    const token = this.getToken();
    
    const config = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` }),
        ...options.headers,
      },
    };

    try {
      const response = await fetch(`${this.baseURL}${endpoint}`, config);
      
      if (response.status === 401) {
        this.clearToken();
        window.location.href = '/login.html';
        throw new Error('Unauthorized');
      }

      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        throw new Error(error.detail || `HTTP error! status: ${response.status}`);
      }

      return response.json();
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  }

  // Auth endpoints - matching backend
  async login(email, password) {
    const response = await this.request('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
    
    if (response.access_token) {
      this.setToken(response.access_token);
    }
    
    return response;
  }

  async register(data) {
    const response = await this.request('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    
    if (response.access_token) {
      this.setToken(response.access_token);
    }
    
    return response;
  }

  async getMe() {
    return this.request('/api/auth/me');
  }

  // Dashboard endpoints - matching backend
  async getDashboardStats() {
    return this.request('/api/dashboard/overview');
  }

  async getRecentDeployments() {
    return this.request('/api/dashboard/recent-deployments');
  }

  // Analytics endpoints - matching backend
  async getAnalytics(apiId = null, period = '30d') {
    const endpoint = apiId 
      ? `/api/analytics/usage-by-consumer?api_id=${apiId}&period=${period}`
      : `/api/analytics/usage-by-consumer?period=${period}`;
    return this.request(endpoint);
  }

  async getGeographicAnalytics(apiId = null, period = '30d') {
    const endpoint = apiId
      ? `/api/analytics/geographic?api_id=${apiId}&period=${period}`
      : `/api/analytics/geographic?period=${period}`;
    return this.request(endpoint);
  }

  async getErrorAnalytics(apiId = null, period = '7d') {
    const endpoint = apiId
      ? `/api/analytics/errors?api_id=${apiId}&period=${period}`
      : `/api/analytics/errors?period=${period}`;
    return this.request(endpoint);
  }

  async getRevenueAnalytics(period = '30d') {
    return this.request(`/api/analytics/revenue?period=${period}`);
  }

  // API management - matching backend
  async getMyAPIs() {
    return this.request('/api/my-apis');
  }

  async getAPIDetails(apiId) {
    return this.request(`/api/my-apis/${apiId}`);
  }

  async deployAPI(data) {
    return this.request('/api/deploy', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateAPI(apiId, data) {
    return this.request(`/api/my-apis/${apiId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteAPI(apiId) {
    return this.request(`/api/my-apis/${apiId}`, {
      method: 'DELETE',
    });
  }

  // API Key management - matching backend
  async getAPIKeys() {
    return this.request('/api/keys');
  }

  async createAPIKey(data) {
    return this.request('/api/keys', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async revokeAPIKey(keyId) {
    return this.request(`/api/keys/${keyId}`, {
      method: 'DELETE',
    });
  }

  // Marketplace endpoints
  async getMarketplaceListings(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    return this.request(`/api/marketplace/listings?${queryString}`);
  }

  async publishToMarketplace(data) {
    return this.request('/api/marketplace/publish', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  // Trial endpoints
  async startTrial(apiId) {
    return this.request(`/api/trials/start?api_id=${apiId}`, {
      method: 'POST',
    });
  }

  async getTrialStatus(apiId) {
    return this.request(`/api/trials/${apiId}/status`);
  }

  // Subscription endpoints
  async getSubscriptionPlans() {
    return this.request('/api/subscription/plans');
  }

  async getCurrentSubscription() {
    return this.request('/api/subscription');
  }

  async createSubscription(plan, paymentMethodId) {
    return this.request('/api/subscription', {
      method: 'POST',
      body: JSON.stringify({ plan, payment_method_id: paymentMethodId }),
    });
  }

  // Payout endpoints
  async requestPayout() {
    return this.request('/api/payouts/request', {
      method: 'POST',
    });
  }

  async getPayoutHistory() {
    return this.request('/api/payouts/history');
  }

  // Additional Deployment endpoints
  async getDeploymentStatus(deploymentId) {
    return this.request(`/api/deployments/${deploymentId}`);
  }

  async getDeployments() {
    return this.request('/api/deployments');
  }

  async createAPI(data) {
    return this.request('/api/apis', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async restartAPI(apiId) {
    return this.request(`/api/apis/${apiId}/restart`, {
      method: 'POST',
    });
  }

  async getAPILogs(apiId, lines = 100) {
    return this.request(`/api/apis/${apiId}/logs?lines=${lines}`);
  }

  async updateAPIConfig(apiId, config) {
    return this.request(`/api/apis/${apiId}/config`, {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  }

  async rollbackDeployment(apiId, deploymentId) {
    return this.request(`/api/deployments/${apiId}/rollback`, {
      method: 'POST',
      body: JSON.stringify({ target_deployment_id: deploymentId }),
    });
  }

  // WebSocket connection for real-time updates
  connectWebSocket() {
    const wsURL = this.baseURL.replace('http', 'ws') + '/ws';
    const token = this.getToken();
    
    if (!token) {
      console.error('No auth token for WebSocket');
      return null;
    }

    const ws = new WebSocket(`${wsURL}?token=${token}`);
    
    ws.onopen = () => {
      console.log('WebSocket connected');
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
    
    ws.onclose = () => {
      console.log('WebSocket disconnected');
      // Reconnect after 5 seconds
      setTimeout(() => this.connectWebSocket(), 5000);
    };
    
    return ws;
  }
}

// Create singleton instance
const apiClient = new APIClient();

// Export for use in dashboard and tests
if (typeof window !== 'undefined') {
  window.apiClient = apiClient;
  
  // Initialize on page load
  document.addEventListener('DOMContentLoaded', () => {
    // Check if user is authenticated
    const token = apiClient.getToken();
    if (!token && !window.location.pathname.includes('login')) {
      window.location.href = '/login.html';
    }
  });
}

// Export for Node.js/Jest
if (typeof module !== 'undefined' && module.exports) {
  module.exports = APIClient;
}