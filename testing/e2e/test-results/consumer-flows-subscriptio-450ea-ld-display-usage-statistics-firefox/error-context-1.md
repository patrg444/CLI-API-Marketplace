# Test info

- Name: Consumer Subscription Journey >> Dashboard & API Usage >> should display usage statistics
- Location: /Users/patrickgloria/Desktop/CLI-API-Marketplace/testing/e2e/tests/consumer-flows/subscription-journey.spec.ts:183:9

# Error details

```
TimeoutError: page.click: Timeout 15000ms exceeded.
Call log:
  - waiting for locator('[data-testid="login-button"]')
    - waiting for navigation to finish...
    - navigated to "http://localhost:3001/"

    at /Users/patrickgloria/Desktop/CLI-API-Marketplace/testing/e2e/tests/consumer-flows/subscription-journey.spec.ts:148:18
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
  114 |       
  115 |       // Complete subscription
  116 |       await page.click('[data-testid="complete-subscription"]');
  117 |       
  118 |       // Verify subscription success
  119 |       await expect(page.locator('text=/subscription.*successful/i')).toBeVisible();
  120 |       await expect(page.locator('[data-testid="api-key-display"]')).toBeVisible();
  121 |     });
  122 |
  123 |     test('should handle failed payment', async () => {
  124 |       // Navigate to API details
  125 |       await page.click('[data-testid="api-card"]:first-of-type');
  126 |       await page.click('[data-testid="select-plan-premium"]');
  127 |       await page.click('[data-testid="subscribe-button"]');
  128 |       
  129 |       // Use declined test card
  130 |       await page.waitForSelector('iframe[title="Secure payment input frame"]');
  131 |       const stripeFrame = page.frameLocator('iframe[title="Secure payment input frame"]');
  132 |       await stripeFrame.locator('[placeholder="Card number"]').fill('4000000000000002');
  133 |       await stripeFrame.locator('[placeholder="MM / YY"]').fill('12/30');
  134 |       await stripeFrame.locator('[placeholder="CVC"]').fill('123');
  135 |       await stripeFrame.locator('[placeholder="ZIP"]').fill('10001');
  136 |       
  137 |       // Attempt subscription
  138 |       await page.click('[data-testid="complete-subscription"]');
  139 |       
  140 |       // Verify error handling
  141 |       await expect(page.locator('text=/payment.*declined/i')).toBeVisible();
  142 |     });
  143 |   });
  144 |
  145 |   test.describe('Dashboard & API Usage', () => {
  146 |     test.beforeEach(async () => {
  147 |       // Login as subscribed user
> 148 |       await page.click('[data-testid="login-button"]');
      |                  ^ TimeoutError: page.click: Timeout 15000ms exceeded.
  149 |       await page.fill('[data-testid="email-input"]', testUser.email);
  150 |       await page.fill('[data-testid="password-input"]', testUser.password);
  151 |       await page.click('[data-testid="submit-login"]');
  152 |       await page.goto('/dashboard');
  153 |     });
  154 |
  155 |     test('should display subscribed APIs', async () => {
  156 |       // Verify dashboard sections
  157 |       await expect(page.locator('[data-testid="subscribed-apis"]')).toBeVisible();
  158 |       await expect(page.locator('[data-testid="usage-stats"]')).toBeVisible();
  159 |       await expect(page.locator('[data-testid="billing-info"]')).toBeVisible();
  160 |       
  161 |       // Check subscribed API count
  162 |       const subscribedApis = await page.locator('[data-testid="subscribed-api-card"]').count();
  163 |       expect(subscribedApis).toBeGreaterThan(0);
  164 |     });
  165 |
  166 |     test('should display API keys', async () => {
  167 |       // Click on subscribed API
  168 |       await page.click('[data-testid="subscribed-api-card"]:first-of-type');
  169 |       
  170 |       // Verify API key section
  171 |       await expect(page.locator('[data-testid="api-key-section"]')).toBeVisible();
  172 |       
  173 |       // Test show/hide API key
  174 |       await page.click('[data-testid="show-api-key"]');
  175 |       const apiKey = await page.locator('[data-testid="api-key-value"]').textContent();
  176 |       expect(apiKey).toMatch(/^api_[a-zA-Z0-9]+$/);
  177 |       
  178 |       // Test copy API key
  179 |       await page.click('[data-testid="copy-api-key"]');
  180 |       await expect(page.locator('text=/copied/i')).toBeVisible();
  181 |     });
  182 |
  183 |     test('should display usage statistics', async () => {
  184 |       // Navigate to usage tab
  185 |       await page.click('[data-testid="usage-tab"]');
  186 |       
  187 |       // Verify usage metrics
  188 |       await expect(page.locator('[data-testid="api-calls-today"]')).toBeVisible();
  189 |       await expect(page.locator('[data-testid="api-calls-month"]')).toBeVisible();
  190 |       await expect(page.locator('[data-testid="usage-chart"]')).toBeVisible();
  191 |       
  192 |       // Check usage limits
  193 |       const usagePercentage = await page.locator('[data-testid="usage-percentage"]').textContent();
  194 |       expect(parseInt(usagePercentage || '0')).toBeGreaterThanOrEqual(0);
  195 |     });
  196 |
  197 |     test('should manage subscription', async () => {
  198 |       // Navigate to billing tab
  199 |       await page.click('[data-testid="billing-tab"]');
  200 |       
  201 |       // Verify billing information
  202 |       await expect(page.locator('[data-testid="current-plan"]')).toBeVisible();
  203 |       await expect(page.locator('[data-testid="next-billing-date"]')).toBeVisible();
  204 |       await expect(page.locator('[data-testid="payment-method"]')).toBeVisible();
  205 |       
  206 |       // Test upgrade plan
  207 |       await page.click('[data-testid="upgrade-plan"]');
  208 |       await expect(page.locator('[data-testid="plan-comparison"]')).toBeVisible();
  209 |       
  210 |       // Test cancel subscription
  211 |       await page.click('[data-testid="cancel-subscription"]');
  212 |       await expect(page.locator('text=/cancel.*subscription/i')).toBeVisible();
  213 |       await page.click('[data-testid="confirm-cancel"]');
  214 |       await expect(page.locator('text=/subscription.*cancelled/i')).toBeVisible();
  215 |     });
  216 |   });
  217 |
  218 |   test.describe('API Testing & Documentation', () => {
  219 |     test.beforeEach(async () => {
  220 |       // Login and navigate to subscribed API
  221 |       await page.click('[data-testid="login-button"]');
  222 |       await page.fill('[data-testid="email-input"]', testUser.email);
  223 |       await page.fill('[data-testid="password-input"]', testUser.password);
  224 |       await page.click('[data-testid="submit-login"]');
  225 |       await page.goto('/dashboard');
  226 |       await page.click('[data-testid="subscribed-api-card"]:first-of-type');
  227 |     });
  228 |
  229 |     test('should test API with Swagger UI', async () => {
  230 |       // Navigate to API documentation
  231 |       await page.click('[data-testid="api-documentation-tab"]');
  232 |       
  233 |       // Wait for Swagger UI to load
  234 |       await page.waitForSelector('.swagger-ui');
  235 |       
  236 |       // Expand first endpoint
  237 |       await page.click('.opblock-summary:first-of-type');
  238 |       
  239 |       // Click "Try it out"
  240 |       await page.click('button:has-text("Try it out")');
  241 |       
  242 |       // Fill in parameters if any
  243 |       const paramInput = page.locator('.parameters input:first-of-type');
  244 |       if (await paramInput.isVisible()) {
  245 |         await paramInput.fill('test-value');
  246 |       }
  247 |       
  248 |       // Execute request
```