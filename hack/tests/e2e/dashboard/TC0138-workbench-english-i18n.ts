import { test, expect } from '../../fixtures/auth';
import { DashboardPage } from '../../pages/DashboardPage';

const chineseTextPattern = /[\u3400-\u9fff]/u;

test.describe('TC-138 Workbench English i18n regression', () => {
  test('TC-138a: English workbench content contains no Chinese system copy', async ({
    adminPage,
    mainLayout,
  }) => {
    const dashboardPage = new DashboardPage(adminPage);

    await adminPage.setViewportSize({ width: 1366, height: 900 });
    await mainLayout.switchLanguage('English');
    await dashboardPage.gotoWorkspace();

    await expect(dashboardPage.workspacePage).toBeVisible();
    await expect(dashboardPage.workspaceDescription).toContainText(
      'Sunny today',
    );
    await expect(
      adminPage.getByRole('heading', { name: 'Projects' }).first(),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceProjects.getByText('LinaPro', { exact: true }),
    ).toBeVisible();
    const linaProDescription = dashboardPage.workspaceProjects.getByText(
      'Let the framework carry the complexity.',
      { exact: true },
    );
    await expect(linaProDescription).toHaveCSS('white-space', 'nowrap');
    await expect(linaProDescription).toHaveCSS('overflow', 'hidden');
    await expect(linaProDescription).toHaveCSS('text-overflow', 'ellipsis');
    await expect(
      dashboardPage.workspaceProjects.getByText('GoFrame', { exact: true }),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceProjects.locator('img[alt="GoFrame"]'),
    ).toHaveAttribute('src', /\/goframe-logo\.webp$/);
    await expect(
      dashboardPage.workspaceProjects.getByText('Ant Design', { exact: true }),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceProjects.getByText('2026-05-01', { exact: true }),
    ).toHaveCount(6);
    await expect(
      adminPage.getByRole('heading', { name: 'Quick Navigation' }).first(),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceQuickNavItem('User Management'),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceQuickNavItem('Menu Management'),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceQuickNavItem('System Parameters'),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceQuickNavItem('Extension Center'),
    ).toBeVisible();
    await expect(dashboardPage.workspaceQuickNavItem('API Docs')).toBeVisible();
    await expect(
      dashboardPage.workspaceQuickNavItem('Scheduled Jobs'),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceTodos
        .getByText('Review Workbench Shortcuts', { exact: true })
        .first(),
    ).toBeVisible();
    await expect(
      dashboardPage.workspaceTrends
        .getByText('workspace quick navigation', { exact: true })
        .first(),
    ).toBeVisible();

    const workspaceText = await dashboardPage.workspacePage.innerText();
    const systemCopyText = workspaceText.replace(/^Good [^\n]+\n?/u, '');
    expect(systemCopyText).not.toMatch(chineseTextPattern);
  });
});
