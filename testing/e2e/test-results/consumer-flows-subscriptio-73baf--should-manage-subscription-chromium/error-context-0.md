# Test info

- Name: Consumer Subscription Journey >> Dashboard & API Usage >> should manage subscription
- Location: /Users/patrickgloria/Desktop/CLI-API-Marketplace/testing/e2e/tests/consumer-flows/subscription-journey.spec.ts:197:9

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:3001/
Call log:
  - navigating to "http://localhost:3001/", waiting until "load"

    at /Users/patrickgloria/Desktop/CLI-API-Marketplace/testing/e2e/tests/consumer-flows/subscription-journey.spec.ts:13:16
```

# Test source

```ts
   1 | import { test, expect, Page } from '@playwright/test';
   2 |
   3 | test.describe('Consumer Subscription Journey', () => {
   4 |   let page: Page;
   5 |   const testUser = {
   6 |     email: 'test.consumer@example.com',
   7 |     password: 'TestPassword123!',
   8 |     name: 'Test Consumer'
   9 |   };
   10 |
   11 |   test.beforeEach(async ({ page: p }) => {
   12 |     page = p;
>  13 |     await page.goto('/');
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:3001/
   14 |   });
   15 |
   16 |   test.describe('User Registration', () => {
   17 |     test('should allow new user registration', async () => {
   18 |       // Navigate to signup
   19 |       await page.click('[data-testid="signup-button"]');
   20 |       
   21 |       // Fill registration form
   22 |       await page.fill('[data-testid="name-input"]', testUser.name);
   23 |       await page.fill('[data-testid="email-input"]', testUser.email);
   24 |       await page.fill('[data-testid="password-input"]', testUser.password);
   25 |       await page.fill('[data-testid="confirm-password-input"]', testUser.password);
   26 |       
   27 |       // Accept terms
   28 |       await page.check('[data-testid="terms-checkbox"]');
   29 |       
   30 |       // Submit form
   31 |       await page.click('[data-testid="submit-signup"]');
   32 |       
   33 |       // Verify email verification page
   34 |       await expect(page.locator('text=/verify.*email/i')).toBeVisible();
   35 |     });
   36 |
   37 |     test('should validate registration form', async () => {
   38 |       await page.click('[data-testid="signup-button"]');
   39 |       
   40 |       // Test email validation
   41 |       await page.fill('[data-testid="email-input"]', 'invalid-email');
   42 |       await page.click('[data-testid="submit-signup"]');
   43 |       await expect(page.locator('text=/valid.*email/i')).toBeVisible();
   44 |       
   45 |       // Test password validation
   46 |       await page.fill('[data-testid="email-input"]', testUser.email);
   47 |       await page.fill('[data-testid="password-input"]', 'weak');
   48 |       await page.fill('[data-testid="confirm-password-input"]', 'weak');
   49 |       await page.click('[data-testid="submit-signup"]');
   50 |       await expect(page.locator('text=/password.*requirements/i')).toBeVisible();
   51 |       
   52 |       // Test password mismatch
   53 |       await page.fill('[data-testid="password-input"]', testUser.password);
   54 |       await page.fill('[data-testid="confirm-password-input"]', 'DifferentPassword123!');
   55 |       await page.click('[data-testid="submit-signup"]');
   56 |       await expect(page.locator('text=/passwords.*match/i')).toBeVisible();
   57 |     });
   58 |   });
   59 |
   60 |   test.describe('API Discovery & Subscription', () => {
   61 |     test.beforeEach(async () => {
   62 |       // Login as test user
   63 |       await page.click('[data-testid="login-button"]');
   64 |       await page.fill('[data-testid="email-input"]', testUser.email);
   65 |       await page.fill('[data-testid="password-input"]', testUser.password);
   66 |       await page.click('[data-testid="submit-login"]');
   67 |       await page.waitForNavigation();
   68 |     });
   69 |
   70 |     test('should browse and filter APIs', async () => {
   71 |       // Browse marketplace
   72 |       await page.click('[data-testid="browse-apis"]');
   73 |       
   74 |       // Apply filters
   75 |       await page.selectOption('[data-testid="category-filter"]', 'Financial Services');
   76 |       await page.click('[data-testid="price-filter-medium"]');
   77 |       await page.check('[data-testid="free-tier-filter"]');
   78 |       
   79 |       // Verify filtered results
   80 |       await page.waitForSelector('[data-testid="api-card"]');
   81 |       const apiCount = await page.locator('[data-testid="api-card"]').count();
   82 |       expect(apiCount).toBeGreaterThan(0);
   83 |     });
   84 |
   85 |     test('should view API details', async () => {
   86 |       // Click on first API card
   87 |       await page.click('[data-testid="api-card"]:first-of-type');
   88 |       
   89 |       // Verify details page elements
   90 |       await expect(page.locator('[data-testid="api-name"]')).toBeVisible();
   91 |       await expect(page.locator('[data-testid="api-description"]')).toBeVisible();
   92 |       await expect(page.locator('[data-testid="api-documentation"]')).toBeVisible();
   93 |       await expect(page.locator('[data-testid="pricing-plans"]')).toBeVisible();
   94 |       await expect(page.locator('[data-testid="api-reviews"]')).toBeVisible();
   95 |     });
   96 |
   97 |     test('should subscribe to an API', async () => {
   98 |       // Navigate to API details
   99 |       await page.click('[data-testid="api-card"]:first-of-type');
  100 |       
  101 |       // Select a pricing plan
  102 |       await page.click('[data-testid="select-plan-basic"]');
  103 |       
  104 |       // Click subscribe
  105 |       await page.click('[data-testid="subscribe-button"]');
  106 |       
  107 |       // Fill payment details (Stripe test card)
  108 |       await page.waitForSelector('iframe[title="Secure payment input frame"]');
  109 |       const stripeFrame = page.frameLocator('iframe[title="Secure payment input frame"]');
  110 |       await stripeFrame.locator('[placeholder="Card number"]').fill('4242424242424242');
  111 |       await stripeFrame.locator('[placeholder="MM / YY"]').fill('12/30');
  112 |       await stripeFrame.locator('[placeholder="CVC"]').fill('123');
  113 |       await stripeFrame.locator('[placeholder="ZIP"]').fill('10001');
```