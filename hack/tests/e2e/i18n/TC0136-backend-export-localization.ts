import type { APIRequestContext, APIResponse } from "@playwright/test";

import * as XLSX from "xlsx";

import { test, expect } from "../../fixtures/auth";
import {
  createAdminApiContext,
  enablePlugin,
  getPlugin,
  installPlugin,
  syncPlugins,
} from "../../support/api/job";

const xlsxRead = (XLSX as any).read || (XLSX as any).default?.read;
const xlsxUtils = (XLSX as any).utils || (XLSX as any).default?.utils;

const sourcePluginIDs = ["monitor-loginlog", "monitor-operlog"] as const;

async function ensureSourcePluginsEnabled(
  api: APIRequestContext,
  pluginIDs: readonly string[],
) {
  await syncPlugins(api);
  for (const pluginID of pluginIDs) {
    let plugin = await getPlugin(api, pluginID);
    if (plugin.installed !== 1) {
      await installPlugin(api, pluginID);
      plugin = await getPlugin(api, pluginID);
    }
    if (plugin.enabled !== 1) {
      await enablePlugin(api, pluginID);
    }
  }
}

async function readWorkbookRows(response: APIResponse): Promise<string[][]> {
  expect(response.ok()).toBeTruthy();
  const workbook = xlsxRead(await response.body(), { type: "buffer" });
  const sheet = workbook.Sheets[workbook.SheetNames[0]];
  return xlsxUtils.sheet_to_json(sheet, { header: 1 }) as string[][];
}

function flattenRows(rows: string[][]) {
  return rows.flat().filter((value): value is string => typeof value === "string");
}

test.describe("TC-136 Backend export localization", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await ensureSourcePluginsEnabled(adminApi, sourcePluginIDs);

    const loginApi = await createAdminApiContext();
    await loginApi.dispose();

    const operlogSeedResponse = await adminApi.get("dict/type/export?pageNum=1&pageSize=1", {
      headers: { "Accept-Language": "en-US" },
    });
    expect(operlogSeedResponse.ok()).toBeTruthy();
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC-136a: login-log export headers and status values follow request locale", async () => {
    const enRows = await readWorkbookRows(
      await adminApi.get("loginlog/export", {
        headers: { "Accept-Language": "en-US" },
      }),
    );
    expect(enRows[0]).toEqual([
      "User Account",
      "Login Status",
      "IP Address",
      "Browser",
      "Operating System",
      "Message",
      "Login Time",
    ]);
    expect(flattenRows(enRows)).toContain("Success");
    expect(flattenRows(enRows)).not.toContain("成功");

    const zhRows = await readWorkbookRows(
      await adminApi.get("loginlog/export", {
        headers: { "Accept-Language": "zh-CN" },
      }),
    );
    expect(zhRows[0]).toEqual([
      "用户账号",
      "登录状态",
      "IP 地址",
      "浏览器",
      "操作系统",
      "提示信息",
      "登录时间",
    ]);
    expect(flattenRows(zhRows)).toContain("成功");
  });

  test("TC-136b: operation-log export headers follow request locale", async () => {
    const enRows = await readWorkbookRows(
      await adminApi.get("operlog/export", {
        headers: { "Accept-Language": "en-US" },
      }),
    );
    expect(enRows[0]).toEqual([
      "Module Name",
      "Operation Summary",
      "Operation Type",
      "Operator",
      "Request Method",
      "Request URL",
      "IP Address",
      "Request Parameters",
      "Response Result",
      "Operation Result",
      "Error Information",
      "Duration (ms)",
      "Operation Time",
    ]);

    const zhRows = await readWorkbookRows(
      await adminApi.get("operlog/export", {
        headers: { "Accept-Language": "zh-CN" },
      }),
    );
    expect(zhRows[0]).toEqual([
      "模块名称",
      "操作摘要",
      "操作类型",
      "操作人员",
      "请求方式",
      "请求地址",
      "IP 地址",
      "请求参数",
      "响应结果",
      "操作结果",
      "错误信息",
      "耗时(ms)",
      "操作时间",
    ]);
  });
});
