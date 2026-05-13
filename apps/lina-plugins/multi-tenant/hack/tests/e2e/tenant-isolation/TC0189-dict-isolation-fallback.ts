import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0189 } from '../../support/multi-tenant-scenarios';

test.describe('TC-189 字典跨租户隔离', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-189a: tenant dictionary override and platform fallback rows stay isolated', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0189();
  });
});
