import type { Locator, Page } from '@playwright/test';

export type AnalyticsRangeKey = 'today' | 'week' | 'month';

export class DashboardPage {
  constructor(private page: Page) {}

  get analyticsPage(): Locator {
    return this.page.getByTestId('dashboard-analytics-page');
  }

  get analyticsInsightCards(): Locator {
    return this.page.getByTestId('dashboard-analytics-insight');
  }

  get analyticsSummary(): Locator {
    return this.page.getByTestId('dashboard-analytics-summary');
  }

  get touchpointCardTitle(): Locator {
    return this.page.getByRole('heading', { name: /触点覆盖/ }).first();
  }

  get sourceCardTitle(): Locator {
    return this.page.getByRole('heading', { name: '来源结构' }).first();
  }

  get salesCardTitle(): Locator {
    return this.page.getByRole('heading', { name: '交付构成' }).first();
  }

  get workspacePage(): Locator {
    return this.page.getByTestId('dashboard-workspace-page');
  }

  get workspaceDescription(): Locator {
    return this.page.getByTestId('dashboard-workspace-description');
  }

  get workspaceQuickActions(): Locator {
    return this.page.locator('button[data-testid^="dashboard-workspace-quick-"]');
  }

  analyticsRangeButton(range: AnalyticsRangeKey): Locator {
    return this.page.getByTestId(`dashboard-range-${range}`);
  }

  workspaceQuickAction(key: string): Locator {
    return this.page.getByTestId(`dashboard-workspace-quick-${key}`);
  }

  async gotoAnalytics() {
    await this.page.goto('/dashboard/analytics');
    await this.page.waitForLoadState('networkidle');
    await this.analyticsPage.waitFor({ state: 'visible' });
  }

  async gotoWorkspace() {
    await this.page.goto('/dashboard/workspace');
    await this.page.waitForLoadState('networkidle');
    await this.workspacePage.waitFor({ state: 'visible' });
  }

  async selectAnalyticsRange(range: AnalyticsRangeKey) {
    await this.analyticsRangeButton(range).click();
  }

  async clickWorkspaceQuickAction(key: string) {
    await this.workspaceQuickAction(key).click();
    await this.page.waitForLoadState('networkidle');
  }
}
