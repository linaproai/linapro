import { test, expect } from '../../fixtures/auth';

test.describe('TC0233 过期会话登出保护', () => {
  test('TC0233a: 旧 token 访问工作台不会触发 logout 401 风暴', async ({
    page,
  }) => {
    let logoutRequestCount = 0;

    await page.route('**/api/v1/auth/logout', async (route) => {
      logoutRequestCount += 1;
      await route.fulfill({
        contentType: 'application/json',
        status: 401,
        body: JSON.stringify({
          code: 401,
          message: 'Unauthorized',
        }),
      });
    });

    await page.addInitScript(() => {
      localStorage.clear();
      const expiredAccessState = JSON.stringify({
        accessCodes: ['*'],
        accessToken: 'expired-token-from-previous-session',
        isLockScreen: false,
        refreshToken: null,
      });
      for (const envName of ['dev', 'prod']) {
        localStorage.setItem(
          `lina-web-antd-5.6.0-${envName}-core-access`,
          expiredAccessState,
        );
      }
    });

    await page.goto('/dashboard/analytics');
    await page.waitForURL(/auth\/login/, { timeout: 15_000 });
    await page.waitForTimeout(1000);

    expect(page.url()).toContain('/auth/login');
    expect(logoutRequestCount).toBe(0);
  });
});
