import { mkdirSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { chromium } from '@playwright/test';

import { LoginPage } from '../pages/LoginPage';
import { config } from './config';

const fixtureDir = path.dirname(fileURLToPath(import.meta.url));
const authStateDir = path.resolve(fixtureDir, '../temp/storage-state');

export const adminStorageStatePath = path.join(authStateDir, 'admin.json');

export async function writeAdminStorageState(baseURL = config.baseURL) {
  mkdirSync(authStateDir, { recursive: true });

  const browser = await chromium.launch();
  const context = await browser.newContext({ baseURL });
  const page = await context.newPage();
  const loginPage = new LoginPage(page);

  try {
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    await page.waitForLoadState('networkidle').catch(() => {});
    await context.storageState({ path: adminStorageStatePath });
  } finally {
    await context.close();
    await browser.close();
  }
}
