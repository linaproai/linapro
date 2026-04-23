import { execFileSync } from 'node:child_process';

import type {
  APIRequestContext,
  APIResponse,
  Browser,
  BrowserContext,
  Page,
} from '@playwright/test';

import { request as playwrightRequest } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { config } from '../../../fixtures/config';
import {
  ensureSourcePluginDisabled,
  ensureSourcePluginEnabled,
  ensureSourcePluginUninstalled,
} from '../../../fixtures/plugin';
import { LoginPage } from '../../../pages/LoginPage';

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? 'http://127.0.0.1:8080/api/v1/';
const mysqlBin = process.env.E2E_MYSQL_BIN ?? 'mysql';
const mysqlUser = process.env.E2E_DB_USER ?? 'root';
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? '12345678';
const mysqlDatabase = process.env.E2E_DB_NAME ?? 'lina';

function unwrapApiData(payload: any) {
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return payload.data;
  }
  return payload;
}

function assertOk(response: APIResponse, message: string) {
  expect(response.ok(), `${message}, status=${response.status()}`).toBeTruthy();
}

function decodeTokenId(accessToken: string): string {
  const payload = accessToken.split('.')[1] ?? '';
  const decoded = JSON.parse(
    Buffer.from(payload, 'base64url').toString('utf8'),
  ) as { tokenId?: string };
  expect(decoded.tokenId, 'JWT 中缺少 tokenId').toBeTruthy();
  return decoded.tokenId!;
}

async function createAdminSessionContext(): Promise<{
  api: APIRequestContext;
  tokenId: string;
}> {
  const loginApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await loginApi.post('auth/login', {
    data: {
      password: config.adminPass,
      username: config.adminUser,
    },
  });
  assertOk(loginResponse, '管理员登录 API 失败');

  const loginResult = unwrapApiData(await loginResponse.json());
  const accessToken = loginResult?.accessToken;
  expect(accessToken, '未获取到 accessToken').toBeTruthy();
  const tokenId = decodeTokenId(accessToken);

  await loginApi.dispose();

  const api = await playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
  return { api, tokenId };
}

function expireOnlineSession(tokenId: string) {
  const escapedTokenId = tokenId.replaceAll("'", "''");
  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      '-e',
      `UPDATE sys_online_session SET last_active_time = DATE_SUB(NOW(), INTERVAL 3 DAY) WHERE token_id = '${escapedTokenId}';`,
    ],
    {
      stdio: 'ignore',
    },
  );
}

async function createIsolatedPage(browser: Browser): Promise<{
  cleanup: () => Promise<void>;
  context: BrowserContext;
  loginPage: LoginPage;
  page: Page;
}> {
  const context = await browser.newContext();
  const page = await context.newPage();
  return {
    cleanup: async () => {
      await context.close();
    },
    context,
    loginPage: new LoginPage(page),
    page,
  };
}

test.describe('TC-99 宿主与监控插件边界回归', () => {
  test('TC0099a: monitor-online 缺失时登录、鉴权与会话过期校验仍由宿主内核保障', async ({
    adminPage,
    browser,
  }) => {
    let isolatedContext: BrowserContext | null = null;

    try {
      await ensureSourcePluginUninstalled(adminPage, 'monitor-online');

      const isolated = await createIsolatedPage(browser);
      isolatedContext = isolated.context;
      await isolated.loginPage.goto();
      await isolated.loginPage.loginAndWaitForRedirect('admin', 'admin123');
      await expect(isolated.page).not.toHaveURL(/\/auth\/login/);
      await isolated.cleanup();
      isolatedContext = null;

      const session = await createAdminSessionContext();
      try {
        const userInfoResponse = await session.api.get('user/info');
        assertOk(userInfoResponse, 'monitor-online 缺失时 user/info 鉴权失败');

        expireOnlineSession(session.tokenId);
        await expect
          .poll(
            async () => {
              const expiredResponse = await session.api.get('user/info');
              return expiredResponse.status();
            },
            {
              intervals: [100, 200, 500, 1000],
              timeout: 5000,
            },
          )
          .toBe(401);
      } finally {
        await session.api.dispose();
      }
    } finally {
      if (isolatedContext) {
        await isolatedContext.close();
      }

      await adminPage.goto('/dashboard/analysis');
      await ensureSourcePluginEnabled(adminPage, 'monitor-online');
    }
  });

  test('TC0099b: 登录日志缺失且操作日志停用时，登录与普通业务请求仍正常', async ({
    adminPage,
    browser,
  }) => {
    let isolatedContext: BrowserContext | null = null;

    try {
      await ensureSourcePluginUninstalled(adminPage, 'monitor-loginlog');
      await ensureSourcePluginDisabled(adminPage, 'monitor-operlog');

      const isolated = await createIsolatedPage(browser);
      isolatedContext = isolated.context;
      const { page } = isolated;

      await isolated.loginPage.goto();
      await isolated.loginPage.loginAndWaitForRedirect('admin', 'admin123');
      await expect(page).not.toHaveURL(/\/auth\/login/);

      const userListResponse = page.waitForResponse(
        (response) =>
          response.url().includes('/api/v1/user?') &&
          response.request().method() === 'GET' &&
          response.status() === 200,
        { timeout: 15000 },
      );
      await page.goto('/system/user');
      await userListResponse;
      await expect(page.locator('.vxe-table')).toBeVisible({ timeout: 10000 });
      await expect(page.locator('.vxe-body--row').first()).toBeVisible();
    } finally {
      if (isolatedContext) {
        await isolatedContext.close();
      }

      await adminPage.goto('/dashboard/analysis');
      await ensureSourcePluginEnabled(adminPage, 'monitor-loginlog');
      await ensureSourcePluginEnabled(adminPage, 'monitor-operlog');
    }
  });
});
