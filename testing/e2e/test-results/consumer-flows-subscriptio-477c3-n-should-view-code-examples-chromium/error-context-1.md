# Test info

- Name: Consumer Subscription Journey >> API Testing & Documentation >> should view code examples
- Location: /Users/patrickgloria/Desktop/CLI-API-Marketplace/testing/e2e/tests/consumer-flows/subscription-journey.spec.ts:271:9

# Error details

```
TimeoutError: page.click: Timeout 15000ms exceeded.
Call log:
  - waiting for locator('[data-testid="login-button"]')

    at /Users/patrickgloria/Desktop/CLI-API-Marketplace/testing/e2e/tests/consumer-flows/subscription-journey.spec.ts:221:18
```

# Test source

```ts
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
  148 |       await page.click('[data-testid="login-button"]');
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
> 221 |       await page.click('[data-testid="login-button"]');
      |                  ^ TimeoutError: page.click: Timeout 15000ms exceeded.
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
  249 |       await page.click('button:has-text("Execute")');
  250 |       
  251 |       // Verify response
  252 |       await expect(page.locator('.responses-inner')).toBeVisible();
  253 |       const responseCode = await page.locator('.response-code').textContent();
  254 |       expect(['200', '201', '204']).toContain(responseCode || '');
  255 |     });
  256 |
  257 |     test('should download SDK', async () => {
  258 |       // Navigate to SDK section
  259 |       await page.click('[data-testid="sdk-tab"]');
  260 |       
  261 |       // Test SDK download
  262 |       const [download] = await Promise.all([
  263 |         page.waitForEvent('download'),
  264 |         page.click('[data-testid="download-sdk-javascript"]')
  265 |       ]);
  266 |       
  267 |       // Verify download
  268 |       expect(download.suggestedFilename()).toContain('.js');
  269 |     });
  270 |
  271 |     test('should view code examples', async () => {
  272 |       // Navigate to code examples
  273 |       await page.click('[data-testid="code-examples-tab"]');
  274 |       
  275 |       // Verify language tabs
  276 |       await expect(page.locator('[data-testid="code-example-javascript"]')).toBeVisible();
  277 |       await expect(page.locator('[data-testid="code-example-python"]')).toBeVisible();
  278 |       await expect(page.locator('[data-testid="code-example-curl"]')).toBeVisible();
  279 |       
  280 |       // Test code copy
  281 |       await page.click('[data-testid="copy-code-javascript"]');
  282 |       await expect(page.locator('text=/copied/i')).toBeVisible();
  283 |     });
  284 |   });
  285 | });
  286 |
```