import type { Locator, Page, Response } from "@playwright/test";

import {
  waitForBusyIndicatorsToClear,
  waitForConfirmOverlay,
  waitForDialogReady,
  waitForRouteReady,
  waitForTableReady,
} from "../support/ui";

export class DictPage {
  constructor(private page: Page) {}

  private resolveLocalizedLabel(scope: Locator, label: string) {
    const labelMap: Record<string, RegExp> = {
      字典名称: /字典名称|Dictionary Name/i,
      字典类型: /字典类型|Dictionary Type/i,
      字典标签: /字典标签|Dictionary Label/i,
      数据标签: /数据标签|Data Label/i,
    };
    const localizedLabel = labelMap[label];
    if (localizedLabel) {
      return scope.getByLabel(localizedLabel).first();
    }
    return scope.getByLabel(label, { exact: true }).first();
  }

  /** The modal/drawer dialog container */
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

  /** Left panel: dict type table */
  private get typePanel() {
    return this.page.locator("#dict-type");
  }

  /** Right panel: dict data table */
  private get dataPanel() {
    return this.page.locator("#dict-data");
  }

  async goto() {
    await this.page.goto("/system/dict");
    await waitForTableReady(this.page);
  }

  /**
   * Resolve a dict type row even when the table has been pushed to later pages by
   * previously created test records. Callers can pass either dict name or dict type.
   */
  private async resolveTypeRow(rowText: string) {
    const row = this.typePanel
      .locator(".vxe-body--row", { hasText: rowText })
      .first();
    if (await row.isVisible({ timeout: 1000 }).catch(() => false)) {
      return row;
    }

    // Seed dictionaries are usually referenced by dict type code in tests.
    await this.clickTypeReset();
    await this.fillTypeSearchField("字典类型", rowText);
    await this.clickTypeSearch();
    if (await row.isVisible({ timeout: 1000 }).catch(() => false)) {
      return row;
    }

    // Imported dictionaries are commonly referenced by the display name instead.
    await this.clickTypeReset();
    await this.fillTypeSearchField("字典名称", rowText);
    await this.clickTypeSearch();
    if (await row.isVisible({ timeout: 3000 }).catch(() => false)) {
      return row;
    }

    throw new Error(`Unable to find dict type row for "${rowText}"`);
  }

  private async resolveDataRow(rowText: string) {
    const row = this.dataPanel
      .locator(".vxe-body--row", { hasText: rowText })
      .first();
    await row.waitFor({ state: "visible", timeout: 5000 });
    return row;
  }

  private getTypeDeleteActionById(id: number) {
    return this.typePanel
      .locator(`[data-testid="dict-type-delete-${id}"]`)
      .first();
  }

  private getDataDeleteActionById(id: number) {
    return this.dataPanel
      .locator(`[data-testid="dict-data-delete-${id}"]`)
      .first();
  }

  private getTypeRowByDeleteActionId(id: number) {
    return this.typePanel
      .locator(".vxe-body--row")
      .filter({ has: this.getTypeDeleteActionById(id) })
      .first();
  }

  private getDataRowByDeleteActionId(id: number) {
    return this.dataPanel
      .locator(".vxe-body--row")
      .filter({ has: this.getDataDeleteActionById(id) })
      .first();
  }

  private async hoverDeleteButtonInRow(row: Locator) {
    const button = row.getByRole("button", { name: /删\s*除|Delete/i }).first();
    await button.waitFor({ state: "visible", timeout: 5000 });
    await button.locator("xpath=..").hover({ force: true });
  }

  // ========== Type operations (left panel) ==========

  async createType(name: string, type: string, remark?: string) {
    // Click the "新增" button in the type panel toolbar
    await this.typePanel.getByRole("button", { name: /新\s*增/ }).click();

    // Wait for modal to open
    await waitForDialogReady(this.dialog);

    // Fill form fields - modal form uses labels
    await this.dialog.getByLabel("字典名称").fill(name);
    await this.dialog.getByLabel("字典类型").fill(type);
    if (remark) {
      await this.dialog.getByLabel("备注").fill(remark);
    }

    // Click confirm button
    await this.dialog.getByRole("button", { name: /确\s*认/ }).click();

    await waitForRouteReady(this.page);
  }

  async hasType(typeName: string): Promise<boolean> {
    try {
      await this.resolveTypeRow(typeName);
      return true;
    } catch {
      return false;
    }
  }

  async editType(typeName: string, fields: { name?: string; type?: string }) {
    // Search for the type first to narrow results
    await this.fillTypeSearchField("字典名称", typeName);
    await this.clickTypeSearch();
    await this.resolveTypeRow(typeName);

    // Click edit button (ghost-button in action column)
    await this.typePanel
      .locator(".ant-btn-sm")
      .filter({ hasText: /编\s*辑/ })
      .first()
      .click();

    // Wait for modal to open
    await waitForDialogReady(this.dialog);

    if (fields.name) {
      const nameInput = this.dialog.getByLabel("字典名称");
      await nameInput.clear();
      await nameInput.fill(fields.name);
    }
    if (fields.type) {
      const typeInput = this.dialog.getByLabel("字典类型");
      await typeInput.clear();
      await typeInput.fill(fields.type);
    }

    // Click confirm button
    await this.dialog.getByRole("button", { name: /确\s*认/ }).click();

    await waitForRouteReady(this.page);
  }

  async deleteType(typeName: string): Promise<Response> {
    // Search for the type first
    await this.fillTypeSearchField("字典名称", typeName);
    await this.clickTypeSearch();

    const deletedRow = await this.resolveTypeRow(typeName);
    const deleteButton = deletedRow
      .getByRole("button", { name: /删\s*除|Delete/i })
      .first();

    if (await deleteButton.isVisible({ timeout: 1000 }).catch(() => false)) {
      await deleteButton.click();
    } else {
      // VXE fixed action columns can render outside the primary row tree.
      await this.typePanel
        .locator(".ant-btn-sm")
        .filter({ hasText: /删\s*除|Delete/i })
        .first()
        .click();
    }

    // Confirm the visible modal directly instead of relying on a global DOM query.
    const modal = await waitForConfirmOverlay(this.page);
    const confirmButton = modal
      .getByRole("button", { name: /确\s*定|确\s*认|OK/i })
      .last();
    await confirmButton.waitFor({ state: "visible", timeout: 10000 });

    const [response] = await Promise.all([
      this.page.waitForResponse(
        (candidate) =>
          /\/dict\/type\/[^/?#]+\/?$/.test(
            new URL(candidate.url()).pathname,
          ) && candidate.request().method() === "DELETE",
        { timeout: 30000 },
      ),
      confirmButton.click(),
    ]);
    await waitForRouteReady(this.page);
    await modal.waitFor({ state: "hidden", timeout: 10000 }).catch(() => {});
    await deletedRow
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
    await waitForBusyIndicatorsToClear(this.page);
    return response;
  }

  async clickCurrentTypeDeleteAction(expectedTypeName?: string) {
    if (expectedTypeName) {
      await this.resolveTypeRow(expectedTypeName);
    }
    await this.typePanel
      .locator(".ant-btn-sm")
      .filter({ hasText: /删\s*除/ })
      .first()
      .click();
  }

  async getTypeDeleteButton(rowText: string): Promise<Locator> {
    const row = await this.resolveTypeRow(rowText);
    return row.getByRole("button", { name: /删\s*除|Delete/i }).first();
  }

  async getTypeEditButton(rowText: string): Promise<Locator> {
    const row = await this.resolveTypeRow(rowText);
    return row.getByRole("button", { name: /编\s*辑|Edit/i }).first();
  }

  async getTypeDeleteButtonById(id: number): Promise<Locator> {
    const row = this.getTypeRowByDeleteActionId(id);
    await row.waitFor({ state: "visible", timeout: 5000 });
    return row.getByRole("button", { name: /删\s*除|Delete/i }).first();
  }

  async getTypeEditButtonById(id: number): Promise<Locator> {
    const row = this.getTypeRowByDeleteActionId(id);
    await row.waitFor({ state: "visible", timeout: 5000 });
    return row.getByRole("button", { name: /编\s*辑|Edit/i }).first();
  }

  async hoverTypeDeleteActionById(id: number) {
    const row = this.getTypeRowByDeleteActionId(id);
    await row.waitFor({ state: "visible", timeout: 5000 });
    await this.hoverDeleteButtonInRow(row);
  }

  async hoverTypeDeleteAction(rowText: string) {
    const row = await this.resolveTypeRow(rowText);
    await this.hoverDeleteButtonInRow(row);
  }

  typeHeader(text: RegExp | string): Locator {
    return this.typePanel.locator(".vxe-header--column, th", { hasText: text });
  }

  dataHeader(text: RegExp | string): Locator {
    return this.dataPanel.locator(".vxe-header--column, th", { hasText: text });
  }

  async clickTypeRow(typeName: string) {
    // Click a row in the type panel to load data in the right panel.
    const row = await this.resolveTypeRow(typeName);
    await row.click();
    await waitForRouteReady(this.page);
  }

  // ========== Data operations (right panel) ==========

  async createData(
    label: string,
    value: string,
    opts?: { sort?: number; remark?: string },
  ) {
    // Click "新增" in the data panel toolbar
    await this.dataPanel.getByRole("button", { name: /新\s*增/ }).click();

    // Wait for drawer to open
    await waitForDialogReady(this.dialog);

    // Fill drawer form fields
    await this.dialog.getByLabel("数据标签").fill(label);
    await this.dialog.getByLabel("数据键值").fill(value);
    if (opts?.sort !== undefined) {
      const sortInput = this.dialog.getByLabel("显示排序");
      await sortInput.clear();
      await sortInput.fill(String(opts.sort));
    }
    if (opts?.remark) {
      await this.dialog.getByLabel("备注").fill(opts.remark);
    }

    // Click confirm button
    await this.dialog.getByRole("button", { name: /确\s*认/ }).click();

    await waitForRouteReady(this.page);
  }

  async editData(label: string, fields: { label?: string; value?: string }) {
    // Search for the data label first
    await this.fillDataSearchField("字典标签", label);
    await this.clickDataSearch();

    // Click edit button in data panel
    await this.dataPanel
      .locator(".ant-btn-sm")
      .filter({ hasText: /编\s*辑/ })
      .first()
      .click();

    // Wait for drawer to open
    await waitForDialogReady(this.dialog);

    if (fields.label) {
      const labelInput = this.dialog.getByLabel("数据标签");
      await labelInput.clear();
      await labelInput.fill(fields.label);
    }
    if (fields.value) {
      const valueInput = this.dialog.getByLabel("数据键值");
      await valueInput.clear();
      await valueInput.fill(fields.value);
    }

    // Click confirm button
    await this.dialog.getByRole("button", { name: /确\s*认/ }).click();

    await waitForRouteReady(this.page);
  }

  async deleteData(label: string) {
    // Search for the data label first
    await this.fillDataSearchField("字典标签", label);
    await this.clickDataSearch();

    // Click delete button in data panel
    await this.dataPanel
      .locator(".ant-btn-sm")
      .filter({ hasText: /删\s*除/ })
      .first()
      .click();

    // Try Popconfirm first (more common pattern)
    const popconfirm = await waitForConfirmOverlay(this.page);
    const modal = this.page.locator(".ant-modal-confirm:visible").first();

    const isPopconfirm = await popconfirm
      .isVisible({ timeout: 1000 })
      .catch(() => false);
    const isModal = await modal.isVisible({ timeout: 1000 }).catch(() => false);

    if (isPopconfirm) {
      await popconfirm.getByRole("button", { name: /确\s*定|OK/i }).click();
    } else if (isModal) {
      await modal.getByRole("button", { name: /确\s*定|OK/i }).click();
    } else {
      // Fallback: try clicking any visible confirm button
      await this.page
        .getByRole("button", { name: /确\s*定|OK/i })
        .first()
        .click();
    }

    // Wait for success message
    await waitForRouteReady(this.page);
  }

  async hasData(label: string): Promise<boolean> {
    return this.dataPanel
      .locator(".vxe-body--row", { hasText: label })
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  async getDataDeleteButton(rowText: string): Promise<Locator> {
    const row = await this.resolveDataRow(rowText);
    return row.getByRole("button", { name: /删\s*除|Delete/i }).first();
  }

  async getDataEditButton(rowText: string): Promise<Locator> {
    const row = await this.resolveDataRow(rowText);
    return row.getByRole("button", { name: /编\s*辑|Edit/i }).first();
  }

  async getDataDeleteButtonById(id: number): Promise<Locator> {
    const row = this.getDataRowByDeleteActionId(id);
    await row.waitFor({ state: "visible", timeout: 5000 });
    return row.getByRole("button", { name: /删\s*除|Delete/i }).first();
  }

  async getDataEditButtonById(id: number): Promise<Locator> {
    const row = this.getDataRowByDeleteActionId(id);
    await row.waitFor({ state: "visible", timeout: 5000 });
    return row.getByRole("button", { name: /编\s*辑|Edit/i }).first();
  }

  async hoverDataDeleteActionById(id: number) {
    const row = this.getDataRowByDeleteActionId(id);
    await row.waitFor({ state: "visible", timeout: 5000 });
    await this.hoverDeleteButtonInRow(row);
  }

  async hoverDataDeleteAction(rowText: string) {
    const row = await this.resolveDataRow(rowText);
    await this.hoverDeleteButtonInRow(row);
  }

  async isBuiltinDeleteTooltipVisible() {
    return this.builtinDeleteTooltip
      .waitFor({ state: "visible", timeout: 3000 })
      .then(() => true)
      .catch(() => false);
  }

  // ========== Export ==========

  async clickTypeExport() {
    await this.typePanel.getByRole("button", { name: /导\s*出/ }).click();
    await waitForDialogReady(this.dialog);
  }

  async clickDataExport() {
    await this.dataPanel.getByRole("button", { name: /导\s*出/ }).click();
    await waitForDialogReady(this.dialog);
  }

  /** Click confirm button in the export confirm modal */
  async clickExportConfirm() {
    const modal = this.page.locator('[role="dialog"]');
    await modal.getByRole("button", { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
  }

  // ========== Search helpers ==========

  /** Fill search field in the type panel (left) */
  async fillTypeSearchField(label: string, value: string) {
    const input = this.resolveLocalizedLabel(this.typePanel, label);
    await input.clear();
    await input.fill(value);
  }

  /** Click search button in the type panel */
  async clickTypeSearch() {
    await this.typePanel
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  /** Click reset button in the type panel */
  async clickTypeReset() {
    await this.typePanel
      .getByRole("button", { name: /重\s*置|Reset/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  /** Fill search field in the data panel (right) */
  async fillDataSearchField(label: string, value: string) {
    const input = this.resolveLocalizedLabel(this.dataPanel, label);
    await input.clear();
    await input.fill(value);
  }

  /** Click search button in the data panel */
  async clickDataSearch() {
    await this.dataPanel
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  /** Click reset button in the data panel */
  async clickDataReset() {
    await this.dataPanel
      .getByRole("button", { name: /重\s*置|Reset/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  /** Get visible row count in the data panel */
  async getDataRowCount(): Promise<number> {
    return this.dataPanel.locator(".vxe-body--row").count();
  }

  async getDataActivePage(): Promise<number> {
    const active = this.dataPanel.locator(".vxe-pager--num-btn.is--active");
    await active.first().waitFor({ state: "visible", timeout: 5000 });
    const text = await active.first().textContent();
    const page = Number(text?.trim());
    if (!Number.isFinite(page)) {
      throw new Error(`Unable to resolve active dict data page from "${text}"`);
    }
    return page;
  }

  async gotoDataPage(pageNumber: number) {
    await this.dataPanel
      .locator(".vxe-pager--num-btn")
      .filter({ hasText: new RegExp(`^${pageNumber}$`) })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  /** Get visible row count in the type panel */
  async getTypeRowCount(): Promise<number> {
    return this.typePanel.locator(".vxe-body--row").count();
  }

  /** Select a row checkbox in the type panel by clicking its checkbox */
  async selectTypeRow(index: number = 0) {
    const checkbox = this.typePanel
      .locator(".vxe-body--row .vxe-checkbox--icon")
      .nth(index);
    await checkbox.click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Select a type row by unique visible text before batch actions. */
  async selectTypeRowByText(rowText: string) {
    const row = await this.resolveTypeRow(rowText);
    await row.locator(".vxe-checkbox--icon").first().click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Select a row checkbox in the data panel by clicking its checkbox */
  async selectDataRow(index: number = 0) {
    const checkbox = this.dataPanel
      .locator(".vxe-body--row .vxe-checkbox--icon")
      .nth(index);
    await checkbox.click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  // ========== Import ==========

  /** Click import button in the type panel */
  async clickTypeImport() {
    await this.typePanel.getByRole("button", { name: /导\s*入/ }).click();
    await waitForDialogReady(this.dialog);
  }

  /** Click import button in the data panel */
  async clickDataImport() {
    await this.dataPanel.getByRole("button", { name: /导\s*入/ }).click();
    await waitForDialogReady(this.dialog);
  }
}
