import { test, expect } from '@playwright/test';

test.describe('Offer Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[type="email"]', 'test4@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard', { timeout: 10000 });
  });

  test('view offers page', async ({ page }) => {
    await page.goto('/offers');

    // Should see offers page with header
    await expect(page.locator('h1:has-text("Offers")')).toBeVisible({ timeout: 5000 });
  });

  test('view offer card', async ({ page }) => {
    await page.goto('/offers');

    // Should see empty state or offer cards
    const emptyState = page.locator('text=No offers yet, text=Offers will appear').first();
    const offerCard = page.locator('[class*="OfferCard"], .rounded-xl.bg-white').first();

    const hasOffers = await offerCard.isVisible({ timeout: 2000 }).catch(() => false);
    const hasEmptyState = await emptyState.isVisible({ timeout: 2000 }).catch(() => false);

    // Either should have offers or empty state
    expect(hasOffers || hasEmptyState).toBeTruthy();
  });

  test('offer accept/decline buttons', async ({ page }) => {
    await page.goto('/offers');

    // Look for action buttons
    const acceptBtn = page.locator('button:has-text("Accept")').first();
    const declineBtn = page.locator('button:has-text("Decline")').first();

    // Buttons might not be visible if no pending offers
    const hasButtons = await acceptBtn.isVisible({ timeout: 2000 }).catch(() => false) ||
                       await declineBtn.isVisible({ timeout: 2000 }).catch(() => false);

    // This is acceptable - no pending offers
    if (!hasButtons) {
      await expect(page.locator('text=No offers yet, text=pending').first()).toBeVisible();
    }
  });
});