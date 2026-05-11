import { test, expect } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";

test.describe("TC-150 Server monitor visibility-aware polling", () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, "monitor-server");
  });

  test("TC-150a: hidden tab pauses polling and visible tab refreshes immediately", async ({
    adminPage,
  }) => {
    const installVisibilityMock = () => {
      let state: DocumentVisibilityState = "visible";
      Object.defineProperty(document, "visibilityState", {
        configurable: true,
        get: () => state,
      });
      Object.defineProperty(document, "hidden", {
        configurable: true,
        get: () => state !== "visible",
      });
      (window as any).__setE2EVisibility = (
        nextState: DocumentVisibilityState,
      ) => {
        state = nextState;
        document.dispatchEvent(new Event("visibilitychange"));
      };
    };

    await adminPage.addInitScript(installVisibilityMock);
    await adminPage.evaluate(installVisibilityMock);

    const monitorRequests: string[] = [];
    adminPage.on("request", (request) => {
      if (
        request.method() === "GET" &&
        request.url().includes("/api/v1/monitor/server")
      ) {
        monitorRequests.push(request.url());
      }
    });

    await adminPage.clock.install();
    const initialMonitorResponse = adminPage.waitForResponse(
      (response) =>
        response.request().method() === "GET" &&
        response.url().includes("/api/v1/monitor/server") &&
        response.status() === 200,
    );
    await adminPage.goto("/monitor/server", { waitUntil: "domcontentloaded" });
    await initialMonitorResponse;

    await adminPage.evaluate(() => {
      (window as any).__setE2EVisibility("hidden");
    });
    const requestsAfterHidden = monitorRequests.length;
    await adminPage.clock.runFor(31_000);
    expect(monitorRequests).toHaveLength(requestsAfterHidden);
    await adminPage.clock.resume();

    const visibleResponse = adminPage.waitForResponse(
      (response) =>
        response.request().method() === "GET" &&
        response.url().includes("/api/v1/monitor/server") &&
        response.status() === 200,
    );
    await adminPage.evaluate(() => {
      (window as any).__setE2EVisibility("visible");
    });
    await visibleResponse;

    expect(monitorRequests.length).toBe(requestsAfterHidden + 1);
  });
});
