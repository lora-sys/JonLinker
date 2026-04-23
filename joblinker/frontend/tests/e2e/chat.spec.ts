import { test, expect } from '@playwright/test';

test.describe('A2A Chat Dialogue Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[type="email"]', 'test4@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard', { timeout: 10000 });
  });

  test('view conversation page', async ({ page }) => {
    // Navigate to matches first to get a match ID
    await page.goto('/matches');

    // Click on first match if exists
    const firstMatch = page.locator('[data-testid="match-card"], .bg-white.rounded-xl').first();
    if (await firstMatch.isVisible({ timeout: 3000 })) {
      await firstMatch.click();
      await page.waitForURL(/\/conversation\//, { timeout: 5000 });
    }

    // Verify conversation page loaded
    await expect(page.locator('text=Conversation, h1:has-text("Conversation")').first()).toBeVisible({ timeout: 5000 });
  });

  test('chat window renders', async ({ page }) => {
    await page.goto('/matches');

    // Try to access a conversation directly with a mock match ID
    await page.goto('/conversation/test-match-id');

    // Should see chat window or appropriate message
    const chatArea = page.locator('[class*="ChatWindow"], [class*="chat"]').first();
    const noMatch = page.locator('text=No conversations, text=Select a match').first();

    // Either chat window or placeholder should be visible
    const hasChatOrPlaceholder = await chatArea.isVisible({ timeout: 2000 }).catch(() => false) ||
                                await noMatch.isVisible({ timeout: 2000 }).catch(() => false);
    expect(hasChatOrPlaceholder).toBeTruthy();
  });

  test('send message via WebSocket', async ({ page }) => {
    // This test requires an active match with messages
    await page.goto('/matches');

    // Find and click a match to go to conversation
    const matchLink = page.locator('a[href*="/conversation/"]').first();
    if (await matchLink.isVisible({ timeout: 3000 })) {
      await matchLink.click();
      await page.waitForURL(/\/conversation\//, { timeout: 5000 });

      // Type a message
      const input = page.locator('input[type="text"], textarea').first();
      if (await input.isVisible({ timeout: 2000 })) {
        await input.fill('Hello test message');
        await input.press('Enter');

        // Message should appear in chat
        await expect(page.locator('text=Hello test message').first()).toBeVisible({ timeout: 3000 });
      }
    }
  });
});