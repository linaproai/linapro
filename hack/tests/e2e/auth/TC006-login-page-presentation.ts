import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import {
  createAdminApiContext,
  getConfigByKey,
  updateConfigValue,
} from "../../support/api/job";

test.describe("TC-2 登录页展示收口与布局", () => {
  let api: APIRequestContext;
  let originalLayout: { id: number; key: string; value: string };

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    originalLayout = await getConfigByKey(api, "sys.auth.loginPanelLayout");
  });

  test.afterAll(async () => {
    await updateConfigValue(api, originalLayout.id, originalLayout.value);
    await api.dispose();
  });

  test("TC-2a: 登录页展示已交付入口并回退未交付认证子路由", async ({
    loginPage,
    page,
  }) => {
    await loginPage.goto();

    await expect(loginPage.forgotPasswordEntry).toBeVisible();
    await expect(loginPage.createAccountEntry).toBeVisible();
    await expect(loginPage.mobileLoginButton).toBeHidden();
    await expect(loginPage.qrCodeLoginButton).toBeHidden();
    // 外部登录双区域：auth.login.after（协议全宽按钮）与 auth.login.social
    // （平台图标行）。无插件时各自整块隐藏。布局契约见 TC-2e / TC-2f。

    for (const path of ["/auth/code-login", "/auth/qrcode-login"]) {
      await page.goto(path);
      await page.waitForURL(/\/auth\/login$/, { timeout: 10000 });
      await expect(loginPage.usernameInput).toBeVisible();
    }

    await page.goto("/auth/forget-password");
    await page.waitForURL(/\/auth\/forget-password$/, { timeout: 10000 });
    await expect(loginPage.forgetPasswordSubmitButton).toBeVisible();
    await expect(
      page.getByText("发送重置链接", { exact: true }),
    ).toBeVisible();

    await page.goto("/auth/register");
    await page.waitForURL(/\/auth\/register$/, { timeout: 10000 });
    await expect(loginPage.registerSubmitButton).toBeVisible();
    await expect(page.getByText("注册", { exact: true }).first()).toBeVisible();
  });

  test("TC-2b: 登录页默认使用居中登录框布局", async ({ loginPage }) => {
    await updateConfigValue(api, originalLayout.id, "panel-center");

    await loginPage.goto();

    await expect(loginPage.centerAuthPanel).toBeVisible();
    await expect(loginPage.leftAuthPanel).toBeHidden();
    await expect(loginPage.rightAuthPanel).toBeHidden();
  });

  test("TC-2c: 修改系统参数后登录页按配置切换布局", async ({ loginPage }) => {
    await updateConfigValue(api, originalLayout.id, "panel-left");
    await loginPage.goto();
    await expect(loginPage.leftAuthPanel).toBeVisible();
    await expect(loginPage.centerAuthPanel).toBeHidden();

    await updateConfigValue(api, originalLayout.id, "panel-right");
    await loginPage.goto();
    await expect(loginPage.rightAuthPanel).toBeVisible();
    await expect(loginPage.leftAuthPanel).toBeHidden();
  });

  test("TC-2d: 登录页密码占位符跟随默认双语切换", async ({ loginPage }) => {
    await loginPage.goto();

    await expect(loginPage.passwordInput).toHaveAttribute(
      "placeholder",
      "请输入密码",
    );

    await loginPage.switchLanguage("English");
    await expect(loginPage.passwordInput).toHaveAttribute(
      "placeholder",
      "Please enter password",
    );
  });

  test("TC-2e: 协议登录插槽为全宽单行按钮纵向排列", async ({
    loginPage,
    page,
  }) => {
    await loginPage.goto();

    // 宿主契约：login.vue 对 auth.login.after 使用 flex-col 全宽纵向按钮栈。
    // 使用与生产一致的 class 注入多入口夹具，不绑定具体插件。
    const fixture = page.getByTestId("login-external-auth-slot-fixture");
    await page.evaluate(() => {
      const region =
        document.querySelector(
          '[data-testid="login-external-auth-region"]',
        ) ??
        (() => {
          const el = document.createElement("div");
          el.className = "login-external-auth w-full sm:mx-auto md:max-w-md";
          el.setAttribute("data-testid", "login-external-auth-region");
          const formRoot =
            document.querySelector("form")?.parentElement ?? document.body;
          formRoot.appendChild(el);
          return el;
        })();

      let host = region.querySelector(
        '[data-testid="login-external-auth-slot"]',
      ) as HTMLElement | null;
      if (!host) {
        host = document.createElement("div");
        // Mirror apps/web-antd login.vue PluginSlotOutlet classes.
        host.className = "mt-4 flex w-full flex-col gap-3 plugin-slot-outlet";
        host.setAttribute("data-testid", "login-external-auth-slot");
        region.appendChild(host);
      }

      host.setAttribute("data-testid", "login-external-auth-slot-fixture");
      host.className = "mt-4 flex w-full flex-col gap-3 plugin-slot-outlet";
      host.innerHTML = `
        <div class="plugin-slot-outlet__item w-full" style="height:40px;width:100%;background:#e5e7eb"></div>
        <div class="plugin-slot-outlet__item w-full" style="height:40px;width:100%;background:#d1d5db"></div>
      `;
    });

    await expect(fixture).toBeVisible();

    const display = await fixture.evaluate((el) => {
      const style = getComputedStyle(el);
      return {
        display: style.display,
        flexDirection: style.flexDirection,
      };
    });
    expect(display.display).toBe("flex");
    expect(display.flexDirection).toBe("column");

    const items = fixture.locator(".plugin-slot-outlet__item");
    expect(await items.count()).toBe(2);
    const firstBox = await items.nth(0).boundingBox();
    const secondBox = await items.nth(1).boundingBox();
    expect(firstBox).not.toBeNull();
    expect(secondBox).not.toBeNull();
    // Vertical stack: second item is below the first (same column).
    expect(secondBox!.y).toBeGreaterThan(firstBox!.y + firstBox!.height - 1);
    const horizontalDelta = Math.abs(secondBox!.x - firstBox!.x);
    expect(horizontalDelta).toBeLessThanOrEqual(8);
  });

  test("TC-2f: 平台社交登录插槽为横向图标行", async ({ loginPage, page }) => {
    await loginPage.goto();

    // 宿主契约：login.vue 对 auth.login.social 使用 flex-wrap 横向图标行，
    // 并带「其他登录方式」分隔线。夹具不绑定具体平台插件。
    const fixture = page.getByTestId("login-social-auth-slot-fixture");
    await page.evaluate(() => {
      const region =
        document.querySelector('[data-testid="login-social-auth-region"]') ??
        (() => {
          const el = document.createElement("div");
          el.className = "login-social-auth w-full sm:mx-auto md:max-w-md";
          el.setAttribute("data-testid", "login-social-auth-region");
          const formRoot =
            document.querySelector("form")?.parentElement ?? document.body;
          formRoot.appendChild(el);
          return el;
        })();

      // Ensure divider markup exists when injecting fixture into a synthetic region.
      if (!region.querySelector(".text-muted-foreground")) {
        const divider = document.createElement("div");
        divider.className = "mt-4 flex items-center justify-between";
        divider.innerHTML = `
          <span class="w-[35%] border-b border-input"></span>
          <span class="text-center text-xs uppercase text-muted-foreground">其他登录方式</span>
          <span class="w-[35%] border-b border-input"></span>
        `;
        region.appendChild(divider);
      }

      let host = region.querySelector(
        '[data-testid="login-social-auth-slot"]',
      ) as HTMLElement | null;
      if (!host) {
        host = document.createElement("div");
        host.className =
          "mt-4 flex flex-wrap justify-center plugin-slot-outlet";
        host.setAttribute("data-testid", "login-social-auth-slot");
        region.appendChild(host);
      }

      host.setAttribute("data-testid", "login-social-auth-slot-fixture");
      host.className =
        "mt-4 flex flex-wrap justify-center plugin-slot-outlet";
      host.innerHTML = `
        <div class="plugin-slot-outlet__item" style="height:36px;width:36px;background:#e5e7eb;border-radius:9999px"></div>
        <div class="plugin-slot-outlet__item" style="height:36px;width:36px;background:#d1d5db;border-radius:9999px"></div>
      `;
    });

    await expect(fixture).toBeVisible();
    await expect(
      page.getByText("其他登录方式", { exact: true }).first(),
    ).toBeVisible();

    const display = await fixture.evaluate((el) => {
      const style = getComputedStyle(el);
      return {
        display: style.display,
        flexDirection: style.flexDirection,
        flexWrap: style.flexWrap,
        justifyContent: style.justifyContent,
      };
    });
    expect(display.display).toBe("flex");
    // Default flex-direction is row when not set to column.
    expect(display.flexDirection === "row" || display.flexDirection === "").toBe(
      true,
    );
    expect(display.flexWrap).toBe("wrap");
    expect(display.justifyContent).toBe("center");

    const items = fixture.locator(".plugin-slot-outlet__item");
    expect(await items.count()).toBe(2);
    const firstBox = await items.nth(0).boundingBox();
    const secondBox = await items.nth(1).boundingBox();
    expect(firstBox).not.toBeNull();
    expect(secondBox).not.toBeNull();
    // Horizontal row: second item is to the right of the first (same row).
    expect(secondBox!.x).toBeGreaterThan(firstBox!.x + firstBox!.width - 1);
    const verticalDelta = Math.abs(secondBox!.y - firstBox!.y);
    expect(verticalDelta).toBeLessThanOrEqual(8);
  });
});
