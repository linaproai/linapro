import type { Page } from '@playwright/test';

import { test, expect } from '../../fixtures/auth';

function countAccessReloadRequests(page: Page) {
  const counter = {
    menuRequests: 0,
    userInfoRequests: 0,
  };
  page.on('request', (request) => {
    if (
      request.method() === 'GET' &&
      request.url().includes('/api/v1/user/info')
    ) {
      counter.userInfoRequests += 1;
    }
    if (
      request.method() === 'GET' &&
      request.url().includes('/api/v1/menus/all')
    ) {
      counter.menuRequests += 1;
    }
  });
  return counter;
}

test.describe('TC-152 Language switch without access reload', () => {
  test('TC-152a: switching locale updates menu text without reloading user or menu access APIs', async ({
    adminPage,
    mainLayout,
  }) => {
    const counter = countAccessReloadRequests(adminPage);

    await mainLayout.switchLanguage('English');
    await expect(
      adminPage.getByText('Settings', { exact: true }).first(),
    ).toBeVisible();
    await mainLayout.switchLanguage('简体中文');
    await expect(
      adminPage.getByText('系统设置', { exact: true }).first(),
    ).toBeVisible();

    expect(counter.userInfoRequests).toBe(0);
    expect(counter.menuRequests).toBe(0);
  });

  test('TC-152b: preferences drawer locale switch avoids full reload and access APIs', async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage('简体中文');
    const counter = countAccessReloadRequests(adminPage);

    await mainLayout.switchLanguageFromPreferences('English');
    await expect(mainLayout.preferencesDrawerTitle).toHaveText('Preferences');
    await expect(
      adminPage.getByText('Settings', { exact: true }).first(),
    ).toBeVisible();

    await mainLayout.switchLanguageFromPreferences('简体中文');
    await expect(mainLayout.preferencesDrawerTitle).toHaveText('偏好设置');
    await expect(
      adminPage.getByText('系统设置', { exact: true }).first(),
    ).toBeVisible();

    expect(counter.userInfoRequests).toBe(0);
    expect(counter.menuRequests).toBe(0);
  });
});
