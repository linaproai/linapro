import type { Locator } from '@playwright/test';

import { test, expect } from '../../fixtures/auth';

async function readTextMetrics(locator: Locator) {
  return await locator.evaluate((node) => {
    const element = node as HTMLElement;
    const style = window.getComputedStyle(element);
    const fontSize = Number.parseFloat(style.fontSize || '16') || 16;
    const rawLineHeight = Number.parseFloat(style.lineHeight || '');
    const lineHeight =
      Number.isFinite(rawLineHeight) && rawLineHeight > 0
        ? rawLineHeight
        : fontSize * 1.2;

    return {
      height: element.getBoundingClientRect().height,
      lineHeight,
      scrollWidth: element.scrollWidth,
      width: element.clientWidth,
    };
  });
}

async function expectSubtitleBelowTitle(title: Locator, subtitle: Locator) {
  const titleBox = await title.boundingBox();
  const subtitleBox = await subtitle.boundingBox();

  expect(titleBox).not.toBeNull();
  expect(subtitleBox).not.toBeNull();
  expect(
    subtitleBox!.y,
    'Preferences subtitle should sit below the title instead of beside it',
  ).toBeGreaterThanOrEqual(titleBox!.y + titleBox!.height);
  expect(
    Math.abs(subtitleBox!.x - titleBox!.x),
    'Preferences subtitle should align with the title start',
  ).toBeLessThanOrEqual(8);
}

test.describe('TC-118 Preferences drawer header', () => {
  test.beforeEach(async ({ adminPage }) => {
    await adminPage.setViewportSize({ width: 1366, height: 900 });
  });

  test('TC-118a: Chinese preferences subtitle is separated from the title', async ({
    mainLayout,
  }) => {
    await mainLayout.switchLanguage('简体中文');
    await mainLayout.openPreferences();

    await expect(mainLayout.preferencesDrawerTitle).toHaveText('偏好设置');
    await expect(mainLayout.preferencesDrawerSubtitle).toHaveText(
      '调整后即时预览',
    );
    await expectSubtitleBelowTitle(
      mainLayout.preferencesDrawerTitle,
      mainLayout.preferencesDrawerSubtitle,
    );
  });

  test('TC-118b: English preferences subtitle stays compact and single-line', async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage('English');
    await mainLayout.openPreferences();

    await expect(mainLayout.preferencesDrawerTitle).toHaveText('Preferences');
    await expect(mainLayout.preferencesDrawerSubtitle).toHaveText('Live preview');
    await expect(
      adminPage.getByText('Customize Preferences & Preview in Real Time'),
    ).toHaveCount(0);
    await expectSubtitleBelowTitle(
      mainLayout.preferencesDrawerTitle,
      mainLayout.preferencesDrawerSubtitle,
    );

    const metrics = await readTextMetrics(mainLayout.preferencesDrawerSubtitle);
    expect(
      metrics.height,
      'English preferences subtitle should remain on a single line',
    ).toBeLessThanOrEqual(metrics.lineHeight * 1.4);
    expect(
      metrics.scrollWidth,
      'English preferences subtitle should fit within the drawer header',
    ).toBeLessThanOrEqual(metrics.width + 2);
  });
});
