import { test, expect } from '../../fixtures/auth';
import { ConfigPage } from '../../pages/ConfigPage';
import { DeptPage } from '../../pages/DeptPage';
import { DictPage } from '../../pages/DictPage';
import { JobGroupPage } from '../../pages/JobGroupPage';
import { JobPage } from '../../pages/JobPage';
import { MenuPage } from '../../pages/MenuPage';
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
    await expect(await deptPage.hasDept('研发部门')).toBe(true);

    await postPage.goto();
    await expect(await postPage.hasPostName('总经理')).toBe(true);

    await rolePage.goto();
    await expect(await rolePage.hasRole('超级管理员')).toBe(true);
  });

  test('TC-110b: 英文环境下菜单字典参数等管理页中的可编辑主数据保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const menuPage = new MenuPage(adminPage);
    const dictPage = new DictPage(adminPage);
    const configPage = new ConfigPage(adminPage);

    await mainLayout.switchLanguage('English');

    await menuPage.goto();
    await menuPage.expandAll();
    await expect(await menuPage.hasMenu('系统设置')).toBe(true);

    await dictPage.goto();
    await expect(await dictPage.hasType('定时任务状态')).toBe(true);

    await configPage.goto();
    await configPage.fillSearchField('参数键名', 'demo.notice.banner');
    await configPage.clickSearch();
    const configRow = configPage.findRowByExactKey('demo.notice.banner');
    await expect(configRow).toBeVisible();
    await expect(configRow).toContainText('演示-首页公告文案');
  });

  test('TC-110c: 英文环境下通知与调度管理页中的可编辑业务记录保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const jobGroupPage = new JobGroupPage(adminPage);
    const jobPage = new JobPage(adminPage);
    const noticePage = new NoticePage(adminPage);

    await mainLayout.switchLanguage('English');

    await jobGroupPage.goto();
    await expect(await jobGroupPage.hasGroup('默认分组')).toBe(true);

    await jobPage.goto();
    await expect(await jobPage.hasJob('服务监控采集')).toBe(true);

    await noticePage.goto();
    await expect(await noticePage.hasNotice('系统升级通知')).toBe(true);
  });
});
