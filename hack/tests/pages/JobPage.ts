import type { Locator, Page } from "@playwright/test";

import {
  waitForBusyIndicatorsToClear,
  waitForConfirmOverlay,
  waitForDialogReady,
  waitForRouteReady,
  waitForTableReady,
} from "../support/ui";

export class JobPage {
  constructor(private page: Page) {}

  private get dialog(): Locator {
    return this.page
      .locator('[role="dialog"]')
      .filter({ hasText: /新增任务|编辑任务|任务详情/ })
      .last();
  }

  async goto() {
    await this.page.goto("/system/job");
    await waitForTableReady(this.page, '[data-testid="job-page"]');
  }

  async fillSearchKeyword(keyword: string) {
    const input = this.page.getByLabel(/关键字|Keyword/i).first();
    await input.clear();
    await input.fill(keyword);
  }

  async clickSearch() {
    await this.page
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  async openCreate() {
    await this.page.getByTestId("job-add").click();
    await waitForDialogReady(this.dialog);
    await this.dialog.getByLabel("任务名称").waitFor({ state: "visible" });
  }

  async openEditSearchedJob() {
    await this.page.locator('[data-testid^="job-edit-"]').first().click();
    await waitForDialogReady(this.dialog);
    await this.dialog.getByLabel("任务名称").waitFor({ state: "visible" });
  }

  async selectCombobox(label: string, optionText: string) {
    const combobox = this.dialog
      .getByRole("combobox", { name: new RegExp(label) })
      .first();
    const select = combobox.locator(
      'xpath=ancestor::*[contains(@class, "ant-select-selector")][1]',
    );
    await combobox.waitFor({ state: "visible" });
    if ((await select.textContent())?.includes(optionText)) {
      return;
    }
    await select.click();
    await this.page.getByText(optionText, { exact: true }).last().click();
  }

  async fillCommonFields(params: {
    groupName?: string;
    name?: string;
    description?: string;
    cronExpr?: string;
    timezone?: string;
  }) {
    if (params.groupName) {
      await this.selectCombobox("所属分组", params.groupName);
    }
    if (params.name) {
      await this.dialog.getByLabel("任务名称").fill(params.name);
    }
    if (params.description !== undefined) {
      const description = this.dialog.getByLabel("任务描述").first();
      await this.replaceFieldValue(description, params.description);
    }
    if (params.cronExpr) {
      const cron = this.dialog.getByLabel("定时表达式");
      await cron.clear();
      await cron.fill(params.cronExpr);
    }
    if (params.timezone) {
      const timezoneLabel = this.getFieldLabel("时区");
      const timezoneFieldId = await timezoneLabel.getAttribute("for");
      if (!timezoneFieldId) {
        throw new Error("未找到时区字段的控件标识");
      }
      const timezone = this.dialog.locator(`#${timezoneFieldId}`).first();
      await timezone.waitFor({ state: "visible", timeout: 5000 });
      await this.replaceFieldValue(timezone, params.timezone);
    }
  }

  async selectTaskTab(tabName: "Handler" | "Shell") {
    const tab = this.dialog.getByRole("tab", { name: tabName, exact: true });
    await tab.click();
    await waitForBusyIndicatorsToClear(this.dialog);
  }

  async isShellTabVisible() {
    return this.dialog
      .getByRole("tab", { name: "Shell", exact: true })
      .isVisible({ timeout: 1500 })
      .catch(() => false);
  }

  async selectHandler(optionText: string) {
    await this.selectCombobox("任务处理器", optionText);
  }

  async fillHandlerParam(label: string, value: string) {
    const target = this.dialog.getByLabel(new RegExp(label)).first();
    await target.waitFor({ state: "visible" });
    await this.replaceFieldValue(target, value);
  }

  async save() {
    await this.dialog.getByRole("button", { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
    const closed = await this.dialog
      .waitFor({ state: "hidden", timeout: 1500 })
      .then(() => true)
      .catch(() => false);
    if (!closed) {
      await waitForBusyIndicatorsToClear(this.dialog);
    }
  }

  async closeDialog() {
    const footerClose = this.dialog.getByRole("button", {
      name: /取\s*消|关\s*闭/,
    });
    if (await footerClose.count()) {
      await footerClose.click();
    } else {
      const iconClose = this.dialog.locator(".ant-modal-close").first();
      if (await iconClose.count()) {
        await iconClose.click();
      } else {
        await this.page.keyboard.press("Escape");
      }
    }
    await this.dialog.waitFor({ state: "hidden", timeout: 5000 }).catch(() => {});
    await waitForBusyIndicatorsToClear(this.page);
  }

  async deleteSearchedJob() {
    await this.page.locator('[data-testid^="job-more-"]').first().click();
    await this.page.locator('[data-testid^="job-delete-"]').first().click();
    await this.confirmPopconfirm();
    await waitForRouteReady(this.page);
  }

  async setTaskStatus(statusLabel: "停用" | "启用") {
    const option = this.dialog
      .locator(".ant-radio-button-wrapper")
      .filter({ hasText: statusLabel })
      .first();
    await option.waitFor({ state: "visible" });
    await option.click();
    await option.waitFor({ state: "visible" });
  }

  async hasSearchedJobMoreButton() {
    return this.page
      .locator('[data-testid^="job-more-"]')
      .first()
      .isVisible({ timeout: 1500 })
      .catch(() => false);
  }

  async isPausedByPluginVisible() {
    return this.page
      .getByText("不可用", { exact: true })
      .first()
      .isVisible();
  }

  async isActionDisabled(prefix: "job-enable-" | "job-trigger-") {
    return this.page
      .locator(`[data-testid^="${prefix}"]`)
      .first()
      .isDisabled()
      .catch(() => false);
  }

  async hasAction(prefix: "job-enable-" | "job-more-" | "job-trigger-") {
    return this.page
      .locator(`[data-testid^="${prefix}"]`)
      .first()
      .isVisible({ timeout: 1500 })
      .catch(() => false);
  }

  async triggerSearchedJob() {
    await this.openTriggerConfirmForSearchedJob();
    await this.confirmTriggerConfirm();
  }

  async openTriggerConfirmForSearchedJob() {
    await this.page.locator('[data-testid^="job-trigger-"]').first().click();
    return waitForConfirmOverlay(this.page);
  }

  async cancelTriggerConfirm() {
    const popconfirm = await waitForConfirmOverlay(this.page);
    await popconfirm
      .getByRole("button", { name: /取\s*消|Cancel/i })
      .click();
    await popconfirm.waitFor({ state: "hidden", timeout: 5000 }).catch(() => {});
  }

  async confirmTriggerConfirm() {
    const popconfirm = await waitForConfirmOverlay(this.page);
    await popconfirm
      .getByRole("button", { name: /确\s*认|确\s*定|OK|Yes|Confirm/i })
      .last()
      .click();
    await waitForRouteReady(this.page);
  }

  async hoverPausedStatusTag() {
    await this.page
      .getByText("不可用", { exact: true })
      .first()
      .hover();
    await this.page
      .locator(
        '[role="tooltip"]:visible, .side-content:visible, .ant-tooltip:visible, .ant-popover:visible',
      )
      .first()
      .waitFor({ state: "visible", timeout: 1500 })
      .catch(() => {});
  }

  async hasJob(name: string) {
    return this.page
      .locator(".vxe-body--row", { hasText: name })
      .first()
      .isVisible({ timeout: 3000 })
      .catch(() => false);
  }

  async getJobRowText(name: string) {
    const row = this.page.locator(".vxe-body--row", { hasText: name }).first();
    await row.waitFor({ state: "visible" });
    return (await row.textContent())?.replace(/\s+/g, " ").trim() ?? "";
  }

  async hoverFieldHelp(label: string) {
    let lastError: unknown;
    for (let attempt = 0; attempt < 3; attempt += 1) {
      const labelNode = this.getFieldLabel(label);
      const trigger = labelNode
        .locator('svg, .anticon, [class*="question"], [data-grace-area-trigger]')
        .first();
      try {
        await trigger.waitFor({ state: "visible", timeout: 5000 });
        await trigger.scrollIntoViewIfNeeded();
        await trigger.hover({ force: true, timeout: 5000 });
        break;
      } catch (error) {
        lastError = error;
        if (attempt === 2) {
          throw lastError;
        }
        await waitForRouteReady(this.page);
      }
    }
    await this.page
      .locator(
        '[role="tooltip"]:visible, .side-content:visible, .ant-tooltip:visible, .ant-popover:visible',
      )
      .first()
      .waitFor({ state: "visible", timeout: 1500 })
      .catch(() => {});
  }

  async isTooltipVisible(text: string) {
    return this.page
      .locator(
        '[role="tooltip"]:visible, .side-content:visible, .ant-tooltip:visible, .ant-popover:visible',
      )
      .getByText(text)
      .first()
      .isVisible({ timeout: 1500 })
      .catch(() => false);
  }

  async getCronDisplayMetrics(jobId: number) {
    const display = this.page.getByTestId(`job-cron-expr-${jobId}`);
    await display.waitFor({ state: "visible" });
    return display.evaluate((node) => {
      const style = getComputedStyle(node);
      return {
        backgroundColor: style.backgroundColor,
        borderColor: style.borderColor,
        color: style.color,
        fieldCount: node.children.length,
        fontFamily: style.fontFamily,
        text: node.textContent?.replace(/\s+/g, " ").trim() ?? "",
      };
    });
  }

  async getCronEditorMetrics() {
    const editor = this.dialog.getByLabel("定时表达式").first();
    await editor.waitFor({ state: "visible" });
    return editor.evaluate((node) => {
      const style = getComputedStyle(node);
      return {
        backgroundColor: style.backgroundColor,
        borderRadius: style.borderRadius,
        fontFamily: style.fontFamily,
      };
    });
  }

  async getElementVerticalPadding(testId: string) {
    const target = this.page.getByTestId(testId);
    await target.waitFor({ state: "visible" });
    return target.evaluate((node) => {
      const style = getComputedStyle(node);
      return {
        paddingBottom: Number.parseFloat(style.paddingBottom),
        paddingTop: Number.parseFloat(style.paddingTop),
      };
    });
  }

  messageNotice(text: string) {
    return this.page.locator(".ant-message-notice").filter({ hasText: text }).last();
  }

  private async confirmPopconfirm() {
    const popconfirm = await waitForConfirmOverlay(this.page);
    await popconfirm.getByRole("button", { name: /确\s*定|OK|是/i }).click();
  }

  private async replaceFieldValue(target: Locator, value: string) {
    await target.click();
    await target.press("Meta+A").catch(() => target.press("Control+A"));
    await target.fill(value);
  }

  private getFieldLabel(label: string) {
    return this.dialog
      .locator("label")
      .filter({ hasText: new RegExp(this.escapeForRegex(label)) })
      .first();
  }

  private escapeForRegex(value: string) {
    return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&").replace(/\s+/g, "\\s*");
  }
}
