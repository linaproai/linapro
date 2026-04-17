import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../fixtures/auth';
import { config } from '../../fixtures/config';
import { ConfigPage } from '../../pages/ConfigPage';
import { LoginPage } from '../../pages/LoginPage';

const publicFrontendParams = [
  { key: 'sys.app.name', name: '品牌展示-应用名称' },
  { key: 'sys.app.logo', name: '品牌展示-应用 Logo' },
  { key: 'sys.app.logoDark', name: '品牌展示-深色 Logo' },
  { key: 'sys.auth.pageTitle', name: '登录展示-页面标题' },
  { key: 'sys.auth.pageDesc', name: '登录展示-页面说明' },
  { key: 'sys.auth.loginSubtitle', name: '登录展示-登录副标题' },
  { key: 'sys.ui.theme.mode', name: '界面风格-主题模式' },
  { key: 'sys.ui.layout', name: '界面风格-工作台布局' },
  { key: 'sys.ui.watermark.enabled', name: '界面风格-是否启用水印' },
  { key: 'sys.ui.watermark.content', name: '界面风格-水印文案' },
];

async function loginAsAdmin(request: APIRequestContext): Promise<string> {
  const response = await request.post('/api/v1/auth/login', {
    data: {
      password: config.adminPass,
      username: config.adminUser,
    },
  });
  expect(response.ok()).toBeTruthy();

  const payload = await response.json();
  expect(payload.code).toBe(0);
  return payload.data.accessToken as string;
}

async function getConfigByKey(
  request: APIRequestContext,
  accessToken: string,
  key: string,
) {
  const response = await request.get(
    `/api/v1/config/key/${encodeURIComponent(key)}`,
    {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    },
  );
  expect(response.ok()).toBeTruthy();

  const payload = await response.json();
  expect(payload.code).toBe(0);
  return payload.data as {
    id: number;
    key: string;
    value: string;
  };
}

async function updateConfigValue(
  request: APIRequestContext,
  accessToken: string,
  id: number,
  value: string,
) {
  const response = await request.put(`/api/v1/config/${id}`, {
    data: { value },
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
  expect(response.ok()).toBeTruthy();

  const payload = await response.json();
  expect(payload.code).toBe(0);
}

test.describe('TC0080 公开前端配置系统参数', () => {
  test('TC0080a: 参数设置页可检索到公开前端配置内置参数', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    for (const item of publicFrontendParams) {
      await configPage.fillSearchField('参数键名', item.key);
      await configPage.clickSearch();

      const targetRow = configPage.findRowByExactKey(item.key);
      await expect(targetRow).toBeVisible();
      await expect(targetRow).toContainText(item.name);
    }
  });

  test('TC0080b: 登录页和主题可消费公开前端配置', async ({ page, request }) => {
    const accessToken = await loginAsAdmin(request);
    const overrides = {
      'sys.app.name': `LinaPro 品牌测试 ${Date.now()}`,
      'sys.auth.pageTitle': '统一品牌登录入口',
      'sys.auth.pageDesc': '宿主工作台与插件能力统一从系统参数读取展示信息',
      'sys.auth.loginSubtitle': '请使用管理员账号登录当前宿主工作区',
      'sys.ui.theme.mode': 'dark',
      'sys.ui.layout': 'header-nav',
      'sys.ui.watermark.enabled': 'true',
      'sys.ui.watermark.content': '品牌测试水印',
    } as const;

    const originals = await Promise.all(
      Object.keys(overrides).map(async (key) => {
        return await getConfigByKey(request, accessToken, key);
      }),
    );

    try {
      for (const original of originals) {
        await updateConfigValue(
          request,
          accessToken,
          original.id,
          overrides[original.key as keyof typeof overrides],
        );
      }

      const publicResponse = await request.get('/api/v1/config/public/frontend');
      expect(publicResponse.ok()).toBeTruthy();
      const publicPayload = await publicResponse.json();
      expect(publicPayload.code).toBe(0);
      expect(publicPayload.data.app.name).toBe(overrides['sys.app.name']);
      expect(publicPayload.data.auth.pageTitle).toBe(overrides['sys.auth.pageTitle']);
      expect(publicPayload.data.ui.themeMode).toBe('dark');
      expect(publicPayload.data.ui.layout).toBe('header-nav');
      expect(publicPayload.data.ui.watermarkEnabled).toBe(true);
      expect(publicPayload.data.ui.watermarkContent).toBe('品牌测试水印');

      const loginPage = new LoginPage(page);
      await loginPage.goto();

      await expect(loginPage.getText(overrides['sys.auth.pageTitle'])).toBeVisible();
      await expect(loginPage.getText(overrides['sys.auth.pageDesc'])).toBeVisible();
      await expect(
        loginPage.getText(overrides['sys.auth.loginSubtitle']),
      ).toBeVisible();
      await expect(loginPage.getText(overrides['sys.app.name'])).toBeVisible();
      await expect
        .poll(async () => await loginPage.getDocumentTitle())
        .toContain(overrides['sys.app.name']);
      await expect
        .poll(async () =>
          page.evaluate(() => document.documentElement.classList.contains('dark')),
        )
        .toBe(true);
      await expect
        .poll(async () =>
          page.evaluate(() => document.documentElement.dataset.theme || ''),
        )
        .toBe('default');

      await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
      await expect
        .poll(async () =>
          page.evaluate(() => document.documentElement.classList.contains('dark')),
        )
        .toBe(true);
    } finally {
      for (const original of originals) {
        await updateConfigValue(request, accessToken, original.id, original.value);
      }
    }
  });

  test('TC0080c: 同一浏览器重新访问时会拉取最新的后台主题配置', async ({
    page,
    request,
  }) => {
    const accessToken = await loginAsAdmin(request);
    const original = await getConfigByKey(request, accessToken, 'sys.ui.theme.mode');
    const loginPage = new LoginPage(page);

    try {
      await updateConfigValue(request, accessToken, original.id, 'light');

      await loginPage.goto();
      await expect
        .poll(async () =>
          page.evaluate(() => document.documentElement.classList.contains('dark')),
        )
        .toBe(false);

      await updateConfigValue(request, accessToken, original.id, 'dark');

      await loginPage.goto();
      await expect
        .poll(async () =>
          page.evaluate(() => document.documentElement.classList.contains('dark')),
        )
        .toBe(true);
    } finally {
      await updateConfigValue(request, accessToken, original.id, original.value);
    }
  });
});
