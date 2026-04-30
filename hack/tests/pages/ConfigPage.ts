import type { Locator, Page } from '@playwright/test';

import {
  waitForBusyIndicatorsToClear,
  waitForConfirmOverlay,
  waitForDialogReady,
  waitForDropdown,
  waitForRouteReady,
  waitForTableReady,
} from '../support/ui';

export class ConfigPage {
  constructor(private page: Page) {}

  private escapeRegex(value: string) {
    return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  }

  private localizedLabelPattern(label: string) {
    const labelMap: Record<string, RegExp> = {
      参数名称: /参数名称|Parameter Name/i,
      参数键名: /参数键名|Parameter Key/i,
      参数键值: /参数键值|Parameter Value/i,
      备注: /备注|Remark/i,
    };
    return labelMap[label] ?? new RegExp(`^\\s*${this.escapeRegex(label)}\\s*$`);
  }

  private resolveLocalizedLabel(scope: Page | Locator, label: string) {
    return scope.getByLabel(this.localizedLabelPattern(label)).first();
  }

  private searchFieldName(label: string) {
    const fieldMap: Record<string, string> = {
      参数名称: 'name',
      参数键名: 'key',
    };
    return fieldMap[label];
  }

  private get searchForm() {
    return this.page.locator('.vxe-grid--form-wrapper').first();
  }

  private async fillInputAndWaitForStableValue(input: Locator, value: string) {
    const deadline = Date.now() + 5000;
    while (Date.now() < deadline) {
      await input.waitFor({ state: 'visible', timeout: 2000 });
      await input.clear();
      await input.fill(value);
      await this.page.waitForTimeout(600);
      await waitForBusyIndicatorsToClear(this.page, 2000);
      if ((await input.inputValue().catch(() => '')) === value) {
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

  async goto() {
    await this.page.goto('/system/config');
    await waitForTableReady(this.page);
  }

  // ========== CRUD operations ==========

  async create(name: string, key: string, value: string, remark?: string) {
    await this.page.getByRole('button', { name: /新\s*增/ }).click();
    await waitForDialogReady(this.dialog);

    await this.dialog.getByLabel('参数名称').fill(name);
    await this.dialog.getByLabel('参数键名').fill(key);
    await this.dialog.getByLabel('参数键值').fill(value);
    if (remark) {
      await this.dialog.getByLabel('备注').fill(remark);
    }

    await this.dialog.getByRole('button', { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
    await this.dialog.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
  }

  async edit(configName: string, fields: { name?: string; key?: string; value?: string; remark?: string }) {
    // Search for the config first to narrow results
    await this.fillSearchField('参数名称', configName);
    await this.clickSearch();

    // Click edit button
    await this.page.locator('.ant-btn-sm').filter({ hasText: /编\s*辑/ }).first().click();
    await waitForDialogReady(this.dialog);

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
    await waitForRouteReady(this.page);
    await this.dialog.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
  }

  async delete(configName: string) {
    // Search for the config first
    await this.fillSearchField('参数名称', configName);
    await this.clickSearch();

    const targetRow = this.page.locator('.vxe-body--row').filter({ hasText: configName }).first();
    await targetRow.waitFor({ state: 'visible', timeout: 5000 });
    await targetRow.hover();

    // Click delete button
    const rowDeleteButton = targetRow.locator('.ant-btn-sm').filter({ hasText: /删\s*除/ }).first();
    if (await rowDeleteButton.isVisible().catch(() => false)) {
      await rowDeleteButton.click();
    } else {
      await this.page.locator('.ant-btn-sm').filter({ hasText: /删\s*除/ }).first().click();
    }

    // Confirm deletion in Popconfirm
    const popconfirm = await waitForConfirmOverlay(this.page);
    const confirmBtn = popconfirm.getByRole('button', { name: /确\s*定|OK|是/i });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      const modal = this.page.locator('.ant-modal-confirm');
      await modal.getByRole('button', { name: /确\s*定|OK/i }).click();
    }

    await waitForRouteReady(this.page);
  }

  async hasConfig(text: string): Promise<boolean> {
    return this.page
      .locator('.vxe-body--row')
      .filter({ hasText: text })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  findRowByExactKey(key: string): Locator {
    const keyPattern = new RegExp(`^\\s*${this.escapeRegex(key)}\\s*$`);

    return this.page
      .locator('.vxe-body--row', {
        has: this.page.locator('.vxe-cell').filter({ hasText: keyPattern }),
      })
      .first();
  }

  async getRowCount(): Promise<number> {
    return this.page.locator('.vxe-body--row').count();
  }

  // ========== Search helpers ==========

  async fillSearchField(label: string, value: string) {
    const fieldName = this.searchFieldName(label);
    if (fieldName) {
      const namedInput = this.searchForm.locator(`input[name="${fieldName}"]`).first();
      if (await namedInput.isVisible().catch(() => false)) {
        await this.fillInputAndWaitForStableValue(namedInput, value);
        return;
      }
    }

    const input = this.searchForm
      .locator('.ant-form-item')
      .filter({ hasText: this.localizedLabelPattern(label) })
      .locator('input')
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
      .getByRole('button', { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  async clickReset() {
    await this.searchForm
      .getByRole('button', { name: /重\s*置|Reset/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  // ========== Row Selection ==========

  async selectRow(configName: string) {
    await this.fillSearchField('参数名称', configName);
    await this.clickSearch();
    // Click the first checkbox in the body rows
    const checkbox = this.page.locator('.vxe-body--row .vxe-checkbox--icon').first();
    await checkbox.click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  // ========== Export ==========

  async clickExport() {
    await this.page.getByRole('button', { name: /导\s*出|Export/i }).click();
    await waitForDialogReady(this.page.locator('[role="dialog"]'));
  }

  /** Click confirm button in the export confirm modal */
  async clickExportConfirm() {
    const modal = this.page.locator('[role="dialog"]');
    await modal.getByRole('button', { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
  }

  // ========== Import ==========

  async clickImport() {
    await this.page.getByRole('button', { name: /导\s*入|Import/i }).click();
    await waitForDialogReady(this.dialog);
  }
}
