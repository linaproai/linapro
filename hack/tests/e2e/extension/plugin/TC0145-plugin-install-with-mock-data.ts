import type { APIRequestContext, APIResponse } from "@playwright/test";

import { request as playwrightRequest } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { config } from "../../../fixtures/config";
import { PluginPage } from "../../../pages/PluginPage";
import {
  execPgSQL,
  pgIdentifier,
  queryPgRows,
  queryPgScalar,
} from "../../../support/postgres";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const targetPluginID = "content-notice";
const noticeTableName = "plugin_content_notice";
// Mock data file 001-content-notice-mock-data.sql ships these notice titles.
// The first one is asserted on screen to prove the mock load committed; the
// table-empty assertion in TC0146 checks the absence of all three together.
const mockNoticeTitle = "系统升级通知";

type PluginListItem = {
  id: string;
  enabled?: number;
  hasMockData?: number;
  installed?: number;
};

function unwrapApiData(payload: unknown) {
  if (payload && typeof payload === "object" && "data" in payload) {
    return (payload as { data: unknown }).data;
  }
  return payload;
}

function assertOk(response: APIResponse, message: string) {
  expect(response.ok(), `${message}, status=${response.status()}`).toBeTruthy();
}

async function createAdminApiContext(): Promise<APIRequestContext> {
  const loginApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await loginApi.post("auth/login", {
    data: {
      username: config.adminUser,
      password: config.adminPass,
    },
  });
  assertOk(loginResponse, "管理员登录 API 失败");

  const loginResult = unwrapApiData(await loginResponse.json()) as {
    accessToken?: string;
  };
  const accessToken = loginResult?.accessToken;
  expect(accessToken, "未获取到 accessToken").toBeTruthy();
  await loginApi.dispose();

  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

async function fetchPlugin(
  adminApi: APIRequestContext,
  pluginID: string,
): Promise<null | PluginListItem> {
  const response = await adminApi.get("plugins");
  assertOk(response, "查询插件列表失败");
  const payload = unwrapApiData(await response.json()) as {
    list?: PluginListItem[];
  };
  return (payload?.list ?? []).find((item) => item.id === pluginID) ?? null;
}

async function ensurePluginUninstalled(
  adminApi: APIRequestContext,
  pluginID: string,
) {
  const item = await fetchPlugin(adminApi, pluginID);
  if (!item || item.installed !== 1) {
    return;
  }
  // Disable first if needed; uninstall API tolerates already-disabled plugins.
  if (item.enabled === 1) {
    const disableResponse = await adminApi.put(`plugins/${pluginID}/disable`);
    assertOk(disableResponse, "卸载前禁用插件失败");
  }
  const uninstallResponse = await adminApi.delete(
    `plugins/${pluginID}?purgeStorageData=true`,
  );
  assertOk(uninstallResponse, "卸载插件失败");
}

function noticeTableExists(): boolean {
  return (
    queryPgScalar(
      `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '${noticeTableName}';`,
    ) === "1"
  );
}

function listNoticeTitlesFromDatabase(): string[] {
  if (!noticeTableExists()) {
    return [];
  }
  return queryPgRows(
    `SELECT title FROM ${pgIdentifier(noticeTableName)} ORDER BY id ASC;`,
  );
}

function dropNoticeTableIfExists() {
  execPgSQL(`DROP TABLE IF EXISTS ${pgIdentifier(noticeTableName)};`);
}

test.describe("TC-145 Install plugin with mock data opt-in", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    if (adminApi) {
      await ensurePluginUninstalled(adminApi, targetPluginID).catch(() => {});
      await adminApi.dispose();
    }
  });

  test.beforeEach(async () => {
    // Each subtest starts from a clean uninstalled state so the install dialog
    // and the post-install navigation reflect a fresh load every time.
    await ensurePluginUninstalled(adminApi, targetPluginID);
    dropNoticeTableIfExists();
  });

  test("TC-145a: install dialog exposes the mock-data checkbox for plugins shipping mock SQL", async ({
    adminPage,
  }) => {
    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(targetPluginID);

    const item = await fetchPlugin(adminApi, targetPluginID);
    expect(item, "插件应可被发现").toBeTruthy();
    expect(item?.hasMockData, "content-notice 应携带 mock-data 标识").toBe(1);
    await expect(pluginPage.tableColumn("示例数据")).toBeVisible();
    await expect(pluginPage.pluginColumnHelpIcon("mockData")).toBeVisible();
    await pluginPage.expectColumnHelpTooltip(
      "mockData",
      /示例数据|Mock Data/iu,
    );
    await expect(pluginPage.pluginMockDataValue(targetPluginID)).toContainText(
      "是",
    );

    await pluginPage.searchByPluginId("demo-control");
    await expect(pluginPage.pluginMockDataValue("demo-control")).toContainText(
      "否",
    );
    await pluginPage.searchByPluginId(targetPluginID);

    await pluginPage.openInstallAuthorization(targetPluginID);
    await expect(pluginPage.pluginInstallMockDataSection()).toBeVisible();
    await expect(pluginPage.pluginInstallMockDataCheckbox()).not.toBeChecked();
    await expect(pluginPage.pluginInstallMockDataHelpIcon()).toBeVisible();

    // Cancel without submitting so beforeEach can restore state for TC-145b.
    await adminPage.keyboard.press("Escape");
    await expect(pluginPage.hostServiceAuthDialog()).toHaveCount(0);
  });

  test("TC-145b: opting in loads the mock notices into the plugin page after install", async ({
    adminPage,
  }) => {
    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(targetPluginID);

    const installRequestPromise = adminPage.waitForRequest(
      (request) =>
        request.url().includes(`/plugins/${targetPluginID}/install`) &&
        request.method() === "POST",
    );

    await pluginPage.installPluginWithMockData(targetPluginID, true);

    const installRequest = await installRequestPromise;
    expect(installRequest.postDataJSON()).toMatchObject({
      installMockData: true,
    });

    // The host returns success without a mock-data warning toast; the install
    // success message is shared with the no-mock path.
    await expect(
      adminPage
        .locator(".ant-message-notice")
        .filter({ hasText: /插件安装成功|Plugin installed/iu })
        .last(),
    ).toBeVisible();

    const titles = listNoticeTitlesFromDatabase();
    expect(
      titles,
      "勾选示例数据后插件页面应包含 mock-data SQL 写入的通知",
    ).toContain(mockNoticeTitle);
  });
});
