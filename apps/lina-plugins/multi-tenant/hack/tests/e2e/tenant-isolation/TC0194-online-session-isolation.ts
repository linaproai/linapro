import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0194 } from '../../support/multi-tenant-scenarios';

test.describe('TC-194 在线会话跨租户隔离', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-194a: online session revocation targets tenant-token pairs', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0194();
  });
});
