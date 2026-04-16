import type { Page } from '@playwright/test';

export class DictPage {
  constructor(private page: Page) {}

  /** The modal/drawer dialog container */
  private get dialog() {
    return this.page.locator('[role="dialog"]');
  }

  /** Left panel: dict type table */
  private get typePanel() {
    return this.page.locator('#dict-type');
  }

  /** Right panel: dict data table */
  private get dataPanel() {
    return this.page.locator('#dict-data');
  }

  async goto() {
    await this.page.goto('/system/dict');
    await this.page.waitForLoadState('networkidle');
    // Wait for at least one VxeGrid table to render
    await this.page.locator('.vxe-table').first().waitFor({ state: 'visible', timeout: 10000 });
  }

  /**
   * Resolve a dict type row even when the table has been pushed to later pages by
   * previously created test records. Callers can pass either dict name or dict type.
   */
  private async resolveTypeRow(rowText: string) {
    const row = this.typePanel.locator('.vxe-body--row', { hasText: rowText }).first();
    if (await row.isVisible({ timeout: 1000 }).catch(() => false)) {
      return row;
    }

    // Seed dictionaries are usually referenced by dict type code in tests.
    await this.clickTypeReset();
    await this.fillTypeSearchField('字典类型', rowText);
    await this.clickTypeSearch();
    if (await row.isVisible({ timeout: 1000 }).catch(() => false)) {
      return row;
    }

    // Imported dictionaries are commonly referenced by the display name instead.
    await this.clickTypeReset();
    await this.fillTypeSearchField('字典名称', rowText);
    await this.clickTypeSearch();
    if (await row.isVisible({ timeout: 3000 }).catch(() => false)) {
      return row;
    }

    throw new Error(`Unable to find dict type row for "${rowText}"`);
  }

  // ========== Type operations (left panel) ==========

  async createType(name: string, type: string, remark?: string) {
    // Click the "新增" button in the type panel toolbar
    await this.typePanel.getByRole('button', { name: /新\s*增/ }).click();

    // Wait for modal to open
    await this.dialog.waitFor({ state: 'visible', timeout: 5000 });

    // Fill form fields - modal form uses labels
    await this.dialog.getByLabel('字典名称').fill(name);
    await this.dialog.getByLabel('字典类型').fill(type);
    if (remark) {
      await this.dialog.getByLabel('备注').fill(remark);
    }

    // Click confirm button
    await this.dialog.getByRole('button', { name: /确\s*认/ }).click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async hasType(typeName: string): Promise<boolean> {
    return this.typePanel
      .locator('.vxe-body--row', { hasText: typeName })
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  async editType(typeName: string, fields: { name?: string; type?: string }) {
    // Search for the type first to narrow results
    await this.fillTypeSearchField('字典名称', typeName);
    await this.clickTypeSearch();

    // Click edit button (ghost-button in action column)
    await this.typePanel.locator('.ant-btn-sm').filter({ hasText: /编\s*辑/ }).first().click();

    // Wait for modal to open
    await this.dialog.waitFor({ state: 'visible', timeout: 5000 });

    if (fields.name) {
      const nameInput = this.dialog.getByLabel('字典名称');
      await nameInput.clear();
      await nameInput.fill(fields.name);
    }
    if (fields.type) {
      const typeInput = this.dialog.getByLabel('字典类型');
      await typeInput.clear();
      await typeInput.fill(fields.type);
    }

    // Click confirm button
    await this.dialog.getByRole('button', { name: /确\s*认/ }).click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async deleteType(typeName: string) {
    // Search for the type first
    await this.fillTypeSearchField('字典名称', typeName);
    await this.clickTypeSearch();

    const deletedRow = this.typePanel
      .locator('.vxe-body--row', { hasText: typeName })
      .first();

    // Click delete button (ghost-button in action column)
    await this.typePanel.locator('.ant-btn-sm').filter({ hasText: /删\s*除/ }).first().click();

    // Confirm the visible modal directly instead of relying on a global DOM query.
    const modal = this.page.locator('.ant-modal-confirm:visible').last();
    await modal.waitFor({ state: 'visible', timeout: 3000 });
    await modal
      .getByRole('button', { name: /确\s*定|确\s*认|OK/i })
      .last()
      .click({ force: true });

    await this.page.waitForLoadState('networkidle');
    await modal.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
    await deletedRow.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
    await this.page.waitForTimeout(300);
  }

  async clickTypeRow(typeName: string) {
    // Click a row in the type panel to load data in the right panel.
    const row = await this.resolveTypeRow(typeName);
    await row.click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  // ========== Data operations (right panel) ==========

  async createData(label: string, value: string, opts?: { sort?: number; remark?: string }) {
    // Click "新增" in the data panel toolbar
    await this.dataPanel.getByRole('button', { name: /新\s*增/ }).click();

    // Wait for drawer to open
    await this.dialog.waitFor({ state: 'visible', timeout: 5000 });

    // Fill drawer form fields
    await this.dialog.getByLabel('数据标签').fill(label);
    await this.dialog.getByLabel('数据键值').fill(value);
    if (opts?.sort !== undefined) {
      const sortInput = this.dialog.getByLabel('显示排序');
      await sortInput.clear();
      await sortInput.fill(String(opts.sort));
    }
    if (opts?.remark) {
      await this.dialog.getByLabel('备注').fill(opts.remark);
    }

    // Click confirm button
    await this.dialog.getByRole('button', { name: /确\s*认/ }).click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async editData(label: string, fields: { label?: string; value?: string }) {
    // Search for the data label first
    await this.fillDataSearchField('字典标签', label);
    await this.clickDataSearch();

    // Click edit button in data panel
    await this.dataPanel.locator('.ant-btn-sm').filter({ hasText: /编\s*辑/ }).first().click();

    // Wait for drawer to open
    await this.dialog.waitFor({ state: 'visible', timeout: 5000 });

    if (fields.label) {
      const labelInput = this.dialog.getByLabel('数据标签');
      await labelInput.clear();
      await labelInput.fill(fields.label);
    }
    if (fields.value) {
      const valueInput = this.dialog.getByLabel('数据键值');
      await valueInput.clear();
      await valueInput.fill(fields.value);
    }

    // Click confirm button
    await this.dialog.getByRole('button', { name: /确\s*认/ }).click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async deleteData(label: string) {
    // Search for the data label first
    await this.fillDataSearchField('字典标签', label);
    await this.clickDataSearch();

    // Click delete button in data panel
    await this.dataPanel.locator('.ant-btn-sm').filter({ hasText: /删\s*除/ }).first().click();

    // Wait for either Popconfirm or modal to appear
    await this.page.waitForTimeout(500);

    // Try Popconfirm first (more common pattern)
    const popconfirm = this.page.locator('.ant-popconfirm:visible, .ant-popover:visible').first();
    const modal = this.page.locator('.ant-modal-confirm:visible').first();

    const isPopconfirm = await popconfirm.isVisible({ timeout: 1000 }).catch(() => false);
    const isModal = await modal.isVisible({ timeout: 1000 }).catch(() => false);

    if (isPopconfirm) {
      await popconfirm.getByRole('button', { name: /确\s*定|OK/i }).click();
    } else if (isModal) {
      await modal.getByRole('button', { name: /确\s*定|OK/i }).click();
    } else {
      // Fallback: try clicking any visible confirm button
      await this.page.getByRole('button', { name: /确\s*定|OK/i }).first().click();
    }

    // Wait for success message
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async hasData(label: string): Promise<boolean> {
    return this.dataPanel
      .locator('.vxe-body--row', { hasText: label })
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  // ========== Export ==========

  async clickTypeExport() {
    await this.typePanel.getByRole('button', { name: /导\s*出/ }).click();
    await this.page.waitForTimeout(1000);
  }

  async clickDataExport() {
    await this.dataPanel.getByRole('button', { name: /导\s*出/ }).click();
    await this.page.waitForTimeout(1000);
  }

  /** Click confirm button in the export confirm modal */
  async clickExportConfirm() {
    const modal = this.page.locator('[role="dialog"]');
    await modal.getByRole('button', { name: /确\s*认/ }).click();
    await this.page.waitForTimeout(500);
  }

  // ========== Search helpers ==========

  /** Fill search field in the type panel (left) */
  async fillTypeSearchField(label: string, value: string) {
    const input = this.typePanel.getByLabel(label, { exact: true }).first();
    await input.clear();
    await input.fill(value);
  }

  /** Click search button in the type panel */
  async clickTypeSearch() {
    await this.typePanel.getByRole('button', { name: /搜\s*索/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Click reset button in the type panel */
  async clickTypeReset() {
    await this.typePanel.getByRole('button', { name: /重\s*置/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Fill search field in the data panel (right) */
  async fillDataSearchField(label: string, value: string) {
    const input = this.dataPanel.getByLabel(label, { exact: true }).first();
    await input.clear();
    await input.fill(value);
  }

  /** Click search button in the data panel */
  async clickDataSearch() {
    await this.dataPanel.getByRole('button', { name: /搜\s*索/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Click reset button in the data panel */
  async clickDataReset() {
    await this.dataPanel.getByRole('button', { name: /重\s*置/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Get visible row count in the data panel */
  async getDataRowCount(): Promise<number> {
    return this.dataPanel.locator('.vxe-body--row').count();
  }

  /** Get visible row count in the type panel */
  async getTypeRowCount(): Promise<number> {
    return this.typePanel.locator('.vxe-body--row').count();
  }

  /** Select a row checkbox in the type panel by clicking its checkbox */
  async selectTypeRow(index: number = 0) {
    const checkbox = this.typePanel.locator('.vxe-body--row .vxe-checkbox--icon').nth(index);
    await checkbox.click();
    await this.page.waitForTimeout(300);
  }

  /** Select a row checkbox in the data panel by clicking its checkbox */
  async selectDataRow(index: number = 0) {
    const checkbox = this.dataPanel.locator('.vxe-body--row .vxe-checkbox--icon').nth(index);
    await checkbox.click();
    await this.page.waitForTimeout(300);
  }

  // ========== Import ==========

  /** Click import button in the type panel */
  async clickTypeImport() {
    await this.typePanel.getByRole('button', { name: /导\s*入/ }).click();
    await this.page.waitForTimeout(500);
  }

  /** Click import button in the data panel */
  async clickDataImport() {
    await this.dataPanel.getByRole('button', { name: /导\s*入/ }).click();
    await this.page.waitForTimeout(500);
  }
}
