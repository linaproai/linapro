import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0218 } from '../../support/multi-tenant-scenarios';

test.describe('TC-218 解析策略无广播', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-218a: removed resolver policy API does not create shared revision', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0218();
  });
});
