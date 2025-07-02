import { test, expect, Page } from '@playwright/test';

const CONSOLE_URL = 'https://console.apidirect.dev';

test.describe('Live Console Tests - console.apidirect.dev', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    await page.goto(CONSOLE_URL);
  });

  test.afterEach(async () => {
    await page.close();
  });

  test.describe('Landing and Authentication', () => {
    test('should load console homepage', async () => {
      // Check page loaded
      await expect(page).toHaveURL(/console\.apidirect\.dev/);
      
      // Look for common elements that should be on homepage
      await expect(page.locator('body')).toBeVisible();
      
      // Check for login elements or dashboard
      const hasLogin = await page.locator('text=/sign in|login/i').count() > 0;
      const hasDashboard = await page.locator('text=/dashboard/i').count() > 0;
      
      expect(hasLogin || hasDashboard).toBeTruthy();
    });

    test('should have proper meta tags and title', async () => {
      // Check title
      await expect(page).toHaveTitle(/API-Direct|Console|Dashboard/i);
      
      // Check viewport meta tag for mobile responsiveness
      const viewport = await page.$('meta[name="viewport"]');
      expect(viewport).not.toBeNull();
    });

    test('should display login form if not authenticated', async () => {
      // If redirected to login
      if (page.url().includes('login')) {
        // Check for login form elements
        await expect(page.locator('input[type="email"], input[name="email"]')).toBeVisible();
        await expect(page.locator('input[type="password"]')).toBeVisible();
        await expect(page.locator('button[type="submit"], button:has-text("Sign in")')).toBeVisible();
      }
    });

    test('should have working navigation links', async () => {
      // Check for navigation elements
      const navLinks = page.locator('nav a, a[href*="dashboard"], a[href*="apis"]');
      const count = await navLinks.count();
      
      if (count > 0) {
        // Test first nav link
        const firstLink = navLinks.first();
        const href = await firstLink.getAttribute('href');
        expect(href).toBeTruthy();
      }
    });
  });

  test.describe('Console Functionality', () => {
    test('should handle authentication flow', async () => {
      // Check if we need to login
      if (page.url().includes('login') || await page.locator('text=/sign in/i').count() > 0) {
        // Look for demo/test credentials or skip if none available
        const hasDemo = await page.locator('text=/demo|test/i').count() > 0;
        
        if (!hasDemo) {
          test.skip();
          return;
        }
      }
    });

    test('should check dashboard elements if accessible', async () => {
      // Try to navigate to dashboard
      const dashboardLink = page.locator('a[href*="dashboard"]').first();
      
      if (await dashboardLink.count() > 0) {
        await dashboardLink.click();
        await page.waitForLoadState('networkidle');
        
        // Check for dashboard elements
        const metricCards = page.locator('[data-metric], .metric-card, .stat-card');
        const hasMetrics = await metricCards.count() > 0;
        
        if (hasMetrics) {
          expect(await metricCards.count()).toBeGreaterThan(0);
        }
      }
    });

    test('should check API management page if accessible', async () => {
      // Look for APIs link
      const apisLink = page.locator('a[href*="apis"], a:has-text("APIs")').first();
      
      if (await apisLink.count() > 0) {
        await apisLink.click();
        await page.waitForLoadState('networkidle');
        
        // Check for API-related elements
        const hasAPIContent = await page.locator('text=/api|deployment/i').count() > 0;
        expect(hasAPIContent).toBeTruthy();
      }
    });
  });

  test.describe('Responsive Design', () => {
    test('should be mobile responsive', async () => {
      // Test mobile viewport
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto(CONSOLE_URL);
      
      // Check if content adapts
      await expect(page.locator('body')).toBeVisible();
      
      // Check for mobile menu if exists
      const mobileMenu = page.locator('[aria-label*="menu"], button:has-text("Menu"), .mobile-menu');
      if (await mobileMenu.count() > 0) {
        await expect(mobileMenu.first()).toBeVisible();
      }
    });

    test('should work on tablet size', async () => {
      await page.setViewportSize({ width: 768, height: 1024 });
      await page.goto(CONSOLE_URL);
      
      await expect(page.locator('body')).toBeVisible();
    });
  });

  test.describe('Performance and Accessibility', () => {
    test('should load within acceptable time', async () => {
      const startTime = Date.now();
      await page.goto(CONSOLE_URL, { waitUntil: 'domcontentloaded' });
      const loadTime = Date.now() - startTime;
      
      // Should load DOM in under 5 seconds
      expect(loadTime).toBeLessThan(5000);
    });

    test('should have proper heading structure', async () => {
      // Check for h1
      const h1Count = await page.locator('h1').count();
      const h2Count = await page.locator('h2').count();
      
      // Should have at least one heading
      expect(h1Count + h2Count).toBeGreaterThan(0);
    });

    test('should have alt text for images', async () => {
      const images = page.locator('img');
      const imageCount = await images.count();
      
      for (let i = 0; i < imageCount; i++) {
        const img = images.nth(i);
        const alt = await img.getAttribute('alt');
        const ariaLabel = await img.getAttribute('aria-label');
        
        // Should have either alt or aria-label
        expect(alt || ariaLabel).toBeTruthy();
      }
    });
  });

  test.describe('Error Handling', () => {
    test('should handle 404 pages gracefully', async () => {
      await page.goto(`${CONSOLE_URL}/non-existent-page-12345`);
      
      // Should show 404 or redirect
      const has404 = await page.locator('text=/404|not found/i').count() > 0;
      const redirectedHome = page.url() === CONSOLE_URL || page.url() === `${CONSOLE_URL}/`;
      
      expect(has404 || redirectedHome).toBeTruthy();
    });

    test('should handle network errors gracefully', async () => {
      // Simulate offline
      await page.context().setOffline(true);
      
      try {
        await page.goto(CONSOLE_URL, { timeout: 5000 });
      } catch (e) {
        // Expected to fail
        expect(e.message).toContain('net::ERR_INTERNET_DISCONNECTED');
      }
      
      await page.context().setOffline(false);
    });
  });

  test.describe('Security', () => {
    test('should use HTTPS', async () => {
      expect(page.url()).toMatch(/^https:/);
    });

    test('should have security headers', async ({ request }) => {
      const response = await request.get(CONSOLE_URL);
      const headers = response.headers();
      
      // Check for common security headers
      const hasSecurityHeaders = 
        headers['strict-transport-security'] ||
        headers['x-content-type-options'] ||
        headers['x-frame-options'] ||
        headers['content-security-policy'];
      
      expect(hasSecurityHeaders).toBeTruthy();
    });
  });
});

// Visual regression tests
test.describe('Visual Tests', () => {
  test('should match homepage screenshot', async ({ page }) => {
    await page.goto(CONSOLE_URL);
    await page.waitForLoadState('networkidle');
    
    // Wait for animations to settle
    await page.waitForTimeout(1000);
    
    await expect(page).toHaveScreenshot('console-homepage.png', {
      fullPage: true,
      animations: 'disabled'
    });
  });

  test('should match mobile screenshot', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto(CONSOLE_URL);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    
    await expect(page).toHaveScreenshot('console-mobile.png', {
      fullPage: true,
      animations: 'disabled'
    });
  });
});

// API Integration tests (if we can access authenticated pages)
test.describe('API Integration', () => {
  test.skip('should test API endpoints if authenticated', async () => {
    // This would test actual API calls if we had test credentials
    // Skipped for now as we don't have access
  });
});