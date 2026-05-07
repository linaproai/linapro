import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
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
} from "../../../support/api/job";

type PluginState = {
  enabled: number;
  installed: number;
};

type DeptCreateResult = {
  id: number;
};

type UserListItem = {
  id: number;
  username: string;
};

type RoleUserItem = {
  id: number;
  username: string;
};

const password = "test123456";
const pluginID = "org-center";

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

async function createDept(
  api: APIRequestContext,
  suffix: string,
  label: string,
) {
  return expectSuccess<DeptCreateResult>(
    await api.post("dept", {
      data: {
        parentId: 0,
        name: `E2E ${label} Dept ${suffix}`,
        code: `e2e_${label}_${suffix}`,
        orderNum: 900,
        status: 1,
      },
    }),
  );
}

async function listUsers(api: APIRequestContext, username: string) {
  return expectSuccess<{ list: UserListItem[]; total: number }>(
    await api.get(
      `user?pageNum=1&pageSize=100&username=${encodeURIComponent(username)}`,
    ),
  );
}

async function listRoleUsers(api: APIRequestContext, roleID: number) {
  return expectSuccess<{ list: RoleUserItem[]; total: number }>(
    await api.get(`role/${roleID}/users?page=1&size=100`),
  );
}

test.describe("TC-170 用户管理数据权限", () => {
  let adminApi: APIRequestContext;
  let limitedApi: APIRequestContext;
  let originalPluginState: PluginState;
  let limitedRoleID = 0;
  let managedRoleID = 0;
  let limitedUserID = 0;
  let sameDeptUserID = 0;
  let otherDeptUserID = 0;
  let deptAID = 0;
  let deptBID = 0;
  let suffix = "";
  let limitedUsername = "";
  let sameDeptUsername = "";
  let otherDeptUsername = "";

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    const plugin = await ensurePluginEnabled(adminApi, pluginID);
    originalPluginState = {
      enabled: plugin.enabled,
      installed: plugin.installed,
    };

    suffix = Date.now().toString();
    const deptA = await createDept(adminApi, suffix, "alpha");
    deptAID = deptA.id;
    const deptB = await createDept(adminApi, suffix, "beta");
    deptBID = deptB.id;

    const menuIds = await getMenuIdsByPermsWithAncestors(adminApi, [
      "system:user:query",
      "system:user:edit",
      "system:role:auth",
    ]);
    const limitedRole = await createRole(adminApi, {
      name: `UserScope${suffix.slice(-10)}`,
      key: `e2e_user_dept_${suffix}`,
      menuIds,
      dataScope: 2,
      sort: 970,
    });
    limitedRoleID = limitedRole.id;

    const managedRole = await createRole(adminApi, {
      name: `Managed${suffix.slice(-10)}`,
      key: `e2e_managed_${suffix}`,
      menuIds: [],
      dataScope: 1,
      sort: 971,
    });
    managedRoleID = managedRole.id;

    limitedUsername = `e2e_scope_limited_${suffix}`;
    sameDeptUsername = `e2e_scope_same_${suffix}`;
    otherDeptUsername = `e2e_scope_other_${suffix}`;

    limitedUserID = (
      await createUser(adminApi, {
        username: limitedUsername,
        password,
        nickname: "E2E Scope Limited",
        deptId: deptAID,
        roleIds: [limitedRoleID],
      })
    ).id;
    sameDeptUserID = (
      await createUser(adminApi, {
        username: sameDeptUsername,
        password,
        nickname: "E2E Scope Same",
        deptId: deptAID,
      })
    ).id;
    otherDeptUserID = (
      await createUser(adminApi, {
        username: otherDeptUsername,
        password,
        nickname: "E2E Scope Other",
        deptId: deptBID,
      })
    ).id;

    await expectSuccess(
      await adminApi.post(`role/${managedRoleID}/users`, {
        data: { userIds: [sameDeptUserID, otherDeptUserID] },
      }),
    );

    limitedApi = await createApiContext(limitedUsername, password);
  });

  test.afterAll(async () => {
    await limitedApi?.post("auth/logout").catch(() => {});
    await limitedApi?.dispose();
    for (const userID of [limitedUserID, sameDeptUserID, otherDeptUserID]) {
      if (userID > 0) {
        await deleteUser(adminApi, userID).catch(() => {});
      }
    }
    for (const roleID of [managedRoleID, limitedRoleID]) {
      if (roleID > 0) {
        await deleteRole(adminApi, roleID).catch(() => {});
      }
    }
    for (const deptID of [deptAID, deptBID]) {
      if (deptID > 0) {
        await adminApi.delete(`dept/${deptID}`).catch(() => {});
      }
    }
    if (originalPluginState) {
      await restorePluginState(adminApi, pluginID, originalPluginState);
    }
    await adminApi?.dispose();
  });

  test("TC-170a~d: 本部门范围过滤列表、详情、写操作和角色授权用户", async () => {
    const scopedUsers = await listUsers(limitedApi, suffix);
    const scopedUsernames = scopedUsers.list.map((item) => item.username);
    expect(scopedUsernames).toContain(limitedUsername);
    expect(scopedUsernames).toContain(sameDeptUsername);
    expect(scopedUsernames).not.toContain(otherDeptUsername);

    await expectSuccess(await limitedApi.get(`user/${sameDeptUserID}`));
    await expectBusinessError(await limitedApi.get(`user/${otherDeptUserID}`));
    await expectBusinessError(
      await limitedApi.put(`user/${otherDeptUserID}`, {
        data: {
          id: otherDeptUserID,
          nickname: "Blocked Update",
        },
      }),
    );

    const roleUsers = await listRoleUsers(limitedApi, managedRoleID);
    const roleUsernames = roleUsers.list.map((item) => item.username);
    expect(roleUsernames).toContain(sameDeptUsername);
    expect(roleUsernames).not.toContain(otherDeptUsername);

    await expectBusinessError(
      await limitedApi.post(`role/${managedRoleID}/users`, {
        data: { userIds: [otherDeptUserID] },
      }),
    );
  });
});
