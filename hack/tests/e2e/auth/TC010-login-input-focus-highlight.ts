import type { Locator } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";

/**
 * Regression coverage for form-field control wrapper layout:
 *
 * 1) Focus ring: Vben Input uses focus-visible:ring-1 (box-shadow). When the
 *    control row applies overflow-x-hidden without matching p-px, the ring is
 *    clipped and only corner fragments remain visible.
 *
 * 2) Validation message: FormMessage is absolutely positioned under the control
 *    column. overflow-x-hidden must stay on the INNER control row only — putting
 *    it on the outer column clips FormMessage (CSS treats overflow-y as auto when
 *    overflow-x is not visible).
 */

type FocusStyleSnapshot = {
  boxShadow: string;
  borderColor: string;
  outlineStyle: string;
};

async function readFocusStyle(locator: Locator): Promise<FocusStyleSnapshot> {
  return locator.evaluate((el) => {
    const style = getComputedStyle(el);
    return {
      boxShadow: style.boxShadow,
      borderColor: style.borderColor,
      outlineStyle: style.outlineStyle,
    };
  });
}

/** True when box-shadow paints a non-empty ring/glow (not "none" / all-zero). */
function hasVisibleBoxShadow(boxShadow: string): boolean {
  if (!boxShadow || boxShadow === "none") {
    return false;
  }
  // Reject fully transparent placeholders if the whole shadow is transparent.
  if (/rgba?\(\s*0\s*,\s*0\s*,\s*0\s*,\s*0\s*\)/.test(boxShadow)) {
    const layers = boxShadow.split(/,(?![^(]*\))/);
    return layers.some(
      (layer) => !/rgba?\(\s*0\s*,\s*0\s*,\s*0\s*,\s*0\s*\)/.test(layer),
    );
  }
  return true;
}

/**
 * Walk ancestors and ensure any overflow-x clip that can cut the focus ring
 * also reserves at least 1px padding (p-px contract with ring-1).
 */
async function assertFocusRingNotClipped(locator: Locator) {
  const clipInfo = await locator.evaluate((el) => {
    let node: HTMLElement | null = el.parentElement;
    while (node && node !== document.body) {
      const style = getComputedStyle(node);
      const clipsX =
        style.overflowX === "hidden" ||
        style.overflowX === "clip" ||
        style.overflow === "hidden" ||
        style.overflow === "clip";
      if (clipsX) {
        const paddingLeft = Number.parseFloat(style.paddingLeft) || 0;
        const paddingRight = Number.parseFloat(style.paddingRight) || 0;
        return {
          clipsX: true,
          paddingLeft,
          paddingRight,
          className: node.className,
        };
      }
      node = node.parentElement;
    }
    return { clipsX: false, paddingLeft: 0, paddingRight: 0, className: "" };
  });

  if (clipInfo.clipsX) {
    expect(
      clipInfo.paddingLeft,
      `overflow-x clip ancestor "${clipInfo.className}" must reserve left padding for focus ring`,
    ).toBeGreaterThanOrEqual(1);
    expect(
      clipInfo.paddingRight,
      `overflow-x clip ancestor "${clipInfo.className}" must reserve right padding for focus ring`,
    ).toBeGreaterThanOrEqual(1);
  }
}

test.describe("TC-10 登录页输入域焦点高亮", () => {
  test("TC-10a: 点击用户名输入域时出现完整焦点 ring（非四角残影）", async ({
    loginPage,
    page,
  }) => {
    await loginPage.goto();
    await expect(loginPage.usernameInput).toBeVisible();

    await loginPage.usernameInput.click();
    await expect(loginPage.usernameInput).toBeFocused();

    const style = await readFocusStyle(loginPage.usernameInput);
    expect(
      hasVisibleBoxShadow(style.boxShadow),
      `expected focus ring box-shadow, got: ${style.boxShadow}`,
    ).toBe(true);

    await assertFocusRingNotClipped(loginPage.usernameInput);

    const day = new Date().toISOString().slice(0, 10).replace(/-/g, "");
    const stamp = new Date()
      .toISOString()
      .replace(/[:.]/g, "-")
      .slice(0, 19);
    await page.screenshot({
      path: `../../temp/${day}/${stamp}-login-input-focus-username.png`,
      fullPage: false,
    });
  });

  test("TC-10b: 空提交显示校验错误，且重新聚焦仍保留完整焦点高亮", async ({
    loginPage,
    page,
  }) => {
    await loginPage.goto();
    await expect(loginPage.loginButton).toBeVisible();

    // Empty submit → field-level validation errors on username/password.
    await loginPage.loginButton.click();

    // FormMessage uses text-destructive; must be visible (not overflow-clipped).
    const fieldErrors = page.locator("p.text-destructive");
    await expect(fieldErrors.first()).toBeVisible({ timeout: 5000 });
    await expect(fieldErrors.first()).not.toBeEmpty();
    // At least one error message has non-zero layout box (not clipped to 0 height).
    const errorBox = await fieldErrors.first().boundingBox();
    expect(
      errorBox && errorBox.height > 0 && errorBox.width > 0,
      "validation FormMessage must have a visible layout box (not overflow-clipped)",
    ).toBe(true);

    await expect(loginPage.usernameInput).toHaveClass(/border-destructive/);

    await loginPage.usernameInput.click();
    await expect(loginPage.usernameInput).toBeFocused();

    const style = await readFocusStyle(loginPage.usernameInput);
    // Error state may use red focus shadow or primary ring; either must be a full
    // non-clipped box-shadow, not corner-only fragments from overflow clipping.
    expect(
      hasVisibleBoxShadow(style.boxShadow),
      `expected focus highlight box-shadow after validation error, got: ${style.boxShadow}`,
    ).toBe(true);

    await assertFocusRingNotClipped(loginPage.usernameInput);

    const day = new Date().toISOString().slice(0, 10).replace(/-/g, "");
    const stamp = new Date()
      .toISOString()
      .replace(/[:.]/g, "-")
      .slice(0, 19);
    await page.screenshot({
      path: `../../temp/${day}/${stamp}-login-input-focus-after-validation.png`,
      fullPage: false,
    });
  });

  test("TC-10c: 密码输入域焦点高亮同样不被裁切", async ({ loginPage }) => {
    await loginPage.goto();
    await loginPage.passwordInput.click();
    await expect(loginPage.passwordInput).toBeFocused();

    const style = await readFocusStyle(loginPage.passwordInput);
    expect(
      hasVisibleBoxShadow(style.boxShadow),
      `expected password focus ring box-shadow, got: ${style.boxShadow}`,
    ).toBe(true);
    await assertFocusRingNotClipped(loginPage.passwordInput);
  });
});
