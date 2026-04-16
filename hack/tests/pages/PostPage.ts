import type { Page } from '@playwright/test';

export class PostPage {
  constructor(private page: Page) {}

  /** The Vben drawer container */
  private get drawer() {
    return this.page.locator('[role="dialog"]');
  }

  async goto() {
    await this.page.goto('/system/post');
    await this.page.waitForLoadState('networkidle');
    // Wait for VxeGrid table to render
    await this.page
      .locator('.vxe-table')
      .waitFor({ state: 'visible', timeout: 10000 });
  }

  /** Click a dept node in the left DeptTree sidebar */
  async selectDept(deptName: string) {
    const treeNode = this.page
      .locator('.ant-tree-node-content-wrapper', { hasText: deptName })
      .first();
    await treeNode.click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Create a new post by clicking toolbar "新增", filling the drawer */
  async createPost(deptName: string, code: string, name: string) {
    await this.page
      .getByRole('button', { name: /新\s*增/ })
      .first()
      .click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Select dept in TreeSelect (first .ant-select-selector in the drawer)
    await this.drawer.locator('.ant-select-selector').first().click();
    // Wait for tree dropdown to appear and select the dept
    await this.page.waitForTimeout(300);
    await this.page
      .locator('.ant-select-tree-node-content-wrapper', { hasText: deptName })
      .first()
      .click();
    await this.page.waitForTimeout(300);

    // Form order: deptId(TreeSelect), name(Input), code(Input), sort(InputNumber)
    // The "请输入" placeholder inputs: first is name, second is code
    const inputs = this.drawer.locator('input[placeholder="请输入"]');
    await inputs.nth(0).fill(name);
    await inputs.nth(1).fill(code);

    // Click confirm button
    await this.drawer
      .getByRole('button', { name: /确\s*认/ })
      .click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Edit a post: search by code, click edit, update name in drawer */
  async editPost(code: string, newName: string) {
    // Search for the post by code first to narrow results
    await this.fillSearchField('岗位编码', code);
    await this.clickSearch();

    // Click the edit button - with fixed column, buttons are in a separate overlay
    // Since we searched first, there's only one row, so click globally
    await this.page
      .getByRole('button', { name: /编\s*辑/ })
      .first()
      .click();

    // Wait for drawer to open
    await this.drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Clear and fill the new name (first "请输入" input is name)
    const nameInput = this.drawer.locator('input[placeholder="请输入"]').first();
    await nameInput.clear();
    await nameInput.fill(newName);

    // Click confirm button
    await this.drawer
      .getByRole('button', { name: /确\s*认/ })
      .click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Delete a post: search by code, click delete, confirm */
  async deletePost(code: string) {
    // Search for the post by code first
    await this.fillSearchField('岗位编码', code);
    await this.clickSearch();

    // Click the delete button - with fixed column, buttons are in a separate overlay
    // Since we searched first, there's only one row, so click globally
    await this.page
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

  /** Check if a post with the given code is visible in the table */
  async hasPost(code: string): Promise<boolean> {
    return this.page
      .locator('.vxe-body--row', { hasText: code })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Click export button */
  async clickExport() {
    await this.page
      .getByRole('button', { name: /导\s*出/ })
      .click();
    await this.page.waitForTimeout(2000);
  }

  /** Select a row by clicking its checkbox (search by code first) */
  async selectRow(code: string) {
    await this.fillSearchField('岗位编码', code);
    await this.clickSearch();
    // Click the first checkbox in body rows
    const checkbox = this.page
      .locator('.vxe-body--row .vxe-checkbox--icon')
      .first();
    await checkbox.click();
    await this.page.waitForTimeout(300);
  }

  /** Click the toolbar batch delete button */
  async batchDelete() {
    // The toolbar delete button is a danger primary button
    await this.page
      .locator('.vxe-grid--toolbar, .vxe-toolbar')
      .getByRole('button', { name: /删\s*除/ })
      .click();
    await this.page.waitForTimeout(500);

    // Confirm in Modal.confirm
    const modal = this.page.locator('.ant-modal-confirm');
    await modal
      .getByRole('button', { name: /确\s*定|OK/i })
      .click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Fill the search form field by label */
  async fillSearchField(label: string, value: string) {
    const input = this.page.getByLabel(label, { exact: true }).first();
    await input.clear();
    await input.fill(value);
  }

  /** Click search/query button */
  async clickSearch() {
    await this.page
      .getByRole('button', { name: /搜\s*索/ })
      .first()
      .click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Click reset button */
  async clickReset() {
    await this.page
      .getByRole('button', { name: /重\s*置/ })
      .first()
      .click();
    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Get the total count from the pager */
  async getTotalCount(): Promise<number> {
    const pager = this.page.locator('.vxe-pager--total');
    const text = await pager.textContent();
    const match = text?.match(/(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }
}
