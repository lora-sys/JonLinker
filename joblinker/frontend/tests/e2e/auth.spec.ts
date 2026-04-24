import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test('register new user', async ({ page }) => {
    const timestamp = Date.now();
    const email = `newuser${timestamp}@example.com`;
    const password = 'password123';

    await page.goto('/register');
    await page.locator('input[type="email"]').fill(email);
    await page.locator('input[type="password"]').nth(0).fill(password);
    await page.locator('input[type="password"]').nth(1).fill(password);
    await page.locator('button:has-text("Job Seeker")').click();
    await page.locator('button[type="submit"]').click();

    // Wait for redirect to dashboard
    await page.waitForURL('**/dashboard', { timeout: 15000 });

    // Verify we're on the dashboard (URL is the key indicator)
    expect(page.url()).toContain('/dashboard');
  });

  test('login page loads', async ({ page }) => {
    await page.goto('/login');
    await expect(page.locator('input[type="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });
});
