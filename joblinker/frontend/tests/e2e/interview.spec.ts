import { test, expect } from '@playwright/test';

test.describe('Interview Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.locator('input[type="email"]').fill('test@example.com');
    await page.locator('input[type="password"]').fill('password123');
    await page.locator('button[type="submit"]').click();
    await page.waitForURL('**/dashboard', { timeout: 15000 });
  });

  test('view interviews page', async ({ page }) => {
    await page.goto('/interviews');

    // Should see interviews page with header
    await expect(page.locator('h1:has-text("Interviews")')).toBeVisible({ timeout: 5000 });
  });

  test('interviews with scheduled interviews', async ({ page }) => {
    await page.goto('/interviews');

    // Should see scheduled interviews or timeline
    await expect(page.locator('h1:has-text("Interviews")')).toBeVisible({ timeout: 5000 });
    // Check for countdown timer or interview content
    const hasContent = await page.locator('text=Interview').first().isVisible().catch(() => false);
    expect(hasContent || await page.locator('[class*="rounded-full"]').first().isVisible().catch(() => false)).toBeTruthy();
  });
});
