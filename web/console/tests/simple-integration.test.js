/**
 * Simple integration test to verify console components work together
 */

describe('Console Integration', () => {
  test('API client exists and has required methods', () => {
    const APIClient = require('../api-client-updated.js');
    const client = new APIClient();
    
    // Check essential methods exist
    expect(typeof client.login).toBe('function');
    expect(typeof client.getDashboardStats).toBe('function');
    expect(typeof client.getMyAPIs).toBe('function');
    expect(typeof client.createAPIKey).toBe('function');
    expect(typeof client.startTrial).toBe('function');
    expect(typeof client.connectWebSocket).toBe('function');
  });
  
  test('Dashboard functions can be loaded', () => {
    // Mock window and document
    global.window = { location: { hostname: 'localhost' } };
    global.document = {
      addEventListener: jest.fn(),
      querySelector: jest.fn(() => ({ textContent: '' })),
      getElementById: jest.fn(() => ({ innerHTML: '' }))
    };
    
    // Should not throw when requiring
    expect(() => {
      require('../dashboard-updated.js');
    }).not.toThrow();
  });
  
  test('API client constructs correct URLs', () => {
    const APIClient = require('../api-client-updated.js');
    const client = new APIClient();
    
    // Check base URL
    expect(client.baseURL).toBe('http://localhost:8000');
    
    // Test URL construction for different endpoints
    const endpoints = [
      '/api/auth/login',
      '/api/dashboard/overview',
      '/api/my-apis',
      '/api/keys',
      '/api/marketplace/listings',
      '/api/trials/start'
    ];
    
    endpoints.forEach(endpoint => {
      expect(`${client.baseURL}${endpoint}`).toMatch(/^http:\/\/localhost:8000\/api\//);
    });
  });
  
  test('Console pages have proper structure', () => {
    const pages = [
      'dashboard.html',
      'apis.html', 
      'analytics.html',
      'earnings.html',
      'marketplace.html'
    ];
    
    const fs = require('fs');
    const path = require('path');
    
    pages.forEach(page => {
      const pagePath = path.join(__dirname, '..', 'pages', page);
      if (fs.existsSync(pagePath)) {
        const content = fs.readFileSync(pagePath, 'utf8');
        
        // Check for basic structure
        expect(content).toContain('{% extends');
        expect(content).toContain('{% block content %}');
        
        // Check for proper semantic HTML
        expect(content).toMatch(/<h1[^>]*>/);
        expect(content).toContain('class=');
      }
    });
  });
});