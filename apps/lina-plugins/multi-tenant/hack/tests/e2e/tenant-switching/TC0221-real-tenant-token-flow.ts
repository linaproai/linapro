import {
  createAdminApiContext,
  createUser,
  deleteUser,
  expectSuccess,
} from '@host-tests/support/api/job';
import {
  execPgSQL,
  pgEscapeLiteral,
  queryPgScalar,
} from '@host-tests/support/postgres';
import {
  addTenantMember,
  createTenant,
  createTenantApiContext,
  deleteTenant,
  ensureMultiTenantPluginEnabled,
  expect,
  listTenantMembers,
  loginRaw,
  removeTenantMember,
  selectTenant,
  switchTenant,
  test,
} from '../../support/multi-tenant';

type APIRequestContext = Awaited<ReturnType<typeof createAdminApiContext>>;

const password = "test123456";

async function getAdminRoleId(api: APIRequestContext) {
  const roles = await expectSuccess<{
    list: Array<{ id: number; key: string }>;
    total: number;
  }>(await api.get("role?page=1&size=100&key=admin"));
  const adminRole = roles.list.find((item) => item.key === "admin");
  expect(adminRole, "built-in admin role should exist").toBeTruthy();
  return adminRole!.id;
}

test.describe("TC-221 multi-tenant real token flow", () => {
  test.use({ multiTenantMode: "multi-tenant-enabled" });

  let adminApi: APIRequestContext;
  let tenantApi: APIRequestContext | undefined;
  let switchedApi: APIRequestContext | undefined;
  let tenantAId = 0;
  let tenantBId = 0;
  let memberAId = 0;
  let memberBId = 0;
  let tenantARoleId = 0;
  let tenantBRoleId = 0;
  let userId = 0;
  let username = "";

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await ensureMultiTenantPluginEnabled(adminApi);

    const suffix = Date.now().toString();
    const adminRoleId = await getAdminRoleId(adminApi);
    username = `e2e_mt_${suffix}`;
    userId = (
      await createUser(adminApi, {
        username,
        password,
        nickname: "E2E Multi Tenant",
        roleIds: [adminRoleId],
      })
    ).id;

    tenantAId = (
      await createTenant(adminApi, {
        code: `e2e-alpha-${suffix}`,
        name: `E2E Alpha ${suffix}`,
      })
    ).id;
    tenantBId = (
      await createTenant(adminApi, {
        code: `e2e-beta-${suffix}`,
        name: `E2E Beta ${suffix}`,
      })
    ).id;

    execPgSQL(
      `UPDATE sys_user SET tenant_id = ${tenantAId} WHERE username = '${pgEscapeLiteral(username)}';`,
    );

    tenantARoleId = Number(
      queryPgScalar(`
      INSERT INTO sys_role (name, key, sort, data_scope, status, remark, tenant_id, created_at, updated_at)
      VALUES ('E2E Tenant Admin A', 'e2e_mt_role_a_${suffix}', 1, 2, 1, 'TC0221 tenant role A', ${tenantAId}, NOW(), NOW())
      RETURNING id;
    `),
    );
    tenantBRoleId = Number(
      queryPgScalar(`
      INSERT INTO sys_role (name, key, sort, data_scope, status, remark, tenant_id, created_at, updated_at)
      VALUES ('E2E Tenant Admin B', 'e2e_mt_role_b_${suffix}', 1, 2, 1, 'TC0221 tenant role B', ${tenantBId}, NOW(), NOW())
      RETURNING id;
    `),
    );
    execPgSQL(
      [
        `INSERT INTO sys_role_menu (role_id, menu_id, tenant_id)
       SELECT ${tenantARoleId}, id, ${tenantAId}
       FROM sys_menu
       WHERE perms IN ('system:user:list', 'system:user:query')
       ON CONFLICT DO NOTHING;`,
        `INSERT INTO sys_role_menu (role_id, menu_id, tenant_id)
       SELECT ${tenantBRoleId}, id, ${tenantBId}
       FROM sys_menu
       WHERE perms IN ('system:user:list', 'system:user:query')
       ON CONFLICT DO NOTHING;`,
        `INSERT INTO sys_user_role (user_id, role_id, tenant_id) VALUES (${userId}, ${tenantARoleId}, ${tenantAId}) ON CONFLICT DO NOTHING;`,
        `INSERT INTO sys_user_role (user_id, role_id, tenant_id) VALUES (${userId}, ${tenantBRoleId}, ${tenantBId}) ON CONFLICT DO NOTHING;`,
      ].join("\n"),
    );

    memberAId = (
      await addTenantMember(adminApi, {
        tenantId: tenantAId,
        userId,
      })
    ).id;
    memberBId = (
      await addTenantMember(adminApi, {
        tenantId: tenantBId,
        userId,
      })
    ).id;
  });

  test.afterAll(async () => {
    await switchedApi?.dispose();
    await tenantApi?.dispose();
    if (memberAId > 0) {
      await removeTenantMember(adminApi, memberAId);
    }
    if (memberBId > 0) {
      await removeTenantMember(adminApi, memberBId);
    }
    if (userId > 0) {
      execPgSQL(`DELETE FROM sys_user_role WHERE user_id = ${userId};`);
      await deleteUser(adminApi, userId).catch(() => {});
    }
    if (tenantARoleId > 0 || tenantBRoleId > 0) {
      execPgSQL(`
        DELETE FROM sys_role_menu WHERE role_id IN (${tenantARoleId || -1}, ${tenantBRoleId || -1});
        DELETE FROM sys_role WHERE id IN (${tenantARoleId || -1}, ${tenantBRoleId || -1});
      `);
    }
    if (tenantAId > 0) {
      await deleteTenant(adminApi, tenantAId);
    }
    if (tenantBId > 0) {
      await deleteTenant(adminApi, tenantBId);
    }
    await adminApi?.dispose();
  });

  test("TC-221a~d: login tenant choices, tenant selection, token switch, and old-token rejection hit real APIs", async ({
    multiTenantMode,
  }) => {
    expect(multiTenantMode).toBe("multi-tenant-enabled");

    const tenantLogin = await loginRaw(username, password);
    expect(tenantLogin.accessToken ?? "").toBe("");
    expect(
      tenantLogin.preToken,
      "multi-tenant login should issue a preToken",
    ).toBeTruthy();
    expect(tenantLogin.tenants?.map((item) => item.id)).toEqual([
      tenantAId,
      tenantBId,
    ]);

    const tenantAToken = await selectTenant(tenantLogin.preToken!, tenantAId);
    const tenantContext = await createTenantApiContext(tenantAToken);
    tenantApi = tenantContext;

    await expectSuccess(await tenantContext.get("user/info"));

    const membersA = await listTenantMembers(tenantContext, tenantAId);
    expect(membersA.list.map((item) => item.userId)).toContain(userId);

    const switchResponse = await switchTenant(tenantContext, tenantBId);
    const switchPayload = await expectSuccess<{ accessToken: string }>(
      switchResponse,
    );
    expect(switchPayload.accessToken).toBeTruthy();
    expect(switchPayload.accessToken).not.toBe(tenantAToken);

    const revokedTokenResponse = await tenantContext.get("user/info");
    expect(revokedTokenResponse.status()).toBe(401);

    const switchedContext = await createTenantApiContext(switchPayload.accessToken);
    switchedApi = switchedContext;
    await expectSuccess(await switchedContext.get("user/info"));
    const membersB = await listTenantMembers(switchedContext, tenantBId);
    expect(membersB.list.map((item) => item.userId)).toContain(userId);
  });
});
