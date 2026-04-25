import type { Locator } from '@playwright/test';

import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';
import { LayoutAuditPage } from '../../pages/LayoutAuditPage';
import { ProfilePage } from '../../pages/ProfilePage';
import { UserPage } from '../../pages/UserPage';

async function readLineMetrics(locator: Locator) {
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

async function expectSingleLine(locator: Locator, label: string) {
  await expect(locator, `${label} should be visible`).toBeVisible();
  const metrics = await readLineMetrics(locator);
  expect(
    metrics.height,
    `${label} wraps unexpectedly`,
  ).toBeLessThanOrEqual(metrics.lineHeight * 1.6);
}

async function expectNoHorizontalClip(locator: Locator, label: string) {
  await expect(locator, `${label} should be visible`).toBeVisible();
  const metrics = await readLineMetrics(locator);
  expect(
    metrics.scrollWidth,
    `${label} is still clipped in the current layout`,
  ).toBeLessThanOrEqual(metrics.width + 2);
}

async function expectPlaceholderFits(locator: Locator, label: string) {
  await expect(locator, `${label} should be visible`).toBeVisible();
  const metrics = await locator.evaluate((node) => {
    const input = node as HTMLInputElement;
    const style = window.getComputedStyle(input);
    const canvas = document.createElement('canvas');
    const context = canvas.getContext('2d');
    const fontStyle = style.fontStyle || 'normal';
    const fontVariant = style.fontVariant || 'normal';
    const fontWeight = style.fontWeight || '400';
    const fontSize = style.fontSize || '14px';
    const fontFamily = style.fontFamily || 'sans-serif';
    const paddingLeft = Number.parseFloat(style.paddingLeft || '0') || 0;
    const paddingRight = Number.parseFloat(style.paddingRight || '0') || 0;

    if (context) {
      context.font =
        `${fontStyle} ${fontVariant} ${fontWeight} ${fontSize} ${fontFamily}`;
    }

    return {
      availableWidth: input.clientWidth - paddingLeft - paddingRight,
      placeholderWidth: context
        ? context.measureText(input.placeholder).width
        : Number.POSITIVE_INFINITY,
    };
  });

  expect(
    metrics.placeholderWidth,
    `${label} placeholder is still clipped in the current layout`,
  ).toBeLessThanOrEqual(metrics.availableWidth + 2);
}

async function expectHeaderSingleLine(locator: Locator, label: string) {
  await expect(locator, `${label} should be visible`).toBeVisible();
  const height = await locator.evaluate((node) => {
    return (node as HTMLElement).getBoundingClientRect().height;
  });
  expect(height, `${label} should stay within a single header row`).toBeLessThanOrEqual(48);
}

async function readWidth(locator: Locator) {
  await expect(locator).toBeVisible();
  return await locator.evaluate((node) => {
    return (node as HTMLElement).getBoundingClientRect().width;
  });
}

test.describe('TC0112 英文布局回归', () => {
  test.beforeEach(async ({ adminPage, mainLayout }) => {
    await adminPage.setViewportSize({ width: 1366, height: 900 });
    await mainLayout.switchLanguage('English');
  });

  test('TC-112a: 侧栏、页签与个人中心表单在英文环境下保持单行可读', async ({
    adminPage,
    mainLayout,
  }) => {
    const profilePage = new ProfilePage(adminPage);
    const layoutPage = new LayoutAuditPage(adminPage);

    const orgMenu = mainLayout.sidebarMenuItem('Organization');
    await orgMenu.scrollIntoViewIfNeeded();
    await expectNoHorizontalClip(orgMenu, 'Organization sidebar item');

    const sidebarBox = await mainLayout.sidebar.boundingBox();
    expect(sidebarBox).not.toBeNull();
    expect(sidebarBox!.width).toBeGreaterThanOrEqual(236);

    const dynamicDemoMenu = mainLayout.sidebarMenuItem('Dynamic Plugin Demo');
    await dynamicDemoMenu.scrollIntoViewIfNeeded();
    await expectNoHorizontalClip(dynamicDemoMenu, 'Dynamic Plugin Demo sidebar item');

    await profilePage.goto();

    await expectSingleLine(
      layoutPage.formLabel(/^Phone$/),
      'Profile phone label',
    );
    await expectPlaceholderFits(
      adminPage.getByPlaceholder(/Please enter a nickname/i).first(),
      'Profile nickname placeholder',
    );
    await expectPlaceholderFits(
      adminPage.getByPlaceholder(/Please enter an email address/i).first(),
      'Profile email placeholder',
    );
    await expectPlaceholderFits(
      adminPage.getByPlaceholder(/Please enter a phone number/i).first(),
      'Profile phone placeholder',
    );
    const baseFormWidth = await readWidth(
      adminPage.getByTestId('profile-base-form'),
    );
    const baseNicknameInputWidth = await readWidth(
      adminPage.getByPlaceholder(/Please enter a nickname/i).first(),
    );

    await profilePage.openPasswordTab();
    await expectSingleLine(
      adminPage.getByText(/Current Password/i).first(),
      'Profile current password label',
    );
    await expectSingleLine(
      adminPage.getByText(/Confirm Password/i).first(),
      'Profile confirm password label',
    );
    await expectPlaceholderFits(
      adminPage.getByPlaceholder(/Please enter the current password/i).first(),
      'Profile current password placeholder',
    );
    await expectPlaceholderFits(
      adminPage.getByPlaceholder(/Please enter a new password/i).first(),
      'Profile new password placeholder',
    );
    await expectPlaceholderFits(
      adminPage.getByPlaceholder(/Please confirm the new password/i).first(),
      'Profile confirm password placeholder',
    );
    const passwordFormWidth = await readWidth(
      adminPage.getByTestId('profile-password-form'),
    );
    const passwordCurrentInputWidth = await readWidth(
      adminPage.getByPlaceholder(/Please enter the current password/i).first(),
    );
    expect(
      Math.abs(baseFormWidth - passwordFormWidth),
      'Profile base/password form containers should keep the same width',
    ).toBeLessThanOrEqual(2);
    expect(
      Math.abs(baseNicknameInputWidth - passwordCurrentInputWidth),
      'Profile base/password input fields should keep the same width',
    ).toBeLessThanOrEqual(2);

    const activeTabTitle = mainLayout.activeTabTitle();
    const tabText = (await activeTabTitle.textContent())?.trim() || '';
    expect(tabText).not.toBe('');
    await expect(activeTabTitle).toHaveAttribute('title', tabText);
  });

  test('TC-112b: 用户、字典与调度日志页在英文环境下保留稳定的搜索与表头布局', async ({
    adminPage,
  }) => {
    const dictPage = new DictPage(adminPage);
    const layoutPage = new LayoutAuditPage(adminPage);
    const userPage = new UserPage(adminPage);

    await userPage.goto();
    await expectSingleLine(
      layoutPage.formLabel(/^User Account$/),
      'User search label: User Account',
    );
    await expectSingleLine(
      layoutPage.formLabel(/^Phone$/),
      'User search label: Phone',
    );

    const deptTree = adminPage.locator('.ant-tree').first();
    if (await deptTree.isVisible().catch(() => false)) {
      const deptTreeBox = await deptTree.boundingBox();
      const gridBox = await adminPage.locator('.vxe-grid').first().boundingBox();
      expect(deptTreeBox).not.toBeNull();
      expect(gridBox).not.toBeNull();
      expect(
        gridBox!.y,
        'User page grid should stack below the department tree at 1366px English layout',
      ).toBeGreaterThan(deptTreeBox!.y + deptTreeBox!.height - 1);
    }

    await dictPage.goto();
    const dictTypePanel = layoutPage.panel('dict-type');
    const dictDataPanel = layoutPage.panel('dict-data');
    const dictTypeBox = await dictTypePanel.boundingBox();
    const dictDataBox = await dictDataPanel.boundingBox();
    expect(dictTypeBox).not.toBeNull();
    expect(dictDataBox).not.toBeNull();
    expect(
      dictDataBox!.y,
      'Dictionary data panel should stack under the type panel at 1366px English layout',
    ).toBeGreaterThan(dictTypeBox!.y + dictTypeBox!.height - 1);
    await expectHeaderSingleLine(
      layoutPage.tableHeader(/Dictionary Name/i, dictTypePanel),
      'Dictionary type header',
    );
    await expectHeaderSingleLine(
      layoutPage.tableHeader(/Dictionary Label/i, dictDataPanel),
      'Dictionary data header',
    );

    await layoutPage.goto('/system/job-log', { tableSelector: '.vxe-table' });
    await expectSingleLine(
      layoutPage.formLabel(/^Node$/),
      'Job log search label: Node',
    );
    await expectHeaderSingleLine(
      layoutPage.tableHeader(/^Status$/),
      'Job log status header',
    );
  });
});
