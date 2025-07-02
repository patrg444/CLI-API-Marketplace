const handler = require('../api/index').default;

// Mock request and response
function createMockReq(method = 'GET', path = '/api/categories', ip = '127.0.0.1', search = '') {
  const url = `http://localhost:3000/api${path}${search ? `?${search}` : ''}`;
  return {
    method,
    url,
    headers: {
      host: 'localhost:3000',
      'x-forwarded-for': ip
    },
    connection: {
      remoteAddress: ip
    }
  };
}

function createMockRes() {
  const res = {
    statusCode: 200,
    headers: {},
    body: null,
    setHeader: jest.fn((key, value) => {
      res.headers[key] = value;
    }),
    status: jest.fn((code) => {
      res.statusCode = code;
      return res;
    }),
    json: jest.fn((data) => {
      res.body = data;
      return res;
    }),
    end: jest.fn()
  };
  return res;
}

// Reset rate limiter before each test
const RateLimiter = require('../api/rate-limiter');

describe('API Rate Limiting', () => {
  beforeEach(() => {
    RateLimiter.limits.clear();
  });

  describe('Categories endpoint', () => {
    test('should apply rate limiting to categories endpoint', () => {
      const ip = '192.168.1.1';
      
      // Make 100 requests (the limit)
      for (let i = 0; i < 100; i++) {
        const req = createMockReq('GET', '/categories', ip);
        const res = createMockRes();
        
        handler(req, res);
        
        expect(res.statusCode).toBe(200);
        expect(res.headers['X-RateLimit-Limit']).toBe(100);
        expect(res.headers['X-RateLimit-Remaining']).toBe(99 - i);
        expect(res.headers['X-RateLimit-Reset']).toBeDefined();
      }
      
      // 101st request should be rate limited
      const req = createMockReq('GET', '/categories', ip);
      const res = createMockRes();
      
      handler(req, res);
      
      expect(res.statusCode).toBe(429);
      expect(res.body.error).toBe('Too many requests. Please try again later.');
      expect(res.headers['Retry-After']).toBeDefined();
    });
  });

  describe('APIs endpoint', () => {
    test('should apply default rate limiting for general API listing', () => {
      const ip = '192.168.1.2';
      
      // Make 100 requests
      for (let i = 0; i < 100; i++) {
        const req = createMockReq('GET', '/apis', ip);
        const res = createMockRes();
        
        handler(req, res);
        
        expect(res.statusCode).toBe(200);
        expect(res.headers['X-RateLimit-Limit']).toBe(100);
      }
      
      // 101st request should be rate limited
      const req = createMockReq('GET', '/apis', ip);
      const res = createMockRes();
      
      handler(req, res);
      
      expect(res.statusCode).toBe(429);
    });

    test('should apply search-specific rate limiting when search parameter present', () => {
      const ip = '192.168.1.3';
      
      // Make 50 search requests (search endpoint limit)
      for (let i = 0; i < 50; i++) {
        const req = createMockReq('GET', '/apis', ip, 'search=weather');
        const res = createMockRes();
        
        handler(req, res);
        
        expect(res.statusCode).toBe(200);
        expect(res.headers['X-RateLimit-Limit']).toBe(50);
        expect(res.headers['X-RateLimit-Remaining']).toBe(49 - i);
      }
      
      // 51st search request should be rate limited
      const req = createMockReq('GET', '/apis', ip, 'search=weather');
      const res = createMockRes();
      
      handler(req, res);
      
      expect(res.statusCode).toBe(429);
    });

    test('should track search and non-search limits separately', () => {
      const ip = '192.168.1.4';
      
      // Use up search limit
      for (let i = 0; i < 50; i++) {
        const req = createMockReq('GET', '/apis', ip, 'search=test');
        const res = createMockRes();
        handler(req, res);
      }
      
      // Non-search requests should still work
      const req = createMockReq('GET', '/apis', ip);
      const res = createMockRes();
      
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.headers['X-RateLimit-Limit']).toBe(100);
      expect(res.headers['X-RateLimit-Remaining']).toBe(99);
    });
  });

  describe('Featured APIs endpoint', () => {
    test('should apply rate limiting to featured endpoint', () => {
      const ip = '192.168.1.5';
      
      // Make requests up to limit
      for (let i = 0; i < 100; i++) {
        const req = createMockReq('GET', '/apis/featured', ip);
        const res = createMockRes();
        
        handler(req, res);
        
        expect(res.statusCode).toBe(200);
      }
      
      // Next request should be rate limited
      const req = createMockReq('GET', '/apis/featured', ip);
      const res = createMockRes();
      
      handler(req, res);
      
      expect(res.statusCode).toBe(429);
    });
  });

  describe('Trending APIs endpoint', () => {
    test('should apply rate limiting to trending endpoint', () => {
      const ip = '192.168.1.6';
      
      // Make requests up to limit
      for (let i = 0; i < 100; i++) {
        const req = createMockReq('GET', '/apis/trending', ip);
        const res = createMockRes();
        
        handler(req, res);
        
        expect(res.statusCode).toBe(200);
      }
      
      // Next request should be rate limited
      const req = createMockReq('GET', '/apis/trending', ip);
      const res = createMockRes();
      
      handler(req, res);
      
      expect(res.statusCode).toBe(429);
    });
  });

  describe('API Details endpoint', () => {
    test('should apply higher rate limit for API details', () => {
      const ip = '192.168.1.7';
      
      // API details has limit of 200
      for (let i = 0; i < 200; i++) {
        const req = createMockReq('GET', '/apis/sentiment-analyzer-pro', ip);
        const res = createMockRes();
        
        handler(req, res);
        
        expect(res.statusCode).toBe(200);
        expect(res.headers['X-RateLimit-Limit']).toBe(200);
        expect(res.headers['X-RateLimit-Remaining']).toBe(199 - i);
      }
      
      // 201st request should be rate limited
      const req = createMockReq('GET', '/apis/sentiment-analyzer-pro', ip);
      const res = createMockRes();
      
      handler(req, res);
      
      expect(res.statusCode).toBe(429);
    });

    test('should apply rate limiting and validate API ID', () => {
      const ip = '192.168.1.8';
      
      // Invalid API ID format should be rejected with 400
      const req = createMockReq('GET', '/apis/invalid..api', ip);
      const res = createMockRes();
      
      handler(req, res);
      
      expect(res.statusCode).toBe(400);
      expect(res.body.error).toBe('Invalid API ID');
      
      // Rate limit headers should still be set (rate limiting was applied)
      expect(res.headers['X-RateLimit-Limit']).toBe(200);
      expect(res.headers['X-RateLimit-Remaining']).toBe(199);
    });
  });

  describe('Rate limit headers', () => {
    test('should always include rate limit headers on successful requests', () => {
      const ip = '192.168.1.9';
      const req = createMockReq('GET', '/categories', ip);
      const res = createMockRes();
      
      handler(req, res);
      
      expect(res.headers['X-RateLimit-Limit']).toBeDefined();
      expect(res.headers['X-RateLimit-Remaining']).toBeDefined();
      expect(res.headers['X-RateLimit-Reset']).toBeDefined();
      expect(res.headers['Retry-After']).toBeUndefined(); // Only on 429
    });

    test('should include Retry-After header on rate limited responses', () => {
      const ip = '192.168.1.10';
      
      // Exhaust limit
      for (let i = 0; i < 100; i++) {
        handler(createMockReq('GET', '/categories', ip), createMockRes());
      }
      
      const req = createMockReq('GET', '/categories', ip);
      const res = createMockRes();
      
      handler(req, res);
      
      expect(res.statusCode).toBe(429);
      expect(res.headers['Retry-After']).toBeDefined();
      expect(typeof res.headers['Retry-After']).toBe('number');
      expect(res.headers['Retry-After']).toBeGreaterThan(0);
      expect(res.headers['Retry-After']).toBeLessThanOrEqual(900); // Max 15 minutes
    });
  });

  describe('IP detection', () => {
    test('should handle different IP header formats', () => {
      // Test x-forwarded-for
      const req1 = createMockReq('GET', '/categories');
      req1.headers['x-forwarded-for'] = '10.0.0.1';
      const res1 = createMockRes();
      
      handler(req1, res1);
      expect(res1.statusCode).toBe(200);
      
      // Test x-real-ip
      const req2 = createMockReq('GET', '/categories');
      delete req2.headers['x-forwarded-for'];
      req2.headers['x-real-ip'] = '10.0.0.2';
      const res2 = createMockRes();
      
      handler(req2, res2);
      expect(res2.statusCode).toBe(200);
      
      // Test connection.remoteAddress
      const req3 = createMockReq('GET', '/categories');
      delete req3.headers['x-forwarded-for'];
      delete req3.headers['x-real-ip'];
      req3.connection = { remoteAddress: '10.0.0.3' };
      const res3 = createMockRes();
      
      handler(req3, res3);
      expect(res3.statusCode).toBe(200);
    });

    test('should handle unknown IP gracefully', () => {
      const req = createMockReq('GET', '/categories');
      delete req.headers['x-forwarded-for'];
      delete req.headers['x-real-ip'];
      delete req.connection;
      
      const res = createMockRes();
      
      handler(req, res);
      
      expect(res.statusCode).toBe(200);
      expect(res.headers['X-RateLimit-Limit']).toBeDefined();
    });
  });

  describe('Cross-endpoint isolation', () => {
    test('should not share limits between different rate limit types', () => {
      const ip = '192.168.1.11';
      
      // Use up search limit (50 requests)
      for (let i = 0; i < 50; i++) {
        handler(createMockReq('GET', '/apis', ip, 'search=test'), createMockRes());
      }
      
      // Search endpoint should be blocked now
      const searchReq = createMockReq('GET', '/apis', ip, 'search=test');
      const searchRes = createMockRes();
      handler(searchReq, searchRes);
      expect(searchRes.statusCode).toBe(429);
      
      // But other endpoint types should still work
      const endpoints = [
        { path: '/apis', expectedStatus: 200 }, // Non-search API listing (different limit)
        { path: '/apis/sentiment-analyzer-pro', expectedStatus: 200 } // API details (different limit)
      ];
      
      for (const { path, expectedStatus } of endpoints) {
        const req = createMockReq('GET', path, ip);
        const res = createMockRes();
        
        handler(req, res);
        
        expect(res.statusCode).toBe(expectedStatus);
        expect(res.statusCode).not.toBe(429);
      }
    });
  });
});