import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { config } from '@host-tests/fixtures/config';
import { LoginPage } from '@host-tests/pages/LoginPage';

const API_BASE = `${config.baseURL}/api/v1`;

async function apiLogin(
  username: string,
  password: string,
): Promise<string> {
  const resp = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
    redirect: 'manual',
  });
  const data = await resp.json();
  return data.data.accessToken;
}

async function apiClearMessages(token: string): Promise<void> {
  await fetch(`${API_BASE}/user/message/clear`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
}

async function apiUnreadCount(token: string): Promise<number> {
  const resp = await fetch(`${API_BASE}/user/message/count`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const data = await resp.json();
  return data.data.count;
}

test.describe('TC0042 用户消息列表页面', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'content-notice');
  });

  test('TC0042a: 消息列表页面可访问', async ({ browser }) => {
    const context = await browser.newContext({ baseURL: config.baseURL });
    const page = await context.newPage();
    const loginPage = new LoginPage(page);

    try {
      await loginPage.goto();
      await loginPage.loginAndWaitForRedirect('user001', config.adminPass);

      await page.goto('/system/message');
      await page.waitForLoadState('networkidle');

      const card = page.locator('.ant-card');
      await expect(card).toBeVisible({ timeout: 10000 });
      await expect(card.locator('.ant-card-head-title')).toHaveText('消息列表');

      await expect(
        page.getByRole('button', { name: /全部已读/ }),
      ).toBeVisible({ timeout: 5000 });
      await expect(
        page.getByRole('button', { name: /清空消息/ }),
      ).toBeVisible({ timeout: 5000 });
    } finally {
      await context.close();
    }
  });

  test('TC0042b: 消息列表展示通知消息', async ({ browser }) => {
    const adminToken = await apiLogin(config.adminUser, config.adminPass);
    const userToken = await apiLogin('user001', config.adminPass);
    await apiClearMessages(userToken);

    const title = `消息列表测试_${Date.now()}`;
    const resp = await fetch(`${API_BASE}/notice`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${adminToken}`,
      },
      body: JSON.stringify({
        title,
        type: 1,
        content: '消息列表测试内容',
        status: 1,
      }),
    });
    const createData = await resp.json();
    expect(createData.code).toBe(0);
    const noticeId = createData.data.id;
    const context = await browser.newContext({ baseURL: config.baseURL });
    const page = await context.newPage();
    const loginPage = new LoginPage(page);

    try {
      await expect.poll(() => apiUnreadCount(userToken), {
        message: 'expected recipient unread-count to include the published notice',
        timeout: 10_000,
      }).toBeGreaterThan(0);

      await loginPage.goto();
      await loginPage.loginAndWaitForRedirect('user001', config.adminPass);

      await page.goto('/system/message');
      await page.waitForLoadState('networkidle');

      const card = page.locator('.ant-card');
      await expect(card).toBeVisible({ timeout: 10000 });
      await expect(card.locator('.ant-card-head-title')).toHaveText('消息列表');
      await expect(page.getByText(title, { exact: true }).first()).toBeVisible({
        timeout: 10_000,
      });
    } finally {
      await fetch(`${API_BASE}/notice/${noticeId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${adminToken}` },
      });
      await apiClearMessages(userToken);
      await context.close();
    }
  });
});
