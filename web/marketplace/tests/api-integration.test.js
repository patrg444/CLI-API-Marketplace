// Comprehensive Marketplace API Integration Tests
const handler = require('../api/index.js').default;

// Enhanced mock request/response utilities
const createMockRequest = (url, method = 'GET', body = null) => ({
  url,
  method,
  headers: { host: 'localhost:3000' },
  body
});

const createMockResponse = () => {
  const res = {
    statusCode: 200,
    headers: {},
    data: null,
    status: function(code) {
      this.statusCode = code;
      return this;
    },
    setHeader: function(key, value) {
      this.headers[key] = value;
      return this;
    },
    setHeaders: function(headers) {
      Object.assign(this.headers, headers);
      return this;
    },
    json: function(data) {
      this.data = data;
      return this;
    },
    end: function() {
      return this;
    }
  };
  return res;
};

describe('Marketplace API Integration Tests', () => {
  // CORS Tests
  describe('CORS handling', () => {
    test('should handle preflight requests', () => {
      const req = createMockRequest('/api/categories', 'OPTIONS');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.headers['Access-Control-Allow-Origin']).toBe('*');
      expect(res.headers['Access-Control-Allow-Methods']).toBeDefined();
    });
    
    test('should include CORS headers on all responses', () => {
      const req = createMockRequest('/api/categories');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.headers['Access-Control-Allow-Origin']).toBe('*');
      expect(res.headers['Content-Type']).toBe('application/json');
    });
  });

  // Categories Endpoint Tests
  describe('GET /api/categories', () => {
    test('should return all categories', () => {
      const req = createMockRequest('/api/categories');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.data.success).toBe(true);
      expect(Array.isArray(res.data.data)).toBe(true);
      expect(res.data.data.length).toBeGreaterThan(0);
      expect(res.data.data[0]).toHaveProperty('id');
      expect(res.data.data[0]).toHaveProperty('name');
    });
  });

  // APIs Endpoint Tests
  describe('GET /api/apis', () => {
    test('should return paginated results', () => {
      const req = createMockRequest('/api/apis');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.data.success).toBe(true);
      expect(Array.isArray(res.data.data)).toBe(true);
      expect(res.data.meta).toBeDefined();
      expect(res.data.meta.total).toBeGreaterThanOrEqual(res.data.data.length);
      expect(res.data.meta.page).toBe(1);
    });

    test('should filter by category', () => {
      const req = createMockRequest('/api/apis?category=ai-ml');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      res.data.data.forEach(api => {
        expect(api.category).toBe('ai-ml');
      });
    });

    test('should handle search queries', () => {
      const req = createMockRequest('/api/apis?search=weather');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.data.data.length).toBeGreaterThan(0);
      
      const hasMatch = res.data.data.some(api => 
        api.name.toLowerCase().includes('weather') ||
        api.description.toLowerCase().includes('weather') ||
        api.tags.some(tag => tag.toLowerCase().includes('weather'))
      );
      expect(hasMatch).toBe(true);
    });

    test('should filter by price', () => {
      const req = createMockRequest('/api/apis?maxPrice=0.005');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      res.data.data.forEach(api => {
        if (api.pricing.type === 'freemium') {
          expect(api.pricing.pricePerCall).toBeLessThanOrEqual(0.005);
        }
      });
    });

    test('should sort by rating', () => {
      const req = createMockRequest('/api/apis?sort=rating');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      for (let i = 1; i < res.data.data.length; i++) {
        expect(res.data.data[i-1].rating).toBeGreaterThanOrEqual(res.data.data[i].rating);
      }
    });

    test('should handle pagination parameters', () => {
      const req = createMockRequest('/api/apis?page=2&limit=2');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.data.data.length).toBeLessThanOrEqual(2);
      expect(res.data.meta.page).toBe(2);
      expect(res.data.meta.limit).toBe(2);
    });

    test('should handle invalid pagination parameters', () => {
      const req = createMockRequest('/api/apis?page=invalid&limit=999');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.data.meta.page).toBe(1); // Should default to 1
      expect(res.data.meta.limit).toBe(10); // Should default to 10
    });

    test('should prevent overly long search queries', () => {
      const longQuery = 'a'.repeat(101);
      const req = createMockRequest(`/api/apis?search=${longQuery}`);
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(400);
      expect(res.data.error).toContain('Search query too long');
    });
  });

  // Featured APIs Tests
  describe('GET /api/apis/featured', () => {
    test('should return only featured APIs', () => {
      const req = createMockRequest('/api/apis/featured');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.data.success).toBe(true);
      res.data.data.forEach(api => {
        expect(api.featured).toBe(true);
      });
    });
  });

  // Trending APIs Tests
  describe('GET /api/apis/trending', () => {
    test('should return trending APIs sorted by growth', () => {
      const req = createMockRequest('/api/apis/trending');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.data.success).toBe(true);
      res.data.data.forEach(api => {
        expect(api.trending).toBe(true);
      });
      
      // Check sorting by growth
      for (let i = 1; i < res.data.data.length; i++) {
        const prevGrowth = res.data.data[i-1].growth || 0;
        const currGrowth = res.data.data[i].growth || 0;
        expect(prevGrowth).toBeGreaterThanOrEqual(currGrowth);
      }
    });
  });

  // Specific API Tests
  describe('GET /api/apis/:id', () => {
    test('should return specific API details', () => {
      const req = createMockRequest('/api/apis/sentiment-analyzer-pro');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.data.success).toBe(true);
      expect(res.data.data.id).toBe('sentiment-analyzer-pro');
    });

    test('should return 404 for non-existent API', () => {
      const req = createMockRequest('/api/apis/non-existent-api');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(404);
      expect(res.data.success).toBe(false);
      expect(res.data.error).toBe('API not found');
    });

    test('should validate API ID format', () => {
      const req = createMockRequest('/api/apis/invalid@id!');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(400);
      expect(res.data.error).toBe('Invalid API ID format');
    });

    test('should prevent path traversal attacks', () => {
      const req = createMockRequest('/api/apis/../../../etc/passwd');
      const res = createMockResponse();
      handler(req, res);
      
      // The URL parsing normalizes the path, so it becomes /etc/passwd
      // which doesn't start with /apis/, resulting in 404
      expect(res.statusCode).toBe(404);
      expect(res.data.error).toBe('Endpoint not found');
      
      // Test a more direct path traversal attempt
      const req2 = createMockRequest('/api/apis/../../etc/passwd');
      const res2 = createMockResponse();
      handler(req2, res2);
      
      expect(res2.statusCode).toBe(404);
    });
  });

  // Error Handling Tests
  describe('Error handling', () => {
    test('should handle invalid endpoints', () => {
      const req = createMockRequest('/api/invalid-endpoint');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(404);
      expect(res.data.success).toBe(false);
      expect(res.data.error).toBe('Endpoint not found');
    });

    test('should reject non-GET methods', () => {
      const req = createMockRequest('/api/categories', 'POST');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(405);
      expect(res.data.error).toBe('Method not allowed');
    });
  });

  // Edge Cases
  describe('Edge cases', () => {
    test('should handle empty search results gracefully', () => {
      const req = createMockRequest('/api/apis?search=xyznonexistentxyz');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.data.success).toBe(true);
      expect(res.data.data).toEqual([]);
    });

    test('should handle multiple filters combined', () => {
      const req = createMockRequest('/api/apis?category=ai-ml&search=sentiment&sort=rating');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.data.success).toBe(true);
      
      // Verify all filters are applied
      res.data.data.forEach(api => {
        expect(api.category).toBe('ai-ml');
      });
      
      // Check at least one result matches search
      if (res.data.data.length > 0) {
        const hasMatch = res.data.data.some(api => 
          api.name.toLowerCase().includes('sentiment') ||
          api.description.toLowerCase().includes('sentiment')
        );
        expect(hasMatch).toBe(true);
      }
    });

    test('should handle URL with trailing slashes', () => {
      const req = createMockRequest('/api/categories/');
      const res = createMockResponse();
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.data.success).toBe(true);
    });
  });
});