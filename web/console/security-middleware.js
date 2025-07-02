// Security middleware for console application
// Implements OWASP security best practices

const crypto = require('crypto');

class SecurityMiddleware {
  constructor() {
    this.rateLimits = new Map();
    this.sessionStore = new Map();
  }
  
  // Content Security Policy
  setSecurityHeaders(req, res, next) {
    // Prevent XSS attacks
    res.setHeader('X-Content-Type-Options', 'nosniff');
    res.setHeader('X-Frame-Options', 'DENY');
    res.setHeader('X-XSS-Protection', '1; mode=block');
    
    // Content Security Policy
    res.setHeader('Content-Security-Policy', 
      "default-src 'self'; " +
      "script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; " +
      "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
      "font-src 'self' https://fonts.gstatic.com; " +
      "img-src 'self' data: https:; " +
      "connect-src 'self' https://api.apidirect.dev"
    );
    
    // Referrer Policy
    res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');
    
    // Feature Policy
    res.setHeader('Feature-Policy', 
      "geolocation 'none'; microphone 'none'; camera 'none'"
    );
    
    if (next) next();
  }
  
  // CSRF Protection
  generateCSRFToken(sessionId) {
    const token = crypto.randomBytes(32).toString('hex');
    const session = this.sessionStore.get(sessionId) || {};
    session.csrfToken = token;
    this.sessionStore.set(sessionId, session);
    return token;
  }
  
  validateCSRFToken(req, res, next) {
    // Skip CSRF for GET requests
    if (req.method === 'GET' || req.method === 'HEAD' || req.method === 'OPTIONS') {
      return next ? next() : true;
    }
    
    const sessionId = req.cookies?.sessionId || req.headers['x-session-id'];
    const session = this.sessionStore.get(sessionId);
    
    if (!session) {
      res.status(403).json({ error: 'Invalid session' });
      return false;
    }
    
    const token = req.body?.csrf_token || 
                  req.headers['x-csrf-token'] || 
                  req.query?.csrf_token;
    
    if (!token || token !== session.csrfToken) {
      res.status(403).json({ error: 'Invalid CSRF token' });
      return false;
    }
    
    if (next) next();
    return true;
  }
  
  // Rate Limiting
  rateLimit(options = {}) {
    const {
      windowMs = 15 * 60 * 1000, // 15 minutes
      max = 100, // limit each IP to 100 requests per windowMs
      message = 'Too many requests from this IP'
    } = options;
    
    return (req, res, next) => {
      const ip = req.ip || req.connection.remoteAddress;
      const key = `${ip}:${req.path}`;
      const now = Date.now();
      
      // Clean up old entries
      for (const [k, v] of this.rateLimits.entries()) {
        if (v.resetTime < now) {
          this.rateLimits.delete(k);
        }
      }
      
      let limit = this.rateLimits.get(key);
      
      if (!limit) {
        limit = {
          count: 0,
          resetTime: now + windowMs
        };
        this.rateLimits.set(key, limit);
      }
      
      if (limit.resetTime < now) {
        limit.count = 0;
        limit.resetTime = now + windowMs;
      }
      
      limit.count++;
      
      // Set rate limit headers
      res.setHeader('X-RateLimit-Limit', max);
      res.setHeader('X-RateLimit-Remaining', Math.max(0, max - limit.count));
      res.setHeader('X-RateLimit-Reset', new Date(limit.resetTime).toISOString());
      
      if (limit.count > max) {
        res.status(429).json({ error: message });
        return;
      }
      
      if (next) next();
    };
  }
  
  // Input Sanitization
  sanitizeInput(input) {
    if (typeof input !== 'string') return input;
    
    // Remove null bytes
    input = input.replace(/\0/g, '');
    
    // Encode HTML entities
    const htmlEntities = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#x27;',
      '/': '&#x2F;'
    };
    
    return input.replace(/[&<>"'/]/g, match => htmlEntities[match]);
  }
  
  // SQL Injection Prevention (parameter validation)
  validateSQLParameter(param, type = 'string') {
    switch (type) {
      case 'number':
        const num = Number(param);
        if (isNaN(num)) throw new Error('Invalid number parameter');
        return num;
        
      case 'boolean':
        return param === 'true' || param === true;
        
      case 'string':
      default:
        // Remove SQL meta-characters
        if (typeof param !== 'string') return '';
        
        // Check for SQL injection patterns
        const sqlInjectionPatterns = [
          /(\b(union|select|insert|update|delete|drop|create|alter|exec|execute)\b)/gi,
          /(--|#|\/\*|\*\/)/g,
          /(\bor\b|\band\b)\s*\d+\s*=\s*\d+/gi,
          /[';]/g
        ];
        
        for (const pattern of sqlInjectionPatterns) {
          if (pattern.test(param)) {
            throw new Error('Invalid input detected');
          }
        }
        
        return param;
    }
  }
  
  // Password validation
  validatePassword(password) {
    const errors = [];
    
    if (password.length < 8) {
      errors.push('Password must be at least 8 characters long');
    }
    
    if (!/[A-Z]/.test(password)) {
      errors.push('Password must contain at least one uppercase letter');
    }
    
    if (!/[a-z]/.test(password)) {
      errors.push('Password must contain at least one lowercase letter');
    }
    
    if (!/[0-9]/.test(password)) {
      errors.push('Password must contain at least one number');
    }
    
    if (!/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
      errors.push('Password must contain at least one special character');
    }
    
    // Check for common passwords
    const commonPasswords = [
      'password', '12345678', 'qwerty', 'abc123', 'password123',
      'admin', 'letmein', 'welcome', 'monkey', '1234567890'
    ];
    
    if (commonPasswords.includes(password.toLowerCase())) {
      errors.push('Password is too common');
    }
    
    return {
      valid: errors.length === 0,
      errors
    };
  }
  
  // Session validation
  validateSession(req, res, next) {
    const sessionId = req.cookies?.sessionId || req.headers['x-session-id'];
    
    if (!sessionId) {
      res.status(401).json({ error: 'No session found' });
      return;
    }
    
    const session = this.sessionStore.get(sessionId);
    
    if (!session) {
      res.status(401).json({ error: 'Invalid session' });
      return;
    }
    
    // Check session expiry
    if (session.expiresAt && new Date(session.expiresAt) < new Date()) {
      this.sessionStore.delete(sessionId);
      res.status(401).json({ error: 'Session expired' });
      return;
    }
    
    // Refresh session
    session.lastActivity = new Date();
    this.sessionStore.set(sessionId, session);
    
    req.session = session;
    if (next) next();
  }
  
  // Create secure session
  createSession(userId, data = {}) {
    const sessionId = crypto.randomBytes(32).toString('hex');
    const session = {
      id: sessionId,
      userId,
      createdAt: new Date(),
      lastActivity: new Date(),
      expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000), // 24 hours
      csrfToken: crypto.randomBytes(32).toString('hex'),
      ...data
    };
    
    this.sessionStore.set(sessionId, session);
    return session;
  }
  
  // Email validation
  validateEmail(email) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    
    if (!emailRegex.test(email)) {
      return { valid: false, error: 'Invalid email format' };
    }
    
    // Additional checks
    if (email.length > 254) {
      return { valid: false, error: 'Email too long' };
    }
    
    const [localPart, domain] = email.split('@');
    
    if (localPart.length > 64) {
      return { valid: false, error: 'Local part too long' };
    }
    
    // Check for consecutive dots
    if (email.includes('..')) {
      return { valid: false, error: 'Invalid email format' };
    }
    
    return { valid: true };
  }
  
  // API key validation
  validateAPIKey(apiKey) {
    // Check format
    if (!apiKey || typeof apiKey !== 'string') {
      return { valid: false, error: 'Invalid API key' };
    }
    
    // Expected format: api_[32 hex characters]
    const apiKeyRegex = /^api_[a-f0-9]{32}$/;
    
    if (!apiKeyRegex.test(apiKey)) {
      return { valid: false, error: 'Invalid API key format' };
    }
    
    return { valid: true };
  }
}

// Export singleton instance
module.exports = new SecurityMiddleware();