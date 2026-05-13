import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0198 } from '../../support/multi-tenant-scenarios';

test.describe('TC-198 header 解析器', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-198a: header hint is configured while formal JWT remains authoritative', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0198();
  });
});
