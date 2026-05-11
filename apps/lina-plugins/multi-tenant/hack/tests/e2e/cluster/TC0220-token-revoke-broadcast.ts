import { test, expect } from '@host-tests/fixtures/multi-tenant';
import { scenarioTC0220 } from '@host-tests/support/multi-tenant-scenarios';

test.describe('TC-220 token revoke 广播', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-220a: old token is rejected after switch through shared revoke state', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0220();
  });
});
