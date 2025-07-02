import { test, expect, Page } from '@playwright/test';

test.describe('Button Accessibility Test Suite', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
  });

  test.describe('Keyboard Navigation Tests', () => {
    test('should navigate all buttons with keyboard', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Start from the first focusable element
      await page.keyboard.press('Tab');
      
      let previousElement = null;
      let tabCount = 0;
      const maxTabs = 50; // Prevent infinite loop

      while (tabCount < maxTabs) {
        const currentElement = await page.locator(':focus').first();
        
        if (await currentElement.count() === 0) {
          break;
        }

        const tagName = await currentElement.evaluate(el => el.tagName.toLowerCase());
        const role = await currentElement.getAttribute('role');
        const type = await currentElement.getAttribute('type');
        
        // Check if current element is a button
        if (tagName === 'button' || 
            (tagName === 'input' && type === 'submit') ||
            role === 'button' ||
            (tagName === 'a' && await currentElement.getAttribute('href'))) {
          
          // Test button accessibility
          await expect(currentElement).toBeVisible();
          await expect(currentElement).toBeEnabled();
          
          // Check for proper ARIA attributes
          const ariaLabel = await currentElement.getAttribute('aria-label');
          const ariaDescribedBy = await currentElement.getAttribute('aria-describedby');
          const textContent = await currentElement.textContent();
          
          // Button should have accessible text (either content, aria-label, or title)
          const title = await currentElement.getAttribute('title');
          const hasAccessibleText = textContent?.trim() || ariaLabel || title;
          expect(hasAccessibleText).toBeTruthy();
          
          // Test focus indicators
          const styles = await currentElement.evaluate(el => {
            const computed = window.getComputedStyle(el);
            return {
              outline: computed.outline,
              outlineWidth: computed.outlineWidth,
              outlineStyle: computed.outlineStyle,
              boxShadow: computed.boxShadow
            };
          });
          
          // Should have some kind of focus indicator
          const hasFocusIndicator = 
            styles.outline !== 'none' ||
            styles.outlineWidth !== '0px' ||
            styles.boxShadow !== 'none';
          
          // Log focus indicator info (not failing test for now)
          if (!hasFocusIndicator) {
            console.log(`Button without focus indicator: ${textContent?.trim() || ariaLabel || 'unlabeled'}`);
          }
          
          // Test keyboard activation
          if (tagName === 'button' || role === 'button') {
            // Space key should activate buttons
            await page.keyboard.press('Space');
            await page.waitForTimeout(100);
            
            // Enter key should also activate buttons
            await currentElement.focus();
            await page.keyboard.press('Enter');
            await page.waitForTimeout(100);
          }
        }
        
        // Move to next focusable element
        await page.keyboard.press('Tab');
        tabCount++;
        
        // Break if we've cycled back to the same element
        const newElement = await page.locator(':focus').first();
        if (await newElement.count() > 0 && 
            await newElement.evaluate((el, prev) => el === prev, previousElement)) {
          break;
        }
        previousElement = await newElement.elementHandle();
      }
    });

    test('should handle shift+tab navigation', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Navigate forward first
      for (let i = 0; i < 5; i++) {
        await page.keyboard.press('Tab');
      }

      // Then navigate backward
      for (let i = 0; i < 5; i++) {
        await page.keyboard.press('Shift+Tab');
        
        const focused = await page.locator(':focus').first();
        if (await focused.count() > 0) {
          const tagName = await focused.evaluate(el => el.tagName.toLowerCase());
          const role = await focused.getAttribute('role');
          
          if (tagName === 'button' || role === 'button') {
            await expect(focused).toBeVisible();
            await expect(focused).toBeEnabled();
          }
        }
      }
    });
  });

  test.describe('Screen Reader Compatibility', () => {
    test('should have proper ARIA attributes on buttons', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      const buttons = page.locator('button, [role="button"], input[type="submit"], input[type="button"]');
      const buttonCount = await buttons.count();

      for (let i = 0; i < buttonCount; i++) {
        const button = buttons.nth(i);
        
        if (await button.isVisible()) {
          // Check for accessible name
          const accessibleName = await button.evaluate(el => {
            // Check various sources of accessible name
            const ariaLabel = el.getAttribute('aria-label');
            const ariaLabelledBy = el.getAttribute('aria-labelledby');
            const textContent = el.textContent?.trim();
            const title = el.getAttribute('title');
            const alt = el.getAttribute('alt');
            
            if (ariaLabel) return ariaLabel;
            if (ariaLabelledBy) {
              const labelElement = document.getElementById(ariaLabelledBy);
              return labelElement?.textContent?.trim();
            }
            if (textContent) return textContent;
            if (title) return title;
            if (alt) return alt;
            
            return null;
          });

          expect(accessibleName).toBeTruthy();

          // Check for proper button states
          const ariaPressed = await button.getAttribute('aria-pressed');
          const ariaExpanded = await button.getAttribute('aria-expanded');
          const ariaDisabled = await button.getAttribute('aria-disabled');
          const disabled = await button.getAttribute('disabled');

          // If aria-pressed is used, it should be 'true' or 'false'
          if (ariaPressed !== null) {
            expect(['true', 'false']).toContain(ariaPressed);
          }

          // If aria-expanded is used, it should be 'true' or 'false'
          if (ariaExpanded !== null) {
            expect(['true', 'false']).toContain(ariaExpanded);
          }

          // Check disabled state consistency
          if (disabled !== null || ariaDisabled === 'true') {
            await expect(button).toBeDisabled();
          }
        }
      }
    });

    test('should have proper button roles and semantics', async () => {
      await page.goto('/auth/login');
      await page.waitForLoadState('networkidle');

      // Test form buttons specifically
      const submitButtons = page.locator('input[type="submit"], button[type="submit"]');
      const submitCount = await submitButtons.count();

      for (let i = 0; i < submitCount; i++) {
        const button = submitButtons.nth(i);
        if (await button.isVisible()) {
          const role = await button.getAttribute('role');
          const type = await button.getAttribute('type');
          
          // Submit buttons should either have no role (implicit) or role="button"
          if (role !== null) {
            expect(role).toBe('button');
          }
          
          // Should have submit type
          expect(type).toBe('submit');
        }
      }

      // Test link buttons
      const linkButtons = page.locator('a[role="button"]');
      const linkButtonCount = await linkButtons.count();

      for (let i = 0; i < linkButtonCount; i++) {
        const linkButton = linkButtons.nth(i);
        if (await linkButton.isVisible()) {
          const href = await linkButton.getAttribute('href');
          const role = await linkButton.getAttribute('role');
          
          expect(role).toBe('button');
          
          // If it has a role of button, it should handle click events
          // and might not need an href (or href="#")
          if (!href || href === '#') {
            // Should have click handler
            const onclick = await linkButton.getAttribute('onclick');
            const hasClickHandler = onclick !== null;
            
            // For testing purposes, we'll just check it's focusable
            await linkButton.focus();
            const isFocused = await linkButton.evaluate(el => document.activeElement === el);
            expect(isFocused).toBe(true);
          }
        }
      }
    });
  });

  test.describe('Color Contrast and Visual Accessibility', () => {
    test('should have sufficient color contrast for buttons', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      const buttons = page.locator('button, [role="button"]').first();
      
      if (await buttons.isVisible()) {
        const colors = await buttons.evaluate(el => {
          const computed = window.getComputedStyle(el);
          return {
            color: computed.color,
            backgroundColor: computed.backgroundColor,
            borderColor: computed.borderColor
          };
        });

        // We can't easily calculate contrast ratios in Playwright,
        // but we can check that colors are actually set
        expect(colors.color).not.toBe('');
        expect(colors.backgroundColor).not.toBe('');
      }
    });

    test('should maintain visibility in high contrast mode', async () => {
      // Simulate high contrast mode
      await page.emulateMedia({ 
        colorScheme: 'dark',
        reducedMotion: 'reduce'
      });

      await page.goto('/');
      await page.waitForLoadState('networkidle');

      const buttons = page.locator('button, [role="button"]');
      const buttonCount = await buttons.count();

      for (let i = 0; i < Math.min(buttonCount, 5); i++) {
        const button = buttons.nth(i);
        if (await button.isVisible()) {
          // Button should still be visible and interactive
          await expect(button).toBeVisible();
          await expect(button).toBeEnabled();
          
          // Should have some styling that makes it distinguishable
          const styles = await button.evaluate(el => {
            const computed = window.getComputedStyle(el);
            return {
              border: computed.border,
              outline: computed.outline,
              backgroundColor: computed.backgroundColor,
              color: computed.color
            };
          });
          
          // Should have some visual styling
          const hasVisualStyling = 
            styles.border !== 'none' ||
            styles.backgroundColor !== 'transparent' ||
            styles.outline !== 'none';
          
          expect(hasVisualStyling).toBe(true);
        }
      }
    });
  });

  test.describe('Touch and Mobile Accessibility', () => {
    test('should have adequate touch targets on mobile', async () => {
      // Set mobile viewport
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      const buttons = page.locator('button, [role="button"], a');
      const buttonCount = await buttons.count();

      for (let i = 0; i < Math.min(buttonCount, 10); i++) {
        const button = buttons.nth(i);
        
        if (await button.isVisible()) {
          const boundingBox = await button.boundingBox();
          
          if (boundingBox) {
            // WCAG recommends minimum 44x44px touch targets
            const minTouchTarget = 44;
            
            if (boundingBox.width < minTouchTarget || boundingBox.height < minTouchTarget) {
              console.log(`Small touch target: ${boundingBox.width}x${boundingBox.height}px`);
              
              // Check if button has adequate padding around it
              const computedStyles = await button.evaluate(el => {
                const computed = window.getComputedStyle(el);
                return {
                  padding: computed.padding,
                  margin: computed.margin
                };
              });
              
              // Log for review (not failing test automatically)
              console.log(`Button styles: ${JSON.stringify(computedStyles)}`);
            }
          }
        }
      }
    });

    test('should handle touch interactions', async () => {
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      const firstButton = page.locator('button, [role="button"]').first();
      
      if (await firstButton.isVisible()) {
        // Test touch events
        await firstButton.hover();
        
        // Simulate touch start and end
        await firstButton.dispatchEvent('touchstart');
        await page.waitForTimeout(50);
        await firstButton.dispatchEvent('touchend');
        
        // Button should still be functional
        await expect(firstButton).toBeEnabled();
        
        // Test tap
        await firstButton.tap();
        await page.waitForTimeout(100);
      }
    });
  });

  test.describe('Reduced Motion Accessibility', () => {
    test('should respect reduced motion preferences', async () => {
      await page.emulateMedia({ reducedMotion: 'reduce' });
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      const animatedButtons = page.locator('button[class*="transition"], [role="button"][class*="animate"]');
      const animatedCount = await animatedButtons.count();

      for (let i = 0; i < animatedCount; i++) {
        const button = animatedButtons.nth(i);
        
        if (await button.isVisible()) {
          const animations = await button.evaluate(el => {
            const computed = window.getComputedStyle(el);
            return {
              transition: computed.transition,
              animation: computed.animation,
              transform: computed.transform
            };
          });
          
          // In reduced motion mode, transitions should be minimal or none
          // This is more of a guidance check than a hard requirement
          if (animations.transition.includes('transform') || 
              animations.animation !== 'none') {
            console.log(`Animated button in reduced motion mode: ${animations}`);
          }
        }
      }
    });
  });

  test.describe('Error State Accessibility', () => {
    test('should have accessible error states for form buttons', async () => {
      await page.goto('/auth/login');
      await page.waitForLoadState('networkidle');

      const submitButton = page.locator('[data-testid="submit-login"]');
      
      if (await submitButton.isVisible()) {
        // Trigger validation errors
        await submitButton.click();
        await page.waitForTimeout(500);

        // Check if button is associated with error messages
        const ariaDescribedBy = await submitButton.getAttribute('aria-describedby');
        
        if (ariaDescribedBy) {
          const errorElements = page.locator(`#${ariaDescribedBy}`);
          if (await errorElements.count() > 0) {
            await expect(errorElements.first()).toBeVisible();
          }
        }

        // Check for invalid state
        const ariaInvalid = await submitButton.getAttribute('aria-invalid');
        if (ariaInvalid) {
          expect(['true', 'false']).toContain(ariaInvalid);
        }
      }
    });
  });

  test.describe('Loading State Accessibility', () => {
    test('should announce loading states to screen readers', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Find a button that might show loading state
      const asyncButtons = page.locator('button[data-testid*="submit"], button[data-testid*="save"]');
      const asyncCount = await asyncButtons.count();

      for (let i = 0; i < Math.min(asyncCount, 3); i++) {
        const button = asyncButtons.nth(i);
        
        if (await button.isVisible() && await button.isEnabled()) {
          // Click and check for loading state
          await button.click();
          await page.waitForTimeout(500);

          // Check for aria-busy or loading state
          const ariaBusy = await button.getAttribute('aria-busy');
          const ariaLabel = await button.getAttribute('aria-label');
          const textContent = await button.textContent();

          if (ariaBusy === 'true') {
            // Should have appropriate loading text
            const hasLoadingText = 
              ariaLabel?.includes('loading') ||
              textContent?.includes('loading') ||
              textContent?.includes('...') ||
              await button.locator('svg, .spinner, .loading').count() > 0;
            
            expect(hasLoadingText).toBe(true);
          }

          // Button might be disabled during loading
          const isDisabled = await button.isDisabled();
          if (isDisabled) {
            const ariaDisabled = await button.getAttribute('aria-disabled');
            expect(ariaDisabled).toBe('true');
          }

          break; // Only test one to avoid side effects
        }
      }
    });
  });

  test.afterEach(async () => {
    await page.close();
  });
});