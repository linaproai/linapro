import type { APIRequestContext } from "@host-tests/support/playwright";

import { test, expect } from "@host-tests/fixtures/auth";
import {
  createAdminApiContext,
  createApiContext,
  createRole,
  createUser,
  deleteRole,
  deleteUser,
  disablePlugin,
  enablePlugin,
  expectBusinessError,
  expectSuccess,
  getMenuIdsByPermsWithAncestors,
  getPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
} from "@host-tests/support/api/job";

type PluginState = {
  enabled: number;
  installed: number;
};

type OnlineUserItem = {
  tokenId: string;
  username: string;
};

const password = "test123456";
const pluginID = "monitor-online";

async function ensurePluginEnabled(api: APIRequestContext, id: string) {
  await syncPlugins(api);
  const plugin = await getPlugin(api, id);
  if (plugin.installed !== 1) {
    await installPlugin(api, id);
  }
  if (plugin.enabled !== 1) {
    await enablePlugin(api, id);
  }
  return plugin;
}

async function restorePluginState(
  api: APIRequestContext,
  id: string,
  state: PluginState,
) {
  if (state.installed !== 1) {
    await uninstallPlugin(api, id).catch(() => {});
    return;
  }
  if (state.enabled !== 1) {
    await disablePlugin(api, id).catch(() => {});
    return;
  }
  await enablePlugin(api, id).catch(() => {});
}

async function listOnline(api: APIRequestContext, username = "") {
  const query = new URLSearchParams({ pageNum: "1", pageSize: "100" });
  if (username) {
    query.set("username", username);
  }
  return expectSuccess<{ items: OnlineUserItem[]; total: number }>(
    await api.get(`monitor/online/list?${query.toString()}`),
  );
}

test.describe("TC-173 在线用户数据权限", () => {
  let adminApi: APIRequestContext;
  let limitedApi: APIRequestContext;
  let otherApi: APIRequestContext;
  let originalPluginState: PluginState;
  let roleID = 0;
  let limitedUserID = 0;
  let otherUserID = 0;
  let limitedUsername = "";
  let otherUsername = "";

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    const plugin = await ensurePluginEnabled(adminApi, pluginID);
    originalPluginState = {
      enabled: plugin.enabled,
      installed: plugin.installed,
    };

    const suffix = Date.now().toString();
    const menuIds = await getMenuIdsByPermsWithAncestors(adminApi, [
      "monitor:online:query",
      "monitor:online:forceLogout",
    ]);
    roleID = (
      await createRole(adminApi, {
        name: `OnlineScope${suffix.slice(-10)}`,
        key: `e2e_online_self_${suffix}`,
        menuIds,
        dataScope: 4,
        sort: 970,
      })
    ).id;

    limitedUsername = `e2e_online_limited_${suffix}`;
    otherUsername = `e2e_online_other_${suffix}`;
    limitedUserID = (
      await createUser(adminApi, {
        username: limitedUsername,
        password,
        nickname: "E2E Online Limited",
        roleIds: [roleID],
      })
    ).id;
    otherUserID = (
      await createUser(adminApi, {
        username: otherUsername,
        password,
        nickname: "E2E Online Other",
      })
    ).id;

    limitedApi = await createApiContext(limitedUsername, password);
    otherApi = await createApiContext(otherUsername, password);
  });

  test.afterAll(async () => {
    await limitedApi?.post("auth/logout").catch(() => {});
    await otherApi?.post("auth/logout").catch(() => {});
    await limitedApi?.dispose();
    await otherApi?.dispose();
    for (const userID of [limitedUserID, otherUserID]) {
      if (userID > 0) {
        await deleteUser(adminApi, userID).catch(() => {});
      }
    }
    if (roleID > 0) {
      await deleteRole(adminApi, roleID).catch(() => {});
    }
    if (originalPluginState) {
      await restorePluginState(adminApi, pluginID, originalPluginState);
    }
    await adminApi?.dispose();
  });

  test("TC-173a~b: 仅本人范围过滤在线列表并拒绝强制下线他人", async () => {
    const scopedOnline = await listOnline(limitedApi);
    const scopedUsernames = scopedOnline.items.map((item) => item.username);
    expect(scopedUsernames).toContain(limitedUsername);
    expect(scopedUsernames).not.toContain(otherUsername);

    const adminOnline = await listOnline(adminApi, otherUsername);
    const otherSession = adminOnline.items.find(
      (item) => item.username === otherUsername,
    );
    expect(otherSession, "other user should have an online session").toBeTruthy();

    await expectBusinessError(
      await limitedApi.delete(`monitor/online/${otherSession!.tokenId}`),
    );

    const otherStillOnline = await listOnline(adminApi, otherUsername);
    expect(otherStillOnline.items.some((item) => item.username === otherUsername))
      .toBeTruthy();
  });
});
