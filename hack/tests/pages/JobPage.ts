import type { Locator, Page } from "@playwright/test";

export class JobPage {
  constructor(private page: Page) {}

  private get dialog(): Locator {
    return this.page.locator('[role="dialog"]').last();
  }

  async goto() {
    await this.page.goto("/system/job");
    await this.page.waitForLoadState("networkidle");
    await this.page.getByTestId("job-page").waitFor({ state: "visible" });
  }

  async fillSearchKeyword(keyword: string) {
    const input = this.page.getByLabel("关键字", { exact: true }).first();
    await input.clear();
    await input.fill(keyword);
  }

  async clickSearch() {
    await this.page
      .getByRole("button", { name: /搜\s*索/ })
      .first()
      .click();
    await this.page.waitForLoadState("networkidle");
    await this.page.waitForTimeout(300);
  }

  async openCreate() {
    await this.page.getByTestId("job-add").click();
    await this.dialog.waitFor({ state: "visible" });
    await this.dialog.getByLabel("任务名称").waitFor({ state: "visible" });
  }

  async openEditSearchedJob() {
    await this.page.locator('[data-testid^="job-edit-"]').first().click();
    await this.dialog.waitFor({ state: "visible" });
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
      const timezone = this.dialog
        .getByPlaceholder("可选择常用时区或输入自定义 IANA 时区")
        .first();
      await timezone.waitFor({ state: "visible" });
      await this.replaceFieldValue(timezone, params.timezone);
    }
  }

  async selectTaskTab(tabName: "Handler" | "Shell") {
    await this.dialog.getByRole("tab", { name: tabName, exact: true }).click();
    await this.page.waitForTimeout(200);
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
    await this.page.waitForLoadState("networkidle");
    await this.page.waitForTimeout(300);
  }

  async closeDialog() {
    await this.dialog.getByRole("button", { name: /取\s*消|关\s*闭/ }).click();
    await this.page.waitForTimeout(300);
  }

  async deleteSearchedJob() {
    await this.page.getByRole("button", { name: "更多" }).first().click();
    await this.page.locator('[data-testid^="job-delete-"]').first().click();
    await this.confirmPopconfirm();
    await this.page.waitForLoadState("networkidle");
    await this.page.waitForTimeout(300);
  }

  async isPausedByPluginVisible() {
    return this.page
      .getByText("插件处理器不可用", { exact: true })
      .first()
      .isVisible();
  }

  async isActionDisabled(prefix: "job-enable-" | "job-trigger-") {
    return this.page.locator(`[data-testid^="${prefix}"]`).first().isDisabled();
  }

  async triggerSearchedJob() {
    await this.page.locator('[data-testid^="job-trigger-"]').first().click();
    await this.page.waitForLoadState("networkidle");
    await this.page.waitForTimeout(300);
  }

  async hoverPausedStatusTag() {
    await this.page
      .getByText("插件处理器不可用", { exact: true })
      .first()
      .hover();
    await this.page.waitForTimeout(150);
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
    const labelNode = this.getFieldLabel(label);
    const trigger = labelNode
      .locator('svg, .anticon, [class*="question"], [data-grace-area-trigger]')
      .first();
    await trigger.waitFor({ state: "visible" });
    await trigger.hover();
    await this.page.waitForTimeout(400);
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
        fieldCount: node.children.length,
        fontFamily: style.fontFamily,
        text: node.textContent?.replace(/\s+/g, " ").trim() ?? "",
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

  private async confirmPopconfirm() {
    const popconfirm = this.page
      .locator(".ant-popconfirm, .ant-popover")
      .last();
    await popconfirm.getByRole("button", { name: /确\s*定|OK|是/i }).click();
  }

  private async replaceFieldValue(target: Locator, value: string) {
    await target.click();
    await target.press("Meta+A").catch(() => target.press("Control+A"));
    await target.fill(value);
  }

  private findFieldItem(label: string) {
    return this.getFieldLabel(label).locator(
      'xpath=ancestor::*[contains(@class, "ant-form-item")][1]',
    );
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
