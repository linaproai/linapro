import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0197 } from '../../support/multi-tenant-scenarios';

test.describe('TC-197 岗位跨租户隔离', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-197a: same post code is allowed across different tenant buckets', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0197();
  });
});
