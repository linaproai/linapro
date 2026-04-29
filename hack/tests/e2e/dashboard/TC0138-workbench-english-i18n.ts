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
      adminPage.getByRole('heading', { name: 'Quick Navigation' }).first(),
    ).toBeVisible();
    await expect(
      adminPage.getByText('Review Frontend Commits', { exact: true }).first(),
    ).toBeVisible();

    const workspaceText = await dashboardPage.workspacePage.innerText();
    expect(workspaceText).not.toMatch(chineseTextPattern);
  });
});
