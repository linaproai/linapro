import { test, expect } from "../../fixtures/auth";

test.describe("TC-144 Management workbench logo dark-mode edge glow", () => {
  test.beforeEach(async ({ adminPage }) => {
    await adminPage.setViewportSize({ width: 1366, height: 900 });
  });

  test("TC-144a: dark mode applies cyan glow to the logo icon edge", async ({
    mainLayout,
  }) => {
    await mainLayout.ensureThemeMode("dark");
    await expect
      .poll(async () => {
        const glow = await mainLayout.getBrandLogoGlowInfo();
        return glow.imageFilter;
      })
      .toContain("drop-shadow");

    const glow = await mainLayout.getBrandLogoGlowInfo();
    expect(glow.isDarkRoot).toBeTruthy();
    expect(glow.beforeBackgroundImage).not.toContain("gradient");
    expect(glow.imageClientWidth).toBeGreaterThan(0);
    expect(glow.imageClientHeight).toBeGreaterThan(0);
    expect(glow.markWidth).toBeGreaterThanOrEqual(30);
    expect(glow.markHeight).toBeLessThanOrEqual(glow.linkHeight);
    expect(glow.linkOverflow).toBe("visible");
  });

  test("TC-144b: light mode keeps the logo without dark-mode glow", async ({
    mainLayout,
  }) => {
    await mainLayout.ensureThemeMode("light");
    await expect
      .poll(async () => {
        const glow = await mainLayout.getBrandLogoGlowInfo();
        return glow.imageFilter;
      })
      .toBe("none");

    const glow = await mainLayout.getBrandLogoGlowInfo();
    expect(glow.isDarkRoot).toBeFalsy();
    expect(glow.beforeBackgroundImage).not.toContain("gradient");
    expect(glow.markWidth).toBeGreaterThanOrEqual(30);
    expect(glow.markHeight).toBeLessThanOrEqual(glow.linkHeight);
    expect(glow.linkOverflow).toBe("visible");
  });
});
