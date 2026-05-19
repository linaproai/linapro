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
const pluginID = "linapro-ops-demo-guard";
const lifecyclePluginID = "linapro-demo-source";
const demoControlMessage = "演示模式已开启，禁止执行写操作";
const demoControlSkipReason =
  "requires linapro-ops-demo-guard to be installed and enabled";

type PluginListItem = {
  autoEnableManaged?: number;
  enabled?: number;
  id: string;
  installed?: number;
};

type LoginTenant = {
  code?: string;
  id: number;
  name?: string;
  status?: string;
};

type TenantLoginResult = {
  accessToken?: string;
  preToken?: string;
  refreshToken?: string;
  tenants?: LoginTenant[];
};

type TenantTokenResult = {
  accessToken?: string;
  refreshToken?: string;
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

async function loginDemoTenantUser(
  api: APIRequestContext,
): Promise<TenantLoginResult | null> {
  const response = await api.post("auth/login", {
    data: {
      password: "admin123",
      username: "tenant_alpha_ops",
    },
  });
  if (!response.ok()) {
    return null;
  }
  const payload = await response.json().catch(() => null);
  if (!payload || payload.code !== 0) {
    return null;
  }
  return unwrapApiData(payload) as TenantLoginResult;
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

test.describe("TC-105 linapro-ops-demo-guard 全局只读保护", () => {
  let adminApi: APIRequestContext;
  let demoControlEnabled = false;
  let demoControlAutoEnableManaged = false;

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
    demoControlEnabled =
      demoControl?.installed === 1 &&
      demoControl?.enabled === 1;
    demoControlAutoEnableManaged =
      demoControl?.installed === 1 &&
      demoControl?.enabled === 1 &&
      demoControl?.autoEnableManaged === 1;
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC-105a: 插件管理页展示 linapro-ops-demo-guard 已启用，并在 autoEnable 托管时显示对应提示", async ({
    adminPage,
  }) => {
    test.skip(!demoControlEnabled, demoControlSkipReason);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    await expect(pluginPage.pluginRow(pluginID)).toContainText("演示控制");
    await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveAttribute(
      "aria-checked",
      "true",
    );

    await pluginPage.openPluginDetail(pluginID);
    await expect(pluginPage.pluginDetailModal()).toContainText(pluginID);
    if (demoControlAutoEnableManaged) {
      await expect(pluginPage.pluginAutoEnableTag(pluginID)).toBeVisible();
      await expect(pluginPage.pluginDetailModal()).toContainText(
        "plugin.autoEnable",
      );
      await expect(pluginPage.pluginAutoEnableDetailAlert()).toContainText(
        "宿主下次重启后会再次安装并启用该插件",
      );
    } else {
      await expect(pluginPage.pluginAutoEnableTag(pluginID)).toHaveCount(0);
      await expect(pluginPage.pluginAutoEnableDetailAlert()).toHaveCount(0);
    }
  });

  test("TC-105b: 演示模式仍允许管理员通过登录页登录并登出", async ({
    page,
  }) => {
    test.skip(!demoControlEnabled, demoControlSkipReason);

    const loginPage = new LoginPage(page);
    const mainLayout = new MainLayout(page);

    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    await expect(page).not.toHaveURL(/auth\/login/);

    await mainLayout.logout();
    await expect(page).toHaveURL(/auth\/login/);
  });

  test("TC-105g: 演示模式允许多租户登录选择租户和会话 token 维护", async () => {
    test.skip(!demoControlEnabled, demoControlSkipReason);

    const sessionApi = await playwrightRequest.newContext({
      baseURL: apiBaseURL,
    });
    try {
      const login = await loginDemoTenantUser(sessionApi);
      const tenants = login?.tenants ?? [];
      test.skip(
        !login?.preToken || tenants.length < 2,
        "requires linapro-tenant-core mock user tenant_alpha_ops with multiple active tenants",
      );

      const preToken = login?.preToken ?? "";
      const selectedTenant = tenants[0]!;
      const switchTargetTenant = tenants[1]!;
      const selectPayload = unwrapApiData(
        await expectApiOK(
          await sessionApi.post("auth/select-tenant", {
            data: {
              preToken,
              tenantId: selectedTenant.id,
            },
          }),
          "演示模式下选择租户不应被只读保护拦截",
        ),
      ) as TenantTokenResult;
      expect(selectPayload.accessToken).toBeTruthy();
      expect(selectPayload.refreshToken).toBeTruthy();
      const refreshToken = selectPayload.refreshToken ?? "";

      const refreshPayload = unwrapApiData(
        await expectApiOK(
          await sessionApi.post("auth/refresh", {
            data: {
              refreshToken,
            },
          }),
          "演示模式下刷新会话 token 不应被只读保护拦截",
        ),
      ) as TenantTokenResult;
      expect(refreshPayload.accessToken).toBeTruthy();
      const accessToken = refreshPayload.accessToken ?? "";

      const tenantApi = await playwrightRequest.newContext({
        baseURL: apiBaseURL,
        extraHTTPHeaders: {
          Authorization: `Bearer ${accessToken}`,
        },
      });
      try {
        const switchPayload = unwrapApiData(
          await expectApiOK(
            await tenantApi.post("auth/switch-tenant", {
              data: {
                tenantId: switchTargetTenant.id,
              },
            }),
            "演示模式下切换已有租户不应被只读保护拦截",
          ),
        ) as TenantTokenResult;
        expect(switchPayload.accessToken).toBeTruthy();
      } finally {
        await tenantApi.dispose();
      }
    } finally {
      await sessionApi.dispose();
    }
  });

  test("TC-105c: 演示模式继续放行已认证查询请求", async () => {
    test.skip(!demoControlEnabled, demoControlSkipReason);

    const infoResponse = await adminApi.get("system/info");
    const infoPayload = await expectApiOK(infoResponse, "查询系统信息失败");
    expect(infoPayload.code, infoPayload.message).toBe(0);
    expect(infoPayload.data?.framework?.name).toBeTruthy();

    const demoControl = await fetchPlugin(adminApi, pluginID);
    expect(demoControl, `未找到插件 ${pluginID}`).toBeTruthy();
    expect(demoControl?.installed).toBe(1);
    expect(demoControl?.enabled).toBe(1);
  });

  test("TC-105d: 演示模式拒绝宿主 API 写操作并返回只读提示", async () => {
    test.skip(!demoControlEnabled, demoControlSkipReason);

    await expectDemoControlRejected(
      await adminApi.post("config", {
        data: {
          key: "e2e.demo.control.blocked",
          name: "E2E Demo Control Blocked",
          remark: "should be blocked by linapro-ops-demo-guard",
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
          remark: "should be blocked by linapro-ops-demo-guard",
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
    test.skip(!demoControlEnabled, demoControlSkipReason);

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
            probe: "linapro-ops-demo-guard-root-scope",
          },
        }),
        "POST / 应被全局作用域拦截",
      );
    } finally {
      await publicRequest.dispose();
    }
  });

  test("TC-105f: 演示模式拒绝插件治理写操作", async () => {
    test.skip(!demoControlEnabled, demoControlSkipReason);

    const originalLifecyclePlugin = await fetchPlugin(adminApi, lifecyclePluginID);
    expect(originalLifecyclePlugin, `未找到插件 ${lifecyclePluginID}`).toBeTruthy();
    if (!originalLifecyclePlugin) {
      throw new Error(`未找到插件 ${lifecyclePluginID}`);
    }

    const rejectedRequests = [
      {
        request: () => adminApi.post(`plugins/${lifecyclePluginID}/install`),
        message: `${lifecyclePluginID} 安装请求应被拦截`,
      },
      {
        request: () => adminApi.put(`plugins/${lifecyclePluginID}/enable`),
        message: `${lifecyclePluginID} 启用请求应被拦截`,
      },
      {
        request: () => adminApi.put(`plugins/${lifecyclePluginID}/disable`),
        message: `${lifecyclePluginID} 禁用请求应被拦截`,
      },
      {
        request: () => adminApi.delete(`plugins/${lifecyclePluginID}`),
        message: `${lifecyclePluginID} 卸载请求应被拦截`,
      },
      {
        request: () => adminApi.put(`plugins/${pluginID}/disable`),
        message: "linapro-ops-demo-guard 自身禁用请求应被拦截",
      },
    ];

    for (const item of rejectedRequests) {
      await expectDemoControlRejected(await item.request(), item.message);
    }
    await expectPluginState(
      adminApi,
      lifecyclePluginID,
      originalLifecyclePlugin.installed ?? 0,
      originalLifecyclePlugin.enabled ?? 0,
    );
    await expectPluginState(adminApi, pluginID, 1, 1);
  });
});
