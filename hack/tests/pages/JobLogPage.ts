import type { Locator, Page } from '@playwright/test';

export class JobLogPage {
  constructor(private page: Page) {}

  private get dialog(): Locator {
    return this.page.locator('[role="dialog"]').last();
  }

  async goto() {
    await this.page.goto('/system/job-log');
    await this.page.waitForLoadState('networkidle');
    await this.page.getByTestId('job-log-page').waitFor({ state: 'visible' });
  }

  async selectJob(jobName: string) {
    const combobox = this.page.getByRole('combobox', { name: '任务名称', exact: true }).first();
    await combobox.click();
    await this.page.getByText(jobName, { exact: true }).last().click();
  }

  async clickSearch() {
    await this.page.getByRole('button', { name: /搜\s*索/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(300);
  }

  async openFirstDetail() {
    await this.page.locator('[data-testid^="job-log-detail-"]').first().click();
    await this.dialog.waitFor({ state: 'visible' });
    await this.dialog.getByText('任务名称', { exact: true }).waitFor({ state: 'visible' });
  }

  async clearLogs() {
    if (await this.dialog.isVisible().catch(() => false)) {
      await this.page.keyboard.press('Escape');
      await this.dialog.waitFor({ state: 'hidden' });
    }
    await this.page.getByTestId('job-log-clear').click();
    await this.confirmPopconfirm();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(300);
  }

  async getVisibleRowCount() {
    return this.page.locator('.vxe-body--row').count();
  }

  async detailContains(text: string) {
    return this.dialog.getByText(text).first().isVisible();
  }

  private async confirmPopconfirm() {
    const modal = this.page.locator('.ant-modal-confirm, .ant-popconfirm, .ant-popover').last();
    const confirm = modal.getByRole('button', { name: /确\s*定|OK|是/i }).last();
    await confirm.click();
  }
}
