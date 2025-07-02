// Higher-order function to add rate limiting to API handlers

const rateLimiter = require('./rate-limiter');

function withRateLimit(handler, endpoint = 'default') {
  return async (req, res) => {
    // Get client IP
    const clientIp = req.headers['x-forwarded-for'] || 
                    req.headers['x-real-ip'] || 
                    req.connection?.remoteAddress || 
                    'unknown';
    
    // Check rate limit
    const limitCheck = rateLimiter.check(clientIp, endpoint);
    
    // Always set rate limit headers
    res.setHeader('X-RateLimit-Limit', limitCheck.limit);
    res.setHeader('X-RateLimit-Remaining', limitCheck.remaining);
    res.setHeader('X-RateLimit-Reset', new Date(limitCheck.resetAt).toISOString());
    
    if (!limitCheck.allowed) {
      res.setHeader('Retry-After', limitCheck.retryAfter);
      
      return res.status(429).json({
        success: false,
        error: 'Too many requests. Please try again later.',
        retryAfter: limitCheck.retryAfter
      });
    }
    
    // Call the original handler
    return handler(req, res);
  };
}

module.exports = withRateLimit;