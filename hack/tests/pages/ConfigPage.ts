import type { Page } from '@playwright/test';

export class ConfigPage {
  constructor(private page: Page) {}

  /** The modal dialog container */
  private get dialog() {
    return this.page.locator('[role="dialog"]');
  }

  async goto() {
    await this.page.goto('/system/config');
    await this.page.waitForLoadState('networkidle');
    await this.page.locator('.vxe-table').first().waitFor({ state: 'visible', timeout: 10000 });
  }

  // ========== CRUD operations ==========

  async create(name: string, key: string, value: string, remark?: string) {
    await this.page.getByRole('button', { name: /新\s*增/ }).click();
    await this.dialog.waitFor({ state: 'visible', timeout: 5000 });

    await this.dialog.getByLabel('参数名称').fill(name);
    await this.dialog.getByLabel('参数键名').fill(key);
    await this.dialog.getByLabel('参数键值').fill(value);
    if (remark) {
      await this.dialog.getByLabel('备注').fill(remark);
    }

    await this.dialog.getByRole('button', { name: /确\s*认/ }).click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async edit(configName: string, fields: { name?: string; key?: string; value?: string; remark?: string }) {
    // Search for the config first to narrow results
    await this.fillSearchField('参数名称', configName);
    await this.clickSearch();

    // Click edit button
    await this.page.locator('.ant-btn-sm').filter({ hasText: /编\s*辑/ }).first().click();
    await this.dialog.waitFor({ state: 'visible', timeout: 5000 });

    if (fields.name) {
      const input = this.dialog.getByLabel('参数名称');
      await input.clear();
      await input.fill(fields.name);
    }
    if (fields.key) {
      const input = this.dialog.getByLabel('参数键名');
      await input.clear();
      await input.fill(fields.key);
    }
    if (fields.value) {
      const input = this.dialog.getByLabel('参数键值');
      await input.clear();
      await input.fill(fields.value);
    }
    if (fields.remark) {
      const input = this.dialog.getByLabel('备注');
      await input.clear();
      await input.fill(fields.remark);
    }

    await this.dialog.getByRole('button', { name: /确\s*认/ }).click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async delete(configName: string) {
    // Search for the config first
    await this.fillSearchField('参数名称', configName);
    await this.clickSearch();

    // Click delete button
    await this.page.locator('.ant-btn-sm').filter({ hasText: /删\s*除/ }).first().click();

    // Confirm deletion in Popconfirm
    await this.page.waitForTimeout(500);
    const popconfirm = this.page.locator('.ant-popconfirm, .ant-popover');
    const confirmBtn = popconfirm.getByRole('button', { name: /确\s*定|OK|是/i });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      const modal = this.page.locator('.ant-modal-confirm');
      await modal.getByRole('button', { name: /确\s*定|OK/i }).click();
    }

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async hasConfig(text: string): Promise<boolean> {
    return this.page
      .locator('.vxe-body--row')
      .filter({ hasText: text })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  async getRowCount(): Promise<number> {
    return this.page.locator('.vxe-body--row').count();
  }

  // ========== Search helpers ==========

  async fillSearchField(label: string, value: string) {
    const input = this.page.getByLabel(label, { exact: true }).first();
    await input.clear();
    await input.fill(value);
  }

  async clickSearch() {
    await this.page.getByRole('button', { name: /搜\s*索/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async clickReset() {
    await this.page.getByRole('button', { name: /重\s*置/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  // ========== Row Selection ==========

  async selectRow(configName: string) {
    await this.fillSearchField('参数名称', configName);
    await this.clickSearch();
    // Click the first checkbox in the body rows
    const checkbox = this.page.locator('.vxe-body--row .vxe-checkbox--icon').first();
    await checkbox.click();
    await this.page.waitForTimeout(300);
  }

  // ========== Export ==========

  async clickExport() {
    await this.page.getByRole('button', { name: /导\s*出/ }).click();
    await this.page.waitForTimeout(1000);
  }

  /** Click confirm button in the export confirm modal */
  async clickExportConfirm() {
    const modal = this.page.locator('[role="dialog"]');
    await modal.getByRole('button', { name: /确\s*认/ }).click();
    await this.page.waitForTimeout(500);
  }

  // ========== Import ==========

  async clickImport() {
    await this.page.getByRole('button', { name: /导\s*入/ }).click();
    await this.page.waitForTimeout(500);
  }
}
