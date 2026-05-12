import { test, expect } from '../../fixtures/auth';

test.describe('TC0234 Token 自动刷新机制', () => {
  test('TC0234a: 单请求 401 自动刷新成功重放', async ({
    page,
  }) => {
    let refreshRequestCount = 0;
    let apiRequestCount = 0;
    let apiRetryAuthorization = '';

    // Mock refresh token 接口 - 第一次调用返回新的 token
    await page.route('**/api/v1/auth/refresh', async (route) => {
      refreshRequestCount += 1;
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          data: {
            accessToken: 'new-access-token-after-refresh',
            refreshToken: 'new-refresh-token',
          },
          message: 'ok',
        }),
      });
    });

    // Mock 业务 API - 第一次返回 401，刷新后第二次返回成功
    await page.route('**/api/v1/user/info', async (route) => {
      apiRequestCount += 1;
      if (apiRequestCount === 1) {
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
      // 记录重试请求使用的 Authorization header
      apiRetryAuthorization =
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

    // 设置过期的 token
    await page.addInitScript(() => {
      localStorage.clear();
      const expiredAccessState = JSON.stringify({
        accessCodes: ['*'],
        accessToken: 'expired-access-token',
        isLockScreen: false,
        refreshToken: 'valid-refresh-token',
      });
      for (const envName of ['dev', 'prod']) {
        localStorage.setItem(
          `lina-web-antd-5.6.0-${envName}-core-access`,
          expiredAccessState,
        );
      }
    });

    await page.goto('/dashboard/analytics');
    
    // 验证 refresh 接口只被调用了一次
    await expect
      .poll(() => refreshRequestCount, { timeout: 15_000 })
      .toBe(1);
    
    // 验证 API 被调用了至少两次（第一次 401，第二次重试成功）
    await expect
      .poll(() => apiRequestCount, { timeout: 15_000 })
      .toBeGreaterThanOrEqual(2);
    
    // 验证重试请求使用了新的 token
    expect(apiRetryAuthorization).toBe('Bearer new-access-token-after-refresh');
    
    // 验证页面正常加载（没有跳转到登录页）
    await expect(page).not.toHaveURL(/auth\/login/);
  });

  test('TC0234b: 多请求并发 401 仅触发一次刷新', async ({
    page,
  }) => {
    let refreshRequestCount = 0;
    let api1RequestCount = 0;
    let api2RequestCount = 0;

    // Mock refresh token 接口
    await page.route('**/api/v1/auth/refresh', async (route) => {
      refreshRequestCount += 1;
      // 模拟网络延迟，让多个请求有机会同时到达
      await new Promise((resolve) => setTimeout(resolve, 500));
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          data: {
            accessToken: 'new-access-token-concurrent',
            refreshToken: 'new-refresh-token',
          },
          message: 'ok',
        }),
      });
    });

    // Mock 两个并发的业务的 API
    await page.route('**/api/v1/user/info', async (route) => {
      api1RequestCount += 1;
      if (api1RequestCount === 1) {
        await route.fulfill({
          contentType: 'application/json',
          status: 401,
          body: JSON.stringify({ code: 401, message: 'Unauthorized' }),
        });
        return;
      }
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
        }),
      });
    });

    await page.route('**/api/v1/menus/all', async (route) => {
      api2RequestCount += 1;
      if (api2RequestCount === 1) {
        await route.fulfill({
          contentType: 'application/json',
          status: 401,
          body: JSON.stringify({ code: 401, message: 'Unauthorized' }),
        });
        return;
      }
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 0,
          data: { list: [] },
        }),
      });
    });

    // 设置过期的 token
    await page.addInitScript(() => {
      localStorage.clear();
      const expiredAccessState = JSON.stringify({
        accessCodes: ['*'],
        accessToken: 'expired-access-token',
        isLockScreen: false,
        refreshToken: 'valid-refresh-token',
      });
      for (const envName of ['dev', 'prod']) {
        localStorage.setItem(
          `lina-web-antd-5.6.0-${envName}-core-access`,
          expiredAccessState,
        );
      }
    });

    await page.goto('/dashboard/analytics');
    
    // 等待所有请求完成
    await page.waitForTimeout(3000);
    
    // 关键验证：refresh 接口只被调用了一次（并发去重）
    expect(refreshRequestCount).toBe(1);
    
    // 验证两个 API 都被调用了至少两次（第一次 401，第二次重试）
    expect(api1RequestCount).toBeGreaterThanOrEqual(2);
    expect(api2RequestCount).toBeGreaterThanOrEqual(2);
  });

  test('TC0234c: Refresh Token 失效跳转登录页', async ({
    page,
  }) => {
    let refreshRequestCount = 0;
    let logoutDetected = false;

    // Mock refresh token 接口 - 返回失败（refresh token 失效）
    await page.route('**/api/v1/auth/refresh', async (route) => {
      refreshRequestCount += 1;
      await route.fulfill({
        contentType: 'application/json',
        status: 200,
        body: JSON.stringify({
          code: 401,
          message: 'Refresh token invalid or expired',
          messageKey: 'error.auth.token.refreshFailed',
        }),
      });
    });

    // Mock 业务 API - 返回 401
    await page.route('**/api/v1/user/info', async (route) => {
      await route.fulfill({
        contentType: 'application/json',
        status: 401,
        body: JSON.stringify({ code: 401, message: 'Unauthorized' }),
      });
    });

    // 监听是否跳转到登录页
    page.on('framenavigated', (frame) => {
      if (frame.url().includes('/auth/login')) {
        logoutDetected = true;
      }
    });

    // 设置过期的 access token 和无效的 refresh token
    await page.addInitScript(() => {
      localStorage.clear();
      const expiredAccessState = JSON.stringify({
        accessCodes: ['*'],
        accessToken: 'expired-access-token',
        isLockScreen: false,
        refreshToken: 'invalid-refresh-token',
      });
      for (const envName of ['dev', 'prod']) {
        localStorage.setItem(
          `lina-web-antd-5.6.0-${envName}-core-access`,
          expiredAccessState,
        );
      }
    });

    await page.goto('/dashboard/analytics');
    
    // 等待 refresh 请求完成
    await expect
      .poll(() => refreshRequestCount, { timeout: 15_000 })
      .toBe(1);
    
    // 等待页面跳转到登录页
    await page.waitForURL(/auth\/login/, { timeout: 15_000 });
    
    // 验证确实跳转到了登录页
    expect(page.url()).toContain('/auth/login');
    expect(logoutDetected).toBe(true);
  });
});
