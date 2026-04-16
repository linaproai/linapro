import type { Page } from '@playwright/test';

export class DeptPage {
  constructor(private page: Page) {}

  /** The Vben drawer container */
  private get drawer() {
    return this.page.locator('[role="dialog"]');
  }

  async goto() {
    await this.page.goto('/system/dept');
    await this.page.waitForLoadState('networkidle');
    // Wait for VxeGrid table to render
    await this.page
      .locator('.vxe-table')
      .waitFor({ state: 'visible', timeout: 10000 });
  }

  /** Click "展开" toolbar button to expand all tree nodes */
  async expandAll() {
    await this.page
      .getByRole('button', { name: /展\s*开/ })
      .first()
      .click();
    await this.page.waitForTimeout(500);
  }

  /** Click "折叠" toolbar button to collapse all tree nodes */
  async collapseAll() {
    await this.page
      .getByRole('button', { name: /折\s*叠/ })
      .first()
      .click();
    await this.page.waitForTimeout(500);
  }

  /** Create a root dept by clicking "新增" toolbar button */
  async createRootDept(name: string, opts?: { code?: string }) {
    // Click the primary "新增" button in toolbar (not the row-level "新增" buttons)
    await this.page
      .locator('.vxe-grid--toolbar')
      .getByRole('button', { name: /新\s*增/ })
      .click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Fill dept name (first text input in drawer)
    const nameInput = this.drawer.locator('input[placeholder="请输入"]').first();
    await nameInput.fill(name);

    // Fill dept code if provided (second text input in drawer)
    if (opts?.code) {
      const codeInput = this.drawer
        .locator('input[placeholder="请输入"]')
        .nth(1);
      await codeInput.fill(opts.code);
    }

    // Click confirm button
    await this.drawer
      .getByRole('button', { name: /确\s*认/ })
      .click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Create a sub dept under the specified parent row */
  async createSubDept(
    parentName: string,
    name: string,
    opts?: { code?: string },
  ) {
    // Find the parent row and click the "新增" action button (green, btn-success)
    const parentRow = this.page.locator('.vxe-body--row', {
      hasText: parentName,
    });
    await parentRow
      .locator('.btn-success')
      .first()
      .click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Fill dept name (first text input in drawer)
    const nameInput = this.drawer.locator('input[placeholder="请输入"]').first();
    await nameInput.fill(name);

    // Fill dept code if provided (second text input in drawer)
    if (opts?.code) {
      const codeInput = this.drawer
        .locator('input[placeholder="请输入"]')
        .nth(1);
      await codeInput.fill(opts.code);
    }

    // Click confirm button
    await this.drawer
      .getByRole('button', { name: /确\s*认/ })
      .click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Edit a dept: find the row, click edit, update fields in drawer */
  async editDept(
    deptName: string,
    newName: string,
    opts?: { code?: string },
  ) {
    // Find the row and click the edit button
    const row = this.page.locator('.vxe-body--row', { hasText: deptName });
    await row
      .getByRole('button', { name: /编\s*辑/ })
      .first()
      .click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Clear and fill the new name (first text input)
    const nameInput = this.drawer.locator('input[placeholder="请输入"]').first();
    await nameInput.clear();
    await nameInput.fill(newName);

    // Fill dept code if provided (second text input)
    if (opts?.code) {
      const codeInput = this.drawer
        .locator('input[placeholder="请输入"]')
        .nth(1);
      await codeInput.clear();
      await codeInput.fill(opts.code);
    }

    // Click confirm button
    await this.drawer
      .getByRole('button', { name: /确\s*认/ })
      .click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Delete a dept: find the row, click delete, confirm in Popconfirm */
  async deleteDept(deptName: string) {
    // Find the row and click the delete ghost button
    const row = this.page.locator('.vxe-body--row', { hasText: deptName });
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

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Check if a dept row with the given name is visible */
  async hasDept(deptName: string): Promise<boolean> {
    return this.page
      .locator('.vxe-body--row', { hasText: deptName })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Check if a dept row with the given name has the expected code */
  async hasDeptWithCode(deptName: string, code: string): Promise<boolean> {
    const row = this.page.locator('.vxe-body--row', { hasText: deptName });
    const hasRow = await row
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
    if (!hasRow) return false;
    return row
      .locator('.vxe-body--column', { hasText: code })
      .first()
      .isVisible({ timeout: 2000 })
      .catch(() => false);
  }
}
