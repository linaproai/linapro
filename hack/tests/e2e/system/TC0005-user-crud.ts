import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0005 用户管理 CRUD', () => {
  const testUsername = `e2e_user_${Date.now()}`;

  test('TC0005a: 创建新用户', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();
    await userPage.createUser(testUsername, 'test123456', 'E2E测试用户');

    await expect(adminPage.getByText(/创建成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0005b: 用户列表中可见新创建的用户', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    const hasUser = await userPage.hasUser(testUsername);
    expect(hasUser).toBeTruthy();
  });

  test('TC0005c: 编辑用户信息', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();
    await userPage.editUser(testUsername, { nickname: '修改后的E2E用户' });

    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0005d: 删除用户', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();
    await userPage.deleteUser(testUsername);

    await expect(adminPage.getByText(/删除成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });
});
