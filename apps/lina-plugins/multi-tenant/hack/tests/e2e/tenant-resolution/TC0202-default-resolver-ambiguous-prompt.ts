import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0202 } from '../../support/multi-tenant-scenarios';

test.describe('TC-202 default 解析器 ambiguous prompt', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-202a: ambiguous login returns preToken and tenant candidates', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0202();
  });
});
