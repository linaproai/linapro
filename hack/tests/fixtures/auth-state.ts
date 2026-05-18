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

  const browser = await chromium.launch({
    channel: process.env.E2E_BROWSER_CHANNEL?.trim() || undefined,
  });
  const context = await browser.newContext({ baseURL });
  const page = await context.newPage();
  const loginPage = new LoginPage(page);

  try {
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    await page.waitForLoadState('networkidle').catch(() => {});
    await page.evaluate(() => {
      const localeKey = Object.keys(localStorage).find((key) =>
        key.endsWith('preferences-locale'),
      );
      if (localeKey) {
        localStorage.setItem(localeKey, JSON.stringify({ value: 'zh-CN' }));
      }

      const preferencesKey = Object.keys(localStorage).find((key) =>
        key.endsWith('preferences'),
      );
      if (!preferencesKey) {
        return;
      }
      const preferences = JSON.parse(localStorage.getItem(preferencesKey) || '{}');
      if (preferences.value?.app) {
        preferences.value.app.locale = 'zh-CN';
        localStorage.setItem(preferencesKey, JSON.stringify(preferences));
      }
    });
    await context.storageState({ path: adminStorageStatePath });
  } finally {
    await context.close();
    await browser.close();
  }
}
