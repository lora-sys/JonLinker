import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests/e2e',
  timeout: 30000,
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'list',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    headless: true,
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'], headless: true },
    },
  ],
  webServer: [
    {
      command: 'cd ../backend && ./server',
      port: 8080,
      reuse: true,
      timeout: 30000,
      reuseExistingServer: true,
    },
    {
      command: 'npm run dev',
      port: 3000,
      reuse: true,
      timeout: 30000,
      reuseExistingServer: true,
    },
  ],
});