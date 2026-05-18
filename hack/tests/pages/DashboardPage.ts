import type { Locator, Page } from '@playwright/test';

export class DashboardPage {
  constructor(private page: Page) {}

  get analyticsPage(): Locator {
    return this.page.getByTestId('dashboard-analytics-page');
  }

  get workspacePage(): Locator {
    return this.page.getByTestId('dashboard-workspace-page');
  }

  get workspaceDescription(): Locator {
    return this.page.getByTestId('dashboard-workspace-description');
  }

  get workspaceProjects(): Locator {
    return this.page.getByTestId('dashboard-workspace-projects');
  }

  get workspaceQuickNav(): Locator {
    return this.page.getByTestId('dashboard-workspace-quick-nav');
  }

  get workspaceTodos(): Locator {
    return this.page.getByTestId('dashboard-workspace-todos');
  }

  get workspaceTrends(): Locator {
    return this.page.getByTestId('dashboard-workspace-trends');
  }

  analyticsMetric(title: string): Locator {
    return this.page.getByTestId('dashboard-analytics-overview').getByText(title, { exact: true });
  }

  analyticsTab(label: string): Locator {
    return this.page.getByTestId('dashboard-analytics-tabs').getByText(label, { exact: true });
  }

  analyticsCardTitle(title: string): Locator {
    return this.page.getByRole('heading', { name: title }).first();
  }

  workspaceQuickNavItem(title: string): Locator {
    return this.workspaceQuickNav.getByText(title, { exact: true }).first();
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

  async clickWorkspaceQuickNav(title: string) {
    await this.workspaceQuickNavItem(title).click();
    await this.page.waitForLoadState('networkidle');
  }
}
