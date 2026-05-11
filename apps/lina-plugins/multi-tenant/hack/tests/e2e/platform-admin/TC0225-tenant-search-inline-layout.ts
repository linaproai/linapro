import { test } from '@host-tests/fixtures/auth';
import { MultiTenantPage } from '@host-tests/pages/MultiTenantPage';

test.describe('TC-225 租户管理搜索区布局', () => {
  test('TC-225a: desktop search fields and action buttons stay on one row', async ({
    page,
  }) => {
    const multiTenantPage = new MultiTenantPage(page);

    await page.setViewportSize({ width: 1440, height: 900 });
    await multiTenantPage.gotoPlatformTenants();
    await multiTenantPage.expectTenantSearchInlineLayout();
  });
});
