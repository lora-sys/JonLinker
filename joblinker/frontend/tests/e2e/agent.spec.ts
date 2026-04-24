import { test, expect } from '@playwright/test';

test.describe('Agent Creation Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.locator('input[type="email"]').fill('test@example.com');
    await page.locator('input[type="password"]').fill('password123');
    await page.locator('button[type="submit"]').click();
    await page.waitForURL('**/dashboard', { timeout: 15000 });
  });

  test('create seeker agent', async ({ page }) => {
    await page.goto('/agents/create');

    // Select seeker type - it's selected by default
    // Click create button directly since seeker is default
    await page.locator('button[type="submit"]').click();

    // Should redirect to dashboard after creation
    await page.waitForURL('**/dashboard', { timeout: 10000 });
  });

  test('create recruiter agent', async ({ page }) => {
    await page.goto('/agents/create');

    // Select recruiter type
    await page.locator('button:has-text("Recruiter")').click();

    await page.locator('button[type="submit"]').click();

    // Should redirect to dashboard after creation
    await page.waitForURL('**/dashboard', { timeout: 10000 });
  });

  test('view agents list', async ({ page }) => {
    await page.goto('/agents');

    // Should show agents page
    await expect(page.locator('h1:has-text("Agents")')).toBeVisible({ timeout: 5000 });
  });
});
