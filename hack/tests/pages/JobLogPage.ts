import type { Locator, Page } from "@playwright/test";

import {
  waitForConfirmOverlay,
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
      .getByRole("combobox", { name: "任务名称", exact: true })
      .first();
    await combobox.click();
    await this.page.getByText(jobName, { exact: true }).last().click();
  }

  async clickSearch() {
    await this.page
      .getByRole("button", { name: /搜\s*索/ })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  async openFirstDetail() {
    await this.page.locator('[data-testid^="job-log-detail-"]').first().click();
    await waitForDialogReady(this.dialog);
    await this.dialog
      .getByText("任务名称", { exact: true })
      .waitFor({ state: "visible" });
  }

  async clearLogs() {
    if (await this.dialog.isVisible().catch(() => false)) {
      await this.page.keyboard.press("Escape");
      await this.dialog.waitFor({ state: "hidden" });
    }
    await this.page.getByTestId("job-log-clear").click();
    await this.confirmPopconfirm();
    await waitForRouteReady(this.page);
  }

  async selectFirstRow() {
    if (await this.dialog.isVisible().catch(() => false)) {
      await this.page.keyboard.press("Escape");
      await this.dialog.waitFor({ state: "hidden" });
    }
    const checkbox = this.page
      .locator(".vxe-table--body .vxe-checkbox--icon")
      .first();
    await checkbox.click();
  }

  async deleteSelectedLogs() {
    if (await this.dialog.isVisible().catch(() => false)) {
      await this.page.keyboard.press("Escape");
      await this.dialog.waitFor({ state: "hidden" });
    }
    await this.page.getByTestId("job-log-delete").click();
    await this.confirmPopconfirm();
    await waitForRouteReady(this.page);
  }

  async getVisibleRowCount() {
    return this.page.locator(".vxe-body--row").count();
  }

  async detailContains(text: string) {
    return this.dialog.getByText(text).first().isVisible();
  }

  private async confirmPopconfirm() {
    const overlay = await waitForConfirmOverlay(this.page);
    const confirm = overlay
      .getByRole("button", { name: /确\s*定|OK|是/i })
      .last();
    await confirm.click();
  }
}
