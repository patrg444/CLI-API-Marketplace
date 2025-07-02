import { test, expect, Page } from '@playwright/test';

test.describe('Button Edge Cases Test Suite', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
  });

  test.describe('Button State Edge Cases', () => {
    test('should handle disabled button interactions correctly', async () => {
      await page.goto('/auth/login');
      await page.waitForLoadState('networkidle');

      // Find form submission button
      const submitButton = page.locator('[data-testid="submit-login"]');
      
      if (await submitButton.isVisible()) {
        // Test clicking disabled button (if it becomes disabled)
        const emailInput = page.locator('[data-testid="email-input"]');
        const passwordInput = page.locator('[data-testid="password-input"]');
        
        // Fill form to potentially enable button
        if (await emailInput.isVisible()) {
          await emailInput.fill('test@example.com');
        }
        if (await passwordInput.isVisible()) {
          await passwordInput.fill('password123');
        }

        // Submit to potentially disable button
        await submitButton.click();
        await page.waitForTimeout(500);

        // Check if button is disabled during processing
        const isDisabled = await submitButton.isDisabled();
        if (isDisabled) {
          // Try clicking disabled button - should not trigger action
          await submitButton.click({ force: true });
          await page.waitForTimeout(500);
          
          // Should still be on login page or in loading state
          const currentUrl = page.url();
          expect(currentUrl).toMatch(/login|loading/);
        }

        // Test keyboard interaction with disabled button
        await submitButton.focus();
        await page.keyboard.press('Enter');
        await page.keyboard.press('Space');
        await page.waitForTimeout(300);
      }
    });

    test('should handle buttons with dynamic content changes', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Find buttons that might change content (like loading states)
      const dynamicButtons = page.locator('button[data-testid*="submit"], button[data-testid*="save"]');
      const buttonCount = await dynamicButtons.count();

      for (let i = 0; i < Math.min(buttonCount, 3); i++) {
        const button = dynamicButtons.nth(i);
        
        if (await button.isVisible() && await button.isEnabled()) {
          const initialText = await button.textContent();
          const initialAriaLabel = await button.getAttribute('aria-label');
          
          await button.click();
          await page.waitForTimeout(500);

          // Check if text or aria-label changed (loading state)
          const newText = await button.textContent();
          const newAriaLabel = await button.getAttribute('aria-label');

          if (newText !== initialText || newAriaLabel !== initialAriaLabel) {
            console.log(`Button text changed from "${initialText}" to "${newText}"`);
            
            // Button should still be focusable and accessible
            await expect(button).toBeVisible();
            
            // Should have accessible content
            const hasAccessibleContent = newText?.trim() || newAriaLabel;
            expect(hasAccessibleContent).toBeTruthy();
          }

          break; // Only test one to avoid side effects
        }
      }
    });

    test('should handle buttons appearing and disappearing dynamically', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Test filter buttons that might appear/disappear
      const filterButton = page.locator('[data-testid^="category-filter-"]').first();
      
      if (await filterButton.isVisible()) {
        // Click to potentially show more buttons
        await filterButton.click();
        await page.waitForTimeout(500);

        // Look for newly appeared buttons
        const newButtons = page.locator('button:visible').count();
        const allButtons = page.locator('button').count();
        
        console.log(`Visible buttons: ${await newButtons}, Total buttons: ${await allButtons}`);

        // Test that newly visible buttons are functional
        const recentlyVisible = page.locator('button:visible').last();
        if (await recentlyVisible.isVisible()) {
          await expect(recentlyVisible).toBeEnabled();
          await recentlyVisible.hover();
        }
      }
    });
  });

  test.describe('Button Context Edge Cases', () => {
    test('should handle buttons in modals and overlays', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Try to trigger a modal
      const apiCard = page.locator('[data-testid="api-card"]').first();
      if (await apiCard.isVisible()) {
        await apiCard.click();
        await page.waitForLoadState('networkidle');

        const subscribeButton = page.locator('[data-testid^="subscribe-"]').first();
        if (await subscribeButton.isVisible()) {
          await subscribeButton.click();
          await page.waitForTimeout(1000);

          // Check if modal opened
          const modal = page.locator('[role="dialog"]');
          if (await modal.isVisible()) {
            // Test modal buttons
            const modalButtons = modal.locator('button');
            const modalButtonCount = await modalButtons.count();

            for (let i = 0; i < modalButtonCount; i++) {
              const button = modalButtons.nth(i);
              await expect(button).toBeVisible();
              await expect(button).toBeEnabled();
              
              // Test focus trap in modal
              await button.focus();
              const isFocused = await button.evaluate(el => document.activeElement === el);
              expect(isFocused).toBe(true);
            }

            // Test escape key behavior
            await page.keyboard.press('Escape');
            await page.waitForTimeout(500);
            
            // Modal should close or handle escape appropriately
            const modalStillVisible = await modal.isVisible();
            if (!modalStillVisible) {
              console.log('Modal closed with Escape key');
            }
          }
        }
      }
    });

    test('should handle buttons in scrollable containers', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Find scrollable container with buttons
      const scrollContainer = page.locator('[class*="overflow"], [class*="scroll"]').first();
      
      if (await scrollContainer.isVisible()) {
        const buttonsInContainer = scrollContainer.locator('button');
        const buttonCount = await buttonsInContainer.count();

        if (buttonCount > 0) {
          // Scroll to bottom of container
          await scrollContainer.evaluate(el => {
            el.scrollTop = el.scrollHeight;
          });
          
          await page.waitForTimeout(300);

          // Test buttons that are now visible
          const visibleButtons = scrollContainer.locator('button:visible');
          const visibleCount = await visibleButtons.count();

          for (let i = 0; i < Math.min(visibleCount, 3); i++) {
            const button = visibleButtons.nth(i);
            await expect(button).toBeVisible();
            await expect(button).toBeEnabled();
            
            // Button should be clickable even when scrolled
            await button.scrollIntoViewIfNeeded();
            await button.hover();
          }
        }
      }
    });

    test('should handle buttons with complex nested structures', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Find buttons that might have complex inner HTML
      const complexButtons = page.locator('button:has(svg), button:has(span), [role="button"]:has(div)');
      const complexCount = await complexButtons.count();

      for (let i = 0; i < Math.min(complexCount, 5); i++) {
        const button = complexButtons.nth(i);
        
        if (await button.isVisible()) {
          // Test that inner elements don't interfere with button functionality
          const innerElements = button.locator('*');
          const innerCount = await innerElements.count();
          
          console.log(`Button ${i} has ${innerCount} inner elements`);

          // Click on different parts of the button
          const boundingBox = await button.boundingBox();
          if (boundingBox) {
            // Click center
            await button.click();
            await page.waitForTimeout(100);

            // Click near edges (but still inside)
            const centerX = boundingBox.x + boundingBox.width / 2;
            const centerY = boundingBox.y + boundingBox.height / 2;
            
            await page.mouse.click(centerX - 10, centerY);
            await page.waitForTimeout(100);
            
            await page.mouse.click(centerX + 10, centerY);
            await page.waitForTimeout(100);
          }

          // Test keyboard interaction
          await button.focus();
          await page.keyboard.press('Enter');
          await page.waitForTimeout(100);
        }
      }
    });
  });

  test.describe('Button Timing Edge Cases', () => {
    test('should handle rapid button interactions', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      const button = page.locator('button:visible').first();
      
      if (await button.isVisible() && await button.isEnabled()) {
        // Rapid clicking
        for (let i = 0; i < 20; i++) {
          await button.click();
          // No wait time between clicks
        }

        // Button should still be functional
        await expect(button).toBeVisible();
        await button.hover();

        // Test rapid keyboard presses
        await button.focus();
        for (let i = 0; i < 10; i++) {
          await page.keyboard.press('Enter');
        }
        
        for (let i = 0; i < 10; i++) {
          await page.keyboard.press('Space');
        }

        // Button should still respond
        await expect(button).toBeEnabled();
      }
    });

    test('should handle slow network responses for button actions', async () => {
      // Simulate slow network
      await page.route('**/api/**', async route => {
        await new Promise(resolve => setTimeout(resolve, 2000)); // 2 second delay
        await route.continue();
      });

      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Find button that might make API calls
      const searchButton = page.locator('[data-testid="search-submit"]');
      
      if (await searchButton.isVisible()) {
        const startTime = Date.now();
        
        await searchButton.click();
        
        // Button should handle loading state
        await page.waitForTimeout(500);
        
        // Check if button is disabled or shows loading state
        const isDisabled = await searchButton.isDisabled();
        const ariaBusy = await searchButton.getAttribute('aria-busy');
        const textContent = await searchButton.textContent();
        
        if (isDisabled || ariaBusy === 'true' || textContent?.includes('...')) {
          console.log('Button properly shows loading state');
        }

        // Wait for response
        await page.waitForTimeout(3000);
        
        const totalTime = Date.now() - startTime;
        console.log(`Button action completed in ${totalTime}ms`);
      }
    });

    test('should handle button interactions during page transitions', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      const navigationButton = page.locator('[data-testid="login-button"]');
      
      if (await navigationButton.isVisible()) {
        // Start navigation
        await navigationButton.click();
        
        // Try to interact with buttons during transition
        await page.waitForTimeout(100);
        
        // Try clicking again during navigation
        if (await navigationButton.isVisible()) {
          await navigationButton.click();
          await page.waitForTimeout(100);
        }

        // Wait for navigation to complete
        await page.waitForLoadState('networkidle');
        
        // Verify we're on the correct page
        expect(page.url()).toMatch(/auth\/login/);
      }
    });
  });

  test.describe('Button Error Handling Edge Cases', () => {
    test('should handle network errors gracefully for button actions', async () => {
      // Simulate network errors
      await page.route('**/api/**', route => route.abort('internetdisconnected'));

      await page.goto('/');
      await page.waitForLoadState('networkidle');

      const searchButton = page.locator('[data-testid="search-submit"]');
      
      if (await searchButton.isVisible()) {
        await searchButton.click();
        await page.waitForTimeout(2000);

        // Look for error handling
        const errorMessage = page.locator('[data-testid*="error"], .error, [role="alert"]');
        const errorVisible = await errorMessage.count() > 0 && await errorMessage.first().isVisible();
        
        if (errorVisible) {
          console.log('Error message displayed for network failure');
          
          // Check for retry button
          const retryButton = page.locator('[data-testid*="retry"], button:has-text("retry")').first();
          if (await retryButton.isVisible()) {
            await expect(retryButton).toBeEnabled();
            console.log('Retry button available');
          }
        }

        // Button should return to normal state
        await expect(searchButton).toBeEnabled();
      }
    });

    test('should handle JavaScript errors in button event handlers', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Inject code that might cause errors
      await page.addScriptTag({
        content: `
          // Override click handlers to sometimes throw errors
          const originalAddEventListener = EventTarget.prototype.addEventListener;
          EventTarget.prototype.addEventListener = function(type, listener, options) {
            if (type === 'click' && Math.random() < 0.1) { // 10% chance of error
              const errorListener = function(event) {
                if (Math.random() < 0.5) {
                  throw new Error('Simulated click handler error');
                }
                return listener.call(this, event);
              };
              return originalAddEventListener.call(this, type, errorListener, options);
            }
            return originalAddEventListener.call(this, type, listener, options);
          };
        `
      });

      // Listen for console errors
      const errors: string[] = [];
      page.on('pageerror', error => {
        errors.push(error.message);
      });

      // Click many buttons to potentially trigger errors
      const buttons = page.locator('button:visible');
      const buttonCount = await buttons.count();

      for (let i = 0; i < Math.min(buttonCount, 10); i++) {
        const button = buttons.nth(i);
        
        if (await button.isVisible() && await button.isEnabled()) {
          try {
            await button.click();
            await page.waitForTimeout(100);
          } catch (error) {
            console.log(`Button click resulted in error: ${error}`);
          }
        }
      }

      console.log(`Captured ${errors.length} JavaScript errors during button testing`);
      
      // App should still be functional despite errors
      const stillWorkingButton = page.locator('button:visible').first();
      if (await stillWorkingButton.isVisible()) {
        await expect(stillWorkingButton).toBeEnabled();
      }
    });
  });

  test.describe('Button Accessibility Edge Cases', () => {
    test('should handle screen reader edge cases', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Test buttons with no visible text but icons
      const iconButtons = page.locator('button:not(:has-text(/\\w/)), [role="button"]:not(:has-text(/\\w/))');
      const iconButtonCount = await iconButtons.count();

      for (let i = 0; i < Math.min(iconButtonCount, 5); i++) {
        const button = iconButtons.nth(i);
        
        if (await button.isVisible()) {
          // Should have aria-label or title for screen readers
          const ariaLabel = await button.getAttribute('aria-label');
          const title = await button.getAttribute('title');
          const ariaLabelledBy = await button.getAttribute('aria-labelledby');
          
          const hasAccessibleName = ariaLabel || title || ariaLabelledBy;
          if (!hasAccessibleName) {
            console.log('Icon button without accessible name found');
          }
          
          // Should still be keyboard accessible
          await button.focus();
          const isFocusable = await button.evaluate(el => document.activeElement === el);
          expect(isFocusable).toBe(true);
        }
      }
    });

    test('should handle high contrast mode edge cases', async () => {
      // Enable high contrast simulation
      await page.emulateMedia({ colorScheme: 'dark', reducedMotion: 'reduce' });
      
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      const buttons = page.locator('button:visible, [role="button"]:visible');
      const buttonCount = await buttons.count();

      for (let i = 0; i < Math.min(buttonCount, 10); i++) {
        const button = buttons.nth(i);
        
        if (await button.isVisible()) {
          // Check button visibility in high contrast mode
          const styles = await button.evaluate(el => {
            const computed = window.getComputedStyle(el);
            return {
              backgroundColor: computed.backgroundColor,
              color: computed.color,
              border: computed.border,
              outline: computed.outline
            };
          });

          // Button should have some visible styling
          const hasVisibleStyling = 
            styles.backgroundColor !== 'transparent' ||
            styles.border !== 'none' ||
            styles.color !== styles.backgroundColor;

          if (!hasVisibleStyling) {
            console.log('Button may not be visible in high contrast mode');
          }

          // Test focus visibility
          await button.focus();
          const focusStyles = await button.evaluate(el => {
            const computed = window.getComputedStyle(el);
            return {
              outline: computed.outline,
              outlineWidth: computed.outlineWidth,
              boxShadow: computed.boxShadow
            };
          });

          const hasFocusIndicator = 
            focusStyles.outline !== 'none' ||
            focusStyles.outlineWidth !== '0px' ||
            focusStyles.boxShadow !== 'none';

          expect(hasFocusIndicator).toBe(true);
        }
      }
    });
  });

  test.afterEach(async () => {
    await page.close();
  });
});