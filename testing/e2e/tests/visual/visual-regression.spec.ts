import { test, expect } from '@playwright/test';

// Visual regression test configuration
const VIEWPORTS = [
  { name: 'desktop', width: 1920, height: 1080 },
  { name: 'laptop', width: 1366, height: 768 },
  { name: 'tablet', width: 768, height: 1024 },
  { name: 'mobile', width: 375, height: 667 }
];

const THEMES = ['light', 'dark'];

test.describe('Visual Regression Tests', () => {
  
  test.describe('Landing Page Components', () => {
    VIEWPORTS.forEach(viewport => {
      test(`should match landing page snapshot on ${viewport.name}`, async ({ page }) => {
        await page.setViewportSize(viewport);
        await page.goto('/');
        
        // Wait for animations to complete
        await page.waitForTimeout(1000);
        
        // Take full page screenshot
        await expect(page).toHaveScreenshot(`landing-${viewport.name}.png`, {
          fullPage: true,
          animations: 'disabled',
          mask: [page.locator('.dynamic-content')] // Mask dynamic content
        });
        
        // Test hero section separately
        const hero = page.locator('.hero-section');
        await expect(hero).toHaveScreenshot(`hero-${viewport.name}.png`);
        
        // Test feature cards
        const features = page.locator('.features-section');
        await expect(features).toHaveScreenshot(`features-${viewport.name}.png`);
      });
    });
  });
  
  test.describe('Marketplace UI Components', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/marketplace');
      await page.waitForLoadState('networkidle');
    });
    
    test('should match API card component styles', async ({ page }) => {
      const apiCard = page.locator('.api-card').first();
      await expect(apiCard).toHaveScreenshot('api-card.png');
      
      // Test hover state
      await apiCard.hover();
      await page.waitForTimeout(300); // Wait for transition
      await expect(apiCard).toHaveScreenshot('api-card-hover.png');
    });
    
    test('should match search and filter components', async ({ page }) => {
      const searchBar = page.locator('.search-container');
      await expect(searchBar).toHaveScreenshot('search-bar.png');
      
      // Test with search input
      await page.fill('[data-testid="search-input"]', 'weather api');
      await expect(searchBar).toHaveScreenshot('search-bar-filled.png');
      
      // Test filter dropdown
      const filterSection = page.locator('.filter-section');
      await expect(filterSection).toHaveScreenshot('filters-closed.png');
      
      await page.click('[data-testid="filter-toggle"]');
      await expect(filterSection).toHaveScreenshot('filters-open.png');
    });
    
    test('should match loading states', async ({ page }) => {
      // Trigger loading state
      await page.evaluate(() => {
        document.querySelectorAll('.api-card').forEach(card => {
          card.classList.add('loading');
        });
      });
      
      const loadingCard = page.locator('.api-card.loading').first();
      await expect(loadingCard).toHaveScreenshot('api-card-loading.png');
      
      // Test skeleton loader
      const skeleton = page.locator('.skeleton-loader');
      await expect(skeleton).toHaveScreenshot('skeleton-loader.png');
    });
  });
  
  test.describe('Console Dashboard Components', () => {
    test.beforeEach(async ({ page }) => {
      // Mock login
      await page.goto('/console/dashboard');
      await page.evaluate(() => {
        localStorage.setItem('auth_token', 'mock_token');
      });
      await page.reload();
    });
    
    test('should match dashboard metric cards', async ({ page }) => {
      const metricsSection = page.locator('.metrics-grid');
      await expect(metricsSection).toHaveScreenshot('dashboard-metrics.png');
      
      // Test individual metric cards
      const metricTypes = ['revenue', 'api-calls', 'active-users', 'growth'];
      for (const type of metricTypes) {
        const card = page.locator(`[data-metric="${type}"]`);
        await expect(card).toHaveScreenshot(`metric-${type}.png`);
      }
    });
    
    test('should match chart components', async ({ page }) => {
      // Wait for charts to render
      await page.waitForFunction(() => {
        return document.querySelector('canvas') !== null;
      });
      
      const usageChart = page.locator('#usage-chart-container');
      await expect(usageChart).toHaveScreenshot('usage-chart.png', {
        animations: 'disabled'
      });
      
      const revenueChart = page.locator('#revenue-chart-container');
      await expect(revenueChart).toHaveScreenshot('revenue-chart.png', {
        animations: 'disabled'
      });
    });
  });
  
  test.describe('Form Components', () => {
    test('should match form input states', async ({ page }) => {
      await page.goto('/console/login');
      
      const form = page.locator('form');
      
      // Default state
      await expect(form).toHaveScreenshot('login-form-default.png');
      
      // Focused state
      await page.focus('[name="email"]');
      await expect(form).toHaveScreenshot('login-form-focused.png');
      
      // Error state
      await page.fill('[name="email"]', 'invalid-email');
      await page.fill('[name="password"]', '123');
      await page.click('[type="submit"]');
      await page.waitForSelector('.error-message');
      await expect(form).toHaveScreenshot('login-form-error.png');
      
      // Filled state
      await page.fill('[name="email"]', 'user@example.com');
      await page.fill('[name="password"]', 'ValidPassword123!');
      await expect(form).toHaveScreenshot('login-form-filled.png');
    });
  });
  
  test.describe('Dark Mode Support', () => {
    test('should match components in dark mode', async ({ page }) => {
      // Enable dark mode
      await page.goto('/');
      await page.evaluate(() => {
        document.documentElement.classList.add('dark');
        localStorage.setItem('theme', 'dark');
      });
      
      // Test key components in dark mode
      await expect(page.locator('.navbar')).toHaveScreenshot('navbar-dark.png');
      await expect(page.locator('.hero-section')).toHaveScreenshot('hero-dark.png');
      
      // Test marketplace in dark mode
      await page.goto('/marketplace');
      await expect(page.locator('.api-card').first()).toHaveScreenshot('api-card-dark.png');
      
      // Test console in dark mode
      await page.goto('/console/dashboard');
      await expect(page.locator('.sidebar')).toHaveScreenshot('sidebar-dark.png');
    });
  });
  
  test.describe('Responsive Design', () => {
    test('should handle responsive navigation', async ({ page }) => {
      // Desktop navigation
      await page.setViewportSize({ width: 1920, height: 1080 });
      await page.goto('/');
      await expect(page.locator('.navbar')).toHaveScreenshot('nav-desktop.png');
      
      // Mobile navigation
      await page.setViewportSize({ width: 375, height: 667 });
      await expect(page.locator('.navbar')).toHaveScreenshot('nav-mobile-closed.png');
      
      // Open mobile menu
      await page.click('[data-testid="mobile-menu-toggle"]');
      await page.waitForTimeout(300); // Wait for animation
      await expect(page.locator('.mobile-menu')).toHaveScreenshot('nav-mobile-open.png');
    });
    
    test('should handle responsive grid layouts', async ({ page }) => {
      await page.goto('/marketplace');
      
      // Test different grid layouts
      const viewportTests = [
        { width: 1920, cols: 4, name: 'desktop-4col' },
        { width: 1366, cols: 3, name: 'laptop-3col' },
        { width: 768, cols: 2, name: 'tablet-2col' },
        { width: 375, cols: 1, name: 'mobile-1col' }
      ];
      
      for (const viewport of viewportTests) {
        await page.setViewportSize({ width: viewport.width, height: 800 });
        await page.waitForTimeout(500); // Wait for reflow
        
        const grid = page.locator('.api-grid');
        await expect(grid).toHaveScreenshot(`grid-${viewport.name}.png`);
      }
    });
  });
  
  test.describe('Animation and Transition States', () => {
    test('should capture button interaction states', async ({ page }) => {
      await page.goto('/marketplace');
      
      const button = page.locator('.primary-button').first();
      
      // Default state
      await expect(button).toHaveScreenshot('button-default.png');
      
      // Hover state
      await button.hover();
      await page.waitForTimeout(150);
      await expect(button).toHaveScreenshot('button-hover.png');
      
      // Active state
      await page.mouse.down();
      await expect(button).toHaveScreenshot('button-active.png');
      await page.mouse.up();
      
      // Focus state
      await button.focus();
      await expect(button).toHaveScreenshot('button-focus.png');
    });
    
    test('should capture modal transitions', async ({ page }) => {
      await page.goto('/marketplace');
      
      // Open API details modal
      await page.click('.api-card .view-details-btn');
      
      // Wait for modal to be fully visible
      await page.waitForSelector('.modal.visible');
      await page.waitForTimeout(300); // Wait for animation
      
      const modal = page.locator('.modal');
      await expect(modal).toHaveScreenshot('modal-open.png');
      
      // Test modal content scrolling
      await page.evaluate(() => {
        const content = document.querySelector('.modal-content');
        if (content) content.scrollTop = 100;
      });
      await expect(modal).toHaveScreenshot('modal-scrolled.png');
    });
  });
  
  test.describe('Error States and Edge Cases', () => {
    test('should match empty state designs', async ({ page }) => {
      await page.goto('/marketplace?search=nonexistentapi12345');
      await page.waitForSelector('.empty-state');
      
      const emptyState = page.locator('.empty-state');
      await expect(emptyState).toHaveScreenshot('empty-state-search.png');
    });
    
    test('should match error page designs', async ({ page }) => {
      await page.goto('/nonexistent-page');
      await expect(page).toHaveScreenshot('404-page.png', {
        fullPage: true
      });
    });
    
    test('should match offline state', async ({ page, context }) => {
      await context.setOffline(true);
      await page.goto('/marketplace');
      
      const offlineMessage = page.locator('.offline-banner');
      await expect(offlineMessage).toHaveScreenshot('offline-banner.png');
    });
  });
});

// Helper function to setup visual regression baseline
export async function setupVisualBaseline(page: Page) {
  // Add CSS to ensure consistent rendering
  await page.addStyleTag({
    content: `
      * {
        -webkit-font-smoothing: antialiased;
        -moz-osx-font-smoothing: grayscale;
      }
      
      /* Disable animations for consistent screenshots */
      *, *::before, *::after {
        animation-duration: 0s !important;
        animation-delay: 0s !important;
        transition-duration: 0s !important;
        transition-delay: 0s !important;
      }
      
      /* Ensure consistent scrollbar rendering */
      ::-webkit-scrollbar {
        width: 12px;
        height: 12px;
      }
      
      ::-webkit-scrollbar-track {
        background: #f1f1f1;
      }
      
      ::-webkit-scrollbar-thumb {
        background: #888;
        border-radius: 6px;
      }
    `
  });
}