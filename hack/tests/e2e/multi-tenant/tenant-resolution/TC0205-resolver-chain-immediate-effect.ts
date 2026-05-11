import { test, expect } from '../../../fixtures/multi-tenant';
import { scenarioTC0205 } from '../../../support/multi-tenant-scenarios';

test.describe('TC-205 解析链固定策略', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-205a: built-in resolver policy no-op write leaves policy unchanged', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0205();
  });
});
