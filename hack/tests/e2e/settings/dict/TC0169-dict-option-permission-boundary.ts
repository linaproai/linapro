import type { APIRequestContext, APIResponse } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import {
  createAdminApiContext,
  createApiContext,
  createRole,
  createUser,
  disablePlugin,
  enablePlugin,
  expectSuccess,
  getPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
} from "../../../support/api/job";

const limitedPassword = "test123456";
const pluginIDs = ["org-center", "content-notice"];
const businessPermissions = [
  "system:dept:list",
  "system:dept:query",
  "system:dept:add",
  "system:dept:edit",
  "system:dept:remove",
  "system:post:list",
  "system:post:query",
  "system:post:add",
  "system:post:edit",
  "system:post:remove",
  "system:post:export",
  "system:notice:list",
  "system:notice:query",
  "system:notice:add",
  "system:notice:edit",
  "system:notice:remove",
];

type PluginState = {
  enabled: number;
  installed: number;
};

type DictOptionItem = {
  label: string;
  value: string;
};

type ErrorEnvelope = {
  code: number;
  errorCode?: string;
  message?: string;
  messageKey?: string;
};

type MenuNode = {
  children?: MenuNode[];
  id: number;
  perms: string;
};

async function expectForbidden(response: APIResponse) {
  expect(response.status()).toBe(403);
  const body = await response.text();
  const jsonStart = body.indexOf("{");
  expect(jsonStart).toBeGreaterThanOrEqual(0);
  const payload = JSON.parse(body.slice(jsonStart)) as ErrorEnvelope;
  expect(payload.code).not.toBe(0);
  expect(
    payload.messageKey ?? payload.errorCode ?? payload.message ?? "",
  ).toContain("permission");
  return payload;
}

function collectMenuIdsByPermsWithAncestors(
  menus: MenuNode[],
  perms: string[],
) {
  const requiredPerms = new Set(perms);
  const selectedIds = new Set<number>();

  function visit(node: MenuNode, ancestors: number[]) {
    const nextAncestors = [...ancestors, node.id];
    if (requiredPerms.has(node.perms)) {
      nextAncestors.forEach((id) => selectedIds.add(id));
    }
    node.children?.forEach((child) => visit(child, nextAncestors));
  }

  menus.forEach((node) => visit(node, []));
  for (const permission of requiredPerms) {
    expect(
      menus.some((node) => hasPermission(node, permission)),
      `missing menu permission: ${permission}`,
    ).toBeTruthy();
  }
  return [...selectedIds];
}

function hasPermission(node: MenuNode, permission: string): boolean {
  return (
    node.perms === permission ||
    Boolean(node.children?.some((child) => hasPermission(child, permission)))
  );
}

test.describe("TC-169 Dictionary option permission boundary", () => {
  let adminApi: APIRequestContext;
  let limitedApi: APIRequestContext;
  let limitedRoleID = 0;
  let limitedUserID = 0;
  const originalPluginStates = new Map<string, PluginState>();

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await syncPlugins(adminApi);

    for (const pluginID of pluginIDs) {
      const plugin = await getPlugin(adminApi, pluginID);
      originalPluginStates.set(pluginID, {
        enabled: plugin.enabled,
        installed: plugin.installed,
      });
      if (plugin.installed !== 1) {
        await installPlugin(adminApi, pluginID);
      }
      if (plugin.enabled !== 1) {
        await enablePlugin(adminApi, pluginID);
      }
    }

    const suffix = Date.now();
    const menuTree = await expectSuccess<{ list: MenuNode[] }>(
      await adminApi.get("menu"),
    );
    const menuIds = collectMenuIdsByPermsWithAncestors(
      menuTree.list,
      businessPermissions,
    );
    const role = await createRole(adminApi, {
      key: `e2e_dict_opt_${suffix}`,
      menuIds,
      name: `E2E Dict Option ${suffix}`,
      sort: 980,
    });
    limitedRoleID = role.id;

    const username = `e2e_dict_option_${suffix}`;
    const user = await createUser(adminApi, {
      nickname: `E2E Dict Option ${suffix}`,
      password: limitedPassword,
      roleIds: [limitedRoleID],
      username,
    });
    limitedUserID = user.id;
    limitedApi = await createApiContext(username, limitedPassword);
  });

  test.afterAll(async () => {
    await limitedApi?.dispose();
    if (limitedUserID > 0) {
      await adminApi.delete(`user?ids=${limitedUserID}`).catch(() => {});
    }
    if (limitedRoleID > 0) {
      await adminApi.delete(`role?ids=${limitedRoleID}`).catch(() => {});
    }

    for (const pluginID of [...pluginIDs].reverse()) {
      const original = originalPluginStates.get(pluginID);
      if (!original) {
        continue;
      }
      if (original.installed !== 1) {
        await uninstallPlugin(adminApi, pluginID).catch(() => {});
      } else if (original.enabled !== 1) {
        await disablePlugin(adminApi, pluginID).catch(() => {});
      } else {
        await enablePlugin(adminApi, pluginID).catch(() => {});
      }
    }
    await adminApi?.dispose();
  });

  test("TC-169a: business-only user can read reusable dictionary options", async () => {
    const statusOptions = await expectSuccess<{ list: DictOptionItem[] }>(
      await limitedApi.get("dict/data/type/sys_normal_disable"),
    );
    const noticeTypeOptions = await expectSuccess<{ list: DictOptionItem[] }>(
      await limitedApi.get("dict/data/type/sys_notice_type"),
    );
    const noticeStatusOptions = await expectSuccess<{ list: DictOptionItem[] }>(
      await limitedApi.get("dict/data/type/sys_notice_status"),
    );

    expect(statusOptions.list.map((item) => item.value)).toContain("1");
    expect(noticeTypeOptions.list.length).toBeGreaterThan(0);
    expect(noticeStatusOptions.list.length).toBeGreaterThan(0);
  });

  test("TC-169b: business-only user still cannot access dictionary management APIs", async () => {
    await expectForbidden(
      await limitedApi.get("dict/data?pageNum=1&pageSize=10"),
    );
    await expectForbidden(
      await limitedApi.get("dict/type?pageNum=1&pageSize=10"),
    );
  });
});
