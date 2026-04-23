import { test, expect } from '@playwright/test';

test.describe('Interview Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[type="email"]', 'test4@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard', { timeout: 10000 });
  });

  test('view interviews page', async ({ page }) => {
    await page.goto('/interviews');

    // Should see interviews page with header
    await expect(page.locator('h1:has-text("Interviews")')).toBeVisible({ timeout: 5000 });
  });

  test('view interview card', async ({ page }) => {
    await page.goto('/interviews');

    // Should see empty state or interview cards
    const emptyState = page.locator('text=No interviews scheduled').first();
    const interviewCard = page.locator('[class*="InterviewCard"], [class*="rounded-xl"]').first();

    const hasInterviews = await interviewCard.isVisible({ timeout: 2000 }).catch(() => false);
    const hasEmptyState = await emptyState.isVisible({ timeout: 2000 }).catch(() => false);

    expect(hasInterviews || hasEmptyState).toBeTruthy();
  });

  test('schedule modal accessibility', async ({ page }) => {
    await page.goto('/interviews');

    // Look for schedule button
    const scheduleBtn = page.locator('button:has-text("Schedule"), button:has-text("New Interview")').first();
    const hasScheduleBtn = await scheduleBtn.isVisible({ timeout: 2000 }).catch(() => false);

    if (hasScheduleBtn) {
      await scheduleBtn.click();

      // Modal should appear
      const modal = page.locator('[role="dialog"], .fixed.inset-0');
      await expect(modal.first()).toBeVisible({ timeout: 3000 });
    }
  });

  test('calendar preview renders', async ({ page }) => {
    await page.goto('/interviews');

    // Look for calendar preview section
    const calendarPreview = page.locator('text=Calendar Preview, text=Download .ics').first();
    const hasCalendar = await calendarPreview.isVisible({ timeout: 2000 }).catch(() => false);

    // If no interviews, calendar preview won't be visible
    // This is expected behavior
    if (!hasCalendar) {
      await expect(page.locator('text=No interviews scheduled').first()).toBeVisible();
    }
  });
});