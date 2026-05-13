import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0201 } from '../../support/multi-tenant-scenarios';

test.describe('TC-201 session 解析器', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-201a: switch flow persists tenant session and revokes the old one', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0201();
  });
});
