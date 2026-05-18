import { test, expect } from '../../fixtures/auth';
import { DashboardPage } from '../../pages/DashboardPage';

test.describe('TC0076 默认分析页', () => {
  test('TC0076a: 分析页恢复参考项目的默认概览与图表卡片', async ({ adminPage }) => {
    const dashboardPage = new DashboardPage(adminPage);

    await dashboardPage.gotoAnalytics();

    await expect(dashboardPage.analyticsMetric('用户量')).toBeVisible();
    await expect(dashboardPage.analyticsMetric('访问量')).toBeVisible();
    await expect(dashboardPage.analyticsMetric('下载量')).toBeVisible();
    await expect(dashboardPage.analyticsMetric('使用量')).toBeVisible();
    await expect(dashboardPage.analyticsCardTitle('访问数量')).toBeVisible();
    await expect(dashboardPage.analyticsCardTitle('访问来源')).toBeVisible();
    await expect(dashboardPage.analyticsCardTitle('商业占比')).toBeVisible();
  });

  test('TC0076b: 分析页标签切换仍可正常工作', async ({ adminPage }) => {
    const dashboardPage = new DashboardPage(adminPage);

    await dashboardPage.gotoAnalytics();

    await dashboardPage.analyticsTab('月访问量').click();
    await expect(dashboardPage.analyticsTab('月访问量')).toHaveAttribute('data-state', 'active');

    await dashboardPage.analyticsTab('流量趋势').click();
    await expect(dashboardPage.analyticsTab('流量趋势')).toHaveAttribute('data-state', 'active');
  });
});
