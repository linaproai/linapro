import { test, expect } from '@host-tests/fixtures/multi-tenant';
import { scenarioTC0181 } from '@host-tests/support/multi-tenant-scenarios';

test.describe('TC-181 租户不暴露归档生命周期', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-181a: archived status transitions are rejected', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0181();
  });
});
