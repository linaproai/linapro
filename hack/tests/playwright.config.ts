import { defineConfig } from '@playwright/test';

const browserChannel = process.env.E2E_BROWSER_CHANNEL?.trim() || undefined;

export default defineConfig({
  // The suite root is the repository root so Playwright can execute host tests
  // and source-plugin-owned tests under apps/lina-plugins/<plugin-id>/hack/tests/e2e.
  testDir: '../..',
  // Anchor TypeScript path resolution (incl. `@host-tests/*`) to this tsconfig
  // so plugin-owned tests can import shared host assets via aliases instead of
  // 6–7 level relative paths.
  tsconfig: './tsconfig.json',
  testMatch: [
    /hack[\\/]tests[\\/]e2e[\\/](?:.*[\\/])?TC\d{4}-[^\\.\\/]+\.ts$/,
    /apps[\\/]lina-plugins[\\/][^\\/]+[\\/]hack[\\/]tests[\\/]e2e[\\/](?:.*[\\/])?TC\d{4}-[^\\.\\/]+\.ts$/,
  ],
  fullyParallel: false,
  globalSetup: './global-setup.ts',
  workers: Number.parseInt(process.env.E2E_WORKERS ?? '1', 10),
  retries: Number.parseInt(process.env.E2E_RETRIES ?? (process.env.CI ? '1' : '0'), 10),
  timeout: 60000,
  expect: {
    timeout: 10000,
  },
  use: {
    baseURL: process.env.E2E_BASE_URL ?? 'http://127.0.0.1:5666',
    headless: true,
    locale: 'zh-CN',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
    video: 'retain-on-failure',
  },
  reporter: [['list'], ['html', { open: 'never' }]],
  projects: [
    {
      name: 'chromium',
      use: { browserName: 'chromium', channel: browserChannel },
    },
  ],
});
