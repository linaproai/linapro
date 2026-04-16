import { test as base, type Page } from '@playwright/test';

import { LoginPage } from '../pages/LoginPage';
import { MainLayout } from '../pages/MainLayout';
import { config } from './config';

export type AuthFixtures = {
  adminPage: Page;
  loginPage: LoginPage;
  mainLayout: MainLayout;
};

export const test = base.extend<AuthFixtures>({
  adminPage: async ({ page }, use) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(
      config.adminUser,
      config.adminPass,
    );
    await use(page);
  },
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page));
  },
  mainLayout: async ({ page }, use) => {
    await use(new MainLayout(page));
  },
});

export { expect } from '@playwright/test';
