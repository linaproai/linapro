import { mkdir } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import path from "node:path";

import { test, expect } from "../../fixtures/auth";
import { DashboardPage } from "../../pages/DashboardPage";
import type { MainLayout } from "../../pages/MainLayout";

// Regression coverage: https://github.com/vbenjs/vue-vben-admin/issues/6919
test.describe("TC007 主题切换性能回归", () => {
  test.beforeEach(async ({ adminPage }) => {
    await adminPage.setViewportSize({ width: 1440, height: 900 });
  });

  test("TC007a: 双向圆形揭幕方向正确且保留 ECharts 实例", async ({
    adminPage,
    mainLayout,
  }) => {
    const dashboardPage = new DashboardPage(adminPage);
    const screenshots = await createScreenshotPaths();

    await dashboardPage.gotoAnalytics();
    await mainLayout.prepareThemeModeWithoutAnimation("light");
    await dashboardPage.waitForAnalyticsChartInstances(4);
    await dashboardPage.waitForAnalyticsChartsPainted(4);

    const chartIdsBeforeSwitch =
      await dashboardPage.getAnalyticsChartInstanceIds();
    expect(chartIdsBeforeSwitch.length).toBeGreaterThanOrEqual(4);
    await dashboardPage.captureAnalyticsScreenshot(screenshots.light);

    expect(await mainLayout.hasStartViewTransitionApi()).toBeTruthy();
    await mainLayout.installThemeTransitionProbe();

    await mainLayout.ensureThemeMode("dark");
    await mainLayout.waitForThemeTransitionFinished(1);
    await expectThemeTransitionRuntimeClean(mainLayout);
    expect(await dashboardPage.getAnalyticsChartInstanceIds()).toEqual(
      chartIdsBeforeSwitch,
    );
    expectPaintedChartSnapshot(
      await dashboardPage.getAnalyticsChartCanvasPaintInfo(),
      4,
    );

    const darkSnapshot = await mainLayout.getThemeTransitionProbeSnapshot();
    expect(darkSnapshot.transitions).toHaveLength(1);
    expect(darkSnapshot.animations).toHaveLength(1);
    expectTransitionRecord(darkSnapshot.transitions[0], "dark");
    expectCircularThemeAnimation(darkSnapshot.animations[0], 1, "collapse");
    await expect
      .poll(async () => dashboardPage.getAnalyticsChartInstanceIds())
      .toEqual(chartIdsBeforeSwitch);
    await dashboardPage.waitForAnalyticsChartsPainted(4);
    await dashboardPage.captureAnalyticsScreenshot(screenshots.dark);

    await mainLayout.ensureThemeMode("light");
    await mainLayout.waitForThemeTransitionFinished(2);
    await expectThemeTransitionRuntimeClean(mainLayout);
    expect(await dashboardPage.getAnalyticsChartInstanceIds()).toEqual(
      chartIdsBeforeSwitch,
    );
    expectPaintedChartSnapshot(
      await dashboardPage.getAnalyticsChartCanvasPaintInfo(),
      4,
    );

    const lightSnapshot = await mainLayout.getThemeTransitionProbeSnapshot();
    expect(lightSnapshot.transitions).toHaveLength(2);
    expect(lightSnapshot.animations).toHaveLength(2);
    expect(
      lightSnapshot.transitions.length - darkSnapshot.transitions.length,
    ).toBe(1);
    expect(
      lightSnapshot.animations.length - darkSnapshot.animations.length,
    ).toBe(1);
    expectTransitionRecord(lightSnapshot.transitions[1], "light");
    expectCircularThemeAnimation(lightSnapshot.animations[1], 2, "expand");
    expect(darkSnapshot.animations[0].clipPath).toEqual([
      lightSnapshot.animations[1].clipPath[1],
      lightSnapshot.animations[1].clipPath[0],
    ]);
    await expect
      .poll(async () => dashboardPage.getAnalyticsChartInstanceIds())
      .toEqual(chartIdsBeforeSwitch);
    await dashboardPage.waitForAnalyticsChartsPainted(4);
    await dashboardPage.captureAnalyticsScreenshot(screenshots.lightRestored);

    console.info(
      `[TC007 timing] ${JSON.stringify(
        lightSnapshot.transitions.map((record) => ({
          call: record.callNumber,
          finishedMs: roundTiming(record.finishedDurationMs),
          readyMs: roundTiming(record.readyDurationMs),
          updateCallbackMs: roundTiming(record.updateCallbackDurationMs),
        })),
      )}`,
    );

    await mainLayout.captureNextThemeTransitionReadyCanvas();
    await mainLayout.ensureThemeMode("dark");
    await mainLayout.waitForThemeTransitionFinished(3);
    await expectThemeTransitionRuntimeClean(mainLayout);

    const readyBoundarySnapshot =
      await mainLayout.getThemeTransitionProbeSnapshot();
    expect(readyBoundarySnapshot.transitions).toHaveLength(3);
    expectReadyCanvasSnapshot(
      readyBoundarySnapshot.transitions[2],
      4,
      chartIdsBeforeSwitch,
    );
    expect(await dashboardPage.getAnalyticsChartInstanceIds()).toEqual(
      chartIdsBeforeSwitch,
    );

    await mainLayout.captureNextThemeTransitionReadyCanvas();
    await mainLayout.ensureThemeMode("light");
    await mainLayout.waitForThemeTransitionFinished(4);
    await expectThemeTransitionRuntimeClean(mainLayout);

    const restoredReadyBoundarySnapshot =
      await mainLayout.getThemeTransitionProbeSnapshot();
    expect(restoredReadyBoundarySnapshot.transitions).toHaveLength(4);
    expectReadyCanvasSnapshot(
      restoredReadyBoundarySnapshot.transitions[3],
      4,
      chartIdsBeforeSwitch,
    );
    expect(await dashboardPage.getAnalyticsChartInstanceIds()).toEqual(
      chartIdsBeforeSwitch,
    );
  });

  test("TC007b: 点击后立即刷新仍保留目标主题", async ({ mainLayout }) => {
    await mainLayout.prepareThemeModeWithoutAnimation("light");

    await mainLayout.switchThemeModeAndReloadImmediately("dark");
    expect(await mainLayout.getThemeMode()).toBe("dark");

    await mainLayout.switchThemeModeAndReloadImmediately("light");
    expect(await mainLayout.getThemeMode()).toBe("light");
  });
});

function expectTransitionRecord(
  record: {
    finishedDurationMs: null | number;
    finishedRejected: boolean;
    finishedSettled: boolean;
    newRootZIndex: string;
    oldRootZIndex: string;
    readyDurationMs: null | number;
    readyDirection: string;
    readyNewRootClipPath: string;
    readyOldRootClipPath: string;
    readyRejected: boolean;
    readySettled: boolean;
    updateCallbackDurationMs: null | number;
    updateCallbackSettled: boolean;
  },
  targetMode: "dark" | "light",
) {
  expect(record.updateCallbackSettled).toBeTruthy();
  expect(record.readySettled).toBeTruthy();
  expect(record.finishedSettled).toBeTruthy();
  expect(record.readyRejected).toBeFalsy();
  expect(record.finishedRejected).toBeFalsy();
  expect(record.updateCallbackDurationMs).not.toBeNull();
  expect(
    record.updateCallbackDurationMs ?? Number.POSITIVE_INFINITY,
  ).toBeLessThanOrEqual(200);
  expect(record.readyDurationMs).not.toBeNull();
  expect(
    record.readyDurationMs ?? Number.POSITIVE_INFINITY,
  ).toBeLessThanOrEqual(300);
  expect(record.finishedDurationMs).not.toBeNull();
  expect(
    record.finishedDurationMs ?? Number.POSITIVE_INFINITY,
  ).toBeLessThanOrEqual(1000);
  expect(record.finishedDurationMs ?? 0).toBeGreaterThanOrEqual(
    record.readyDurationMs ?? Number.POSITIVE_INFINITY,
  );

  const newRootZIndex = Number(record.newRootZIndex);
  const oldRootZIndex = Number(record.oldRootZIndex);
  expect(Number.isFinite(newRootZIndex)).toBeTruthy();
  expect(Number.isFinite(oldRootZIndex)).toBeTruthy();
  if (targetMode === "dark") {
    expect(oldRootZIndex).toBeGreaterThan(newRootZIndex);
    expect(record.readyDirection).toBe("to-dark");
    expect(record.readyOldRootClipPath).toMatch(
      /^circle\((?!0px )\d+(?:\.\d+)?px at \d+(?:\.\d+)?px \d+(?:\.\d+)?px\)$/u,
    );
  } else {
    expect(newRootZIndex).toBeGreaterThan(oldRootZIndex);
    expect(record.readyDirection).toBe("to-light");
    expect(record.readyNewRootClipPath).toMatch(
      /^circle\(0px at \d+(?:\.\d+)?px \d+(?:\.\d+)?px\)$/u,
    );
    expect(record.readyOldRootClipPath).toBe("none");
  }
}

async function expectThemeTransitionRuntimeClean(mainLayout: MainLayout) {
  await expect
    .poll(() => mainLayout.getThemeTransitionRuntimeState())
    .toEqual({
      direction: "",
      pseudoAnimationCount: 0,
      radius: "",
      x: "",
      y: "",
    });
}

function expectReadyCanvasSnapshot(
  record: {
    readyCanvasCount: null | number;
    readyChartInstanceIds: null | string[];
    readyPaintedCanvasCount: null | number;
    readyRejected: boolean;
    readySettled: boolean;
  },
  minimumCount: number,
  expectedInstanceIds: string[],
) {
  expect(record.readySettled).toBeTruthy();
  expect(record.readyRejected).toBeFalsy();
  expect(record.readyCanvasCount).not.toBeNull();
  expect(record.readyCanvasCount ?? 0).toBeGreaterThanOrEqual(minimumCount);
  expect(record.readyPaintedCanvasCount).toBe(record.readyCanvasCount);
  expect(record.readyChartInstanceIds).toEqual(expectedInstanceIds);
}

function expectPaintedChartSnapshot(
  paintInfo: Array<{
    canvasCount: number;
    paintedCanvasCount: number;
  }>,
  minimumCount: number,
) {
  expect(paintInfo.length).toBeGreaterThanOrEqual(minimumCount);
  expect(
    paintInfo.every(
      ({ canvasCount, paintedCanvasCount }) =>
        canvasCount > 0 && paintedCanvasCount > 0,
    ),
  ).toBeTruthy();
}

function expectCircularThemeAnimation(
  animation: {
    clipPath: string[];
    duration: null | number;
    easing: string;
    fill: FillMode | "";
    pseudoElement: string;
    transitionCall: number;
  },
  transitionCall: number,
  direction: "collapse" | "expand",
) {
  expect(animation).toMatchObject({
    duration: 450,
    easing: "ease-in",
    fill: "both",
    pseudoElement:
      direction === "collapse"
        ? "::view-transition-old(root)"
        : "::view-transition-new(root)",
    transitionCall,
  });
  expect(animation.clipPath).toHaveLength(2);
  const zeroCirclePattern =
    /^circle\(0px at \d+(?:\.\d+)?px \d+(?:\.\d+)?px\)$/u;
  const fullCirclePattern =
    /^circle\((?!0px )\d+(?:\.\d+)?px at \d+(?:\.\d+)?px \d+(?:\.\d+)?px\)$/u;
  const expectedPatterns =
    direction === "collapse"
      ? [fullCirclePattern, zeroCirclePattern]
      : [zeroCirclePattern, fullCirclePattern];

  expect(animation.clipPath[0]).toMatch(expectedPatterns[0]);
  expect(animation.clipPath[1]).toMatch(expectedPatterns[1]);
}

function roundTiming(value: null | number) {
  return value === null ? null : Number(value.toFixed(2));
}

async function createScreenshotPaths() {
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
  const directory = fileURLToPath(
    new URL(`../../../../temp/${day}/`, import.meta.url),
  );

  await mkdir(directory, { recursive: true });

  return {
    dark: path.join(directory, `${time}-theme-switch-dark.png`),
    light: path.join(directory, `${time}-theme-switch-light.png`),
    lightRestored: path.join(
      directory,
      `${time}-theme-switch-light-restored.png`,
    ),
  };
}
