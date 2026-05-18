import { test, expect } from '../../fixtures/auth';
import { RolePage } from '../../pages/RolePage';
import {
  createAdminApiContext,
  expectSuccess,
} from '../../support/api/job';

test.describe('TC0115 内置超级管理员角色英文展示专项回归', () => {
  test('TC-115a: 英文环境下角色管理页将内置超级管理员投影为英文且不影响其他可编辑角色', async ({
    adminPage,
    mainLayout,
  }) => {
    const rolePage = new RolePage(adminPage);
    const api = await createAdminApiContext();
    try {
      const data = await expectSuccess<{
        list: Array<{ key: string; name: string }>;
        total: number;
      }>(
        await api.get('role?key=admin&page=1&size=10', {
          headers: { 'Accept-Language': 'en-US' },
        }),
      );
      const adminRole = data.list.find((item) => item.key === 'admin');
      expect(adminRole?.name).toBe('Administrator');
    } finally {
      await api.dispose();
    }

    await mainLayout.switchLanguage('English');
    await rolePage.goto();

    const adminRow = rolePage.roleRowByKey('admin');
    await expect(adminRow).toContainText('Administrator');
    await expect(adminRow).not.toContainText('超级管理员');

    const standardUserRow = rolePage.roleRowByKey('user');
    await expect(standardUserRow).toContainText('User');
    await expect(standardUserRow).not.toContainText('普通用户');
  });
});
