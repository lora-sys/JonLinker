import { test, expect } from '@playwright/test';

test.describe('Matching Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('input[type="email"]', 'test4@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard', { timeout: 10000 });
  });

  test('view matches page', async ({ page }) => {
    await page.goto('/matches');

    // Should see matches or empty state
    const heading = page.locator('h1, h2').first();
    await expect(heading).toBeVisible({ timeout: 5000 });
  });

  test('view jobs page', async ({ page }) => {
    await page.goto('/jobs');

    // Should see jobs page
    const heading = page.locator('h1, h2').first();
    await expect(heading).toBeVisible({ timeout: 5000 });
  });
});