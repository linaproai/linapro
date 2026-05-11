import { test, expect } from '@host-tests/fixtures/multi-tenant';
import { scenarioTC0215 } from '@host-tests/support/multi-tenant-scenarios';

test.describe('TC-215 impersonation 审计日志', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-215a: impersonation writes dual-track audit fields', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0215();
  });
});
