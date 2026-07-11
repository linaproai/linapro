import type { Locator, Page } from '@playwright/test';

import { test, expect } from '../../fixtures/auth';
import { prepareSourcePluginsBaseline } from '../../fixtures/plugin';
import { ConfigPage } from '../../pages/ConfigPage';
import { workspacePath } from '../../fixtures/config';
import { waitForRouteReady } from '../../support/ui';

async function readBorderRadius(locator: Locator): Promise<string> {
  await expect(locator).toBeVisible();
  return locator.evaluate((el) => getComputedStyle(el).borderRadius);
}

async function readThemeRadiusXl(page: Page): Promise<string> {
  return page.evaluate(() => {
    const root = document.documentElement;
    const radius = getComputedStyle(root).getPropertyValue('--radius').trim();
    // Mirror design token: --radius-xl = calc(var(--radius) + 4px)
    const probe = document.createElement('div');
    probe.style.borderRadius = `calc(${radius || '0.5rem'} + 4px)`;
    probe.style.position = 'absolute';
    probe.style.visibility = 'hidden';
    document.body.appendChild(probe);
    const resolved = getComputedStyle(probe).borderRadius;
    probe.remove();
    return resolved;
  });
}

async function readCardBoxRadius(page: Page): Promise<string> {
  await page.goto('/about/system-info');
  await waitForRouteReady(page);
  return readBorderRadius(
    page.locator('[id="__vben_main_content"] .card-box').first(),
  );
}

function gridPanel(page: Page) {
  return page
    .locator('[id="__vben_main_content"] .vxe-grid')
    .first()
    .locator('xpath=..');
}

test.describe('TC-4 后台页面板块圆角统一', () => {
  test.beforeAll(async () => {
    await prepareSourcePluginsBaseline(['linapro-ai-core']);
  });

  test('TC-4a: 版本信息 card-box 使用主题 radius-xl 且非尖角', async ({
    adminPage,
  }) => {
    await adminPage.goto('/about/system-info');
    await waitForRouteReady(adminPage);

    const content = adminPage.locator('[id="__vben_main_content"]');
    const firstCard = content.locator('.card-box').first();
    const radius = await readBorderRadius(firstCard);
    const expected = await readThemeRadiusXl(adminPage);

    expect(radius).not.toBe('0px');
    expect(radius).toBe(expected);
  });

  test('TC-4b: 列表页表格板块与 card-box 使用同一圆角', async ({
    adminPage,
  }) => {
    const cardBoxRadius = await readCardBoxRadius(adminPage);

    const configPage = new ConfigPage(adminPage);
    await configPage.goto();
    const gridRadius = await readBorderRadius(gridPanel(adminPage));

    expect(gridRadius).not.toBe('0px');
    expect(gridRadius).toBe(cardBoxRadius);
  });

  test('TC-4c: Ant Card 表单板块与 card-box 使用同一圆角', async ({
    adminPage,
  }) => {
    const cardBoxRadius = await readCardBoxRadius(adminPage);

    await adminPage.goto('/system/message');
    await waitForRouteReady(adminPage);
    const antCard = adminPage
      .locator('[id="__vben_main_content"] .ant-card')
      .first();
    const antCardRadius = await readBorderRadius(antCard);

    expect(antCardRadius).not.toBe('0px');
    expect(antCardRadius).toBe(cardBoxRadius);
  });

  test('TC-4d: 智能中心调用日志板块与 card-box 使用同一圆角', async ({
    adminPage,
  }) => {
    const cardBoxRadius = await readCardBoxRadius(adminPage);

    await adminPage.goto(workspacePath('/ai/invocations'));
    await waitForRouteReady(adminPage);

    const invocationPanel = adminPage
      .getByTestId('ai-invocation-logs-panel')
      .or(gridPanel(adminPage))
      .first();
    const invocationRadius = await readBorderRadius(invocationPanel);

    expect(invocationRadius).not.toBe('0px');
    expect(invocationRadius).toBe(cardBoxRadius);
  });
});
