import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0208 } from '../../support/multi-tenant-scenarios';

test.describe('TC-208 tenant-scoped 插件启停', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-208a: tenant plugin API toggles tenant-scoped enablement rows', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0208();
  });
});
