import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0216 } from '../../support/multi-tenant-scenarios';

test.describe('TC-216 强制操作审计', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-216a: force lifecycle request is guarded and observable through stable state', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0216();
  });
});
