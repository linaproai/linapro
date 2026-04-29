import { test, expect } from '../../fixtures/auth';

test.describe('TC0045 版本信息页面', () => {
  test('TC0045a: 版本信息页面显示三个区块', async ({ adminPage }) => {
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

  test('TC0045b: 后端组件从配置文件动态加载', async ({ adminPage }) => {
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
      content.getByText('MySQL', { exact: true }),
    ).toBeVisible();
    await expect(
      content.getByText('JWT', { exact: true }),
    ).toBeVisible();
    await expect(
      content.getByText('Excelize', { exact: true }),
    ).toBeVisible();
  });

  test('TC0045c: 页面顶部不显示标题栏和版本信息介绍板块', async ({ adminPage }) => {
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

  test('TC0045d: 关于项目区块展示官网和仓库地址', async ({ adminPage }) => {
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
});
