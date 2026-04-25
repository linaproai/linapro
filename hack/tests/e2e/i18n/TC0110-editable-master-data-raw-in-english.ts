import { test, expect } from '../../fixtures/auth';
import { ConfigPage } from '../../pages/ConfigPage';
import { DeptPage } from '../../pages/DeptPage';
import { NoticePage } from '../../pages/NoticePage';
import { PostPage } from '../../pages/PostPage';
import { RolePage } from '../../pages/RolePage';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0110 可编辑主数据退出 i18n 投影专项回归', () => {
  test('TC-110a: 英文环境下用户与组织管理页面中的可编辑主数据保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const userPage = new UserPage(adminPage);
    const deptPage = new DeptPage(adminPage);
    const postPage = new PostPage(adminPage);
    const rolePage = new RolePage(adminPage);

    await mainLayout.switchLanguage('English');

    await userPage.goto();
    await expect(await userPage.hasDeptTreeNode('研发部门')).toBe(true);

    await deptPage.goto();
    await expect(await deptPage.hasDeptInExpandedTree('研发部门')).toBe(true);

    await postPage.goto();
    await expect(await postPage.hasPostName('总经理')).toBe(true);

    await rolePage.goto();
    await expect(await rolePage.hasRole('普通用户')).toBe(true);
  });

  test('TC-110b: 英文环境下参数管理页中的可编辑主数据保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const configPage = new ConfigPage(adminPage);

    await mainLayout.switchLanguage('English');

    await configPage.goto();
    await configPage.fillSearchField('参数键名', 'demo.notice.banner');
    await configPage.clickSearch();
    const configRow = configPage.findRowByExactKey('demo.notice.banner');
    await expect(configRow).toBeVisible();
    await expect(configRow).toContainText('demo.notice.banner');
    await expect(configRow).toContainText('欢迎使用 LinaPro');
  });

  test('TC-110c: 英文环境下通知管理页中的可编辑业务记录保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const noticePage = new NoticePage(adminPage);

    await mainLayout.switchLanguage('English');

    await noticePage.goto();
    await expect(await noticePage.hasNotice('系统升级通知')).toBe(true);
  });
});
