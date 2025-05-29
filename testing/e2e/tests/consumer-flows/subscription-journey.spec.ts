import { test, expect, Page } from '@playwright/test';

test.describe('Consumer Subscription Journey', () => {
  let page: Page;
  const testUser = {
    email: 'test.consumer@example.com',
    password: 'TestPassword123!',
    name: 'Test Consumer'
  };

  test.beforeEach(async ({ page: p }) => {
    page = p;
    await page.goto('/');
  });

  test.describe('User Registration', () => {
    test('should allow new user registration', async () => {
      // Navigate to signup
      await page.click('[data-testid="signup-button"]');
      
      // Fill registration form
      await page.fill('[data-testid="name-input"]', testUser.name);
      await page.fill('[data-testid="email-input"]', testUser.email);
      await page.fill('[data-testid="password-input"]', testUser.password);
      await page.fill('[data-testid="confirm-password-input"]', testUser.password);
      
      // Accept terms
      await page.check('[data-testid="terms-checkbox"]');
      
      // Submit form
      await page.click('[data-testid="submit-signup"]');
      
      // Verify email verification page
      await expect(page.locator('text=/verify.*email/i')).toBeVisible();
    });

    test('should validate registration form', async () => {
      await page.click('[data-testid="signup-button"]');
      
      // Test email validation
      await page.fill('[data-testid="email-input"]', 'invalid-email');
      await page.click('[data-testid="submit-signup"]');
      await expect(page.locator('text=/valid.*email/i')).toBeVisible();
      
      // Test password validation
      await page.fill('[data-testid="email-input"]', testUser.email);
      await page.fill('[data-testid="password-input"]', 'weak');
      await page.fill('[data-testid="confirm-password-input"]', 'weak');
      await page.click('[data-testid="submit-signup"]');
      await expect(page.locator('text=/password.*requirements/i')).toBeVisible();
      
      // Test password mismatch
      await page.fill('[data-testid="password-input"]', testUser.password);
      await page.fill('[data-testid="confirm-password-input"]', 'DifferentPassword123!');
      await page.click('[data-testid="submit-signup"]');
      await expect(page.locator('text=/passwords.*match/i')).toBeVisible();
    });
  });

  test.describe('API Discovery & Subscription', () => {
    test.beforeEach(async () => {
      // Login as test user
      await page.click('[data-testid="login-button"]');
      await page.fill('[data-testid="email-input"]', testUser.email);
      await page.fill('[data-testid="password-input"]', testUser.password);
      await page.click('[data-testid="submit-login"]');
      await page.waitForNavigation();
    });

    test('should browse and filter APIs', async () => {
      // Browse marketplace
      await page.click('[data-testid="browse-apis"]');
      
      // Apply filters
      await page.selectOption('[data-testid="category-filter"]', 'Financial Services');
      await page.click('[data-testid="price-filter-medium"]');
      await page.check('[data-testid="free-tier-filter"]');
      
      // Verify filtered results
      await page.waitForSelector('[data-testid="api-card"]');
      const apiCount = await page.locator('[data-testid="api-card"]').count();
      expect(apiCount).toBeGreaterThan(0);
    });

    test('should view API details', async () => {
      // Click on first API card
      await page.click('[data-testid="api-card"]:first-of-type');
      
      // Verify details page elements
      await expect(page.locator('[data-testid="api-name"]')).toBeVisible();
      await expect(page.locator('[data-testid="api-description"]')).toBeVisible();
      await expect(page.locator('[data-testid="api-documentation"]')).toBeVisible();
      await expect(page.locator('[data-testid="pricing-plans"]')).toBeVisible();
      await expect(page.locator('[data-testid="api-reviews"]')).toBeVisible();
    });

    test('should subscribe to an API', async () => {
      // Navigate to API details
      await page.click('[data-testid="api-card"]:first-of-type');
      
      // Select a pricing plan
      await page.click('[data-testid="select-plan-basic"]');
      
      // Click subscribe
      await page.click('[data-testid="subscribe-button"]');
      
      // Fill payment details (Stripe test card)
      await page.waitForSelector('iframe[title="Secure payment input frame"]');
      const stripeFrame = page.frameLocator('iframe[title="Secure payment input frame"]');
      await stripeFrame.locator('[placeholder="Card number"]').fill('4242424242424242');
      await stripeFrame.locator('[placeholder="MM / YY"]').fill('12/30');
      await stripeFrame.locator('[placeholder="CVC"]').fill('123');
      await stripeFrame.locator('[placeholder="ZIP"]').fill('10001');
      
      // Complete subscription
      await page.click('[data-testid="complete-subscription"]');
      
      // Verify subscription success
      await expect(page.locator('text=/subscription.*successful/i')).toBeVisible();
      await expect(page.locator('[data-testid="api-key-display"]')).toBeVisible();
    });

    test('should handle failed payment', async () => {
      // Navigate to API details
      await page.click('[data-testid="api-card"]:first-of-type');
      await page.click('[data-testid="select-plan-premium"]');
      await page.click('[data-testid="subscribe-button"]');
      
      // Use declined test card
      await page.waitForSelector('iframe[title="Secure payment input frame"]');
      const stripeFrame = page.frameLocator('iframe[title="Secure payment input frame"]');
      await stripeFrame.locator('[placeholder="Card number"]').fill('4000000000000002');
      await stripeFrame.locator('[placeholder="MM / YY"]').fill('12/30');
      await stripeFrame.locator('[placeholder="CVC"]').fill('123');
      await stripeFrame.locator('[placeholder="ZIP"]').fill('10001');
      
      // Attempt subscription
      await page.click('[data-testid="complete-subscription"]');
      
      // Verify error handling
      await expect(page.locator('text=/payment.*declined/i')).toBeVisible();
    });
  });

  test.describe('Dashboard & API Usage', () => {
    test.beforeEach(async () => {
      // Login as subscribed user
      await page.click('[data-testid="login-button"]');
      await page.fill('[data-testid="email-input"]', testUser.email);
      await page.fill('[data-testid="password-input"]', testUser.password);
      await page.click('[data-testid="submit-login"]');
      await page.goto('/dashboard');
    });

    test('should display subscribed APIs', async () => {
      // Verify dashboard sections
      await expect(page.locator('[data-testid="subscribed-apis"]')).toBeVisible();
      await expect(page.locator('[data-testid="usage-stats"]')).toBeVisible();
      await expect(page.locator('[data-testid="billing-info"]')).toBeVisible();
      
      // Check subscribed API count
      const subscribedApis = await page.locator('[data-testid="subscribed-api-card"]').count();
      expect(subscribedApis).toBeGreaterThan(0);
    });

    test('should display API keys', async () => {
      // Click on subscribed API
      await page.click('[data-testid="subscribed-api-card"]:first-of-type');
      
      // Verify API key section
      await expect(page.locator('[data-testid="api-key-section"]')).toBeVisible();
      
      // Test show/hide API key
      await page.click('[data-testid="show-api-key"]');
      const apiKey = await page.locator('[data-testid="api-key-value"]').textContent();
      expect(apiKey).toMatch(/^api_[a-zA-Z0-9]+$/);
      
      // Test copy API key
      await page.click('[data-testid="copy-api-key"]');
      await expect(page.locator('text=/copied/i')).toBeVisible();
    });

    test('should display usage statistics', async () => {
      // Navigate to usage tab
      await page.click('[data-testid="usage-tab"]');
      
      // Verify usage metrics
      await expect(page.locator('[data-testid="api-calls-today"]')).toBeVisible();
      await expect(page.locator('[data-testid="api-calls-month"]')).toBeVisible();
      await expect(page.locator('[data-testid="usage-chart"]')).toBeVisible();
      
      // Check usage limits
      const usagePercentage = await page.locator('[data-testid="usage-percentage"]').textContent();
      expect(parseInt(usagePercentage || '0')).toBeGreaterThanOrEqual(0);
    });

    test('should manage subscription', async () => {
      // Navigate to billing tab
      await page.click('[data-testid="billing-tab"]');
      
      // Verify billing information
      await expect(page.locator('[data-testid="current-plan"]')).toBeVisible();
      await expect(page.locator('[data-testid="next-billing-date"]')).toBeVisible();
      await expect(page.locator('[data-testid="payment-method"]')).toBeVisible();
      
      // Test upgrade plan
      await page.click('[data-testid="upgrade-plan"]');
      await expect(page.locator('[data-testid="plan-comparison"]')).toBeVisible();
      
      // Test cancel subscription
      await page.click('[data-testid="cancel-subscription"]');
      await expect(page.locator('text=/cancel.*subscription/i')).toBeVisible();
      await page.click('[data-testid="confirm-cancel"]');
      await expect(page.locator('text=/subscription.*cancelled/i')).toBeVisible();
    });
  });

  test.describe('API Testing & Documentation', () => {
    test.beforeEach(async () => {
      // Login and navigate to subscribed API
      await page.click('[data-testid="login-button"]');
      await page.fill('[data-testid="email-input"]', testUser.email);
      await page.fill('[data-testid="password-input"]', testUser.password);
      await page.click('[data-testid="submit-login"]');
      await page.goto('/dashboard');
      await page.click('[data-testid="subscribed-api-card"]:first-of-type');
    });

    test('should test API with Swagger UI', async () => {
      // Navigate to API documentation
      await page.click('[data-testid="api-documentation-tab"]');
      
      // Wait for Swagger UI to load
      await page.waitForSelector('.swagger-ui');
      
      // Expand first endpoint
      await page.click('.opblock-summary:first-of-type');
      
      // Click "Try it out"
      await page.click('button:has-text("Try it out")');
      
      // Fill in parameters if any
      const paramInput = page.locator('.parameters input:first-of-type');
      if (await paramInput.isVisible()) {
        await paramInput.fill('test-value');
      }
      
      // Execute request
      await page.click('button:has-text("Execute")');
      
      // Verify response
      await expect(page.locator('.responses-inner')).toBeVisible();
      const responseCode = await page.locator('.response-code').textContent();
      expect(['200', '201', '204']).toContain(responseCode || '');
    });

    test('should download SDK', async () => {
      // Navigate to SDK section
      await page.click('[data-testid="sdk-tab"]');
      
      // Test SDK download
      const [download] = await Promise.all([
        page.waitForEvent('download'),
        page.click('[data-testid="download-sdk-javascript"]')
      ]);
      
      // Verify download
      expect(download.suggestedFilename()).toContain('.js');
    });

    test('should view code examples', async () => {
      // Navigate to code examples
      await page.click('[data-testid="code-examples-tab"]');
      
      // Verify language tabs
      await expect(page.locator('[data-testid="code-example-javascript"]')).toBeVisible();
      await expect(page.locator('[data-testid="code-example-python"]')).toBeVisible();
      await expect(page.locator('[data-testid="code-example-curl"]')).toBeVisible();
      
      // Test code copy
      await page.click('[data-testid="copy-code-javascript"]');
      await expect(page.locator('text=/copied/i')).toBeVisible();
    });
  });
});
