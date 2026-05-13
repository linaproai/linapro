import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0217 } from '../../support/multi-tenant-scenarios';

test.describe('TC-217 平台与租户视图差异', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-217a: platform and tenant API menu views differ', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0217();
  });
});
