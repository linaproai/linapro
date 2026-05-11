import { test, expect } from '../../../fixtures/multi-tenant';
import { scenarioTC0206 } from '../../../support/multi-tenant-scenarios';

test.describe('TC-206 tenant-aware 插件 install_mode', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-206a: tenant-aware plugin registry exposes controllable install mode', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0206();
  });
});
