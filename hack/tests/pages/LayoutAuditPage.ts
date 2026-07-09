import { expect, type Locator, type Page } from "@playwright/test";

import { waitForRouteReady, waitForTableReady } from "../support/ui";

type ElementBox = {
  height: number;
  width: number;
  x: number;
  y: number;
};

export class LayoutAuditPage {
  constructor(private page: Page) {}

  async goto(path: string, options?: { tableSelector?: string }) {
    await this.page.goto(path);
    if (options?.tableSelector) {
      await waitForTableReady(this.page, options.tableSelector);
      return;
    }
    await waitForRouteReady(this.page);
  }

  panel(id: string): Locator {
    return this.page.locator(`#${id}`).first();
  }

  formLabel(text: RegExp | string, scope?: Locator): Locator {
    return (scope ?? this.page)
      .locator(".ant-form-item-label, .ant-form-item-label label, label", {
        hasText: text,
      })
      .first();
  }

  searchForm(): Locator {
    return this.page
      .locator(".vxe-grid form")
      .filter({
        has: this.page.getByRole("button", { name: /搜\s*索|Search/u }),
      })
      .filter({
        has: this.page.getByRole("button", { name: /重\s*置|Reset/u }),
      })
      .first();
  }

  searchCollapseToggle(): Locator {
    return this.searchForm()
      .locator(".vben-link", { hasText: /展\s*开|收\s*起|Expand|Collapse/u })
      .first();
  }

  searchFormLabel(text: RegExp | string): Locator {
    return this.formLabel(text, this.searchForm());
  }

  searchResetButton(): Locator {
    return this.searchForm()
      .getByRole("button", { name: /重\s*置|Reset/u })
      .first();
  }

  searchSubmitButton(): Locator {
    return this.searchForm()
      .getByRole("button", { name: /搜\s*索|Search/u })
      .first();
  }

  async expectSearchCollapseHidden() {
    await expect(this.searchCollapseToggle()).toHaveCount(0);
  }

  async expectSearchCollapseVisible() {
    await expect(this.searchCollapseToggle()).toBeVisible();
  }

  async expectSearchLabelHidden(text: RegExp | string) {
    await expect(this.searchFormLabel(text)).toBeHidden();
  }

  async expectSearchLabelVisible(text: RegExp | string) {
    await expect(this.searchFormLabel(text)).toBeVisible();
  }

  async toggleSearchCollapse() {
    const toggle = this.searchCollapseToggle();
    await expect(toggle).toBeVisible();
    await toggle.click();
  }

  async expectSearchControlsOnOneRow(labels: string[]) {
    const controls = [
      ...labels.map((label) =>
        this.searchForm().getByLabel(label, { exact: true }).first(),
      ),
      this.searchResetButton(),
      this.searchSubmitButton(),
    ];

    const boxes = await Promise.all(
      controls.map((control, index) =>
        this.visibleBoundingBox(control, `search-control-${index}`),
      ),
    );
    const centerYList = boxes.map((box) => box.y + box.height / 2);
    expect(Math.max(...centerYList) - Math.min(...centerYList)).toBeLessThan(
      32,
    );

    for (let index = 0; index < boxes.length; index += 1) {
      for (
        let nextIndex = index + 1;
        nextIndex < boxes.length;
        nextIndex += 1
      ) {
        expect(this.boxesOverlap(boxes[index]!, boxes[nextIndex]!)).toBe(false);
      }
    }
  }

  tableHeader(text: RegExp | string, scope?: Locator): Locator {
    return (scope ?? this.page)
      .locator(".vxe-header--column, th", { hasText: text })
      .first();
  }

  private boxesOverlap(first: ElementBox, second: ElementBox) {
    return (
      first.x < second.x + second.width - 1 &&
      first.x + first.width > second.x + 1 &&
      first.y < second.y + second.height - 1 &&
      first.y + first.height > second.y + 1
    );
  }

  private async visibleBoundingBox(locator: Locator, name: string) {
    await expect(locator, `${name} should be visible`).toBeVisible();
    const box = await locator.boundingBox();
    expect(box, `${name} should have a bounding box`).not.toBeNull();
    return box!;
  }
}
