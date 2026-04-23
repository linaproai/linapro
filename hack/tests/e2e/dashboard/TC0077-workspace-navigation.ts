import { test, expect } from '../../fixtures/auth';
import { DashboardPage } from '../../pages/DashboardPage';

test.describe('TC0077 默认工作台', () => {
  test('TC0077a: 工作台恢复参考项目的项目卡片与快捷导航内容', async ({ adminPage }) => {
    const dashboardPage = new DashboardPage(adminPage);

    await dashboardPage.gotoWorkspace();

    await expect(dashboardPage.workspaceDescription).toContainText('今日晴');
    await expect(dashboardPage.workspaceProjects.getByText('Github', { exact: true })).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('首页')).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('仪表盘')).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('组件')).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('系统管理')).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('权限管理')).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('图表')).toBeVisible();
    await expect(adminPage.getByText('审查前端代码提交', { exact: true })).toBeVisible();
  });

  test('TC0077b: 工作台快捷导航跳转到当前项目的可达页面', async ({ adminPage }) => {
    const dashboardPage = new DashboardPage(adminPage);

    await dashboardPage.gotoWorkspace();

    await dashboardPage.clickWorkspaceQuickNav('系统管理');
    await expect(adminPage).toHaveURL(/\/system\/user/);
    await expect(adminPage.getByText('用户列表', { exact: true })).toBeVisible();

    await dashboardPage.gotoWorkspace();

    await dashboardPage.clickWorkspaceQuickNav('图表');
    await expect(adminPage).toHaveURL(/\/dashboard\/analytics/);
    await expect(dashboardPage.analyticsPage).toBeVisible();
  });
});
