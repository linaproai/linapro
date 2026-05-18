import { test, expect } from '../../../fixtures/auth';
import {
  createAdminApiContext,
  ensureSourcePluginEnabled,
} from '../../../fixtures/plugin';
import { UserPage } from '../../../pages/UserPage';

test.describe('TC0005 用户管理 CRUD', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'org-center');
  });

  function createTestUsername(scope: string) {
    return `e2e_user_${scope}_${Date.now()}`;
  }

  async function searchUser(userPage: UserPage, username: string) {
    await userPage.searchByUsername(username);
  }

  async function expectUserVisible(userPage: UserPage, username: string) {
    await searchUser(userPage, username);
    await expect(userPage.getUserRow(username)).toBeVisible({ timeout: 20_000 });
  }

  function unwrapApiData(payload: any) {
    if (payload && typeof payload === 'object' && 'data' in payload) {
      return payload.data;
    }
    return payload;
  }

  async function deleteUserIfExists(username: string) {
    const adminApi = await createAdminApiContext();
    try {
      const listResponse = await adminApi.get(
        `user?pageNum=1&pageSize=20&username=${encodeURIComponent(username)}`,
      );
      expect(listResponse.ok(), `查询测试用户失败: ${username}`).toBeTruthy();
      const payload = unwrapApiData(await listResponse.json());
      const user = (payload?.list ?? []).find(
        (item: { id?: number; username?: string }) => item.username === username,
      );
      if (!user?.id) {
        return;
      }

      const deleteResponse = await adminApi.delete(`user?ids=${user.id}`);
      expect(deleteResponse.ok(), `清理测试用户失败: ${username}`).toBeTruthy();
    } finally {
      await adminApi.dispose();
    }
  }

  test('TC0005a: 创建新用户', async ({ adminPage }) => {
    const testUsername = createTestUsername('create');
    const userPage = new UserPage(adminPage);
    try {
      await userPage.goto();
      await userPage.createUser(testUsername, 'test123456', 'E2E测试用户');
      await expectUserVisible(userPage, testUsername);
    } finally {
      await deleteUserIfExists(testUsername);
    }
  });

  test('TC0005b: 用户列表中可见新创建的用户', async ({ adminPage }) => {
    const testUsername = createTestUsername('list');
    const userPage = new UserPage(adminPage);
    try {
      await userPage.goto();
      await userPage.createUser(testUsername, 'test123456', 'E2E测试用户');
      await expectUserVisible(userPage, testUsername);
    } finally {
      await deleteUserIfExists(testUsername);
    }
  });

  test('TC0005c: 编辑用户信息', async ({ adminPage }) => {
    const testUsername = createTestUsername('edit');
    const userPage = new UserPage(adminPage);
    try {
      await userPage.goto();
      await userPage.createUser(testUsername, 'test123456', 'E2E测试用户');
      await expectUserVisible(userPage, testUsername);
      await userPage.editUser(testUsername, { nickname: '修改后的E2E用户' });

      await searchUser(userPage, testUsername);
      await expect(
        adminPage.locator('.vxe-body--row', { hasText: testUsername }).first(),
      ).toContainText('修改后的E2E用户');
    } finally {
      await deleteUserIfExists(testUsername);
    }
  });

  test('TC0005d: 删除用户', async ({ adminPage }) => {
    const testUsername = createTestUsername('delete');
    const userPage = new UserPage(adminPage);
    try {
      await userPage.goto();
      await userPage.createUser(testUsername, 'test123456', 'E2E测试用户');
      await expectUserVisible(userPage, testUsername);
      await userPage.deleteUser(testUsername);

      await searchUser(userPage, testUsername);
      expect(await userPage.hasUser(testUsername)).toBeFalsy();
    } finally {
      await deleteUserIfExists(testUsername);
    }
  });
});
