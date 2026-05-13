import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0181 } from '../../support/multi-tenant-scenarios';

test.describe('TC-181 租户不暴露归档生命周期', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-181a: archived status transitions are rejected', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0181();
  });
});
