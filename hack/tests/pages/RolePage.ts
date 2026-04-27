import type { Locator, Page } from "@playwright/test";

import {
  waitForBusyIndicatorsToClear,
  waitForConfirmOverlay,
  waitForDialogReady,
  waitForRouteReady,
  waitForTableReady,
} from "../support/ui";

export class RolePage {
  constructor(private page: Page) {}

  private get roleNameSearchInput() {
    return this.page.getByLabel(/角色名称|Role Name/i).first();
  }

  /** The Vben drawer container */
  private get drawer() {
    return this.page.locator('[role="dialog"]');
  }

  async goto() {
    await this.page.goto("/system/role");
    await waitForTableReady(this.page);
    await this.roleNameSearchInput.waitFor({
      state: "visible",
      timeout: 10000,
    });
  }

  async openCreateDrawer() {
    await waitForRouteReady(this.page);
    await this.page
      .getByRole("button", { name: /新\s*增/ })
      .first()
      .click({ force: true });

    const drawer = await waitForDialogReady(this.drawer);
    await this.dismissTourOverlayIfPresent();
    return drawer;
  }

  /** Create a new role by clicking "新增" toolbar button */
  async createRole(params: {
    name: string;
    code: string;
    sort?: number;
    status?: number;
    remark?: string;
  }) {
    // Wait for page to be ready first
    await this.page.waitForLoadState("load");
    await waitForBusyIndicatorsToClear(this.page);

    const drawer = await this.openCreateDrawer();

    // Fill text fields first (these work even with tour overlay present)
    const nameInput = drawer.locator('input[placeholder="请输入角色名称"]');
    await nameInput.waitFor({ state: "visible", timeout: 5000 });
    await nameInput.fill(params.name);

    const codeInput = drawer.locator('input[placeholder="如: admin, user等"]');
    await codeInput.fill(params.code);

    if (params.sort !== undefined) {
      const sortInput = drawer.getByRole("spinbutton");
      await sortInput.fill(String(params.sort));
    }

    if (params.remark) {
      const remarkInput = drawer.locator('textarea[placeholder="请输入备注"]');
      await remarkInput.fill(params.remark);
    }

    // Wait for form to fully render
    await waitForBusyIndicatorsToClear(this.page);

    // Select status - RadioGroup with button style
    // Default is already '正常' (value 1), so we only need to click if status is 0
    const statusValue = params.status ?? 1;
    if (statusValue === 0) {
      // Click on "停用" radio button
      await drawer
        .locator(".ant-radio-button-wrapper")
        .filter({ hasText: "停用" })
        .click();
      await waitForBusyIndicatorsToClear(this.page);
    }

    // Select data scope (required field) - click the label text to select
    // The RadioGroup uses ant-radio-button-wrapper with button style
    const dataScopeLabel = drawer.getByText("全部数据权限", { exact: true });
    await dataScopeLabel.waitFor({ state: "visible", timeout: 5000 });
    await dataScopeLabel.click({ force: true });
    await waitForBusyIndicatorsToClear(this.page);

    // Verify the radio is selected (ant-radio-button-wrapper-checked class)
    const checkedRadio = drawer.locator(".ant-radio-button-wrapper-checked");
    await checkedRadio
      .waitFor({ state: "visible", timeout: 3000 })
      .catch(() => {
        // If not visible, try clicking again
        return dataScopeLabel.click({ force: true });
      });

    // Select menus if needed - for basic test we skip menu selection
    // Menu selection is tested separately in TC0061e

    // Click confirm button - scroll into view first since dialog may be taller than viewport
    const confirmBtn = drawer.getByRole("button", { name: /确\s*认/ });
    await confirmBtn.scrollIntoViewIfNeeded();
    await confirmBtn.click({ force: true });

    await this.page.waitForLoadState("load");
    await this.drawer
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Edit a role: find the row, click edit, update fields in drawer */
  async editRole(roleName: string, newName: string) {
    // Find the row and click the edit button
    const row = this.page.locator(".vxe-body--row", { hasText: roleName });
    await row
      .getByRole("button", { name: /编\s*辑/ })
      .first()
      .click();

    const drawer = await waitForDialogReady(this.drawer);
    await this.dismissTourOverlayIfPresent();

    // Clear and fill the new name
    const nameInput = drawer.locator('input[placeholder="请输入角色名称"]');
    await nameInput.clear();
    await nameInput.fill(newName);

    // Click confirm button - scroll into view first since dialog may be taller than viewport
    const confirmBtn = drawer.getByRole("button", { name: /确\s*认/ });
    await confirmBtn.scrollIntoViewIfNeeded();
    await confirmBtn.click({ force: true });

    await this.page.waitForLoadState("load");
    await this.drawer
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Delete a role: find the row, click delete, confirm in Popconfirm */
  async deleteRole(roleName: string) {
    // Find the row and click the delete ghost button
    const row = this.page.locator(".vxe-body--row", { hasText: roleName });
    await row
      .locator(".ant-btn-sm")
      .filter({ hasText: /删\s*除/ })
      .first()
      .click();

    // Wait for whichever Ant Design confirm overlay the page renders and then
    // prefer the semantic confirm label, falling back to the primary action.
    await waitForBusyIndicatorsToClear(this.page);
    const confirmOverlay = await waitForConfirmOverlay(this.page);
    const confirmBtn = confirmOverlay.getByRole("button", {
      name: /确\s*认|确\s*定|OK|是/i,
    });
    if (
      await confirmBtn
        .first()
        .isVisible({ timeout: 2000 })
        .catch(() => false)
    ) {
      await confirmBtn.first().click();
    } else {
      await confirmOverlay.locator("button.ant-btn-primary").last().click();
    }

    await this.page.waitForLoadState("load");
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Check if a role row with the given name is visible */
  async hasRole(roleName: string): Promise<boolean> {
    return this.page
      .locator(".vxe-body--row", { hasText: roleName })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Find the first role row containing the specified permission key. */
  roleRowByKey(roleKey: string): Locator {
    return this.page.locator(".vxe-body--row", { hasText: roleKey }).first();
  }

  /** Search role by name */
  async searchRole(name: string) {
    // Prefer the accessible label because it stays stable even when the form DOM is re-created.
    await waitForRouteReady(this.page);
    const searchInput = this.roleNameSearchInput;
    await searchInput.waitFor({ state: "visible", timeout: 10000 });
    await searchInput.fill(name);

    // Click search button
    await this.page
      .getByRole("button", { name: /搜\s*索/ })
      .first()
      .click();
    await waitForRouteReady(this.page);
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Reset search */
  async resetSearch() {
    await this.page
      .getByRole("button", { name: /重\s*置/ })
      .first()
      .click();
    await waitForRouteReady(this.page);
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Toggle role status */
  async toggleStatus(roleName: string) {
    const row = this.page.locator(".vxe-body--row", { hasText: roleName });
    const switchBtn = row.locator(".ant-switch");
    await switchBtn.click();
    await this.page.waitForLoadState("load");
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Click assign button to go to role-auth page */
  async clickAssign(roleName: string) {
    const row = this.page.locator(".vxe-body--row", { hasText: roleName });
    await row
      .getByRole("button", { name: /分\s*配/ })
      .first()
      .click();
    await this.page.waitForLoadState("load");
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Check menu in the menu tree table (for role edit) */
  async checkMenu(menuName: string) {
    const menuTree = this.drawer.locator(".vxe-table");
    const menuRow = menuTree.locator(".vxe-body--row", { hasText: menuName });
    const checkbox = menuRow.locator(".vxe-checkbox--icon");
    await checkbox.click({ force: true });
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Uncheck menu in the menu tree table (for role edit) */
  async uncheckMenu(menuName: string) {
    const menuTree = this.drawer.locator(".vxe-table");
    const menuRow = menuTree.locator(".vxe-body--row", { hasText: menuName });
    const checkbox = menuRow.locator(".vxe-checkbox--icon");
    await checkbox.click({ force: true });
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Get checked menu count in drawer */
  async getCheckedMenuCount(): Promise<number> {
    const menuTree = this.drawer.locator(".vxe-table");
    const checkedRows = menuTree.locator(".vxe-body--row.is--checked");
    return await checkedRows.count();
  }

  /** Create role with specific menus */
  async createRoleWithMenus(params: {
    name: string;
    code: string;
    sort?: number;
    remark?: string;
    menuNames?: string[];
  }) {
    await this.page.waitForLoadState("load");
    await waitForBusyIndicatorsToClear(this.page);

    const drawer = await this.openCreateDrawer();

    const nameInput = drawer.locator('input[placeholder="请输入角色名称"]');
    await nameInput.waitFor({ state: "visible", timeout: 5000 });
    await nameInput.fill(params.name);

    const codeInput = drawer.locator('input[placeholder="如: admin, user等"]');
    await codeInput.fill(params.code);

    if (params.sort !== undefined) {
      const sortInput = drawer.getByRole("spinbutton");
      await sortInput.fill(String(params.sort));
    }

    if (params.remark) {
      const remarkInput = drawer.locator('textarea[placeholder="请输入备注"]');
      await remarkInput.fill(params.remark);
    }

    const dataScopeLabel = drawer.getByText("全部数据权限", { exact: true });
    await dataScopeLabel.waitFor({ state: "visible", timeout: 5000 });
    await dataScopeLabel.click({ force: true });
    await waitForBusyIndicatorsToClear(this.page);

    if (params.menuNames && params.menuNames.length > 0) {
      await drawer
        .locator(".vxe-table")
        .waitFor({ state: "visible", timeout: 5000 });
      for (const menuName of params.menuNames) {
        await this.checkMenu(menuName);
      }
    }

    const confirmBtn = drawer.getByRole("button", { name: /确\s*认/ });
    await confirmBtn.scrollIntoViewIfNeeded();
    await confirmBtn.click({ force: true });

    await this.page.waitForLoadState("load");
    await this.drawer
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Assign menus to existing role */
  async assignMenusToRole(roleName: string, menuNames: string[]) {
    const row = this.page.locator(".vxe-body--row", { hasText: roleName });
    await row
      .getByRole("button", { name: /编\s*辑/ })
      .first()
      .click();

    await this.drawer.waitFor({ state: "visible", timeout: 5000 });

    // Wait for menu tree
    await this.drawer
      .locator(".vxe-table")
      .waitFor({ state: "visible", timeout: 3000 });

    // Clear existing selections - expand all and uncheck all
    const menuTree = this.drawer.locator(".vxe-table");
    const allCheckboxes = menuTree.locator(".vxe-checkbox--icon");
    const count = await allCheckboxes.count();
    for (let i = 0; i < count; i++) {
      const checkbox = allCheckboxes.nth(i);
      const row = checkbox.locator("xpath=..");
      const isChecked = await row.evaluate((el) =>
        el.classList.contains("is--checked"),
      );
      if (isChecked) {
        await checkbox.click();
        await waitForBusyIndicatorsToClear(this.page);
      }
    }

    // Select new menus
    for (const menuName of menuNames) {
      await this.checkMenu(menuName);
    }

    const confirmBtn = this.drawer.getByRole("button", { name: /确\s*认/ });
    await confirmBtn.scrollIntoViewIfNeeded();
    await confirmBtn.click({ force: true });
    await this.page.waitForLoadState("load");
    await this.drawer
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Navigate to role management page */
  async navigateTo() {
    await this.page.goto("/system/role");
    await waitForTableReady(this.page);
  }

  /** Check if status switch is disabled for a role */
  async isStatusSwitchDisabled(roleName: string): Promise<boolean> {
    await this.searchRole(roleName);
    const switchEl = this.page.locator(".vxe-body--row .ant-switch").first();
    return switchEl.evaluate((el) =>
      el.classList.contains("ant-switch-disabled"),
    );
  }

  /** Check if checkbox is disabled for a role */
  async isCheckboxDisabled(roleName: string): Promise<boolean> {
    await this.searchRole(roleName);
    const checkbox = this.page
      .locator(".vxe-body--row .vxe-cell--checkbox")
      .first();
    return checkbox.evaluate((el) => el.classList.contains("is--disabled"));
  }

  private async dismissTourOverlayIfPresent() {
    const endTourBtn = this.page.getByRole("button", { name: "结束导览" });
    if (await endTourBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await endTourBtn.click({ force: true });
      await waitForBusyIndicatorsToClear(this.page);
    }

    const tourClose = this.page.locator(".ant-tour-close");
    if (await tourClose.isVisible({ timeout: 300 }).catch(() => false)) {
      await tourClose.click({ force: true });
      await waitForBusyIndicatorsToClear(this.page);
    }
  }
}
