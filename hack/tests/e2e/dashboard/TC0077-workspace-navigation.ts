import { test, expect } from '../../fixtures/auth';
import { DashboardPage } from '../../pages/DashboardPage';

test.describe('TC0077 默认工作台', () => {
  test('TC0077a: 工作台展示真实快捷入口与运维内容', async ({ adminPage }) => {
    const dashboardPage = new DashboardPage(adminPage);

    await dashboardPage.gotoWorkspace();

    await expect(dashboardPage.workspaceDescription).toContainText('OpenSpec');
    await expect(dashboardPage.workspaceQuickActions).toHaveCount(6);
    await expect(adminPage.getByText('复核插件发布授权快照')).toBeVisible();
    await expect(
      adminPage.getByTestId('dashboard-workspace-projects').getByText('默认管理工作台', { exact: true }),
    ).toBeVisible();
  });

  test('TC0077b: 工作台快捷入口跳转到实际管理页面', async ({ adminPage }) => {
    const dashboardPage = new DashboardPage(adminPage);

    await dashboardPage.gotoWorkspace();

    await dashboardPage.clickWorkspaceQuickAction('plugin-management');
    await expect(adminPage).toHaveURL(/\/system\/plugin/);
    await expect(adminPage.getByText('插件列表')).toBeVisible();

    await dashboardPage.gotoWorkspace();

    await dashboardPage.clickWorkspaceQuickAction('api-docs');
    await expect(adminPage).toHaveURL(/\/about\/api-docs/);
    await expect(adminPage.locator('iframe.api-docs-iframe')).toBeVisible({ timeout: 10_000 });
  });
});
