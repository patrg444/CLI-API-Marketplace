// API Client for Console Dashboard
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000';

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

    const response = await fetch(`${this.baseURL}${endpoint}`, config);
    
    if (response.status === 401) {
      this.clearToken();
      window.location.href = '/login';
      throw new Error('Unauthorized');
    }

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.detail || `HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  // Auth endpoints
  async login(email, password) {
    const response = await this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
    
    if (response.access_token) {
      this.setToken(response.access_token);
    }
    
    return response;
  }

  async register(data) {
    const response = await this.request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    
    if (response.access_token) {
      this.setToken(response.access_token);
    }
    
    return response;
  }

  async getMe() {
    return this.request('/auth/me');
  }

  // Dashboard endpoints
  async getDashboardStats() {
    return this.request('/dashboard/stats');
  }

  async getAnalyticsOverview(period = '7d') {
    return this.request(`/analytics/overview?period=${period}`);
  }

  // API management
  async getMyAPIs() {
    return this.request('/apis');
  }

  async createAPI(data) {
    return this.request('/apis', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  // API Keys
  async getAPIKeys() {
    return this.request('/api-keys');
  }

  async createAPIKey(data) {
    return this.request('/api-keys', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  // Billing
  async getBillingSummary() {
    return this.request('/billing/summary');
  }

  // WebSocket connection for real-time updates
  connectWebSocket(onMessage) {
    const wsURL = this.baseURL.replace('http', 'ws') + '/ws';
    const ws = new WebSocket(wsURL);
    
    ws.onopen = () => {
      console.log('WebSocket connected');
    };
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      onMessage(data);
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
    
    ws.onclose = () => {
      console.log('WebSocket disconnected');
      // Reconnect after 5 seconds
      setTimeout(() => this.connectWebSocket(onMessage), 5000);
    };
    
    return ws;
  }
}

// Create singleton instance
const apiClient = new APIClient();

// React hooks
export function useAuth() {
  const [user, setUser] = React.useState(null);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    async function checkAuth() {
      try {
        const token = apiClient.getToken();
        if (token) {
          const userData = await apiClient.getMe();
          setUser(userData);
        }
      } catch (error) {
        console.error('Auth check failed:', error);
      } finally {
        setLoading(false);
      }
    }
    
    checkAuth();
  }, []);

  const login = async (email, password) => {
    const response = await apiClient.login(email, password);
    setUser(response.user);
    return response;
  };

  const logout = () => {
    apiClient.clearToken();
    setUser(null);
    window.location.href = '/login';
  };

  return { user, loading, login, logout };
}

export function useDashboardStats() {
  const [stats, setStats] = React.useState(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(null);

  React.useEffect(() => {
    async function fetchStats() {
      try {
        const data = await apiClient.getDashboardStats();
        setStats(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    
    fetchStats();
    
    // Set up WebSocket for real-time updates
    const ws = apiClient.connectWebSocket((message) => {
      if (message.type === 'stats_update') {
        setStats(prev => ({ ...prev, ...message.data }));
      }
    });
    
    return () => ws.close();
  }, []);

  return { stats, loading, error };
}

export function useAnalytics(period = '7d') {
  const [analytics, setAnalytics] = React.useState(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(null);

  React.useEffect(() => {
    async function fetchAnalytics() {
      try {
        const data = await apiClient.getAnalyticsOverview(period);
        setAnalytics(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    }
    
    fetchAnalytics();
  }, [period]);

  return { analytics, loading, error };
}

export default apiClient;