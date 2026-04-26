import { test, expect } from '../../fixtures/auth';

test.describe('TC0119 管理工作台底栏版权偏好', () => {
  test.beforeEach(async ({ adminPage, mainLayout }) => {
    await adminPage.setViewportSize({ width: 1366, height: 900 });
    await mainLayout.switchLanguage('简体中文');
  });

  test('TC-119a: 默认显示底栏版权信息', async ({ mainLayout }) => {
    const currentYear = String(new Date().getFullYear());

    await expect(mainLayout.workspaceFooterCopyright).toContainText(
      `Copyright © ${currentYear}`,
    );
    await expect(mainLayout.workspaceFooterCopyright).toContainText(
      'LinaPro.AI',
    );
  });

  test('TC-119b: 布局页隐藏底栏与版权设置板块', async ({
    mainLayout,
  }) => {
    await mainLayout.openPreferencesTab('布局');

    await expect(
      mainLayout.preferencesDrawer.getByRole('heading', {
        exact: true,
        name: '底栏',
      }),
    ).toHaveCount(0);
    await expect(
      mainLayout.preferencesDrawer.getByRole('heading', {
        exact: true,
        name: '版权',
      }),
    ).toHaveCount(0);
  });
});
