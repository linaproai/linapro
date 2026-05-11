import { test, expect } from '@host-tests/fixtures/multi-tenant';
import { scenarioTC0180 } from '@host-tests/support/multi-tenant-scenarios';

test.describe('TC-180 租户暂停恢复', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-180a: suspended tenant blocks login and resumed tenant allows it again', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0180();
  });
});
