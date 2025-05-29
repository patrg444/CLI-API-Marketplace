import { test, expect, Page } from '@playwright/test';

test.describe('Creator Earnings & Payout Journey', () => {
  let page: Page;
  const testCreator = {
    email: 'test.creator@example.com',
    password: 'CreatorPass123!',
    name: 'Test Creator',
    apiName: 'Test Payment API',
    apiDescription: 'A test API for payment processing'
  };

  test.beforeEach(async ({ page: p }) => {
    page = p;
    await page.goto('/creator-portal');
  });

  test.describe('API Publishing & Pricing', () => {
    test.beforeEach(async () => {
      // Login as creator
      await page.fill('[data-testid="email-input"]', testCreator.email);
      await page.fill('[data-testid="password-input"]', testCreator.password);
      await page.click('[data-testid="submit-login"]');
      await page.waitForNavigation();
    });

    test('should create and publish new API', async () => {
      // Navigate to APIs page
      await page.click('[data-testid="apis-nav"]');
      
      // Create new API
      await page.click('[data-testid="create-api-button"]');
      await page.fill('[data-testid="api-name-input"]', testCreator.apiName);
      await page.fill('[data-testid="api-description-input"]', testCreator.apiDescription);
      await page.selectOption('[data-testid="api-category-select"]', 'Financial Services');
      
      // Upload OpenAPI spec
      const fileInput = page.locator('[data-testid="openapi-upload"]');
      await fileInput.setInputFiles('./test-fixtures/sample-openapi.json');
      
      // Submit
      await page.click('[data-testid="create-api-submit"]');
      await expect(page.locator('text=/api.*created/i')).toBeVisible();
    });

    test('should set API pricing plans', async () => {
      // Navigate to API settings
      await page.click('[data-testid="apis-nav"]');
      await page.click(`[data-testid="api-card-${testCreator.apiName}"]`);
      await page.click('[data-testid="marketplace-settings-tab"]');
      
      // Add pricing plan
      await page.click('[data-testid="add-pricing-plan"]');
      
      // Fill plan details
      await page.fill('[data-testid="plan-name-input"]', 'Basic Plan');
      await page.fill('[data-testid="plan-description-input"]', 'Basic access with rate limits');
      await page.fill('[data-testid="price-per-call-input"]', '0.01');
      await page.fill('[data-testid="monthly-price-input"]', '29.99');
      await page.fill('[data-testid="rate-limit-input"]', '1000');
      
      // Test negative price validation (bug fix verification)
      await page.fill('[data-testid="price-per-call-input"]', '-5');
      await page.click('[data-testid="save-plan-button"]');
      await expect(page.locator('[data-testid="price-per-call-input"]')).toHaveValue('0.01');
      
      // Save plan
      await page.click('[data-testid="save-plan-button"]');
      await expect(page.locator('text=/plan.*saved/i')).toBeVisible();
      
      // Add free tier
      await page.click('[data-testid="add-pricing-plan"]');
      await page.fill('[data-testid="plan-name-input"]', 'Free Tier');
      await page.fill('[data-testid="monthly-price-input"]', '0');
      await page.fill('[data-testid="rate-limit-input"]', '100');
      await page.click('[data-testid="save-plan-button"]');
    });

    test('should publish API to marketplace', async () => {
      // Navigate to API
      await page.click('[data-testid="apis-nav"]');
      await page.click(`[data-testid="api-card-${testCreator.apiName}"]`);
      
      // Verify all requirements are met
      await expect(page.locator('[data-testid="requirement-openapi"] .checkmark')).toBeVisible();
      await expect(page.locator('[data-testid="requirement-pricing"] .checkmark')).toBeVisible();
      await expect(page.locator('[data-testid="requirement-documentation"] .checkmark')).toBeVisible();
      
      // Publish API
      await page.click('[data-testid="publish-api-button"]');
      await page.click('[data-testid="confirm-publish"]');
      
      // Verify published status
      await expect(page.locator('text=/published.*marketplace/i')).toBeVisible();
      await expect(page.locator('[data-testid="api-status"]')).toHaveText('Published');
    });
  });

  test.describe('Stripe Connect Onboarding', () => {
    test.beforeEach(async () => {
      // Login as creator
      await page.fill('[data-testid="email-input"]', testCreator.email);
      await page.fill('[data-testid="password-input"]', testCreator.password);
      await page.click('[data-testid="submit-login"]');
      await page.goto('/creator-portal/payouts');
    });

    test('should complete Stripe Connect onboarding', async () => {
      // Check if already connected
      const isConnected = await page.locator('[data-testid="stripe-connected"]').isVisible();
      
      if (!isConnected) {
        // Start onboarding
        await page.click('[data-testid="connect-stripe-button"]');
        
        // Fill business details (in Stripe iframe)
        await page.waitForSelector('iframe[name="stripe-connect-onboarding"]');
        const stripeFrame = page.frameLocator('iframe[name="stripe-connect-onboarding"]');
        
        // Business type
        await stripeFrame.locator('[data-testid="business-type-individual"]').click();
        
        // Personal details
        await stripeFrame.locator('[name="first_name"]').fill('Test');
        await stripeFrame.locator('[name="last_name"]').fill('Creator');
        await stripeFrame.locator('[name="email"]').fill(testCreator.email);
        await stripeFrame.locator('[name="phone"]').fill('+12125551234');
        
        // Address
        await stripeFrame.locator('[name="address_line1"]').fill('123 Test St');
        await stripeFrame.locator('[name="city"]').fill('New York');
        await stripeFrame.locator('[name="state"]').selectOption('NY');
        await stripeFrame.locator('[name="zip"]').fill('10001');
        
        // SSN (test value)
        await stripeFrame.locator('[name="ssn_last_4"]').fill('0000');
        
        // Bank account
        await stripeFrame.locator('[name="routing_number"]').fill('110000000');
        await stripeFrame.locator('[name="account_number"]').fill('000123456789');
        
        // Submit
        await stripeFrame.locator('button[type="submit"]').click();
        
        // Return to platform
        await page.waitForNavigation();
        await expect(page.locator('[data-testid="stripe-connected"]')).toBeVisible();
      }
    });

    test('should display payout settings', async () => {
      // Verify payout schedule
      await expect(page.locator('[data-testid="payout-schedule"]')).toBeVisible();
      await expect(page.locator('[data-testid="payout-schedule"]')).toContainText('Monthly');
      
      // Verify minimum payout
      await expect(page.locator('[data-testid="minimum-payout"]')).toBeVisible();
      await expect(page.locator('[data-testid="minimum-payout"]')).toContainText('$100');
      
      // Verify bank account
      await expect(page.locator('[data-testid="bank-account"]')).toBeVisible();
      await expect(page.locator('[data-testid="bank-account"]')).toContainText('****6789');
    });
  });

  test.describe('Earnings Tracking', () => {
    test.beforeEach(async () => {
      // Login as creator with earnings
      await page.fill('[data-testid="email-input"]', testCreator.email);
      await page.fill('[data-testid="password-input"]', testCreator.password);
      await page.click('[data-testid="submit-login"]');
      await page.goto('/creator-portal/payouts');
    });

    test('should display earnings dashboard', async () => {
      // Verify earnings overview
      await expect(page.locator('[data-testid="total-earnings"]')).toBeVisible();
      await expect(page.locator('[data-testid="pending-earnings"]')).toBeVisible();
      await expect(page.locator('[data-testid="available-balance"]')).toBeVisible();
      
      // Verify earnings chart
      await expect(page.locator('[data-testid="earnings-chart"]')).toBeVisible();
      
      // Test date range filter
      await page.selectOption('[data-testid="date-range-select"]', 'last_30_days');
      await page.waitForLoadState('networkidle');
      
      // Verify updated data
      const totalEarnings = await page.locator('[data-testid="total-earnings"]').textContent();
      expect(totalEarnings).toMatch(/\$[\d,]+\.\d{2}/);
    });

    test('should display earnings by API', async () => {
      // Switch to API breakdown view
      await page.click('[data-testid="earnings-by-api-tab"]');
      
      // Verify API earnings table
      await expect(page.locator('[data-testid="api-earnings-table"]')).toBeVisible();
      
      // Check table has data
      const apiRows = await page.locator('[data-testid="api-earnings-row"]').count();
      expect(apiRows).toBeGreaterThan(0);
      
      // Verify earnings details
      const firstApiEarnings = await page.locator('[data-testid="api-earnings-row"]:first-of-type').textContent();
      expect(firstApiEarnings).toContain(testCreator.apiName);
      expect(firstApiEarnings).toMatch(/\$[\d,]+\.\d{2}/);
    });

    test('should display transaction history', async () => {
      // Navigate to transactions
      await page.click('[data-testid="transactions-tab"]');
      
      // Verify transaction list
      await expect(page.locator('[data-testid="transaction-list"]')).toBeVisible();
      
      // Check transaction details
      const transactions = await page.locator('[data-testid="transaction-item"]').count();
      if (transactions > 0) {
        const firstTransaction = page.locator('[data-testid="transaction-item"]:first-of-type');
        await expect(firstTransaction.locator('[data-testid="transaction-type"]')).toBeVisible();
        await expect(firstTransaction.locator('[data-testid="transaction-amount"]')).toBeVisible();
        await expect(firstTransaction.locator('[data-testid="transaction-date"]')).toBeVisible();
        await expect(firstTransaction.locator('[data-testid="transaction-status"]')).toBeVisible();
      }
      
      // Test transaction filtering
      await page.selectOption('[data-testid="transaction-type-filter"]', 'api_usage');
      await page.waitForLoadState('networkidle');
      
      // Export transactions
      await page.click('[data-testid="export-transactions"]');
      const [download] = await Promise.all([
        page.waitForEvent('download'),
        page.click('[data-testid="export-csv"]')
      ]);
      expect(download.suggestedFilename()).toContain('transactions');
    });
  });

  test.describe('Payout Management', () => {
    test.beforeEach(async () => {
      // Login as creator with available balance
      await page.fill('[data-testid="email-input"]', testCreator.email);
      await page.fill('[data-testid="password-input"]', testCreator.password);
      await page.click('[data-testid="submit-login"]');
      await page.goto('/creator-portal/payouts');
    });

    test('should display payout history', async () => {
      // Navigate to payout history
      await page.click('[data-testid="payout-history-tab"]');
      
      // Verify payout list
      await expect(page.locator('[data-testid="payout-list"]')).toBeVisible();
      
      // Check payout details
      const payouts = await page.locator('[data-testid="payout-item"]').count();
      if (payouts > 0) {
        const firstPayout = page.locator('[data-testid="payout-item"]:first-of-type');
        await expect(firstPayout.locator('[data-testid="payout-amount"]')).toBeVisible();
        await expect(firstPayout.locator('[data-testid="payout-date"]')).toBeVisible();
        await expect(firstPayout.locator('[data-testid="payout-status"]')).toBeVisible();
        
        // View payout details
        await firstPayout.click();
        await expect(page.locator('[data-testid="payout-details-modal"]')).toBeVisible();
        await expect(page.locator('[data-testid="payout-breakdown"]')).toBeVisible();
        await page.click('[data-testid="close-modal"]');
      }
    });

    test('should handle manual payout request', async () => {
      // Check if manual payout is available
      const availableBalance = await page.locator('[data-testid="available-balance"]').textContent();
      const balance = parseFloat(availableBalance?.replace(/[$,]/g, '') || '0');
      
      if (balance >= 100) {
        // Request manual payout
        await page.click('[data-testid="request-payout-button"]');
        
        // Confirm payout
        await expect(page.locator('[data-testid="payout-confirmation-modal"]')).toBeVisible();
        await expect(page.locator('[data-testid="payout-amount-confirm"]')).toContainText(availableBalance || '');
        
        // Submit request
        await page.click('[data-testid="confirm-payout-request"]');
        
        // Verify success
        await expect(page.locator('text=/payout.*requested/i')).toBeVisible();
        await expect(page.locator('[data-testid="available-balance"]')).toContainText('$0.00');
      }
    });

    test('should update payout preferences', async () => {
      // Navigate to settings
      await page.click('[data-testid="payout-settings-tab"]');
      
      // Update minimum payout threshold
      await page.selectOption('[data-testid="minimum-payout-select"]', '500');
      
      // Update payout schedule
      await page.selectOption('[data-testid="payout-schedule-select"]', 'weekly');
      
      // Save changes
      await page.click('[data-testid="save-payout-settings"]');
      
      // Verify success
      await expect(page.locator('text=/settings.*updated/i')).toBeVisible();
      
      // Verify changes persisted
      await page.reload();
      await expect(page.locator('[data-testid="minimum-payout-select"]')).toHaveValue('500');
      await expect(page.locator('[data-testid="payout-schedule-select"]')).toHaveValue('weekly');
    });
  });

  test.describe('Analytics & Insights', () => {
    test.beforeEach(async () => {
      // Login as creator
      await page.fill('[data-testid="email-input"]', testCreator.email);
      await page.fill('[data-testid="password-input"]', testCreator.password);
      await page.click('[data-testid="submit-login"]');
      await page.goto('/creator-portal/dashboard');
    });

    test('should display revenue analytics', async () => {
      // Verify analytics dashboard
      await expect(page.locator('[data-testid="revenue-chart"]')).toBeVisible();
      await expect(page.locator('[data-testid="api-usage-chart"]')).toBeVisible();
      await expect(page.locator('[data-testid="subscriber-growth-chart"]')).toBeVisible();
      
      // Test time range selector
      await page.selectOption('[data-testid="analytics-timerange"]', 'last_90_days');
      await page.waitForLoadState('networkidle');
      
      // Verify KPIs
      await expect(page.locator('[data-testid="total-revenue-kpi"]')).toBeVisible();
      await expect(page.locator('[data-testid="active-subscribers-kpi"]')).toBeVisible();
      await expect(page.locator('[data-testid="api-calls-kpi"]')).toBeVisible();
      await expect(page.locator('[data-testid="conversion-rate-kpi"]')).toBeVisible();
    });

    test('should export analytics data', async () => {
      // Export revenue report
      const [download] = await Promise.all([
        page.waitForEvent('download'),
        page.click('[data-testid="export-analytics"]')
      ]);
      
      // Verify download
      expect(download.suggestedFilename()).toMatch(/analytics.*\.csv$/);
    });
  });
});
