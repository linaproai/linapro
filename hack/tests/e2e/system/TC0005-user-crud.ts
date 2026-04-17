import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0005 用户管理 CRUD', () => {
  function createTestUsername(scope: string) {
    return `e2e_user_${scope}_${Date.now()}`;
  }

  async function deleteUserIfExists(userPage: UserPage, username: string) {
    await userPage.goto();
    await userPage.fillSearchField('用户账号', username);
    await userPage.clickSearch();
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

      await userPage.goto();
      await userPage.fillSearchField('用户账号', testUsername);
      await userPage.clickSearch();
      expect(await userPage.hasUser(testUsername)).toBeTruthy();
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

      await userPage.goto();
      await userPage.fillSearchField('用户账号', testUsername);
      await userPage.clickSearch();

      const hasUser = await userPage.hasUser(testUsername);
      expect(hasUser).toBeTruthy();
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
      await userPage.goto();
      await userPage.editUser(testUsername, { nickname: '修改后的E2E用户' });

      await userPage.goto();
      await userPage.fillSearchField('用户账号', testUsername);
      await userPage.clickSearch();
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
    await userPage.goto();
    await userPage.deleteUser(testUsername);

    await userPage.goto();
    await userPage.fillSearchField('用户账号', testUsername);
    await userPage.clickSearch();
    expect(await userPage.hasUser(testUsername)).toBeFalsy();
  });
});
