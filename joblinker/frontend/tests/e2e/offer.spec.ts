import { test, expect } from '@playwright/test';

test.describe('Offer Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.locator('input[type="email"]').fill('test@example.com');
    await page.locator('input[type="password"]').fill('password123');
    await page.locator('button[type="submit"]').click();
    await page.waitForURL('**/dashboard', { timeout: 15000 });
  });

  test('view offers page', async ({ page }) => {
    await page.goto('/offers');

    // Should see offers page with header
    await expect(page.locator('h1:has-text("Offers")')).toBeVisible({ timeout: 5000 });
  });

  test('offers page with mock offers', async ({ page }) => {
    await page.goto('/offers');

    // Should see offers page header
    await expect(page.locator('h1:has-text("Offers")')).toBeVisible({ timeout: 5000 });
    // Check for compensation card or accept/negotiate buttons
    const hasButtons = await page.locator('button:has-text("Accept")').isVisible().catch(() => false);
    expect(hasButtons || await page.locator('text=TechCorp').isVisible().catch(() => false)).toBeTruthy();
  });
});
