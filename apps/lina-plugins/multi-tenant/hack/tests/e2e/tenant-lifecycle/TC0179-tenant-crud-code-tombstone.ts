import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0179 } from '../../support/multi-tenant-scenarios';

test.describe('TC-179 平台管理员创建租户', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-179a: tenant CRUD validates code uniqueness and tombstones', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0179();
  });
});
