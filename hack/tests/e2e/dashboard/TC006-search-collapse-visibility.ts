import { mkdir } from "node:fs/promises";
import path from "node:path";

import type { Page } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { LayoutAuditPage } from "../../pages/LayoutAuditPage";

test.describe("TC-6 Search collapse visibility", () => {
  test.beforeEach(async ({ adminPage }) => {
    await adminPage.setViewportSize({ width: 1440, height: 900 });
  });

  test("TC-6a: single-row search does not show collapse toggle", async ({
    adminPage,
  }) => {
    const layout = new LayoutAuditPage(adminPage);

    await layout.goto("/system/job-group", {
      tableSelector: '[data-testid="job-group-page"]',
    });

    await expect(layout.searchFormLabel("分组编码")).toBeVisible();
    await expect(layout.searchFormLabel("分组名称")).toBeVisible();
    await expect(layout.searchResetButton()).toContainText(/重\s*置/u);
    await expect(layout.searchSubmitButton()).toContainText(/搜\s*索/u);
    await layout.expectSearchCollapseHidden();
    await layout.expectSearchControlsOnOneRow(["分组编码", "分组名称"]);
    await expect(adminPage.locator("body")).not.toContainText(
      /pages\.system\.jobGroup|common\.(reset|search)/u,
    );
    await captureEvidence(adminPage, "search-collapse-single-row");
  });

  test("TC-6b: multi-row search keeps collapsible conditions accessible", async ({
    adminPage,
  }) => {
    const layout = new LayoutAuditPage(adminPage);

    await layout.goto("/system/plugin", {
      tableSelector: ".vxe-table",
    });

    await expect(layout.searchFormLabel("插件标识")).toBeVisible();
    await expect(layout.searchFormLabel("插件名称")).toBeVisible();
    await expect(layout.searchFormLabel("插件类型")).toBeVisible();
    await expect(layout.searchFormLabel("安装状态")).toBeVisible();
    await expect(layout.searchFormLabel("状态")).toBeVisible();
    await layout.expectSearchCollapseVisible();
    await expect(layout.searchCollapseToggle()).toContainText("收起");
    await expect(adminPage.locator("body")).not.toContainText(
      /pages\.system\.plugin|common\.(reset|search)/u,
    );
    await captureEvidence(adminPage, "search-collapse-multi-row-expanded");

    await layout.toggleSearchCollapse();

    await expect(layout.searchCollapseToggle()).toContainText("展开");
    await layout.expectSearchLabelHidden("状态");
    await layout.expectSearchLabelVisible("插件标识");
    await layout.expectSearchLabelVisible("插件名称");
    await layout.expectSearchLabelVisible("插件类型");
    await captureEvidence(adminPage, "search-collapse-multi-row-collapsed");

    await layout.toggleSearchCollapse();

    await expect(layout.searchCollapseToggle()).toContainText("收起");
    await layout.expectSearchLabelVisible("状态");
    await captureEvidence(adminPage, "search-collapse-multi-row-reexpanded");
  });
});

async function captureEvidence(page: Page, name: string) {
  await page.waitForTimeout(300);
  const now = new Date();
  const day = new Intl.DateTimeFormat("en-CA", {
    day: "2-digit",
    month: "2-digit",
    timeZone: "Asia/Shanghai",
    year: "numeric",
  })
    .format(now)
    .replaceAll("-", "");
  const time = new Intl.DateTimeFormat("en-GB", {
    hour: "2-digit",
    hour12: false,
    minute: "2-digit",
    second: "2-digit",
    timeZone: "Asia/Shanghai",
  })
    .format(now)
    .replaceAll(":", "");
  const dir = path.resolve(process.cwd(), "..", "..", "temp", day);
  await mkdir(dir, { recursive: true });
  await page.screenshot({
    fullPage: true,
    path: path.join(dir, `${time}-${name}.png`),
  });
}
