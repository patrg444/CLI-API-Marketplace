/**
 * Unit tests for API Client
 */

// Mock localStorage for testing
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
global.localStorage = localStorageMock;

// Mock fetch
global.fetch = jest.fn();

// Import API client (we'll use require to ensure mocks are in place)
let apiClient;

describe('API Client Tests', () => {
  beforeEach(() => {
    // Clear all mocks
    jest.clearAllMocks();
    fetch.mockClear();
    localStorageMock.getItem.mockClear();
    localStorageMock.setItem.mockClear();
    localStorageMock.removeItem.mockClear();
    
    // Reset modules to get fresh instance
    jest.resetModules();
    
    // Mock window location
    delete window.location;
    window.location = { 
      href: '',
      hostname: 'localhost'
    };
  });

  describe('Authentication', () => {
    test('should login successfully and store token', async () => {
      const mockToken = 'test-jwt-token';
      const mockUser = { id: '123', email: 'test@example.com' };
      
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          access_token: mockToken,
          user: mockUser
        })
      });

      // Create new instance
      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();

      const response = await apiClient.login('test@example.com', 'password123');

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8000/api/auth/login',
        expect.objectContaining({
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            email: 'test@example.com',
            password: 'password123'
          })
        })
      );

      expect(localStorageMock.setItem).toHaveBeenCalledWith('api_token', mockToken);
      expect(response.access_token).toBe(mockToken);
      expect(response.user).toEqual(mockUser);
    });

    test('should handle login failure', async () => {
      fetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({
          detail: 'Invalid credentials'
        })
      });

      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();

      await expect(apiClient.login('test@example.com', 'wrong'))
        .rejects.toThrow('Invalid credentials');
    });

    test('should clear token on logout', () => {
      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();
      apiClient.setToken('test-token');

      apiClient.clearToken();

      expect(localStorageMock.removeItem).toHaveBeenCalledWith('api_token');
      expect(apiClient.token).toBeNull();
    });
  });

  describe('API Requests', () => {
    beforeEach(() => {
      localStorageMock.getItem.mockReturnValue('test-token');
    });

    test('should get dashboard stats with auth header', async () => {
      const mockStats = {
        metrics: {
          total_revenue_30d: 1500.50,
          total_calls_30d: 10000,
          active_deployments: 5
        }
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockStats
      });

      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();

      const stats = await apiClient.getDashboardStats();

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8000/api/dashboard/overview',
        expect.objectContaining({
          headers: expect.objectContaining({
            'Authorization': 'Bearer test-token',
            'Content-Type': 'application/json'
          })
        })
      );

      expect(stats).toEqual(mockStats);
    });

    test('should handle 401 unauthorized', async () => {
      fetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({})
      });

      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();

      await expect(apiClient.getMyAPIs()).rejects.toThrow('Unauthorized');
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('api_token');
      expect(window.location.href).toBe('/login.html');
    });

    test('should get APIs with pagination', async () => {
      const mockAPIs = [
        { id: '1', name: 'API 1', status: 'running' },
        { id: '2', name: 'API 2', status: 'running' }
      ];

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockAPIs
      });

      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();

      const apis = await apiClient.getMyAPIs();

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8000/api/my-apis',
        expect.any(Object)
      );
      expect(apis).toEqual(mockAPIs);
    });
  });

  describe('Marketplace Integration', () => {
    test('should search marketplace with filters', async () => {
      const params = {
        search: 'weather',
        category: 'data',
        maxPrice: 50,
        hasFreeTier: true
      };

      const mockResults = {
        listings: [],
        total: 0
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResults
      });

      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();

      const results = await apiClient.getMarketplaceListings(params);

      const expectedUrl = 'http://localhost:8000/api/marketplace/listings?search=weather&category=data&maxPrice=50&hasFreeTier=true';
      expect(fetch).toHaveBeenCalledWith(expectedUrl, expect.any(Object));
      expect(results).toEqual(mockResults);
    });

    test('should start API trial', async () => {
      const apiId = 'test-api-123';
      const mockTrial = {
        trial_id: 'trial-456',
        requests_limit: 1000,
        expires_at: '2024-01-01T00:00:00Z'
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockTrial
      });

      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();

      const trial = await apiClient.startTrial(apiId);

      expect(fetch).toHaveBeenCalledWith(
        `http://localhost:8000/api/trials/start?api_id=${apiId}`,
        expect.objectContaining({
          method: 'POST'
        })
      );
      expect(trial).toEqual(mockTrial);
    });
  });

  describe('WebSocket Connection', () => {
    let mockWebSocket;

    beforeEach(() => {
      mockWebSocket = {
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        close: jest.fn(),
        send: jest.fn(),
      };
      
      global.WebSocket = jest.fn(() => mockWebSocket);
    });

    test('should create WebSocket connection with token', () => {
      localStorageMock.getItem.mockReturnValue('test-token');
      
      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();

      const ws = apiClient.connectWebSocket();

      expect(WebSocket).toHaveBeenCalledWith('ws://localhost:8000/ws?token=test-token');
      expect(ws).toBe(mockWebSocket);
    });

    test('should not create WebSocket without token', () => {
      localStorageMock.getItem.mockReturnValue(null);
      
      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();

      const ws = apiClient.connectWebSocket();

      expect(WebSocket).not.toHaveBeenCalled();
      expect(ws).toBeNull();
    });
  });

  describe('Error Handling', () => {
    test('should handle network errors', async () => {
      fetch.mockRejectedValueOnce(new Error('Network error'));

      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();

      await expect(apiClient.getDashboardStats())
        .rejects.toThrow('Network error');
    });

    test('should handle invalid JSON response', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => {
          throw new Error('Invalid JSON');
        }
      });

      const APIClient = require('../api-client-updated.js');
      apiClient = new APIClient();

      await expect(apiClient.getDashboardStats())
        .rejects.toThrow('Invalid JSON');
    });
  });
});

// Export for running with Jest
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {};
}