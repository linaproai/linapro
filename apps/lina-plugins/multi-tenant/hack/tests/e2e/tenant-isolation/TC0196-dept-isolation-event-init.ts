import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0196 } from '../../support/multi-tenant-scenarios';

test.describe('TC-196 部门跨租户隔离', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-196a: tenant department rows use tenant scope without event outbox', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0196();
  });
});
