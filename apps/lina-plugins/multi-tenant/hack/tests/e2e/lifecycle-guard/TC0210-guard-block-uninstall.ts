import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0210 } from '../../support/multi-tenant-scenarios';

test.describe('TC-210 LifecycleGuard 否决卸载', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-210a: lifecycle guard blocks multi-tenant uninstall', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0210();
  });
});
