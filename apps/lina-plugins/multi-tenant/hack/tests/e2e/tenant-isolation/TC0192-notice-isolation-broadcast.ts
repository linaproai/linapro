import { test, expect } from '@host-tests/fixtures/multi-tenant';
import { scenarioTC0192 } from '@host-tests/support/multi-tenant-scenarios';

test.describe('TC-192 通知跨租户隔离', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-192a: tenant notices and platform broadcast messages persist separately', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0192();
  });
});
