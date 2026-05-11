import type { Page } from '@host-tests/support/playwright';

import {
  waitForBusyIndicatorsToClear,
  waitForConfirmOverlay,
  waitForDialogReady,
  waitForRouteReady,
  waitForTableReady,
} from '@host-tests/support/ui';

export class DeptPage {
  constructor(private page: Page) {}

  private resolveLocalizedLabel(label: string) {
    const labelMap: Record<string, RegExp> = {
      部门名称: /部门名称|Department Name|plugin\.org-center\.dept\.fields\.name/i,
      部门编码: /部门编码|Department Code|plugin\.org-center\.dept\.fields\.code/i,
    };
    const localizedLabel = labelMap[label];
    if (localizedLabel) {
      return this.page.getByLabel(localizedLabel).first();
    }
    return this.page.getByLabel(label, { exact: true }).first();
  }

  /** The Vben drawer container */
  private get drawer() {
    return this.page.locator('[role="dialog"]');
  }

  async goto() {
    await this.page.goto('/system/dept');
    await waitForTableReady(this.page);
  }

  /** Click "展开" toolbar button to expand all tree nodes */
  async expandAll() {
    await this.page
      .getByRole('button', { name: /展\s*开|Expand/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  /** Fill the search form field by label */
  async fillSearchField(label: string, value: string) {
    const input = this.resolveLocalizedLabel(label);
    await input.waitFor({ state: 'visible', timeout: 10000 });
    await input.clear();
    await input.fill(value);
  }

  /** Click search/query button */
  async clickSearch() {
    await this.page
      .getByRole('button', { name: /搜\s*索|Search/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  /** Click reset button */
  async clickReset() {
    await this.page
      .getByRole('button', { name: /重\s*置|Reset/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  /** Click "折叠" toolbar button to collapse all tree nodes */
  async collapseAll() {
    await this.page
      .getByRole('button', { name: /折\s*叠|Collapse/i })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  /** Create a root dept by clicking "新增" toolbar button */
  async createRootDept(name: string, opts?: { code?: string }) {
    // Click the primary "新增" button in toolbar (not the row-level "新增" buttons)
    await this.page
      .locator('.vxe-grid--toolbar')
      .getByRole('button', { name: /新\s*增/ })
      .click();

    await waitForDialogReady(this.drawer);

    // Fill dept name (first text input in drawer)
    const nameInput = this.drawer.getByLabel('部门名称', { exact: true });
    await nameInput.fill(name);

    // Fill dept code if provided.
    if (opts?.code) {
      const codeInput = this.drawer.getByLabel('部门编码', { exact: true });
      await codeInput.fill(opts.code);
    }

    // Click confirm button
    await this.drawer
      .getByRole('button', { name: /确\s*认/ })
      .click();

    await this.waitForDrawerSubmitToSettle();
  }

  /** Create a sub dept under the specified parent row */
  async createSubDept(
    parentName: string,
    name: string,
    opts?: { code?: string },
  ) {
    await this.fillSearchField('部门名称', parentName);
    await this.clickSearch();

    const parentRow = this.page.locator('.vxe-body--row:visible', {
      hasText: parentName,
    });
    await parentRow.first().waitFor({ state: 'visible', timeout: 10000 });
    await parentRow
      .locator('button:visible')
      .filter({ hasText: /新\s*增/ })
      .first()
      .click();

    await waitForDialogReady(this.drawer);

    // The drawer marks required labels with a leading "*", so match by role/name.
    const nameInput = this.drawer.getByRole('textbox', { name: /部门名称/ });
    await nameInput.fill(name);

    // Fill dept code if provided.
    if (opts?.code) {
      const codeInput = this.drawer.getByRole('textbox', { name: /部门编码/ });
      await codeInput.fill(opts.code);
    }

    // Click confirm button
    await this.drawer
      .getByRole('button', { name: /确\s*认/ })
      .click();

    await this.waitForDrawerSubmitToSettle();
  }

  /** Edit a dept: find the row, click edit, update fields in drawer */
  async editDept(
    deptName: string,
    newName: string,
    opts?: { code?: string },
  ) {
    await this.fillSearchField('部门名称', deptName);
    await this.clickSearch();

    const row = this.page.locator('.vxe-body--row:visible', { hasText: deptName });
    await row.first().waitFor({ state: 'visible', timeout: 10000 });
    await row
      .locator('button:visible')
      .filter({ hasText: /编\s*辑/ })
      .first()
      .click();

    await waitForDialogReady(this.drawer);

    // Clear and fill the new name (first text input)
    const nameInput = this.drawer.getByRole('textbox', { name: /部门名称/ });
    await nameInput.clear();
    await nameInput.fill(newName);

    // Fill dept code if provided.
    if (opts?.code) {
      const codeInput = this.drawer.getByRole('textbox', { name: /部门编码/ });
      await codeInput.clear();
      await codeInput.fill(opts.code);
    }

    // Click confirm button
    await this.drawer
      .getByRole('button', { name: /确\s*认/ })
      .click();

    await this.waitForDrawerSubmitToSettle();
  }

  /** Delete a dept: find the row, click delete, confirm in Popconfirm */
  async deleteDept(deptName: string) {
    await this.expandAll();
    await this.fillSearchField('部门名称', deptName);
    await this.clickSearch();

    // Tree rows may be collapsed or mirrored in fixed columns; use a visible
    // row plus a visible delete button inside that row.
    let row = this.page.locator('.vxe-body--row:visible', { hasText: deptName });
    const hasFilteredRow = await row
      .first()
      .isVisible({ timeout: 1500 })
      .catch(() => false);
    if (!hasFilteredRow) {
      await this.clickReset();
      await this.expandAll();
      await this.fillSearchField('部门名称', deptName);
      await this.clickSearch();
      row = this.page.locator('.vxe-body--row:visible', { hasText: deptName });
    }
    await row.first().waitFor({ state: 'visible', timeout: 10000 });
    await row
      .locator('button:visible')
      .filter({ hasText: /删\s*除/ })
      .first()
      .click();

    const popconfirm = await waitForConfirmOverlay(this.page);
    const confirmBtn = popconfirm.getByRole('button', {
      name: /确\s*定|OK|是/i,
    });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      const modal = this.page.locator('.ant-modal-confirm');
      await modal.getByRole('button', { name: /确\s*定|OK/i }).click();
    }

    await waitForRouteReady(this.page);
  }

  /** Check if a dept row with the given name is visible */
  async hasDept(deptName: string): Promise<boolean> {
    await this.fillSearchField('部门名称', deptName);
    await this.clickSearch();
    let row = this.page.locator('.vxe-body--row:visible', { hasText: deptName });
    const hasFilteredRow = await row
      .first()
      .isVisible({ timeout: 1500 })
      .catch(() => false);
    if (!hasFilteredRow) {
      await this.clickReset();
      await this.expandAll();
      await this.fillSearchField('部门名称', deptName);
      await this.clickSearch();
      row = this.page.locator('.vxe-body--row:visible', { hasText: deptName });
    }
    return row.first().isVisible({ timeout: 5000 }).catch(() => false);
  }

  /** Check whether the expanded tree contains the specified department name. */
  async hasDeptInExpandedTree(deptName: string): Promise<boolean> {
    await this.expandAll();
    return this.page
      .locator('.vxe-body--row:visible', { hasText: deptName })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Check if a dept row with the given name has the expected code */
  async hasDeptWithCode(deptName: string, code: string): Promise<boolean> {
    await this.fillSearchField('部门名称', deptName);
    await this.clickSearch();
    let row = this.page.locator('.vxe-body--row:visible', { hasText: deptName });
    let hasRow = await row
      .first()
      .isVisible({ timeout: 1500 })
      .catch(() => false);
    if (!hasRow) {
      await this.clickReset();
      await this.expandAll();
      await this.fillSearchField('部门名称', deptName);
      await this.clickSearch();
      row = this.page.locator('.vxe-body--row:visible', { hasText: deptName });
      hasRow = await row
        .first()
        .isVisible({ timeout: 5000 })
        .catch(() => false);
      if (!hasRow) {
        return false;
      }
    }
    const rowText = await row.first().textContent();
    return rowText?.includes(code) ?? false;
  }

  private async waitForDrawerSubmitToSettle() {
    await waitForRouteReady(this.page);
    const closed = await this.drawer
      .waitFor({ state: 'hidden', timeout: 1500 })
      .then(() => true)
      .catch(() => false);
    if (!closed) {
      await waitForBusyIndicatorsToClear(this.drawer);
    }
  }
}
