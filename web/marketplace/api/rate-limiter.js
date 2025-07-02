// Simple in-memory rate limiter for API endpoints
// In production, use Redis or similar for distributed rate limiting

class RateLimiter {
  constructor() {
    this.limits = new Map();
    this.config = {
      windowMs: 15 * 60 * 1000, // 15 minutes
      maxRequests: {
        default: 100,
        search: 50,
        apiDetails: 200
      }
    };
  }
  
  check(ip, endpoint = 'default') {
    const now = Date.now();
    const key = `${ip}:${endpoint}`;
    const maxRequests = this.config.maxRequests[endpoint] || this.config.maxRequests.default;
    
    // Clean up expired entries periodically
    if (Math.random() < 0.01) { // 1% chance to clean up
      this.cleanup();
    }
    
    let record = this.limits.get(key);
    
    if (!record) {
      record = {
        count: 0,
        resetAt: now + this.config.windowMs
      };
      this.limits.set(key, record);
    }
    
    // Reset if window expired
    if (now > record.resetAt) {
      record.count = 0;
      record.resetAt = now + this.config.windowMs;
    }
    
    record.count++;
    
    return {
      allowed: record.count <= maxRequests,
      limit: maxRequests,
      remaining: Math.max(0, maxRequests - record.count),
      resetAt: record.resetAt,
      retryAfter: record.count > maxRequests ? Math.ceil((record.resetAt - now) / 1000) : null
    };
  }
  
  cleanup() {
    const now = Date.now();
    for (const [key, record] of this.limits.entries()) {
      if (now > record.resetAt) {
        this.limits.delete(key);
      }
    }
  }
  
  reset(ip, endpoint) {
    const key = `${ip}:${endpoint}`;
    this.limits.delete(key);
  }
}

module.exports = new RateLimiter();