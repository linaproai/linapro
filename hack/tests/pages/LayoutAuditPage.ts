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

  async measureSidebarToggle(reverseDuringAnimation = false) {
    const header = this.page.locator("header");
    const toggle = header.locator(":scope > button").first();
    await expect(toggle).toBeVisible();
    await this.page.evaluate(() => document.fonts.ready.then(() => undefined));

    // 点击前安装探针，避免跨进程往返漏掉动画开头；不修改生产样式。
    const probe = await header.evaluateHandle((element, reverse) => {
      const wrapper = element.parentElement!;
      const toolbar = element.lastElementChild!;
      const button =
        element.querySelector<HTMLButtonElement>(":scope > button")!;
      const aside = document.querySelector("aside")!;
      const tools = Array.from(toolbar.querySelectorAll("button"));
      const read = () => ({
        headerRight: wrapper.getBoundingClientRect().right,
        headerWidth: wrapper.getBoundingClientRect().width,
        sidebarWidth: aside.getBoundingClientRect().width,
        toolbarLeft: toolbar.getBoundingClientRect().left,
        toolbarRight: toolbar.getBoundingClientRect().right,
        toolPositions: tools.map((tool) => tool.getBoundingClientRect().left),
      });
      type Frame = ReturnType<typeof read>;
      let frameId = 0;
      let timeoutId = 0;
      let onClick: () => void;
      const cancel = () => {
        cancelAnimationFrame(frameId);
        clearTimeout(timeoutId);
        button.removeEventListener("click", onClick, true);
      };
      const result = new Promise<{
        frames: Frame[];
        initial: Frame;
        reversedWhileAnimating: boolean;
      }>((resolve, reject) => {
        onClick = () => {
          const initial = read();
          const frames: Frame[] = [];
          const startedAt = performance.now();
          let reversedWhileAnimating = false;
          let stableFrames = 0;
          const sample = () => {
            const frame = read();
            frames.push(frame);
            // 反向时两种过渡的剩余时长不同，必须等待顶栏和侧栏都结束。
            const running = [
              ...wrapper.getAnimations(),
              ...aside.getAnimations(),
            ].some((animation) => animation.playState === "running");
            if (
              reverse &&
              !reversedWhileAnimating &&
              running &&
              performance.now() - startedAt >= 60 &&
              Math.abs(frame.headerWidth - initial.headerWidth) > 1
            ) {
              // 普通 Playwright 点击会等待元素稳定；此处需要在真实过渡中反向。
              reversedWhileAnimating = true;
              button.click();
              stableFrames = 0;
            } else {
              stableFrames = running ? 0 : stableFrames + 1;
            }
            if (stableFrames >= 2) {
              cancel();
              resolve({ frames, initial, reversedWhileAnimating });
              return;
            }
            frameId = requestAnimationFrame(sample);
          };
          frameId = requestAnimationFrame(sample);
        };
        button.addEventListener("click", onClick, {
          capture: true,
          once: true,
        });
        timeoutId = window.setTimeout(() => {
          cancel();
          reject(new Error("侧栏动画采样未在 5 秒内结束"));
        }, 5000);
      });
      return { cancel, result };
    }, reverseDuringAnimation);

    try {
      await toggle.click();
      return await probe.evaluate((value) => value.result);
    } finally {
      await probe.evaluate((value) => value.cancel());
      await probe.dispose();
    }
  }

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
