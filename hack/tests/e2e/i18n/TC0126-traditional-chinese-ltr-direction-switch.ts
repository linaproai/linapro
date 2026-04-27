import type { Page } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";

async function expectHtmlDirection(page: Page, lang: string) {
  await expect
    .poll(async () => page.locator("html").getAttribute("lang"))
    .toBe(lang);
  await expect
    .poll(async () => page.locator("html").getAttribute("dir"))
    .toBe("ltr");
  await expect(page.getByTestId("app-direction-root")).toHaveAttribute(
    "data-app-direction",
    "ltr",
  );
}

test.describe("TC0126 固定 LTR 方向切换", () => {
  test("TC-126a: html 与 Ant Design Vue 方向在多语言切换时保持 LTR", async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage("繁體中文");
    await expectHtmlDirection(adminPage, "zh-TW");

    await mainLayout.switchLanguage("English");
    await expectHtmlDirection(adminPage, "en-US");

    await mainLayout.switchLanguage("简体中文");
    await expectHtmlDirection(adminPage, "zh-CN");
  });
});
