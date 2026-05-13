import { createAdminApiContext } from '@host-tests/support/api/job';
import {
  execPgSQL,
  pgEscapeLiteral,
  queryPgScalar,
} from '@host-tests/support/postgres';
import {
  createTenant,
  deleteTenant,
  ensureMultiTenantPluginEnabled,
  expect,
  test,
} from '../../support/multi-tenant';

const pluginId = 'monitor-loginlog';
const enablementKey = '__tenant_enabled__';

function scalarNumber(sql: string) {
  return Number(queryPgScalar(sql));
}

function pluginMode() {
  return queryPgScalar(
    `SELECT install_mode FROM sys_plugin WHERE plugin_id = '${pgEscapeLiteral(pluginId)}';`,
  );
}

function tenantStateCount(tenantId: number) {
  return scalarNumber(`
    SELECT COUNT(1)
    FROM sys_plugin_state
    WHERE plugin_id = '${pgEscapeLiteral(pluginId)}'
      AND tenant_id = ${tenantId}
      AND state_key = '${enablementKey}';
  `);
}

function enabledTenantStateCount(tenantId: number) {
  return scalarNumber(`
    SELECT COUNT(1)
    FROM sys_plugin_state
    WHERE plugin_id = '${pgEscapeLiteral(pluginId)}'
      AND tenant_id = ${tenantId}
      AND state_key = '${enablementKey}'
      AND enabled = TRUE;
  `);
}

function migrateGlobalToTenantScoped(plugin: string) {
  execPgSQL(`
    UPDATE sys_plugin
    SET install_mode = 'tenant_scoped',
        updated_at = NOW()
    WHERE plugin_id = '${pgEscapeLiteral(plugin)}'
      AND scope_nature = 'tenant_aware';

    INSERT INTO sys_plugin_state (plugin_id, tenant_id, state_key, state_value, enabled, created_at, updated_at)
    SELECT '${pgEscapeLiteral(plugin)}', t.id, '${enablementKey}', 'enabled', TRUE, NOW(), NOW()
    FROM plugin_multi_tenant_tenant t
    WHERE t.status = 'active'
      AND t.deleted_at IS NULL
    ON CONFLICT DO NOTHING;
  `);
}

function migrateTenantScopedToGlobal(plugin: string) {
  execPgSQL(`
    UPDATE sys_plugin
    SET install_mode = 'global',
        status = 1,
        updated_at = NOW()
    WHERE plugin_id = '${pgEscapeLiteral(plugin)}'
      AND scope_nature = 'tenant_aware';

    DELETE FROM sys_plugin_state
    WHERE plugin_id = '${pgEscapeLiteral(plugin)}'
      AND tenant_id > 0
      AND state_key = '${enablementKey}';

    INSERT INTO sys_plugin_state (plugin_id, tenant_id, state_key, state_value, enabled, created_at, updated_at)
    VALUES ('${pgEscapeLiteral(plugin)}', 0, '${enablementKey}', 'enabled', TRUE, NOW(), NOW())
    ON CONFLICT DO NOTHING;
  `);
}

test.describe('TC-209 install_mode 切换', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  let tenantId = 0;
  let previousMode = '';

  test.beforeAll(async () => {
    const adminApi = await createAdminApiContext();
    await ensureMultiTenantPluginEnabled(adminApi);
    const suffix = Date.now().toString();
    tenantId = (
      await createTenant(adminApi, {
        code: `tc209-${suffix}`,
        name: `TC209 Tenant ${suffix}`,
      })
    ).id;
    previousMode = pluginMode();
    await adminApi.dispose();
  });

  test.afterAll(async () => {
    execPgSQL(`
      UPDATE sys_plugin
      SET install_mode = '${pgEscapeLiteral(previousMode || 'tenant_scoped')}',
          updated_at = NOW()
      WHERE plugin_id = '${pgEscapeLiteral(pluginId)}';
      DELETE FROM sys_plugin_state
      WHERE plugin_id = '${pgEscapeLiteral(pluginId)}'
        AND tenant_id = ${tenantId}
        AND state_key = '${enablementKey}';
    `);
    if (tenantId > 0) {
      const adminApi = await createAdminApiContext();
      await deleteTenant(adminApi, tenantId);
      await adminApi.dispose();
    }
  });

  test('TC-209a: global to tenant_scoped creates active tenant enablement rows and tenant_scoped to global collapses state', async ({
    multiTenantMode,
  }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');

    migrateGlobalToTenantScoped(pluginId);
    expect(pluginMode()).toBe('tenant_scoped');
    expect(tenantStateCount(tenantId)).toBe(1);
    expect(enabledTenantStateCount(tenantId)).toBe(1);

    migrateTenantScopedToGlobal(pluginId);
    expect(pluginMode()).toBe('global');
    expect(tenantStateCount(tenantId)).toBe(0);
    expect(
      enabledTenantStateCount(0),
      'global mode should keep a platform-level enabled state row',
    ).toBe(1);
  });
});
