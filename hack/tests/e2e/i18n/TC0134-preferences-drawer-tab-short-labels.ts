import type { Locator } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";

async function readTabMetrics(tab: Locator) {
  return await tab.evaluate((node) => {
    const element = node as HTMLElement;

    return {
      clientWidth: element.clientWidth,
      scrollWidth: element.scrollWidth,
      text: element.textContent?.trim() ?? "",
    };
  });
}

test.describe("TC-134 Preferences drawer short tab labels", () => {
  test.beforeEach(async ({ adminPage }) => {
    await adminPage.setViewportSize({ width: 1366, height: 900 });
  });

  test("TC-134a: English preferences tabs use compact labels without overflow", async ({
    mainLayout,
  }) => {
    await mainLayout.switchLanguage("English");
    await mainLayout.openPreferences();

    const expectedTabs = ["Theme", "Layout", "Shortcuts", "General"];
    for (const label of expectedTabs) {
      const tab = mainLayout.preferencesDrawer.getByRole("tab", {
        exact: true,
        name: label,
      });
      await expect(tab).toBeVisible();

      const metrics = await readTabMetrics(tab);
      expect(
        metrics.scrollWidth,
        `${metrics.text} tab text should fit within the button background`,
      ).toBeLessThanOrEqual(metrics.clientWidth + 1);
    }

    await expect(
      mainLayout.preferencesDrawer.getByRole("tab", {
        exact: true,
        name: "Appearance",
      }),
    ).toHaveCount(0);
    await expect(
      mainLayout.preferencesDrawer.getByRole("tab", {
        exact: true,
        name: "Shortcut Keys",
      }),
    ).toHaveCount(0);
  });
});
