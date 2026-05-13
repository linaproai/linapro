import { expect, test } from "@host-tests/fixtures/auth";
import { config } from "@host-tests/fixtures/config";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";
import { LoginPage } from "@host-tests/pages/LoginPage";
import {
  createAdminApiContext,
  expectBusinessError,
  expectSuccess,
} from "@host-tests/support/api/job";
import { waitForRouteReady } from "@host-tests/support/ui";

type AdminApiContext = Awaited<ReturnType<typeof createAdminApiContext>>;

type CreatedId = {
  id: number;
};

type ListResult<T> = {
  list: T[];
  total: number;
};

type ResolveResult = {
  matched: boolean;
  source: string;
  strategyId: number;
};

type StrategyDetail = {
  enable: number;
  global: number;
  id: number;
  name: string;
  strategy: string;
};

type DeviceBindingItem = {
  deviceId: string;
  strategyId: number;
};

type TenantBindingItem = {
  tenantId?: string;
  strategyId: number;
};

type TenantDeviceBindingItem = {
  tenantId: string;
  deviceId?: string;
  strategyId: number;
};

type AliasDetail = {
  id: number;
  alias: string;
  autoRemove: number;
  streamPath: string;
};

async function expectPageHeightStable(page: any, pageName: string) {
  const samples = await page.evaluate(async () => {
    const values: number[] = [];
    for (let index = 0; index < 5; index += 1) {
      values.push(document.documentElement.scrollHeight);
      if (index < 4) {
        await new Promise<void>((resolve) => {
          requestAnimationFrame(() => requestAnimationFrame(() => resolve()));
        });
      }
    }
    return values;
  });

  expect(
    Math.max(...samples) - Math.min(...samples),
    `${pageName}高度未稳定，采样结果: ${samples.join(", ")}`,
  ).toBeLessThanOrEqual(16);
}

async function expectNoPageErrors(
  errors: Error[],
  allowedMessagePattern?: RegExp,
) {
  const unexpectedErrors = errors.filter(
    (error) => !allowedMessagePattern?.test(error.message),
  );
  expect(
    unexpectedErrors.map((error) => error.message),
    "媒体管理页面不应触发未捕获前端异常",
  ).toEqual([]);
}

async function expectApiResponseSuccess(response: any) {
  expect(response.ok()).toBeTruthy();
  const payload = await response.json();
  expect(payload.code).toBe(0);
  return payload.data;
}

function visibleModalRoot(page: any) {
  return page.locator("body");
}

async function confirmModal(modal: any) {
  await modal
    .getByRole("button", { name: /确\s*(定|认)|OK/i })
    .last()
    .click();
}

async function confirmPopconfirm(page: any) {
  await page
    .locator(".ant-popover")
    .getByRole("button", { name: /确\s*(定|认)|OK/i })
    .last()
    .click();
}

function tableRowByText(page: any, text: string) {
  return page.locator(".vxe-body--row").filter({ hasText: text }).first();
}

async function expectCheckedRadioLabel(
  root: any,
  expectedLabel: string,
) {
  await expect(
    root.locator(".ant-radio-button-wrapper-checked"),
  ).toContainText(expectedLabel);
}

function rowKeyDevice(deviceId: string) {
  return `device:${deviceId}`;
}

function rowKeyTenant(tenantId: string) {
  return `tenant:${tenantId}`;
}

function rowKeyTenantDevice(tenantId: string, deviceId: string) {
  return `tenantDevice:${tenantId}:${deviceId}`;
}

async function createStrategy(
  api: AdminApiContext,
  name: string,
  body: string,
) {
  const result = await expectSuccess<CreatedId>(
    await api.post("media/strategies", {
      data: {
        enable: 1,
        global: 2,
        name,
        strategy: body,
      },
    }),
  );
  return result.id;
}

function pathSegment(value: string) {
  return encodeURIComponent(value);
}

function hostStaticBaseURL() {
  const configuredBaseURL = process.env.E2E_HOST_BASE_URL?.trim();
  if (configuredBaseURL) {
    return configuredBaseURL.replace(/\/$/, "");
  }

  const baseURL = new URL(config.baseURL);
  baseURL.port = process.env.E2E_HOST_PORT?.trim() || "8080";
  return baseURL.toString().replace(/\/$/, "");
}

async function saveDeviceBinding(
  api: AdminApiContext,
  deviceId: string,
  strategyId: number,
) {
  await expectSuccess(
    await api.put(`media/device-bindings/${pathSegment(deviceId)}`, {
      data: {
        deviceId,
        strategyId,
      },
    }),
  );
}

async function saveTenantBinding(
  api: AdminApiContext,
  tenantId: string,
  strategyId: number,
) {
  await expectSuccess(
    await api.put(`media/tenant-bindings/${pathSegment(tenantId)}`, {
      data: {
        tenantId,
        strategyId,
      },
    }),
  );
}

async function saveTenantDeviceBinding(
  api: AdminApiContext,
  tenantId: string,
  deviceId: string,
  strategyId: number,
) {
  await expectSuccess(
    await api.put(
      `media/tenant-device-bindings/${pathSegment(tenantId)}/${pathSegment(
        deviceId,
      )}`,
      {
        data: {
          tenantId,
          deviceId,
          strategyId,
        },
      },
    ),
  );
}

async function deleteDeviceBinding(api: AdminApiContext, deviceId: string) {
  await expectSuccess(
    await api.delete(`media/device-bindings/${pathSegment(deviceId)}`),
  );
}

async function deleteTenantBinding(
  api: AdminApiContext,
  tenantId: string,
) {
  await expectSuccess(
    await api.delete(`media/tenant-bindings/${pathSegment(tenantId)}`),
  );
}

async function deleteTenantDeviceBinding(
  api: AdminApiContext,
  tenantId: string,
  deviceId: string,
) {
  await expectSuccess(
    await api.delete(
      `media/tenant-device-bindings/${pathSegment(tenantId)}/${pathSegment(
        deviceId,
      )}`,
    ),
  );
}

async function resolveStrategy(
  api: AdminApiContext,
  data: {
    tenantId?: string;
    deviceId?: string;
  },
) {
  const params = new URLSearchParams();
  if (data.tenantId) {
    params.set("tenantId", data.tenantId);
  }
  if (data.deviceId) {
    params.set("deviceId", data.deviceId);
  }
  return expectSuccess<ResolveResult>(
    await api.get(`media/strategies/resolve?${params.toString()}`),
  );
}

test.describe("TC-234 media plugin owned E2E discovery", () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, "media");
  });

  test("TC-234a: 媒体管理页面加载、切换页签且高度稳定", async ({
    adminPage,
  }) => {
    const pageErrors: Error[] = [];
    adminPage.on("pageerror", (error) => pageErrors.push(error));

    const strategyResponse = adminPage.waitForResponse(
      (res) =>
        res.url().includes("/api/v1/media/strategies") &&
        res.request().method() === "GET" &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.goto("/media");
    await strategyResponse;
    await waitForRouteReady(adminPage);

    await expect(adminPage.getByTestId("media-management-page")).toBeVisible();
    await expect(adminPage.getByText("媒体策略").first()).toBeVisible();
    await expect(adminPage.locator(".vxe-table").first()).toBeVisible();
    await expectPageHeightStable(adminPage, "媒体策略页签");

    const deviceBindingResponse = adminPage.waitForResponse(
      (res) =>
        res.url().includes("/api/v1/media/device-bindings") &&
        res.request().method() === "GET" &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.getByRole("tab", { exact: true, name: "设备绑定" }).click();
    await deviceBindingResponse;
    await waitForRouteReady(adminPage);
    await expect(adminPage.getByText("设备策略绑定").first()).toBeVisible();
    await expectPageHeightStable(adminPage, "设备绑定页签");

    const tenantBindingResponse = adminPage.waitForResponse(
      (res) =>
        res.url().includes("/api/v1/media/tenant-bindings") &&
        res.request().method() === "GET" &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.getByRole("tab", { exact: true, name: "租户绑定" }).click();
    await tenantBindingResponse;
    await waitForRouteReady(adminPage);
    await expect(adminPage.getByText("租户策略绑定").first()).toBeVisible();
    await expectPageHeightStable(adminPage, "租户绑定页签");

    const tenantDeviceBindingResponse = adminPage.waitForResponse(
      (res) =>
        res.url().includes("/api/v1/media/tenant-device-bindings") &&
        res.request().method() === "GET" &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage
      .getByRole("tab", { exact: true, name: "租户设备绑定" })
      .click();
    await tenantDeviceBindingResponse;
    await waitForRouteReady(adminPage);
    await expect(adminPage.getByText("租户设备策略绑定").first()).toBeVisible();
    await expectPageHeightStable(adminPage, "租户设备绑定页签");

    await adminPage.getByRole("tab", { exact: true, name: "策略解析" }).click();
    await waitForRouteReady(adminPage);
    await expect(adminPage.getByText("解析生效策略").first()).toBeVisible();
    await expectPageHeightStable(adminPage, "策略解析页签");

    const aliasResponse = adminPage.waitForResponse(
      (res) =>
        res.url().includes("/api/v1/media/stream-aliases") &&
        res.request().method() === "GET" &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.getByRole("tab", { exact: true, name: "流别名" }).click();
    await aliasResponse;
    await waitForRouteReady(adminPage);
    await expect(adminPage.getByText("流别名").first()).toBeVisible();
    await expectPageHeightStable(adminPage, "流别名页签");

    await expectNoPageErrors(pageErrors, /ResizeObserver loop/i);
  });

  test("TC-234b: 媒体策略绑定优先级和流别名接口可用", async () => {
    const api = await createAdminApiContext();
    const suffix = Date.now().toString();
    const tenantId = `tenant-e2e-${suffix}`;
    const deviceId = `3402000000132${suffix.slice(-7).padStart(7, "0")}`;
    const alias = `e2e-alias-${suffix}`;

    const strategyIds: number[] = [];
    let aliasId = 0;

    try {
      const tenantStrategyId = await createStrategy(
        api,
        `E2E租户策略-${suffix}`,
        `record: tenant-${suffix}`,
      );
      const deviceStrategyId = await createStrategy(
        api,
        `E2E设备策略-${suffix}`,
        `record: device-${suffix}`,
      );
      const tenantDeviceStrategyId = await createStrategy(
        api,
        `E2E租户设备策略-${suffix}`,
        `record: tenant-device-${suffix}`,
      );
      strategyIds.push(
        tenantDeviceStrategyId,
        deviceStrategyId,
        tenantStrategyId,
      );

      await saveTenantBinding(api, tenantId, tenantStrategyId);
      await saveDeviceBinding(api, deviceId, deviceStrategyId);
      await saveTenantDeviceBinding(
        api,
        tenantId,
        deviceId,
        tenantDeviceStrategyId,
      );

      await expectBusinessError(
        await api.delete(`media/strategies/${tenantStrategyId}`),
      );

      await expect(resolveStrategy(api, { tenantId, deviceId })).resolves.toMatchObject({
        matched: true,
        source: "tenantDevice",
        strategyId: tenantDeviceStrategyId,
      });

      await deleteTenantDeviceBinding(api, tenantId, deviceId);
      await expect(resolveStrategy(api, { tenantId, deviceId })).resolves.toMatchObject({
        matched: true,
        source: "device",
        strategyId: deviceStrategyId,
      });

      await deleteDeviceBinding(api, deviceId);
      await expect(resolveStrategy(api, { tenantId, deviceId })).resolves.toMatchObject({
        matched: true,
        source: "tenant",
        strategyId: tenantStrategyId,
      });

      const createdAlias = await expectSuccess<CreatedId>(
        await api.post("media/stream-aliases", {
          data: {
            alias,
            autoRemove: 0,
            streamPath: `live/${alias}`,
          },
        }),
      );
      aliasId = createdAlias.id;

      await expectSuccess(
        await api.put(`media/stream-aliases/${aliasId}`, {
          data: {
            alias,
            autoRemove: 1,
            streamPath: `live/${alias}-updated`,
          },
        }),
      );

      const aliasDetail = await expectSuccess<AliasDetail>(
        await api.get(`media/stream-aliases/${aliasId}`),
      );
      expect(aliasDetail).toMatchObject({
        alias,
        autoRemove: 1,
        streamPath: `live/${alias}-updated`,
      });

      await expectSuccess(await api.delete(`media/stream-aliases/${aliasId}`));
      aliasId = 0;
    } finally {
      await deleteTenantDeviceBinding(api, tenantId, deviceId).catch(
        () => undefined,
      );
      await deleteDeviceBinding(api, deviceId).catch(() => undefined);
      await deleteTenantBinding(api, tenantId).catch(() => undefined);
      if (aliasId > 0) {
        await api
          .delete(`media/stream-aliases/${aliasId}`)
          .catch(() => undefined);
      }
      for (const strategyId of strategyIds) {
        await api
          .delete(`media/strategies/${strategyId}`)
          .catch(() => undefined);
      }
      await api.dispose();
    }
  });

  test("TC-234c: 媒体管理全部 REST 接口语义正确", async () => {
    const api = await createAdminApiContext();
    const suffix = Date.now().toString();
    const strategyName = `E2E接口策略-${suffix}`;
    const updatedStrategyName = `E2E接口策略更新-${suffix}`;
    const strategyBody = `record: api-${suffix}`;
    const updatedStrategyBody = `record: api-updated-${suffix}`;
    const deviceId = `3402000000139${suffix.slice(-7).padStart(7, "0")}`;
    const tenantId = `tenant-api-${suffix}`;
    const alias = `e2e-api-alias-${suffix}`;

    let strategyId = 0;
    let replacementStrategyId = 0;
    let aliasId = 0;

    try {
      strategyId = await createStrategy(api, strategyName, strategyBody);
      const strategyDetail = await expectSuccess<StrategyDetail>(
        await api.get(`media/strategies/${strategyId}`),
      );
      expect(strategyDetail).toMatchObject({
        enable: 1,
        global: 2,
        id: strategyId,
        name: strategyName,
        strategy: strategyBody,
      });

      const listedStrategies = await expectSuccess<ListResult<StrategyDetail>>(
        await api.get(
          `media/strategies?pageNum=1&pageSize=20&keyword=${encodeURIComponent(strategyName)}`,
        ),
      );
      expect(
        listedStrategies.list.some((item) => item.id === strategyId),
      ).toBeTruthy();

      await expectSuccess(
        await api.put(`media/strategies/${strategyId}`, {
          data: {
            enable: 1,
            global: 2,
            name: updatedStrategyName,
            strategy: updatedStrategyBody,
          },
        }),
      );
      await expect(
        expectSuccess<StrategyDetail>(
          await api.get(`media/strategies/${strategyId}`),
        ),
      ).resolves.toMatchObject({
        name: updatedStrategyName,
        strategy: updatedStrategyBody,
      });

      await expectSuccess(
        await api.put(`media/strategies/${strategyId}/enable`, {
          data: { enable: 2 },
        }),
      );
      await expect(
        expectSuccess<StrategyDetail>(
          await api.get(`media/strategies/${strategyId}`),
        ),
      ).resolves.toMatchObject({ enable: 2 });

      await expectSuccess(await api.put(`media/strategies/${strategyId}/global`));
      await expect(
        expectSuccess<StrategyDetail>(
          await api.get(`media/strategies/${strategyId}`),
        ),
      ).resolves.toMatchObject({ enable: 1, global: 1 });

      replacementStrategyId = await createStrategy(
        api,
        `E2E接口替换策略-${suffix}`,
        `record: replacement-${suffix}`,
      );

      await saveDeviceBinding(api, deviceId, strategyId);
      await saveDeviceBinding(api, deviceId, replacementStrategyId);
      const deviceBindings = await expectSuccess<
        ListResult<DeviceBindingItem>
      >(
        await api.get(
          `media/device-bindings?pageNum=1&pageSize=20&keyword=${encodeURIComponent(deviceId)}`,
        ),
      );
      expect(deviceBindings.list).toEqual([
        expect.objectContaining({
          deviceId,
          strategyId: replacementStrategyId,
        }),
      ]);

      await saveTenantBinding(api, tenantId, strategyId);
      const tenantBindings = await expectSuccess<
        ListResult<TenantBindingItem>
      >(
        await api.get(
          `media/tenant-bindings?pageNum=1&pageSize=20&keyword=${encodeURIComponent(tenantId)}`,
        ),
      );
      expect(tenantBindings.list).toEqual([
        expect.objectContaining({
          tenantId,
          strategyId,
        }),
      ]);

      await saveTenantDeviceBinding(api, tenantId, deviceId, strategyId);
      const tenantDeviceBindings = await expectSuccess<
        ListResult<TenantDeviceBindingItem>
      >(
        await api.get(
          `media/tenant-device-bindings?pageNum=1&pageSize=20&keyword=${encodeURIComponent(tenantId)}`,
        ),
      );
      expect(tenantDeviceBindings.list).toEqual([
        expect.objectContaining({
          tenantId,
          deviceId,
          strategyId,
        }),
      ]);

      await expect(resolveStrategy(api, { tenantId, deviceId })).resolves.toMatchObject({
        matched: true,
        source: "tenantDevice",
        strategyId,
      });
      await deleteTenantDeviceBinding(api, tenantId, deviceId);
      await deleteTenantBinding(api, tenantId);
      await deleteDeviceBinding(api, deviceId);

      await expect(resolveStrategy(api, { tenantId, deviceId })).resolves.toMatchObject({
        matched: true,
        source: "global",
        strategyId,
      });

      const createdAlias = await expectSuccess<CreatedId>(
        await api.post("media/stream-aliases", {
          data: {
            alias,
            autoRemove: 0,
            streamPath: `live/${alias}`,
          },
        }),
      );
      aliasId = createdAlias.id;

      const listedAliases = await expectSuccess<ListResult<AliasDetail>>(
        await api.get(
          `media/stream-aliases?pageNum=1&pageSize=20&keyword=${encodeURIComponent(alias)}`,
        ),
      );
      expect(listedAliases.list).toEqual([
        expect.objectContaining({ alias, id: aliasId }),
      ]);

      await expect(
        expectSuccess<AliasDetail>(
          await api.get(`media/stream-aliases/${aliasId}`),
        ),
      ).resolves.toMatchObject({
        alias,
        autoRemove: 0,
        streamPath: `live/${alias}`,
      });

      await expectSuccess(
        await api.put(`media/stream-aliases/${aliasId}`, {
          data: {
            alias,
            autoRemove: 1,
            streamPath: `live/${alias}-updated`,
          },
        }),
      );
      await expect(
        expectSuccess<AliasDetail>(
          await api.get(`media/stream-aliases/${aliasId}`),
        ),
      ).resolves.toMatchObject({
        autoRemove: 1,
        streamPath: `live/${alias}-updated`,
      });

      await expectSuccess(await api.delete(`media/stream-aliases/${aliasId}`));
      aliasId = 0;
      await expectBusinessError(await api.get(`media/stream-aliases/${createdAlias.id}`));

      await expectSuccess(await api.delete(`media/strategies/${replacementStrategyId}`));
      replacementStrategyId = 0;
      await expectSuccess(await api.delete(`media/strategies/${strategyId}`));
      strategyId = 0;
    } finally {
      await deleteTenantDeviceBinding(api, tenantId, deviceId).catch(
        () => undefined,
      );
      await deleteTenantBinding(api, tenantId).catch(() => undefined);
      await deleteDeviceBinding(api, deviceId).catch(() => undefined);
      if (aliasId > 0) {
        await api
          .delete(`media/stream-aliases/${aliasId}`)
          .catch(() => undefined);
      }
      for (const id of [replacementStrategyId, strategyId]) {
        if (id > 0) {
          await api.delete(`media/strategies/${id}`).catch(() => undefined);
        }
      }
      await api.dispose();
    }
  });

  test("TC-234d: 宿主静态入口可加载媒体管理页面", async ({ browser }) => {
    const context = await browser.newContext({
      baseURL: hostStaticBaseURL(),
      locale: "zh-CN",
    });
    const page = await context.newPage();
    const pageErrors: Error[] = [];
    page.on("pageerror", (error) => pageErrors.push(error));

    try {
      const loginPage = new LoginPage(page);
      await page.goto("/#/auth/login");
      await loginPage.usernameInput.waitFor({ state: "visible" });
      await loginPage.login(config.adminUser, config.adminPass);
      await page.waitForURL((url) => !url.hash.includes("/auth/login"), {
        timeout: 15000,
      });
      await waitForRouteReady(page, 15000);

      await page.evaluate(() => {
        window.location.hash = "#/media";
      });
      await page.waitForURL((url) => url.hash === "#/media", {
        timeout: 15000,
      });
      await waitForRouteReady(page, 15000);

      await expect(page.getByTestId("media-management-page")).toBeVisible();
      await expect(page.getByText("插件页面未找到")).toHaveCount(0);
      await expect(page.getByText("媒体策略").first()).toBeVisible();
      await expectNoPageErrors(pageErrors, /ResizeObserver loop/i);
    } finally {
      await context.close();
    }
  });

  test("TC-234e: 媒体管理界面编辑回显和接口执行正确", async ({
    adminPage,
  }) => {
    const api = await createAdminApiContext();
    const suffix = Date.now().toString();
    const strategyName = `E2E界面策略-${suffix}`;
    const updatedStrategyName = `E2E界面策略更新-${suffix}`;
    const strategyBody = `record: ui-${suffix}`;
    const updatedStrategyBody = `record: ui-updated-${suffix}`;
    const replacementStrategyName = `E2E界面备用策略-${suffix}`;
    const deviceId = `000-e2e-device-${suffix}`;
    const tenantId = `000-e2e-tenant-${suffix}`;
    const tenantDeviceId = `000-e2e-tenant-device-${suffix}`;
    const alias = `e2e-ui-alias-${suffix}`;
    const createdAfterEditStrategyName = `E2E界面新增策略-${suffix}`;
    const createdAfterEditStrategyBody = `record: ui-created-after-edit-${suffix}`;
    const createdAfterEditAlias = `e2e-ui-alias-new-${suffix}`;
    const createdAfterEditDeviceId = `000-e2e-device-new-${suffix}`;
    const createdAfterEditTenantId = `000-e2e-tenant-new-${suffix}`;
    const createdAfterEditTenantDeviceTenantId = `000-e2e-td-tenant-new-${suffix}`;
    const createdAfterEditTenantDeviceId = `000-e2e-td-device-new-${suffix}`;

    let strategyId = 0;
    let createdAfterEditStrategyId = 0;
    let replacementStrategyId = 0;
    let aliasId = 0;
    let createdAfterEditAliasId = 0;
    let previousGlobalStrategyId = 0;

    try {
      strategyId = await createStrategy(api, strategyName, strategyBody);
      const previousGlobalStrategies = await expectSuccess<
        ListResult<StrategyDetail>
      >(await api.get("media/strategies?pageNum=1&pageSize=1&global=1"));
      previousGlobalStrategyId = previousGlobalStrategies.list[0]?.id || 0;
      replacementStrategyId = await createStrategy(
        api,
        replacementStrategyName,
        `record: ui-replacement-${suffix}`,
      );
      await saveDeviceBinding(api, deviceId, strategyId);
      await saveTenantBinding(api, tenantId, strategyId);
      await saveTenantDeviceBinding(
        api,
        tenantId,
        tenantDeviceId,
        strategyId,
      );
      const createdAlias = await expectSuccess<CreatedId>(
        await api.post("media/stream-aliases", {
          data: {
            alias,
            autoRemove: 0,
            streamPath: `live/${alias}`,
          },
        }),
      );
      aliasId = createdAlias.id;

      const strategyListResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/strategies") &&
          res.request().method() === "GET" &&
          res.status() === 200,
        { timeout: 15000 },
      );
      await adminPage.goto("/media");
      await strategyListResponse;
      await waitForRouteReady(adminPage);

      const strategyRow = tableRowByText(adminPage, strategyName);
      await expect(strategyRow).toBeVisible();
      const strategyDetailResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes(`/api/v1/media/strategies/${strategyId}`) &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage.getByTestId(`media-strategy-edit-${strategyId}`).click();
      await expectApiResponseSuccess(await strategyDetailResponse);
      const strategyModal = visibleModalRoot(adminPage);
      await expect(strategyModal.getByText("编辑媒体策略")).toBeVisible();
      await expect(
        strategyModal.getByTestId("media-strategy-name"),
      ).toHaveValue(strategyName);
      await expect(
        strategyModal.getByTestId("media-strategy-body"),
      ).toHaveValue(strategyBody);
      await expectCheckedRadioLabel(
        strategyModal.getByTestId("media-strategy-enable"),
        "开启",
      );
      await expectCheckedRadioLabel(
        strategyModal.getByTestId("media-strategy-global"),
        "否",
      );
      await strategyModal
        .getByTestId("media-strategy-name")
        .fill(updatedStrategyName);
      await strategyModal
        .getByTestId("media-strategy-body")
        .fill(updatedStrategyBody);
      const strategyUpdateResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes(`/api/v1/media/strategies/${strategyId}`) &&
          res.request().method() === "PUT",
        { timeout: 15000 },
      );
      await confirmModal(strategyModal);
      await expectApiResponseSuccess(await strategyUpdateResponse);
      await expect(
        adminPage.getByText("编辑媒体策略", { exact: true }),
      ).toBeHidden({ timeout: 15000 });
      await expect(
        expectSuccess<StrategyDetail>(
          await api.get(`media/strategies/${strategyId}`),
        ),
      ).resolves.toMatchObject({
        name: updatedStrategyName,
        strategy: updatedStrategyBody,
      });

      await adminPage.getByTestId("media-strategy-add").click();
      await expect(strategyModal.getByText("新增媒体策略")).toBeVisible();
      await expect(
        strategyModal.getByTestId("media-strategy-name"),
      ).toHaveValue("");
      await expect(
        strategyModal.getByTestId("media-strategy-body"),
      ).toHaveValue("");
      await strategyModal
        .getByTestId("media-strategy-name")
        .fill(createdAfterEditStrategyName);
      await strategyModal
        .getByTestId("media-strategy-body")
        .fill(createdAfterEditStrategyBody);
      const strategyCreateResponse = adminPage.waitForResponse(
        (res) =>
          res.url().endsWith("/api/v1/media/strategies") &&
          res.request().method() === "POST",
        { timeout: 15000 },
      );
      await confirmModal(strategyModal);
      const createdStrategyPayload = await expectApiResponseSuccess(
        await strategyCreateResponse,
      );
      createdAfterEditStrategyId = createdStrategyPayload.id;
      expect(createdAfterEditStrategyId).toBeGreaterThan(0);
      await expect(
        adminPage.getByText("新增媒体策略", { exact: true }),
      ).toBeHidden({ timeout: 15000 });
      await expect(
        expectSuccess<StrategyDetail>(
          await api.get(`media/strategies/${createdAfterEditStrategyId}`),
        ),
      ).resolves.toMatchObject({
        name: createdAfterEditStrategyName,
        strategy: createdAfterEditStrategyBody,
      });

      const strategyToggleResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/strategies/${createdAfterEditStrategyId}/enable`,
            ) && res.request().method() === "PUT",
        { timeout: 15000 },
      );
      await adminPage
        .getByTestId(`media-strategy-toggle-${createdAfterEditStrategyId}`)
        .click();
      await expectApiResponseSuccess(await strategyToggleResponse);
      await expect(
        expectSuccess<StrategyDetail>(
          await api.get(`media/strategies/${createdAfterEditStrategyId}`),
        ),
      ).resolves.toMatchObject({ enable: 2 });

      const strategyGlobalResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/strategies/${createdAfterEditStrategyId}/global`,
            ) && res.request().method() === "PUT",
        { timeout: 15000 },
      );
      await adminPage
        .getByTestId(`media-strategy-global-${createdAfterEditStrategyId}`)
        .click();
      await expectApiResponseSuccess(await strategyGlobalResponse);
      await expect(
        expectSuccess<StrategyDetail>(
          await api.get(`media/strategies/${createdAfterEditStrategyId}`),
        ),
      ).resolves.toMatchObject({ enable: 1, global: 1 });

      const deviceBindingListResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/device-bindings") &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage
        .getByRole("tab", { exact: true, name: "设备绑定" })
        .click();
      await expectApiResponseSuccess(await deviceBindingListResponse);
      const deviceRow = tableRowByText(adminPage, deviceId);
      await expect(deviceRow).toBeVisible();
      const deviceStrategyOptionsResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/strategies") &&
          res.url().includes("enable=1") &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage
        .getByTestId(`media-device-binding-edit-${rowKeyDevice(deviceId)}`)
        .click();
      await expectApiResponseSuccess(await deviceStrategyOptionsResponse);
      const deviceModal = visibleModalRoot(adminPage);
      await expect(deviceModal.getByText("编辑设备策略绑定")).toBeVisible();
      await expect(
        deviceModal.getByTestId("media-binding-device-id"),
      ).toHaveValue(deviceId);
      await expect(
        deviceModal.getByTestId("media-binding-device-id"),
      ).toBeDisabled();
      await expect(
        deviceModal
          .getByTestId("media-binding-strategy")
          .locator(".ant-select-selection-item"),
      ).toContainText(`#${strategyId}`);
      const deviceUpdateResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(`/api/v1/media/device-bindings/${pathSegment(deviceId)}`) &&
          res.request().method() === "PUT",
        { timeout: 15000 },
      );
      await confirmModal(deviceModal);
      await expectApiResponseSuccess(await deviceUpdateResponse);
      await expect(
        adminPage.getByText("编辑设备策略绑定", { exact: true }),
      ).toBeHidden({ timeout: 15000 });

      const deviceAddOptionsResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/strategies") &&
          res.url().includes("enable=1") &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage.getByTestId("media-device-binding-add").click();
      await expectApiResponseSuccess(await deviceAddOptionsResponse);
      await expect(deviceModal.getByText("新增设备策略绑定")).toBeVisible();
      await expect(
        deviceModal.getByTestId("media-binding-device-id"),
      ).toHaveValue("");
      await expect(
        deviceModal.getByTestId("media-binding-device-id"),
      ).toBeEnabled();
      await deviceModal
        .getByTestId("media-binding-device-id")
        .fill(createdAfterEditDeviceId);
      const deviceCreateResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/device-bindings/${pathSegment(
                createdAfterEditDeviceId,
              )}`,
            ) && res.request().method() === "PUT",
        { timeout: 15000 },
      );
      await confirmModal(deviceModal);
      await expect(
        await expectApiResponseSuccess(await deviceCreateResponse),
      ).toMatchObject({
        deviceId: createdAfterEditDeviceId,
      });
      await expect(
        adminPage.getByText("新增设备策略绑定", { exact: true }),
      ).toBeHidden({ timeout: 15000 });

      const tenantBindingListResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/tenant-bindings") &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage
        .getByRole("tab", { exact: true, name: "租户绑定" })
        .click();
      await expectApiResponseSuccess(await tenantBindingListResponse);
      const tenantRow = tableRowByText(adminPage, tenantId);
      await expect(tenantRow).toBeVisible();
      const tenantStrategyOptionsResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/strategies") &&
          res.url().includes("enable=1") &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage
        .getByTestId(
          `media-tenant-binding-edit-${rowKeyTenant(tenantId)}`,
        )
        .click();
      await expectApiResponseSuccess(await tenantStrategyOptionsResponse);
      const tenantModal = visibleModalRoot(adminPage);
      await expect(tenantModal.getByText("编辑租户策略绑定")).toBeVisible();
      await expect(
        tenantModal.getByTestId("media-binding-tenant-id"),
      ).toHaveValue(tenantId);
      await expect(
        tenantModal.getByTestId("media-binding-tenant-id"),
      ).toBeDisabled();
      await expect(
        tenantModal
          .getByTestId("media-binding-strategy")
          .locator(".ant-select-selection-item"),
      ).toContainText(`#${strategyId}`);
      const tenantUpdateResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/tenant-bindings/${pathSegment(tenantId)}`,
            ) && res.request().method() === "PUT",
        { timeout: 15000 },
      );
      await confirmModal(tenantModal);
      await expectApiResponseSuccess(await tenantUpdateResponse);
      await expect(
        adminPage.getByText("编辑租户策略绑定", { exact: true }),
      ).toBeHidden({ timeout: 15000 });

      const tenantAddOptionsResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/strategies") &&
          res.url().includes("enable=1") &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage.getByTestId("media-tenant-binding-add").click();
      await expectApiResponseSuccess(await tenantAddOptionsResponse);
      await expect(tenantModal.getByText("新增租户策略绑定")).toBeVisible();
      await expect(
        tenantModal.getByTestId("media-binding-tenant-id"),
      ).toHaveValue("");
      await expect(
        tenantModal.getByTestId("media-binding-tenant-id"),
      ).toBeEnabled();
      await tenantModal
        .getByTestId("media-binding-tenant-id")
        .fill(createdAfterEditTenantId);
      const tenantCreateResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/tenant-bindings/${pathSegment(
                createdAfterEditTenantId,
              )}`,
            ) && res.request().method() === "PUT",
        { timeout: 15000 },
      );
      await confirmModal(tenantModal);
      await expect(
        await expectApiResponseSuccess(await tenantCreateResponse),
      ).toMatchObject({
        tenantId: createdAfterEditTenantId,
      });
      await expect(
        adminPage.getByText("新增租户策略绑定", { exact: true }),
      ).toBeHidden({ timeout: 15000 });

      const tenantDeviceBindingListResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/tenant-device-bindings") &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage
        .getByRole("tab", { exact: true, name: "租户设备绑定" })
        .click();
      await expectApiResponseSuccess(await tenantDeviceBindingListResponse);
      const tenantDeviceRow = tableRowByText(adminPage, tenantDeviceId);
      await expect(tenantDeviceRow).toBeVisible();
      const tenantDeviceStrategyOptionsResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/strategies") &&
          res.url().includes("enable=1") &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage
        .getByTestId(
          `media-tenant-device-binding-edit-${rowKeyTenantDevice(
            tenantId,
            tenantDeviceId,
          )}`,
        )
        .click();
      await expectApiResponseSuccess(await tenantDeviceStrategyOptionsResponse);
      const tenantDeviceModal = visibleModalRoot(adminPage);
      await expect(
        tenantDeviceModal.getByText("编辑租户设备策略绑定"),
      ).toBeVisible();
      await expect(
        tenantDeviceModal.getByTestId("media-binding-tenant-id"),
      ).toHaveValue(tenantId);
      await expect(
        tenantDeviceModal.getByTestId("media-binding-device-id"),
      ).toHaveValue(tenantDeviceId);
      await expect(
        tenantDeviceModal.getByTestId("media-binding-tenant-id"),
      ).toBeDisabled();
      await expect(
        tenantDeviceModal.getByTestId("media-binding-device-id"),
      ).toBeDisabled();
      await expect(
        tenantDeviceModal
          .getByTestId("media-binding-strategy")
          .locator(".ant-select-selection-item"),
      ).toContainText(`#${strategyId}`);
      const tenantDeviceUpdateResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/tenant-device-bindings/${pathSegment(
                tenantId,
              )}/${pathSegment(tenantDeviceId)}`,
            ) && res.request().method() === "PUT",
        { timeout: 15000 },
      );
      await confirmModal(tenantDeviceModal);
      await expectApiResponseSuccess(await tenantDeviceUpdateResponse);
      await expect(
        adminPage.getByText("编辑租户设备策略绑定", { exact: true }),
      ).toBeHidden({ timeout: 15000 });

      const tenantDeviceAddOptionsResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/strategies") &&
          res.url().includes("enable=1") &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage.getByTestId("media-tenant-device-binding-add").click();
      await expectApiResponseSuccess(await tenantDeviceAddOptionsResponse);
      await expect(
        tenantDeviceModal.getByText("新增租户设备策略绑定"),
      ).toBeVisible();
      await expect(
        tenantDeviceModal.getByTestId("media-binding-tenant-id"),
      ).toHaveValue("");
      await expect(
        tenantDeviceModal.getByTestId("media-binding-device-id"),
      ).toHaveValue("");
      await expect(
        tenantDeviceModal.getByTestId("media-binding-tenant-id"),
      ).toBeEnabled();
      await expect(
        tenantDeviceModal.getByTestId("media-binding-device-id"),
      ).toBeEnabled();
      await tenantDeviceModal
        .getByTestId("media-binding-tenant-id")
        .fill(createdAfterEditTenantDeviceTenantId);
      await tenantDeviceModal
        .getByTestId("media-binding-device-id")
        .fill(createdAfterEditTenantDeviceId);
      const tenantDeviceCreateResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/tenant-device-bindings/${pathSegment(
                createdAfterEditTenantDeviceTenantId,
              )}/${pathSegment(createdAfterEditTenantDeviceId)}`,
            ) && res.request().method() === "PUT",
        { timeout: 15000 },
      );
      await confirmModal(tenantDeviceModal);
      await expect(
        await expectApiResponseSuccess(await tenantDeviceCreateResponse),
      ).toMatchObject({
        deviceId: createdAfterEditTenantDeviceId,
        tenantId: createdAfterEditTenantDeviceTenantId,
      });
      await expect(
        adminPage.getByText("新增租户设备策略绑定", { exact: true }),
      ).toBeHidden({ timeout: 15000 });

      const resolveResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/strategies/resolve") &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage
        .getByRole("tab", { exact: true, name: "策略解析" })
        .click();
      const resolvePane = adminPage.locator(".ant-tabs-tabpane-active");
      await resolvePane.getByPlaceholder("tenant-a").fill(tenantId);
      await resolvePane
        .getByPlaceholder("34020000001320000001")
        .fill(tenantDeviceId);
      await resolvePane.getByRole("button", { name: "解析生效策略" }).click();
      const resolvePayload = await expectApiResponseSuccess(
        await resolveResponse,
      );
      expect(resolvePayload).toMatchObject({
        matched: true,
        source: "tenantDevice",
        strategyId,
      });
      await expect(resolvePane.getByText("租户设备策略")).toBeVisible();

      const aliasListResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes("/api/v1/media/stream-aliases") &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage.getByRole("tab", { exact: true, name: "流别名" }).click();
      await expectApiResponseSuccess(await aliasListResponse);
      const aliasRow = tableRowByText(adminPage, alias);
      await expect(aliasRow).toBeVisible();
      const aliasDetailResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes(`/api/v1/media/stream-aliases/${aliasId}`) &&
          res.request().method() === "GET",
        { timeout: 15000 },
      );
      await adminPage.getByTestId(`media-alias-edit-${aliasId}`).click();
      await expectApiResponseSuccess(await aliasDetailResponse);
      const aliasModal = visibleModalRoot(adminPage);
      await expect(aliasModal.getByText("编辑流别名")).toBeVisible();
      await expect(aliasModal.getByTestId("media-alias-name")).toHaveValue(
        alias,
      );
      await expect(
        aliasModal.getByTestId("media-alias-stream-path"),
      ).toHaveValue(`live/${alias}`);
      await expectCheckedRadioLabel(
        aliasModal.getByTestId("media-alias-auto-remove"),
        "否",
      );
      await aliasModal
        .getByTestId("media-alias-stream-path")
        .fill(`live/${alias}-ui-updated`);
      const aliasUpdateResponse = adminPage.waitForResponse(
        (res) =>
          res.url().includes(`/api/v1/media/stream-aliases/${aliasId}`) &&
          res.request().method() === "PUT",
        { timeout: 15000 },
      );
      await confirmModal(aliasModal);
      await expectApiResponseSuccess(await aliasUpdateResponse);
      await expect(
        adminPage.getByText("编辑流别名", { exact: true }),
      ).toBeHidden({ timeout: 15000 });
      await expect(
        expectSuccess<AliasDetail>(
          await api.get(`media/stream-aliases/${aliasId}`),
        ),
      ).resolves.toMatchObject({
        alias,
        streamPath: `live/${alias}-ui-updated`,
      });

      await adminPage.getByTestId("media-alias-add").click();
      await expect(aliasModal.getByText("新增流别名")).toBeVisible();
      await expect(aliasModal.getByTestId("media-alias-name")).toHaveValue("");
      await expect(
        aliasModal.getByTestId("media-alias-stream-path"),
      ).toHaveValue("");
      await aliasModal
        .getByTestId("media-alias-name")
        .fill(createdAfterEditAlias);
      await aliasModal
        .getByTestId("media-alias-stream-path")
        .fill(`live/${createdAfterEditAlias}`);
      const aliasCreateResponse = adminPage.waitForResponse(
        (res) =>
          res.url().endsWith("/api/v1/media/stream-aliases") &&
          res.request().method() === "POST",
        { timeout: 15000 },
      );
      await confirmModal(aliasModal);
      const createdAliasPayload = await expectApiResponseSuccess(
        await aliasCreateResponse,
      );
      createdAfterEditAliasId = createdAliasPayload.id;
      expect(createdAfterEditAliasId).toBeGreaterThan(0);
      await expect(
        adminPage.getByText("新增流别名", { exact: true }),
      ).toBeHidden({ timeout: 15000 });
      await expect(
        expectSuccess<AliasDetail>(
          await api.get(`media/stream-aliases/${createdAfterEditAliasId}`),
        ),
      ).resolves.toMatchObject({
        alias: createdAfterEditAlias,
        streamPath: `live/${createdAfterEditAlias}`,
      });

      await adminPage.getByRole("tab", { exact: true, name: "设备绑定" }).click();
      const deviceDeleteResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/device-bindings/${pathSegment(
                createdAfterEditDeviceId,
              )}`,
            ) && res.request().method() === "DELETE",
        { timeout: 15000 },
      );
      await adminPage
        .getByTestId(
          `media-device-binding-delete-${rowKeyDevice(
            createdAfterEditDeviceId,
          )}`,
        )
        .click();
      await confirmPopconfirm(adminPage);
      await expectApiResponseSuccess(await deviceDeleteResponse);

      await adminPage.getByRole("tab", { exact: true, name: "租户绑定" }).click();
      const tenantDeleteResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/tenant-bindings/${pathSegment(
                createdAfterEditTenantId,
              )}`,
            ) && res.request().method() === "DELETE",
        { timeout: 15000 },
      );
      await adminPage
        .getByTestId(
          `media-tenant-binding-delete-${rowKeyTenant(
            createdAfterEditTenantId,
          )}`,
        )
        .click();
      await confirmPopconfirm(adminPage);
      await expectApiResponseSuccess(await tenantDeleteResponse);

      await adminPage
        .getByRole("tab", { exact: true, name: "租户设备绑定" })
        .click();
      const tenantDeviceDeleteResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/tenant-device-bindings/${pathSegment(
                createdAfterEditTenantDeviceTenantId,
              )}/${pathSegment(createdAfterEditTenantDeviceId)}`,
            ) && res.request().method() === "DELETE",
        { timeout: 15000 },
      );
      await adminPage
        .getByTestId(
          `media-tenant-device-binding-delete-${rowKeyTenantDevice(
            createdAfterEditTenantDeviceTenantId,
            createdAfterEditTenantDeviceId,
          )}`,
        )
        .click();
      await confirmPopconfirm(adminPage);
      await expectApiResponseSuccess(await tenantDeviceDeleteResponse);

      await adminPage.getByRole("tab", { exact: true, name: "流别名" }).click();
      const aliasDeleteResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/stream-aliases/${createdAfterEditAliasId}`,
            ) && res.request().method() === "DELETE",
        { timeout: 15000 },
      );
      await adminPage
        .getByTestId(`media-alias-delete-${createdAfterEditAliasId}`)
        .click();
      await confirmPopconfirm(adminPage);
      await expectApiResponseSuccess(await aliasDeleteResponse);
      createdAfterEditAliasId = 0;

      await adminPage
        .getByRole("tab", { exact: true, name: "策略管理" })
        .click();
      const strategyDeleteResponse = adminPage.waitForResponse(
        (res) =>
          res
            .url()
            .includes(
              `/api/v1/media/strategies/${createdAfterEditStrategyId}`,
            ) && res.request().method() === "DELETE",
        { timeout: 15000 },
      );
      await adminPage
        .getByTestId(`media-strategy-delete-${createdAfterEditStrategyId}`)
        .click();
      await confirmPopconfirm(adminPage);
      await expectApiResponseSuccess(await strategyDeleteResponse);
      createdAfterEditStrategyId = 0;
      if (previousGlobalStrategyId > 0) {
        await expectSuccess(
          await api.put(`media/strategies/${previousGlobalStrategyId}/global`),
        );
        previousGlobalStrategyId = 0;
      }
    } finally {
      if (previousGlobalStrategyId > 0) {
        await api
          .put(`media/strategies/${previousGlobalStrategyId}/global`)
          .catch(() => undefined);
      }
      await deleteTenantDeviceBinding(
        api,
        createdAfterEditTenantDeviceTenantId,
        createdAfterEditTenantDeviceId,
      ).catch(() => undefined);
      await deleteTenantBinding(api, createdAfterEditTenantId).catch(
        () => undefined,
      );
      await deleteDeviceBinding(api, createdAfterEditDeviceId).catch(
        () => undefined,
      );
      await deleteTenantDeviceBinding(api, tenantId, tenantDeviceId).catch(
        () => undefined,
      );
      await deleteTenantBinding(api, tenantId).catch(() => undefined);
      await deleteDeviceBinding(api, deviceId).catch(() => undefined);
      if (createdAfterEditAliasId > 0) {
        await api
          .delete(`media/stream-aliases/${createdAfterEditAliasId}`)
          .catch(() => undefined);
      }
      if (aliasId > 0) {
        await api
          .delete(`media/stream-aliases/${aliasId}`)
          .catch(() => undefined);
      }
      for (const id of [
        createdAfterEditStrategyId,
        replacementStrategyId,
        strategyId,
      ]) {
        if (id > 0) {
          await api.delete(`media/strategies/${id}`).catch(() => undefined);
        }
      }
      await api.dispose();
    }
  });
});
