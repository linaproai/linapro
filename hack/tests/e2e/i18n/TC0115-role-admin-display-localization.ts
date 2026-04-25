import { test, expect } from '../../fixtures/auth';
import { RolePage } from '../../pages/RolePage';

test.describe('TC0115 内置超级管理员角色英文展示专项回归', () => {
  test('TC-115a: 英文环境下角色管理页将内置超级管理员投影为英文且不影响其他可编辑角色', async ({
    adminPage,
    mainLayout,
  }) => {
    const rolePage = new RolePage(adminPage);

    await mainLayout.switchLanguage('English');
    await rolePage.goto();

    const adminRow = rolePage.roleRowByKey('admin');
    await expect(adminRow).toContainText('Administrator');
    await expect(adminRow).not.toContainText('超级管理员');

    const standardUserRow = rolePage.roleRowByKey('user');
    await expect(standardUserRow).toContainText('普通用户');
  });
});
