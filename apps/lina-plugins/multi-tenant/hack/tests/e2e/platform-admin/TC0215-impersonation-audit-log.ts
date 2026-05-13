import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0215 } from '../../support/multi-tenant-scenarios';

test.describe('TC-215 impersonation 审计日志', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-215a: impersonation writes dual-track audit fields', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0215();
  });
});
