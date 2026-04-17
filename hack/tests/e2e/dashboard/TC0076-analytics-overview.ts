import { test, expect } from '../../fixtures/auth';
import { DashboardPage } from '../../pages/DashboardPage';

test.describe('TC0076 默认分析页', () => {
  test('TC0076a: 分析页默认展示宿主概览与关键洞察', async ({ adminPage }) => {
    const dashboardPage = new DashboardPage(adminPage);

    await dashboardPage.gotoAnalytics();

    await expect(dashboardPage.analyticsSummary).toContainText('最近 7 天');
    await expect(dashboardPage.analyticsInsightCards).toHaveCount(3);
    await expect(dashboardPage.touchpointCardTitle).toContainText('近 7 天触点覆盖');
    await expect(dashboardPage.sourceCardTitle).toBeVisible();
    await expect(dashboardPage.salesCardTitle).toBeVisible();
  });

  test('TC0076b: 切换时间范围时同步刷新摘要与图表标题', async ({ adminPage }) => {
    const dashboardPage = new DashboardPage(adminPage);

    await dashboardPage.gotoAnalytics();

    await dashboardPage.selectAnalyticsRange('month');
    await expect(dashboardPage.analyticsSummary).toContainText('最近 30 天');
    await expect(dashboardPage.touchpointCardTitle).toContainText('近 30 天触点覆盖');

    await dashboardPage.selectAnalyticsRange('today');
    await expect(dashboardPage.analyticsSummary).toContainText('近 24 小时');
    await expect(dashboardPage.touchpointCardTitle).toContainText('今日触点覆盖');
  });
});
