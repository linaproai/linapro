import { test, expect } from "../../../fixtures/auth";
import { RolePage } from "../../../pages/RolePage";
import {
  dismissTourOverlayIfPresent,
  waitForConfirmOverlay,
} from "../../../support/ui";

test.describe("TC-167 角色权限抽屉关闭交互", () => {
  test("TC-167a: 修改权限后取消关闭会提示并可确认关闭", async ({
    adminPage,
  }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();
    await adminPage.evaluate(() => {
      localStorage.setItem("menu_select_fullscreen_read", "true");
    });

    const drawer = await rolePage.openCreateDrawer();

    const selectedCount = drawer.getByTestId("menu-permission-selected-count");
    await expect(selectedCount).toBeVisible({ timeout: 5000 });

    const spacing = await drawer
      .getByTestId("menu-permission-toolbar")
      .evaluate((toolbar) => {
        const mode = toolbar.querySelector(
          '[data-testid="menu-permission-association-mode"]',
        );
        const count = toolbar.querySelector(
          '[data-testid="menu-permission-selected-count"]',
        );
        const modeRect = mode?.getBoundingClientRect();
        const countRect = count?.getBoundingClientRect();
        const columnGap = Number.parseFloat(
          getComputedStyle(toolbar).columnGap,
        );
        return {
          columnGap,
          gap: modeRect && countRect ? countRect.left - modeRect.right : 0,
          hasClass: count?.classList.contains("permission-selection-count"),
          toolbarWidth: toolbar.getBoundingClientRect().width,
        };
      });
    expect(spacing.hasClass).toBeTruthy();
    expect(spacing.columnGap).toBeGreaterThanOrEqual(16);
    expect(spacing.gap).toBeGreaterThanOrEqual(12);
    expect(spacing.toolbarWidth).toBeGreaterThan(0);

    await dismissTourOverlayIfPresent(adminPage);

    const permissionCheckbox = drawer
      .locator(".ant-checkbox-wrapper", {
        hasText: /查询|Search|新增|Add|编辑|Edit/,
      })
      .first();
    await permissionCheckbox.waitFor({ state: "visible", timeout: 10000 });
    await permissionCheckbox.click();

    await drawer.getByRole("button", { name: /取\s*消|Cancel/i }).click();
    const confirmOverlay = await waitForConfirmOverlay(adminPage);
    await expect(confirmOverlay).toContainText(/未保存的修改|unsaved changes/i);
    await dismissTourOverlayIfPresent(adminPage);
    await confirmOverlay
      .getByRole("button", { name: /确\s*认|OK|Confirm/i })
      .last()
      .click();

    await expect(drawer).toBeHidden({ timeout: 10000 });
  });
});
