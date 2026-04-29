import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';
import { createAdminApiContext, expectSuccess } from '../../support/api/job';
import { waitForRouteReady, waitForTableReady } from '../../support/ui';

type RoleListResult = {
  list: Array<{ key: string; name: string }>;
  total: number;
};

type UserListResult = {
  list: Array<{ roleNames: string[]; username: string }>;
  total: number;
};

test.describe('TC-139 Built-in role English display regression', () => {
  test('TC-139a: Role and user APIs project built-in role names in English', async () => {
    const api = await createAdminApiContext();
    try {
      const roleData = await expectSuccess<RoleListResult>(
        await api.get('role?page=1&size=100', {
          headers: { 'Accept-Language': 'en-US' },
        }),
      );
      const adminRole = roleData.list.find((item) => item.key === 'admin');
      const userRole = roleData.list.find((item) => item.key === 'user');

      expect(adminRole?.name).toBe('Administrator');
      expect(userRole?.name).toBe('User');
      expect(roleData.list.map((item) => item.name)).not.toContain('普通用户');

      const userData = await expectSuccess<UserListResult>(
        await api.get('user?pageNum=1&pageSize=100&username=admin', {
          headers: { 'Accept-Language': 'en-US' },
        }),
      );
      const adminUser = userData.list.find((item) => item.username === 'admin');
      expect(adminUser?.roleNames).toContain('Administrator');
      expect(adminUser?.roleNames).not.toContain('超级管理员');
    } finally {
      await api.dispose();
    }
  });

  test('TC-139b: User and role management tables show the same English built-in names', async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage('English');

    await adminPage.goto('/system/role');
    await waitForTableReady(adminPage);
    const adminRoleRow = adminPage
      .locator('.vxe-body--row:visible', { hasText: 'admin' })
      .first();
    const userRoleRow = adminPage
      .locator('.vxe-body--row:visible', { hasText: 'user' })
      .first();

    await expect(adminRoleRow).toContainText('Administrator');
    await expect(userRoleRow).toContainText('User');
    await expect(userRoleRow).not.toContainText('普通用户');

    const userPage = new UserPage(adminPage);
    await userPage.goto();
    await userPage.searchByUsername('admin');
    const adminUserRow = userPage.getUserRow('admin');
    await expect(adminUserRow).toContainText('Administrator');
    await expect(adminUserRow).not.toContainText('超级管理员');

    await waitForRouteReady(adminPage);
  });
});
