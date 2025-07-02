import { test, expect, Page } from '@playwright/test';

// Test data
const TEST_USER = {
  email: 'console-test@example.com',
  password: 'TestPassword123!',
  apiName: 'test-weather-api',
  apiDescription: 'Test weather API for integration testing'
};

test.describe('Console Integration Tests', () => {
  let page: Page;
  
  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    // Login before each test
    await login(page);
  });
  
  test.afterEach(async () => {
    await page.close();
  });
  
  test.describe('Dashboard Functionality', () => {
    test('should display dashboard metrics correctly', async () => {
      await page.goto('/console/pages/dashboard.html');
      
      // Check for key dashboard elements
      await expect(page.locator('h1')).toContainText('Dashboard');
      
      // Verify metric cards are present
      const metricCards = page.locator('.metric-card');
      await expect(metricCards).toHaveCount(4);
      
      // Check specific metrics
      await expect(page.locator('[data-metric="total-apis"]')).toBeVisible();
      await expect(page.locator('[data-metric="total-calls"]')).toBeVisible();
      await expect(page.locator('[data-metric="monthly-revenue"]')).toBeVisible();
      await expect(page.locator('[data-metric="active-users"]')).toBeVisible();
      
      // Verify chart containers
      await expect(page.locator('#usage-chart')).toBeVisible();
      await expect(page.locator('#revenue-chart')).toBeVisible();
    });
    
    test('should update metrics in real-time', async () => {
      await page.goto('/console/pages/dashboard.html');
      
      // Get initial call count
      const initialCalls = await page.locator('[data-metric="total-calls"] .metric-value').textContent();
      
      // Simulate API call (trigger metric update)
      await page.evaluate(() => {
        window.dispatchEvent(new CustomEvent('api-call-made', { detail: { count: 1 } }));
      });
      
      // Wait for update
      await page.waitForTimeout(1000);
      
      // Verify count increased
      const updatedCalls = await page.locator('[data-metric="total-calls"] .metric-value').textContent();
      expect(parseInt(updatedCalls || '0')).toBeGreaterThan(parseInt(initialCalls || '0'));
    });
  });
  
  test.describe('API Management', () => {
    test('should create a new API', async () => {
      await page.goto('/console/pages/apis.html');
      
      // Click create API button
      await page.click('[data-action="create-api"]');
      
      // Fill in API details
      await page.fill('#api-name', TEST_USER.apiName);
      await page.fill('#api-description', TEST_USER.apiDescription);
      await page.selectOption('#api-category', 'weather');
      
      // Set pricing
      await page.click('#pricing-freemium');
      await page.fill('#free-calls', '1000');
      await page.fill('#price-per-call', '0.001');
      
      // Submit form
      await page.click('[type="submit"]');
      
      // Verify success message
      await expect(page.locator('.success-message')).toContainText('API created successfully');
      
      // Verify API appears in list
      await expect(page.locator(`[data-api-name="${TEST_USER.apiName}"]`)).toBeVisible();
    });
    
    test('should edit API configuration', async () => {
      await page.goto('/console/pages/apis.html');
      
      // Find and click edit button for test API
      await page.click(`[data-api-name="${TEST_USER.apiName}"] [data-action="edit"]`);
      
      // Update description
      const newDescription = 'Updated weather API description';
      await page.fill('#api-description', newDescription);
      
      // Update pricing
      await page.fill('#price-per-call', '0.002');
      
      // Save changes
      await page.click('[data-action="save-changes"]');
      
      // Verify success
      await expect(page.locator('.success-message')).toContainText('API updated successfully');
      
      // Verify changes persisted
      await page.reload();
      await page.click(`[data-api-name="${TEST_USER.apiName}"] [data-action="edit"]`);
      await expect(page.locator('#api-description')).toHaveValue(newDescription);
      await expect(page.locator('#price-per-call')).toHaveValue('0.002');
    });
    
    test('should manage API keys', async () => {
      await page.goto('/console/pages/apis.html');
      
      // Navigate to API keys section
      await page.click(`[data-api-name="${TEST_USER.apiName}"] [data-action="manage-keys"]`);
      
      // Generate new key
      await page.click('[data-action="generate-key"]');
      await page.fill('#key-name', 'Test Integration Key');
      await page.click('[data-action="confirm-generate"]');
      
      // Verify key was created
      await expect(page.locator('.api-key-display')).toBeVisible();
      const apiKey = await page.locator('.api-key-display').textContent();
      expect(apiKey).toMatch(/^api_[a-zA-Z0-9]{32}$/);
      
      // Test key operations
      await page.click('[data-action="copy-key"]');
      await expect(page.locator('.copy-success')).toContainText('Copied!');
      
      // Revoke key
      await page.click('[data-action="revoke-key"]');
      await page.click('[data-action="confirm-revoke"]');
      await expect(page.locator('.success-message')).toContainText('Key revoked');
    });
  });
  
  test.describe('Analytics Features', () => {
    test('should display analytics data correctly', async () => {
      await page.goto('/console/pages/analytics.html');
      
      // Check date range selector
      await expect(page.locator('#date-range-selector')).toBeVisible();
      
      // Verify analytics sections
      await expect(page.locator('#usage-analytics')).toBeVisible();
      await expect(page.locator('#performance-metrics')).toBeVisible();
      await expect(page.locator('#geographic-distribution')).toBeVisible();
      await expect(page.locator('#endpoint-breakdown')).toBeVisible();
      
      // Test date range change
      await page.selectOption('#date-range-selector', '30d');
      await page.waitForLoadState('networkidle');
      
      // Verify data updated
      await expect(page.locator('.analytics-updated')).toBeVisible();
    });
    
    test('should export analytics data', async () => {
      await page.goto('/console/pages/analytics.html');
      
      // Test CSV export
      const [download] = await Promise.all([
        page.waitForEvent('download'),
        page.click('[data-action="export-csv"]')
      ]);
      
      expect(download.suggestedFilename()).toContain('analytics');
      expect(download.suggestedFilename()).toContain('.csv');
      
      // Test PDF report generation
      await page.click('[data-action="generate-report"]');
      await expect(page.locator('.report-generating')).toBeVisible();
      await expect(page.locator('.report-ready')).toBeVisible({ timeout: 10000 });
    });
  });
  
  test.describe('Earnings and Payouts', () => {
    test('should display earnings overview', async () => {
      await page.goto('/console/pages/earnings.html');
      
      // Check earnings summary
      await expect(page.locator('#current-balance')).toBeVisible();
      await expect(page.locator('#pending-earnings')).toBeVisible();
      await expect(page.locator('#total-earned')).toBeVisible();
      
      // Verify earnings table
      const earningsTable = page.locator('#earnings-table');
      await expect(earningsTable).toBeVisible();
      
      // Check for transaction details
      const transactions = page.locator('.transaction-row');
      const count = await transactions.count();
      expect(count).toBeGreaterThan(0);
    });
    
    test('should process payout request', async () => {
      await page.goto('/console/pages/earnings.html');
      
      // Check if balance is sufficient
      const balance = await page.locator('#current-balance').textContent();
      const balanceValue = parseFloat(balance?.replace('$', '') || '0');
      
      if (balanceValue >= 10) {
        // Request payout
        await page.click('[data-action="request-payout"]');
        
        // Fill payout details
        await page.fill('#payout-amount', '10.00');
        await page.selectOption('#payout-method', 'bank_transfer');
        
        // Submit request
        await page.click('[data-action="submit-payout"]');
        
        // Verify confirmation
        await expect(page.locator('.payout-confirmation')).toBeVisible();
        await expect(page.locator('.payout-status')).toContainText('Pending');
      }
    });
  });
  
  test.describe('Marketplace Integration', () => {
    test('should preview API in marketplace', async () => {
      await page.goto('/console/pages/apis.html');
      
      // Click preview button
      await page.click(`[data-api-name="${TEST_USER.apiName}"] [data-action="preview-marketplace"]`);
      
      // Switch to marketplace preview tab
      const [marketplacePage] = await Promise.all([
        page.waitForEvent('popup'),
        page.click('[data-action="open-preview"]')
      ]);
      
      // Verify API details in marketplace view
      await expect(marketplacePage.locator('h1')).toContainText(TEST_USER.apiName);
      await expect(marketplacePage.locator('.api-description')).toContainText(TEST_USER.apiDescription);
      
      await marketplacePage.close();
    });
    
    test('should manage marketplace listing', async () => {
      await page.goto('/console/pages/marketplace.html');
      
      // Find API listing
      const apiListing = page.locator(`[data-api="${TEST_USER.apiName}"]`);
      await expect(apiListing).toBeVisible();
      
      // Toggle visibility
      await page.click(`[data-api="${TEST_USER.apiName}"] [data-action="toggle-visibility"]`);
      await expect(page.locator('.success-message')).toContainText('Visibility updated');
      
      // Update tags
      await page.click(`[data-api="${TEST_USER.apiName}"] [data-action="edit-tags"]`);
      await page.fill('#api-tags', 'weather, forecast, climate');
      await page.click('[data-action="save-tags"]');
      
      // Verify tags updated
      await expect(page.locator(`[data-api="${TEST_USER.apiName}"] .api-tags`)).toContainText('weather');
    });
  });
  
  test.describe('Creator Portal', () => {
    test('should access creator resources', async () => {
      await page.goto('/console/creator-portal.html');
      
      // Verify sections
      await expect(page.locator('#documentation-section')).toBeVisible();
      await expect(page.locator('#sdk-downloads')).toBeVisible();
      await expect(page.locator('#api-templates')).toBeVisible();
      await expect(page.locator('#support-resources')).toBeVisible();
      
      // Test documentation links
      const docLinks = page.locator('.documentation-link');
      const linkCount = await docLinks.count();
      expect(linkCount).toBeGreaterThan(5);
      
      // Test SDK download
      const [download] = await Promise.all([
        page.waitForEvent('download'),
        page.click('[data-sdk="javascript"]')
      ]);
      
      expect(download.suggestedFilename()).toContain('sdk');
    });
  });
});

// Helper function to login
async function login(page: Page) {
  await page.goto('/console/login.html');
  await page.fill('#email', TEST_USER.email);
  await page.fill('#password', TEST_USER.password);
  await page.click('[type="submit"]');
  await page.waitForURL(/dashboard/);
}