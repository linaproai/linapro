import {
  createAdminApiContext,
  createUser,
  deleteUser,
  expectSuccess,
} from '@host-tests/support/api/job';
import { execPgSQL } from '@host-tests/support/postgres';
import {
  addTenantMember,
  createTenant,
  createTenantApiContext,
  deleteTenant,
  ensureMultiTenantPluginEnabled,
  expect,
  grantTenantPermissions,
  listTenantMembers,
  loginRaw,
  removeTenantMember,
  revokeTenantPermissionGrants,
  selectTenant,
  test,
  type TenantUserGrant,
  updateUserPrimaryTenant,
} from '../../support/multi-tenant';

type APIRequestContext = Awaited<ReturnType<typeof createAdminApiContext>>;

const password = "test123456";

test.describe("TC-187 用户跨租户隔离", () => {
  test.use({ multiTenantMode: "multi-tenant-enabled" });

  let adminApi: APIRequestContext;
  let tenantApi: APIRequestContext | undefined;
  let tenantAId = 0;
  let tenantBId = 0;
  let userAId = 0;
  let userBId = 0;
  let memberAId = 0;
  let memberBId = 0;
  let usernameA = "";
  let usernameB = "";
  let grants: TenantUserGrant[] = [];

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await ensureMultiTenantPluginEnabled(adminApi);

    const suffix = Date.now().toString();
    tenantAId = (
      await createTenant(adminApi, {
        code: `tc187-a-${suffix}`,
        name: `TC187 Tenant A ${suffix}`,
      })
    ).id;
    tenantBId = (
      await createTenant(adminApi, {
        code: `tc187-b-${suffix}`,
        name: `TC187 Tenant B ${suffix}`,
      })
    ).id;

    usernameA = `tc187_a_${suffix}`;
    usernameB = `tc187_b_${suffix}`;
    userAId = (
      await createUser(adminApi, {
        username: usernameA,
        password,
        nickname: "TC187 Tenant A User",
      })
    ).id;
    userBId = (
      await createUser(adminApi, {
        username: usernameB,
        password,
        nickname: "TC187 Tenant B User",
      })
    ).id;

    updateUserPrimaryTenant(usernameA, tenantAId);
    updateUserPrimaryTenant(usernameB, tenantBId);

    memberAId = (
      await addTenantMember(adminApi, {
        tenantId: tenantAId,
        userId: userAId,
      })
    ).id;
    memberBId = (
      await addTenantMember(adminApi, {
        tenantId: tenantBId,
        userId: userBId,
      })
    ).id;
    grants = [
      await grantTenantPermissions(adminApi, {
        roleKey: `tc187_user_reader_${suffix}`,
        roleName: `TC187 User Reader ${suffix}`,
        tenantId: tenantAId,
        userId: userAId,
        permissions: ["system:user:list", "system:user:query"],
      }),
    ];
  });

  test.afterAll(async () => {
    await tenantApi?.dispose();
    revokeTenantPermissionGrants(grants);
    if (memberAId > 0) {
      await removeTenantMember(adminApi, memberAId);
    }
    if (memberBId > 0) {
      await removeTenantMember(adminApi, memberBId);
    }
    if (userAId > 0) {
      execPgSQL(`DELETE FROM sys_user_role WHERE user_id = ${userAId};`);
      await deleteUser(adminApi, userAId).catch(() => {});
    }
    if (userBId > 0) {
      execPgSQL(`DELETE FROM sys_user_role WHERE user_id = ${userBId};`);
      await deleteUser(adminApi, userBId).catch(() => {});
    }
    if (tenantAId > 0) {
      await deleteTenant(adminApi, tenantAId);
    }
    if (tenantBId > 0) {
      await deleteTenant(adminApi, tenantBId);
    }
    await adminApi?.dispose();
  });

  test("TC-187a: tenant A user visibility does not expose tenant B users", async ({
    multiTenantMode,
  }) => {
    expect(multiTenantMode).toBe("multi-tenant-enabled");

    const login = await loginRaw(usernameA, password);
    if (login.tenants) {
      expect(login.tenants.map((tenant) => tenant.id)).toEqual([tenantAId]);
    }

    const tenantAToken = login.accessToken
      ? login.accessToken
      : login.preToken
        ? await selectTenant(login.preToken, tenantAId)
        : "";
    expect(tenantAToken).toBeTruthy();
    const tenantContext = await createTenantApiContext(tenantAToken);
    tenantApi = tenantContext;

    const membersA = await listTenantMembers(tenantContext, tenantAId);
    expect(membersA.list.map((member) => member.userId)).toContain(userAId);
    expect(membersA.list.map((member) => member.userId)).not.toContain(userBId);

    const tenantBListViaTenantAToken = await expectSuccess<{
      list: Array<{ id: number }>;
      total: number;
    }>(await tenantContext.get(`user?pageNum=1&pageSize=100&tenantId=${tenantBId}`));
    expect(tenantBListViaTenantAToken.list.map((user) => user.id)).not.toContain(
      userBId,
    );
  });
});
