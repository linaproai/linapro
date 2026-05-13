import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0203 } from '../../support/multi-tenant-scenarios';

test.describe('TC-203 固定 prompt 歧义策略', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-203a: resolver policy stays code-owned and rejects reject mode', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0203();
  });
});
