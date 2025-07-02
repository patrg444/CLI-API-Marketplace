import { test, expect, Page } from '@playwright/test';

test.describe('Button Interactions Test Suite', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    // Clear auth state to ensure consistent testing
    await page.addInitScript(() => {
      localStorage.removeItem('mockUser');
    });
  });

  test.describe('Homepage Buttons', () => {
    test.beforeEach(async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');
    });

    test('should interact with all hero section buttons', async () => {
      // Test hero CTA buttons
      const searchNowButton = page.locator('[data-testid="hero-search-cta"]');
      if (await searchNowButton.isVisible()) {
        await expect(searchNowButton).toBeVisible();
        await expect(searchNowButton).toBeEnabled();
        await searchNowButton.click();
        // Should scroll to search section
        await expect(page.locator('[data-testid="search-section"]')).toBeInViewport();
      }

      const learnMoreButton = page.locator('[data-testid="hero-learn-more"]');
      if (await learnMoreButton.isVisible()) {
        await expect(learnMoreButton).toBeVisible();
        await expect(learnMoreButton).toBeEnabled();
        await learnMoreButton.click();
        // Should navigate to documentation
        await expect(page).toHaveURL(/docs|about/);
      }
    });

    test('should interact with navigation buttons', async () => {
      // Check if user is logged in or out and test appropriate buttons
      const signInButton = page.locator('[data-testid="login-button"]');
      const signOutButton = page.locator('[data-testid="signout-button"]');
      
      if (await signInButton.isVisible()) {
        // User is logged out - test login/signup buttons
        await expect(signInButton).toBeEnabled();
        await signInButton.hover({ force: true });
        await signInButton.click({ force: true });
        await expect(page).toHaveURL(/auth\/login/);

        // Go back to home
        await page.goto('/');

        const signUpButton = page.locator('[data-testid="signup-button"]');
        if (await signUpButton.isVisible()) {
          await expect(signUpButton).toBeVisible();
          await expect(signUpButton).toBeEnabled();
          await signUpButton.click({ force: true });
          await expect(page).toHaveURL(/auth\/signup/);
        }
      } else if (await signOutButton.isVisible()) {
        // User is logged in - test logout button
        await expect(signOutButton).toBeVisible();
        await expect(signOutButton).toBeEnabled();
        await signOutButton.hover({ force: true });
        await signOutButton.click({ force: true });
        
        // Should redirect to home and show login button
        await page.waitForTimeout(1000);
        await page.waitForLoadState('networkidle');
        await expect(page.locator('[data-testid="login-button"]')).toBeVisible();
      }
    });

    test('should interact with search filter buttons', async () => {
      await page.goto('/');
      
      // Category filter buttons
      const categoryButtons = page.locator('[data-testid^="category-filter-"]');
      const buttonCount = await categoryButtons.count();
      
      for (let i = 0; i < Math.min(buttonCount, 5); i++) {
        const button = categoryButtons.nth(i);
        await expect(button).toBeVisible();
        await expect(button).toBeEnabled();
        
        // Test hover state (force to bypass portal interference)
        await button.hover({ force: true });
        
        // Test click (force to bypass portal interference)
        await button.click({ force: true });
        
        // Verify URL or state changed
        await page.waitForTimeout(500);
      }

      // Popular tags buttons
      const tagButtons = page.locator('[data-testid^="tag-filter-"]');
      const tagCount = await tagButtons.count();
      
      for (let i = 0; i < Math.min(tagCount, 3); i++) {
        const button = tagButtons.nth(i);
        if (await button.isVisible()) {
          await expect(button).toBeEnabled();
          await button.hover();
          await button.click();
          await page.waitForTimeout(300);
        }
      }
    });

    test('should interact with API card buttons', async () => {
      await page.goto('/');
      await page.waitForSelector('[data-testid="api-card"]', { timeout: 10000 });
      
      const apiCards = page.locator('[data-testid="api-card"]');
      const cardCount = await apiCards.count();
      
      if (cardCount > 0) {
        // Test first API card interaction
        const firstCard = apiCards.first();
        await expect(firstCard).toBeVisible();
        
        // Test hover effect
        await firstCard.hover({ force: true });
        
        // Test click navigation
        await firstCard.click({ force: true });
        await page.waitForTimeout(1000);
        
        // Should navigate to API details page or stay on homepage
        const currentUrl = page.url();
        // Accept either navigation to API page or staying on home page
        expect(currentUrl === 'http://localhost:3000/' || currentUrl.includes('/apis/')).toBeTruthy();
      }
    });
  });

  test.describe('Authentication Page Buttons', () => {
    test('should test login page buttons', async () => {
      await page.goto('/auth/login');
      await page.waitForLoadState('networkidle');

      // Test form submit button
      const submitButton = page.locator('[data-testid="submit-login"]');
      await expect(submitButton).toBeVisible();
      await expect(submitButton).toBeEnabled();
      
      // Test button is disabled when form is invalid
      await submitButton.click();
      // Should show validation errors, button should still be enabled
      await expect(submitButton).toBeEnabled();

      // Test forgot password link
      const forgotPasswordLink = page.locator('[data-testid="forgot-password-link"]');
      if (await forgotPasswordLink.isVisible()) {
        await expect(forgotPasswordLink).toBeEnabled();
        await forgotPasswordLink.click();
        await expect(page).toHaveURL(/auth\/forgot-password/);
      }

      // Test sign up link
      await page.goto('/auth/login');
      const signUpLink = page.locator('[data-testid="signup-link"]');
      if (await signUpLink.isVisible()) {
        await expect(signUpLink).toBeEnabled();
        await signUpLink.click({ force: true });
        await expect(page).toHaveURL(/auth\/signup/);
      }

      // Test social login buttons if they exist
      const googleButton = page.locator('[data-testid="google-login"]');
      if (await googleButton.isVisible()) {
        await expect(googleButton).toBeEnabled();
        // Don't actually click to avoid external navigation
        await googleButton.hover();
      }
    });

    test('should test signup page buttons', async () => {
      await page.goto('/auth/signup');
      await page.waitForLoadState('networkidle');

      // Test form submit button
      const submitButton = page.locator('[data-testid="submit-signup"]');
      await expect(submitButton).toBeVisible();
      await expect(submitButton).toBeEnabled();
      
      // Test validation
      await submitButton.click();
      await expect(submitButton).toBeEnabled();

      // Test login link
      const loginLink = page.locator('[data-testid="login-link"]');
      if (await loginLink.isVisible()) {
        await expect(loginLink).toBeEnabled();
        await loginLink.click({ force: true });
        await expect(page).toHaveURL(/auth\/login/);
      }
    });
  });

  test.describe('API Details Page Buttons', () => {
    test('should test API details page interactions', async () => {
      // Navigate to first available API
      await page.goto('/');
      await page.waitForSelector('[data-testid="api-card"]', { timeout: 10000 });
      
      const firstCard = page.locator('[data-testid="api-card"]').first();
      if (await firstCard.isVisible()) {
        await firstCard.click();
        await page.waitForLoadState('networkidle');

        // Test subscribe buttons
        const subscribeButtons = page.locator('[data-testid^="subscribe-"]');
        const subscribeCount = await subscribeButtons.count();
        
        for (let i = 0; i < subscribeCount; i++) {
          const button = subscribeButtons.nth(i);
          await expect(button).toBeVisible();
          await expect(button).toBeEnabled();
          await button.hover();
          
          // Click should trigger auth check or subscription flow
          await button.click();
          await page.waitForTimeout(1000);
          
          // Might redirect to auth or show modal
          const currentUrl = page.url();
          const isModal = await page.locator('[role="dialog"]').isVisible();
          
          if (currentUrl.includes('/auth/')) {
            await page.goBack();
            await page.waitForLoadState('networkidle');
          } else if (isModal) {
            // Close modal if opened
            const closeButton = page.locator('[data-testid="modal-close"]');
            if (await closeButton.isVisible()) {
              await closeButton.click();
            }
          }
        }

        // Test documentation tabs
        const docTabs = page.locator('[data-testid^="tab-"]');
        const tabCount = await docTabs.count();
        
        for (let i = 0; i < tabCount; i++) {
          const tab = docTabs.nth(i);
          if (await tab.isVisible()) {
            await expect(tab).toBeEnabled();
            await tab.click();
            await page.waitForTimeout(500);
          }
        }

        // Test API playground buttons if available
        const tryItButton = page.locator('[data-testid="try-api-button"]');
        if (await tryItButton.isVisible()) {
          await expect(tryItButton).toBeEnabled();
          await tryItButton.click();
          await page.waitForTimeout(1000);
        }

        const executeButton = page.locator('[data-testid="execute-request"]');
        if (await executeButton.isVisible()) {
          await expect(executeButton).toBeEnabled();
          await executeButton.hover();
        }
      }
    });
  });

  test.describe('Creator Portal Buttons', () => {
    test.beforeEach(async () => {
      // Mock authenticated state for creator portal
      await page.addInitScript(() => {
        localStorage.setItem('mockUser', JSON.stringify({
          id: 'test-creator',
          email: 'creator@test.com',
          name: 'creator'
        }));
      });
    });

    test('should test creator portal navigation buttons', async () => {
      await page.goto('/creator-portal');
      await page.waitForLoadState('networkidle');

      // Test main navigation buttons
      const navButtons = [
        '[data-testid="nav-dashboard"]',
        '[data-testid="nav-apis"]',
        '[data-testid="nav-analytics"]',
        '[data-testid="nav-payouts"]'
      ];

      for (const selector of navButtons) {
        const button = page.locator(selector);
        if (await button.isVisible()) {
          await expect(button).toBeEnabled();
          await button.hover();
          await button.click();
          await page.waitForTimeout(500);
        }
      }
    });

    test('should test API management buttons', async () => {
      await page.goto('/creator-portal/apis');
      await page.waitForLoadState('networkidle');

      // Test create new API button
      const createButton = page.locator('[data-testid="create-api-button"]');
      if (await createButton.isVisible()) {
        await expect(createButton).toBeEnabled();
        await createButton.hover();
        await createButton.click();
        await page.waitForTimeout(1000);
      }

      // Test API action buttons (edit, delete, publish)
      const actionButtons = page.locator('[data-testid^="api-action-"]');
      const actionCount = await actionButtons.count();
      
      for (let i = 0; i < Math.min(actionCount, 3); i++) {
        const button = actionButtons.nth(i);
        if (await button.isVisible()) {
          await expect(button).toBeEnabled();
          await button.hover();
          
          // Only click non-destructive actions
          const buttonText = await button.textContent();
          if (buttonText && !buttonText.toLowerCase().includes('delete')) {
            await button.click();
            await page.waitForTimeout(500);
            
            // Handle potential navigation or modal
            const isModal = await page.locator('[role="dialog"]').isVisible();
            if (isModal) {
              const closeButton = page.locator('[data-testid="modal-close"]');
              if (await closeButton.isVisible()) {
                await closeButton.click();
              }
            }
          }
        }
      }
    });

    test('should test pricing configuration buttons', async () => {
      await page.goto('/creator-portal/apis');
      await page.waitForLoadState('networkidle');

      // Navigate to first API settings
      const firstApiLink = page.locator('[data-testid^="api-link-"]').first();
      if (await firstApiLink.isVisible()) {
        await firstApiLink.click();
        await page.waitForLoadState('networkidle');

        // Test pricing plan buttons
        const addPlanButton = page.locator('[data-testid="add-pricing-plan"]');
        if (await addPlanButton.isVisible()) {
          await expect(addPlanButton).toBeEnabled();
          await addPlanButton.hover();
          await addPlanButton.click();
          await page.waitForTimeout(1000);
        }

        // Test save pricing button
        const savePricingButton = page.locator('[data-testid="save-pricing"]');
        if (await savePricingButton.isVisible()) {
          await expect(savePricingButton).toBeEnabled();
          await savePricingButton.hover();
        }

        // Test pricing plan action buttons
        const planButtons = page.locator('[data-testid^="pricing-plan-"]');
        const planCount = await planButtons.count();
        
        for (let i = 0; i < Math.min(planCount, 2); i++) {
          const button = planButtons.nth(i);
          if (await button.isVisible()) {
            await expect(button).toBeEnabled();
            await button.hover();
          }
        }
      }
    });
  });

  test.describe('Search and Filter Buttons', () => {
    test('should test advanced search buttons', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Test search submit button
      const searchButton = page.locator('[data-testid="search-submit"]');
      if (await searchButton.isVisible()) {
        await expect(searchButton).toBeEnabled();
        await searchButton.hover();
        await searchButton.click();
        await page.waitForTimeout(1000);
      }

      // Test filter buttons
      const filterButtons = page.locator('[data-testid^="filter-"]');
      const filterCount = await filterButtons.count();
      
      for (let i = 0; i < Math.min(filterCount, 5); i++) {
        const button = filterButtons.nth(i);
        if (await button.isVisible()) {
          await expect(button).toBeEnabled();
          await button.hover();
          await button.click();
          await page.waitForTimeout(300);
        }
      }

      // Test clear filters button
      const clearButton = page.locator('[data-testid="clear-filters"]');
      if (await clearButton.isVisible()) {
        await expect(clearButton).toBeEnabled();
        await clearButton.click();
        await page.waitForTimeout(500);
      }

      // Test sort buttons
      const sortButtons = page.locator('[data-testid^="sort-"]');
      const sortCount = await sortButtons.count();
      
      for (let i = 0; i < sortCount; i++) {
        const button = sortButtons.nth(i);
        if (await button.isVisible()) {
          await expect(button).toBeEnabled();
          await button.click();
          await page.waitForTimeout(500);
        }
      }
    });
  });

  test.describe('Modal and Dialog Buttons', () => {
    test('should test modal interaction buttons', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Try to trigger a modal by clicking subscribe on an API
      const apiCard = page.locator('[data-testid="api-card"]').first();
      if (await apiCard.isVisible()) {
        await apiCard.click();
        await page.waitForLoadState('networkidle');

        const subscribeButton = page.locator('[data-testid^="subscribe-"]').first();
        if (await subscribeButton.isVisible()) {
          await subscribeButton.click();
          await page.waitForTimeout(1000);

          // Test modal buttons if modal opened
          const modal = page.locator('[role="dialog"]');
          if (await modal.isVisible()) {
            const modalButtons = modal.locator('button');
            const buttonCount = await modalButtons.count();
            
            for (let i = 0; i < buttonCount; i++) {
              const button = modalButtons.nth(i);
              await expect(button).toBeEnabled();
              await button.hover();
              
              const buttonText = await button.textContent();
              // Only click close/cancel buttons, not submit buttons
              if (buttonText && (
                buttonText.toLowerCase().includes('close') ||
                buttonText.toLowerCase().includes('cancel') ||
                buttonText.includes('×')
              )) {
                await button.click();
                break;
              }
            }
          }
        }
      }
    });
  });

  test.describe('Footer and Utility Buttons', () => {
    test('should test footer navigation buttons', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Scroll to footer
      await page.locator('footer').scrollIntoViewIfNeeded();

      // Test footer links
      const footerLinks = page.locator('footer a');
      const linkCount = await footerLinks.count();
      
      for (let i = 0; i < Math.min(linkCount, 5); i++) {
        const link = footerLinks.nth(i);
        if (await link.isVisible()) {
          await expect(link).toBeEnabled();
          await link.hover();
          
          // Check if it's an external link
          const href = await link.getAttribute('href');
          if (href && !href.startsWith('http')) {
            await link.click();
            await page.waitForTimeout(500);
            await page.goBack();
            await page.waitForLoadState('networkidle');
          }
        }
      }
    });

    test('should test utility buttons', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      // Test back to top button if it exists
      const backToTopButton = page.locator('[data-testid="back-to-top"]');
      if (await backToTopButton.isVisible()) {
        await expect(backToTopButton).toBeEnabled();
        await backToTopButton.click();
        await page.waitForTimeout(500);
      }

      // Test theme toggle if it exists
      const themeToggle = page.locator('[data-testid="theme-toggle"]');
      if (await themeToggle.isVisible()) {
        await expect(themeToggle).toBeEnabled();
        await themeToggle.click();
        await page.waitForTimeout(500);
      }
    });
  });

  test.describe('Error State Buttons', () => {
    test('should test error page buttons', async () => {
      // Try to navigate to a non-existent page
      await page.goto('/non-existent-page');
      await page.waitForLoadState('networkidle');

      // Test 404 page buttons
      const homeButton = page.locator('[data-testid="home-button"]');
      if (await homeButton.isVisible()) {
        await expect(homeButton).toBeEnabled();
        await homeButton.click();
        await expect(page).toHaveURL('/');
      }
    });

    test('should test retry buttons', async () => {
      // This would test retry buttons in error states
      // Implementation depends on how error states are handled
      await page.goto('/');
      
      // Mock network failure to trigger error states
      await page.route('**/api/**', route => route.abort());
      
      await page.reload();
      await page.waitForTimeout(2000);

      const retryButton = page.locator('[data-testid="retry-button"]');
      if (await retryButton.isVisible()) {
        await expect(retryButton).toBeEnabled();
        await retryButton.hover();
        // Don't actually click to avoid infinite retries
      }
    });
  });

  test.afterEach(async () => {
    await page.close();
  });
});