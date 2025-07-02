const { chromium } = require('playwright');

(async () => {
  console.log('🧪 Testing console.apidirect.dev...\n');
  
  const browser = await chromium.launch({ 
    headless: false,
    slowMo: 50 
  });
  
  const context = await browser.newContext();
  const page = await context.newPage();
  
  try {
    // Test 1: Load homepage
    console.log('✓ Test 1: Loading homepage...');
    await page.goto('https://console.apidirect.dev');
    console.log(`  URL: ${page.url()}`);
    console.log(`  Title: ${await page.title()}`);
    
    // Test 2: Check for key elements
    console.log('\n✓ Test 2: Checking page elements...');
    
    // Check if it's a login page or dashboard
    const hasLogin = await page.locator('input[type="email"], input[type="password"]').count() > 0;
    const hasDashboard = await page.locator('text=/dashboard/i').count() > 0;
    
    if (hasLogin) {
      console.log('  Found: Login page');
      console.log('  - Email input:', await page.locator('input[type="email"]').count() > 0);
      console.log('  - Password input:', await page.locator('input[type="password"]').count() > 0);
      console.log('  - Submit button:', await page.locator('button[type="submit"]').count() > 0);
    } else if (hasDashboard) {
      console.log('  Found: Dashboard page');
    } else {
      console.log('  Found: Other page type');
    }
    
    // Test 3: Take screenshot
    console.log('\n✓ Test 3: Taking screenshot...');
    await page.screenshot({ 
      path: 'console-homepage.png',
      fullPage: true 
    });
    console.log('  Screenshot saved: console-homepage.png');
    
    // Test 4: Check responsive design
    console.log('\n✓ Test 4: Testing responsive design...');
    await page.setViewportSize({ width: 375, height: 667 });
    await page.screenshot({ path: 'console-mobile.png' });
    console.log('  Mobile screenshot saved: console-mobile.png');
    
    // Test 5: Performance check
    console.log('\n✓ Test 5: Checking performance...');
    const startTime = Date.now();
    await page.reload();
    const loadTime = Date.now() - startTime;
    console.log(`  Page load time: ${loadTime}ms`);
    
    // Test 6: Security check
    console.log('\n✓ Test 6: Security checks...');
    console.log(`  HTTPS: ${page.url().startsWith('https')}`);
    
    // Test 7: Try navigation
    console.log('\n✓ Test 7: Testing navigation...');
    const links = await page.locator('a:visible').all();
    console.log(`  Found ${links.length} visible links`);
    
    if (links.length > 0) {
      const firstLink = links[0];
      const href = await firstLink.getAttribute('href');
      console.log(`  First link: ${href}`);
    }
    
    // Test 8: Check for API endpoints or features
    console.log('\n✓ Test 8: Looking for console features...');
    const features = [
      'API', 'Dashboard', 'Analytics', 'Deploy', 'Marketplace', 
      'Sign in', 'Login', 'Register', 'Console'
    ];
    
    for (const feature of features) {
      const found = await page.locator(`text=/${feature}/i`).count() > 0;
      if (found) {
        console.log(`  ✓ Found: ${feature}`);
      }
    }
    
    console.log('\n✅ All tests completed!');
    
  } catch (error) {
    console.error('\n❌ Test failed:', error.message);
  } finally {
    await browser.close();
  }
})();