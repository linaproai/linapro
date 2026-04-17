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

type UserDetail = {
  code: number;
  data: {
    id: number;
    username: string;
    nickname: string;
    email: string;
    phone: string;
    sex: number;
    status: number;
    remark: string;
    deptId: number;
    postIds: number[];
    roleIds: number[];
  };
};

async function apiGetUserDetailByUsername(
  token: string,
  username: string,
): Promise<UserDetail['data']> {
  const listResp = await fetch(
    `${API_BASE}/user?username=${encodeURIComponent(username)}&pageNum=1&pageSize=10`,
    {
      headers: { Authorization: `Bearer ${token}` },
    },
  );
  const listData = await listResp.json();
  expect(listData.code).toBe(0);
  expect(listData.data.list.length).toBeGreaterThan(0);

  const userId = listData.data.list[0].id;
  const detailResp = await fetch(`${API_BASE}/user/${userId}`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  const detailData = (await detailResp.json()) as UserDetail;
  expect(detailData.code).toBe(0);
  return detailData.data;
}

async function apiUpdateUserRoles(
  token: string,
  user: UserDetail['data'],
  roleIds: number[],
): Promise<void> {
  const resp = await fetch(`${API_BASE}/user/${user.id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      id: user.id,
      username: user.username,
      nickname: user.nickname,
      email: user.email,
      phone: user.phone,
      sex: user.sex,
      status: user.status,
      remark: user.remark,
      deptId: user.deptId,
      postIds: user.postIds,
      roleIds,
    }),
  });
  const data = await resp.json();
  expect(data.code).toBe(0);
}

test.describe('TC0041 消息列表预览弹窗查看通知详情', () => {
  test('TC0041a: 从消息列表点击消息弹出预览窗口', async ({ adminPage }) => {
    const adminToken = await apiLogin(config.adminUser, config.adminPass);
    await apiClearMessages(adminToken);

    const senderUser = await apiGetUserDetailByUsername(adminToken, 'user001');
    const originalRoleIds = [...senderUser.roleIds];
    const elevatedRoleIds = Array.from(new Set([...originalRoleIds, 1]));
    let noticeId = 0;

    await apiUpdateUserRoles(adminToken, senderUser, elevatedRoleIds);

    const senderToken = await apiLogin('user001', config.adminPass);
    const title = `预览测试通知_${Date.now()}`;
    const content = '<p>这是预览测试的通知内容</p>';

    try {
      const resp = await fetch(`${API_BASE}/notice`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${senderToken}`,
        },
        body: JSON.stringify({ title, type: 1, content, status: 1 }),
      });
      const createData = await resp.json();
      expect(createData.code).toBe(0);
      noticeId = createData.data.id;

      await adminPage.goto('/system/message');
      await adminPage.waitForLoadState('networkidle');
      await adminPage.getByText(title).click();

      const modal = adminPage.locator('[role="dialog"]');
      await expect(modal).toBeVisible({ timeout: 10000 });
      await expect(
        modal.getByText('这是预览测试的通知内容'),
      ).toBeVisible({ timeout: 5000 });
      await expect(
        modal.locator('.ant-descriptions'),
      ).toBeVisible({ timeout: 5000 });
    } finally {
      if (noticeId > 0) {
        await fetch(`${API_BASE}/notice/${noticeId}`, {
          method: 'DELETE',
          headers: { Authorization: `Bearer ${adminToken}` },
        });
      }
      await apiClearMessages(adminToken);
      await apiUpdateUserRoles(adminToken, senderUser, originalRoleIds);
    }
  });
});
