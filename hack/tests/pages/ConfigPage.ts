import type { Locator, Page } from "@playwright/test";

import {
  waitForBusyIndicatorsToClear,
  waitForConfirmOverlay,
  waitForDialogReady,
  waitForDropdown,
  waitForRouteReady,
  waitForTableReady,
} from "../support/ui";

export class ConfigPage {
  constructor(private page: Page) {}

  private escapeRegex(value: string) {
    return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  }

  private localizedLabelPattern(label: string) {
    const labelMap: Record<string, RegExp> = {
      参数名称: /参数名称|Parameter Name/i,
      参数键名: /参数键名|Parameter Key/i,
      参数键值: /参数键值|Parameter Value/i,
      参数类型: /参数类型|Value Type/i,
      选项列表: /选项列表|Options/i,
      备注: /备注|Remark/i,
    };
    return (
      labelMap[label] ?? new RegExp(`^\\s*${this.escapeRegex(label)}\\s*$`)
    );
  }

  private resolveLocalizedLabel(scope: Page | Locator, label: string) {
    return scope.getByLabel(this.localizedLabelPattern(label)).first();
  }

  private searchFieldName(label: string) {
    const fieldMap: Record<string, string> = {
      参数名称: "name",
      参数键名: "key",
    };
    return fieldMap[label];
  }

  private get searchForm() {
    return this.page.locator(".vxe-grid--form-wrapper").first();
  }

  private async fillInputAndWaitForStableValue(input: Locator, value: string) {
    const deadline = Date.now() + 5000;
    while (Date.now() < deadline) {
      await input.waitFor({ state: "visible", timeout: 2000 });
      await input.clear();
      await input.fill(value);
      await this.page.waitForTimeout(600);
      await waitForBusyIndicatorsToClear(this.page, 2000);
      if ((await input.inputValue().catch(() => "")) === value) {
        return;
      }
    }

    await input.clear();
    await input.fill(value);
  }

  /** The modal dialog container */
  private get dialog() {
    return this.page.locator('[role="dialog"]');
  }

  private get builtinDeleteTooltip() {
    return this.page
      .locator(
        '[role="tooltip"]:visible, .ant-tooltip:visible, .ant-popover:visible',
      )
      .filter({
        hasText:
          /System built-in data cannot be deleted|系统内置数据不支持删除|系統內置數據不支援刪除/,
      })
      .first();
  }

  private getDeleteActionById(id: number) {
    return this.page.locator(`[data-testid="config-delete-${id}"]`).first();
  }

  private getRowByDeleteActionId(id: number) {
    return this.page
      .locator(".vxe-body--row")
      .filter({ has: this.getDeleteActionById(id) })
      .first();
  }

  private async hoverDeleteButtonInRow(row: Locator) {
    const button = row.getByRole("button", { name: /删\s*除|Delete/i }).first();
    await button.waitFor({ state: "visible", timeout: 5000 });
    await button.locator("xpath=..").hover({ force: true });
  }

  async goto() {
    await this.page.goto("/system/config");
    await waitForTableReady(this.page);
  }

  // ========== CRUD operations ==========

  async create(name: string, key: string, value: string, remark?: string) {
    await this.page.getByRole("button", { name: /新\s*增/ }).click();
    await waitForDialogReady(this.dialog);

    await this.fillDialogField("参数名称", name);
    await this.fillDialogField("参数键名", key);
    await this.fillDialogField("参数键值", value);
    if (remark) {
      await this.fillDialogField("备注", remark);
    }

    await this.dialog.getByRole("button", { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
    await this.dialog
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
  }

  /**
   * Create a select-typed parameter with simple-line options and choose one option value.
   */
  async createSelect(
    name: string,
    key: string,
    optionsText: string,
    optionLabel: string,
    remark?: string,
  ) {
    await this.page.getByRole("button", { name: /新\s*增/ }).click();
    await waitForDialogReady(this.dialog);

    await this.fillDialogField("参数名称", name);
    await this.fillDialogField("参数键名", key);
    await this.selectDialogOption("参数类型", /下拉单选|Select/i);
    await this.fillDialogField("选项列表", optionsText);
    await this.page.waitForTimeout(300);
    await this.selectDialogOption("参数键值", optionLabel);
    if (remark) {
      await this.fillDialogField("备注", remark);
    }

    await this.dialog.getByRole("button", { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
    await this.dialog
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
  }

  async openEditByKey(key: string) {
    await this.fillSearchField("参数键名", key);
    await this.clickSearch();
    const editButton = await this.getEditButtonByKey(key);
    await editButton.click();
    await waitForDialogReady(this.dialog);
  }

  async chooseBooleanValue(value: "true" | "false") {
    const label = value === "true" ? /是|True/i : /否|False/i;
    // Ant Design button radios hide the native input; click the visible wrapper.
    const wrapper = this.dialog
      .locator(".ant-radio-button-wrapper")
      .filter({ hasText: label })
      .first();
    if (await wrapper.isVisible().catch(() => false)) {
      await wrapper.click();
      return;
    }
    await this.dialog.getByRole("radio", { name: label }).click({ force: true });
  }

  /**
   * Visible boolean option control (button-style Radio hides the native input).
   */
  booleanOption(label: RegExp | string) {
    const pattern =
      typeof label === "string"
        ? new RegExp(this.escapeRegex(label), "i")
        : label;
    return this.dialog
      .locator(".ant-radio-button-wrapper")
      .filter({ hasText: pattern })
      .first();
  }

  async selectDialogOption(fieldLabel: string, option: string | RegExp) {
    // Vben form rows may not always expose .ant-form-item; cover both layouts.
    const item = this.dialog
      .locator(".ant-form-item, .form-is-required, .form-valid-error, .flex.flex-col, [class*='form-item']")
      .filter({ hasText: this.localizedLabelPattern(fieldLabel) })
      .first();
    const selector = item.locator(".ant-select").first();
    await selector.click();
    const dropdown = await waitForDropdown(this.page);
    const optionPattern =
      typeof option === "string"
        ? new RegExp(this.escapeRegex(option), "i")
        : option;
    // Ant Select options can sit in a portal; prefer role + evaluate click when
    // the portal is clipped by modal overflow (outside viewport).
    const optionLoc = dropdown
      .locator(".ant-select-item-option, [role='option']")
      .filter({ hasText: optionPattern })
      .first();
    await optionLoc.waitFor({ state: "attached", timeout: 5000 });
    await optionLoc.evaluate((el: HTMLElement) => {
      el.scrollIntoView({ block: "nearest", inline: "nearest" });
      el.click();
    });
  }

  /**
   * Vben FormItem row (shadcn FormItem). Prefer form-valid-error after validate;
   * form-is-required covers empty required fields before/after validate.
   */
  dialogFieldItem(label: string): Locator {
    const pattern = this.localizedLabelPattern(label);
    return this.dialog
      .locator(".form-valid-error, .form-is-required")
      .filter({ hasText: pattern })
      .first();
  }

  async openCreateDialog() {
    await this.page.getByRole("button", { name: /新\s*增/ }).click();
    await waitForDialogReady(this.dialog);
  }

  /** Root of the create/edit parameter dialog. */
  get createEditDialog() {
    return this.dialog;
  }

  /**
   * Select a value type on the create form (labels are bilingual).
   * Triggers schema remount and modal density layout refresh.
   */
  async selectValueType(option: string | RegExp) {
    await this.selectDialogOption("参数类型", option);
    await this.page.waitForTimeout(300);
  }

  /** Richtext editor shell inside the open dialog. */
  get richtextEditor() {
    return this.dialog.getByTestId("tiptap-editor");
  }

  get richtextEditorContent() {
    return this.dialog.getByTestId("tiptap-editor-content");
  }

  /** Fullscreen toggle in the Vben modal header (visible for spacious types). */
  get dialogFullscreenButton() {
    return this.dialog.locator("button.absolute.top-3.right-10");
  }

  /**
   * Click confirm without waiting for dialog close — used for validation failures.
   */
  async clickDialogConfirm() {
    await this.dialog.getByRole("button", { name: /确\s*认/ }).click();
  }

  async fillDialogField(label: string, value: string) {
    // Prefer accessible label association from Vben FormControl id/for.
    const byLabel = this.resolveLocalizedLabel(this.dialog, label);
    if (await byLabel.isVisible().catch(() => false)) {
      const tag = await byLabel.evaluate((el) => el.tagName.toLowerCase());
      if (tag === "textarea" || tag === "input") {
        await byLabel.clear();
        await byLabel.fill(value);
        return;
      }
    }

    const item = this.dialog
      .locator(".ant-form-item, .form-is-required, .form-valid-error")
      .filter({ hasText: this.localizedLabelPattern(label) })
      .first();
    const textarea = item.locator("textarea").first();
    if (await textarea.isVisible().catch(() => false)) {
      await textarea.clear();
      await textarea.fill(value);
      return;
    }
    const input = item.locator("input").first();
    await input.clear();
    await input.fill(value);
  }

  /**
   * Validation error message under a dialog field (vee-validate FormMessage).
   */
  dialogFieldError(label: string): Locator {
    return this.dialogFieldItem(label).locator("p.text-destructive").first();
  }

  dialogFieldControl(label: string): Locator {
    const byLabel = this.resolveLocalizedLabel(this.dialog, label);
    return byLabel;
  }

  async confirmDialog() {
    await this.dialog.getByRole("button", { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
    await this.dialog
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
  }

  async edit(
    configName: string,
    fields: { name?: string; key?: string; value?: string; remark?: string },
  ) {
    // Search for the config first to narrow results
    await this.fillSearchField("参数名称", configName);
    await this.clickSearch();

    // Click edit button
    await this.page
      .locator(".ant-btn-sm")
      .filter({ hasText: /编\s*辑/ })
      .first()
      .click();
    await waitForDialogReady(this.dialog);

    if (fields.name) {
      await this.fillDialogField("参数名称", fields.name);
    }
    if (fields.key) {
      await this.fillDialogField("参数键名", fields.key);
    }
    if (fields.value) {
      await this.fillDialogField("参数键值", fields.value);
    }
    if (fields.remark) {
      await this.fillDialogField("备注", fields.remark);
    }

    await this.dialog.getByRole("button", { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
    await this.dialog
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
  }

  async delete(configName: string) {
    // Search for the config first
    await this.fillSearchField("参数名称", configName);
    await this.clickSearch();

    const targetRow = this.page
      .locator(".vxe-body--row")
      .filter({ hasText: configName })
      .first();
    await targetRow.waitFor({ state: "visible", timeout: 5000 });
    await targetRow.hover();

    // Click delete button
    const rowDeleteButton = targetRow
      .locator(".ant-btn-sm")
      .filter({ hasText: /删\s*除/ })
      .first();
    if (await rowDeleteButton.isVisible().catch(() => false)) {
      await rowDeleteButton.click();
    } else {
      await this.page
        .locator(".ant-btn-sm")
        .filter({ hasText: /删\s*除/ })
        .first()
        .click();
    }

    // Confirm deletion in Popconfirm
    const popconfirm = await waitForConfirmOverlay(this.page);
    const confirmBtn = popconfirm.getByRole("button", {
      name: /确\s*定|OK|是/i,
    });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      const modal = this.page.locator(".ant-modal-confirm");
      await modal.getByRole("button", { name: /确\s*定|OK/i }).click();
    }

    await waitForRouteReady(this.page);
  }

  async hasConfig(text: string): Promise<boolean> {
    return this.page
      .locator(".vxe-body--row")
      .filter({ hasText: text })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  findRowByExactKey(key: string): Locator {
    const keyPattern = new RegExp(`^\\s*${this.escapeRegex(key)}\\s*$`);

    return this.page
      .locator(".vxe-body--row", {
        has: this.page.locator(".vxe-cell").filter({ hasText: keyPattern }),
      })
      .first();
  }

  async getDeleteButtonByKey(key: string): Promise<Locator> {
    const row = this.findRowByExactKey(key);
    await row.waitFor({ state: "visible", timeout: 5000 });
    return row.getByRole("button", { name: /删\s*除|Delete/i }).first();
  }

  async getEditButtonByKey(key: string): Promise<Locator> {
    const row = this.findRowByExactKey(key);
    await row.waitFor({ state: "visible", timeout: 5000 });
    return row.getByRole("button", { name: /编\s*辑|Edit/i }).first();
  }

  async getDeleteButtonById(id: number): Promise<Locator> {
    const row = this.getRowByDeleteActionId(id);
    await row.waitFor({ state: "visible", timeout: 5000 });
    return row.getByRole("button", { name: /删\s*除|Delete/i }).first();
  }

  async getEditButtonById(id: number): Promise<Locator> {
    const row = this.getRowByDeleteActionId(id);
    await row.waitFor({ state: "visible", timeout: 5000 });
    return row.getByRole("button", { name: /编\s*辑|Edit/i }).first();
  }

  async hoverDeleteActionById(id: number) {
    const row = this.getRowByDeleteActionId(id);
    await row.waitFor({ state: "visible", timeout: 5000 });
    await this.hoverDeleteButtonInRow(row);
  }

  async hoverDeleteActionByKey(key: string) {
    const row = this.findRowByExactKey(key);
    await row.waitFor({ state: "visible", timeout: 5000 });
    await this.hoverDeleteButtonInRow(row);
  }

  async isBuiltinDeleteTooltipVisible() {
    return this.builtinDeleteTooltip
      .waitFor({ state: "visible", timeout: 3000 })
      .then(() => true)
      .catch(() => false);
  }

  async getRowCount(): Promise<number> {
    return this.page.locator(".vxe-body--row").count();
  }

  /**
   * Resolve the header cell for a list column by localized title, then return
   * whether the header and first body cell are left-aligned (vxe col--left).
   */
  async getColumnAlignment(headerLabel: string): Promise<{
    headerLeft: boolean;
    bodyLeft: boolean;
  }> {
    const headerCell = this.page
      .locator(".vxe-header--column")
      .filter({ hasText: this.localizedLabelPattern(headerLabel) })
      .first();
    await headerCell.waitFor({ state: "visible", timeout: 5000 });

    const headerLeft = await headerCell
      .evaluate((el) => el.classList.contains("col--left"))
      .catch(() => false);

    const colIdClass = await headerCell.evaluate((el) => {
      const classes = Array.from(el.classList);
      return (
        classes.find((name) => /^col_[A-Za-z0-9]+$/.test(name)) ??
        classes.find((name) => name.startsWith("col_") && !name.startsWith("col--")) ??
        ""
      );
    });

    if (!colIdClass) {
      return { headerLeft, bodyLeft: false };
    }

    const bodyCell = this.page
      .locator(`.vxe-body--row .vxe-body--column.${colIdClass}`)
      .first();
    const bodyVisible = await bodyCell
      .isVisible({ timeout: 5000 })
      .catch(() => false);
    if (!bodyVisible) {
      return { headerLeft, bodyLeft: false };
    }

    const bodyLeft = await bodyCell.evaluate((el) =>
      el.classList.contains("col--left"),
    );
    return { headerLeft, bodyLeft };
  }

  // ========== Search helpers ==========

  async fillSearchField(label: string, value: string) {
    const fieldName = this.searchFieldName(label);
    if (fieldName) {
      const namedInput = this.searchForm
        .locator(`input[name="${fieldName}"]`)
        .first();
      if (await namedInput.isVisible().catch(() => false)) {
        await this.fillInputAndWaitForStableValue(namedInput, value);
        return;
      }
    }

    const input = this.searchForm
      .locator(".ant-form-item")
      .filter({ hasText: this.localizedLabelPattern(label) })
      .locator("input")
      .first();
    if (await input.isVisible().catch(() => false)) {
      await this.fillInputAndWaitForStableValue(input, value);
      return;
    }
    const fallbackInput = this.resolveLocalizedLabel(this.searchForm, label);
    await this.fillInputAndWaitForStableValue(fallbackInput, value);
  }

  async clickSearch() {
    await this.searchForm
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  async clickReset() {
    await this.searchForm
      .getByRole("button", { name: /重\s*置|Reset/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  // ========== Row Selection ==========

  async selectRow(configName: string) {
    await this.fillSearchField("参数名称", configName);
    await this.clickSearch();
    // Click the first checkbox in the body rows
    const checkbox = this.page
      .locator(".vxe-body--row .vxe-checkbox--icon")
      .first();
    await checkbox.click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  // ========== Export ==========

  async clickExport() {
    await this.page.getByRole("button", { name: /导\s*出|Export/i }).click();
    await waitForDialogReady(this.page.locator('[role="dialog"]'));
  }

  /** Click confirm button in the export confirm modal */
  async clickExportConfirm() {
    const modal = this.page.locator('[role="dialog"]');
    await modal.getByRole("button", { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
  }

  // ========== Import ==========

  async clickImport() {
    await this.page.getByRole("button", { name: /导\s*入|Import/i }).click();
    await waitForDialogReady(this.dialog);
  }
}
