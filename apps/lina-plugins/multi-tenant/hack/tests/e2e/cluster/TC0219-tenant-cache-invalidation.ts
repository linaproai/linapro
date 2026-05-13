import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0219 } from '../../support/multi-tenant-scenarios';

test.describe('TC-219 租户缓存失效隔离', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-219a: cache revision rows are isolated by tenant scope', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0219();
  });
});
