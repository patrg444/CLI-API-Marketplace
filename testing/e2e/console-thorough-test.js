const { chromium } = require('playwright');
const fs = require('fs');

(async () => {
  console.log('🧪 Thorough Console Testing for console.apidirect.dev\n');
  console.log('═'.repeat(60) + '\n');
  
  const browser = await chromium.launch({ 
    headless: true,  // Run headless for stability
    slowMo: 100
  });
  
  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36'
  });
  
  const page = await context.newPage();
  
  // Test results storage
  const testResults = {
    timestamp: new Date().toISOString(),
    url: 'https://console.apidirect.dev',
    tests: {}
  };
  
  // Helper function to safely test elements
  async function safeTest(testName, testFn) {
    try {
      const result = await testFn();
      testResults.tests[testName] = { status: 'PASS', ...result };
      console.log(`✅ ${testName}: PASS`);
      if (result.details) {
        Object.entries(result.details).forEach(([key, value]) => {
          console.log(`   - ${key}: ${value}`);
        });
      }
    } catch (error) {
      testResults.tests[testName] = { status: 'FAIL', error: error.message };
      console.log(`❌ ${testName}: FAIL - ${error.message}`);
    }
  }
  
  try {
    // Navigate to the console
    await safeTest('Page Load', async () => {
      const startTime = Date.now();
      const response = await page.goto('https://console.apidirect.dev', {
        waitUntil: 'domcontentloaded',
        timeout: 30000
      });
      const loadTime = Date.now() - startTime;
      
      return {
        details: {
          'Load Time': `${loadTime}ms`,
          'Status Code': response.status(),
          'URL': page.url(),
          'HTTPS': page.url().startsWith('https')
        }
      };
    });
    
    // Wait for page to stabilize
    await page.waitForTimeout(2000);
    
    // Test page metadata
    await safeTest('Page Metadata', async () => {
      const title = await page.title();
      const description = await page.$eval('meta[name="description"]', el => el.content).catch(() => null);
      const viewport = await page.$eval('meta[name="viewport"]', el => el.content).catch(() => null);
      
      return {
        details: {
          'Title': title,
          'Has Description': !!description,
          'Has Viewport': !!viewport,
          'Title Length': `${title.length} chars`
        }
      };
    });
    
    // Test navigation structure
    await safeTest('Navigation Analysis', async () => {
      const links = await page.$$eval('a', elements => 
        elements.map(el => ({
          text: el.textContent.trim(),
          href: el.href,
          isExternal: !el.href.includes(window.location.hostname)
        }))
      );
      
      const navSections = ['APIs', 'Analytics', 'Dashboard', 'Settings'];
      const foundSections = navSections.filter(section => 
        links.some(link => link.text.toLowerCase().includes(section.toLowerCase()))
      );
      
      return {
        details: {
          'Total Links': links.length,
          'External Links': links.filter(l => l.isExternal).length,
          'Navigation Sections': `${foundSections.length}/${navSections.length} found`,
          'Found Sections': foundSections.join(', ')
        }
      };
    });
    
    // Test authentication elements
    await safeTest('Authentication Elements', async () => {
      const authSelectors = {
        email: 'input[type="email"], input[name="email"], input[id="email"]',
        password: 'input[type="password"]',
        submit: 'button[type="submit"], input[type="submit"], button:has-text("Sign in"), button:has-text("Login")',
        oauth: 'button:has-text("Google"), button:has-text("GitHub")'
      };
      
      const results = {};
      for (const [name, selector] of Object.entries(authSelectors)) {
        const elements = await page.$$(selector);
        results[name] = elements.length;
      }
      
      // Check if elements are visible
      const emailInput = await page.$(authSelectors.email);
      const isEmailVisible = emailInput ? await emailInput.isVisible() : false;
      
      return {
        details: {
          'Email Inputs': results.email,
          'Password Inputs': results.password,
          'Submit Buttons': results.submit,
          'OAuth Buttons': results.oauth,
          'Email Input Visible': isEmailVisible
        }
      };
    });
    
    // Test page structure
    await safeTest('Page Structure', async () => {
      const structure = await page.evaluate(() => {
        const elements = {
          headers: document.querySelectorAll('h1, h2, h3').length,
          forms: document.querySelectorAll('form').length,
          buttons: document.querySelectorAll('button').length,
          inputs: document.querySelectorAll('input').length,
          images: document.querySelectorAll('img').length,
          scripts: document.querySelectorAll('script').length,
          stylesheets: document.querySelectorAll('link[rel="stylesheet"]').length
        };
        
        // Check for common sections
        const sections = {
          header: !!document.querySelector('header, .header, [role="banner"]'),
          nav: !!document.querySelector('nav, .nav, [role="navigation"]'),
          main: !!document.querySelector('main, .main, [role="main"]'),
          footer: !!document.querySelector('footer, .footer, [role="contentinfo"]'),
          sidebar: !!document.querySelector('aside, .sidebar, [role="complementary"]')
        };
        
        return { elements, sections };
      });
      
      return {
        details: {
          'Headings': structure.elements.headers,
          'Forms': structure.elements.forms,
          'Buttons': structure.elements.buttons,
          'Inputs': structure.elements.inputs,
          'Has Header': structure.sections.header,
          'Has Navigation': structure.sections.nav,
          'Has Main Content': structure.sections.main,
          'Has Sidebar': structure.sections.sidebar
        }
      };
    });
    
    // Test accessibility
    await safeTest('Accessibility', async () => {
      const a11y = await page.evaluate(() => {
        const images = Array.from(document.querySelectorAll('img'));
        const imagesWithAlt = images.filter(img => img.alt || img.getAttribute('aria-label'));
        
        const buttons = Array.from(document.querySelectorAll('button'));
        const buttonsWithText = buttons.filter(btn => 
          btn.textContent.trim() || btn.getAttribute('aria-label') || btn.getAttribute('title')
        );
        
        const inputsWithLabels = Array.from(document.querySelectorAll('input')).filter(input => {
          const id = input.id;
          return (id && document.querySelector(`label[for="${id}"]`)) || 
                 input.getAttribute('aria-label') || 
                 input.placeholder;
        });
        
        return {
          images: { total: images.length, withAlt: imagesWithAlt.length },
          buttons: { total: buttons.length, withText: buttonsWithText.length },
          inputs: { total: document.querySelectorAll('input').length, withLabels: inputsWithLabels.length },
          lang: document.documentElement.lang,
          ariaElements: document.querySelectorAll('[role], [aria-label], [aria-describedby]').length
        };
      });
      
      return {
        details: {
          'Images with Alt': `${a11y.images.withAlt}/${a11y.images.total}`,
          'Buttons with Text': `${a11y.buttons.withText}/${a11y.buttons.total}`,
          'Inputs with Labels': `${a11y.inputs.withLabels}/${a11y.inputs.total}`,
          'Has Lang Attribute': !!a11y.lang,
          'ARIA Elements': a11y.ariaElements
        }
      };
    });
    
    // Test responsive design
    await safeTest('Responsive Design', async () => {
      const viewports = [
        { name: 'Mobile', width: 375, height: 667 },
        { name: 'Tablet', width: 768, height: 1024 },
        { name: 'Desktop', width: 1920, height: 1080 }
      ];
      
      const results = {};
      
      for (const viewport of viewports) {
        await page.setViewportSize({ width: viewport.width, height: viewport.height });
        await page.waitForTimeout(500);
        
        const layout = await page.evaluate(() => {
          const body = document.body;
          const hasScrollbar = body.scrollWidth > window.innerWidth;
          const mobileMenu = document.querySelector('.mobile-menu, .hamburger, [aria-label*="menu"]');
          const isMobileMenuVisible = mobileMenu ? window.getComputedStyle(mobileMenu).display !== 'none' : false;
          
          return {
            hasHorizontalScroll: hasScrollbar,
            hasMobileMenu: !!mobileMenu,
            isMobileMenuVisible: isMobileMenuVisible
          };
        });
        
        results[viewport.name] = layout;
      }
      
      // Reset to desktop
      await page.setViewportSize({ width: 1280, height: 720 });
      
      return {
        details: {
          'Mobile Menu (Mobile)': results.Mobile.isMobileMenuVisible,
          'Mobile Menu (Desktop)': results.Desktop.isMobileMenuVisible,
          'No Horizontal Scroll': !results.Mobile.hasHorizontalScroll && !results.Tablet.hasHorizontalScroll
        }
      };
    });
    
    // Test JavaScript functionality
    await safeTest('JavaScript Functionality', async () => {
      const jsTest = await page.evaluate(() => {
        // Check if common frameworks are loaded
        const frameworks = {
          React: typeof React !== 'undefined' || !!document.querySelector('[data-reactroot]'),
          Vue: typeof Vue !== 'undefined' || !!document.querySelector('#app'),
          Angular: typeof ng !== 'undefined' || !!document.querySelector('[ng-app]'),
          jQuery: typeof jQuery !== 'undefined' || typeof $ !== 'undefined'
        };
        
        // Check for API client
        const hasAPIClient = typeof APIClient !== 'undefined' || 
                           typeof api !== 'undefined' ||
                           window.apiClient !== undefined;
        
        // Check localStorage
        const hasLocalStorage = typeof localStorage !== 'undefined';
        const storageKeys = hasLocalStorage ? Object.keys(localStorage) : [];
        
        return {
          frameworks: Object.entries(frameworks).filter(([_, loaded]) => loaded).map(([name]) => name),
          hasAPIClient,
          hasLocalStorage,
          storageKeys: storageKeys.length
        };
      });
      
      return {
        details: {
          'Frameworks': jsTest.frameworks.join(', ') || 'None detected',
          'API Client': jsTest.hasAPIClient,
          'LocalStorage': jsTest.hasLocalStorage,
          'Storage Keys': jsTest.storageKeys
        }
      };
    });
    
    // Test console errors
    await safeTest('Console Errors', async () => {
      const messages = [];
      page.on('console', msg => {
        if (msg.type() === 'error' || msg.type() === 'warning') {
          messages.push({
            type: msg.type(),
            text: msg.text()
          });
        }
      });
      
      // Reload page to catch all console messages
      await page.reload();
      await page.waitForTimeout(2000);
      
      const errors = messages.filter(m => m.type === 'error');
      const warnings = messages.filter(m => m.type === 'warning');
      
      return {
        details: {
          'Errors': errors.length,
          'Warnings': warnings.length,
          'First Error': errors[0]?.text || 'None',
          'Clean Console': errors.length === 0
        }
      };
    });
    
    // Test security headers
    await safeTest('Security Headers', async () => {
      const response = await page.goto('https://console.apidirect.dev');
      const headers = response.headers();
      
      const securityHeaders = {
        'strict-transport-security': headers['strict-transport-security'],
        'x-content-type-options': headers['x-content-type-options'],
        'x-frame-options': headers['x-frame-options'],
        'content-security-policy': headers['content-security-policy'],
        'x-xss-protection': headers['x-xss-protection']
      };
      
      const hasSecurityHeaders = Object.values(securityHeaders).filter(h => h).length;
      
      return {
        details: {
          'HSTS': !!securityHeaders['strict-transport-security'],
          'X-Content-Type-Options': !!securityHeaders['x-content-type-options'],
          'X-Frame-Options': !!securityHeaders['x-frame-options'],
          'CSP': !!securityHeaders['content-security-policy'],
          'Security Headers': `${hasSecurityHeaders}/5`
        }
      };
    });
    
    // Performance test
    await safeTest('Performance Metrics', async () => {
      const metrics = await page.evaluate(() => {
        const timing = performance.timing;
        const paint = performance.getEntriesByType('paint');
        const fcp = paint.find(p => p.name === 'first-contentful-paint');
        
        return {
          domContentLoaded: timing.domContentLoadedEventEnd - timing.navigationStart,
          loadComplete: timing.loadEventEnd - timing.navigationStart,
          firstContentfulPaint: fcp ? Math.round(fcp.startTime) : null,
          resources: performance.getEntriesByType('resource').length
        };
      });
      
      return {
        details: {
          'DOM Ready': `${metrics.domContentLoaded}ms`,
          'Page Load': `${metrics.loadComplete}ms`,
          'First Paint': metrics.firstContentfulPaint ? `${metrics.firstContentfulPaint}ms` : 'Not measured',
          'Resources': metrics.resources,
          'Performance': metrics.loadComplete < 3000 ? 'Good' : 'Needs improvement'
        }
      };
    });
    
    // Generate summary
    console.log('\n' + '═'.repeat(60));
    console.log('\n📊 TEST SUMMARY\n');
    
    const passed = Object.values(testResults.tests).filter(t => t.status === 'PASS').length;
    const failed = Object.values(testResults.tests).filter(t => t.status === 'FAIL').length;
    const total = passed + failed;
    
    console.log(`Total Tests: ${total}`);
    console.log(`Passed: ${passed} (${Math.round(passed/total*100)}%)`);
    console.log(`Failed: ${failed} (${Math.round(failed/total*100)}%)`);
    
    // Key findings
    console.log('\n🔍 Key Findings:');
    const pageLoad = testResults.tests['Page Load'];
    if (pageLoad && pageLoad.status === 'PASS') {
      console.log(`- Page loads successfully in ${pageLoad.details['Load Time']}`);
      console.log(`- HTTPS enabled: ${pageLoad.details.HTTPS}`);
    }
    
    const auth = testResults.tests['Authentication Elements'];
    if (auth && auth.status === 'PASS') {
      console.log(`- Authentication elements found: ${auth.details['Email Inputs']} email inputs`);
    }
    
    const responsive = testResults.tests['Responsive Design'];
    if (responsive && responsive.status === 'PASS') {
      console.log(`- Responsive design: ${responsive.details['No Horizontal Scroll'] ? 'Good' : 'Issues found'}`);
    }
    
    const security = testResults.tests['Security Headers'];
    if (security && security.status === 'PASS') {
      console.log(`- Security headers: ${security.details['Security Headers']}`);
    }
    
    // Save detailed report
    const reportPath = 'console-test-report.json';
    fs.writeFileSync(reportPath, JSON.stringify(testResults, null, 2));
    console.log(`\n📄 Detailed report saved to: ${reportPath}`);
    
  } catch (error) {
    console.error('\n❌ Critical test failure:', error.message);
  } finally {
    await browser.close();
    console.log('\n✅ Testing completed!');
  }
})();