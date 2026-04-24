import type { APIRequestContext, APIResponse } from "@playwright/test";

import { request as playwrightRequest } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { config } from "../../../fixtures/config";
import { createAdminApiContext } from "../../../fixtures/plugin";
import { LoginPage } from "../../../pages/LoginPage";
import { MainLayout } from "../../../pages/MainLayout";
import { PluginPage } from "../../../pages/PluginPage";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const publicBaseURL =
  process.env.E2E_PUBLIC_BASE_URL ?? apiBaseURL.replace(/\/api\/v1\/?$/, "");
const pluginID = "demo-control";
const lifecyclePluginID = "plugin-demo-source";
const demoControlMessage = "演示模式已开启，禁止执行写操作";
const demoControlSkipReason =
  "requires demo-control to be enabled via plugin.autoEnable";

type PluginListItem = {
  autoEnableManaged?: number;
  enabled?: number;
  id: string;
  installed?: number;
};

function unwrapApiData(payload: any) {
  if (payload && typeof payload === "object" && "data" in payload) {
    return payload.data;
  }
  return payload;
}

async function expectApiOK(response: APIResponse, message: string) {
  expect(response.ok(), `${message}, status=${response.status()}`).toBeTruthy();
  const payload = await response.json();
  expect(payload.code, payload.message).toBe(0);
  return payload;
}

async function expectDemoControlRejected(
  response: APIResponse,
  message: string,
) {
  expect(response.status(), `${message}, status=${response.status()}`).toBe(
    403,
  );
  expect(await response.text(), message).toContain(demoControlMessage);
}

async function fetchPlugin(
  adminApi: APIRequestContext,
  targetPluginID: string,
): Promise<PluginListItem | null> {
  const response = await adminApi.get("plugins");
  const payload = await expectApiOK(response, "查询插件列表失败");
  const list = unwrapApiData(payload)?.list ?? [];
  return (
    list.find(
      (item: PluginListItem) => item.id === targetPluginID,
    ) ?? null
  );
}

async function expectPluginState(
  adminApi: APIRequestContext,
  targetPluginID: string,
  installed: number,
  enabled: number,
) {
  const plugin = await fetchPlugin(adminApi, targetPluginID);
  expect(plugin, `未找到插件 ${targetPluginID}`).toBeTruthy();
  expect(plugin?.installed, `${targetPluginID} installed 状态不符合预期`).toBe(
    installed,
  );
  expect(plugin?.enabled, `${targetPluginID} enabled 状态不符合预期`).toBe(
    enabled,
  );
}

test.describe("TC-105 demo-control 全局只读保护", () => {
  let adminApi: APIRequestContext;
  let demoControlManaged = false;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    const pluginResponse = await adminApi.get("plugins");
    expect(
      pluginResponse.ok(),
      `查询插件列表失败, status=${pluginResponse.status()}`,
    ).toBeTruthy();
    const pluginPayload = unwrapApiData(await pluginResponse.json());
    const demoControl = (pluginPayload?.list ?? []).find(
      (item: Record<string, unknown>) => item.id === pluginID,
    );
    demoControlManaged =
      demoControl?.installed === 1 &&
      demoControl?.enabled === 1 &&
      demoControl?.autoEnableManaged === 1;
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC-105a: 插件管理页展示 demo-control 已被 plugin.autoEnable 管理并处于启用状态", async ({
    adminPage,
  }) => {
    test.skip(!demoControlManaged, demoControlSkipReason);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    await expect(pluginPage.pluginRow(pluginID)).toContainText("演示控制");
    await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveAttribute(
      "aria-checked",
      "true",
    );
    await expect(pluginPage.pluginAutoEnableTag(pluginID)).toBeVisible();

    await pluginPage.openPluginDetail(pluginID);
    await expect(pluginPage.pluginDetailModal()).toContainText(pluginID);
    await expect(pluginPage.pluginDetailModal()).toContainText(
      "plugin.autoEnable",
    );
    await expect(pluginPage.pluginAutoEnableDetailAlert()).toContainText(
      "宿主下次重启后会再次安装并启用该插件",
    );
  });

  test("TC-105b: 演示模式仍允许管理员通过登录页登录并登出", async ({
    page,
  }) => {
    test.skip(!demoControlManaged, demoControlSkipReason);

    const loginPage = new LoginPage(page);
    const mainLayout = new MainLayout(page);

    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    await expect(page).not.toHaveURL(/auth\/login/);

    await mainLayout.logout();
    await expect(page).toHaveURL(/auth\/login/);
  });

  test("TC-105c: 演示模式继续放行已认证查询请求", async () => {
    test.skip(!demoControlManaged, demoControlSkipReason);

    const infoResponse = await adminApi.get("system/info");
    const infoPayload = await expectApiOK(infoResponse, "查询系统信息失败");
    expect(infoPayload.code, infoPayload.message).toBe(0);
    expect(infoPayload.data?.framework?.name).toBeTruthy();

    const demoControl = await fetchPlugin(adminApi, pluginID);
    expect(demoControl, `未找到插件 ${pluginID}`).toBeTruthy();
    expect(demoControl?.installed).toBe(1);
    expect(demoControl?.enabled).toBe(1);
    expect(demoControl?.autoEnableManaged).toBe(1);
  });

  test("TC-105d: 演示模式拒绝宿主 API 写操作并返回只读提示", async () => {
    test.skip(!demoControlManaged, demoControlSkipReason);

    await expectDemoControlRejected(
      await adminApi.post("config", {
        data: {
          key: "e2e.demo.control.blocked",
          name: "E2E Demo Control Blocked",
          remark: "should be blocked by demo-control",
          value: "1",
        },
      }),
      "POST /api/v1/config 应被拦截",
    );

    await expectDemoControlRejected(
      await adminApi.put("config/999999999", {
        data: {
          key: "e2e.demo.control.blocked",
          name: "E2E Demo Control Blocked",
          remark: "should be blocked by demo-control",
          value: "2",
        },
      }),
      "PUT /api/v1/config/999999999 应被拦截",
    );

    await expectDemoControlRejected(
      await adminApi.delete("config/999999999"),
      "DELETE /api/v1/config/999999999 应被拦截",
    );
  });

  test("TC-105e: 演示模式在 /* 作用域下拦截非 API 写请求并放行只读访问", async () => {
    test.skip(!demoControlManaged, demoControlSkipReason);

    const publicRequest = await playwrightRequest.newContext({
      baseURL: publicBaseURL,
    });

    try {
      const getResponse = await publicRequest.get("/");
      expect(
        getResponse.ok(),
        `GET / 应继续放行, status=${getResponse.status()}`,
      ).toBeTruthy();

      await expectDemoControlRejected(
        await publicRequest.post("/", {
          data: {
            probe: "demo-control-root-scope",
          },
        }),
        "POST / 应被全局作用域拦截",
      );
    } finally {
      await publicRequest.dispose();
    }
  });

  test("TC-105f: 演示模式允许其他插件治理白名单操作但拒绝修改 demo-control 自身", async () => {
    test.skip(!demoControlManaged, demoControlSkipReason);

    const originalLifecyclePlugin = await fetchPlugin(adminApi, lifecyclePluginID);
    expect(originalLifecyclePlugin, `未找到插件 ${lifecyclePluginID}`).toBeTruthy();

    const restoreLifecyclePluginState = async () => {
      if (!originalLifecyclePlugin) {
        return;
      }
      const currentPlugin = await fetchPlugin(adminApi, lifecyclePluginID);
      if (!currentPlugin) {
        return;
      }

      if (originalLifecyclePlugin.installed !== 1) {
        if (currentPlugin.installed === 1) {
          await expectApiOK(
            await adminApi.delete(`plugins/${lifecyclePluginID}`),
            `恢复卸载插件 ${lifecyclePluginID} 失败`,
          );
        }
        return;
      }

      if (currentPlugin.installed !== 1) {
        await expectApiOK(
          await adminApi.post(`plugins/${lifecyclePluginID}/install`),
          `恢复安装插件 ${lifecyclePluginID} 失败`,
        );
        await expectPluginState(adminApi, lifecyclePluginID, 1, 0);
      }

      if (originalLifecyclePlugin.enabled === 1) {
        const refreshedPlugin = await fetchPlugin(adminApi, lifecyclePluginID);
        if (refreshedPlugin?.enabled !== 1) {
          await expectApiOK(
            await adminApi.put(`plugins/${lifecyclePluginID}/enable`),
            `恢复启用插件 ${lifecyclePluginID} 失败`,
          );
        }
      } else {
        const refreshedPlugin = await fetchPlugin(adminApi, lifecyclePluginID);
        if (refreshedPlugin?.enabled === 1) {
          await expectApiOK(
            await adminApi.put(`plugins/${lifecyclePluginID}/disable`),
            `恢复禁用插件 ${lifecyclePluginID} 失败`,
          );
        }
      }
    };

    try {
      const currentPlugin = await fetchPlugin(adminApi, lifecyclePluginID);
      if (currentPlugin?.installed === 1) {
        await expectApiOK(
          await adminApi.delete(`plugins/${lifecyclePluginID}`),
          `卸载插件 ${lifecyclePluginID} 失败`,
        );
        await expectPluginState(adminApi, lifecyclePluginID, 0, 0);
      }

      await expectApiOK(
        await adminApi.post(`plugins/${lifecyclePluginID}/install`),
        `安装插件 ${lifecyclePluginID} 失败`,
      );
      await expectPluginState(adminApi, lifecyclePluginID, 1, 0);

      await expectApiOK(
        await adminApi.put(`plugins/${lifecyclePluginID}/enable`),
        `启用插件 ${lifecyclePluginID} 失败`,
      );
      await expectPluginState(adminApi, lifecyclePluginID, 1, 1);

      await expectApiOK(
        await adminApi.put(`plugins/${lifecyclePluginID}/disable`),
        `禁用插件 ${lifecyclePluginID} 失败`,
      );
      await expectPluginState(adminApi, lifecyclePluginID, 1, 0);

      await expectApiOK(
        await adminApi.delete(`plugins/${lifecyclePluginID}`),
        `卸载插件 ${lifecyclePluginID} 失败`,
      );
      await expectPluginState(adminApi, lifecyclePluginID, 0, 0);

      await expectDemoControlRejected(
        await adminApi.put(`plugins/${pluginID}/disable`),
        "demo-control 自身禁用请求应继续被拦截",
      );
      await expectPluginState(adminApi, pluginID, 1, 1);
    } finally {
      await restoreLifecyclePluginState();
    }
  });
});
