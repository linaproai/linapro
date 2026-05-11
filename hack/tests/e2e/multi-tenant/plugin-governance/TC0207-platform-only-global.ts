import { test, expect } from '../../../fixtures/multi-tenant';
import { scenarioTC0207 } from '../../../support/multi-tenant-scenarios';

test.describe('TC-207 platform-only 插件强制 global', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-207a: platform-only plugin remains global and hidden from tenant plugin API', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0207();
  });
});
