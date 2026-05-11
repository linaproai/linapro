import { test, expect } from "../../../fixtures/auth";
import { UserPage } from "../../../pages/UserPage";
import { waitForDialogReady } from "../../../support/ui";

test.describe("TC-226 用户管理账号字段展示", () => {
  test("TC-226a: 列表和新增抽屉使用账号文案", async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    const header = (title: string) =>
      adminPage.locator(".vxe-header--column.fixed--visible", {
        hasText: title,
      });
    const expectHeaderMinWidth = async (title: string, minWidth: number) => {
      const columnHeader = header(title).first();
      await expect(columnHeader).toBeVisible();
      const width = await columnHeader.evaluate((element) => element.clientWidth);
      expect(width).toBeGreaterThanOrEqual(minWidth);
    };

    await expectHeaderMinWidth("账号", 180);
    await expectHeaderMinWidth("昵称", 160);
    await expectHeaderMinWidth("角色", 220);
    await expectHeaderMinWidth("邮箱", 240);
    await expect(
      adminPage.locator(".vxe-header--column", { hasText: "名称" }),
    ).toHaveCount(0);

    await adminPage.getByRole("button", { name: /新\s*增/ }).click();
    const drawer = await waitForDialogReady(
      adminPage
        .locator('[role="dialog"]')
        .filter({ has: adminPage.getByPlaceholder("请输入账号") })
        .last(),
    );

    await expect(drawer.getByText(/账号/).first()).toBeVisible();
    await expect(drawer.getByPlaceholder("请输入账号")).toBeVisible();
    await expect(drawer.getByText(/用户名/)).toHaveCount(0);
  });
});
