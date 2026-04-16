import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0062 用户角色关联', () => {
  const testPassword = 'test123456';
  const testNickname = 'E2E用户角色测试';
  const initialRoleName = '普通用户';
  const updatedRoleName = '超级管理员';

  function createTestUsername(scope: string) {
    return `e2e_user_role_${scope}_${Date.now()}`;
  }

  async function deleteUserIfExists(userPage: UserPage, username: string) {
    await userPage.goto();
    const hasUser = await userPage.hasUser(username);
    if (hasUser) {
      await userPage.deleteUser(username);
    }
  }

  test('TC0062a: 创建用户时选择角色', async ({ adminPage }) => {
    const testUsername = createTestUsername('create');
    const userPage = new UserPage(adminPage);

    try {
      await userPage.goto();

      // Create a dedicated user for this assertion so the test remains valid
      // even if Playwright restarts the worker after an earlier failure.
      await userPage.createUserWithRoles(
        testUsername,
        testPassword,
        testNickname,
        [initialRoleName],
      );

      await expect(adminPage.getByText('创建成功')).toBeVisible({
        timeout: 5000,
      });
    } finally {
      await deleteUserIfExists(userPage, testUsername);
    }
  });

  test('TC0062b: 用户列表显示角色信息', async ({ adminPage }) => {
    const testUsername = createTestUsername('list');
    const userPage = new UserPage(adminPage);

    try {
      await userPage.goto();
      await userPage.createUserWithRoles(
        testUsername,
        testPassword,
        testNickname,
        [initialRoleName],
      );
      await expect(adminPage.getByText('创建成功')).toBeVisible({
        timeout: 5000,
      });

      // Search for the test user in a fresh list state and verify the visible
      // role column reflects the assigned role.
      await userPage.goto();
      await userPage.fillSearchField('用户账号', testUsername);
      await userPage.clickSearch();

      const hasUser = await userPage.hasUser(testUsername);
      expect(hasUser).toBeTruthy();

      const roleNames = await userPage.getRoleNames(testUsername);
      expect(roleNames).toContain(initialRoleName);
    } finally {
      await deleteUserIfExists(userPage, testUsername);
    }
  });

  test('TC0062c: 编辑用户修改角色', async ({ adminPage }) => {
    const testUsername = createTestUsername('edit');
    const userPage = new UserPage(adminPage);

    try {
      await userPage.goto();
      await userPage.createUserWithRoles(
        testUsername,
        testPassword,
        testNickname,
        [initialRoleName],
      );
      await expect(adminPage.getByText('创建成功')).toBeVisible({
        timeout: 5000,
      });

      // Replace the user's role with the second dedicated role and verify the
      // list reflects the new assignment after the drawer is saved.
      await userPage.goto();
      await userPage.editUserRoles(testUsername, [updatedRoleName]);

      await expect(adminPage.getByText('更新成功')).toBeVisible({
        timeout: 5000,
      });

      await userPage.goto();
      await userPage.fillSearchField('用户账号', testUsername);
      await userPage.clickSearch();
      const hasUser = await userPage.hasUser(testUsername);
      expect(hasUser).toBeTruthy();

      const roleNames = await userPage.getRoleNames(testUsername);
      expect(roleNames).toContain(updatedRoleName);
    } finally {
      await deleteUserIfExists(userPage, testUsername);
    }
  });

  test('TC0062d: 删除用户时清理角色关联', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Create a new user for testing cleanup
    const cleanupUsername = `e2e_cleanup_${Date.now()}`;
    await userPage.createUser(cleanupUsername, testPassword, 'E2E清理测试');

    // Delete the user
    await userPage.goto();
    await userPage.deleteUser(cleanupUsername);

    await expect(adminPage.getByText(/删除成功|success/i)).toBeVisible({
      timeout: 5000,
    });

    // Verify user is deleted
    await userPage.goto();
    const hasUser = await userPage.hasUser(cleanupUsername);
    expect(hasUser).toBeFalsy();
  });
});
