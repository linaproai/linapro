import { test, expect } from '../../../fixtures/auth';
import { ensureSourcePluginEnabled } from '../../../fixtures/plugin';
import { UserPage } from '../../../pages/UserPage';

test.describe('TC0005 用户管理 CRUD', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'org-center');
  });

  function createTestUsername(scope: string) {
    return `e2e_user_${scope}_${Date.now()}`;
  }

  async function searchUser(userPage: UserPage, username: string) {
    await userPage.goto();
    await userPage.searchByUsername(username);
  }

  async function expectUserVisible(userPage: UserPage, username: string) {
    await expect
      .poll(
        async () => {
          await searchUser(userPage, username);
          return userPage.hasUser(username);
        },
        {
          message: `expected user ${username} to appear in list`,
          timeout: 20_000,
        },
      )
      .toBeTruthy();
  }

  async function deleteUserIfExists(userPage: UserPage, username: string) {
    await searchUser(userPage, username);
    if (await userPage.hasUser(username)) {
      await userPage.deleteUser(username);
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
      await deleteUserIfExists(userPage, testUsername);
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
      await deleteUserIfExists(userPage, testUsername);
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
      await deleteUserIfExists(userPage, testUsername);
    }
  });

  test('TC0005d: 删除用户', async ({ adminPage }) => {
    const testUsername = createTestUsername('delete');
    const userPage = new UserPage(adminPage);
    await userPage.goto();
    await userPage.createUser(testUsername, 'test123456', 'E2E测试用户');
    await expectUserVisible(userPage, testUsername);
    await userPage.deleteUser(testUsername);

    await searchUser(userPage, testUsername);
    expect(await userPage.hasUser(testUsername)).toBeFalsy();
  });
});
