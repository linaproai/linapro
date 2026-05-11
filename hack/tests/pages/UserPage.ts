import { expect, type Page } from "@playwright/test";

import {
  waitForBusyIndicatorsToClear,
  waitForConfirmOverlay,
  waitForDialogReady,
  waitForDropdown,
  waitForRouteReady,
  waitForTableReady,
} from "../support/ui";

export class UserPage {
  constructor(private page: Page) {}

  /** Role column index within the main user table row. */
  private static readonly ROLE_COLUMN_INDEX = 5;

  /** Drawer submit can settle slowly in full-suite parallel runs. */
  private static readonly DRAWER_HIDDEN_TIMEOUT = 20000;

  /** The Vben drawer (Sheet/Dialog) container */
  private get drawer() {
    return this.page
      .locator('[role="dialog"]')
      .filter({
        has: this.page.getByPlaceholder(/请输入(?:账号|用户名)|account|username/i),
      })
      .last();
  }

  /** User drawer account input. */
  private get drawerAccountInput() {
    return this.drawer.getByPlaceholder(/请输入(?:账号|用户名)|account|username/i);
  }

  /** Username search input in the list filter form. */
  private get usernameSearchInput() {
    return this.page.getByLabel(/用户账号|User Account/i).first();
  }

  /** User drawer role combobox */
  private get roleCombobox() {
    return this.drawer
      .getByRole("combobox", { name: "角色", exact: true })
      .first();
  }

  /** User drawer role select wrapper */
  private get roleSelect() {
    return this.roleCombobox
      .locator('xpath=ancestor::*[contains(@class,"ant-select")]')
      .first();
  }

  /** Wait until the user drawer has finished async schema/data initialization. */
  private async waitForDrawerReady(expectedUsername: string) {
    await waitForDialogReady(this.drawer);

    const usernameInput = this.drawerAccountInput;
    await usernameInput.waitFor({ state: "visible", timeout: 10000 });
    await expect(usernameInput).toHaveValue(expectedUsername, {
      timeout: 10000,
    });

    await this.roleCombobox.waitFor({ state: "visible", timeout: 10000 });
    await waitForBusyIndicatorsToClear(this.drawer);
  }

  /**
   * Resolve the main table row for the given username.
   *
   * VXE renders fixed action columns in a separate table tree, so callers that
   * need business data should always work with the primary data row first.
   */
  private getUserDataRow(username: string) {
    return this.page
      .locator(".vxe-body--row:visible", { hasText: username })
      .first();
  }

  /** Public row locator for assertions after filtering. */
  getUserRow(username: string) {
    return this.getUserDataRow(username);
  }

  /** Check whether the left department tree shows the expected raw department label. */
  async hasDeptTreeNode(label: string): Promise<boolean> {
    return this.page
      .locator(".ant-tree")
      .getByText(label, { exact: false })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Wait for the VXE grid loading mask to settle before interacting. */
  private async waitForGridIdle() {
    await waitForBusyIndicatorsToClear(this.page);
  }

  async goto() {
    await this.page.goto("/system/user");
    await waitForTableReady(this.page);
    await this.waitForGridIdle();
  }

  async createUser(username: string, password: string, nickname?: string) {
    // The "新 增" button is in the toolbar (spaced text)
    await this.page.getByRole("button", { name: /新\s*增/ }).click();

    await this.waitForDrawerReady("");

    // Fill form fields scoped to the drawer to avoid conflict with the search form
    await this.drawerAccountInput.fill(username);
    await this.drawer.getByPlaceholder("请输入密码").fill(password);
    if (nickname) {
      await this.drawer.getByPlaceholder("请输入昵称").fill(nickname);
    }

    // Click the drawer's confirm button (确 认 - note space in Ant Design)
    await this.drawer.getByRole("button", { name: /确\s*认/ }).click();

    await this.page.waitForLoadState("networkidle");
    await this.waitForGridIdle();
    await this.drawer.waitFor({
      state: "hidden",
      timeout: UserPage.DRAWER_HIDDEN_TIMEOUT,
    });
  }

  async editUser(username: string, fields: { nickname?: string }) {
    // VXE-Grid with fixed: 'right' action column renders buttons in a separate
    // fixed overlay DOM tree. Search for the user first to narrow to one row.
    await this.searchByUsername(username);

    // With search filtering to one row, click the first visible edit button
    // Note: Ant Design adds space between Chinese chars in buttons ("编 辑")
    await this.page
      .getByRole("button", { name: /编\s*辑/ })
      .first()
      .click();

    await this.waitForDrawerReady(username);

    if (fields.nickname) {
      const nicknameInput = this.drawer.getByPlaceholder("请输入昵称");
      await nicknameInput.waitFor({ state: "visible", timeout: 5000 });
      await nicknameInput.clear();
      await nicknameInput.fill(fields.nickname);
    }

    // Click the drawer's confirm button
    await this.drawer.getByRole("button", { name: /确\s*认/ }).click();

    await this.page.waitForLoadState("networkidle");
    await this.waitForGridIdle();
    await this.drawer.waitFor({
      state: "hidden",
      timeout: UserPage.DRAWER_HIDDEN_TIMEOUT,
    });
  }

  async deleteUser(username: string) {
    // VXE-Grid with fixed: 'right' action column - search to narrow to one row
    await this.searchByUsername(username);

    await this.getUserDataRow(username).waitFor({
      state: "visible",
      timeout: 10000,
    });

    const deleteButton = this.page
      .locator(
        "button.ant-btn-primary.ant-btn-background-ghost.ant-btn-sm:not([disabled])",
      )
      .filter({ hasText: /删\s*除/ })
      .first();
    await deleteButton.waitFor({ state: "visible", timeout: 5000 });
    await deleteButton.click();

    // Confirm deletion in the Popconfirm
    // Popconfirm uses ant-popover
    const popconfirm = await waitForConfirmOverlay(this.page);
    const confirmBtn = popconfirm.getByRole("button", {
      name: /确\s*定|OK|是/i,
    });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      // Fallback: Ant Design Modal confirm
      const modal = this.page.locator(".ant-modal-confirm");
      await modal.getByRole("button", { name: /确\s*定|OK/i }).click();
    }

    await this.page.waitForLoadState("networkidle");
    await this.waitForGridIdle();
    await this.getUserDataRow(username)
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
  }

  async hasUser(username: string): Promise<boolean> {
    await this.waitForGridIdle();
    const rowCount = await this.page
      .locator(".vxe-body--row:visible", { hasText: username })
      .count();
    return rowCount > 0;
  }

  /** Click a column header to trigger sorting */
  async clickColumnSort(columnTitle: string) {
    // VXE-Grid has duplicate headers (visible + fixed-hidden), use .first() for visible one
    const header = this.page
      .locator(".vxe-header--column.fixed--visible", { hasText: columnTitle })
      .first();
    await header.click();
    await waitForRouteReady(this.page);
  }

  /** Get all cell values for a column by field name */
  async getColumnValues(field: string): Promise<string[]> {
    const cells = this.page.locator(`.vxe-body--column[colid] .vxe-cell`);
    // Use a more reliable way: get all rows and extract the specific column
    const rows = this.page.locator(".vxe-body--row");
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
    return this.page.locator(".vxe-body--row").count();
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
    const form = this.page
      .locator(".vxe-grid--form-wrapper, .vben-form-wrapper")
      .first();
    const select = form.locator(".ant-select").first();
    await select.click();
    const dropdown = await waitForDropdown(this.page);
    await dropdown.getByText(statusLabel, { exact: true }).click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Click search/query button */
  async clickSearch() {
    await this.page
      .getByRole("button", { name: /搜\s*索|Search/i })
      .first()
      .click();
    await this.page.waitForLoadState("networkidle");
    await this.waitForGridIdle();
  }

  /** Click reset button */
  async clickReset() {
    await this.page
      .getByRole("button", { name: /重\s*置|Reset/i })
      .first()
      .click();
    await this.page.waitForLoadState("networkidle");
    await this.waitForGridIdle();
  }

  /** Reset filters and search by username in a deterministic fresh state. */
  async searchByUsername(username: string) {
    await this.clickReset();
    await this.usernameSearchInput.waitFor({
      state: "visible",
      timeout: 10000,
    });
    await this.usernameSearchInput.clear();
    await this.usernameSearchInput.fill(username);
    await this.clickSearch();
  }

  /** Search by a username keyword without forcing exact-match semantics. */
  async searchByUsernameKeyword(keyword: string) {
    await this.clickReset();
    await this.usernameSearchInput.waitFor({
      state: "visible",
      timeout: 10000,
    });
    await this.usernameSearchInput.clear();
    await this.usernameSearchInput.fill(keyword);
    await this.clickSearch();
  }

  /** Select multiple visible user rows by username. */
  async selectVisibleUserRows(usernames: string[]) {
    for (const username of usernames) {
      const row = this.getUserDataRow(username);
      await expect(row).toBeVisible();
      await row.locator(".vxe-checkbox--icon").first().click();
      await waitForBusyIndicatorsToClear(this.page);
    }
  }

  /** Click toolbar batch delete and confirm the overlay. */
  async confirmSelectedUserBatchDelete() {
    await this.page.getByTestId("user-batch-delete-button").click();
    const confirmOverlay = await waitForConfirmOverlay(this.page);
    await confirmOverlay
      .getByRole("button", { name: /确\s*定|OK|是/i })
      .last()
      .click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Click export button */
  async clickExport() {
    await this.page.getByRole("button", { name: /导\s*出/ }).click();
    await waitForDialogReady(this.page.locator('[role="dialog"]'));
  }

  /** Click confirm button in the export confirm modal */
  async clickExportConfirm() {
    const modal = this.page.locator('[role="dialog"]');
    await modal.getByRole("button", { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
  }

  /** Select a row by clicking its checkbox (search for the user first) */
  async selectRow(username: string) {
    await this.fillSearchField("用户账号", username);
    await this.clickSearch();
    // Click the first checkbox in the body rows
    const checkbox = this.page
      .locator(".vxe-body--row .vxe-checkbox--icon")
      .first();
    await checkbox.click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Check if the export button is visible */
  async isExportVisible(): Promise<boolean> {
    return this.page
      .getByRole("button", { name: /导\s*出/ })
      .isVisible({ timeout: 2000 })
      .catch(() => false);
  }

  /** Check if the toolbar delete button is visible */
  async isToolbarDeleteVisible(): Promise<boolean> {
    // Toolbar delete button is a primary danger button (not the ghost button in rows)
    return this.page
      .locator(".vxe-grid--toolbar")
      .getByRole("button", { name: /删\s*除/ })
      .isVisible({ timeout: 2000 })
      .catch(() => false);
  }

  /** Check if action buttons (edit/delete/more) are visible for a row */
  async hasActionButtons(username: string): Promise<boolean> {
    await this.searchByUsername(username);
    const editBtn = this.page.getByRole("button", { name: /编\s*辑/ }).first();
    return editBtn.isVisible({ timeout: 2000 }).catch(() => false);
  }

  /** Check if the status switch is disabled for a row */
  async isStatusSwitchDisabled(username: string): Promise<boolean> {
    await this.searchByUsername(username);
    const switchEl = this.page.locator(".vxe-body--row .ant-switch").first();
    return switchEl.evaluate((el) =>
      el.classList.contains("ant-switch-disabled"),
    );
  }

  /** Check if the row checkbox is disabled */
  async isCheckboxDisabled(username: string): Promise<boolean> {
    await this.searchByUsername(username);
    const checkbox = this.page
      .locator(".vxe-body--row .vxe-cell--checkbox")
      .first();
    return checkbox.evaluate((el) => el.classList.contains("is--disabled"));
  }

  /** Click import button to open import modal */
  async clickImport() {
    await this.page
      .getByRole("button", { name: /导\s*入/ })
      .first()
      .click();
    await waitForDialogReady(this.page.locator('[role="dialog"]'));
  }

  /** Get the total count from the pager */
  async getTotalCount(): Promise<number> {
    const pager = this.page.locator(".vxe-pager--total");
    const text = await pager.textContent();
    const match = text?.match(/(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }

  /** Select roles in the user drawer */
  async selectRoles(roleNames: string[]) {
    await this.roleCombobox.waitFor({ state: "visible", timeout: 5000 });

    for (const roleName of roleNames) {
      await this.roleCombobox.click();
      await waitForBusyIndicatorsToClear(this.page);
      // Filter the dropdown first so we do not depend on the option already being in view.
      await this.roleCombobox.fill(roleName);

      const dropdown = await waitForDropdown(this.page);
      const option = dropdown.getByText(roleName, { exact: true }).first();
      await option.waitFor({ state: "visible", timeout: 5000 });
      await option.click();
      await waitForBusyIndicatorsToClear(this.page);
    }
  }

  /** Get visible role names from user list table */
  async getRoleNames(username: string): Promise<string> {
    await this.searchByUsername(username);

    const row = this.getUserDataRow(username);
    await row.waitFor({ state: "visible", timeout: 10000 });

    // The role cell is stable by visible column order even when VXE duplicates
    // DOM fragments for fixed columns. Using accessible cells avoids depending
    // on internal `field` attributes that are not rendered consistently.
    const roleCell = row.getByRole("cell").nth(UserPage.ROLE_COLUMN_INDEX);
    const roleText = await roleCell.textContent();
    return roleText?.trim() || "";
  }

  /** Get role count from user drawer */
  async getSelectedRoleCount(): Promise<number> {
    const roleSelect = this.roleSelect;
    // Ant Design multi-select shows selected items as tags
    const selectedTags = roleSelect.locator(".ant-select-selection-item");
    return await selectedTags.count();
  }

  /** Create user with roles */
  async createUserWithRoles(
    username: string,
    password: string,
    nickname: string,
    roleNames: string[],
  ) {
    await this.page.getByRole("button", { name: /新\s*增/ }).click();
    await this.waitForDrawerReady("");

    await this.drawerAccountInput.fill(username);
    await this.drawer.getByPlaceholder("请输入密码").fill(password);
    await this.drawer.getByPlaceholder("请输入昵称").fill(nickname);

    // Select roles
    await this.selectRoles(roleNames);

    await this.drawer.getByRole("button", { name: /确\s*认/ }).click();
    await this.page.waitForLoadState("networkidle");
    await this.waitForGridIdle();
    await this.drawer.waitFor({
      state: "hidden",
      timeout: UserPage.DRAWER_HIDDEN_TIMEOUT,
    });
  }

  /** Edit user's roles */
  async editUserRoles(username: string, roleNames: string[]) {
    await this.searchByUsername(username);

    // Ensure the searched row is rendered before interacting with the fixed
    // action column. The action buttons live in a separate fixed table, but the
    // visible edit button becomes unique once the main data row is filtered.
    await this.getUserDataRow(username).waitFor({
      state: "visible",
      timeout: 10000,
    });
    await this.page
      .getByRole("button", { name: /编\s*辑/ })
      .first()
      .click();
    await this.waitForDrawerReady(username);

    // Clear existing roles first by clicking clear button
    const roleSelect = this.roleSelect;
    const clearBtn = roleSelect.locator(".ant-select-clear");
    if (await clearBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await clearBtn.click();
      await waitForBusyIndicatorsToClear(this.page);
    }

    // Select new roles
    await this.selectRoles(roleNames);

    await this.drawer.getByRole("button", { name: /确\s*认/ }).click();
    await this.page.waitForLoadState("networkidle");
    await this.waitForGridIdle();
    await this.drawer.waitFor({
      state: "hidden",
      timeout: UserPage.DRAWER_HIDDEN_TIMEOUT,
    });
  }
}
