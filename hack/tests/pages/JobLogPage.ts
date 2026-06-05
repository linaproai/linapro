import type { Locator, Page } from "@playwright/test";

import {
  waitForDialogReady,
  waitForRouteReady,
  waitForTableReady,
} from "../support/ui";

export class JobLogPage {
  constructor(private page: Page) {}

  private get dialog(): Locator {
    return this.page.locator('[role="dialog"]').last();
  }

  async goto() {
    await this.page.goto("/system/job-log");
    await waitForTableReady(this.page, '[data-testid="job-log-page"]');
  }

  async selectJob(jobName: string) {
    const combobox = this.page
      .getByRole("combobox", { name: /任务名称|Job Name/i })
      .first();
    await combobox.click();
    await this.page.getByText(jobName, { exact: true }).last().click();
  }

  async clickSearch() {
    await this.page
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  async openFirstDetail() {
    await this.page.locator('[data-testid^="job-log-detail-"]').first().click();
    await waitForDialogReady(this.dialog);
    await this.dialog
      .getByText(/任务名称|Job Name/i)
      .waitFor({ state: "visible" });
  }

  async openDeleteDialog() {
    if (await this.dialog.isVisible().catch(() => false)) {
      await this.page.keyboard.press("Escape");
      await this.dialog.waitFor({ state: "hidden" });
    }
    await this.page.getByTestId("job-log-delete").click();
    return waitForDialogReady(
      this.page.locator(".ant-modal-wrap:visible").filter({
        hasText: /删除执行日志|Delete Execution Logs/i,
      }),
    );
  }

  async confirmDeleteDialog(dialog: Locator) {
    await dialog
      .getByRole("button", { name: /确\s*(认|定)|Confirm/i })
      .click();
    await waitForRouteReady(this.page);
  }

  async getVisibleRowCount() {
    return this.page.locator(".vxe-body--row").count();
  }

  async detailContains(text: RegExp | string) {
    const matcher =
      typeof text === 'string'
        ? new RegExp(text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'i')
        : text;
    return this.dialog.getByText(matcher).first().isVisible();
  }
}
