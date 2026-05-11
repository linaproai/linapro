import { test } from '@host-tests/fixtures/auth';
import { MultiTenantPage } from '@host-tests/pages/MultiTenantPage';

test.describe('TC-224 多租户上下文刷新默认页', () => {
  test('TC-224a: tenant context changes refresh permissions and enter default pages', async ({
    page,
  }) => {
    const multiTenantPage = new MultiTenantPage(page);

    await multiTenantPage.exerciseDirectImpersonationDefaultRoute();
    await multiTenantPage.exerciseTenantSwitch();
    await multiTenantPage.exerciseTenantUserSwitch();
  });
});
