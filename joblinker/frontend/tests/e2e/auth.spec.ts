import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test('register and login flow', async ({ page }) => {
    const timestamp = Date.now();
    const email = `test${timestamp}@example.com`;
    const password = 'password123';

    // Register
    await page.goto('/register');
    await page.fill('input[type="email"]', email);
    await page.fill('input[type="password"]', password);
    await page.fill('input[placeholder="••••••••"]:nth-of-type(2)', password);

    // Select seeker role
    await page.click('button:has-text("Job Seeker")');

    await page.click('button[type="submit"]');

    // Should redirect to dashboard
    await page.waitForURL('**/dashboard', { timeout: 10000 });

    // Verify user is logged in
    await expect(page.locator('text=Dashboard')).toBeVisible();
  });

  test('login with existing user', async ({ page }) => {
    // Use the test user from before
    await page.goto('/login');
    await page.fill('input[type="email"]', 'test4@example.com');
    await page.fill('input[type="password"]', 'password123');

    await page.click('button[type="submit"]');

    // Should redirect to dashboard
    await page.waitForURL('**/dashboard', { timeout: 10000 });
  });

  test('logout flow', async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('input[type="email"]', 'test4@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard', { timeout: 10000 });

    // Click logout (in header)
    const logoutBtn = page.locator('button:has-text("Logout"), button:has-text("Sign Out")');
    if (await logoutBtn.isVisible()) {
      await logoutBtn.click();
    }

    // Should redirect to home
    await page.waitForURL('**/', { timeout: 5000 });
  });
});