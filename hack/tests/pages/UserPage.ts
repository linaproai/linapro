import type { Page } from '@playwright/test';

export class UserPage {
  constructor(private page: Page) {}

  /** Role column index within the main user table row. */
  private static readonly ROLE_COLUMN_INDEX = 5;

  /** The Vben drawer (Sheet/Dialog) container */
  private get drawer() {
    return this.page.locator('[role="dialog"]');
  }

  /** User drawer role combobox */
  private get roleCombobox() {
    return this.drawer.getByRole('combobox', { name: '角色', exact: true }).first();
  }

  /** User drawer role select wrapper */
  private get roleSelect() {
    return this.roleCombobox
      .locator('xpath=ancestor::*[contains(@class,"ant-select")]')
      .first();
  }

  /**
   * Resolve the main table row for the given username.
   *
   * VXE renders fixed action columns in a separate table tree, so callers that
   * need business data should always work with the primary data row first.
   */
  private getUserDataRow(username: string) {
    return this.page.locator('.vxe-body--row', { hasText: username }).first();
  }

  async goto() {
    await this.page.goto('/system/user');
    await this.page.waitForLoadState('networkidle');
    // Wait for VxeGrid table to render
    await this.page.locator('.vxe-table').waitFor({ state: 'visible', timeout: 10000 });
  }

  async createUser(
    username: string,
    password: string,
    nickname?: string,
  ) {
    // The "新 增" button is in the toolbar (spaced text)
    await this.page.getByRole('button', { name: /新\s*增/ }).click();

    // Wait for drawer (Sheet dialog) to open
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Fill form fields scoped to the drawer to avoid conflict with the search form
    await this.drawer.getByPlaceholder('请输入用户名').fill(username);
    await this.drawer.getByPlaceholder('请输入密码').fill(password);
    if (nickname) {
      await this.drawer.getByPlaceholder('请输入昵称').fill(nickname);
    }

    // Click the drawer's confirm button (确 认 - note space in Ant Design)
    await this.drawer.getByRole('button', { name: /确\s*认/ }).click();

    // Wait for API response
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async editUser(username: string, fields: { nickname?: string }) {
    // VXE-Grid with fixed: 'right' action column renders buttons in a separate
    // fixed overlay DOM tree. Search for the user first to narrow to one row.
    await this.fillSearchField('用户账号', username);
    await this.clickSearch();

    // With search filtering to one row, click the first visible edit button
    // Note: Ant Design adds space between Chinese chars in buttons ("编 辑")
    await this.page.getByRole('button', { name: /编\s*辑/ }).first().click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    if (fields.nickname) {
      const nicknameInput = this.drawer.getByPlaceholder('请输入昵称');
      await nicknameInput.clear();
      await nicknameInput.fill(fields.nickname);
    }

    // Click the drawer's confirm button
    await this.drawer.getByRole('button', { name: /确\s*认/ }).click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async deleteUser(username: string) {
    // VXE-Grid with fixed: 'right' action column - search to narrow to one row
    await this.fillSearchField('用户账号', username);
    await this.clickSearch();

    // Click the first visible delete button (ghost-button = ant-btn-sm, not toolbar's full button)
    // Note: Ant Design adds space between Chinese chars in buttons ("删 除")
    await this.page.locator('.ant-btn-sm').filter({ hasText: /删\s*除/ }).first().click();

    // Confirm deletion in the Popconfirm
    await this.page.waitForTimeout(500);
    // Popconfirm uses ant-popover
    const popconfirm = this.page.locator('.ant-popconfirm, .ant-popover');
    const confirmBtn = popconfirm.getByRole('button', { name: /确\s*定|OK|是/i });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      // Fallback: Ant Design Modal confirm
      const modal = this.page.locator('.ant-modal-confirm');
      await modal.getByRole('button', { name: /确\s*定|OK/i }).click();
    }

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  async hasUser(username: string): Promise<boolean> {
    return this.page
      .locator('.vxe-body--row', { hasText: username })
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Click a column header to trigger sorting */
  async clickColumnSort(columnTitle: string) {
    // VXE-Grid has duplicate headers (visible + fixed-hidden), use .first() for visible one
    const header = this.page.locator('.vxe-header--column.fixed--visible', { hasText: columnTitle }).first();
    await header.click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Get all cell values for a column by field name */
  async getColumnValues(field: string): Promise<string[]> {
    const cells = this.page.locator(`.vxe-body--column[colid] .vxe-cell`);
    // Use a more reliable way: get all rows and extract the specific column
    const rows = this.page.locator('.vxe-body--row');
    const count = await rows.count();
    const values: string[] = [];
    for (let i = 0; i < count; i++) {
      const row = rows.nth(i);
      // Try to get the cell text for the column
      const cell = row.locator(`td[field="${field}"] .vxe-cell, td .vxe-cell`);
      // Fallback: use column index mapping
    }
    return values;
  }

  /** Get visible row count */
  async getVisibleRowCount(): Promise<number> {
    return this.page.locator('.vxe-body--row').count();
  }

  /** Fill the search form field by label */
  async fillSearchField(label: string, value: string) {
    // The Vben5 form renders labels as text followed by input fields
    // Use getByLabel which matches aria-label or associated label text
    const input = this.page.getByLabel(label, { exact: true }).first();
    await input.clear();
    await input.fill(value);
  }

  /** Select status in search form */
  async selectSearchStatus(statusLabel: string) {
    const form = this.page.locator('.vxe-grid--form-wrapper, .vben-form-wrapper').first();
    const select = form.locator('.ant-select').first();
    await select.click();
    await this.page.getByText(statusLabel, { exact: true }).click();
    await this.page.waitForTimeout(300);
  }

  /** Click search/query button */
  async clickSearch() {
    await this.page.getByRole('button', { name: /搜\s*索/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Click reset button */
  async clickReset() {
    await this.page.getByRole('button', { name: /重\s*置/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Click export button */
  async clickExport() {
    await this.page.getByRole('button', { name: /导\s*出/ }).click();
    await this.page.waitForTimeout(2000);
  }

  /** Click confirm button in the export confirm modal */
  async clickExportConfirm() {
    const modal = this.page.locator('[role="dialog"]');
    await modal.getByRole('button', { name: /确\s*认/ }).click();
    await this.page.waitForTimeout(500);
  }

  /** Select a row by clicking its checkbox (search for the user first) */
  async selectRow(username: string) {
    await this.fillSearchField('用户账号', username);
    await this.clickSearch();
    // Click the first checkbox in the body rows
    const checkbox = this.page.locator('.vxe-body--row .vxe-checkbox--icon').first();
    await checkbox.click();
    await this.page.waitForTimeout(300);
  }

  /** Check if the export button is visible */
  async isExportVisible(): Promise<boolean> {
    return this.page
      .getByRole('button', { name: /导\s*出/ })
      .isVisible({ timeout: 2000 })
      .catch(() => false);
  }

  /** Check if the toolbar delete button is visible */
  async isToolbarDeleteVisible(): Promise<boolean> {
    // Toolbar delete button is a primary danger button (not the ghost button in rows)
    return this.page
      .locator('.vxe-grid--toolbar')
      .getByRole('button', { name: /删\s*除/ })
      .isVisible({ timeout: 2000 })
      .catch(() => false);
  }

  /** Check if action buttons (edit/delete/more) are visible for a row */
  async hasActionButtons(username: string): Promise<boolean> {
    await this.fillSearchField('用户账号', username);
    await this.clickSearch();
    const editBtn = this.page.getByRole('button', { name: /编\s*辑/ }).first();
    return editBtn.isVisible({ timeout: 2000 }).catch(() => false);
  }

  /** Check if the status switch is disabled for a row */
  async isStatusSwitchDisabled(username: string): Promise<boolean> {
    await this.fillSearchField('用户账号', username);
    await this.clickSearch();
    const switchEl = this.page.locator('.vxe-body--row .ant-switch').first();
    return switchEl.evaluate((el) => el.classList.contains('ant-switch-disabled'));
  }

  /** Check if the row checkbox is disabled */
  async isCheckboxDisabled(username: string): Promise<boolean> {
    await this.fillSearchField('用户账号', username);
    await this.clickSearch();
    const checkbox = this.page.locator('.vxe-body--row .vxe-cell--checkbox').first();
    return checkbox.evaluate((el) => el.classList.contains('is--disabled'));
  }

  /** Click import button to open import modal */
  async clickImport() {
    await this.page.getByRole('button', { name: /导\s*入/ }).first().click();
    await this.page.waitForTimeout(500);
  }

  /** Get the total count from the pager */
  async getTotalCount(): Promise<number> {
    const pager = this.page.locator('.vxe-pager--total');
    const text = await pager.textContent();
    const match = text?.match(/(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }

  /** Select roles in the user drawer */
  async selectRoles(roleNames: string[]) {
    await this.roleCombobox.waitFor({ state: 'visible', timeout: 5000 });

    for (const roleName of roleNames) {
      await this.roleCombobox.click();
      await this.page.waitForTimeout(300);
      // Filter the dropdown first so we do not depend on the option already being in view.
      await this.roleCombobox.fill(roleName);

      const dropdown = this.page.locator('.ant-select-dropdown:visible').last();
      const option = dropdown.getByText(roleName, { exact: true }).first();
      await option.waitFor({ state: 'visible', timeout: 5000 });
      await option.click();
      await this.page.waitForTimeout(300);
    }
  }

  /** Get visible role names from user list table */
  async getRoleNames(username: string): Promise<string> {
    await this.fillSearchField('用户账号', username);
    await this.clickSearch();

    const row = this.getUserDataRow(username);
    await row.waitFor({ state: 'visible', timeout: 10000 });

    // The role cell is stable by visible column order even when VXE duplicates
    // DOM fragments for fixed columns. Using accessible cells avoids depending
    // on internal `field` attributes that are not rendered consistently.
    const roleCell = row.getByRole('cell').nth(UserPage.ROLE_COLUMN_INDEX);
    const roleText = await roleCell.textContent();
    return roleText?.trim() || '';
  }

  /** Get role count from user drawer */
  async getSelectedRoleCount(): Promise<number> {
    const roleSelect = this.roleSelect;
    // Ant Design multi-select shows selected items as tags
    const selectedTags = roleSelect.locator('.ant-select-selection-item');
    return await selectedTags.count();
  }

  /** Create user with roles */
  async createUserWithRoles(
    username: string,
    password: string,
    nickname: string,
    roleNames: string[],
  ) {
    await this.page.getByRole('button', { name: /新\s*增/ }).click();
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    await this.drawer.getByPlaceholder('请输入用户名').fill(username);
    await this.drawer.getByPlaceholder('请输入密码').fill(password);
    await this.drawer.getByPlaceholder('请输入昵称').fill(nickname);

    // Select roles
    await this.selectRoles(roleNames);

    await this.drawer.getByRole('button', { name: /确\s*认/ }).click();
    await this.page.waitForLoadState('networkidle');
    await this.drawer.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
    await this.page.waitForTimeout(500);
  }

  /** Edit user's roles */
  async editUserRoles(username: string, roleNames: string[]) {
    await this.fillSearchField('用户账号', username);
    await this.clickSearch();

    // Ensure the searched row is rendered before interacting with the fixed
    // action column. The action buttons live in a separate fixed table, but the
    // visible edit button becomes unique once the main data row is filtered.
    await this.getUserDataRow(username).waitFor({ state: 'visible', timeout: 10000 });
    await this.page.getByRole('button', { name: /编\s*辑/ }).first().click();
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Clear existing roles first by clicking clear button
    const roleSelect = this.roleSelect;
    const clearBtn = roleSelect.locator('.ant-select-clear');
    if (await clearBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await clearBtn.click();
      await this.page.waitForTimeout(300);
    }

    // Select new roles
    await this.selectRoles(roleNames);

    await this.drawer.getByRole('button', { name: /确\s*认/ }).click();
    await this.page.waitForLoadState('networkidle');
    await this.drawer.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
    await this.page.waitForTimeout(500);
  }
}
