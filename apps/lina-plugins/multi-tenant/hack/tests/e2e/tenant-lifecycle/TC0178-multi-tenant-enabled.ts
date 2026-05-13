import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0178 } from '../../support/multi-tenant-scenarios';

test.describe('TC-178 多租户启用', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-178a: multi-tenant plugin is installed and exposes real tenant APIs', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0178();
  });
});
