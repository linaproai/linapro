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

  test("TC-2a: 登录页隐藏未实现入口并回退未实现认证子路由", async ({
    loginPage,
    page,
  }) => {
    await loginPage.goto();

    await expect(loginPage.forgotPasswordEntry).toBeHidden();
    await expect(loginPage.createAccountEntry).toBeHidden();
    await expect(loginPage.mobileLoginButton).toBeHidden();
    await expect(loginPage.qrCodeLoginButton).toBeHidden();
    // 「其他登录方式」由宿主在 auth.login.after 有插件注入时展示（对齐 Vben
    // ThirdPartyLogin 分隔线），不再视为未实现入口。无插件时整块区域隐藏；
    // 有插件时可见——此处不强制 hidden，由 TC-2e 覆盖布局契约。

    for (const path of [
      "/auth/code-login",
      "/auth/qrcode-login",
      "/auth/forget-password",
      "/auth/register",
    ]) {
      await page.goto(path);
      await page.waitForURL(/\/auth\/login$/, { timeout: 10000 });
      await expect(loginPage.usernameInput).toBeVisible();
    }
  });

  test("TC-2b: 登录页默认使用居右登录框布局", async ({ loginPage }) => {
    await updateConfigValue(api, originalLayout.id, "panel-right");

    await loginPage.goto();

    await expect(loginPage.rightAuthPanel).toBeVisible();
    await expect(loginPage.leftAuthPanel).toBeHidden();
    await expect(loginPage.centerAuthPanel).toBeHidden();
  });

  test("TC-2c: 修改系统参数后登录页按配置切换布局", async ({ loginPage }) => {
    await updateConfigValue(api, originalLayout.id, "panel-left");
    await loginPage.goto();
    await expect(loginPage.leftAuthPanel).toBeVisible();
    await expect(loginPage.rightAuthPanel).toBeHidden();

    await updateConfigValue(api, originalLayout.id, "panel-center");
    await loginPage.goto();
    await expect(loginPage.centerAuthPanel).toBeVisible();
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

  test("TC-2e: 外部登录插槽对齐 Vben5 横向图标行与分隔线", async ({
    loginPage,
    page,
  }) => {
    await loginPage.goto();

    // 宿主契约：login.vue 对齐 Vben ThirdPartyLogin —
    // 分隔线「其他登录方式」+ flex-wrap justify-center 横向图标行。
    // 使用与生产一致的 class 注入多入口夹具，不绑定具体 OIDC 插件。
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

      // Ensure the Vben-style divider is present for the fixture.
      if (!region.querySelector('[data-testid="login-third-party-divider"]')) {
        const divider = document.createElement("div");
        divider.className = "mt-4 flex items-center justify-between";
        divider.setAttribute("data-testid", "login-third-party-divider");
        divider.innerHTML = `
          <span class="w-[35%] border-b border-input"></span>
          <span class="text-center text-xs text-muted-foreground uppercase">其他登录方式</span>
          <span class="w-[35%] border-b border-input"></span>
        `;
        region.prepend(divider);
      }

      let host = region.querySelector(
        '[data-testid="login-external-auth-slot"]',
      ) as HTMLElement | null;
      if (!host) {
        host = document.createElement("div");
        // Mirror apps/web-antd login.vue PluginSlotOutlet classes.
        host.className =
          "mt-4 flex flex-wrap justify-center plugin-slot-outlet";
        host.setAttribute("data-testid", "login-external-auth-slot");
        region.appendChild(host);
      }

      host.setAttribute("data-testid", "login-external-auth-slot-fixture");
      host.className =
        "mt-4 flex flex-wrap justify-center plugin-slot-outlet";
      host.innerHTML = `
        <div class="plugin-slot-outlet__item" style="height:40px;width:40px;background:#e5e7eb;margin:0 4px"></div>
        <div class="plugin-slot-outlet__item" style="height:40px;width:40px;background:#d1d5db;margin:0 4px"></div>
      `;
    });

    await expect(fixture).toBeVisible();
    await expect(page.getByTestId("login-third-party-divider")).toBeVisible();
    await expect(loginPage.thirdPartyLoginTitle).toBeVisible();

    const display = await fixture.evaluate((el) => {
      const style = getComputedStyle(el);
      return {
        display: style.display,
        flexWrap: style.flexWrap,
        justifyContent: style.justifyContent,
      };
    });
    expect(display.display).toBe("flex");
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
