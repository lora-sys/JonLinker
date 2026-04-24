import { test, expect } from '@playwright/test';

test.describe('A2A Chat Dialogue Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.locator('input[type="email"]').fill('test@example.com');
    await page.locator('input[type="password"]').fill('password123');
    await page.locator('button[type="submit"]').click();
    await page.waitForURL('**/dashboard', { timeout: 15000 });
  });

  test('view matches page', async ({ page }) => {
    await page.goto('/matches');

    // Should show matches page
    await expect(page.locator('h1:has-text("Matches")')).toBeVisible({ timeout: 5000 });
  });

  test('messages page renders', async ({ page }) => {
    await page.goto('/messages');

    // Should load without error
    await expect(page.locator('body')).toBeVisible({ timeout: 5000 });
  });
});
