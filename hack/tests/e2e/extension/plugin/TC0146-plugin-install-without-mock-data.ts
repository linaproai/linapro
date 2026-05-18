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
// When mock data is NOT loaded, none of these should appear on the plugin page.
const mockNoticeTitles = [
  "系统升级通知",
  "关于规范使用系统的公告",
  "新功能上线预告",
];

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

test.describe("TC-146 Install plugin without mock data leaves plugin tables empty", () => {
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
    await ensurePluginUninstalled(adminApi, targetPluginID);
    dropNoticeTableIfExists();
  });

  test("TC-146a: not opting in keeps the plugin's mock SQL out of the live table", async ({
    adminPage,
  }) => {
    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(targetPluginID);

    await pluginPage.installPluginWithMockData(targetPluginID, false);

    // Install message confirms the lifecycle ran without the mock-warning
    // toast — the helper used in TC-145 only fires when the user opted in.
    await expect(
      adminPage
        .locator(".ant-message-notice")
        .filter({ hasText: /插件安装成功|Plugin installed/iu })
        .last(),
    ).toBeVisible();

    const titles = new Set(listNoticeTitlesFromDatabase());
    for (const expectedAbsent of mockNoticeTitles) {
      expect(
        titles.has(expectedAbsent),
        `不勾选 mock-data 复选框时，mock 标题 "${expectedAbsent}" 不应出现在插件页面`,
      ).toBeFalsy();
    }
  });
});
