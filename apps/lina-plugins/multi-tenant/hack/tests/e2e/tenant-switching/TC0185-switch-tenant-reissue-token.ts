import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0185 } from '../../support/multi-tenant-scenarios';

test.describe('TC-185 切换租户重签 token', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-185a: switch-tenant reissues token and revokes the previous token', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0185();
  });
});
