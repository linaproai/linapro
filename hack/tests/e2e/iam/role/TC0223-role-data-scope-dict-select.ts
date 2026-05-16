import type { Page } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { RolePage } from "../../../pages/RolePage";
import { createAdminApiContext, expectSuccess } from "../../../support/api/job";

type CreatedID = {
  id: number;
};

test.describe("TC-223 角色数据权限字典下拉框", () => {
  test("TC-223a~c: 列表显示数据权限标签,下拉框按租户插件状态过滤本租户数据", async ({
    adminPage,
  }) => {
    await mockMultiTenantPluginState(adminPage, true);
    const adminApi = await createAdminApiContext();
    const suffix = Date.now();
    const tenantRoleName = `E2E Tenant Scope ${suffix}`;
    const tenantRoleKey = `e2e_tenant_scope_${suffix}`;
    const selfRoleName = `E2E Self Scope ${suffix}`;
    const selfRoleKey = `e2e_self_scope_${suffix}`;
    const roleIds: number[] = [];

    try {
      const tenantCreated = await expectSuccess<CreatedID>(
        await adminApi.post("role", {
          data: {
            dataScope: 2,
            key: tenantRoleKey,
            name: tenantRoleName,
            sort: 999,
            status: 1,
          },
        }),
      );
      roleIds.push(tenantCreated.id);
      const selfCreated = await expectSuccess<CreatedID>(
        await adminApi.post("role", {
          data: {
            dataScope: 4,
            key: selfRoleKey,
            name: selfRoleName,
            sort: 999,
            status: 1,
          },
        }),
      );
      roleIds.push(selfCreated.id);

      const rolePage = new RolePage(adminPage);
      await rolePage.goto();
      await rolePage.searchRole(selfRoleName);

      const row = rolePage.roleRowByKey(selfRoleKey);
      await expect(row).toContainText("本人数据");
      await expect(row.locator(".vxe-cell", { hasText: /^4$/ })).toHaveCount(0);

      const drawer = await rolePage.openCreateDrawer();
      expect(await rolePage.selectedDataScopeText(drawer)).toBe("本租户数据");
      const dataScopeOptions = await rolePage.getDataScopeOptions(drawer);
      expect(dataScopeOptions).toEqual(
        expect.arrayContaining(["全部数据", "本租户数据", "本人数据"]),
      );
      expect(dataScopeOptions).not.toContain("4");

      await mockMultiTenantPluginState(adminPage, false);
      await rolePage.goto();
      await rolePage.searchRole(tenantRoleName);
      const tenantRowText = await rolePage.roleRowTextByKey(tenantRoleKey);
      expect(tenantRowText).toContain("全部数据");
      expect(tenantRowText).not.toContain("本租户数据");

      const singleTenantDrawer = await rolePage.openCreateDrawer();
      expect(await rolePage.selectedDataScopeText(singleTenantDrawer)).toBe(
        "全部数据",
      );
      const singleTenantOptions =
        await rolePage.getDataScopeOptions(singleTenantDrawer);
      expect(singleTenantOptions).toEqual(
        expect.arrayContaining(["全部数据", "本人数据"]),
      );
      expect(singleTenantOptions).not.toContain("本租户数据");
    } finally {
      for (const roleId of roleIds) {
        await adminApi.delete(`role/${roleId}`).catch(() => {});
      }
      await adminApi.dispose();
    }
  });

  test("TC-223d: 源码多租户启用态不依赖动态插件状态列表", async ({
    adminPage,
  }) => {
    await mockPluginRuntimeStates(adminPage, []);
    await mockUserInfoWithTenantMenu(adminPage);
    await adminPage.reload({ waitUntil: "domcontentloaded" });

    const roleName = "普通用户";

    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    const createDrawer = await rolePage.openCreateDrawer();
    expect(await rolePage.selectedDataScopeText(createDrawer)).toBe(
      "本租户数据",
    );
    expect(await rolePage.getDataScopeOptions(createDrawer)).toEqual(
      expect.arrayContaining(["本租户数据"]),
    );

    await rolePage.goto();
    await rolePage.searchRole(roleName);
    const editDrawer = await rolePage.openEditDrawer(roleName);
    expect(await rolePage.selectedDataScopeText(editDrawer)).toBe("本人数据");
    expect(await rolePage.getDataScopeOptions(editDrawer)).toEqual(
      expect.arrayContaining(["本租户数据"]),
    );
  });
});

async function mockUserInfoWithTenantMenu(page: Page) {
  await page.unroute("**/api/v1/user/info").catch(() => {});
  await page.route("**/api/v1/user/info", async (route) => {
    const response = await route.fetch();
    const payload = (await response.json()) as Record<string, unknown>;
    const data = (payload.data ?? {}) as Record<string, unknown>;
    const menus = Array.isArray(data.menus) ? data.menus : [];

    await route.fulfill({
      body: JSON.stringify({
        ...payload,
        data: {
          ...data,
          menus: appendTenantMenu(menus),
        },
      }),
      contentType: "application/json",
      response,
    });
  });
}

function appendTenantMenu(menus: unknown[]) {
  return [
    ...menus,
    {
      children: [],
      id: 990223,
      name: "PlatformTenantManagement",
      path: "/platform/tenants",
    },
  ];
}

async function mockMultiTenantPluginState(page: Page, enabled: boolean) {
  await mockPluginRuntimeStates(page, [
    {
      enabled: enabled ? 1 : 0,
      generation: 1,
      id: "multi-tenant",
      installed: 1,
      statusKey: "sys_plugin.status:multi-tenant",
      version: "e2e",
    },
  ]);
}

async function mockPluginRuntimeStates(
  page: Page,
  items: Array<Record<string, unknown>>,
) {
  await page.unroute("**/api/v1/plugins/dynamic**").catch(() => {});
  await page.route("**/api/v1/plugins/dynamic**", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        code: 0,
        data: {
          list: items,
        },
        message: "success",
      }),
      status: 200,
    });
  });
  await page.evaluate(() => {
    const registryGlobal = globalThis as any;
    registryGlobal.__linaPluginStatePromise = null;
    registryGlobal.__linaPluginStateSignature = null;
  });
}
