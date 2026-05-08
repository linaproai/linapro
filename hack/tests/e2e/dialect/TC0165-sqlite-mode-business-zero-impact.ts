import { test, expect } from "../../fixtures/auth";
import {
  createAdminApiContext,
  findPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
  updatePluginStatus,
} from "../../fixtures/plugin";
import {
  expectApiSuccess,
  requireSQLiteE2E,
  sqliteSourcePluginId,
} from "../../support/sqlite-e2e";
import { listLogs } from "../../support/api/job";

type UserListItem = {
  id: number;
  nickname: string;
  username: string;
};

type UserCreateResult = {
  id: number;
};

test.describe("TC-165 SQLite mode business zero impact", () => {
  requireSQLiteE2E();

  test("TC-165a~c: user CRUD, execution log list, and source plugin lifecycle work on SQLite", async () => {
    const api = await createAdminApiContext();
    const username = `sqlite_e2e_${Date.now()}`;
    let createdUserId = 0;

    try {
      const created = await expectApiSuccess<UserCreateResult>(
        await api.post("user", {
          data: {
            nickname: "SQLite E2E User",
            password: "test123456",
            username,
          },
        }),
        "create user in SQLite mode",
      );
      createdUserId = created.id;

      let users = await expectApiSuccess<{ list: UserListItem[] }>(
        await api.get("user", { params: { username } }),
        "query created user in SQLite mode",
      );
      expect(users.list.some((item) => item.username === username)).toBe(true);

      await expectApiSuccess<unknown>(
        await api.put(`user/${createdUserId}`, {
          data: {
            nickname: "SQLite E2E User Updated",
            username,
          },
        }),
        "update user in SQLite mode",
      );

      users = await expectApiSuccess<{ list: UserListItem[] }>(
        await api.get("user", { params: { username } }),
        "query updated user in SQLite mode",
      );
      expect(
        users.list.find((item) => item.username === username)?.nickname,
      ).toBe("SQLite E2E User Updated");

      await expectApiSuccess<unknown>(
        await api.delete(`user/${createdUserId}`),
        "delete user in SQLite mode",
      );
      createdUserId = 0;

      users = await expectApiSuccess<{ list: UserListItem[] }>(
        await api.get("user", { params: { username } }),
        "query deleted user in SQLite mode",
      );
      expect(users.list.some((item) => item.username === username)).toBe(false);

      const logs = await listLogs(api);
      expect(logs.total).toBeGreaterThanOrEqual(0);
      expect(Array.isArray(logs.list)).toBe(true);

      await syncPlugins(api);
      let plugin = await findPlugin(api, sqliteSourcePluginId);
      expect(plugin, `expected ${sqliteSourcePluginId} to be discoverable`).toBeTruthy();
      if (plugin?.installed === 1 && plugin.enabled === 1) {
        await updatePluginStatus(api, sqliteSourcePluginId, false);
      }
      if (plugin?.installed !== 1) {
        await installPlugin(api, sqliteSourcePluginId);
      }
      await updatePluginStatus(api, sqliteSourcePluginId, true);
      plugin = await findPlugin(api, sqliteSourcePluginId);
      expect(plugin?.installed).toBe(1);
      expect(plugin?.enabled).toBe(1);

      await updatePluginStatus(api, sqliteSourcePluginId, false);
      plugin = await findPlugin(api, sqliteSourcePluginId);
      expect(plugin?.enabled).toBe(0);
    } finally {
      if (createdUserId > 0) {
        await api.delete(`user/${createdUserId}`).catch(() => undefined);
      }
      await uninstallPlugin(api, sqliteSourcePluginId, true).catch(
        () => undefined,
      );
      await api.dispose();
    }
  });
});
