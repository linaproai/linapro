import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  testMatch: /TC\d{4}.*\.ts$/,
  fullyParallel: false,
  globalSetup: './global-setup.ts',
  workers: Number.parseInt(process.env.E2E_WORKERS ?? '1', 10),
  retries: 0,
  timeout: 60000,
  expect: {
    timeout: 10000,
  },
  use: {
    baseURL: process.env.E2E_BASE_URL ?? 'http://127.0.0.1:5666',
    headless: true,
    locale: 'zh-CN',
    screenshot: 'only-on-failure',
    trace: 'on-first-retry',
  },
  reporter: [['list'], ['html', { open: 'never' }]],
  projects: [
    {
      name: 'chromium',
      use: { browserName: 'chromium' },
    },
  ],
});
