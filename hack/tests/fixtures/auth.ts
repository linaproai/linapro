import { test as base, type BrowserContext, type Page } from '@playwright/test';

import { adminStorageStatePath } from './auth-state';
import { LoginPage } from '../pages/LoginPage';
import { MainLayout } from '../pages/MainLayout';
import { config, isWorkspaceManagedPath, workspacePath } from './config';
import { waitForRouteReady } from '../support/ui';

export type AuthFixtures = {
  adminContext: BrowserContext;
  adminPage: Page;
  authenticatedPage: Page;
  loginPage: LoginPage;
  mainLayout: MainLayout;
};

const workspaceGotoInstalled = Symbol('workspaceGotoInstalled');

type WorkspaceAwarePage = Page & {
  [workspaceGotoInstalled]?: boolean;
};

function shouldPrefixWorkspacePath(url: Parameters<Page['goto']>[0]) {
  if (typeof url !== 'string' || !url.startsWith('/')) {
    return false;
  }
  return isWorkspaceManagedPath(url);
}

async function installWorkspaceGoto(page: Page) {
  const workspaceAwarePage = page as WorkspaceAwarePage;
  if (workspaceAwarePage[workspaceGotoInstalled]) {
    return;
  }
  const originalGoto = workspaceAwarePage.goto.bind(page);
  workspaceAwarePage.goto = ((url, options) => {
    const nextURL = shouldPrefixWorkspacePath(url)
      ? workspacePath(url as string)
      : url;
    return originalGoto(nextURL, options);
  }) as Page['goto'];
  workspaceAwarePage[workspaceGotoInstalled] = true;
}

export const test = base.extend<AuthFixtures>({
  page: async ({ page }, use) => {
    await installWorkspaceGoto(page);
    await use(page);
  },
  adminContext: async ({ browser }, use) => {
    const context = await browser.newContext({
      baseURL: config.baseURL,
      storageState: adminStorageStatePath,
    });
    await use(context);
    await context.close();
  },
  authenticatedPage: async ({ adminContext }, use) => {
    const page = await adminContext.newPage();
    await installWorkspaceGoto(page);
    await use(page);
    await page.close();
  },
  adminPage: async ({ adminContext }, use) => {
    const page = await adminContext.newPage();
    await installWorkspaceGoto(page);
    await page.goto(workspacePath('/dashboard/analytics'), { waitUntil: 'domcontentloaded' });
    await waitForRouteReady(page, 15000);
    await use(page);
    await page.close();
  },
  loginPage: async ({ page }, use) => {
    await installWorkspaceGoto(page);
    await use(new LoginPage(page));
  },
  mainLayout: async ({ adminPage }, use) => {
    await use(new MainLayout(adminPage));
  },
});

export { expect } from '@playwright/test';
export type { Locator, Page, Route } from '@playwright/test';
