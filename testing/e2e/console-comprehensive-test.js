const { chromium } = require('playwright');

(async () => {
  console.log('🧪 Comprehensive Console Testing for console.apidirect.dev\n');
  console.log('═'.repeat(60) + '\n');
  
  const browser = await chromium.launch({ 
    headless: false,
    slowMo: 500  // Slow down for visibility
  });
  
  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
    recordVideo: {
      dir: './test-videos',
      size: { width: 1280, height: 720 }
    }
  });
  
  const page = await context.newPage();
  
  // Track console messages and errors
  const consoleMessages = [];
  const pageErrors = [];
  
  page.on('console', msg => {
    consoleMessages.push({
      type: msg.type(),
      text: msg.text(),
      location: msg.location()
    });
  });
  
  page.on('pageerror', error => {
    pageErrors.push(error.toString());
  });
  
  // Track network requests
  const apiRequests = [];
  const resourceTimings = {};
  
  page.on('request', request => {
    const url = request.url();
    if (url.includes('api') || url.includes('auth')) {
      apiRequests.push({
        url: url,
        method: request.method(),
        headers: request.headers(),
        timestamp: Date.now()
      });
    }
  });
  
  page.on('response', response => {
    const url = response.url();
    resourceTimings[url] = {
      status: response.status(),
      size: response.headers()['content-length'] || 'unknown'
    };
  });
  
  try {
    // Test 1: Initial Load and Performance
    console.log('📍 TEST 1: Initial Load and Performance\n');
    const startTime = Date.now();
    
    const response = await page.goto('https://console.apidirect.dev', {
      waitUntil: 'networkidle',
      timeout: 30000
    });
    
    const loadTime = Date.now() - startTime;
    
    console.log(`✓ Page loaded in ${loadTime}ms`);
    console.log(`✓ Response status: ${response.status()}`);
    console.log(`✓ Response headers:`);
    const headers = response.headers();
    for (const [key, value] of Object.entries(headers)) {
      if (key.toLowerCase().includes('security') || 
          key.toLowerCase().includes('content') ||
          key.toLowerCase().includes('x-')) {
        console.log(`  - ${key}: ${value}`);
      }
    }
    
    // Test 2: Page Metadata and SEO
    console.log('\n📍 TEST 2: Page Metadata and SEO\n');
    
    const title = await page.title();
    const description = await page.$eval('meta[name="description"]', el => el.content).catch(() => 'Not found');
    const viewport = await page.$eval('meta[name="viewport"]', el => el.content).catch(() => 'Not found');
    const charset = await page.$eval('meta[charset]', el => el.getAttribute('charset')).catch(() => 'Not found');
    
    console.log(`✓ Title: ${title}`);
    console.log(`✓ Description: ${description}`);
    console.log(`✓ Viewport: ${viewport}`);
    console.log(`✓ Charset: ${charset}`);
    
    // Test 3: Authentication System
    console.log('\n📍 TEST 3: Authentication System Analysis\n');
    
    // Check for auth forms
    const authElements = {
      'Email input': await page.$$('input[type="email"], input[name="email"], input[placeholder*="email" i]'),
      'Password input': await page.$$('input[type="password"]'),
      'Username input': await page.$$('input[name="username"], input[placeholder*="username" i]'),
      'Submit button': await page.$$('button[type="submit"], input[type="submit"]'),
      'Login link': await page.$$('a[href*="login"], button:has-text("Login"), button:has-text("Sign in")'),
      'Register link': await page.$$('a[href*="register"], a[href*="signup"], button:has-text("Register")'),
      'OAuth buttons': await page.$$('button:has-text("Google"), button:has-text("GitHub"), button:has-text("Microsoft")')
    };
    
    for (const [name, elements] of Object.entries(authElements)) {
      console.log(`${elements.length > 0 ? '✓' : '✗'} ${name}: ${elements.length} found`);
    }
    
    // Test 4: Navigation Structure
    console.log('\n📍 TEST 4: Navigation Structure\n');
    
    const navItems = await page.$$eval('nav a, header a, .sidebar a, [role="navigation"] a', links => 
      links.map(link => ({
        text: link.textContent.trim(),
        href: link.href,
        target: link.target
      }))
    );
    
    console.log(`✓ Found ${navItems.length} navigation links:`);
    navItems.forEach(item => {
      console.log(`  - ${item.text || 'No text'}: ${item.href}`);
    });
    
    // Test 5: Form Validation
    console.log('\n📍 TEST 5: Form Validation Testing\n');
    
    const emailInput = await page.$('input[type="email"], input[name="email"]');
    if (emailInput) {
      // Test invalid email
      await emailInput.fill('invalid-email');
      await emailInput.press('Tab');
      await page.waitForTimeout(500);
      
      const validationMessage = await emailInput.evaluate(el => el.validationMessage);
      console.log(`✓ Email validation: ${validationMessage || 'No browser validation'}`);
      
      // Check for custom validation messages
      const errorMessages = await page.$$('.error, .invalid, [class*="error"], [class*="invalid"]');
      console.log(`✓ Custom error elements: ${errorMessages.length} found`);
    } else {
      console.log('✗ No email input found to test validation');
    }
    
    // Test 6: Accessibility
    console.log('\n📍 TEST 6: Accessibility Testing\n');
    
    // Check for ARIA labels
    const ariaElements = await page.$$('[aria-label], [aria-describedby], [role]');
    console.log(`✓ ARIA elements: ${ariaElements.length} found`);
    
    // Check for alt texts
    const images = await page.$$('img');
    let imagesWithAlt = 0;
    for (const img of images) {
      const alt = await img.getAttribute('alt');
      if (alt) imagesWithAlt++;
    }
    console.log(`✓ Images with alt text: ${imagesWithAlt}/${images.length}`);
    
    // Check heading structure
    const headings = await page.$$eval('h1, h2, h3, h4, h5, h6', elements =>
      elements.map(el => ({
        level: el.tagName,
        text: el.textContent.trim()
      }))
    );
    console.log(`✓ Heading structure: ${headings.length} headings found`);
    headings.forEach(h => {
      console.log(`  - ${h.level}: ${h.text.substring(0, 50)}${h.text.length > 50 ? '...' : ''}`);
    });
    
    // Test 7: Interactive Elements
    console.log('\n📍 TEST 7: Interactive Elements Testing\n');
    
    const buttons = await page.$$('button, input[type="button"], input[type="submit"], a[role="button"]');
    console.log(`✓ Interactive buttons: ${buttons.length} found`);
    
    // Test first clickable button
    if (buttons.length > 0) {
      const firstButton = buttons[0];
      const buttonText = await firstButton.textContent();
      console.log(`✓ Testing click on: "${buttonText?.trim()}"`);
      
      try {
        await firstButton.click({ timeout: 5000 });
        await page.waitForTimeout(1000);
        console.log('  - Click successful');
        
        // Check if navigation occurred
        const newUrl = page.url();
        console.log(`  - Current URL: ${newUrl}`);
      } catch (e) {
        console.log('  - Click failed or button not interactive');
      }
    }
    
    // Test 8: API Integration
    console.log('\n📍 TEST 8: API Integration Analysis\n');
    
    // Check for API configuration
    const scripts = await page.$$eval('script', scripts => 
      scripts.map(s => s.src).filter(src => src)
    );
    
    const apiScripts = scripts.filter(src => 
      src.includes('api') || src.includes('client') || src.includes('sdk')
    );
    
    console.log(`✓ API-related scripts: ${apiScripts.length} found`);
    apiScripts.forEach(script => {
      console.log(`  - ${script}`);
    });
    
    // Check localStorage for tokens
    const localStorageData = await page.evaluate(() => {
      const data = {};
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key.includes('token') || key.includes('auth') || key.includes('user')) {
          data[key] = localStorage.getItem(key) ? 'Present' : 'Empty';
        }
      }
      return data;
    });
    
    console.log('✓ LocalStorage auth data:');
    Object.entries(localStorageData).forEach(([key, value]) => {
      console.log(`  - ${key}: ${value}`);
    });
    
    // Test 9: Responsive Design
    console.log('\n📍 TEST 9: Responsive Design Testing\n');
    
    const viewports = [
      { name: 'Desktop HD', width: 1920, height: 1080 },
      { name: 'Desktop', width: 1366, height: 768 },
      { name: 'Tablet Landscape', width: 1024, height: 768 },
      { name: 'Tablet Portrait', width: 768, height: 1024 },
      { name: 'Mobile', width: 375, height: 667 },
      { name: 'Mobile Small', width: 320, height: 568 }
    ];
    
    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await page.waitForTimeout(500);
      
      // Check if mobile menu appears
      const mobileMenu = await page.$('.mobile-menu, .hamburger, [aria-label*="menu" i], .menu-toggle');
      const isMenuVisible = mobileMenu ? await mobileMenu.isVisible() : false;
      
      // Check layout changes
      const mainContent = await page.$('main, .main, #main, .content');
      const contentWidth = mainContent ? await mainContent.evaluate(el => el.offsetWidth) : 0;
      
      console.log(`✓ ${viewport.name} (${viewport.width}x${viewport.height}):`);
      console.log(`  - Mobile menu: ${isMenuVisible ? 'Visible' : 'Hidden'}`);
      console.log(`  - Content width: ${contentWidth}px`);
      
      // Take screenshot
      await page.screenshot({ 
        path: `console-${viewport.name.toLowerCase().replace(' ', '-')}.png`,
        fullPage: false 
      });
    }
    
    // Test 10: Security Headers and Practices
    console.log('\n📍 TEST 10: Security Analysis\n');
    
    // Check cookies
    const cookies = await context.cookies();
    console.log(`✓ Cookies: ${cookies.length} found`);
    cookies.forEach(cookie => {
      console.log(`  - ${cookie.name}:`);
      console.log(`    - Secure: ${cookie.secure}`);
      console.log(`    - HttpOnly: ${cookie.httpOnly}`);
      console.log(`    - SameSite: ${cookie.sameSite}`);
    });
    
    // Check for mixed content
    const mixedContent = consoleMessages.filter(msg => 
      msg.text.includes('Mixed Content') || msg.text.includes('was loaded over HTTPS')
    );
    console.log(`✓ Mixed content warnings: ${mixedContent.length}`);
    
    // Test 11: Performance Metrics
    console.log('\n📍 TEST 11: Performance Metrics\n');
    
    const performanceMetrics = await page.evaluate(() => {
      const timing = performance.timing;
      const navigation = performance.getEntriesByType('navigation')[0];
      
      return {
        dns: timing.domainLookupEnd - timing.domainLookupStart,
        tcp: timing.connectEnd - timing.connectStart,
        ttfb: timing.responseStart - timing.navigationStart,
        domContentLoaded: timing.domContentLoadedEventEnd - timing.navigationStart,
        load: timing.loadEventEnd - timing.navigationStart,
        domInteractive: timing.domInteractive - timing.navigationStart,
        resources: performance.getEntriesByType('resource').length,
        transferSize: navigation?.transferSize || 0,
        encodedBodySize: navigation?.encodedBodySize || 0
      };
    });
    
    console.log('✓ Performance timings:');
    Object.entries(performanceMetrics).forEach(([key, value]) => {
      console.log(`  - ${key}: ${value}${key.includes('Size') ? ' bytes' : key === 'resources' ? ' resources' : 'ms'}`);
    });
    
    // Test 12: Error Summary
    console.log('\n📍 TEST 12: Error and Warning Summary\n');
    
    console.log(`✓ Page errors: ${pageErrors.length}`);
    pageErrors.forEach(error => {
      console.log(`  - ${error}`);
    });
    
    const warnings = consoleMessages.filter(msg => msg.type === 'warning');
    console.log(`✓ Console warnings: ${warnings.length}`);
    warnings.forEach(warning => {
      console.log(`  - ${warning.text}`);
    });
    
    const errors = consoleMessages.filter(msg => msg.type === 'error');
    console.log(`✓ Console errors: ${errors.length}`);
    errors.forEach(error => {
      console.log(`  - ${error.text}`);
    });
    
    // Test 13: API Request Summary
    console.log('\n📍 TEST 13: API Request Analysis\n');
    
    console.log(`✓ Total API requests: ${apiRequests.length}`);
    apiRequests.forEach(req => {
      console.log(`  - ${req.method} ${req.url}`);
    });
    
    // Final Summary
    console.log('\n' + '═'.repeat(60));
    console.log('\n📊 FINAL TEST SUMMARY\n');
    
    const summary = {
      'Page Load': loadTime < 3000 ? '✅ PASS' : '⚠️  SLOW',
      'HTTPS': response.url().startsWith('https') ? '✅ PASS' : '❌ FAIL',
      'Authentication': Object.values(authElements).some(el => el.length > 0) ? '✅ FOUND' : '⚠️  MISSING',
      'Navigation': navItems.length > 0 ? '✅ PASS' : '❌ FAIL',
      'Accessibility': imagesWithAlt === images.length ? '✅ PASS' : '⚠️  PARTIAL',
      'Responsive': '✅ TESTED',
      'Security Headers': headers['strict-transport-security'] ? '✅ PASS' : '⚠️  MISSING',
      'Console Errors': errors.length === 0 ? '✅ NONE' : `⚠️  ${errors.length} ERRORS`,
      'Performance': performanceMetrics.load < 5000 ? '✅ GOOD' : '⚠️  SLOW'
    };
    
    Object.entries(summary).forEach(([test, result]) => {
      console.log(`${test}: ${result}`);
    });
    
    console.log('\n✅ Comprehensive testing completed!');
    console.log('📸 Screenshots saved for all viewports');
    console.log('🎥 Video recording saved in ./test-videos');
    
  } catch (error) {
    console.error('\n❌ Test failed:', error.message);
    console.error('Stack trace:', error.stack);
  } finally {
    console.log('\n🏁 Closing browser in 10 seconds...');
    await page.waitForTimeout(10000);
    await browser.close();
  }
})();