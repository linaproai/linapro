import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0186 } from '../../support/multi-tenant-scenarios';

test.describe('TC-186 平台管理员 impersonation', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-186a: platform impersonation starts and then revokes its online session', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0186();
  });
});
