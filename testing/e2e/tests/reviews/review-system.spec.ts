import { test, expect, Page } from '@playwright/test';

test.describe('Review & Rating System', () => {
  let page: Page;
  const testApiId = 'test-payment-api';
  
  test.beforeEach(async ({ page: p }) => {
    page = p;
    // Assume user is already logged in as a consumer with subscription
    await page.goto(`/api/${testApiId}`);
    await page.waitForLoadState('networkidle');
  });

  test.describe('Review Display', () => {
    test('should display review statistics', async () => {
      // Check review stats section
      const statsSection = page.locator('[data-testid="review-stats"]');
      await expect(statsSection).toBeVisible();
      
      // Verify average rating
      const avgRating = page.locator('[data-testid="average-rating"]');
      await expect(avgRating).toBeVisible();
      const rating = await avgRating.textContent();
      expect(parseFloat(rating || '0')).toBeGreaterThanOrEqual(0);
      expect(parseFloat(rating || '0')).toBeLessThanOrEqual(5);
      
      // Verify total review count
      const reviewCount = page.locator('[data-testid="total-reviews"]');
      await expect(reviewCount).toBeVisible();
    });

    test('should display rating distribution', async () => {
      // Check each star rating bar
      for (let i = 5; i >= 1; i--) {
        const ratingBar = page.locator(`[data-testid="rating-bar-${i}"]`);
        await expect(ratingBar).toBeVisible();
        
        const percentage = page.locator(`[data-testid="rating-percentage-${i}"]`);
        await expect(percentage).toBeVisible();
      }
    });

    test('should display individual reviews', async () => {
      // Verify review list
      const reviews = page.locator('[data-testid="review-item"]');
      const count = await reviews.count();
      expect(count).toBeGreaterThan(0);
      
      // Check first review structure
      const firstReview = reviews.first();
      await expect(firstReview.locator('[data-testid="review-rating"]')).toBeVisible();
      await expect(firstReview.locator('[data-testid="review-author"]')).toBeVisible();
      await expect(firstReview.locator('[data-testid="review-date"]')).toBeVisible();
      await expect(firstReview.locator('[data-testid="review-comment"]')).toBeVisible();
    });

    test('should show verified purchase badge', async () => {
      // Find reviews with verified badge
      const verifiedBadges = page.locator('[data-testid="verified-purchase-badge"]');
      const count = await verifiedBadges.count();
      expect(count).toBeGreaterThan(0);
    });

    test('should display creator responses', async () => {
      // Find a review with creator response
      const creatorResponses = page.locator('[data-testid="creator-response"]');
      if (await creatorResponses.count() > 0) {
        const response = creatorResponses.first();
        await expect(response).toBeVisible();
        await expect(response.locator('[data-testid="response-author"]')).toContainText('Creator');
      }
    });
  });

  test.describe('Review Submission', () => {
    test('should show review form only for subscribers', async () => {
      // Logged in subscriber should see form
      const reviewForm = page.locator('[data-testid="review-form"]');
      await expect(reviewForm).toBeVisible();
    });

    test('should not show review form for non-subscribers', async () => {
      // Navigate as non-subscriber
      await page.evaluate(() => {
        // Mock non-subscriber state
        localStorage.setItem('hasSubscription', 'false');
      });
      await page.reload();
      
      // Should see subscription prompt instead
      const subscribePrompt = page.locator('[data-testid="subscribe-to-review"]');
      await expect(subscribePrompt).toBeVisible();
    });

    test('should submit a new review', async () => {
      // Fill review form
      await page.click('[data-testid="star-rating-5"]');
      await page.fill('[data-testid="review-title"]', 'Excellent API!');
      await page.fill('[data-testid="review-comment"]', 'This API has great documentation and is very easy to integrate. The response times are fast and the pricing is fair.');
      
      // Submit review
      await page.click('[data-testid="submit-review"]');
      
      // Wait for success message
      await expect(page.locator('[data-testid="review-success"]')).toBeVisible();
      
      // Verify review appears in list
      await page.waitForTimeout(1000); // Wait for review to be added
      const newReview = page.locator('[data-testid="review-item"]').filter({
        hasText: 'Excellent API!'
      });
      await expect(newReview).toBeVisible();
    });

    test('should validate review form', async () => {
      // Try to submit without rating
      await page.click('[data-testid="submit-review"]');
      await expect(page.locator('text=Please select a rating')).toBeVisible();
      
      // Select rating but leave comment empty
      await page.click('[data-testid="star-rating-4"]');
      await page.click('[data-testid="submit-review"]');
      await expect(page.locator('text=Please write a review')).toBeVisible();
      
      // Add very short comment
      await page.fill('[data-testid="review-comment"]', 'Good');
      await page.click('[data-testid="submit-review"]');
      await expect(page.locator('text=Review must be at least 10 characters')).toBeVisible();
    });

    test('should enforce character limits', async () => {
      // Test title character limit
      const longTitle = 'a'.repeat(101);
      await page.fill('[data-testid="review-title"]', longTitle);
      const titleValue = await page.locator('[data-testid="review-title"]').inputValue();
      expect(titleValue.length).toBeLessThanOrEqual(100);
      
      // Test comment character limit
      const longComment = 'a'.repeat(1001);
      await page.fill('[data-testid="review-comment"]', longComment);
      const commentValue = await page.locator('[data-testid="review-comment"]').inputValue();
      expect(commentValue.length).toBeLessThanOrEqual(1000);
    });

    test('should show character count', async () => {
      await page.fill('[data-testid="review-comment"]', 'This is my review');
      const charCount = page.locator('[data-testid="char-count"]');
      await expect(charCount).toContainText('17 / 1000');
    });
  });

  test.describe('Review Interactions', () => {
    test('should vote review as helpful', async () => {
      const firstReview = page.locator('[data-testid="review-item"]').first();
      const helpfulButton = firstReview.locator('[data-testid="vote-helpful"]');
      
      // Get initial count
      const initialCount = await helpfulButton.locator('[data-testid="helpful-count"]').textContent();
      
      // Click helpful
      await helpfulButton.click();
      
      // Verify count increased
      await page.waitForTimeout(500);
      const newCount = await helpfulButton.locator('[data-testid="helpful-count"]').textContent();
      expect(parseInt(newCount || '0')).toBe(parseInt(initialCount || '0') + 1);
      
      // Button should be disabled after voting
      await expect(helpfulButton).toBeDisabled();
    });

    test('should vote review as not helpful', async () => {
      const firstReview = page.locator('[data-testid="review-item"]').first();
      const notHelpfulButton = firstReview.locator('[data-testid="vote-not-helpful"]');
      
      await notHelpfulButton.click();
      
      // Button should be disabled after voting
      await expect(notHelpfulButton).toBeDisabled();
    });

    test('should prevent duplicate voting', async () => {
      const firstReview = page.locator('[data-testid="review-item"]').first();
      const helpfulButton = firstReview.locator('[data-testid="vote-helpful"]');
      
      // Vote once
      await helpfulButton.click();
      
      // Try to vote again
      await expect(helpfulButton).toBeDisabled();
      
      // Reload page
      await page.reload();
      
      // Button should still be disabled
      await expect(helpfulButton).toBeDisabled();
    });
  });

  test.describe('Review Sorting', () => {
    test('should sort reviews by most recent', async () => {
      await page.selectOption('[data-testid="review-sort"]', 'recent');
      
      // Get dates of first two reviews
      const dates = await page.locator('[data-testid="review-date"]').allTextContents();
      const firstDate = new Date(dates[0]);
      const secondDate = new Date(dates[1]);
      
      // First should be more recent
      expect(firstDate.getTime()).toBeGreaterThanOrEqual(secondDate.getTime());
    });

    test('should sort reviews by most helpful', async () => {
      await page.selectOption('[data-testid="review-sort"]', 'helpful');
      
      // Get helpful counts
      const counts = await page.locator('[data-testid="helpful-count"]').allTextContents();
      const firstCount = parseInt(counts[0] || '0');
      const secondCount = parseInt(counts[1] || '0');
      
      // First should have more helpful votes
      expect(firstCount).toBeGreaterThanOrEqual(secondCount);
    });

    test('should sort reviews by highest rating', async () => {
      await page.selectOption('[data-testid="review-sort"]', 'highest');
      
      // Get ratings
      const ratings = await page.locator('[data-testid="review-rating"]').allTextContents();
      const firstRating = parseFloat(ratings[0]);
      const secondRating = parseFloat(ratings[1]);
      
      // First should have higher rating
      expect(firstRating).toBeGreaterThanOrEqual(secondRating);
    });

    test('should sort reviews by lowest rating', async () => {
      await page.selectOption('[data-testid="review-sort"]', 'lowest');
      
      // Get ratings
      const ratings = await page.locator('[data-testid="review-rating"]').allTextContents();
      const firstRating = parseFloat(ratings[0]);
      const secondRating = parseFloat(ratings[1]);
      
      // First should have lower rating
      expect(firstRating).toBeLessThanOrEqual(secondRating);
    });
  });

  test.describe('Review Pagination', () => {
    test('should paginate reviews', async () => {
      // Check if pagination exists
      const pagination = page.locator('[data-testid="review-pagination"]');
      
      if (await pagination.isVisible()) {
        // Click next page
        await page.click('[data-testid="review-next-page"]');
        
        // Verify new reviews loaded
        await page.waitForTimeout(500);
        const reviews = page.locator('[data-testid="review-item"]');
        await expect(reviews.first()).toBeVisible();
        
        // Go back to first page
        await page.click('[data-testid="review-prev-page"]');
      }
    });

    test('should show correct review count', async () => {
      const showingText = page.locator('[data-testid="showing-reviews"]');
      await expect(showingText).toContainText(/Showing \d+-\d+ of \d+ reviews/);
    });
  });

  test.describe('Creator Response Flow', () => {
    test('creator should be able to respond to reviews', async () => {
      // Switch to creator view
      await page.goto(`${process.env.CREATOR_PORTAL_URL}/apis/${testApiId}/reviews`);
      
      // Find a review without response
      const unrepliedReview = page.locator('[data-testid="review-item"]').filter({
        hasNot: page.locator('[data-testid="creator-response"]')
      }).first();
      
      // Click respond button
      await unrepliedReview.locator('[data-testid="respond-button"]').click();
      
      // Fill response
      await page.fill('[data-testid="response-text"]', 'Thank you for your feedback! We appreciate your review and are glad you find our API useful.');
      
      // Submit response
      await page.click('[data-testid="submit-response"]');
      
      // Verify response appears
      await expect(unrepliedReview.locator('[data-testid="creator-response"]')).toBeVisible();
    });

    test('should limit creator to one response per review', async () => {
      // Find a review with existing response
      const repliedReview = page.locator('[data-testid="review-item"]').filter({
        has: page.locator('[data-testid="creator-response"]')
      }).first();
      
      // Respond button should not be visible
      await expect(repliedReview.locator('[data-testid="respond-button"]')).not.toBeVisible();
    });
  });

  test.describe('Edge Cases', () => {
    test('should handle API with no reviews', async () => {
      await page.goto('/api/new-api-no-reviews');
      
      // Should show no reviews message
      await expect(page.locator('[data-testid="no-reviews"]')).toBeVisible();
      await expect(page.locator('text=Be the first to review')).toBeVisible();
      
      // Stats should show zeros
      await expect(page.locator('[data-testid="average-rating"]')).toContainText('0');
      await expect(page.locator('[data-testid="total-reviews"]')).toContainText('0');
    });

    test('should handle review submission errors', async () => {
      // Mock network error
      await page.route('**/api/v1/marketplace/apis/*/reviews', route => {
        route.abort('failed');
      });
      
      // Try to submit review
      await page.click('[data-testid="star-rating-5"]');
      await page.fill('[data-testid="review-comment"]', 'Test review');
      await page.click('[data-testid="submit-review"]');
      
      // Should show error message
      await expect(page.locator('[data-testid="review-error"]')).toBeVisible();
    });

    test('should handle concurrent review updates', async () => {
      // Open same API in two tabs
      const context = page.context();
      const page2 = await context.newPage();
      await page2.goto(`/api/${testApiId}`);
      
      // Submit review in first tab
      await page.click('[data-testid="star-rating-5"]');
      await page.fill('[data-testid="review-comment"]', 'Review from tab 1');
      await page.click('[data-testid="submit-review"]');
      
      // Refresh second tab
      await page2.reload();
      
      // New review should appear
      const newReview = page2.locator('[data-testid="review-item"]').filter({
        hasText: 'Review from tab 1'
      });
      await expect(newReview).toBeVisible();
      
      await page2.close();
    });
  });
});
