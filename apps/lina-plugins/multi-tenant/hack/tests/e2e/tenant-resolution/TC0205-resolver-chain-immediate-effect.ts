import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0205 } from '../../support/multi-tenant-scenarios';

test.describe('TC-205 解析链固定策略', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-205a: removed resolver policy API leaves code-owned policy unchanged', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0205();
  });
});
