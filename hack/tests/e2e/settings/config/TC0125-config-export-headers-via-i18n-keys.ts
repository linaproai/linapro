import type { APIRequestContext } from "@playwright/test";

import * as XLSX from "xlsx";

import { test, expect } from "../../../fixtures/auth";
import { createAdminApiContext } from "../../../support/api/job";

const xlsxRead = (XLSX as any).read || (XLSX as any).default?.read;
const xlsxUtils = (XLSX as any).utils || (XLSX as any).default?.utils;

test.describe("TC0125 参数设置导出表头通过 i18n key 本地化", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC-125a: 中文导出表头来自 config.field 翻译键", async () => {
    const headers = await exportConfigHeaders(adminApi, "zh-CN");

    expect(headers).toEqual([
      "参数名称",
      "参数键名",
      "参数键值",
      "备注",
      "创建时间",
      "修改时间",
    ]);
  });

  test("TC-125b: 英文导出表头随语言切换为英文", async () => {
    const headers = await exportConfigHeaders(adminApi, "en-US");

    expect(headers).toEqual([
      "Parameter Name",
      "Parameter Key",
      "Parameter Value",
      "Remark",
      "Created At",
      "Updated At",
    ]);
  });

  test("TC-125c: 繁体中文导出表头随语言切换为繁体中文", async () => {
    const headers = await exportConfigHeaders(adminApi, "zh-TW");

    expect(headers).toEqual([
      "參數名稱",
      "參數鍵名",
      "參數鍵值",
      "備註",
      "創建時間",
      "修改時間",
    ]);
  });
});

async function exportConfigHeaders(api: APIRequestContext, locale: string) {
  const response = await api.get("config/export?key=sys.jwt.expire", {
    headers: {
      "Accept-Language": locale,
    },
  });

  expect(response.ok()).toBeTruthy();
  expect(response.headers()["content-type"]).toContain("spreadsheetml");

  const workbook = xlsxRead(await response.body(), { type: "buffer" });
  const sheet = workbook.Sheets[workbook.SheetNames[0]];
  const rows = xlsxUtils.sheet_to_json(sheet, { header: 1 }) as string[][];
  return rows[0];
}
