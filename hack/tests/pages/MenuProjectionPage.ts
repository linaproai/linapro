import { mkdir, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { expect, type Page } from "@playwright/test";

import { workspacePath } from "../fixtures/config";
import { waitForBusyIndicatorsToClear, waitForRouteReady } from "../support/ui";
import { DashboardPage } from "./DashboardPage";
import { MainLayout } from "./MainLayout";

export const menuLayouts = [
  "sidebar-nav",
  "header-nav",
  "mixed-nav",
  "sidebar-mixed-nav",
  "header-mixed-nav",
  "header-sidebar-nav",
] as const;

export type MenuLayout = (typeof menuLayouts)[number];
type MenuLocale = "en-US" | "zh-CN";
type DashboardView = "analytics" | "workspace";

const layoutLabels: Record<MenuLocale, Record<MenuLayout, string>> = {
  "zh-CN": {
    "sidebar-nav": "垂直",
    "header-nav": "水平",
    "mixed-nav": "混合垂直",
    "sidebar-mixed-nav": "双列菜单",
    "header-mixed-nav": "混合双列",
    "header-sidebar-nav": "侧边导航",
  },
  "en-US": {
    "sidebar-nav": "Vertical",
    "header-nav": "Horizontal",
    "mixed-nav": "Mixed Menu",
    "sidebar-mixed-nav": "Two Column",
    "header-mixed-nav": "Header Two Column",
    "header-sidebar-nav": "Header Vertical",
  },
};

const menuLabels = {
  "zh-CN": { root: "工作台", analytics: "分析页", workspace: "工作台" },
  "en-US": {
    root: "Dashboard",
    analytics: "Analytics",
    workspace: "Workspace",
  },
} as const;

export interface MenuPreferencesSnapshot {
  collapsed: boolean;
  hidden: boolean;
  layout: MenuLayout;
  locale: MenuLocale;
  url: string;
}

export interface MenuFrameSample {
  clickToFirstFrameMs: number;
  direction: "hide" | "show";
  elapsedMs: number;
  frameCount: number;
  frameIntervalsMs: number[];
}

export class MenuProjectionPage {
  private readonly mainLayout: MainLayout;
  private readonly dashboard: DashboardPage;
  private dashboardGroupIndex: number | undefined;

  constructor(private readonly page: Page) {
    this.mainLayout = new MainLayout(page);
    this.dashboard = new DashboardPage(page);
  }

  private get verticalMenu() {
    return this.page.locator("aside .vben-menu.is-vertical").first();
  }

  private get horizontalMenu() {
    return this.page.locator("header .vben-menu.is-horizontal").first();
  }

  private get collapseToggle() {
    // 对应 layout-sidebar.vue 内 SidebarCollapseButton 的真实底部控件。
    return this.page
      .locator("aside > div > div.absolute.bottom-2.left-3.cursor-pointer")
      .first();
  }

  private get headerSidebarToggle() {
    // 对应 vben-layout.vue 的 header toggle-button 插槽，控制 sidebar.hidden。
    return this.page.locator("header > button.my-0.mr-1").first();
  }

  async isSidebarHidden() {
    return this.page
      .locator("aside")
      .first()
      .evaluate((aside) => aside.style.width === "0px");
  }

  async setSidebarHidden(hidden: boolean) {
    if ((await this.isSidebarHidden()) !== hidden) {
      await this.headerSidebarToggle.click();
    }
    await expect.poll(() => this.isSidebarHidden()).toBe(hidden);
  }

  async setViewport() {
    await this.page.setViewportSize({ width: 1600, height: 1000 });
  }

  private async locale(): Promise<MenuLocale> {
    const locale = await this.page.locator("html").getAttribute("lang");
    if (locale !== "zh-CN" && locale !== "en-US") {
      throw new Error(`菜单测试不支持当前语言：${locale}`);
    }
    return locale;
  }

  private layoutCard(layout: MenuLayout, locale: MenuLocale) {
    return this.mainLayout.preferencesDrawer
      .locator("div.cursor-pointer")
      .filter({
        has: this.page.getByText(layoutLabels[locale][layout], { exact: true }),
      });
  }

  private async closePreferences() {
    // 菜单弹层也响应 Escape；点击抽屉自身的 SheetClose 以明确关闭目标。
    await this.mainLayout.preferencesDrawer
      .locator('button[class~="ml-[2px]"]')
      .click();
    await expect(this.mainLayout.preferencesDrawer).toBeHidden();
  }

  async rememberPreferences(): Promise<MenuPreferencesSnapshot> {
    const locale = await this.locale();
    const url = this.page.url();
    const hidden = await this.isSidebarHidden();
    await this.mainLayout.openPreferencesTab(
      locale === "zh-CN" ? "布局" : "Layout",
    );
    let layout: MenuLayout | undefined;
    for (const candidate of menuLayouts) {
      if (
        await this.layoutCard(candidate, locale)
          .locator(".outline-box-active")
          .count()
      ) {
        layout = candidate;
        break;
      }
    }
    if (!layout) {
      throw new Error("初始布局不是本用例支持的菜单布局");
    }
    const collapseLabel = locale === "zh-CN" ? "折叠菜单" : "Collpase Menu";
    const collapsed = await this.mainLayout.preferencesDrawer
      .getByText(collapseLabel, { exact: true })
      .locator("..")
      .getByRole("switch")
      .getAttribute("aria-checked");
    expect(["true", "false"]).toContain(collapsed);
    await this.closePreferences();
    return { collapsed: collapsed === "true", hidden, layout, locale, url };
  }

  async restorePreferences(initial: MenuPreferencesSnapshot) {
    if (await this.mainLayout.preferencesDrawer.isVisible()) {
      await this.closePreferences();
    }
    await this.switchLanguage(initial.locale);
    await this.page.goto(initial.url, { waitUntil: "domcontentloaded" });
    await expect(this.page).toHaveURL(initial.url);
    await waitForRouteReady(this.page);
    await this.switchLayout("sidebar-nav");
    await this.setSidebarHidden(false);
    await this.setCollapsed(initial.collapsed);
    await this.setSidebarHidden(initial.hidden);
    await this.switchLayout(initial.layout);
  }

  async switchLanguage(locale: MenuLocale) {
    if ((await this.locale()) !== locale) {
      await this.mainLayout.switchLanguage(
        locale === "zh-CN" ? "简体中文" : "English",
      );
    }
    await expect(this.page.locator("html")).toHaveAttribute("lang", locale);
  }

  async switchLayout(layout: MenuLayout, capturePreferences = false) {
    const locale = await this.locale();
    await this.mainLayout.openPreferencesTab(
      locale === "zh-CN" ? "布局" : "Layout",
    );
    const card = this.layoutCard(layout, locale);
    await expect(card).toHaveCount(1);
    await card.click();
    await expect(card.locator(".outline-box")).toHaveClass(
      /outline-box-active/,
    );
    const preferencesScreenshot = capturePreferences
      ? await this.screenshot(`menu-projection-preferences-${layout}-${locale}`)
      : undefined;
    await this.closePreferences();
    await waitForRouteReady(this.page);
    if (
      layout === "header-nav" ||
      layout === "mixed-nav" ||
      layout === "header-mixed-nav"
    ) {
      await expect(this.horizontalMenu).toBeVisible();
    }
    if (layout === "sidebar-mixed-nav" || layout === "header-mixed-nav") {
      await expect(this.page.locator("aside .vben-normal-menu")).toBeVisible();
    } else if (layout !== "header-nav") {
      await expect(this.verticalMenu).toBeAttached();
    }
    return preferencesScreenshot;
  }

  async isCollapsed() {
    await expect(this.verticalMenu).toBeVisible();
    return this.verticalMenu.evaluate((menu) =>
      menu.classList.contains("is-collapse"),
    );
  }

  async setCollapsed(collapsed: boolean) {
    const wasCollapsed = await this.isCollapsed();
    if (wasCollapsed && !collapsed) {
      // 先完成当前悬浮菜单的关闭，避免其延迟 mouseleave 在展开后关闭分组。
      await this.page.mouse.move(800, 90);
      await expect(
        this.page.locator(".vben-menu__popup-container:visible"),
      ).toHaveCount(0);
    }
    if (wasCollapsed !== collapsed) {
      await this.collapseToggle.click();
    }
    await expect.poll(() => this.isCollapsed()).toBe(collapsed);
  }

  async topLevelNames() {
    await expect(this.verticalMenu).not.toHaveClass(/is-collapse/);
    const names = await this.verticalMenu
      .locator(
        ":scope > .vben-sub-menu > .vben-sub-menu-content .vben-sub-menu-content__title",
      )
      .allTextContents();
    return names.map((name) => name.trim());
  }

  private async openVerticalDashboard(locale: MenuLocale) {
    const groups = this.verticalMenu.locator(":scope > .vben-sub-menu");
    if (!(await this.isCollapsed())) {
      const names = await this.topLevelNames();
      this.dashboardGroupIndex = names.indexOf(menuLabels[locale].root);
      expect(
        this.dashboardGroupIndex,
        "工作台顶级菜单应存在",
      ).toBeGreaterThanOrEqual(0);
      const group = groups.nth(this.dashboardGroupIndex);
      if (
        !(await group.evaluate((node) => node.classList.contains("is-opened")))
      ) {
        await group.locator(":scope > .vben-sub-menu-content").click();
      }
      await expect(group).toHaveClass(/is-opened/);
      return group;
    }
    if (this.dashboardGroupIndex === undefined) {
      throw new Error("收起前必须先从展开菜单记录工作台分组位置");
    }
    const group = groups.nth(this.dashboardGroupIndex);
    await group.locator(".vben-sub-menu-content").first().hover();
    await expect(group).toHaveClass(/is-opened/);
    return group;
  }

  private async revealDashboard(layout: MenuLayout, locale: MenuLocale) {
    const rootLabel = menuLabels[locale].root;
    if (layout === "sidebar-nav" || layout === "header-sidebar-nav") {
      return this.openVerticalDashboard(locale);
    }
    if (layout === "header-nav") {
      const root = this.horizontalMenu
        .locator(":scope > .vben-sub-menu")
        .filter({ has: this.page.getByText(rootLabel, { exact: true }) });
      await expect(root).toHaveCount(1);
      await root.locator(".vben-sub-menu-content").first().hover();
      await expect(root).toHaveClass(/is-opened/);
      return root;
    }
    const roots =
      layout === "sidebar-mixed-nav"
        ? this.page.locator("aside .vben-normal-menu__item")
        : this.horizontalMenu.locator(":scope > .vben-menu-item");
    const root = roots.filter({
      has: this.page.getByText(rootLabel, { exact: true }),
    });
    await expect(root).toHaveCount(1);
    await root.click();
    await expect(root).toHaveClass(/is-active/);
    return root;
  }

  private dashboardLeaf(layout: MenuLayout, label: string) {
    const selector =
      layout === "header-mixed-nav"
        ? "aside .vben-normal-menu__item:visible"
        : layout === "mixed-nav" || layout === "sidebar-mixed-nav"
          ? "aside .vben-menu-item:visible"
          : ".vben-menu-item:visible";
    return this.page
      .locator(selector)
      .filter({ has: this.page.getByText(label, { exact: true }) });
  }

  async selectDashboard(
    layout: MenuLayout,
    view: DashboardView,
    locale: MenuLocale,
  ) {
    const root = await this.revealDashboard(layout, locale);
    const leaf = this.dashboardLeaf(layout, menuLabels[locale][view]);
    await expect(leaf).toHaveCount(1);
    await expect(leaf).toBeVisible();
    await expect(leaf).not.toHaveClass(/is-disabled/);
    await leaf.click();
    await expect(this.page).toHaveURL(
      new RegExp(`${workspacePath(`/dashboard/${view}`)}(?:[?#].*)?$`),
    );
    await this.waitForDashboard(view);
    // 水平和收起菜单选中后关闭弹层，重新悬浮以审查实际选中项。
    if (
      layout === "header-nav" ||
      ((layout === "sidebar-nav" || layout === "header-sidebar-nav") &&
        (await this.isCollapsed()))
    ) {
      await this.revealDashboard(layout, locale);
    }
    await expect(
      this.dashboardLeaf(layout, menuLabels[locale][view]),
    ).toHaveClass(/is-active/);
    await expect(root).toHaveClass(/is-active/);
    await expect(this.mainLayout.activeTabTitle()).toHaveText(
      menuLabels[locale][view],
    );
  }

  async expectDashboardLabels(layout: MenuLayout, locale: MenuLocale) {
    await this.revealDashboard(layout, locale);
    for (const view of ["analytics", "workspace"] as const) {
      await expect(
        this.dashboardLeaf(layout, menuLabels[locale][view]),
      ).toBeVisible();
    }
    const staleLabel = locale === "zh-CN" ? "Analytics" : "分析页";
    await expect(this.dashboardLeaf(layout, staleLabel)).toHaveCount(0);
    const visibleMenus = this.page.locator(
      ".vben-menu:visible, .vben-normal-menu:visible",
    );
    for (const menu of await visibleMenus.all()) {
      await expect(menu).not.toContainText(
        /menu\.[\w:.-]+\.title|page\.dashboard\./,
      );
    }
  }

  async waitForDashboard(view: DashboardView) {
    await waitForRouteReady(this.page);
    if (view === "analytics") {
      await expect(this.dashboard.analyticsPage).toBeVisible();
      await this.dashboard.waitForAnalyticsChartInstances(4);
      await this.dashboard.waitForAnalyticsChartsPainted(4);
      await waitForBusyIndicatorsToClear(this.dashboard.analyticsPage);
      // 已绘制不等于绘制完成；等待实际画布内容稳定后再截图和开始预热。
      let previous = "";
      let stableSince = Date.now();
      await expect
        .poll(
          async () => {
            const fingerprints = await this.dashboard.analyticsPage
              .locator("canvas")
              .evaluateAll((canvases) =>
                canvases.map((node) => {
                  const canvas = node as HTMLCanvasElement;
                  const context = canvas.getContext("2d");
                  if (!context) throw new Error("分析页图表缺少二维画布");
                  const pixels = context.getImageData(
                    0,
                    0,
                    canvas.width,
                    canvas.height,
                  ).data;
                  let fingerprint = 2166136261;
                  for (let index = 0; index < pixels.length; index += 16) {
                    const pixel =
                      pixels[index]! |
                      (pixels[index + 1]! << 8) |
                      (pixels[index + 2]! << 16) |
                      (pixels[index + 3]! << 24);
                    fingerprint = Math.imul(fingerprint ^ pixel, 16777619);
                  }
                  return `${canvas.width}x${canvas.height}:${fingerprint}`;
                }),
              );
            const current = fingerprints.join("|");
            if (current !== previous) {
              previous = current;
              stableSince = Date.now();
            }
            return Date.now() - stableSince >= 400;
          },
          { intervals: [200], timeout: 10_000 },
        )
        .toBe(true);
    } else {
      await expect(this.dashboard.workspacePage).toBeVisible();
      await expect(this.dashboard.workspaceQuickNav).toBeVisible();
      await expect(this.dashboard.workspaceTodos).toBeVisible();
      await waitForBusyIndicatorsToClear(this.dashboard.workspacePage);
    }
  }

  private async artifactPath(description: string, extension: string) {
    const now = new Date();
    const pad = (value: number, length = 2) =>
      String(value).padStart(length, "0");
    const day = `${now.getFullYear()}${pad(now.getMonth() + 1)}${pad(now.getDate())}`;
    const time = `${pad(now.getHours())}${pad(now.getMinutes())}${pad(now.getSeconds())}${pad(now.getMilliseconds(), 3)}`;
    const root = path.resolve(
      path.dirname(fileURLToPath(import.meta.url)),
      "../../..",
    );
    const directory = path.join(root, "temp", day);
    await mkdir(directory, { recursive: true });
    return path.join(directory, `${time}-${description}.${extension}`);
  }

  async screenshot(description: string) {
    const screenshotPath = await this.artifactPath(description, "png");
    await this.page.screenshot({ path: screenshotPath, fullPage: false });
    return screenshotPath;
  }

  async savePerformance(json: string) {
    const reportPath = await this.artifactPath(
      "menu-projection-performance",
      "json",
    );
    await writeFile(reportPath, json, "utf8");
    return reportPath;
  }

  async performanceEnvironment() {
    return {
      viewport: this.page.viewportSize(),
      userAgent: await this.page.evaluate(() => navigator.userAgent),
    };
  }

  async sampleSidebarToggle(durationMs = 500): Promise<MenuFrameSample> {
    const hiddenBefore = await this.isSidebarHidden();
    const probe = await this.headerSidebarToggle.evaluateHandle(
      (element, durationMs) => {
        let raf = 0;
        let timeout = 0;
        let settled = false;
        let resolveResult!: (value: {
          error: string | null;
          sample: Omit<MenuFrameSample, "direction"> | null;
        }) => void;
        const result = new Promise<{
          error: string | null;
          sample: Omit<MenuFrameSample, "direction"> | null;
        }>((resolve) => {
          resolveResult = resolve;
        });
        const cleanup = () => {
          cancelAnimationFrame(raf);
          clearTimeout(timeout);
          element.removeEventListener("click", onClick, true);
        };
        const finish = (value: Parameters<typeof resolveResult>[0]) => {
          if (settled) return;
          settled = true;
          cleanup();
          resolveResult(value);
        };
        const onClick = () => {
          const start = performance.now();
          let firstFrame: number | undefined;
          let previousTimestamp: number | undefined;
          let frameCount = 0;
          const frameIntervalsMs: number[] = [];
          const onFrame = (timestamp: number) => {
            const now = performance.now();
            firstFrame ??= now;
            frameCount += 1;
            if (previousTimestamp !== undefined)
              frameIntervalsMs.push(timestamp - previousTimestamp);
            previousTimestamp = timestamp;
            if (now - start >= durationMs) {
              finish({
                error: null,
                sample: {
                  clickToFirstFrameMs: firstFrame - start,
                  elapsedMs: now - start,
                  frameCount,
                  frameIntervalsMs,
                },
              });
            } else {
              raf = requestAnimationFrame(onFrame);
            }
          };
          raf = requestAnimationFrame(onFrame);
        };
        element.addEventListener("click", onClick, {
          capture: true,
          once: true,
        });
        timeout = window.setTimeout(
          () => finish({ error: "菜单帧采样未在10秒内完成", sample: null }),
          10_000,
        );
        return {
          result,
          dispose: () => finish({ error: "菜单帧采样已清理", sample: null }),
        };
      },
      durationMs,
    );
    try {
      await this.headerSidebarToggle.click();
      const result = await probe.evaluate(async (probe) => await probe.result);
      if (result.error || !result.sample)
        throw new Error(result.error ?? "菜单帧采样为空");
      await expect.poll(() => this.isSidebarHidden()).toBe(!hiddenBefore);
      return { ...result.sample, direction: hiddenBefore ? "show" : "hide" };
    } finally {
      // 清理当前采样创建的帧回调、点击监听器和超时，包括点击失败路径。
      await probe.evaluate((probe) => probe.dispose());
      await probe.dispose();
    }
  }
}
