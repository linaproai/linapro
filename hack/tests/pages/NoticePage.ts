import type { Page } from '@playwright/test';

export class NoticePage {
  constructor(private page: Page) {}

  /** The Vben modal container */
  private get modal() {
    return this.page.locator('[role="dialog"]');
  }

  async goto() {
    await this.page.goto('/system/notice');
    await this.page.waitForLoadState('networkidle');
    await this.page
      .locator('.vxe-table')
      .waitFor({ state: 'visible', timeout: 10000 });
  }

  /** Create a new notice */
  async createNotice(
    title: string,
    type: '通知' | '公告',
    status: '草稿' | '已发布',
    content?: string,
  ) {
    await this.page
      .getByRole('button', { name: /新\s*增/ })
      .first()
      .click();

    // Wait for modal to open
    await this.modal.waitFor({ state: 'visible', timeout: 5000 });
    await this.page.waitForTimeout(500);

    // Fill title - use placeholder to find the input inside the modal
    const titleInput = this.modal
      .getByPlaceholder('请输入公告标题')
      .first();
    await titleInput.fill(title);

    // Select status (RadioButton) - using label text since they're button-style radios
    await this.modal
      .locator('.ant-radio-button-wrapper', { hasText: status })
      .click();

    // Select type (RadioButton)
    await this.modal
      .locator('.ant-radio-button-wrapper', { hasText: type })
      .click();

    // Type content in Tiptap editor if provided
    if (content) {
      const editor = this.modal.locator('.tiptap[contenteditable="true"]');
      await editor.waitFor({ state: 'visible', timeout: 5000 });
      await editor.click();
      await this.page.keyboard.type(content, { delay: 20 });
    }

    // Click confirm button (modal footer)
    await this.modal
      .getByRole('button', { name: /确\s*认/ })
      .click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Edit a notice: search by title, click edit, update title */
  async editNotice(searchTitle: string, newTitle: string) {
    await this.fillSearchField('公告标题', searchTitle);
    await this.clickSearch();

    await this.page
      .getByRole('button', { name: /编\s*辑/ })
      .first()
      .click();

    await this.modal.waitFor({ state: 'visible', timeout: 5000 });
    await this.page.waitForTimeout(1000);

    const titleInput = this.modal
      .getByPlaceholder('请输入公告标题')
      .first();
    await titleInput.clear();
    await titleInput.fill(newTitle);

    await this.modal
      .getByRole('button', { name: /确\s*认/ })
      .click();

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Delete a notice: search by title, click delete, confirm */
  async deleteNotice(title: string) {
    await this.fillSearchField('公告标题', title);
    await this.clickSearch();

    await this.page
      .locator('.ant-btn-sm')
      .filter({ hasText: /删\s*除/ })
      .first()
      .click();

    await this.page.waitForTimeout(500);
    const popconfirm = this.page.locator('.ant-popconfirm, .ant-popover');
    const confirmBtn = popconfirm.getByRole('button', {
      name: /确\s*定|OK|是/i,
    });
    if (await confirmBtn.isVisible({ timeout: 2000 }).catch(() => false)) {
      await confirmBtn.click();
    }

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(500);
  }

  /** Check if a notice with the given title is visible */
  async hasNotice(title: string): Promise<boolean> {
    return this.page
      .locator('.vxe-body--row', { hasText: title })
      .first()
      .isVisible({ timeout: 5000 })
      .catch(() => false);
  }

  /** Preview a notice: search by title, click preview button */
  async previewNotice(title: string) {
    await this.fillSearchField('公告标题', title);
    await this.clickSearch();

    await this.page
      .getByRole('button', { name: /预\s*览/ })
      .first()
      .click();

    await this.modal.waitFor({ state: 'visible', timeout: 5000 });
    await this.page.waitForTimeout(500);
  }

  /** Fill search form field by label */
  async fillSearchField(label: string, value: string) {
    const input = this.page.getByLabel(label, { exact: true }).first();
    await input.clear();
    await input.fill(value);
  }

  /** Click search button */
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

  /** Get total count from pager */
  async getTotalCount(): Promise<number> {
    const pager = this.page.locator('.vxe-pager--total');
    const text = await pager.textContent();
    const match = text?.match(/(\d+)/);
    return match ? parseInt(match[1], 10) : 0;
  }
}
