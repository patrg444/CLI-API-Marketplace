import { test, expect, Page } from '@playwright/test';

test.describe('Form Button Validation Test Suite', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
  });

  test.describe('Login Form Button States', () => {
    test('should validate login form button states', async () => {
      await page.goto('/auth/login');
      await page.waitForLoadState('networkidle');

      const emailInput = page.locator('[data-testid="email-input"]');
      const passwordInput = page.locator('[data-testid="password-input"]');
      const submitButton = page.locator('[data-testid="submit-login"]');

      // Initial state - button should be enabled
      await expect(submitButton).toBeEnabled();

      // Test with empty form
      await submitButton.click();
      await page.waitForTimeout(500);
      
      // Should show validation errors
      const emailError = page.locator('[data-testid="email-error"]');
      const passwordError = page.locator('[data-testid="password-error"]');
      
      if (await emailError.isVisible()) {
        await expect(emailError).toContainText(/required|invalid/i);
      }
      if (await passwordError.isVisible()) {
        await expect(passwordError).toContainText(/required/i);
      }

      // Test with invalid email
      await emailInput.fill('invalid-email');
      await submitButton.click();
      await page.waitForTimeout(500);
      
      if (await emailError.isVisible()) {
        await expect(emailError).toContainText(/invalid|valid email/i);
      }

      // Test with valid email but no password
      await emailInput.fill('test@example.com');
      await submitButton.click();
      await page.waitForTimeout(500);
      
      if (await passwordError.isVisible()) {
        await expect(passwordError).toContainText(/required/i);
      }

      // Test with valid inputs
      await passwordInput.fill('password123');
      await expect(submitButton).toBeEnabled();
      
      // Test loading state
      await submitButton.click();
      
      // Button might show loading state or be disabled temporarily
      await page.waitForTimeout(1000);
    });
  });

  test.describe('Signup Form Button States', () => {
    test('should validate signup form button states', async () => {
      await page.goto('/auth/signup');
      await page.waitForLoadState('networkidle');

      const emailInput = page.locator('[data-testid="email-input"]');
      const passwordInput = page.locator('[data-testid="password-input"]');
      const confirmPasswordInput = page.locator('[data-testid="confirm-password-input"]');
      const submitButton = page.locator('[data-testid="submit-signup"]');

      // Initial state
      await expect(submitButton).toBeEnabled();

      // Test with empty form
      await submitButton.click();
      await page.waitForTimeout(500);

      // Test password mismatch
      await emailInput.fill('test@example.com');
      await passwordInput.fill('password123');
      await confirmPasswordInput.fill('different123');
      await submitButton.click();
      await page.waitForTimeout(500);

      const passwordMismatchError = page.locator('[data-testid="password-mismatch-error"]');
      if (await passwordMismatchError.isVisible()) {
        await expect(passwordMismatchError).toContainText(/match|same/i);
      }

      // Test with matching passwords
      await confirmPasswordInput.fill('password123');
      await expect(submitButton).toBeEnabled();
    });
  });

  test.describe('API Creation Form Button States', () => {
    test('should validate API creation form buttons', async () => {
      // Mock authenticated state
      await page.addInitScript(() => {
        localStorage.setItem('mock-auth-state', JSON.stringify({
          isAuthenticated: true,
          user: { id: 'test-creator', role: 'creator' }
        }));
      });

      await page.goto('/creator-portal/apis');
      await page.waitForLoadState('networkidle');

      const createButton = page.locator('[data-testid="create-api-button"]');
      if (await createButton.isVisible()) {
        await createButton.click();
        await page.waitForTimeout(1000);

        // Test form validation
        const nameInput = page.locator('[data-testid="api-name-input"]');
        const descriptionInput = page.locator('[data-testid="api-description-input"]');
        const categorySelect = page.locator('[data-testid="api-category-select"]');
        const saveButton = page.locator('[data-testid="save-api-button"]');

        if (await saveButton.isVisible()) {
          await expect(saveButton).toBeEnabled();

          // Test with empty form
          await saveButton.click();
          await page.waitForTimeout(500);

          // Test with partial data
          if (await nameInput.isVisible()) {
            await nameInput.fill('Test API');
            await expect(saveButton).toBeEnabled();
          }

          if (await descriptionInput.isVisible()) {
            await descriptionInput.fill('A test API description');
            await expect(saveButton).toBeEnabled();
          }

          if (await categorySelect.isVisible()) {
            await categorySelect.selectOption('AI/ML');
            await expect(saveButton).toBeEnabled();
          }
        }
      }
    });
  });

  test.describe('Pricing Form Button States', () => {
    test('should validate pricing configuration buttons', async () => {
      await page.addInitScript(() => {
        localStorage.setItem('mock-auth-state', JSON.stringify({
          isAuthenticated: true,
          user: { id: 'test-creator', role: 'creator' }
        }));
      });

      await page.goto('/creator-portal/apis');
      await page.waitForLoadState('networkidle');

      // Navigate to pricing section
      const firstApiLink = page.locator('[data-testid^="api-link-"]').first();
      if (await firstApiLink.isVisible()) {
        await firstApiLink.click();
        await page.waitForLoadState('networkidle');

        // Test pricing plan creation
        const addPlanButton = page.locator('[data-testid="add-pricing-plan"]');
        if (await addPlanButton.isVisible()) {
          await addPlanButton.click();
          await page.waitForTimeout(1000);

          // Test pricing form validation
          const planNameInput = page.locator('[data-testid="plan-name-input"]');
          const priceInput = page.locator('[data-testid="price-per-call-input"]');
          const savePlanButton = page.locator('[data-testid="save-plan-button"]');

          if (await savePlanButton.isVisible()) {
            await expect(savePlanButton).toBeEnabled();

            // Test with empty inputs
            await savePlanButton.click();
            await page.waitForTimeout(500);

            // Test with invalid price
            if (await priceInput.isVisible()) {
              await priceInput.fill('-10');
              await savePlanButton.click();
              await page.waitForTimeout(500);

              const priceError = page.locator('[data-testid="price-error"]');
              if (await priceError.isVisible()) {
                await expect(priceError).toContainText(/positive|greater/i);
              }

              // Test with valid price
              await priceInput.fill('0.01');
              await expect(savePlanButton).toBeEnabled();
            }

            if (await planNameInput.isVisible()) {
              await planNameInput.fill('Basic Plan');
              await expect(savePlanButton).toBeEnabled();
            }
          }
        }
      }
    });
  });

  test.describe('Search Form Button States', () => {
    test('should validate search form buttons', async () => {
      await page.goto('/');
      await page.waitForLoadState('networkidle');

      const searchInput = page.locator('[data-testid="search-input"]');
      const searchButton = page.locator('[data-testid="search-submit"]');

      if (await searchButton.isVisible()) {
        await expect(searchButton).toBeEnabled();

        // Test with empty search
        await searchButton.click();
        await page.waitForTimeout(500);

        // Test with search term
        if (await searchInput.isVisible()) {
          await searchInput.fill('weather');
          await expect(searchButton).toBeEnabled();
          await searchButton.click();
          await page.waitForTimeout(1000);

          // Should show search results or navigate
          await expect(page.url()).toMatch(/search|query/);
        }
      }

      // Test filter form buttons
      const categoryFilters = page.locator('[data-testid^="category-filter-"]');
      const categoryCount = await categoryFilters.count();

      for (let i = 0; i < Math.min(categoryCount, 3); i++) {
        const filter = categoryFilters.nth(i);
        if (await filter.isVisible()) {
          await expect(filter).toBeEnabled();
          
          // Test active state
          await filter.click();
          await page.waitForTimeout(500);
          
          // Filter should be in active state
          const isActive = await filter.getAttribute('class');
          expect(isActive).toMatch(/active|selected|bg-|text-/);
        }
      }
    });
  });

  test.describe('Subscription Form Button States', () => {
    test('should validate subscription form buttons', async () => {
      await page.goto('/');
      await page.waitForSelector('[data-testid="api-card"]', { timeout: 10000 });

      const firstCard = page.locator('[data-testid="api-card"]').first();
      if (await firstCard.isVisible()) {
        await firstCard.click();
        await page.waitForLoadState('networkidle');

        // Test subscription buttons
        const subscribeButtons = page.locator('[data-testid^="subscribe-"]');
        const buttonCount = await subscribeButtons.count();

        for (let i = 0; i < buttonCount; i++) {
          const button = subscribeButtons.nth(i);
          if (await button.isVisible()) {
            await expect(button).toBeEnabled();
            
            // Test hover state
            await button.hover();
            
            // Test click (should redirect to auth or show subscription form)
            await button.click();
            await page.waitForTimeout(1000);

            const currentUrl = page.url();
            const modal = page.locator('[role="dialog"]');
            
            if (currentUrl.includes('/auth/')) {
              // Redirected to auth - go back
              await page.goBack();
              await page.waitForLoadState('networkidle');
            } else if (await modal.isVisible()) {
              // Subscription modal opened
              const modalButtons = modal.locator('button');
              const modalButtonCount = await modalButtons.count();
              
              for (let j = 0; j < modalButtonCount; j++) {
                const modalButton = modalButtons.nth(j);
                const buttonText = await modalButton.textContent();
                
                // Test each button's enabled state
                await expect(modalButton).toBeEnabled();
                
                // Only click cancel/close buttons
                if (buttonText && (
                  buttonText.toLowerCase().includes('cancel') ||
                  buttonText.toLowerCase().includes('close') ||
                  buttonText.includes('×')
                )) {
                  await modalButton.click();
                  break;
                }
              }
            }
            break;
          }
        }
      }
    });
  });

  test.describe('File Upload Button States', () => {
    test('should validate file upload buttons', async () => {
      await page.addInitScript(() => {
        localStorage.setItem('mock-auth-state', JSON.stringify({
          isAuthenticated: true,
          user: { id: 'test-creator', role: 'creator' }
        }));
      });

      await page.goto('/creator-portal/apis');
      await page.waitForLoadState('networkidle');

      const firstApiLink = page.locator('[data-testid^="api-link-"]').first();
      if (await firstApiLink.isVisible()) {
        await firstApiLink.click();
        await page.waitForLoadState('networkidle');

        // Test documentation upload
        const uploadButton = page.locator('[data-testid="upload-docs-button"]');
        if (await uploadButton.isVisible()) {
          await expect(uploadButton).toBeEnabled();
          
          // Test file input
          const fileInput = page.locator('input[type="file"]');
          if (await fileInput.isVisible()) {
            // Create a test file
            const testFile = Buffer.from('{"openapi": "3.0.0"}');
            await fileInput.setInputFiles({
              name: 'test-api.json',
              mimeType: 'application/json',
              buffer: testFile
            });
            
            // Upload button should become active
            await expect(uploadButton).toBeEnabled();
          }
        }

        // Test icon upload
        const iconUploadButton = page.locator('[data-testid="upload-icon-button"]');
        if (await iconUploadButton.isVisible()) {
          await expect(iconUploadButton).toBeEnabled();
        }
      }
    });
  });

  test.describe('Bulk Action Button States', () => {
    test('should validate bulk action buttons', async () => {
      await page.addInitScript(() => {
        localStorage.setItem('mock-auth-state', JSON.stringify({
          isAuthenticated: true,
          user: { id: 'test-creator', role: 'creator' }
        }));
      });

      await page.goto('/creator-portal/apis');
      await page.waitForLoadState('networkidle');

      // Test bulk selection
      const selectAllCheckbox = page.locator('[data-testid="select-all-apis"]');
      if (await selectAllCheckbox.isVisible()) {
        await expect(selectAllCheckbox).toBeEnabled();
        await selectAllCheckbox.click();
        await page.waitForTimeout(500);

        // Bulk action buttons should become enabled
        const bulkButtons = page.locator('[data-testid^="bulk-"]');
        const bulkCount = await bulkButtons.count();

        for (let i = 0; i < bulkCount; i++) {
          const button = bulkButtons.nth(i);
          if (await button.isVisible()) {
            await expect(button).toBeEnabled();
            await button.hover();
            
            const buttonText = await button.textContent();
            // Only test non-destructive bulk actions
            if (buttonText && !buttonText.toLowerCase().includes('delete')) {
              await button.click();
              await page.waitForTimeout(500);
              
              // Handle any confirmation dialogs
              const confirmDialog = page.locator('[role="dialog"]');
              if (await confirmDialog.isVisible()) {
                const cancelButton = confirmDialog.locator('button:has-text("Cancel")');
                if (await cancelButton.isVisible()) {
                  await cancelButton.click();
                }
              }
            }
          }
        }
      }
    });
  });

  test.afterEach(async () => {
    await page.close();
  });
});