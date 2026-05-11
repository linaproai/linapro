import { test, expect } from '@host-tests/fixtures/multi-tenant';
import { scenarioTC0204 } from '@host-tests/support/multi-tenant-scenarios';

test.describe('TC-204 override 解析器', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-204a: platform override creates impersonation and ordinary users cannot override tenant context', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0204();
  });
});
