import { test, expect } from '@playwright/test';

test.describe('Admin Dashboard Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.locator('input[type="email"]').fill('test@example.com');
    await page.locator('input[type="password"]').fill('password123');
    await page.locator('button[type="submit"]').click();
    await page.waitForURL('**/dashboard', { timeout: 15000 });
  });

  test('view admin dashboard', async ({ page }) => {
    await page.goto('/admin');

    // Should see admin dashboard header
    await expect(page.locator('h1:has-text("Admin Dashboard")')).toBeVisible({ timeout: 5000 });
  });

  test('metrics panel renders', async ({ page }) => {
    await page.goto('/admin');

    // Should see overview section with metrics
    await expect(page.locator('text=Overview')).toBeVisible({ timeout: 5000 });
  });

  test('job management section', async ({ page }) => {
    await page.goto('/admin');

    // Look for Jobs section
    await expect(page.locator('h2:has-text("Jobs")')).toBeVisible({ timeout: 3000 });
  });

  test('agent management section', async ({ page }) => {
    await page.goto('/admin');

    // Look for Agents section
    await expect(page.locator('h2:has-text("Agents")')).toBeVisible({ timeout: 3000 });
  });
});
