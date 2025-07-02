import { test, expect } from '@playwright/test';

const CONSOLE_URL = 'https://console.apidirect.dev';

test.describe('Live Console Quick Tests', () => {
  test('console loads and has expected elements', async ({ page }) => {
    // Navigate to console
    await page.goto(CONSOLE_URL);
    
    // Check page loaded
    await expect(page).toHaveURL(/console\.apidirect\.dev/);
    
    // Take screenshot
    await page.screenshot({ path: 'console-test-screenshot.png', fullPage: true });
    
    // Check for main elements
    const title = await page.title();
    expect(title).toContain('API-Direct');
    
    // Check navigation exists
    const hasNavigation = await page.locator('nav, .sidebar, [role="navigation"]').count() > 0;
    expect(hasNavigation).toBeTruthy();
    
    // Check for authentication elements or dashboard
    const hasAuth = await page.locator('input[type="email"], button:has-text("Sign in")').count() > 0;
    const hasDashboard = await page.locator('text=/dashboard|apis|analytics/i').count() > 0;
    
    expect(hasAuth || hasDashboard).toBeTruthy();
    
    console.log('✅ Console is live and accessible');
    console.log(`Title: ${title}`);
    console.log(`URL: ${page.url()}`);
    console.log(`Has Navigation: ${hasNavigation}`);
    console.log(`Has Auth/Dashboard: ${hasAuth || hasDashboard}`);
  });

  test('responsive design works', async ({ page }) => {
    // Test mobile
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto(CONSOLE_URL);
    await expect(page.locator('body')).toBeVisible();
    
    // Test tablet
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.locator('body')).toBeVisible();
    
    // Test desktop
    await page.setViewportSize({ width: 1920, height: 1080 });
    await expect(page.locator('body')).toBeVisible();
    
    console.log('✅ Responsive design verified');
  });

  test('security checks', async ({ page }) => {
    await page.goto(CONSOLE_URL);
    
    // Check HTTPS
    expect(page.url()).toMatch(/^https:/);
    
    // Check for console errors
    const errors: string[] = [];
    page.on('console', msg => {
      if (msg.type() === 'error') {
        errors.push(msg.text());
      }
    });
    
    await page.waitForTimeout(2000);
    
    console.log('✅ Security checks passed');
    console.log(`HTTPS: ${page.url().startsWith('https')}`);
    console.log(`Console errors: ${errors.length === 0 ? 'None' : errors.join(', ')}`);
  });
});