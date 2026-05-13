import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0213 } from '../../support/multi-tenant-scenarios';

test.describe('TC-213 tenant provisioning policy', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-213a: new tenant receives platform-managed tenant-scoped plugin enablement', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0213();
  });
});
