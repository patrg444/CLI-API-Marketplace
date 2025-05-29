# Test info

- Name: Creator Earnings & Payout Journey >> API Publishing & Pricing >> should set API pricing plans
- Location: /Users/patrickgloria/Desktop/CLI-API-Marketplace/testing/e2e/tests/creator-flows/earnings-payout.spec.ts:46:9

# Error details

```
TimeoutError: page.fill: Timeout 15000ms exceeded.
Call log:
  - waiting for locator('[data-testid="email-input"]')
    - waiting for navigation to finish...
    - navigated to "http://localhost:3001/creator-portal"

    at /Users/patrickgloria/Desktop/CLI-API-Marketplace/testing/e2e/tests/creator-flows/earnings-payout.spec.ts:21:18
```

# Page snapshot

```yaml
- heading "Unable to connect" [level=1]
- paragraph: Firefox can’t establish a connection to the server at localhost:3001.
- paragraph
- list:
  - listitem: The site could be temporarily unavailable or too busy. Try again in a few moments.
  - listitem: If you are unable to load any pages, check your computer’s network connection.
  - listitem: If your computer or network is protected by a firewall or proxy, make sure that Nightly is permitted to access the web.
  - listitem: If you are trying to load a local network page, please check that Nightly has been granted Local Network permissions in the macOS Privacy & Security settings.
- button "Try Again"
```

# Test source

```ts
   1 | import { test, expect, Page } from '@playwright/test';
   2 |
   3 | test.describe('Creator Earnings & Payout Journey', () => {
   4 |   let page: Page;
   5 |   const testCreator = {
   6 |     email: 'test.creator@example.com',
   7 |     password: 'CreatorPass123!',
   8 |     name: 'Test Creator',
   9 |     apiName: 'Test Payment API',
   10 |     apiDescription: 'A test API for payment processing'
   11 |   };
   12 |
   13 |   test.beforeEach(async ({ page: p }) => {
   14 |     page = p;
   15 |     await page.goto('/creator-portal');
   16 |   });
   17 |
   18 |   test.describe('API Publishing & Pricing', () => {
   19 |     test.beforeEach(async () => {
   20 |       // Login as creator
>  21 |       await page.fill('[data-testid="email-input"]', testCreator.email);
      |                  ^ TimeoutError: page.fill: Timeout 15000ms exceeded.
   22 |       await page.fill('[data-testid="password-input"]', testCreator.password);
   23 |       await page.click('[data-testid="submit-login"]');
   24 |       await page.waitForNavigation();
   25 |     });
   26 |
   27 |     test('should create and publish new API', async () => {
   28 |       // Navigate to APIs page
   29 |       await page.click('[data-testid="apis-nav"]');
   30 |       
   31 |       // Create new API
   32 |       await page.click('[data-testid="create-api-button"]');
   33 |       await page.fill('[data-testid="api-name-input"]', testCreator.apiName);
   34 |       await page.fill('[data-testid="api-description-input"]', testCreator.apiDescription);
   35 |       await page.selectOption('[data-testid="api-category-select"]', 'Financial Services');
   36 |       
   37 |       // Upload OpenAPI spec
   38 |       const fileInput = page.locator('[data-testid="openapi-upload"]');
   39 |       await fileInput.setInputFiles('./test-fixtures/sample-openapi.json');
   40 |       
   41 |       // Submit
   42 |       await page.click('[data-testid="create-api-submit"]');
   43 |       await expect(page.locator('text=/api.*created/i')).toBeVisible();
   44 |     });
   45 |
   46 |     test('should set API pricing plans', async () => {
   47 |       // Navigate to API settings
   48 |       await page.click('[data-testid="apis-nav"]');
   49 |       await page.click(`[data-testid="api-card-${testCreator.apiName}"]`);
   50 |       await page.click('[data-testid="marketplace-settings-tab"]');
   51 |       
   52 |       // Add pricing plan
   53 |       await page.click('[data-testid="add-pricing-plan"]');
   54 |       
   55 |       // Fill plan details
   56 |       await page.fill('[data-testid="plan-name-input"]', 'Basic Plan');
   57 |       await page.fill('[data-testid="plan-description-input"]', 'Basic access with rate limits');
   58 |       await page.fill('[data-testid="price-per-call-input"]', '0.01');
   59 |       await page.fill('[data-testid="monthly-price-input"]', '29.99');
   60 |       await page.fill('[data-testid="rate-limit-input"]', '1000');
   61 |       
   62 |       // Test negative price validation (bug fix verification)
   63 |       await page.fill('[data-testid="price-per-call-input"]', '-5');
   64 |       await page.click('[data-testid="save-plan-button"]');
   65 |       await expect(page.locator('[data-testid="price-per-call-input"]')).toHaveValue('0.01');
   66 |       
   67 |       // Save plan
   68 |       await page.click('[data-testid="save-plan-button"]');
   69 |       await expect(page.locator('text=/plan.*saved/i')).toBeVisible();
   70 |       
   71 |       // Add free tier
   72 |       await page.click('[data-testid="add-pricing-plan"]');
   73 |       await page.fill('[data-testid="plan-name-input"]', 'Free Tier');
   74 |       await page.fill('[data-testid="monthly-price-input"]', '0');
   75 |       await page.fill('[data-testid="rate-limit-input"]', '100');
   76 |       await page.click('[data-testid="save-plan-button"]');
   77 |     });
   78 |
   79 |     test('should publish API to marketplace', async () => {
   80 |       // Navigate to API
   81 |       await page.click('[data-testid="apis-nav"]');
   82 |       await page.click(`[data-testid="api-card-${testCreator.apiName}"]`);
   83 |       
   84 |       // Verify all requirements are met
   85 |       await expect(page.locator('[data-testid="requirement-openapi"] .checkmark')).toBeVisible();
   86 |       await expect(page.locator('[data-testid="requirement-pricing"] .checkmark')).toBeVisible();
   87 |       await expect(page.locator('[data-testid="requirement-documentation"] .checkmark')).toBeVisible();
   88 |       
   89 |       // Publish API
   90 |       await page.click('[data-testid="publish-api-button"]');
   91 |       await page.click('[data-testid="confirm-publish"]');
   92 |       
   93 |       // Verify published status
   94 |       await expect(page.locator('text=/published.*marketplace/i')).toBeVisible();
   95 |       await expect(page.locator('[data-testid="api-status"]')).toHaveText('Published');
   96 |     });
   97 |   });
   98 |
   99 |   test.describe('Stripe Connect Onboarding', () => {
  100 |     test.beforeEach(async () => {
  101 |       // Login as creator
  102 |       await page.fill('[data-testid="email-input"]', testCreator.email);
  103 |       await page.fill('[data-testid="password-input"]', testCreator.password);
  104 |       await page.click('[data-testid="submit-login"]');
  105 |       await page.goto('/creator-portal/payouts');
  106 |     });
  107 |
  108 |     test('should complete Stripe Connect onboarding', async () => {
  109 |       // Check if already connected
  110 |       const isConnected = await page.locator('[data-testid="stripe-connected"]').isVisible();
  111 |       
  112 |       if (!isConnected) {
  113 |         // Start onboarding
  114 |         await page.click('[data-testid="connect-stripe-button"]');
  115 |         
  116 |         // Fill business details (in Stripe iframe)
  117 |         await page.waitForSelector('iframe[name="stripe-connect-onboarding"]');
  118 |         const stripeFrame = page.frameLocator('iframe[name="stripe-connect-onboarding"]');
  119 |         
  120 |         // Business type
  121 |         await stripeFrame.locator('[data-testid="business-type-individual"]').click();
```