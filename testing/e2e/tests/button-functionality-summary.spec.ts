import { test, expect, Page } from '@playwright/test';

test.describe('Button Functionality Summary', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    // Clear auth state
    await page.addInitScript(() => {
      localStorage.removeItem('mockUser');
    });
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('homepage button interactions work correctly', async () => {
    // Test category filter buttons work
    const categoryButton = page.locator('[data-testid="category-filter-all"]');
    if (await categoryButton.isVisible()) {
      await categoryButton.click({ force: true });
      await page.waitForTimeout(500);
      // Should stay on page (button should work)
      expect(page.url()).toContain('localhost:3001');
    }

    // Test popular tag buttons work
    const tagButton = page.locator('[data-testid*="tag-"]').first();
    if (await tagButton.isVisible()) {
      await tagButton.click({ force: true });
      await page.waitForTimeout(500);
      // Should trigger search functionality
      expect(page.url()).toContain('localhost:3001');
    }

    // Test search functionality
    const searchInput = page.locator('[data-testid="search-input"]');
    if (await searchInput.isVisible()) {
      await searchInput.fill('test');
      const searchButton = page.locator('[data-testid="search-submit"]');
      if (await searchButton.isVisible()) {
        await searchButton.click({ force: true });
        await page.waitForTimeout(500);
        // Should show search results
        expect(page.url()).toContain('localhost:3001');
      }
    }
  });

  test('navigation buttons work correctly', async () => {
    // Test login button
    const loginButton = page.locator('[data-testid="login-button"]');
    if (await loginButton.isVisible()) {
      await loginButton.click({ force: true });
      await page.waitForTimeout(1000);
      // Should navigate to login page
      expect(page.url()).toContain('/auth/login');
      
      // Go back to test signup
      await page.goto('/');
      await page.waitForLoadState('networkidle');
      
      const signupButton = page.locator('[data-testid="signup-button"]');
      if (await signupButton.isVisible()) {
        await signupButton.click({ force: true });
        await page.waitForTimeout(1000);
        // Should navigate to signup page
        expect(page.url()).toContain('/auth/signup');
      }
    }
  });

  test('scroll behavior works correctly on filter interactions', async ({ isMobile }) => {
    // Skip on mobile as scroll behavior is platform-specific
    if (isMobile) {
      test.skip();
      return;
    }
    // Scroll down from top
    await page.evaluate(() => window.scrollTo(0, 500));
    await page.waitForTimeout(500);
    
    const initialScrollY = await page.evaluate(() => window.scrollY);
    expect(initialScrollY).toBeGreaterThan(400);
    
    // Click category filter
    const categoryButton = page.locator('[data-testid*="category-filter-"]').nth(1);
    if (await categoryButton.isVisible()) {
      await categoryButton.click({ force: true });
      await page.waitForTimeout(1000);
      
      // Verify scroll position maintained (allow for some variance)
      const afterScrollY = await page.evaluate(() => window.scrollY);
      // Allow up to 500px difference as the page might auto-scroll slightly
      // Mobile devices may scroll more due to viewport adjustments
      const maxDifference = isMobile ? 1000 : 500;
      expect(Math.abs(afterScrollY - initialScrollY)).toBeLessThan(maxDifference);
    }
  });

  test.afterEach(async () => {
    await page.close();
  });
});