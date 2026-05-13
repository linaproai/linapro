import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0190 } from '../../support/multi-tenant-scenarios';

test.describe('TC-190 配置跨租户隔离', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-190a: tenant config override and platform fallback rows stay isolated', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0190();
  });
});
