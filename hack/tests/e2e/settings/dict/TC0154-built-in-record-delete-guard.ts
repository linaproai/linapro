import type { APIRequestContext, APIResponse } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { ConfigPage } from "../../../pages/ConfigPage";
import { DictPage } from "../../../pages/DictPage";
import { createAdminApiContext, expectSuccess } from "../../../support/api/job";

type ListResult<T> = {
  list: T[];
  total: number;
};

type ErrorEnvelope = {
  code: number;
  errorCode?: string;
  messageKey?: string;
  message?: string;
};

type DictTypeItem = {
  id: number;
  isBuiltin: number;
  type: string;
};

type DictDataItem = {
  id: number;
  dictType: string;
  isBuiltin: number;
  value: string;
};

type ConfigItem = {
  id: number;
  isBuiltin: number;
  key: string;
};

type BuiltInRecords = {
  config: ConfigItem;
  dictData: DictDataItem;
  dictType: DictTypeItem;
};

async function expectBusinessCode(
  response: APIResponse,
  errorCode: string,
  messageKey: string,
) {
  const payload = (await response.json()) as ErrorEnvelope;
  expect(payload.code).not.toBe(0);
  expect(payload.errorCode).toBe(errorCode);
  expect(payload.messageKey).toBe(messageKey);
  return payload;
}

function expectBuiltIn<T extends { isBuiltin: number }>(
  item: T | undefined,
  name: string,
): T {
  expect(item, `${name} should exist`).toBeTruthy();
  expect(item!.isBuiltin, `${name} should be marked built-in`).toBe(1);
  return item!;
}

async function loadBuiltInRecords(
  api: APIRequestContext,
): Promise<BuiltInRecords> {
  const dictTypes = await expectSuccess<ListResult<DictTypeItem>>(
    await api.get("dict/type?pageNum=1&pageSize=20&type=sys_normal_disable"),
  );
  const dictType = expectBuiltIn(
    dictTypes.list.find((item) => item.type === "sys_normal_disable"),
    "sys_normal_disable dictionary type",
  );

  const dictData = await expectSuccess<ListResult<DictDataItem>>(
    await api.get(
      "dict/data?pageNum=1&pageSize=20&dictType=sys_normal_disable",
    ),
  );
  const normalData = expectBuiltIn(
    dictData.list.find((item) => item.value === "1"),
    "sys_normal_disable=1 dictionary data",
  );

  const configs = await expectSuccess<ListResult<ConfigItem>>(
    await api.get("config?pageNum=1&pageSize=20&key=sys.jwt.expire"),
  );
  const config = expectBuiltIn(
    configs.list.find((item) => item.key === "sys.jwt.expire"),
    "sys.jwt.expire system parameter",
  );

  return {
    config,
    dictData: normalData,
    dictType,
  };
}

test.describe("TC-154 Built-in record delete guard", () => {
  let api: APIRequestContext;
  let records: BuiltInRecords;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    records = await loadBuiltInRecords(api);
  });

  test.afterAll(async () => {
    await api.dispose();
  });

  test("TC-154a: built-in dictionary type delete action is disabled with tooltip while edit stays enabled", async ({
    adminPage,
  }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();
    await dictPage.fillTypeSearchField("字典类型", records.dictType.type);
    await dictPage.clickTypeSearch();

    await expect(
      await dictPage.getTypeDeleteButton(records.dictType.type),
    ).toBeDisabled();
    await expect(
      await dictPage.getTypeEditButton(records.dictType.type),
    ).toBeEnabled();

    await dictPage.hoverTypeDeleteAction(records.dictType.type);
    expect(await dictPage.isBuiltinDeleteTooltipVisible()).toBeTruthy();
  });

  test("TC-154b: built-in dictionary data delete action is disabled with tooltip while edit stays enabled", async ({
    adminPage,
  }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();
    await dictPage.clickTypeRow(records.dictType.type);

    await expect(
      await dictPage.getDataDeleteButton(records.dictData.value),
    ).toBeDisabled();
    await expect(
      await dictPage.getDataEditButton(records.dictData.value),
    ).toBeEnabled();

    await dictPage.hoverDataDeleteAction(records.dictData.value);
    expect(await dictPage.isBuiltinDeleteTooltipVisible()).toBeTruthy();
  });

  test("TC-154c: built-in system parameter delete action is disabled with tooltip while edit stays enabled", async ({
    adminPage,
  }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();
    await configPage.fillSearchField("参数键名", records.config.key);
    await configPage.clickSearch();

    await expect(
      await configPage.getDeleteButtonById(records.config.id),
    ).toBeDisabled();
    await expect(
      await configPage.getEditButtonById(records.config.id),
    ).toBeEnabled();

    await configPage.hoverDeleteActionByKey(records.config.key);
    expect(await configPage.isBuiltinDeleteTooltipVisible()).toBeTruthy();
  });

  test("TC-154d: backend rejects direct deletion of built-in records and preserves them", async () => {
    await expectBusinessCode(
      await api.delete(`dict/type/${records.dictType.id}`),
      "DICT_TYPE_BUILTIN_DELETE_DENIED",
      "error.dict.type.builtin.delete.denied",
    );

    await expectBusinessCode(
      await api.delete(`dict/data/${records.dictData.id}`),
      "DICT_DATA_BUILTIN_DELETE_DENIED",
      "error.dict.data.builtin.delete.denied",
    );

    await expectBusinessCode(
      await api.delete(`config/${records.config.id}`),
      "SYSCONFIG_BUILTIN_DELETE_DENIED",
      "error.sysconfig.builtin.delete.denied",
    );

    records = await loadBuiltInRecords(api);
  });
});
