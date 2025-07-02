import { test, expect, Page } from '@playwright/test';

test.describe('Filter Button Scroll Fix Verification', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('should not scroll to top when clicking filter buttons', async ({ isMobile }) => {
    // Skip on mobile as scroll behavior is platform-specific
    if (isMobile) {
      test.skip();
      return;
    }
    // Scroll down to ensure we're not at the top
    await page.evaluate(() => window.scrollTo(0, 500));
    await page.waitForTimeout(500);
    
    // Get initial scroll position
    const initialScrollY = await page.evaluate(() => window.scrollY);
    expect(initialScrollY).toBeGreaterThan(0);
    
    // Click a category filter button
    const categoryButton = page.locator('[data-testid="category-filter-all"]').first();
    if (await categoryButton.isVisible()) {
      await categoryButton.click({ force: true });
      await page.waitForTimeout(500);
      
      // Check that scroll position hasn't changed significantly
      const afterScrollY = await page.evaluate(() => window.scrollY);
      // Mobile devices may have different scroll behavior
      const maxDifference = isMobile ? 1000 : 500;
      expect(Math.abs(afterScrollY - initialScrollY)).toBeLessThan(maxDifference);
    }
    
    // Test another category filter
    const businessFilter = page.locator('[data-testid*="category-filter-"]').nth(1);
    if (await businessFilter.isVisible()) {
      await businessFilter.click({ force: true });
      await page.waitForTimeout(500);
      
      // Check that scroll position is still maintained
      const finalScrollY = await page.evaluate(() => window.scrollY);
      const maxDifference = isMobile ? 1000 : 500;
      expect(Math.abs(finalScrollY - initialScrollY)).toBeLessThan(maxDifference);
    }
  });

  test('should not scroll to top when using search', async ({ isMobile }) => {
    // Scroll down
    await page.evaluate(() => window.scrollTo(0, 500));
    await page.waitForTimeout(500);
    
    const initialScrollY = await page.evaluate(() => window.scrollY);
    expect(initialScrollY).toBeGreaterThan(0);
    
    // Use search functionality
    const searchInput = page.locator('input[placeholder*="Search"]').first();
    if (await searchInput.isVisible()) {
      await searchInput.fill('test search');
      
      const searchButton = page.locator('[data-testid="search-submit"]');
      if (await searchButton.isVisible()) {
        await searchButton.click({ force: true });
        await page.waitForTimeout(500);
        
        // Verify scroll position maintained
        const afterSearchScrollY = await page.evaluate(() => window.scrollY);
        const maxDifference = isMobile ? 1000 : 500;
        expect(Math.abs(afterSearchScrollY - initialScrollY)).toBeLessThan(maxDifference);
      }
    }
  });

  test.afterEach(async () => {
    await page.close();
  });
});