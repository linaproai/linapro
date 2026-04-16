import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";

export class MenuPage {
  constructor(private page: Page) {}

  /** The Vben drawer container */
  private get drawer() {
    return this.page.locator('[role="dialog"]');
  }

  /** The main sidebar container rendered by the application layout */
  private get sidebar() {
    return this.page.locator("aside").first();
  }

  async goto() {
    await this.page.goto("/system/menu");
    await this.page.waitForLoadState("networkidle");
    // Wait for VxeGrid table to render - short timeout, continue if not found
    await this.page
      .locator(".vxe-table")
      .waitFor({ state: "visible", timeout: 5000 })
      .catch(() => {});
  }

  /** Click "展开" toolbar button to expand all tree nodes */
  async expandAll() {
    const expandButton = this.page
      .getByRole("button", { name: /展\s*开/ })
      .first();
    if (await expandButton.isVisible({ timeout: 1000 }).catch(() => false)) {
      await expandButton.click();
      await this.page.waitForTimeout(500);
      return;
    }

    const collapseButton = this.page
      .getByRole("button", { name: /折\s*叠/ })
      .first();
    if (await collapseButton.isVisible({ timeout: 1000 }).catch(() => false)) {
      await this.page.waitForTimeout(500);
      return;
    }

    throw new Error("菜单工具栏未找到“展开”或“折叠”按钮");
  }

  /** Click "折叠" toolbar button to collapse all tree nodes */
  async collapseAll() {
    const collapseButton = this.page
      .getByRole("button", { name: /折\s*叠/ })
      .first();
    if (await collapseButton.isVisible({ timeout: 1000 }).catch(() => false)) {
      await collapseButton.click();
      await this.page.waitForTimeout(500);
      return;
    }

    const expandButton = this.page
      .getByRole("button", { name: /展\s*开/ })
      .first();
    if (await expandButton.isVisible({ timeout: 1000 }).catch(() => false)) {
      await this.page.waitForTimeout(500);
      return;
    }

    throw new Error("菜单工具栏未找到“折叠”或“展开”按钮");
  }

  /** Create a root menu by clicking "新增" toolbar button */
  async createRootMenu(params: {
    name: string;
    type: "D" | "M" | "B";
    path?: string;
    component?: string;
    perms?: string;
    icon?: string;
    sort?: number;
    visible?: number;
    status?: number;
  }) {
    // Click the primary "新增" button
    await this.page
      .getByRole("button", { name: /新\s*增/ })
      .first()
      .click();

    // Wait for drawer dialog to open
    const drawer = this.page.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });

    // Wait for loading skeleton to disappear (form is loading)
    // The skeleton has class 'ant-skeleton' - wait for it to be hidden
    const skeleton = drawer.locator(".ant-skeleton");
    await skeleton.waitFor({ state: "hidden", timeout: 10000 });

    // Small additional wait for form to fully render
    await this.page.waitForTimeout(500);

    // First select parent menu - click the 上级菜单 TreeSelect
    // Use a more specific selector for the visible TreeSelect
    const parentSelect = drawer
      .locator(".ant-select:not(.ant-select-hidden)")
      .first();
    await parentSelect.waitFor({ state: "visible", timeout: 5000 });
    await parentSelect.click();
    await this.page.waitForTimeout(500);

    // Wait for tree dropdown and select "根菜单" (root option)
    // The tree has role="tree" attribute
    const treeDropdown = this.page.locator('[role="tree"]');
    await treeDropdown.waitFor({ state: "visible", timeout: 5000 });
    // Click on "根菜单" text which is the first/root option
    await treeDropdown.getByText("根菜单", { exact: true }).click();
    await this.page.waitForTimeout(500);

    // Fill menu name
    const nameInput = drawer.locator('input[placeholder="请输入菜单名称"]');
    await nameInput.fill(params.name);
    await this.page.waitForTimeout(200);

    // Select menu type (RadioGroup with buttons) - click on the visible button label
    // Default is 'D' (目录), only click if different type is needed
    const typeLabels = { D: "目录", M: "菜单", B: "按钮" };
    if (params.type !== "D") {
      // The radio button input is hidden, click on the label/span that contains the text
      await drawer
        .locator(".ant-radio-button-wrapper")
        .filter({ hasText: typeLabels[params.type] })
        .click();
      await this.page.waitForTimeout(500);
    }

    // Fill path for directory/menu type (in createRootMenu)
    if (params.path && params.type !== "B") {
      // Find the path input directly by placeholder
      const pathInput = this.drawer.locator('input[placeholder*="路由地址"]');
      await pathInput.waitFor({ state: "visible", timeout: 5000 });
      await pathInput.fill(params.path);
      await this.page.waitForTimeout(200);
    }

    // Fill component for menu type
    if (params.component && params.type === "M") {
      const componentInput = this.drawer
        .locator(".ant-form-item")
        .filter({ hasText: "组件路径" })
        .locator("input");
      await componentInput.fill(params.component);
    }

    // Fill permissions for menu/button type
    if (params.perms && params.type !== "D") {
      const permsInput = this.drawer
        .locator(".ant-form-item")
        .filter({ hasText: "权限标识" })
        .locator("input");
      await permsInput.fill(params.perms);
    }

    // Fill sort if provided - use spinbutton role for the input number
    if (params.sort !== undefined) {
      const sortInput = this.drawer
        .locator('input[role="spinbutton"], .ant-input-number-input')
        .first();
      await sortInput.waitFor({ state: "visible", timeout: 5000 });
      await sortInput.fill(String(params.sort));
      await this.page.waitForTimeout(200);
    }

    // Fill visible if provided
    if (params.visible !== undefined) {
      const visibleSelect = this.drawer.locator(".ant-select").nth(1);
      await visibleSelect.click();
      await this.page.waitForTimeout(300);
      const visibleLabel = params.visible === 1 ? "显示" : "隐藏";
      await this.page
        .locator(".ant-select-dropdown")
        .getByText(visibleLabel, { exact: true })
        .click();
    }

    // Fill status if provided
    if (params.status !== undefined) {
      const statusSelect = this.drawer.locator(".ant-select").nth(2);
      await statusSelect.click();
      await this.page.waitForTimeout(300);
      const statusLabel = params.status === 1 ? "正常" : "停用";
      await this.page
        .locator(".ant-select-dropdown")
        .getByText(statusLabel, { exact: true })
        .click();
    }

    // Click confirm button
    await this.drawer.getByRole("button", { name: /确\s*认/ }).click();

    await this.page.waitForLoadState("networkidle");
    await this.page.waitForTimeout(500);
  }

  /** Create a sub menu under the specified parent row */
  async createSubMenu(
    parentName: string,
    params: {
      name: string;
      type: "D" | "M" | "B";
      path?: string;
      component?: string;
      perms?: string;
      sort?: number;
      visible?: number;
      status?: number;
    },
  ) {
    // Expand all to ensure parent is visible
    await this.expandAll();

    // Find the parent row and click the "新增" action button (green, btn-success)
    const parentRow = this.page.locator(".vxe-body--row", {
      hasText: parentName,
    });
    await parentRow.locator(".btn-success").first().click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: "visible", timeout: 5000 });
    const skeleton = this.drawer.locator(".ant-skeleton");
    await skeleton.waitFor({ state: "hidden", timeout: 10000 }).catch(() => {});
    await this.page.waitForTimeout(500);

    // Fill menu name
    const nameInput = this.drawer.locator(
      'input[placeholder="请输入菜单名称"]',
    );
    await nameInput.fill(params.name);

    // Select menu type
    const typeLabels = { D: "目录", M: "菜单", B: "按钮" };
    if (params.type !== "D") {
      await this.drawer
        .locator(".ant-radio-button-wrapper")
        .filter({ hasText: typeLabels[params.type] })
        .click();
      await this.page.waitForTimeout(500);
    }

    // Fill path for directory/menu type
    if (params.path && params.type !== "B") {
      const pathInput = this.drawer
        .locator('input[placeholder*="路由地址"]')
        .first();
      await pathInput.waitFor({ state: "visible", timeout: 5000 });
      await pathInput.fill(params.path);
    }

    // Fill component for menu type
    if (params.component && params.type === "M") {
      const componentInput = this.drawer
        .locator('input[placeholder="请输入"]')
        .last();
      await componentInput.waitFor({ state: "visible", timeout: 5000 });
      await componentInput.fill(params.component);
    }

    // Fill permissions for menu/button type
    if (params.perms && params.type !== "D") {
      const permsInput = this.drawer
        .locator('input[placeholder="请输入权限标识"]')
        .first();
      await permsInput.waitFor({ state: "visible", timeout: 5000 });
      await permsInput.fill(params.perms);
    }

    // Fill sort if provided - use spinbutton role for the input number
    if (params.sort !== undefined) {
      const sortInput = this.drawer
        .locator('input[role="spinbutton"], .ant-input-number-input')
        .first();
      await sortInput.waitFor({ state: "visible", timeout: 5000 });
      await sortInput.fill(String(params.sort));
      await this.page.waitForTimeout(200);
    }

    // Click confirm button
    await this.drawer.getByRole("button", { name: /确\s*认/ }).click();

    await this.page.waitForLoadState("networkidle");
    await this.page.waitForTimeout(500);
  }

  /** Edit a menu: find the row, click edit, update fields in drawer */
  async editMenu(menuName: string, newName: string) {
    // Find the row and click the edit button
    const row = this.page.locator(".vxe-body--row", { hasText: menuName });
    await row
      .getByRole("button", { name: /编\s*辑/ })
      .first()
      .click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: "visible", timeout: 5000 });

    // Clear and fill the new name
    const nameInput = this.drawer.locator(
      'input[placeholder="请输入菜单名称"]',
    );
    await nameInput.clear();
    await nameInput.fill(newName);

    // Click confirm button
    await this.drawer.getByRole("button", { name: /确\s*认/ }).click();

    await this.page.waitForLoadState("networkidle");
    await this.page.waitForTimeout(500);
  }

  /** Delete a menu: find the row, click delete, confirm in Popconfirm */
  async deleteMenu(menuName: string, cascade: boolean = false) {
    // Enable cascade delete if needed
    if (cascade) {
      const cascadeSwitch = this.page.locator(".ant-switch");
      const switchContainer = this.page.locator("text=级联删除").locator("..");
      const cascadeSwitchInContainer = switchContainer.locator(".ant-switch");
      if (!(await cascadeSwitchInContainer.isChecked())) {
        await cascadeSwitchInContainer.click();
        await this.page.waitForTimeout(300);
      }
    }

    // Find the row and click the delete ghost button
    const row = this.page.locator(".vxe-body--row", { hasText: menuName });
    await row
      .locator(".ant-btn-sm")
      .filter({ hasText: /删\s*除/ })
      .first()
      .click();

    // Confirm in Popconfirm
    await this.page.waitForTimeout(500);
    const popconfirm = this.page.locator(".ant-popconfirm, .ant-popover");
    const confirmBtn = popconfirm.getByRole("button", {
      name: /确\s*定|OK|是/i,
    });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      const modal = this.page.locator(".ant-modal-confirm");
      await modal.getByRole("button", { name: /确\s*定|OK/i }).click();
    }

    await this.page.waitForLoadState("networkidle");
    await this.page.waitForTimeout(500);
  }

  /** Delete menu by name with optional cascade */
  async deleteMenuByName(menuName: string, cascade: boolean = false) {
    // If cascade, enable the cascade switch first
    if (cascade) {
      const cascadeContainer = this.page.locator("text=级联删除").locator("..");
      const cascadeSwitchInContainer = cascadeContainer.locator(".ant-switch");
      const isChecked = await cascadeSwitchInContainer.evaluate(
        (el: any) => el.checked,
      );
      if (!isChecked) {
        await cascadeSwitchInContainer.click();
        await this.page.waitForTimeout(300);
      }
    }

    // Search for the menu first to narrow down
    await this.searchMenu(menuName);
    await this.page.waitForTimeout(500);

    // Find and click delete
    const row = this.page.locator(".vxe-body--row", { hasText: menuName });
    await row
      .locator(".ant-btn-sm")
      .filter({ hasText: /删\s*除/ })
      .first()
      .click();

    // Confirm deletion
    await this.page.waitForTimeout(500);
    const popconfirm = this.page.locator(".ant-popconfirm, .ant-popover");
    const confirmBtn = popconfirm.getByRole("button", {
      name: /确\s*定|OK|是/i,
    });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      const modal = this.page.locator(".ant-modal-confirm");
      await modal.getByRole("button", { name: /确\s*定|OK/i }).click();
    }

    await this.page.waitForLoadState("networkidle");
    await this.page.waitForTimeout(500);

    // Reset search to show all menus
    await this.resetSearch();
  }

  /** Check if a menu row with the given name is visible */
  async hasMenu(menuName: string): Promise<boolean> {
    return this.page
      .locator(".vxe-body--row", { hasText: menuName })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Search menu by name */
  async searchMenu(name: string) {
    const searchInput = this.page
      .getByRole("textbox", { name: "菜单名称" })
      .first();
    await searchInput.waitFor({ state: "visible", timeout: 10000 });
    await searchInput.fill(name);

    // Click search button or press enter
    await this.page
      .getByRole("button", { name: /搜\s*索/ })
      .first()
      .click();
    await this.page.waitForLoadState("networkidle");
    await this.page.waitForTimeout(500);
  }

  /** Reset search */
  async resetSearch() {
    await this.page
      .getByRole("button", { name: /重\s*置/ })
      .first()
      .click();
    await this.page.waitForLoadState("networkidle");
    await this.page.waitForTimeout(500);
  }

  /** Navigate to menu management page */
  async navigateTo() {
    await this.page.goto("/system/menu");
    await this.page.waitForLoadState("networkidle");
    await this.page
      .locator(".vxe-table")
      .waitFor({ state: "visible", timeout: 10000 });
  }

  /** Update the menu visibility flag from the edit drawer. */
  async updateMenuVisibility(menuName: string, visible: 0 | 1) {
    const targetLabel = visible === 1 ? "显示" : "隐藏";

    await this.searchMenu(menuName);

    try {
      const row = this.page
        .locator(".vxe-body--row", { hasText: menuName })
        .first();
      await row.waitFor({ state: "visible", timeout: 10000 });
      await row.getByRole("button", { name: /编\s*辑/ }).click();

      await this.drawer.waitFor({ state: "visible", timeout: 10000 });
      await this.drawer
        .locator(".ant-skeleton")
        .waitFor({ state: "hidden", timeout: 10000 })
        .catch(() => {});

      await this.drawer
        .locator("label.ant-radio-button-wrapper")
        .filter({ hasText: targetLabel })
        .click();

      await this.drawer.getByRole("button", { name: /确\s*认/ }).click();
      await this.drawer.waitFor({ state: "hidden", timeout: 15000 });
      await this.page.waitForLoadState("networkidle");
      await this.page.waitForTimeout(500);
    } finally {
      await this.resetSearch();
    }
  }

  /** Assert that the sidebar eventually contains the target menu label. */
  async expectSidebarContains(label: string, timeout: number = 10000) {
    await expect(this.sidebar).toContainText(label, { timeout });
  }

  /** Assert that the sidebar eventually removes the target menu label. */
  async expectSidebarNotContains(label: string, timeout: number = 10000) {
    await expect(this.sidebar).not.toContainText(label, { timeout });
  }

  async expectLayoutHeightStable(
    sampleCount: number = 4,
    intervalMs: number = 400,
  ) {
    const samples: number[] = [];

    for (let index = 0; index < sampleCount; index += 1) {
      const height = await this.page.evaluate(() => {
        return document.documentElement.scrollHeight;
      });
      samples.push(height);
      if (index < sampleCount - 1) {
        await this.page.waitForTimeout(intervalMs);
      }
    }

    const minHeight = Math.min(...samples);
    const maxHeight = Math.max(...samples);

    if (maxHeight - minHeight > 16) {
      throw new Error(`菜单管理页高度未稳定，采样结果: ${samples.join(", ")}`);
    }
  }
}
