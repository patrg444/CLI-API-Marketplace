const { chromium } = require('playwright');

(async () => {
  console.log('🚀 Testing console.apidirect.dev with minimal script...\n');
  
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();
  
  try {
    // Test 1: Basic accessibility
    console.log('Test 1: Checking if console is accessible...');
    const response = await page.goto('https://console.apidirect.dev', {
      waitUntil: 'domcontentloaded',
      timeout: 30000
    });
    
    console.log(`✓ Status: ${response.status()}`);
    console.log(`✓ URL: ${page.url()}`);
    console.log(`✓ Title: ${await page.title()}`);
    
    // Test 2: Page structure
    console.log('\nTest 2: Analyzing page structure...');
    const pageText = await page.textContent('body');
    
    const features = {
      'Navigation': await page.locator('nav, .sidebar').count() > 0,
      'Login Form': await page.locator('input[type="email"]').count() > 0,
      'Dashboard Links': pageText.includes('Dashboard') || pageText.includes('dashboard'),
      'API Links': pageText.includes('API') || pageText.includes('apis'),
      'Analytics Links': pageText.includes('Analytics') || pageText.includes('analytics')
    };
    
    for (const [feature, found] of Object.entries(features)) {
      console.log(`${found ? '✓' : '✗'} ${feature}`);
    }
    
    // Test 3: Performance
    console.log('\nTest 3: Performance metrics...');
    const metrics = await page.evaluate(() => {
      const timing = performance.timing;
      return {
        domContentLoaded: timing.domContentLoadedEventEnd - timing.navigationStart,
        loadComplete: timing.loadEventEnd - timing.navigationStart
      };
    });
    console.log(`✓ DOM Content Loaded: ${metrics.domContentLoaded}ms`);
    console.log(`✓ Page Load Complete: ${metrics.loadComplete}ms`);
    
    // Test 4: Screenshot
    console.log('\nTest 4: Taking screenshots...');
    await page.screenshot({ 
      path: 'console-test-results.png',
      fullPage: true 
    });
    console.log('✓ Screenshot saved: console-test-results.png');
    
    // Test 5: API endpoint check
    console.log('\nTest 5: Checking for API integration...');
    const scripts = await page.locator('script').all();
    let hasApiClient = false;
    for (const script of scripts) {
      const src = await script.getAttribute('src');
      if (src && src.includes('api')) {
        hasApiClient = true;
        console.log(`✓ Found API script: ${src}`);
      }
    }
    if (!hasApiClient) {
      console.log('✗ No API client scripts found');
    }
    
    console.log('\n✅ All tests completed successfully!');
    
  } catch (error) {
    console.error('\n❌ Test failed:', error.message);
  } finally {
    await browser.close();
  }
})();