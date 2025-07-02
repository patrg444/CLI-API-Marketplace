import { test, expect } from '@playwright/test';
import crypto from 'crypto';

// Security test configuration
const SECURITY_TESTS = {
  sqlInjectionPayloads: [
    "' OR '1'='1",
    "admin'--",
    "1' UNION SELECT NULL--",
    "'; DROP TABLE users;--",
    "1' OR '1'='1' /*"
  ],
  xssPayloads: [
    '<script>alert("XSS")</script>',
    'javascript:alert("XSS")',
    '<img src=x onerror=alert("XSS")>',
    '<svg onload=alert("XSS")>',
    '"><script>alert(String.fromCharCode(88,83,83))</script>'
  ],
  bruteForceAttempts: 10,
  passwordComplexityTests: [
    { password: '123456', shouldFail: true, reason: 'Too simple' },
    { password: 'password', shouldFail: true, reason: 'Common password' },
    { password: 'Pass123!', shouldFail: true, reason: 'Too short' },
    { password: 'ValidPassword123!', shouldFail: false, reason: 'Valid password' }
  ]
};

test.describe('Authentication Security Tests', () => {
  
  test.describe('SQL Injection Prevention', () => {
    for (const payload of SECURITY_TESTS.sqlInjectionPayloads) {
      test(`should prevent SQL injection with payload: ${payload.substring(0, 20)}...`, async ({ page }) => {
        await page.goto('/login');
        
        // Try SQL injection in email field
        await page.fill('[name="email"]', payload);
        await page.fill('[name="password"]', 'anypassword');
        await page.click('[type="submit"]');
        
        // Should not log in or cause database error
        await expect(page.locator('.error-message')).toBeVisible();
        await expect(page).not.toHaveURL(/dashboard/);
        
        // Check that error message doesn't reveal database structure
        const errorText = await page.locator('.error-message').textContent();
        expect(errorText).not.toContain('SQL');
        expect(errorText).not.toContain('database');
        expect(errorText).not.toContain('syntax');
      });
    }
  });
  
  test.describe('XSS Prevention', () => {
    for (const payload of SECURITY_TESTS.xssPayloads) {
      test(`should prevent XSS with payload: ${payload.substring(0, 20)}...`, async ({ page }) => {
        await page.goto('/register');
        
        // Try XSS in username field
        await page.fill('[name="username"]', payload);
        await page.fill('[name="email"]', 'test@example.com');
        await page.fill('[name="password"]', 'ValidPassword123!');
        
        // Listen for any alert dialogs (XSS indicator)
        let alertFired = false;
        page.on('dialog', async dialog => {
          alertFired = true;
          await dialog.dismiss();
        });
        
        await page.click('[type="submit"]');
        
        // Wait to see if XSS executes
        await page.waitForTimeout(2000);
        
        expect(alertFired).toBe(false);
        
        // If registration succeeds, check profile page
        if (await page.url().includes('dashboard')) {
          await page.goto('/profile');
          
          // Username should be escaped
          const displayedUsername = await page.locator('.username-display').textContent();
          expect(displayedUsername).not.toContain('<script>');
          expect(displayedUsername).not.toContain('<img');
          expect(displayedUsername).not.toContain('javascript:');
        }
      });
    }
  });
  
  test.describe('Brute Force Protection', () => {
    test('should implement rate limiting after failed login attempts', async ({ page }) => {
      await page.goto('/login');
      
      const attempts = [];
      
      // Try multiple failed login attempts
      for (let i = 0; i < SECURITY_TESTS.bruteForceAttempts; i++) {
        const startTime = Date.now();
        
        await page.fill('[name="email"]', 'bruteforce@test.com');
        await page.fill('[name="password"]', `wrongpassword${i}`);
        await page.click('[type="submit"]');
        
        const endTime = Date.now();
        attempts.push(endTime - startTime);
        
        // Check for rate limiting
        if (i >= 5) {
          // Should see rate limit message or CAPTCHA
          const rateLimitMessage = page.locator('.rate-limit-message');
          const captcha = page.locator('.captcha-container');
          
          const hasRateLimit = await rateLimitMessage.isVisible() || await captcha.isVisible();
          
          if (hasRateLimit) {
            console.log(`Rate limiting activated after ${i + 1} attempts`);
            break;
          }
        }
        
        // Small delay between attempts
        await page.waitForTimeout(500);
      }
      
      // Verify that rate limiting was applied
      const hasRateLimit = await page.locator('.rate-limit-message').isVisible() || 
                          await page.locator('.captcha-container').isVisible();
      expect(hasRateLimit).toBe(true);
    });
  });
  
  test.describe('Password Security', () => {
    for (const testCase of SECURITY_TESTS.passwordComplexityTests) {
      test(`should ${testCase.shouldFail ? 'reject' : 'accept'} password: ${testCase.reason}`, async ({ page }) => {
        await page.goto('/register');
        
        await page.fill('[name="username"]', 'testuser');
        await page.fill('[name="email"]', 'test@example.com');
        await page.fill('[name="password"]', testCase.password);
        
        // Check real-time password validation
        const passwordError = page.locator('.password-error');
        
        if (testCase.shouldFail) {
          await expect(passwordError).toBeVisible();
          
          // Try to submit anyway
          await page.click('[type="submit"]');
          
          // Should not proceed
          await expect(page).not.toHaveURL(/dashboard/);
        } else {
          await expect(passwordError).not.toBeVisible();
        }
      });
    }
  });
  
  test.describe('Session Security', () => {
    test('should invalidate session after password change', async ({ page, context }) => {
      // First, log in
      await page.goto('/login');
      await page.fill('[name="email"]', 'test@example.com');
      await page.fill('[name="password"]', 'ValidPassword123!');
      await page.click('[type="submit"]');
      await expect(page).toHaveURL(/dashboard/);
      
      // Get session cookie
      const cookies = await context.cookies();
      const sessionCookie = cookies.find(c => c.name === 'session' || c.name === 'auth_token');
      expect(sessionCookie).toBeTruthy();
      
      // Change password
      await page.goto('/settings/security');
      await page.fill('[name="currentPassword"]', 'ValidPassword123!');
      await page.fill('[name="newPassword"]', 'NewValidPassword123!');
      await page.fill('[name="confirmPassword"]', 'NewValidPassword123!');
      await page.click('[type="submit"]');
      
      // Old session should be invalidated
      await page.goto('/dashboard');
      await expect(page).toHaveURL(/login/);
    });
    
    test('should use secure cookie flags', async ({ page, context }) => {
      await page.goto('/login');
      await page.fill('[name="email"]', 'test@example.com');
      await page.fill('[name="password"]', 'ValidPassword123!');
      await page.click('[type="submit"]');
      
      const cookies = await context.cookies();
      const authCookies = cookies.filter(c => 
        c.name === 'session' || c.name === 'auth_token' || c.name.includes('auth')
      );
      
      authCookies.forEach(cookie => {
        expect(cookie.httpOnly).toBe(true);
        expect(cookie.sameSite).toBe('Strict' || 'Lax');
        // In production, should also check: expect(cookie.secure).toBe(true);
      });
    });
  });
  
  test.describe('CSRF Protection', () => {
    test('should include CSRF token in forms', async ({ page }) => {
      await page.goto('/login');
      
      // Check for CSRF token in form
      const csrfToken = await page.locator('[name="csrf_token"], [name="_csrf"], input[type="hidden"][name*="csrf"]').first();
      await expect(csrfToken).toBeAttached();
      
      const tokenValue = await csrfToken.getAttribute('value');
      expect(tokenValue).toBeTruthy();
      expect(tokenValue.length).toBeGreaterThan(20);
    });
    
    test('should reject requests without valid CSRF token', async ({ page }) => {
      // Attempt to submit form without CSRF token
      const response = await page.evaluate(async () => {
        const formData = new FormData();
        formData.append('email', 'test@example.com');
        formData.append('password', 'password');
        
        const res = await fetch('/api/auth/login', {
          method: 'POST',
          body: formData
        });
        
        return {
          status: res.status,
          text: await res.text()
        };
      });
      
      expect(response.status).toBe(403);
    });
  });
  
  test.describe('Authorization Checks', () => {
    test('should prevent unauthorized access to admin endpoints', async ({ page }) => {
      // Try to access admin area without auth
      await page.goto('/admin/dashboard');
      await expect(page).toHaveURL(/login/);
      
      // Log in as regular user
      await page.fill('[name="email"]', 'user@example.com');
      await page.fill('[name="password"]', 'UserPassword123!');
      await page.click('[type="submit"]');
      
      // Try to access admin area as regular user
      await page.goto('/admin/dashboard');
      await expect(page).not.toHaveURL(/admin/);
      
      // Should see error or be redirected
      const errorMessage = page.locator('.error-message, .unauthorized-message');
      const isUnauthorized = await errorMessage.isVisible() || !page.url().includes('/admin');
      expect(isUnauthorized).toBe(true);
    });
  });
  
  test.describe('API Security Headers', () => {
    test('should include security headers in responses', async ({ page }) => {
      const response = await page.goto('/');
      const headers = response.headers();
      
      // Check for security headers
      expect(headers['x-content-type-options']).toBe('nosniff');
      expect(headers['x-frame-options']).toMatch(/DENY|SAMEORIGIN/);
      expect(headers['x-xss-protection']).toBe('1; mode=block');
      
      // Check for CSP
      const csp = headers['content-security-policy'];
      if (csp) {
        expect(csp).toContain("default-src");
        expect(csp).not.toContain("unsafe-inline");
        expect(csp).not.toContain("unsafe-eval");
      }
    });
  });
  
  test.describe('Input Validation', () => {
    test('should validate email format', async ({ page }) => {
      await page.goto('/register');
      
      const invalidEmails = ['notanemail', '@example.com', 'user@', 'user..name@example.com'];
      
      for (const email of invalidEmails) {
        await page.fill('[name="email"]', email);
        await page.fill('[name="password"]', 'ValidPassword123!');
        
        const emailError = page.locator('.email-error');
        await expect(emailError).toBeVisible();
      }
    });
    
    test('should sanitize user input in profile', async ({ page }) => {
      // Log in first
      await page.goto('/login');
      await page.fill('[name="email"]', 'test@example.com');
      await page.fill('[name="password"]', 'ValidPassword123!');
      await page.click('[type="submit"]');
      
      // Update profile with potentially dangerous input
      await page.goto('/profile/edit');
      
      const dangerousInputs = {
        bio: '<script>alert("XSS")</script>Some bio text',
        website: 'javascript:alert("XSS")',
        location: '"><img src=x onerror=alert("XSS")>'
      };
      
      await page.fill('[name="bio"]', dangerousInputs.bio);
      await page.fill('[name="website"]', dangerousInputs.website);
      await page.fill('[name="location"]', dangerousInputs.location);
      await page.click('[type="submit"]');
      
      // Check that inputs are sanitized on display
      await page.goto('/profile');
      
      const bioText = await page.locator('.bio-display').textContent();
      expect(bioText).not.toContain('<script>');
      
      const websiteLink = await page.locator('.website-link').getAttribute('href');
      expect(websiteLink).not.toContain('javascript:');
      
      const locationText = await page.locator('.location-display').textContent();
      expect(locationText).not.toContain('<img');
    });
  });
});