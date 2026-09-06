import { test, expect } from "../../fixtures/auth";
import {
  menuLayouts,
  MenuProjectionPage,
  type MenuFrameSample,
} from "../../pages/MenuProjectionPage";

test.describe("TC-1 菜单投影回归", () => {
  test("TC-1a: 侧栏收放保持中文名称、导航和选中状态", async ({
    adminPage,
  }, testInfo) => {
    const menu = new MenuProjectionPage(adminPage);
    await menu.setViewport();
    const initial = await menu.rememberPreferences();
    try {
      await menu.switchLanguage("zh-CN");
      await menu.switchLayout("sidebar-nav");
      await menu.setSidebarHidden(false);
      await menu.setCollapsed(false);
      await menu.selectDashboard("sidebar-nav", "analytics", "zh-CN");
      await menu.expectDashboardLabels("sidebar-nav", "zh-CN");
      const namesBefore = await menu.topLevelNames();
      expect(namesBefore).toContain("权限管理");
      expect(namesBefore).toContain("系统设置");
      await testInfo.attach("中文展开菜单", {
        path: await menu.screenshot("menu-projection-expanded-zh"),
        contentType: "image/png",
      });

      await menu.setCollapsed(true);
      await menu.expectDashboardLabels("sidebar-nav", "zh-CN");
      await menu.selectDashboard("sidebar-nav", "workspace", "zh-CN");
      expect(await menu.isCollapsed()).toBe(true);
      await testInfo.attach("中文收起悬浮菜单", {
        path: await menu.screenshot("menu-projection-collapsed-zh"),
        contentType: "image/png",
      });

      await menu.setCollapsed(false);
      await menu.selectDashboard("sidebar-nav", "analytics", "zh-CN");
      await menu.expectDashboardLabels("sidebar-nav", "zh-CN");
      expect(await menu.topLevelNames()).toEqual(namesBefore);
      await testInfo.attach("中文恢复展开菜单", {
        path: await menu.screenshot("menu-projection-restored-zh"),
        contentType: "image/png",
      });
    } catch (error) {
      await testInfo.attach("收放失败时页面", {
        path: await menu.screenshot("menu-projection-collapse-failure"),
        contentType: "image/png",
      });
      throw error;
    } finally {
      await menu.restorePreferences(initial);
    }
  });

  test("TC-1b: 六种布局的菜单随语言更新并正确导航", async ({
    adminPage,
  }, testInfo) => {
    test.setTimeout(300_000);
    const menu = new MenuProjectionPage(adminPage);
    await menu.setViewport();
    const initial = await menu.rememberPreferences();
    try {
      await menu.switchLayout("sidebar-nav");
      await menu.setSidebarHidden(false);
      await menu.setCollapsed(false);
      for (const layout of menuLayouts) {
        await test.step(layout, async () => {
          await menu.switchLanguage("zh-CN");
          const preferencesScreenshot = await menu.switchLayout(layout, true);
          await testInfo.attach(`${layout}布局设置`, {
            path: preferencesScreenshot!,
            contentType: "image/png",
          });
          await menu.selectDashboard(layout, "workspace", "zh-CN");
          await menu.expectDashboardLabels(layout, "zh-CN");
          await testInfo.attach(`${layout}中文`, {
            path: await menu.screenshot(`menu-projection-${layout}-zh`),
            contentType: "image/png",
          });

          await menu.switchLanguage("en-US");
          // 宿主切换语言会同步公共布局配置；重新通过UI选择本轮待测布局。
          if (layout !== "sidebar-nav") {
            await menu.expectDashboardLabels("sidebar-nav", "en-US");
            await menu.switchLayout(layout);
          }
          await menu.expectDashboardLabels(layout, "en-US");
          await menu.selectDashboard(layout, "analytics", "en-US");
          await testInfo.attach(`${layout}英文`, {
            path: await menu.screenshot(`menu-projection-${layout}-en`),
            contentType: "image/png",
          });

          await menu.switchLanguage("zh-CN");
          if (layout !== "sidebar-nav") {
            await menu.expectDashboardLabels("sidebar-nav", "zh-CN");
            await menu.switchLayout(layout);
          }
          await menu.expectDashboardLabels(layout, "zh-CN");
          await menu.selectDashboard(layout, "workspace", "zh-CN");
        });
      }
    } catch (error) {
      await testInfo.attach("布局失败时页面", {
        path: await menu.screenshot("menu-projection-layout-failure"),
        contentType: "image/png",
      });
      throw error;
    } finally {
      await menu.restorePreferences(initial);
    }
  });

  test("TC-1c: 顶栏按钮隐藏展开侧栏的多轮帧采样供版本对照", async ({
    adminPage,
  }, testInfo) => {
    const menu = new MenuProjectionPage(adminPage);
    await menu.setViewport();
    const initial = await menu.rememberPreferences();
    try {
      await menu.switchLanguage("zh-CN");
      await menu.switchLayout("sidebar-nav");
      await menu.setSidebarHidden(false);
      await menu.setCollapsed(false);
      await menu.selectDashboard("sidebar-nav", "analytics", "zh-CN");
      await testInfo.attach("性能采样初始页面", {
        path: await menu.screenshot("menu-projection-performance-before"),
        contentType: "image/png",
      });

      // 预热顶栏按钮的隐藏和展开；500ms窗口包含动画结束后的空闲帧。
      for (let warmup = 0; warmup < 2; warmup += 1) {
        await menu.sampleSidebarToggle();
        if (warmup === 0) {
          await testInfo.attach("顶栏按钮隐藏侧栏", {
            path: await menu.screenshot("menu-projection-header-hidden"),
            contentType: "image/png",
          });
        }
        await menu.sampleSidebarToggle();
      }
      const samples: MenuFrameSample[] = [];
      for (let round = 0; round < 6; round += 1) {
        samples.push(await menu.sampleSidebarToggle());
        samples.push(await menu.sampleSidebarToggle());
      }
      expect(await menu.isCollapsed()).toBe(false);
      expect(await menu.isSidebarHidden()).toBe(false);
      for (const sample of samples) {
        expect(sample.frameCount).toBeGreaterThan(1);
        expect(sample.clickToFirstFrameMs).toBeGreaterThanOrEqual(0);
        expect(sample.frameIntervalsMs).toHaveLength(sample.frameCount - 1);
        expect(
          sample.frameIntervalsMs.every(
            (interval) => interval > 0 && Number.isFinite(interval),
          ),
        ).toBe(true);
      }
      await menu.waitForDashboard("analytics");
      await menu.expectDashboardLabels("sidebar-nav", "zh-CN");
      await testInfo.attach("性能采样结束页面", {
        path: await menu.screenshot("menu-projection-performance-after"),
        contentType: "image/png",
      });

      const result = {
        variant: process.env.E2E_MENU_PROJECTION_VARIANT ?? "unspecified",
        run: process.env.E2E_MENU_PROJECTION_RUN ?? "unspecified",
        environment: await menu.performanceEnvironment(),
        warmupCycles: 2,
        measuredCycles: 6,
        sampleWindowMs: 500,
        trigger: "header-sidebar-toggle",
        summary: (["hide", "show"] as const).map((direction) => {
          const selected = samples.filter(
            (sample) => sample.direction === direction,
          );
          return {
            direction,
            clickToFirstFrameMs: summarize(
              selected.map((sample) => sample.clickToFirstFrameMs),
            ),
            frameIntervalsMs: summarize(
              selected.flatMap((sample) => sample.frameIntervalsMs),
            ),
          };
        }),
        samples,
      };
      const json = JSON.stringify(result, null, 2);
      console.info(`[TC001 menu-projection] ${JSON.stringify(result)}`);
      await testInfo.attach("menu-projection-performance.json", {
        path: await menu.savePerformance(json),
        contentType: "application/json",
      });
    } finally {
      await menu.restorePreferences(initial);
    }
  });
});

function summarize(values: number[]) {
  const sorted = [...values].sort((left, right) => left - right);
  return {
    count: sorted.length,
    min: sorted[0],
    median:
      (sorted[Math.floor((sorted.length - 1) / 2)]! +
        sorted[Math.ceil((sorted.length - 1) / 2)]!) /
      2,
    p95: sorted[Math.ceil(sorted.length * 0.95) - 1],
    max: sorted.at(-1),
  };
}
