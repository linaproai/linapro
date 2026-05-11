import { test, expect } from '@host-tests/fixtures/auth';
import { MainLayout } from '@host-tests/pages/MainLayout';
import { UserPage } from '@host-tests/pages/UserPage';

test.describe("TC-229 多租户插件未安装时 UI 退化", () => {
  test("TC-229a: 租户状态缓存残留时顶部切换器隐藏且用户管理不请求租户插件接口", async ({
    adminPage,
  }) => {
    await adminPage.route("**/api/v1/plugins/dynamic**", async (route) => {
      await route.fulfill({
        contentType: "application/json",
        json: {
          code: 0,
          data: {
            list: [
              {
                enabled: 0,
                generation: 1,
                id: "multi-tenant",
                installed: 0,
                statusKey: "sys_plugin.status:multi-tenant",
                version: "v0.1.0",
              },
            ],
          },
          message: "ok",
        },
      });
    });

    const platformTenantRequests: string[] = [];
    await adminPage.route("**/api/v1/platform/tenants**", async (route) => {
      platformTenantRequests.push(route.request().url());
      await route.fulfill({
        contentType: "application/json",
        status: 404,
        json: {
          code: 404,
          message: "未找到, 请求的资源不存在。",
        },
      });
    });

    await adminPage.evaluate(() => {
      localStorage.setItem(
        "linapro:tenant-state",
        JSON.stringify({
          currentTenant: {
            code: "stale",
            id: 99,
            name: "Stale Tenant",
            status: "active",
          },
          enabled: true,
          impersonation: { active: false },
          tenants: [
            {
              code: "stale",
              id: 99,
              name: "Stale Tenant",
              status: "active",
            },
          ],
        }),
      );
    });

    await adminPage.reload({ waitUntil: "domcontentloaded" });
    const layout = new MainLayout(adminPage);
    await expect(layout.tenantSwitcher).toHaveCount(0);

    const userPage = new UserPage(adminPage);
    await userPage.goto();
    await expect(userPage.tenantFilter).toHaveCount(0);
    await expect(userPage.tenantMembershipHeader).toHaveCount(0);
    expect(platformTenantRequests).toEqual([]);
  });
});
