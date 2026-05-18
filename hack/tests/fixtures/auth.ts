import { test as base, type BrowserContext, type Page } from '@playwright/test';

import { adminStorageStatePath } from './auth-state';
import { LoginPage } from '../pages/LoginPage';
import { MainLayout } from '../pages/MainLayout';
import { config } from './config';
import { waitForRouteReady } from '../support/ui';

export type AuthFixtures = {
  adminContext: BrowserContext;
  adminPage: Page;
  loginPage: LoginPage;
  mainLayout: MainLayout;
};

export const test = base.extend<AuthFixtures>({
  adminContext: async ({ browser }, use) => {
    const context = await browser.newContext({
      baseURL: config.baseURL,
      storageState: adminStorageStatePath,
    });
    await use(context);
    await context.close();
  },
  adminPage: async ({ adminContext }, use) => {
    const page = await adminContext.newPage();
    await page.goto('/dashboard/analytics', { waitUntil: 'domcontentloaded' });
    await waitForRouteReady(page, 15000);
    await use(page);
    await page.close();
  },
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page));
  },
  mainLayout: async ({ adminPage }, use) => {
    await use(new MainLayout(adminPage));
  },
});

export { expect } from '@playwright/test';
