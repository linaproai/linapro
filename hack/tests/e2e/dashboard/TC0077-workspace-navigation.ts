import { test, expect } from '../../fixtures/auth';
import { DashboardPage } from '../../pages/DashboardPage';

test.describe('TC0077 默认工作台', () => {
  test('TC0077a: 工作台展示管理后台技术栈项目卡片、快捷导航和 LinaPro 示例内容', async ({
    adminPage,
  }) => {
    const dashboardPage = new DashboardPage(adminPage);

    await dashboardPage.gotoWorkspace();

    await expect(dashboardPage.workspaceDescription).toContainText('今日晴');
    await expect(
      dashboardPage.workspaceProjects.getByText('LinaPro', { exact: true }),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceProjects.locator('img[alt="LinaPro"]'),
    ).toHaveAttribute('src', /\/logo\.webp$/);
    await expect(
      dashboardPage.workspaceProjects.getByText('GoFrame', { exact: true }),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceProjects.locator('img[alt="GoFrame"]'),
    ).toHaveAttribute('src', /\/goframe-logo\.webp$/);
    await expect(
      dashboardPage.workspaceProjects.getByText('Vue', { exact: true }),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceProjects.getByText('Vben', { exact: true }),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceProjects.locator('img[alt="Vben"]'),
    ).toHaveAttribute('src', /\/vben-logo\.webp$/);
    await expect(
      dashboardPage.workspaceProjects.getByText('Ant Design', { exact: true }),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceProjects.getByText('TypeScript', { exact: true }),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceProjects.getByText('2026-05-01', { exact: true }),
    ).toHaveCount(6);
    await expect(dashboardPage.workspaceQuickNavItem('用户管理')).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('菜单管理')).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('系统参数')).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('扩展中心')).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('接口文档')).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('定时任务')).toBeVisible();
    await expect(
      dashboardPage.workspaceTodos.getByText('检查工作台快捷入口', {
        exact: true,
      }),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceTrends.getByText('工作台快捷导航', {
        exact: true,
      }),
    ).toBeVisible();
  });

  test('TC0077b: 工作台快捷导航跳转到当前项目的可达页面', async ({ adminPage }) => {
    const dashboardPage = new DashboardPage(adminPage);
    const quickNavCases: Array<[string, RegExp]> = [
      ['用户管理', /\/system\/user/],
      ['菜单管理', /\/system\/menu/],
      ['系统参数', /\/system\/config/],
      ['扩展中心', /\/system\/plugin/],
      ['接口文档', /\/about\/api-docs/],
      ['定时任务', /\/system\/job/],
    ];

    for (const [label, expectedUrl] of quickNavCases) {
      await dashboardPage.gotoWorkspace();
      await dashboardPage.clickWorkspaceQuickNav(label);
      await expect(adminPage).toHaveURL(expectedUrl);
    }
  });
});
