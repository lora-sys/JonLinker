import { test, expect } from '@playwright/test';

test.describe('Agent Creation Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('input[type="email"]', 'test4@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard', { timeout: 10000 });
  });

  test('create seeker agent', async ({ page }) => {
    await page.goto('/agents/create');

    // Fill agent creation form
    await page.fill('input[name="name"]', 'Test Seeker Agent');

    // Select seeker type if needed
    const seekerBtn = page.locator('button:has-text("Seeker")');
    if (await seekerBtn.isVisible()) {
      await seekerBtn.click();
    }

    // Submit
    await page.click('button[type="submit"]');

    // Should see success or redirect
    await page.waitForURL(/\/agents/, { timeout: 5000 });
  });

  test('view agents list', async ({ page }) => {
    await page.goto('/agents');

    // Should see agents page
    await expect(page.locator('text=My Agents')).toBeVisible({ timeout: 5000 });
  });
});