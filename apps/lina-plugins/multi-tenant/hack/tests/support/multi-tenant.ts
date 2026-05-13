import { config } from "@host-tests/fixtures/config";
import { test as authTest, expect } from "@host-tests/fixtures/auth";
import {
  enablePlugin,
  expectSuccess,
  getPlugin,
  getMenuIdsByPerms,
  installPlugin,
  playwrightRequest,
  syncPlugins,
  type APIRequestContext,
  type APIResponse,
} from "@host-tests/support/api/job";
import {
  execPgSQL,
  pgEscapeLiteral,
  queryPgRows,
  queryPgScalar,
} from "@host-tests/support/postgres";

export type MultiTenantMode = "multi-tenant-disabled" | "multi-tenant-enabled";

export type MultiTenantFixtures = {
  multiTenantMode: MultiTenantMode;
};

const apiBaseURL =
  process.env.E2E_API_BASE_URL ??
  new URL("/api/v1/", config.baseURL).toString();

export type LoginTenant = {
  id: number;
  code: string;
  name: string;
  status: string;
};

export type TenantCreateResult = {
  id: number;
};

export type TenantMember = {
  id: number;
  tenantId: number;
  userId: number;
  username: string;
  status: number;
};

export type TenantUserGrant = {
  roleId: number;
  tenantId: number;
};

export const test = authTest.extend<MultiTenantFixtures>({
  multiTenantMode: ["multi-tenant-enabled", { option: true }],
});

export { expect };

export async function ensureMultiTenantPluginEnabled(api: APIRequestContext) {
  await syncPlugins(api);
  const plugin = await getPlugin(api, "multi-tenant");
  if (plugin.installed !== 1) {
    await installPlugin(api, "multi-tenant");
  }
  if (plugin.enabled !== 1) {
    await enablePlugin(api, "multi-tenant");
  }
  return plugin;
}

export async function loginRaw(username = config.adminUser, password = config.adminPass) {
  const api = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const response = await api.post("auth/login", {
    data: { username, password },
  });
  expect(response.ok()).toBeTruthy();
  const payload = await response.json();
  expect(payload.code).toBe(0);
  await api.dispose();
  return payload.data as {
    accessToken?: string;
    preToken?: string;
    tenants?: LoginTenant[];
  };
}

export async function createTenant(
  api: APIRequestContext,
  payload: { code: string; name: string; remark?: string },
) {
  return expectSuccess<TenantCreateResult>(
    await api.post("platform/tenants", {
      data: {
        remark: "",
        ...payload,
      },
    }),
  );
}

export function updateUserPrimaryTenant(username: string, tenantId: number) {
  execPgSQL(
    `UPDATE sys_user SET tenant_id = ${tenantId} WHERE username = '${pgEscapeLiteral(username)}';`,
  );
}

export async function deleteTenant(api: APIRequestContext, id: number) {
  await api.delete(`platform/tenants/${id}`).catch(() => {});
}

export async function addTenantMember(
  _api: APIRequestContext,
  payload: { tenantId: number; userId: number },
) {
  execPgSQL(`
    INSERT INTO plugin_multi_tenant_user_membership ("user_id", "tenant_id", "status", "created_by", "updated_by")
    VALUES (${payload.userId}, ${payload.tenantId}, 1, 0, 0)
    ON CONFLICT ("user_id", "tenant_id") DO UPDATE SET
      "status" = 1,
      "deleted_at" = NULL,
      "updated_at" = NOW();
  `);
  return {
    id: Number(
      queryPgScalar(`
        SELECT id
        FROM plugin_multi_tenant_user_membership
        WHERE user_id = ${payload.userId}
          AND tenant_id = ${payload.tenantId}
        LIMIT 1;
      `),
    ),
  };
}

export async function grantTenantPermissions(
  api: APIRequestContext,
  payload: {
    roleKey: string;
    roleName: string;
    tenantId: number;
    userId: number;
    permissions: string[];
  },
): Promise<TenantUserGrant> {
  await getMenuIdsByPerms(api, payload.permissions);
  const roleId = Number(
    queryPgScalar(`
      INSERT INTO sys_role (name, key, sort, data_scope, status, remark, tenant_id, created_at, updated_at)
      VALUES (
        '${pgEscapeLiteral(payload.roleName)}',
        '${pgEscapeLiteral(payload.roleKey)}',
        1,
        2,
        1,
        'Multi-tenant E2E tenant role',
        ${payload.tenantId},
        NOW(),
        NOW()
      )
      RETURNING id;
    `),
  );
  execPgSQL(
    `
      INSERT INTO sys_role_menu (role_id, menu_id, tenant_id)
      SELECT ${roleId}, id, ${payload.tenantId}
      FROM sys_menu
      WHERE perms IN (${payload.permissions
        .map((permission) => `'${pgEscapeLiteral(permission)}'`)
        .join(", ")})
      ON CONFLICT DO NOTHING;

      INSERT INTO sys_user_role (user_id, role_id, tenant_id)
      VALUES (${payload.userId}, ${roleId}, ${payload.tenantId})
      ON CONFLICT DO NOTHING;
    `,
  );
  return { roleId, tenantId: payload.tenantId };
}

export function revokeTenantPermissionGrants(grants: TenantUserGrant[]) {
  const roleIds = grants
    .map((grant) => grant.roleId)
    .filter((roleId) => roleId > 0);
  if (roleIds.length === 0) {
    return;
  }
  const idList = roleIds.join(", ");
  execPgSQL(`
    DELETE FROM sys_role_menu WHERE role_id IN (${idList});
    DELETE FROM sys_user_role WHERE role_id IN (${idList});
    DELETE FROM sys_role WHERE id IN (${idList});
  `);
}

export async function removeTenantMember(_api: APIRequestContext, id: number) {
  if (id <= 0) {
    return;
  }
  execPgSQL(
    `DELETE FROM plugin_multi_tenant_user_membership WHERE id = ${id};`,
  );
}

export function listTenantMembers(_api: APIRequestContext, tenantId: number) {
  const list = queryPgRows(`
    SELECT m.id || '|' || m.tenant_id || '|' || m.user_id || '|' || u.username || '|' || m.status
    FROM plugin_multi_tenant_user_membership m
    JOIN sys_user u ON u.id = m.user_id
    WHERE m.tenant_id = ${tenantId}
      AND m.deleted_at IS NULL
    ORDER BY m.id;
  `).map((row) => {
    const [id, rowTenantId, userId, username, status] = row.split("|");
    return {
      id: Number(id),
      tenantId: Number(rowTenantId),
      userId: Number(userId),
      username,
      status: Number(status),
    };
  });
  return {
    list,
    total: list.length,
  };
}

export async function selectTenant(preToken: string, tenantId: number) {
  const api = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const response = await api.post("auth/select-tenant", {
    data: { preToken, tenantId },
  });
  const data = await expectSuccess<{ accessToken: string }>(response);
  await api.dispose();
  return data.accessToken;
}

export async function createTenantApiContext(accessToken: string) {
  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

export async function switchTenant(
  api: APIRequestContext,
  tenantId: number,
): Promise<APIResponse> {
  return api.post("auth/switch-tenant", {
    data: { tenantId },
  });
}
