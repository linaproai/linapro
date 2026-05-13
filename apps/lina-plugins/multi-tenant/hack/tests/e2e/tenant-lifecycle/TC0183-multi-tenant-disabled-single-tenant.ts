import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0183 } from '../../support/multi-tenant-scenarios';

test.describe('TC-183 多租户禁用退化', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-183a: multi-tenant lifecycle guard keeps the platform-only plugin installed', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0183();
  });
});
