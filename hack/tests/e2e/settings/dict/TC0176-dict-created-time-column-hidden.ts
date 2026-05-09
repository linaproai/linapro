import { test, expect } from "../../../fixtures/auth";
import { DictPage } from "../../../pages/DictPage";

test.describe("TC-176 字典管理列表列展示", () => {
  test("TC-176a: 字典类型和字典数据列表不展示创建时间列", async ({
    adminPage,
  }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await expect(
      dictPage.typeHeader(/字典名称|Dictionary Name/i).first(),
    ).toBeVisible();
    await expect(dictPage.typeHeader(/创建时间|Created At/i)).toHaveCount(0);

    await dictPage.clickTypeRow("sys_normal_disable");

    await expect(
      dictPage.dataHeader(/字典标签|Dictionary Label/i).first(),
    ).toBeVisible();
    await expect(dictPage.dataHeader(/创建时间|Created At/i)).toHaveCount(0);
  });
});
