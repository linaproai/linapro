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

  test('TC0233b: access token 过期时先刷新并重试原请求', async ({
    page,
  }) => {
    let refreshRequestCount = 0;
    let userInfoRequestCount = 0;
    let userInfoRetryAuthorization = '';

    await page.route('**/api/v1/auth/logout', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        status: 500,
        body: JSON.stringify({
          code: 500,
          message: 'logout should not be called',
        }),
      });
    });
    await page.route('**/api/v1/auth/refresh', async (route) => {
      refreshRequestCount += 1;
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          data: {
            accessToken: 'new-access-token',
            refreshToken: 'new-refresh-token',
          },
          message: 'ok',
        }),
      });
    });
    await page.route('**/api/v1/user/info', async (route) => {
      userInfoRequestCount += 1;
      if (userInfoRequestCount === 1) {
        await route.fulfill({
          contentType: 'application/json',
          status: 401,
          body: JSON.stringify({
            code: 401,
            message: 'Unauthorized',
          }),
        });
        return;
      }
      userInfoRetryAuthorization =
        route.request().headers().authorization ?? '';
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          data: {
            avatar: '',
            homePath: '/dashboard/analytics',
            permissions: ['*'],
            realName: 'Admin',
            roles: ['admin'],
            userId: 1,
            username: 'admin',
          },
          message: 'ok',
        }),
      });
    });
    await page.route('**/api/v1/menus/all', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          data: {
            list: [
              {
                children: [
                  {
                    component: 'dashboard/analytics/index',
                    id: 2,
                    meta: {
                      authority: ['*'],
                      icon: 'lucide:area-chart',
                      title: 'Analytics',
                    },
                    name: 'DashboardAnalytics',
                    parentId: 1,
                    path: 'analytics',
                    type: 'M',
                  },
                ],
                component: 'BasicLayout',
                id: 1,
                meta: {
                  authority: ['*'],
                  icon: 'lucide:layout-dashboard',
                  title: 'Dashboard',
                },
                name: 'Dashboard',
                parentId: 0,
                path: '/dashboard',
                type: 'D',
              },
            ],
          },
          message: 'ok',
        }),
      });
    });
    await page.route('**/api/v1/user/message/count', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          data: 0,
          message: 'ok',
        }),
      });
    });
    await page.route('**/api/v1/user/message?*', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          data: {
            list: [],
            pageNum: 1,
            pageSize: 20,
            total: 0,
          },
          message: 'ok',
        }),
      });
    });
    await page.route('**/api/v1/platform/tenants?*', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          data: {
            list: [],
            pageNum: 1,
            pageSize: 100,
            total: 0,
          },
          message: 'ok',
        }),
      });
    });

    await page.addInitScript(() => {
      localStorage.clear();
      const expiredAccessState = JSON.stringify({
        accessCodes: ['*'],
        accessToken: 'expired-access-token',
        isLockScreen: false,
        refreshToken: 'stored-refresh-token',
      });
      for (const envName of ['dev', 'prod']) {
        localStorage.setItem(
          `lina-web-antd-5.6.0-${envName}-core-access`,
          expiredAccessState,
        );
      }
    });

    await page.goto('/dashboard/analytics');
    await expect
      .poll(() => refreshRequestCount, { timeout: 15_000 })
      .toBeGreaterThanOrEqual(1);
    await expect
      .poll(() => userInfoRequestCount, { timeout: 15_000 })
      .toBeGreaterThanOrEqual(2);
    await expect(page).not.toHaveURL(/auth\/login/);
    expect(userInfoRetryAuthorization).toBe('Bearer new-access-token');
  });
});
