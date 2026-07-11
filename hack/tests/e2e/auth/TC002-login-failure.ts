import { LoginPage } from '../../pages/LoginPage';
import { config } from '../../fixtures/config';
import { test, expect } from '../../fixtures/auth';
import { waitForRouteReady } from '../../support/ui';

test.describe('TC002 登录失败', () => {
  test('TC002a: 错误密码显示本地化错误提示', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await waitForRouteReady(page);
    if ((await page.locator('html').getAttribute('lang')) !== 'zh-CN') {
      await loginPage.switchLanguage('简体中文');
    }

    await loginPage.login(config.adminUser, 'wrongpassword');

    // Runtime i18n key error.auth.invalid.credentials (not English fallback).
    await expect(
      page.getByText('用户名或密码错误', { exact: true }),
    ).toBeVisible({ timeout: 5000 });
    await expect(
      page.getByText('Invalid username or password', { exact: true }),
    ).toHaveCount(0);
  });

  test('TC002b: 错误密码后仍在登录页', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login(config.adminUser, 'wrongpassword');

    await expect(page).toHaveURL(/\/auth\/login/, { timeout: 5000 });
  });
});
