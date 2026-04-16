import { test, expect } from '../../fixtures/auth';
import { config } from '../../fixtures/config';

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

test.describe('TC0041 消息列表预览弹窗查看通知详情', () => {
  test('TC0041a: 从消息列表点击消息弹出预览窗口', async ({ adminPage }) => {
    // Create a notice as user001 so admin receives the message via fan-out
    const adminToken = await apiLogin(config.adminUser, config.adminPass);
    await apiClearMessages(adminToken);

    const user001Token = await apiLogin('user001', config.adminPass);
    const title = `预览测试通知_${Date.now()}`;
    const content = '<p>这是预览测试的通知内容</p>';
    const resp = await fetch(`${API_BASE}/notice`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${user001Token}`,
      },
      body: JSON.stringify({ title, type: 1, content, status: 1 }),
    });
    const createData = await resp.json();
    expect(createData.code).toBe(0);
    const noticeId = createData.data.id;

    // Navigate admin to message list page
    await adminPage.goto('/system/message');
    await adminPage.waitForLoadState('networkidle');

    // Click on the message
    await adminPage.getByText(title).click();

    // Preview modal should appear
    const modal = adminPage.locator('[role="dialog"]');
    await expect(modal).toBeVisible({ timeout: 10000 });

    // Should show the notice content in the preview modal
    await expect(
      modal.getByText('这是预览测试的通知内容'),
    ).toBeVisible({ timeout: 5000 });

    // Should show descriptions with type info
    const descArea = modal.locator('.ant-descriptions');
    await expect(descArea).toBeVisible({ timeout: 5000 });

    // Cleanup: delete the notice and clear messages
    await fetch(`${API_BASE}/notice/${noticeId}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${user001Token}` },
    });
    await apiClearMessages(adminToken);
  });
});
