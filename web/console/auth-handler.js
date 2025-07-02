// Secure authentication handler for console
// Implements best practices for authentication and session management

const crypto = require('crypto');
const bcrypt = require('bcryptjs');
const jwt = require('jsonwebtoken');
const securityMiddleware = require('./security-middleware');

class AuthHandler {
  constructor() {
    // In production, use environment variables
    this.jwtSecret = process.env.JWT_SECRET || crypto.randomBytes(64).toString('hex');
    this.sessionTimeout = 24 * 60 * 60 * 1000; // 24 hours
    this.refreshTokenExpiry = 30 * 24 * 60 * 60 * 1000; // 30 days
    
    // Temporary in-memory store (use Redis in production)
    this.users = new Map();
    this.refreshTokens = new Map();
    this.loginAttempts = new Map();
  }
  
  // Register new user
  async register(email, password, username) {
    try {
      // Validate email
      const emailValidation = securityMiddleware.validateEmail(email);
      if (!emailValidation.valid) {
        return { success: false, error: emailValidation.error };
      }
      
      // Validate password
      const passwordValidation = securityMiddleware.validatePassword(password);
      if (!passwordValidation.valid) {
        return { success: false, errors: passwordValidation.errors };
      }
      
      // Check if user exists
      if (this.users.has(email)) {
        return { success: false, error: 'User already exists' };
      }
      
      // Hash password
      const salt = await bcrypt.genSalt(12);
      const hashedPassword = await bcrypt.hash(password, salt);
      
      // Create user
      const userId = crypto.randomBytes(16).toString('hex');
      const user = {
        id: userId,
        email: email.toLowerCase(),
        username: securityMiddleware.sanitizeInput(username),
        password: hashedPassword,
        createdAt: new Date(),
        emailVerified: false,
        twoFactorEnabled: false,
        role: 'consumer'
      };
      
      this.users.set(email.toLowerCase(), user);
      
      // Generate verification token
      const verificationToken = this.generateVerificationToken(userId);
      
      return {
        success: true,
        userId,
        verificationToken,
        message: 'Registration successful. Please verify your email.'
      };
      
    } catch (error) {
      console.error('Registration error:', error);
      return { success: false, error: 'Registration failed' };
    }
  }
  
  // Login user
  async login(email, password, req) {
    try {
      const clientIp = req.headers['x-forwarded-for'] || req.connection.remoteAddress;
      
      // Check login attempts
      if (!this.checkLoginAttempts(clientIp)) {
        return { 
          success: false, 
          error: 'Too many login attempts. Please try again later.',
          retryAfter: 300 // 5 minutes
        };
      }
      
      // Validate email
      const emailValidation = securityMiddleware.validateEmail(email);
      if (!emailValidation.valid) {
        this.recordFailedAttempt(clientIp);
        return { success: false, error: 'Invalid credentials' };
      }
      
      // Get user
      const user = this.users.get(email.toLowerCase());
      if (!user) {
        this.recordFailedAttempt(clientIp);
        return { success: false, error: 'Invalid credentials' };
      }
      
      // Verify password
      const isValid = await bcrypt.compare(password, user.password);
      if (!isValid) {
        this.recordFailedAttempt(clientIp);
        return { success: false, error: 'Invalid credentials' };
      }
      
      // Check if email is verified
      if (!user.emailVerified) {
        return { success: false, error: 'Please verify your email before logging in' };
      }
      
      // Generate tokens
      const accessToken = this.generateAccessToken(user);
      const refreshToken = this.generateRefreshToken(user);
      
      // Create session
      const session = securityMiddleware.createSession(user.id, {
        email: user.email,
        role: user.role
      });
      
      // Clear failed attempts
      this.clearFailedAttempts(clientIp);
      
      return {
        success: true,
        accessToken,
        refreshToken,
        sessionId: session.id,
        csrfToken: session.csrfToken,
        user: {
          id: user.id,
          email: user.email,
          username: user.username,
          role: user.role
        }
      };
      
    } catch (error) {
      console.error('Login error:', error);
      return { success: false, error: 'Login failed' };
    }
  }
  
  // Verify JWT token
  verifyAccessToken(token) {
    try {
      const decoded = jwt.verify(token, this.jwtSecret);
      return { valid: true, decoded };
    } catch (error) {
      return { valid: false, error: error.message };
    }
  }
  
  // Refresh access token
  async refreshAccessToken(refreshToken) {
    try {
      // Verify refresh token
      const decoded = jwt.verify(refreshToken, this.jwtSecret);
      
      // Check if refresh token exists and is valid
      const storedToken = this.refreshTokens.get(decoded.jti);
      if (!storedToken || storedToken.userId !== decoded.userId) {
        return { success: false, error: 'Invalid refresh token' };
      }
      
      // Get user
      const user = Array.from(this.users.values()).find(u => u.id === decoded.userId);
      if (!user) {
        return { success: false, error: 'User not found' };
      }
      
      // Generate new access token
      const accessToken = this.generateAccessToken(user);
      
      return {
        success: true,
        accessToken
      };
      
    } catch (error) {
      return { success: false, error: 'Invalid refresh token' };
    }
  }
  
  // Logout user
  logout(sessionId, refreshToken) {
    // Remove session
    securityMiddleware.sessionStore.delete(sessionId);
    
    // Revoke refresh token
    if (refreshToken) {
      try {
        const decoded = jwt.verify(refreshToken, this.jwtSecret);
        this.refreshTokens.delete(decoded.jti);
      } catch (error) {
        // Ignore invalid tokens
      }
    }
    
    return { success: true };
  }
  
  // Change password
  async changePassword(userId, currentPassword, newPassword) {
    try {
      // Get user
      const user = Array.from(this.users.values()).find(u => u.id === userId);
      if (!user) {
        return { success: false, error: 'User not found' };
      }
      
      // Verify current password
      const isValid = await bcrypt.compare(currentPassword, user.password);
      if (!isValid) {
        return { success: false, error: 'Current password is incorrect' };
      }
      
      // Validate new password
      const passwordValidation = securityMiddleware.validatePassword(newPassword);
      if (!passwordValidation.valid) {
        return { success: false, errors: passwordValidation.errors };
      }
      
      // Hash new password
      const salt = await bcrypt.genSalt(12);
      user.password = await bcrypt.hash(newPassword, salt);
      
      // Invalidate all sessions and refresh tokens
      this.invalidateAllUserSessions(userId);
      
      return {
        success: true,
        message: 'Password changed successfully. Please login again.'
      };
      
    } catch (error) {
      console.error('Password change error:', error);
      return { success: false, error: 'Failed to change password' };
    }
  }
  
  // Helper methods
  generateAccessToken(user) {
    return jwt.sign(
      {
        userId: user.id,
        email: user.email,
        role: user.role
      },
      this.jwtSecret,
      {
        expiresIn: '1h',
        issuer: 'apidirect',
        audience: 'apidirect-console'
      }
    );
  }
  
  generateRefreshToken(user) {
    const jti = crypto.randomBytes(16).toString('hex');
    const token = jwt.sign(
      {
        userId: user.id,
        jti
      },
      this.jwtSecret,
      {
        expiresIn: '30d',
        issuer: 'apidirect'
      }
    );
    
    // Store refresh token
    this.refreshTokens.set(jti, {
      userId: user.id,
      createdAt: new Date(),
      expiresAt: new Date(Date.now() + this.refreshTokenExpiry)
    });
    
    return token;
  }
  
  generateVerificationToken(userId) {
    return jwt.sign(
      { userId, type: 'email-verification' },
      this.jwtSecret,
      { expiresIn: '24h' }
    );
  }
  
  checkLoginAttempts(ip) {
    const attempts = this.loginAttempts.get(ip) || { count: 0, firstAttempt: Date.now() };
    
    // Reset after 5 minutes
    if (Date.now() - attempts.firstAttempt > 5 * 60 * 1000) {
      this.loginAttempts.delete(ip);
      return true;
    }
    
    return attempts.count < 5;
  }
  
  recordFailedAttempt(ip) {
    const attempts = this.loginAttempts.get(ip) || { count: 0, firstAttempt: Date.now() };
    attempts.count++;
    this.loginAttempts.set(ip, attempts);
  }
  
  clearFailedAttempts(ip) {
    this.loginAttempts.delete(ip);
  }
  
  invalidateAllUserSessions(userId) {
    // Remove all sessions
    for (const [sessionId, session] of securityMiddleware.sessionStore.entries()) {
      if (session.userId === userId) {
        securityMiddleware.sessionStore.delete(sessionId);
      }
    }
    
    // Revoke all refresh tokens
    for (const [jti, token] of this.refreshTokens.entries()) {
      if (token.userId === userId) {
        this.refreshTokens.delete(jti);
      }
    }
  }
  
  // Verify email
  async verifyEmail(token) {
    try {
      const decoded = jwt.verify(token, this.jwtSecret);
      
      if (decoded.type !== 'email-verification') {
        return { success: false, error: 'Invalid token type' };
      }
      
      // Find user and update
      for (const user of this.users.values()) {
        if (user.id === decoded.userId) {
          user.emailVerified = true;
          return { success: true, message: 'Email verified successfully' };
        }
      }
      
      return { success: false, error: 'User not found' };
      
    } catch (error) {
      return { success: false, error: 'Invalid or expired token' };
    }
  }
}

// Export singleton instance
module.exports = new AuthHandler();