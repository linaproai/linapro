import type { Page } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { waitForRouteReady } from "../../support/ui";

function unwrapApiData(payload: any) {
  if (payload && typeof payload === "object" && "data" in payload) {
    return payload.data;
  }
  return payload;
}

async function mockRuntimeI18nDisabled(page: Page) {
  await page.route("**/api/v1/i18n/runtime/locales**", async (route) => {
    const response = await route.fetch();
    const payload = await response.json();
    const data = unwrapApiData(payload);
    if (data && typeof data === "object") {
      data.enabled = false;
      data.locale = "zh-CN";
      data.items = [
        {
          direction: "ltr",
          isDefault: true,
          locale: "zh-CN",
          name: "简体中文",
          nativeName: "简体中文",
        },
      ];
    }

    await route.fulfill({
      json: payload,
      response,
    });
  });
}

test.describe("TC-130 runtime i18n disabled switch", () => {
  test("TC-130a: disabled runtime i18n hides language switcher and uses default locale", async ({
    adminContext,
  }) => {
    const page = await adminContext.newPage();
    await mockRuntimeI18nDisabled(page);

    await page.goto("/dashboard/analytics", { waitUntil: "domcontentloaded" });
    await waitForRouteReady(page, 15000);

    await expect
      .poll(async () => page.locator("html").getAttribute("lang"))
      .toBe("zh-CN");
    await expect(page.getByTestId("language-toggle-trigger")).toHaveCount(0);

    await page.close();
  });
});
