import { test, expect } from '@host-tests/fixtures/multi-tenant';
import { scenarioTC0212 } from '@host-tests/support/multi-tenant-scenarios';

test.describe('TC-212 钩子 fail-safe', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-212a: tenant delete does not rely on lifecycle event outbox', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0212();
  });
});
