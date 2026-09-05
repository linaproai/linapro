import { mkdir } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { test, expect } from "../../fixtures/auth";
import { LayoutAuditPage } from "../../pages/LayoutAuditPage";

test.describe("TC-8 侧栏收放时顶栏工具保持定位", () => {
  for (const width of [1081, 1440]) {
    test(
      "TC-8" +
        (width === 1081 ? "a" : "b") +
        ": " +
        width +
        "px 下双向收放与中途反向",
      async ({ authenticatedPage }, testInfo) => {
        await authenticatedPage.setViewportSize({ width, height: 900 });
        const layout = new LayoutAuditPage(authenticatedPage);
        await layout.goto("/platform/tenants", { tableSelector: ".vxe-table" });
        await expect(
          authenticatedPage.getByText("租户列表", { exact: true }),
        ).toBeVisible();
        await expect(
          authenticatedPage.getByText("加载菜单中...", { exact: true }),
        ).toBeHidden();

        const stamp = new Date().toISOString().replaceAll(/[-:.TZ]/g, "");
        const directory = fileURLToPath(
          new URL(
            "../../../../temp/" + stamp.slice(0, 8) + "/",
            import.meta.url,
          ),
        );
        await mkdir(directory, { recursive: true });
        // 完成四轮后恢复初始状态，认证固件会销毁当前测试的独立浏览器会话。
        for (const [index, reverse] of [false, true, false, true].entries()) {
          await authenticatedPage.screenshot({
            path: path.join(
              directory,
              stamp.slice(8) +
                "-sidebar-" +
                width +
                "-before-" +
                index +
                ".png",
            ),
          });
          const result = await layout.measureSidebarToggle(reverse);
          await testInfo.attach("sidebar-" + index, {
            body: JSON.stringify(result),
            contentType: "application/json",
          });
          await authenticatedPage.screenshot({
            path: path.join(
              directory,
              stamp.slice(8) + "-sidebar-" + width + "-after-" + index + ".png",
            ),
          });
          const { initial, frames } = result;
          const final = frames.at(-1)!;
          expect(initial.toolPositions.length).toBeGreaterThan(0);
          expect(Math.abs(initial.headerRight - width)).toBeLessThanOrEqual(1);
          expect(Math.abs(initial.toolbarRight - width)).toBeLessThanOrEqual(1);
          expect(result.reversedWhileAnimating).toBe(reverse);
          // 至少两帧宽度有实际变化，避免只采到终态或关闭动画后误通过。
          expect(
            frames.filter(
              (frame) =>
                Math.abs(frame.headerWidth - initial.headerWidth) > 1 &&
                (reverse ||
                  Math.abs(frame.headerWidth - final.headerWidth) > 1),
            ).length,
          ).toBeGreaterThanOrEqual(2);
          expect(final.sidebarWidth > 1).toBe(
            reverse ? initial.sidebarWidth > 1 : initial.sidebarWidth <= 1,
          );
          for (const frame of frames) {
            expect(
              Math.abs(frame.headerRight - initial.headerRight),
            ).toBeLessThanOrEqual(1);
            expect(
              Math.abs(frame.toolbarLeft - initial.toolbarLeft),
            ).toBeLessThanOrEqual(1);
            expect(
              Math.abs(frame.toolbarRight - initial.toolbarRight),
            ).toBeLessThanOrEqual(1);
            for (const [toolIndex, position] of frame.toolPositions.entries()) {
              expect(
                Math.abs(position - initial.toolPositions[toolIndex]!),
              ).toBeLessThanOrEqual(1);
            }
          }
        }
      },
    );
  }
});
