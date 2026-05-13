import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0200 } from '../../support/multi-tenant-scenarios';

test.describe('TC-200 jwt 解析器', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-200a: JWT tenant claim authorizes tenant member access', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0200();
  });
});
