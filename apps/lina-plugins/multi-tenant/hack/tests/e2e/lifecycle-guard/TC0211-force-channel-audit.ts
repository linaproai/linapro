import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0211 } from '../../support/multi-tenant-scenarios';

test.describe('TC-211 force 通道审计', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-211a: force uninstall request stays governed and leaves plugin installed', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0211();
  });
});
