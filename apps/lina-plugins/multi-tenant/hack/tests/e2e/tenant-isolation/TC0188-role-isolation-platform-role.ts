import { test, expect } from '@host-tests/fixtures/multi-tenant';
import { scenarioTC0188 } from '@host-tests/support/multi-tenant-scenarios';

test.describe('TC-188 角色跨租户隔离', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-188a: tenant roles and platform roles persist in disjoint tenant buckets', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0188();
  });
});
