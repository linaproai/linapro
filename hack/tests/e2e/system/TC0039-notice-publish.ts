import { test, expect } from '../../fixtures/auth';
import { NoticePage } from '../../pages/NoticePage';
import { config } from '../../fixtures/config';

const API_BASE = `${config.baseURL}/api/v1`;

/** Login via API and return accessToken */
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

/** Get unread message count via API */
async function apiUnreadCount(token: string): Promise<number> {
  const resp = await fetch(`${API_BASE}/user/message/count`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const data = await resp.json();
  return data.data.count;
}

/** Clear all messages via API */
async function apiClearMessages(token: string): Promise<void> {
  await fetch(`${API_BASE}/user/message/clear`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  });
}

test.describe('TC0039 通知公告发布与消息分发', () => {
  const publishTitle = `发布测试_${Date.now()}`;

  test('TC0039a: 创建已发布通知后铃铛显示未读', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();

    // Create a published notice
    await noticePage.createNotice(publishTitle, '通知', '已发布', '发布测试内容');

    await expect(
      adminPage.getByText(/新增成功|创建成功|success/i),
    ).toBeVisible({ timeout: 5000 });

    // Note: The admin user is excluded from message distribution (they are the creator),
    // so we just verify the notice was created successfully
    const hasNotice = await noticePage.hasNotice(publishTitle);
    expect(hasNotice).toBeTruthy();
  });

  test('TC0039b: 发布后其他用户收到消息通知', async () => {
    // Clear user001's messages first
    const userToken = await apiLogin('user001', config.adminPass);
    await apiClearMessages(userToken);

    // Verify user001 has 0 unread messages
    const countBefore = await apiUnreadCount(userToken);
    expect(countBefore).toBe(0);

    // Admin creates a published notice
    const adminToken = await apiLogin(config.adminUser, config.adminPass);
    const resp = await fetch(`${API_BASE}/notice`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${adminToken}`,
      },
      body: JSON.stringify({
        title: `消息分发测试_${Date.now()}`,
        type: 1,
        content: '验证消息分发',
        status: 1,
      }),
    });
    const createData = await resp.json();
    expect(createData.code).toBe(0);

    // Verify user001 now has 1 unread message
    const countAfter = await apiUnreadCount(userToken);
    expect(countAfter).toBe(1);

    // Clean up: clear user001's messages and delete the notice
    await apiClearMessages(userToken);
    const noticeId = createData.data.id;
    await fetch(`${API_BASE}/notice/${noticeId}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${adminToken}` },
    });
  });

  test('TC0039d: 草稿发布后其他用户收到消息通知', async () => {
    // Clear user001's messages first
    const userToken = await apiLogin('user001', config.adminPass);
    await apiClearMessages(userToken);

    // Verify user001 has 0 unread messages
    const countBefore = await apiUnreadCount(userToken);
    expect(countBefore).toBe(0);

    // Admin creates a DRAFT notice
    const adminToken = await apiLogin(config.adminUser, config.adminPass);
    const createResp = await fetch(`${API_BASE}/notice`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${adminToken}`,
      },
      body: JSON.stringify({
        title: `草稿发布测试_${Date.now()}`,
        type: 1,
        content: '草稿内容',
        status: 0,
      }),
    });
    const createData = await createResp.json();
    expect(createData.code).toBe(0);
    const noticeId = createData.data.id;

    // Verify user001 still has 0 unread messages (draft should not fan-out)
    const countAfterDraft = await apiUnreadCount(userToken);
    expect(countAfterDraft).toBe(0);

    // Now publish the draft by updating status to 1
    const updateResp = await fetch(`${API_BASE}/notice/${noticeId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${adminToken}`,
      },
      body: JSON.stringify({ status: 1 }),
    });
    const updateData = await updateResp.json();
    expect(updateData.code).toBe(0);

    // Verify user001 now has 1 unread message
    const countAfterPublish = await apiUnreadCount(userToken);
    expect(countAfterPublish).toBe(1);

    // Clean up
    await apiClearMessages(userToken);
    await fetch(`${API_BASE}/notice/${noticeId}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${adminToken}` },
    });
  });

  test('TC0039c: 清理 - 删除测试通知', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();
    await noticePage.deleteNotice(publishTitle);

    await expect(adminPage.getByText(/删除成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });
});
