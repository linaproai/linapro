import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";

import { waitForRouteReady } from "../support/ui";

type SidebarMenuLabel = RegExp | string;

export class MainLayout {
  constructor(private page: Page) {}

  private async waitForLocalePersistence(locale: string) {
    await expect
      .poll(async () => {
        try {
          return await this.page.evaluate(() => {
            const key = Object.keys(localStorage).find((item) =>
              item.endsWith("preferences-locale"),
            );
            if (!key) {
              return "";
            }
            try {
              return JSON.parse(localStorage.getItem(key) || "{}")?.value || "";
            } catch {
              return "";
            }
          });
        } catch {
          return "";
        }
      })
      .toBe(locale);
  }

  get sidebar() {
    return this.page
      .locator('[class*="sidebar"], [class*="menu"], nav')
      .first();
  }

  get languageToggleTrigger() {
    return this.page.getByTestId("language-toggle-trigger").first();
  }

  get brandLogoImage() {
    return this.page.locator('img[alt^="LinaPro"]:visible').first();
  }

  get brandLogoMark() {
    return this.page.locator(".vben-logo__mark:visible").first();
  }

  sidebarMenuItem(label: SidebarMenuLabel) {
    if (typeof label !== "string") {
      return this.sidebar.getByText(label).first();
    }
    return this.sidebar.getByText(label, { exact: true }).first();
  }

  private sidebarSubmenuTitle(label: SidebarMenuLabel) {
    return this.sidebar
      .locator(".ant-menu-submenu-title, .vben-sub-menu-content")
      .filter({ hasText: label })
      .first();
  }

  async expandSidebarGroup(label: SidebarMenuLabel) {
    const title = this.sidebarSubmenuTitle(label);
    await expect(title).toBeVisible();
    const submenu = title
      .locator(
        "xpath=ancestor::*[contains(concat(' ', normalize-space(@class), ' '), ' ant-menu-submenu ') or contains(concat(' ', normalize-space(@class), ' '), ' vben-sub-menu ')][1]",
      )
      .first();
    const className =
      (await submenu.getAttribute("class").catch(() => "")) ?? "";
    if (
      !className.includes("ant-menu-submenu-open") &&
      !className.includes("is-opened")
    ) {
      await title.click();
    }
  }

  async expectSidebarMenuVisible(label: SidebarMenuLabel) {
    for (let attempt = 1; attempt <= 3; attempt += 1) {
      const menuItem = this.sidebarMenuItem(label);
      if (await menuItem.isVisible().catch(() => false)) {
        await expect(menuItem).toBeVisible();
        return menuItem;
      }

      await this.expandSidebarGroup(/Extension Center|Extensions|扩展中心/);
      if (await menuItem.isVisible({ timeout: 1000 }).catch(() => false)) {
        await expect(menuItem).toBeVisible();
        return menuItem;
      }

      if (attempt < 3) {
        await this.page.reload({ waitUntil: "domcontentloaded" });
        await waitForRouteReady(this.page, 15000);
      }
    }

    const menuItem = this.sidebarMenuItem(label);
    await expect(menuItem).toBeVisible();
    return menuItem;
  }

  tabTitle(label: string) {
    return this.page
      .locator('[data-tab-item="true"] span[title]')
      .filter({ hasText: label })
      .first();
  }

  breadcrumbItem(label: string) {
    return this.page
      .locator('nav[aria-label="breadcrumb"]')
      .getByText(label, { exact: true })
      .first();
  }

  activeTabTitle() {
    return this.page
      .locator('[data-tab-item="true"].is-active span[title]')
      .first();
  }

  get userDropdownTrigger() {
    return this.page.getByTestId("layout-user-dropdown-trigger").first();
  }

  get userDropdownMenu() {
    return this.page.getByTestId("layout-user-dropdown-menu");
  }

  get userDropdownProfile() {
    return this.page.getByTestId("layout-user-dropdown-profile");
  }

  get userDropdownName() {
    return this.page.getByTestId("layout-user-dropdown-name");
  }

  get preferencesTrigger() {
    return this.page.getByTestId("preferences-trigger").first();
  }

  get preferencesDrawerTitle() {
    return this.page.getByTestId("preferences-drawer-title").first();
  }

  get preferencesDrawerSubtitle() {
    return this.page.getByTestId("preferences-drawer-subtitle").first();
  }

  get preferencesDrawer() {
    return this.page
      .locator('[role="dialog"], [data-slot="sheet-content"]')
      .filter({
        has: this.page.locator('[data-testid="preferences-drawer-title"]'),
      })
      .first();
  }

  get tenantSwitcher() {
    return this.page.getByTestId("tenant-switcher");
  }

  get workspaceFooterCopyright() {
    return this.page
      .locator("footer")
      .filter({ hasText: "Copyright ©" })
      .first()
      .getByText(/Copyright ©/);
  }

  async navigateTo(menuGroup: string, menuItem: string) {
    await this.page.getByText(menuGroup).click();
    await this.page.getByText(menuItem).click();
    await this.page.waitForLoadState("networkidle");
  }

  async switchLanguage(label: "English" | "简体中文") {
    const localeMap = {
      English: "en-US",
      简体中文: "zh-CN",
    } as const;
    const locale = localeMap[label];
    await this.languageToggleTrigger.click();
    await this.page.getByText(label, { exact: true }).last().click();
    await this.waitForLocalePersistence(locale);
    await expect
      .poll(async () => await this.page.locator("html").getAttribute("lang"))
      .toBe(locale);
    await this.page.waitForLoadState("networkidle");
    await waitForRouteReady(this.page);
  }

  async switchLanguageFromPreferences(label: "English" | "简体中文") {
    const localeMap = {
      English: "en-US",
      简体中文: "zh-CN",
    } as const;
    const locale = localeMap[label];
    if (!(await this.preferencesDrawer.isVisible().catch(() => false))) {
      await this.openPreferences();
    }
    await this.preferencesDrawer
      .getByRole("tab", { name: /General|通用/ })
      .click();
    await this.preferencesDrawer.getByRole("combobox").first().click();
    await this.page.getByRole("option", { name: label }).click();
    await this.waitForLocalePersistence(locale);
    await expect
      .poll(async () => await this.page.locator("html").getAttribute("lang"))
      .toBe(locale);
    await waitForRouteReady(this.page);
  }

  async getBrandLogoInfo() {
    await expect(this.brandLogoImage).toBeVisible();
    await expect
      .poll(async () =>
        this.brandLogoImage.evaluate(
          (img) => (img as HTMLImageElement).naturalWidth,
        ),
      )
      .toBeGreaterThan(0);

    return this.brandLogoImage.evaluate((node) => {
      const img = node as HTMLImageElement;
      return {
        currentSrc: img.currentSrc,
        height: img.clientHeight,
        naturalHeight: img.naturalHeight,
        naturalWidth: img.naturalWidth,
        parentText:
          (img.closest("a") ?? img.parentElement)?.textContent?.trim() ?? "",
        src: img.getAttribute("src") ?? "",
        width: img.clientWidth,
      };
    });
  }

  async getBrandLogoGlowInfo() {
    await expect(this.brandLogoMark).toBeVisible();

    return this.brandLogoMark.evaluate((mark) => {
      const root = mark.closest(".vben-logo");
      const link = mark.closest("a");
      const image = mark.querySelector("img") as HTMLElement;
      const beforeStyle = window.getComputedStyle(mark, "::before");
      const imageStyle = image ? window.getComputedStyle(image) : null;
      const markRect = mark.getBoundingClientRect();
      const linkRect = link?.getBoundingClientRect();

      return {
        beforeBackgroundImage: beforeStyle.backgroundImage,
        beforeFilter: beforeStyle.filter,
        beforeOpacity: beforeStyle.opacity,
        imageClientHeight: image?.clientHeight ?? 0,
        imageClientWidth: image?.clientWidth ?? 0,
        imageFilter: imageStyle?.filter ?? "",
        isDarkRoot: root?.classList.contains("dark") ?? false,
        linkHeight: linkRect?.height ?? 0,
        linkOverflow: link ? window.getComputedStyle(link).overflow : "",
        markHeight: markRect.height,
        markWidth: markRect.width,
      };
    });
  }

  async ensureThemeMode(mode: "dark" | "light") {
    const shouldBeDark = mode === "dark";
    const isDark = (await this.getThemeMode()) === "dark";

    if (isDark !== shouldBeDark) {
      await this.themeToggleButton(mode).click();
    }

    await expect
      .poll(async () => (await this.getThemeMode()) === "dark")
      .toBe(shouldBeDark);
    await waitForRouteReady(this.page);
  }

  private themeToggleButton(mode: "dark" | "light") {
    return this.page.locator(`button[aria-label="${mode}"]:visible`).first();
  }

  async getThemeMode(): Promise<"dark" | "light"> {
    const isDark = await this.page.evaluate(() =>
      document.documentElement.classList.contains("dark"),
    );
    return isDark ? "dark" : "light";
  }

  async prepareThemeModeWithoutAnimation(mode: "dark" | "light") {
    await this.page.emulateMedia({ reducedMotion: "reduce" });
    try {
      await this.ensureThemeMode(mode);
    } finally {
      await this.page.emulateMedia({ reducedMotion: "no-preference" });
    }
  }

  async installThemeTransitionProbe() {
    await this.page.evaluate(() => {
      type NativeViewTransition = {
        finished: Promise<unknown>;
        ready: Promise<unknown>;
      };
      type ThemeTransitionRecord = {
        callNumber: number;
        finishedDurationMs: null | number;
        finishedRejected: boolean;
        finishedSettled: boolean;
        newRootZIndex: string;
        oldRootZIndex: string;
        readyCanvasCount: null | number;
        readyChartInstanceIds: null | string[];
        readyDurationMs: null | number;
        readyDirection: string;
        readyNewRootClipPath: string;
        readyOldRootClipPath: string;
        readyPaintedCanvasCount: null | number;
        readyRejected: boolean;
        readySettled: boolean;
        updateCallbackDurationMs: null | number;
        updateCallbackSettled: boolean;
      };
      type RootAnimationRecord = {
        clipPath: string[];
        duration: null | number;
        easing: string;
        fill: FillMode | "";
        pseudoElement: string;
        transitionCall: number;
      };
      type ThemeTransitionProbe = {
        animations: RootAnimationRecord[];
        captureNextReadyCanvas: boolean;
        pendingClickStartedAt?: number;
        records: ThemeTransitionRecord[];
        transitions: NativeViewTransition[];
      };

      const transitionDocument = document as Document & {
        startViewTransition?: (
          updateCallback: () => unknown,
        ) => NativeViewTransition;
      };
      const originalStartViewTransition =
        transitionDocument.startViewTransition;
      if (typeof originalStartViewTransition !== "function") {
        throw new Error("Native document.startViewTransition is unavailable");
      }

      const probeWindow = window as Window & {
        __themeTransitionProbe?: ThemeTransitionProbe;
      };
      const probe: ThemeTransitionProbe = {
        animations: [],
        captureNextReadyCanvas: false,
        records: [],
        transitions: [],
      };
      probeWindow.__themeTransitionProbe = probe;

      const root = document.documentElement;
      document.addEventListener(
        "click",
        (event) => {
          const target = event.target;
          if (
            target instanceof Element &&
            target.closest("button.theme-toggle")
          ) {
            probe.pendingClickStartedAt = performance.now();
          }
        },
        true,
      );

      const getCanvasPaintState = () => {
        const chartInstances = Array.from(
          document.querySelectorAll<HTMLElement>(
            '[data-testid="dashboard-analytics-page"] [_echarts_instance_]',
          ),
        );
        const canvases = chartInstances.flatMap((chart) =>
          Array.from(chart.querySelectorAll<HTMLCanvasElement>("canvas")),
        );
        const paintedCanvasCount = canvases.filter((canvas) => {
          if (canvas.width === 0 || canvas.height === 0) {
            return false;
          }

          const context = canvas.getContext("2d", {
            willReadFrequently: true,
          });
          if (!context) {
            return false;
          }

          const pixels = context.getImageData(
            0,
            0,
            canvas.width,
            canvas.height,
          ).data;
          const firstRed = pixels[0];
          const firstGreen = pixels[1];
          const firstBlue = pixels[2];
          const firstAlpha = pixels[3];

          for (let index = 4; index < pixels.length; index += 4) {
            if (
              pixels[index] !== firstRed ||
              pixels[index + 1] !== firstGreen ||
              pixels[index + 2] !== firstBlue ||
              pixels[index + 3] !== firstAlpha
            ) {
              return true;
            }
          }

          return false;
        }).length;

        return {
          canvasCount: canvases.length,
          instanceIds: chartInstances
            .map((chart) => chart.getAttribute("_echarts_instance_"))
            .filter((id): id is string => Boolean(id)),
          paintedCanvasCount,
        };
      };

      const originalRootAnimate = root.animate.bind(root);
      Object.defineProperty(root, "animate", {
        configurable: true,
        value: (
          keyframes: Keyframe[] | PropertyIndexedKeyframes,
          options?: KeyframeAnimationOptions | number,
        ) => {
          const animationOptions =
            typeof options === "number"
              ? { duration: options }
              : (options ?? {});
          const rawClipPath = Array.isArray(keyframes)
            ? keyframes.flatMap((keyframe) => {
                const value = (keyframe as Record<string, unknown>).clipPath;
                return Array.isArray(value) ? value : [value];
              })
            : (keyframes as Record<string, unknown>).clipPath;
          const clipPath = (
            Array.isArray(rawClipPath) ? rawClipPath : [rawClipPath]
          ).filter((value): value is string => typeof value === "string");
          const extendedOptions =
            animationOptions as KeyframeAnimationOptions & {
              pseudoElement?: string;
            };

          probe.animations.push({
            clipPath,
            duration:
              typeof extendedOptions.duration === "number"
                ? extendedOptions.duration
                : null,
            easing: extendedOptions.easing ?? "",
            fill: extendedOptions.fill ?? "",
            pseudoElement: extendedOptions.pseudoElement ?? "",
            transitionCall: probe.records.length,
          });

          return originalRootAnimate(keyframes, options);
        },
      });

      Object.defineProperty(document, "startViewTransition", {
        configurable: true,
        value: (updateCallback: () => unknown) => {
          const startedAt = probe.pendingClickStartedAt ?? performance.now();
          const captureReadyCanvas = probe.captureNextReadyCanvas;
          probe.captureNextReadyCanvas = false;
          probe.pendingClickStartedAt = undefined;
          const record: ThemeTransitionRecord = {
            callNumber: probe.records.length + 1,
            finishedDurationMs: null,
            finishedRejected: false,
            finishedSettled: false,
            newRootZIndex: "",
            oldRootZIndex: "",
            readyCanvasCount: null,
            readyChartInstanceIds: null,
            readyDurationMs: null,
            readyDirection: "",
            readyNewRootClipPath: "",
            readyOldRootClipPath: "",
            readyPaintedCanvasCount: null,
            readyRejected: false,
            readySettled: false,
            updateCallbackDurationMs: null,
            updateCallbackSettled: false,
          };
          probe.records.push(record);

          const transition = originalStartViewTransition.call(
            document,
            async () => {
              const updateStartedAt = performance.now();
              try {
                return await updateCallback();
              } finally {
                record.updateCallbackDurationMs =
                  performance.now() - updateStartedAt;
                record.updateCallbackSettled = true;
              }
            },
          );
          probe.transitions.push(transition);

          void transition.ready.then(
            () => {
              record.readyDurationMs = performance.now() - startedAt;
              record.newRootZIndex = getComputedStyle(
                root,
                "::view-transition-new(root)",
              ).zIndex;
              record.oldRootZIndex = getComputedStyle(
                root,
                "::view-transition-old(root)",
              ).zIndex;
              record.readyDirection =
                root.dataset.themeTransitionDirection ?? "";
              record.readyNewRootClipPath = getComputedStyle(
                root,
                "::view-transition-new(root)",
              ).clipPath;
              record.readyOldRootClipPath = getComputedStyle(
                root,
                "::view-transition-old(root)",
              ).clipPath;
              if (captureReadyCanvas) {
                const canvasState = getCanvasPaintState();
                record.readyCanvasCount = canvasState.canvasCount;
                record.readyChartInstanceIds = canvasState.instanceIds;
                record.readyPaintedCanvasCount = canvasState.paintedCanvasCount;
              }
              record.readySettled = true;
            },
            () => {
              record.readyDurationMs = performance.now() - startedAt;
              record.readyRejected = true;
              record.readySettled = true;
            },
          );
          void transition.finished.then(
            () => {
              record.finishedDurationMs = performance.now() - startedAt;
              record.finishedSettled = true;
            },
            () => {
              record.finishedDurationMs = performance.now() - startedAt;
              record.finishedRejected = true;
              record.finishedSettled = true;
            },
          );

          return transition;
        },
      });
    });
  }

  async captureNextThemeTransitionReadyCanvas() {
    await this.page.evaluate(() => {
      const probe = (
        window as Window & {
          __themeTransitionProbe?: {
            captureNextReadyCanvas: boolean;
          };
        }
      ).__themeTransitionProbe;
      if (!probe) {
        throw new Error("Theme transition probe is not installed");
      }
      probe.captureNextReadyCanvas = true;
    });
  }

  async hasStartViewTransitionApi() {
    return this.page.evaluate(
      () =>
        typeof (
          document as Document & {
            startViewTransition?: (...args: unknown[]) => unknown;
          }
        ).startViewTransition === "function",
    );
  }

  async waitForThemeTransitionReady(expectedCalls: number) {
    await expect
      .poll(async () => {
        const snapshot = await this.getThemeTransitionProbeSnapshot();
        return snapshot.transitions.length;
      })
      .toBe(expectedCalls);

    await this.page.evaluate(async (callCount) => {
      const probe = (
        window as Window & {
          __themeTransitionProbe?: {
            transitions: Array<{ ready: Promise<unknown> }>;
          };
        }
      ).__themeTransitionProbe;
      if (!probe || probe.transitions.length < callCount) {
        throw new Error(`Expected ${callCount} captured theme transitions`);
      }

      await Promise.all(
        probe.transitions
          .slice(0, callCount)
          .map((transition) => transition.ready),
      );
    }, expectedCalls);

    await expect
      .poll(async () => {
        const snapshot = await this.getThemeTransitionProbeSnapshot();
        return snapshot.transitions
          .slice(0, expectedCalls)
          .every(
            (record) =>
              record.updateCallbackSettled &&
              record.readySettled &&
              !record.readyRejected,
          );
      })
      .toBeTruthy();
  }

  async waitForThemeTransitionFinished(expectedCalls: number) {
    await expect
      .poll(async () => {
        const snapshot = await this.getThemeTransitionProbeSnapshot();
        return snapshot.transitions.length;
      })
      .toBe(expectedCalls);

    await this.page.evaluate(async (callCount) => {
      const probe = (
        window as Window & {
          __themeTransitionProbe?: {
            transitions: Array<{ finished: Promise<unknown> }>;
          };
        }
      ).__themeTransitionProbe;
      if (!probe || probe.transitions.length < callCount) {
        throw new Error(`Expected ${callCount} captured theme transitions`);
      }

      await Promise.all(
        probe.transitions
          .slice(0, callCount)
          .map((transition) => transition.finished),
      );
    }, expectedCalls);

    await expect
      .poll(async () => {
        const snapshot = await this.getThemeTransitionProbeSnapshot();
        return snapshot.transitions
          .slice(0, expectedCalls)
          .every(
            (record) =>
              record.updateCallbackSettled &&
              record.readySettled &&
              record.finishedSettled,
          );
      })
      .toBeTruthy();
  }

  async getThemeTransitionProbeSnapshot() {
    return this.page.evaluate(() => {
      const probe = (
        window as Window & {
          __themeTransitionProbe?: {
            animations: Array<{
              clipPath: string[];
              duration: null | number;
              easing: string;
              fill: FillMode | "";
              pseudoElement: string;
              transitionCall: number;
            }>;
            records: Array<{
              callNumber: number;
              finishedDurationMs: null | number;
              finishedRejected: boolean;
              finishedSettled: boolean;
              newRootZIndex: string;
              oldRootZIndex: string;
              readyCanvasCount: null | number;
              readyChartInstanceIds: null | string[];
              readyDurationMs: null | number;
              readyDirection: string;
              readyNewRootClipPath: string;
              readyOldRootClipPath: string;
              readyPaintedCanvasCount: null | number;
              readyRejected: boolean;
              readySettled: boolean;
              updateCallbackDurationMs: null | number;
              updateCallbackSettled: boolean;
            }>;
          };
        }
      ).__themeTransitionProbe;
      if (!probe) {
        throw new Error("Theme transition probe is not installed");
      }

      return {
        animations: probe.animations.map((record) => ({
          ...record,
          clipPath: [...record.clipPath],
        })),
        transitions: probe.records.map((record) => ({ ...record })),
      };
    });
  }

  async getThemeTransitionRuntimeState() {
    return this.page.evaluate(() => {
      const root = document.documentElement;
      const pseudoAnimationCount = document.getAnimations().filter((item) => {
        const effect = item.effect as KeyframeEffect | null;
        return effect?.pseudoElement?.startsWith("::view-transition-");
      }).length;

      return {
        direction: root.dataset.themeTransitionDirection ?? "",
        pseudoAnimationCount,
        radius: root.style.getPropertyValue("--theme-transition-radius"),
        x: root.style.getPropertyValue("--theme-transition-x"),
        y: root.style.getPropertyValue("--theme-transition-y"),
      };
    });
  }

  async switchThemeModeAndReloadImmediately(mode: "dark" | "light") {
    const initialMode = await this.getThemeMode();
    if (initialMode === mode) {
      throw new Error(
        `Theme must differ from target mode before switching to ${mode}`,
      );
    }

    await this.page.evaluate(() => {
      const never = new Promise<never>(() => {});
      Object.defineProperty(document, "startViewTransition", {
        configurable: true,
        value: () => ({
          finished: never,
          ready: never,
          updateCallbackDone: never,
        }),
      });
    });
    await this.themeToggleButton(mode).click();
    expect(await this.getThemeMode()).toBe(initialMode);
    await this.page.reload({ waitUntil: "domcontentloaded" });
    await waitForRouteReady(this.page, 15000);
  }

  async openUserDropdown() {
    await this.userDropdownTrigger.click();
    await expect(this.userDropdownMenu).toBeVisible();
  }

  async openPreferences() {
    await expect(this.preferencesTrigger).toBeVisible();
    await this.preferencesTrigger.click();
    await expect(this.preferencesDrawerTitle).toBeVisible();
    await expect(this.preferencesDrawer).toBeVisible();
  }

  async openPreferencesTab(label: string | RegExp) {
    await this.openPreferences();
    await this.preferencesDrawer.getByRole("tab", { name: label }).click();
  }

  async logout() {
    // Use keyboard shortcut Alt+Q to trigger the logout modal
    // This avoids the complex DOM interaction with the user dropdown
    await this.page.keyboard.press("Alt+KeyQ");

    // Wait for the confirmation modal to appear
    // The modal asks "是否退出登录？" with 确认/取消 buttons
    const confirmBtn = this.page.getByRole("button", {
      name: /确\s*认|confirm/i,
    });
    await confirmBtn.waitFor({ state: "visible", timeout: 1500 });
    await confirmBtn.click();

    // Wait for redirect to login page
    await this.page.waitForURL(/auth\/login/, { timeout: 10000 });
  }
}
