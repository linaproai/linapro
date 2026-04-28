import { MainLayout } from "../../pages/MainLayout";
import { config } from "../../fixtures/config";
import { test, expect } from "../../fixtures/auth";
import { waitForRouteReady } from "../../support/ui";

test.describe("TC-117 默认品牌 Logo 资源", () => {
  test("TC-117a: 公共前端配置返回本地图标 Logo", async ({ request }) => {
    const response = await request.get("/api/v1/config/public/frontend");
    expect(response.ok()).toBeTruthy();

    const payload = await response.json();
    expect(payload.code).toBe(0);
    expect(payload.data.app.logo).toBe("/logo.png");
    expect(payload.data.app.logoDark).toBe("/logo.png");
  });

  test("TC-117b: 登录页使用图标 Logo 且渲染应用名文本", async ({
    loginPage,
  }) => {
    await loginPage.goto();

    await expect(loginPage.brandLogoImage).toBeVisible();
    const logo = await loginPage.getBrandLogoInfo();
    expect(logo.src).toBe("/logo.png");
    expect(logo.currentSrc).toContain("/logo.png");
    expect(logo.naturalWidth).toBe(405);
    expect(logo.naturalHeight).toBe(405);
    expect(logo.width).toBeLessThan(60);
    expect(logo.parentText).toContain("LinaPro");
  });

  test("TC-117c: 登录后工作台左上角使用图标 Logo 且渲染应用名文本", async ({
    loginPage,
    page,
  }) => {
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    await waitForRouteReady(page);

    const mainLayout = new MainLayout(page);
    await expect(mainLayout.brandLogoImage).toBeVisible();
    const logo = await mainLayout.getBrandLogoInfo();
    expect(logo.src).toBe("/logo.png");
    expect(logo.currentSrc).toContain("/logo.png");
    expect(logo.naturalWidth).toBe(405);
    expect(logo.naturalHeight).toBe(405);
    expect(logo.width).toBeLessThan(60);
    expect(logo.parentText).toContain("LinaPro");
  });
});
