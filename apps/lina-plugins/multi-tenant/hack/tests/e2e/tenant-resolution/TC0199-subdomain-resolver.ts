import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0199 } from '../../support/multi-tenant-scenarios';

test.describe('TC-199 subdomain 解析器', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-199a: subdomain root is fixed empty and reserved labels are enforced', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0199();
  });
});
