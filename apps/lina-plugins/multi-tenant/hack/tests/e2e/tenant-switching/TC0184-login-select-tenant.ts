import { test, expect } from '../../support/multi-tenant';
import { MultiTenantPage } from '@host-tests/pages/MultiTenantPage';
import { scenarioTC0184 } from '../../support/multi-tenant-scenarios';

test.describe('TC-184 登录选择租户', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-184a: multi-membership login returns preToken and select-tenant issues a JWT', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0184();
  });

  test('TC-184b: login page swaps credentials for a tenant dropdown after retoken login', async ({
    page,
  }) => {
    const multiTenantPage = new MultiTenantPage(page);

    await multiTenantPage.exerciseTenantSelectionLogin();
  });
});
