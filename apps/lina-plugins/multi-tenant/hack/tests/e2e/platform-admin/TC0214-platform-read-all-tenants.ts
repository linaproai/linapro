import { test, expect } from '@host-tests/fixtures/multi-tenant';
import { scenarioTC0214 } from '@host-tests/support/multi-tenant-scenarios';

test.describe('TC-214 平台管理员跨租户读', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-214a: platform tenant list can read tenants across scopes', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0214();
  });
});
