import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { LoginPage } from '@host-tests/pages/LoginPage';
import { config } from '@host-tests/fixtures/config';
import { waitForRouteReady } from '@host-tests/support/ui';

test.describe('TC0030 登录日志自动记录', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'monitor-loginlog');
  });

  test('TC0030a: 登录成功后登录日志中记录成功日志', async ({ adminPage }) => {
    // The adminPage fixture already logged in, so a login log should exist
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/loginlog') &&
        res.request().method() === 'GET' &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.goto('/monitor/loginlog');
    await responsePromise;
    await waitForRouteReady(adminPage);

    // Should see at least one successful admin login row without depending on
    // global log ordering from other tests.
    const rows = adminPage.locator('.vxe-table--main-wrapper .vxe-body--row');
    await expect(rows.first()).toBeVisible();
    const adminSuccessRow = rows.filter({ hasText: 'admin' }).filter({
      hasText: /成功|Success/i,
    }).first();
    await expect(adminSuccessRow).toBeVisible({ timeout: 10_000 });
  });

  test('TC0030b: 登录失败后登录日志中记录失败日志', async ({ browser }) => {
    // The adminPage fixture used in beforeEach authenticates the default page.
    // Use a fresh browser context here so the failed-login path starts truly
    // unauthenticated and can reach /auth/login without redirecting away.
    const context = await browser.newContext({ baseURL: config.baseURL });
    const page = await context.newPage();
    const loginPage = new LoginPage(page);

    try {
      await loginPage.goto();
      await loginPage.login('admin', 'wrongpassword');
      await expect(loginPage.errorMessage).toBeVisible();

      // Now login correctly in the same fresh context.
      await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);

      // Navigate to login log page.
      const responsePromise = page.waitForResponse(
        (res) =>
          res.url().includes('/api/v1/loginlog') &&
          res.request().method() === 'GET' &&
          res.status() === 200,
        { timeout: 15000 },
      );
      await page.goto('/monitor/loginlog');
      await responsePromise;
      await waitForRouteReady(page);

      // Should see login logs including failure-state rows.
      const rows = page.locator('.vxe-body--row');
      await expect(rows.first()).toBeVisible();
      const failedRow = rows
        .filter({ hasText: /失败|登录失败|用户名或密码错误/ })
        .first();
      await expect(failedRow).toBeVisible({ timeout: 10_000 });
      await expect(failedRow).toContainText('admin');
    } finally {
      await context.close();
    }
  });
});
