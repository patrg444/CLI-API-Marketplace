const RateLimiter = require('../api/rate-limiter');

describe('Rate Limiter', () => {
  beforeEach(() => {
    // Reset rate limiter state before each test
    RateLimiter.limits.clear();
  });

  describe('Basic functionality', () => {
    test('should allow requests within limit', () => {
      const ip = '127.0.0.1';
      
      for (let i = 0; i < 100; i++) {
        const result = RateLimiter.check(ip);
        expect(result.allowed).toBe(true);
        expect(result.remaining).toBe(99 - i);
      }
    });

    test('should block requests exceeding limit', () => {
      const ip = '127.0.0.1';
      
      // Use up all requests
      for (let i = 0; i < 100; i++) {
        RateLimiter.check(ip);
      }
      
      // Next request should be blocked
      const result = RateLimiter.check(ip);
      expect(result.allowed).toBe(false);
      expect(result.remaining).toBe(0);
      expect(result.retryAfter).toBeGreaterThan(0);
    });

    test('should track different IPs separately', () => {
      const ip1 = '127.0.0.1';
      const ip2 = '127.0.0.2';
      
      // Use up limit for IP1
      for (let i = 0; i < 100; i++) {
        RateLimiter.check(ip1);
      }
      
      // IP2 should still have full quota
      const result = RateLimiter.check(ip2);
      expect(result.allowed).toBe(true);
      expect(result.remaining).toBe(99);
    });
  });

  describe('Endpoint-specific limits', () => {
    test('should apply different limits for different endpoints', () => {
      const ip = '127.0.0.1';
      
      // Default endpoint (100 requests)
      const defaultResult = RateLimiter.check(ip, 'default');
      expect(defaultResult.limit).toBe(100);
      
      // Search endpoint (50 requests)
      const searchResult = RateLimiter.check(ip, 'search');
      expect(searchResult.limit).toBe(50);
      
      // API details endpoint (200 requests)
      const detailsResult = RateLimiter.check(ip, 'apiDetails');
      expect(detailsResult.limit).toBe(200);
    });

    test('should track endpoints separately for same IP', () => {
      const ip = '127.0.0.1';
      
      // Use up search limit
      for (let i = 0; i < 50; i++) {
        RateLimiter.check(ip, 'search');
      }
      
      // Search should be blocked
      const searchResult = RateLimiter.check(ip, 'search');
      expect(searchResult.allowed).toBe(false);
      
      // Default endpoint should still work
      const defaultResult = RateLimiter.check(ip, 'default');
      expect(defaultResult.allowed).toBe(true);
    });
  });

  describe('Time window behavior', () => {
    test('should reset after time window expires', () => {
      const ip = '127.0.0.1';
      const now = Date.now();
      
      // Use up limit
      for (let i = 0; i < 100; i++) {
        RateLimiter.check(ip);
      }
      
      // Get the record and manually expire it
      const key = `${ip}:default`;
      const record = RateLimiter.limits.get(key);
      record.resetAt = now - 1000; // Set to past
      
      // Next request should reset the window
      const result = RateLimiter.check(ip);
      expect(result.allowed).toBe(true);
      expect(result.remaining).toBe(99);
    });

    test('should provide correct resetAt time', () => {
      const ip = '127.0.0.1';
      const before = Date.now();
      
      const result = RateLimiter.check(ip);
      
      const after = Date.now();
      const expectedResetAt = before + RateLimiter.config.windowMs;
      
      expect(result.resetAt).toBeGreaterThanOrEqual(expectedResetAt);
      expect(result.resetAt).toBeLessThanOrEqual(after + RateLimiter.config.windowMs);
    });
  });

  describe('Cleanup functionality', () => {
    test('should clean up expired entries', () => {
      const ip = '127.0.0.1';
      
      // Create an entry
      RateLimiter.check(ip);
      expect(RateLimiter.limits.size).toBe(1);
      
      // Manually expire it
      const key = `${ip}:default`;
      const record = RateLimiter.limits.get(key);
      record.resetAt = Date.now() - 1000;
      
      // Run cleanup
      RateLimiter.cleanup();
      
      expect(RateLimiter.limits.size).toBe(0);
    });

    test('should not clean up active entries', () => {
      const ip = '127.0.0.1';
      
      // Create an entry
      RateLimiter.check(ip);
      
      // Run cleanup
      RateLimiter.cleanup();
      
      // Entry should still exist
      expect(RateLimiter.limits.size).toBe(1);
    });
  });

  describe('Reset functionality', () => {
    test('should reset specific IP and endpoint', () => {
      const ip = '127.0.0.1';
      
      // Use up some requests
      for (let i = 0; i < 50; i++) {
        RateLimiter.check(ip);
      }
      
      // Reset
      RateLimiter.reset(ip, 'default');
      
      // Should have full quota again
      const result = RateLimiter.check(ip);
      expect(result.allowed).toBe(true);
      expect(result.remaining).toBe(99);
    });
  });

  describe('Edge cases', () => {
    test('should handle undefined endpoint gracefully', () => {
      const ip = '127.0.0.1';
      const result = RateLimiter.check(ip, 'unknownEndpoint');
      
      expect(result.limit).toBe(100); // Should use default
      expect(result.allowed).toBe(true);
    });

    test('should handle concurrent requests correctly', () => {
      const ip = '127.0.0.1';
      const promises = [];
      
      // Simulate 150 concurrent requests
      for (let i = 0; i < 150; i++) {
        promises.push(Promise.resolve(RateLimiter.check(ip)));
      }
      
      return Promise.all(promises).then(results => {
        const allowed = results.filter(r => r.allowed).length;
        const blocked = results.filter(r => !r.allowed).length;
        
        expect(allowed).toBe(100);
        expect(blocked).toBe(50);
      });
    });

    test('should provide correct retry-after header value', () => {
      const ip = '127.0.0.1';
      
      // Use up limit
      for (let i = 0; i < 100; i++) {
        RateLimiter.check(ip);
      }
      
      // Check retry after
      const result = RateLimiter.check(ip);
      expect(result.retryAfter).toBeLessThanOrEqual(15 * 60); // Max 15 minutes
      expect(result.retryAfter).toBeGreaterThan(0);
    });
  });

  describe('Performance', () => {
    test('should handle large number of IPs efficiently', () => {
      const start = Date.now();
      
      // Simulate 1000 different IPs
      for (let i = 0; i < 1000; i++) {
        RateLimiter.check(`192.168.1.${i % 256}`);
      }
      
      const duration = Date.now() - start;
      expect(duration).toBeLessThan(100); // Should complete in under 100ms
    });

    test('cleanup should not block requests', () => {
      // Add many expired entries
      for (let i = 0; i < 100; i++) {
        const key = `192.168.1.${i}:default`;
        RateLimiter.limits.set(key, {
          count: 100,
          resetAt: Date.now() - 1000
        });
      }
      
      const start = Date.now();
      
      // Trigger cleanup with 1% chance - force it
      const originalRandom = Math.random;
      Math.random = () => 0.001; // Force cleanup
      
      RateLimiter.check('127.0.0.1');
      
      Math.random = originalRandom;
      
      const duration = Date.now() - start;
      expect(duration).toBeLessThan(10); // Cleanup should be fast
    });
  });
});