import type { Page } from '@playwright/test';

export class RolePage {
  constructor(private page: Page) {}

  /** The Vben drawer container */
  private get drawer() {
    return this.page.locator('[role="dialog"]');
  }

  async goto() {
    await this.page.goto('/system/role');
    await this.page.waitForLoadState('networkidle');
    await this.page.locator('.vxe-table').first().waitFor({ state: 'visible', timeout: 10000 });
    await this.page.getByLabel('角色名称', { exact: true }).first().waitFor({
      state: 'visible',
      timeout: 10000,
    });
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
    await this.page.waitForLoadState('load');
    await this.page.waitForTimeout(2000);

    // Click the primary "新增" button in toolbar (use first() as there may be multiple buttons)
    await this.page
      .getByRole('button', { name: /新\s*增/ })
      .first()
      .click({ force: true });

    // Wait for drawer to open
    await this.drawer.waitFor({ state: 'visible', timeout: 10000 });
    await this.page.waitForTimeout(500);

    // Fill text fields first (these work even with tour overlay present)
    const nameInput = this.drawer.locator('input[placeholder="请输入角色名称"]');
    await nameInput.waitFor({ state: 'visible', timeout: 5000 });
    await nameInput.fill(params.name);

    const codeInput = this.drawer.locator('input[placeholder="如: admin, user等"]');
    await codeInput.fill(params.code);

    if (params.sort !== undefined) {
      const sortInput = this.drawer.getByRole('spinbutton');
      await sortInput.fill(String(params.sort));
    }

    if (params.remark) {
      const remarkInput = this.drawer.locator('textarea[placeholder="请输入备注"]');
      await remarkInput.fill(params.remark);
    }

    // Wait for form to fully render
    await this.page.waitForTimeout(1000);

    // Dismiss tour overlay if present
    const endTourBtn = this.page.getByRole('button', { name: '结束导览' });
    if (await endTourBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await endTourBtn.click({ force: true });
      await this.page.waitForTimeout(500);
    }
    const tourClose = this.page.locator('.ant-tour-close');
    if (await tourClose.isVisible({ timeout: 300 }).catch(() => false)) {
      await tourClose.click({ force: true });
      await this.page.waitForTimeout(300);
    }

    // Select status - RadioGroup with button style
    // Default is already '正常' (value 1), so we only need to click if status is 0
    const statusValue = params.status ?? 1;
    if (statusValue === 0) {
      // Click on "停用" radio button
      await this.drawer.locator('.ant-radio-button-wrapper').filter({ hasText: '停用' }).click();
      await this.page.waitForTimeout(300);
    }

    // Select data scope (required field) - click the label text to select
    // The RadioGroup uses ant-radio-button-wrapper with button style
    const dataScopeLabel = this.drawer.getByText('全部数据权限', { exact: true });
    await dataScopeLabel.waitFor({ state: 'visible', timeout: 5000 });
    await dataScopeLabel.click({ force: true });
    await this.page.waitForTimeout(500);

    // Verify the radio is selected (ant-radio-button-wrapper-checked class)
    const checkedRadio = this.drawer.locator('.ant-radio-button-wrapper-checked');
    await checkedRadio.waitFor({ state: 'visible', timeout: 3000 }).catch(() => {
      // If not visible, try clicking again
      return dataScopeLabel.click({ force: true });
    });

    // Select menus if needed - for basic test we skip menu selection
    // Menu selection is tested separately in TC0061e

    // Click confirm button - scroll into view first since dialog may be taller than viewport
    const confirmBtn = this.drawer.getByRole('button', { name: /确\s*认/ });
    await confirmBtn.scrollIntoViewIfNeeded();
    await confirmBtn.click({ force: true });

    await this.page.waitForLoadState('load');
    await this.drawer.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
    await this.page.waitForTimeout(500);
  }

  /** Edit a role: find the row, click edit, update fields in drawer */
  async editRole(roleName: string, newName: string) {
    // Find the row and click the edit button
    const row = this.page.locator('.vxe-body--row', { hasText: roleName });
    await row.getByRole('button', { name: /编\s*辑/ }).first().click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });
    await this.page.waitForTimeout(500);

    // Dismiss any tour overlay that might block clicks
    const endTourBtn = this.page.getByRole('button', { name: '结束导览' });
    if (await endTourBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await endTourBtn.click({ force: true });
      await this.page.waitForTimeout(500);
    }
    const tourClose = this.page.locator('.ant-tour-close');
    if (await tourClose.isVisible({ timeout: 300 }).catch(() => false)) {
      await tourClose.click({ force: true });
      await this.page.waitForTimeout(300);
    }

    // Clear and fill the new name
    const nameInput = this.drawer.locator('input[placeholder="请输入角色名称"]');
    await nameInput.clear();
    await nameInput.fill(newName);

    // Click confirm button - scroll into view first since dialog may be taller than viewport
    const confirmBtn = this.drawer.getByRole('button', { name: /确\s*认/ });
    await confirmBtn.scrollIntoViewIfNeeded();
    await confirmBtn.click({ force: true });

    await this.page.waitForLoadState('load');
    await this.drawer.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
    await this.page.waitForTimeout(500);
  }

  /** Delete a role: find the row, click delete, confirm in Popconfirm */
  async deleteRole(roleName: string) {
    // Find the row and click the delete ghost button
    const row = this.page.locator('.vxe-body--row', { hasText: roleName });
    await row
      .locator('.ant-btn-sm')
      .filter({ hasText: /删\s*除/ })
      .first()
      .click();

    // Confirm in Popconfirm
    await this.page.waitForTimeout(500);
    const popconfirm = this.page.locator('.ant-popconfirm, .ant-popover');
    const confirmBtn = popconfirm.getByRole('button', {
      name: /确\s*定|OK|是/i,
    });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      const modal = this.page.locator('.ant-modal-confirm');
      await modal.getByRole('button', { name: /确\s*定|OK/i }).click();
    }

    await this.page.waitForLoadState('load');
    await this.page.waitForTimeout(500);
  }

  /** Check if a role row with the given name is visible */
  async hasRole(roleName: string): Promise<boolean> {
    return this.page
      .locator('.vxe-body--row', { hasText: roleName })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Search role by name */
  async searchRole(name: string) {
    // Prefer the accessible label because it stays stable even when the form DOM is re-created.
    await this.page.waitForLoadState('networkidle');
    const searchInput = this.page.getByLabel('角色名称', { exact: true }).first();
    await searchInput.waitFor({ state: 'visible', timeout: 10000 });
    await searchInput.fill(name);

    // Click search button
    await this.page.getByRole('button', { name: /搜\s*索/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Reset search */
  async resetSearch() {
    await this.page.getByRole('button', { name: /重\s*置/ }).first().click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Toggle role status */
  async toggleStatus(roleName: string) {
    const row = this.page.locator('.vxe-body--row', { hasText: roleName });
    const switchBtn = row.locator('.ant-switch');
    await switchBtn.click();
    await this.page.waitForLoadState('load');
    await this.page.waitForTimeout(500);
  }

  /** Click assign button to go to role-auth page */
  async clickAssign(roleName: string) {
    const row = this.page.locator('.vxe-body--row', { hasText: roleName });
    await row.getByRole('button', { name: /分\s*配/ }).first().click();
    await this.page.waitForLoadState('load');
    await this.page.waitForTimeout(500);
  }

  /** Check menu in the menu tree table (for role edit) */
  async checkMenu(menuName: string) {
    const menuTree = this.drawer.locator('.vxe-table');
    const menuRow = menuTree.locator('.vxe-body--row', { hasText: menuName });
    const checkbox = menuRow.locator('.vxe-checkbox--icon');
    await checkbox.click({ force: true });
    await this.page.waitForTimeout(300);
  }

  /** Uncheck menu in the menu tree table (for role edit) */
  async uncheckMenu(menuName: string) {
    const menuTree = this.drawer.locator('.vxe-table');
    const menuRow = menuTree.locator('.vxe-body--row', { hasText: menuName });
    const checkbox = menuRow.locator('.vxe-checkbox--icon');
    await checkbox.click({ force: true });
    await this.page.waitForTimeout(300);
  }

  /** Get checked menu count in drawer */
  async getCheckedMenuCount(): Promise<number> {
    const menuTree = this.drawer.locator('.vxe-table');
    const checkedRows = menuTree.locator('.vxe-body--row.is--checked');
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
    await this.page.waitForLoadState('load');
    await this.page.waitForTimeout(2000);

    await this.page
      .getByRole('button', { name: /新\s*增/ })
      .first()
      .click({ force: true });

    await this.drawer.waitFor({ state: 'visible', timeout: 10000 });
    await this.page.waitForTimeout(500);

    const nameInput = this.drawer.locator('input[placeholder="请输入角色名称"]');
    await nameInput.waitFor({ state: 'visible', timeout: 5000 });
    await nameInput.fill(params.name);

    const codeInput = this.drawer.locator('input[placeholder="如: admin, user等"]');
    await codeInput.fill(params.code);

    if (params.sort !== undefined) {
      const sortInput = this.drawer.getByRole('spinbutton');
      await sortInput.fill(String(params.sort));
    }

    if (params.remark) {
      const remarkInput = this.drawer.locator('textarea[placeholder="请输入备注"]');
      await remarkInput.fill(params.remark);
    }

    const endTourBtn = this.page.getByRole('button', { name: '结束导览' });
    if (await endTourBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await endTourBtn.click({ force: true });
      await this.page.waitForTimeout(500);
    }
    const tourClose = this.page.locator('.ant-tour-close');
    if (await tourClose.isVisible({ timeout: 300 }).catch(() => false)) {
      await tourClose.click({ force: true });
      await this.page.waitForTimeout(300);
    }

    const dataScopeLabel = this.drawer.getByText('全部数据权限', { exact: true });
    await dataScopeLabel.waitFor({ state: 'visible', timeout: 5000 });
    await dataScopeLabel.click({ force: true });
    await this.page.waitForTimeout(500);

    if (params.menuNames && params.menuNames.length > 0) {
      await this.drawer.locator('.vxe-table').waitFor({ state: 'visible', timeout: 5000 });
      for (const menuName of params.menuNames) {
        await this.checkMenu(menuName);
      }
    }

    const confirmBtn = this.drawer.getByRole('button', { name: /确\s*认/ });
    await confirmBtn.scrollIntoViewIfNeeded();
    await confirmBtn.click({ force: true });

    await this.page.waitForLoadState('load');
    await this.drawer.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
    await this.page.waitForTimeout(500);
  }

  /** Assign menus to existing role */
  async assignMenusToRole(roleName: string, menuNames: string[]) {
    const row = this.page.locator('.vxe-body--row', { hasText: roleName });
    await row.getByRole('button', { name: /编\s*辑/ }).first().click();

    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Wait for menu tree
    await this.drawer.locator('.vxe-table').waitFor({ state: 'visible', timeout: 3000 });

    // Clear existing selections - expand all and uncheck all
    const menuTree = this.drawer.locator('.vxe-table');
    const allCheckboxes = menuTree.locator('.vxe-checkbox--icon');
    const count = await allCheckboxes.count();
    for (let i = 0; i < count; i++) {
      const checkbox = allCheckboxes.nth(i);
      const row = checkbox.locator('xpath=..');
      const isChecked = await row.evaluate((el) => el.classList.contains('is--checked'));
      if (isChecked) {
        await checkbox.click();
        await this.page.waitForTimeout(100);
      }
    }

    // Select new menus
    for (const menuName of menuNames) {
      await this.checkMenu(menuName);
    }

    await this.drawer.getByRole('button', { name: /确\s*认/ }).click();
    await this.page.waitForLoadState('load');
    await this.drawer.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
    await this.page.waitForTimeout(500);
  }

  /** Navigate to role management page */
  async navigateTo() {
    await this.page.goto('/system/role');
    await this.page.waitForLoadState('load');
    await this.page
      .locator('.vxe-table')
      .waitFor({ state: 'visible', timeout: 5000 })
      .catch(() => {});
  }

  /** Check if status switch is disabled for a role */
  async isStatusSwitchDisabled(roleName: string): Promise<boolean> {
    await this.searchRole(roleName);
    const switchEl = this.page.locator('.vxe-body--row .ant-switch').first();
    return switchEl.evaluate((el) => el.classList.contains('ant-switch-disabled'));
  }

  /** Check if checkbox is disabled for a role */
  async isCheckboxDisabled(roleName: string): Promise<boolean> {
    await this.searchRole(roleName);
    const checkbox = this.page.locator('.vxe-body--row .vxe-cell--checkbox').first();
    return checkbox.evaluate((el) => el.classList.contains('is--disabled'));
  }
}
