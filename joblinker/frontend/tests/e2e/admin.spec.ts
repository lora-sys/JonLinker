import { test, expect } from '@playwright/test';

test.describe('Admin Dashboard Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[type="email"]', 'test4@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard', { timeout: 10000 });
  });

  test('view admin dashboard', async ({ page }) => {
    await page.goto('/admin');

    // Should see admin dashboard with metrics
    await expect(page.locator('text=Admin Dashboard, h1:has-text("Admin")').first()).toBeVisible({ timeout: 5000 });
  });

  test('metrics panel renders', async ({ page }) => {
    await page.goto('/admin');

    // Should see metrics cards
    const metricsPanel = page.locator('[class*="grid"], .gap-4').first();
    await expect(metricsPanel).toBeVisible({ timeout: 5000 });

    // Should see metric labels
    const totalAgents = page.locator('text=Total Agents').first();
    const activeJobs = page.locator('text=Active Jobs').first();
    await expect(totalAgents.or(activeJobs).first()).toBeVisible({ timeout: 3000 });
  });

  test('job management section', async ({ page }) => {
    await page.goto('/admin');

    // Look for job management section
    const jobSection = page.locator('text=Job Management, h3:has-text("Job")').first();
    await expect(jobSection).toBeVisible({ timeout: 3000 });
  });

  test('agent management section', async ({ page }) => {
    await page.goto('/admin');

    // Look for agent management section
    const agentSection = page.locator('text=Agent Management, h3:has-text("Agent")').first();
    await expect(agentSection).toBeVisible({ timeout: 3000 });
  });
});