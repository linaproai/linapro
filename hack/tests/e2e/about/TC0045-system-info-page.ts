import type { Page, Route } from '@playwright/test';

import { test, expect } from '../../fixtures/auth';

const longPostgresVersion =
  'PostgreSQL 16.8 (Homebrew) on arm64-apple-darwin24.4.0, compiled by Apple clang version 16.0.0 (clang-1600.0.26.6), 64-bit';

async function mockLongPostgreSqlVersion(page: Page) {
  await page.route('**/api/v1/system/info**', async (route: Route) => {
    const response = await route.fetch();
    const payload = await response.json();
    const data = payload?.data ?? payload;
    const backendComponents = data?.backendComponents ?? [];
    const postgresql = backendComponents.find(
      (component: { name?: string }) => component?.name === 'PostgreSQL',
    );
    if (postgresql) {
      postgresql.version = longPostgresVersion;
    }
    if (data) {
      data.dbVersion = longPostgresVersion;
    }

    await route.fulfill({ json: payload, response });
  });
}

test.describe('TC-45 版本信息页面', () => {
  test('TC-45a: 版本信息页面显示三个区块', async ({ adminPage }) => {
    await adminPage.goto('/about/system-info');
    await adminPage.waitForLoadState('networkidle');

    const content = adminPage.locator('[id="__vben_main_content"]');

    // 关于项目区块 - 第一行：项目名称 + 项目介绍
    await expect(content.getByText('关于项目')).toBeVisible();
    await expect(content.getByText('项目名称')).toBeVisible();
    await expect(content.getByText('项目介绍')).toBeVisible();
    await expect(
      content.getByText('面向可持续交付的 AI 原生全栈框架。', {
        exact: true,
      }),
    ).toBeVisible();

    // 关于项目区块 - 第二行：版本号、开源许可、项目官网、仓库地址
    await expect(content.getByText('版本号')).toBeVisible();
    await expect(content.getByText('开源许可')).toBeVisible();
    await expect(content.getByText('项目官网')).toBeVisible();
    await expect(content.getByText('仓库地址')).toBeVisible();

    // 后端组件区块（从 API 加载）
    await expect(content.getByText('后端组件')).toBeVisible();
    await expect(
      content.getByText('GoFrame', { exact: true }),
    ).toBeVisible({ timeout: 10_000 });

    // 前端组件区块（从 API 加载）
    await expect(content.getByText('前端组件')).toBeVisible();
    await expect(
      content.getByText('Vue', { exact: true }).first(),
    ).toBeVisible({ timeout: 10_000 });
  });

  test('TC-45b: 后端组件从配置文件动态加载', async ({ adminPage }) => {
    const systemInfoResponsePromise = adminPage.waitForResponse(
      (response) =>
        response.request().method() === 'GET' &&
        response.url().includes('/api/v1/system/info') &&
        response.ok(),
    );

    await adminPage.goto('/about/system-info');
    await adminPage.waitForLoadState('networkidle');

    const content = adminPage.locator('[id="__vben_main_content"]');
    const systemInfoPayload = await (await systemInfoResponsePromise).json();
    const backendComponents =
      systemInfoPayload?.data?.backendComponents ??
      systemInfoPayload?.backendComponents ??
      [];
    const goframeComponent = backendComponents.find(
      (component: { description?: string; name?: string }) =>
        component?.name === 'GoFrame',
    );

    // 后端组件应从 API 加载，包含关键组件
    await expect(
      content.getByText('GoFrame', { exact: true }),
    ).toBeVisible({ timeout: 10_000 });
    expect(goframeComponent?.description).toBeTruthy();
    await expect(
      content.getByText(goframeComponent!.description!, { exact: true }),
    ).toBeVisible();
    await expect(
      content.getByText('PostgreSQL', { exact: true }),
    ).toBeVisible();
    await expect(
      content.getByText('JWT', { exact: true }),
    ).toBeVisible();
    await expect(
      content.getByText('Excelize', { exact: true }),
    ).toBeVisible();
  });

  test('TC-45c: 页面顶部不显示标题栏和版本信息介绍板块', async ({ adminPage }) => {
    await adminPage.goto('/about/system-info');
    await adminPage.waitForLoadState('networkidle');

    const content = adminPage.locator('[id="__vben_main_content"]');

    // 页面顶部不应有"版本信息"标题栏（Page 组件的 title）
    const pageHeader = content.locator('.page-header, header').first();
    await expect(pageHeader).not.toBeVisible();

    // 第一个 card-box 应直接是"关于项目"
    const firstCard = content.locator('.card-box').first();
    await expect(firstCard.getByText('关于项目')).toBeVisible();
  });

  test('TC-45d: 关于项目区块展示官网和仓库地址', async ({ adminPage }) => {
    const systemInfoResponsePromise = adminPage.waitForResponse(
      (response) =>
        response.request().method() === 'GET' &&
        response.url().includes('/api/v1/system/info') &&
        response.ok(),
    );

    await adminPage.goto('/about/system-info');
    await adminPage.waitForLoadState('networkidle');

    const content = adminPage.locator('[id="__vben_main_content"]');
    const aboutCard = content.locator('.card-box').first();
    const systemInfoPayload = await (await systemInfoResponsePromise).json();
    const framework =
      systemInfoPayload?.data?.framework ??
      systemInfoPayload?.framework ??
      {};

    await expect(aboutCard.getByText('项目官网')).toBeVisible();
    await expect(aboutCard.getByText('仓库地址')).toBeVisible();
    await expect(aboutCard.getByRole('link', { name: '点击查看' }).first()).toHaveAttribute(
      'href',
      framework.homepage,
    );
    await expect(aboutCard.getByRole('link', { name: '点击查看' }).nth(1)).toHaveAttribute(
      'href',
      framework.repositoryUrl,
    );
  });

  test('TC-45e: PostgreSQL 长版本信息单行省略且保留完整内容', async ({
    adminPage,
  }) => {
    await mockLongPostgreSqlVersion(adminPage);
    await adminPage.setViewportSize({ height: 900, width: 1440 });

    await adminPage.goto('/about/system-info');
    await adminPage.waitForLoadState('networkidle');

    const postgresqlItem = adminPage.getByTestId(
      'system-info-component-postgresql',
    );
    const jwtItem = adminPage.getByTestId('system-info-component-jwt');
    const versionText = adminPage.getByTestId(
      'system-info-component-version-postgresql',
    );

    await expect(versionText).toHaveAttribute('title', longPostgresVersion);
    await expect(versionText).toHaveCSS('overflow', 'hidden');
    await expect(versionText).toHaveCSS('text-overflow', 'ellipsis');
    await expect(versionText).toHaveCSS('white-space', 'nowrap');

    const box = await versionText.boundingBox();
    expect(box?.height).toBeLessThanOrEqual(24);

    const postgresqlBox = await postgresqlItem.boundingBox();
    const jwtBox = await jwtItem.boundingBox();
    expect(
      postgresqlBox && jwtBox ? jwtBox.x - (postgresqlBox.x + postgresqlBox.width) : 0,
    ).toBeGreaterThanOrEqual(20);

    await adminPage.setViewportSize({ height: 844, width: 390 });
    await adminPage.reload({ waitUntil: 'domcontentloaded' });
    await adminPage.waitForLoadState('networkidle');

    const mobileVersion = adminPage.getByTestId(
      'system-info-component-version-postgresql',
    );
    const mobileLink = postgresqlItem.getByRole('link', { name: '关系型数据库' });
    const mobileVersionBox = await mobileVersion.boundingBox();
    const mobileLinkBox = await mobileLink.boundingBox();

    expect(mobileVersionBox?.width).toBeGreaterThanOrEqual(100);
    expect(
      mobileVersionBox && mobileLinkBox ? mobileLinkBox.y - mobileVersionBox.y : 0,
    ).toBeGreaterThanOrEqual(18);
  });
});
