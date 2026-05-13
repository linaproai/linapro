import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0195 } from '../../support/multi-tenant-scenarios';

test.describe('TC-195 审计日志隔离与 impersonation 标记', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-195a: login and operation logs keep tenant and impersonation fields', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0195();
  });
});
