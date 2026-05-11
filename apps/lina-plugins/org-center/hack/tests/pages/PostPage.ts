import type { Page } from "@host-tests/support/playwright";

import {
  waitForBusyIndicatorsToClear,
  waitForConfirmOverlay,
  waitForDialogReady,
  waitForDropdown,
  waitForRouteReady,
  waitForTableReady,
} from "@host-tests/support/ui";

export class PostPage {
  constructor(private page: Page) {}

  /** The Vben drawer container */
  private get drawer() {
    return this.page.locator('[role="dialog"]');
  }

  async goto() {
    await this.page.goto("/system/post");
    await waitForTableReady(this.page);
  }

  /** Click a dept node in the left DeptTree sidebar */
  async selectDept(deptName: string) {
    const treeNode = this.page
      .locator(".ant-tree-node-content-wrapper", { hasText: deptName })
      .first();
    await treeNode.click();
    await waitForRouteReady(this.page);
  }

  /** Create a new post by clicking toolbar "新增", filling the drawer */
  async createPost(deptName: string, code: string, name: string) {
    await this.page
      .getByRole("button", { name: /新\s*增/ })
      .first()
      .click();

    // Wait for drawer to open
    await waitForDialogReady(this.drawer);

    // The drawer marks required labels with a leading "*", so match by role/name.
    await this.drawer.getByRole("combobox", { name: /所属部门/ }).click();
    // Wait for tree dropdown to appear and select the dept
    const dropdown = await waitForDropdown(this.page);
    const deptNode = dropdown
      .locator(".ant-select-tree-treenode", { hasText: deptName })
      .first()
      .locator(".ant-select-tree-title")
      .first();
    await deptNode.click({ force: true });
    await dropdown
      .waitFor({ state: "hidden", timeout: 5000 })
      .catch(async () => {
        await this.page.keyboard.press("Escape");
      });
    await waitForBusyIndicatorsToClear(this.page);

    await this.drawer.getByRole("textbox", { name: /岗位名称/ }).fill(name);
    await this.drawer.getByRole("textbox", { name: /岗位编码/ }).fill(code);

    // Click confirm button
    await this.drawer.getByRole("button", { name: /确\s*认/ }).click();

    await waitForRouteReady(this.page);
    await this.drawer
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
  }

  /** Edit a post: search by code, click edit, update name in drawer */
  async editPost(code: string, newName: string) {
    // Search for the post by code first to narrow results
    await this.fillSearchField("岗位编码", code);
    await this.clickSearch();

    await this.page
      .locator("button:visible:not([disabled])")
      .filter({ hasText: /编\s*辑/ })
      .first()
      .click();

    // Wait for drawer to open
    await waitForDialogReady(this.drawer);

    const nameInput = this.drawer.getByRole("textbox", { name: /岗位名称/ });
    await nameInput.clear();
    await nameInput.fill(newName);

    // Click confirm button
    await this.drawer.getByRole("button", { name: /确\s*认/ }).click();

    await waitForRouteReady(this.page);
    await this.drawer
      .waitFor({ state: "hidden", timeout: 10000 })
      .catch(() => {});
  }

  /** Delete a post: search by code, click delete, confirm */
  async deletePost(code: string) {
    // Search for the post by code first
    await this.fillSearchField("岗位编码", code);
    await this.clickSearch();

    await this.page
      .locator("button:visible:not([disabled])")
      .filter({ hasText: /删\s*除/ })
      .first()
      .click();

    // Confirm in Popconfirm
    const popconfirm = await waitForConfirmOverlay(this.page);
    const confirmBtn = popconfirm.getByRole("button", {
      name: /确\s*定|OK|是/i,
    });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    } else {
      const modal = this.page.locator(".ant-modal-confirm");
      await modal.getByRole("button", { name: /确\s*定|OK/i }).click();
    }

    await waitForRouteReady(this.page);
  }

  /** Check if a post with the given code is visible in the table */
  async hasPost(code: string): Promise<boolean> {
    await this.fillSearchField("岗位编码", code);
    await this.clickSearch();
    return this.page
      .locator(".vxe-body--row:visible", { hasText: code })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Check if a post row with the given display name is visible in the table. */
  async hasPostName(name: string): Promise<boolean> {
    return this.page
      .locator(".vxe-body--row:visible", { hasText: name })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Click export button */
  async clickExport() {
    await this.page.getByRole("button", { name: /导\s*出/ }).click();
    await waitForDialogReady(this.page.locator('[role="dialog"]'));
  }

  /** Select a row by clicking its checkbox (search by code first) */
  async selectRow(code: string) {
    await this.fillSearchField("岗位编码", code);
    await this.clickSearch();
    // Click the first checkbox in body rows
    const checkbox = this.page
      .locator(".vxe-body--row .vxe-checkbox--icon")
      .first();
    await checkbox.click();
    await waitForBusyIndicatorsToClear(this.page);
  }

  /** Click the toolbar batch delete button */
  async batchDelete() {
    // The toolbar delete button is a danger primary button
    await this.page
      .locator(".vxe-grid--toolbar, .vxe-toolbar")
      .getByRole("button", { name: /删\s*除/ })
      .click();

    // Confirm in Modal.confirm
    const overlay = await waitForConfirmOverlay(this.page);
    const modalConfirm = overlay
      .getByRole("button", { name: /确\s*定|OK/i })
      .last();
    if (await modalConfirm.isVisible({ timeout: 1000 }).catch(() => false)) {
      await modalConfirm.click();
    } else {
      const modal = this.page.locator(".ant-modal-confirm");
      await modal.getByRole("button", { name: /确\s*定|OK/i }).click();
    }

    await waitForRouteReady(this.page);
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
      .getByRole("button", { name: /搜\s*索/ })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  /** Click reset button */
  async clickReset() {
    await this.page
      .getByRole("button", { name: /重\s*置/ })
      .first()
      .click();
    await waitForRouteReady(this.page);
  }

  /** Get the total count from the pager */
  async getTotalCount(): Promise<number> {
    const pager = this.page.locator(".vxe-pager--total");
    const text = await pager.textContent();
    const match = text?.match(/(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }
}
