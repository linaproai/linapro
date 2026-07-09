import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { test, expect } from '../../fixtures/auth';
import { workspacePath } from '../../fixtures/config';

const currentDir = dirname(fileURLToPath(import.meta.url));
const openApiMethods = new Set([
  'connect',
  'delete',
  'get',
  'head',
  'options',
  'patch',
  'post',
  'put',
  'trace',
]);
const chineseLocalizedPathPrefixes = [
  '/api/v1/',
];

type OpenApiDocument = {
  paths?: Record<string, Record<string, OpenApiOperation | unknown>>;
};

type OpenApiOperation = {
  summary?: unknown;
  tags?: unknown;
};

type OpenApiTitle = {
  method: string;
  path: string;
  kind: 'summary' | 'tag';
  value: string;
};

test.describe('TC001 系统接口页面', () => {
  test('TC001a: 系统接口页面通过 iframe 加载 Stoplight Elements', async ({
    adminPage,
  }) => {
    await adminPage.goto('/about/api-docs');
    // Verify the iframe is visible
    const iframe = adminPage.locator('iframe.api-docs-iframe');
    await expect(iframe).toBeVisible({ timeout: 10_000 });
    await expect(iframe).toHaveAttribute(
      'src',
      new RegExp(`${escapeRegExp(workspacePath('/stoplight/apidocs.html'))}\\?`),
    );
    // Wait for Stoplight Elements to render inside the iframe
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    const apiElement = frame.locator('elements-api');
    await expect(apiElement).toBeAttached({ timeout: 15_000 });
    // Verify Overview is visible in sidebar
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // Verify ENDPOINTS section is visible
    await expect(
      frame
        .locator('.sl-uppercase.sl-font-bold')
        .filter({ hasText: /^Endpoints$/i }),
    ).toBeVisible();
  });

  test('TC001b: 系统接口页面不污染主页面样式', async ({ adminPage }) => {
    await adminPage.goto('/about/api-docs');
    const iframe = adminPage.locator('iframe.api-docs-iframe');
    await expect(iframe).toBeVisible({ timeout: 10_000 });
    // Main page should not have any Stoplight stylesheets injected
    const stoplightLinks = await adminPage
      .locator('link[href*="stoplight/styles"]')
      .count();
    expect(stoplightLinks).toBe(0);
  });

  test('TC001c: Overview 页面显示 API 标题和描述', async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage('English');
    const apiResponse = await adminPage.request.get('/api.json?lang=en-US', {
      headers: { 'Accept-Language': 'en-US' },
    });
    expect(apiResponse.ok()).toBeTruthy();
    const apiDocument = await apiResponse.json();

    await adminPage.goto('/about/api-docs');
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    // Wait for content to load
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // Verify the right panel shows API title and description
    await expect(
      frame.locator('h1', { hasText: apiDocument.info.title.trim() }),
    ).toBeVisible({ timeout: 10_000 });
    await expect(frame.getByText(apiDocument.info.version)).toBeVisible();
    await expect(frame.getByText(apiDocument.info.description)).toBeVisible();
  });

  test('TC001d: 隐藏 powered by Stoplight 标识', async ({ adminPage }) => {
    await adminPage.goto('/about/api-docs');
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // "powered by Stoplight" link should be hidden
    const poweredBy = frame.locator('a', { hasText: 'Stoplight' });
    await expect(poweredBy).toBeHidden();
  });

  test('TC001e: 模块名称粗体、接口名称非粗体', async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage('简体中文');
    await adminPage.goto('/about/api-docs');
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // Click on a module to expand it
    const moduleItem = frame.locator('[title="身份认证"]');
    await moduleItem.click();
    // Module name should be bold (font-weight 700)
    const moduleText = frame
      .locator('[title="身份认证"] .sl-flex-1')
      .first();
    await expect(moduleText).toHaveCSS('font-weight', '700');
    // Endpoint name should not be bold (font-weight 400)
    // Use "用户登录" endpoint which exists in auth module
    const endpointText = frame
      .locator('[title="用户登录"] .sl-flex-1')
      .first();
    await expect(endpointText).toBeVisible();
    const fontWeight = await endpointText.evaluate(
      (el) => window.getComputedStyle(el).fontWeight,
    );
    expect(fontWeight).toBe('400');
  });

  test('TC001f: 接口地址背景块全宽展示，GET 在左、路径在右', async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage('简体中文');
    await adminPage.goto('/about/api-docs');
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // Expand module and click endpoint
    await frame.locator('[title="用户管理"]').click();
    await frame.locator('[title="获取用户列表"]').click();
    // Find the method/path block
    const pathBlock = frame.locator(
      'div[title*="/api/v1/user"]',
    );
    await expect(pathBlock).toBeVisible({ timeout: 10_000 });
    // Block should be full width (display: flex, not inline-flex)
    await expect(pathBlock).toHaveCSS('display', 'flex');
    await expect(pathBlock).toHaveCSS('width', /.+/);
    await expect(pathBlock).toHaveCSS('justify-content', 'space-between');
  });

  test('TC001g: SCHEMAS 区域默认折叠且可展开', async ({ adminPage }) => {
    await adminPage.goto('/about/api-docs');
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // SCHEMAS header should be visible
    const schemasHeader = frame.locator('.schemas-section-header');
    await expect(schemasHeader).toBeVisible();
    // Schema items should be hidden by default (collapsed)
    const firstSchema = frame
      .locator('.ElementsTableOfContentsItem[href*="/schemas/"]')
      .first();
    await expect(firstSchema).toBeHidden();
    // Click to expand
    await schemasHeader.click();
    await expect(firstSchema).toBeVisible();
    // Click again to collapse
    await schemasHeader.click();
    await expect(firstSchema).toBeHidden();
  });

  test('TC001h: 英文环境下系统接口文档使用英文接口源文案', async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage('English');
    await adminPage.goto('/about/api-docs');

    const iframe = adminPage.locator('iframe.api-docs-iframe');
    await expect(iframe).toBeVisible({ timeout: 10_000 });
    await expect(iframe).toHaveAttribute('src', /lang=en-US/);

    const apiResponse = await adminPage.request.get('/api.json?lang=en-US', {
      headers: { 'Accept-Language': 'en-US' },
    });
    expect(apiResponse.ok()).toBeTruthy();
    const apiDocument = await apiResponse.text();
    expect(apiDocument).toContain('"User Management"');
    expect(apiDocument).toContain('"Get user list"');
    expect(apiDocument).toContain('"Page number"');

    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    await frame.locator('[title="User Management"]').click();
    await expect(
      frame.locator('[title="Get user list"]').first(),
    ).toBeVisible({ timeout: 10_000 });
    await expect(frame.locator('[title="用户管理"]')).toHaveCount(0);
  });

  test('TC001i: 中文接口文档加载拆分层级 apidoc 资源和公共 fallback', async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage('简体中文');

    const apiResponse = await adminPage.request.get('/api.json?lang=zh-CN', {
      headers: { 'Accept-Language': 'zh-CN' },
    });
    expect(apiResponse.ok()).toBeTruthy();
    const apiDocument = (await apiResponse.json()) as OpenApiDocument;
    const apiDocumentText = JSON.stringify(apiDocument);

    expect(apiDocumentText).toContain('"用户登录"');
    expect(apiDocumentText).toContain('"身份认证"');
    expect(apiDocumentText).toContain('"错误码"');
    expect(apiDocumentText).toContain('"错误消息"');
    expect(apiDocumentText).toContain('"按接口定义返回的结果数据"');

    const englishFallbacks = collectUnlocalizedChineseOperationTitles(
      apiDocument,
    );
    expect(
      englishFallbacks,
      `中文接口文档仍存在英文标题：\n${formatOpenApiTitles(englishFallbacks)}`,
    ).toHaveLength(0);
    expectOperationTitle(apiDocument, 'GET', '/api/v1/user', {
      tag: '用户管理',
      summary: '获取用户列表',
    });

    await adminPage.goto('/about/api-docs');
    const iframe = adminPage.locator('iframe.api-docs-iframe');
    await expect(iframe).toBeVisible({ timeout: 10_000 });
    await expect(iframe).toHaveAttribute('src', /lang=zh-CN/);

    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    await frame.locator('[title="身份认证"]').click();
    await expect(frame.locator('[title="用户登录"]').first()).toBeVisible({
      timeout: 10_000,
    });
  });

  test('TC001j: 接口文档内容加载期间显示 iframe 内 Loading', async ({
    adminPage,
  }) => {
    let releaseApiResponse!: () => void;
    let markApiRequestStarted!: () => void;
    let released = false;
    const apiRequestStarted = new Promise<void>((resolve) => {
      markApiRequestStarted = resolve;
    });
    const apiResponseRelease = new Promise<void>((resolve) => {
      releaseApiResponse = () => {
        if (released) return;
        released = true;
        resolve();
      };
    });
    const apiDocsHtml = readFileSync(
      resolve(
        currentDir,
        '../../../../apps/lina-vben/apps/web-antd/public/stoplight/apidocs.html',
      ),
      'utf8',
    );
    const delayedApiDocument = {
      openapi: '3.0.0',
      info: {
        title: 'Loading Test API',
        version: 'v-test',
        description: 'Minimal OpenAPI document for loading state verification',
      },
      paths: {
        '/api/v1/loading-test': {
          get: {
            tags: ['Loading Test'],
            summary: 'Loading test endpoint',
            responses: {
              '200': {
                description: 'OK',
              },
            },
          },
        },
      },
    };

    await adminPage.route('**/stoplight/apidocs.html?**', async (route) => {
      await route.fulfill({
        contentType: 'text/html; charset=utf-8',
        body: apiDocsHtml,
      });
    });

    await adminPage.route('**/api.json?**', async (route) => {
      markApiRequestStarted();
      await apiResponseRelease;
      await route.fulfill({
        contentType: 'application/json; charset=utf-8',
        body: JSON.stringify(delayedApiDocument),
      });
    });

    try {
      await adminPage.goto('/about/api-docs', { waitUntil: 'domcontentloaded' });
      await Promise.race([
        apiRequestStarted,
        adminPage.waitForTimeout(5_000).then(() => {
          throw new Error('Expected Stoplight iframe to request /api.json');
        }),
      ]);

      const frame = adminPage.frameLocator('iframe.api-docs-iframe');
      const loading = frame.locator('#api-docs-loading');
      await expect(frame.locator('body')).toBeAttached({ timeout: 5_000 });
      await expect(loading).toBeVisible({ timeout: 10_000 });
      await expect(frame.locator('#api-docs-loading-title')).toContainText(
        /接口文档加载中|Loading API documentation/,
      );

      releaseApiResponse();
      await expect(frame.getByText('Overview')).toBeVisible({
        timeout: 15_000,
      });
      await expect(loading).toBeHidden({ timeout: 10_000 });
    } finally {
      releaseApiResponse();
      await adminPage.unroute('**/api.json?**');
      await adminPage.unroute('**/stoplight/apidocs.html?**');
    }
  });
});

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function collectUnlocalizedChineseOperationTitles(apiDocument: OpenApiDocument) {
  const titles: OpenApiTitle[] = [];
  for (const [path, pathItem] of Object.entries(apiDocument.paths ?? {})) {
    if (!shouldRequireChineseOperationTitle(path)) {
      continue;
    }
    for (const [method, operation] of Object.entries(pathItem)) {
      if (!openApiMethods.has(method.toLowerCase())) {
        continue;
      }
      const typedOperation = operation as OpenApiOperation;
      const summary = stringValue(typedOperation.summary);
      if (summary !== '' && !containsCjk(summary)) {
        titles.push({
          method: method.toUpperCase(),
          path,
          kind: 'summary',
          value: summary,
        });
      }
      for (const tag of arrayStringValues(typedOperation.tags)) {
        if (tag !== '' && !containsCjk(tag)) {
          titles.push({
            method: method.toUpperCase(),
            path,
            kind: 'tag',
            value: tag,
          });
        }
      }
    }
  }
  return titles;
}

function shouldRequireChineseOperationTitle(path: string) {
  return chineseLocalizedPathPrefixes.some((prefix) => path.startsWith(prefix));
}

function expectOperationTitle(
  apiDocument: OpenApiDocument,
  method: string,
  path: string,
  expected: { tag: string; summary: string },
) {
  const operation = apiDocument.paths?.[path]?.[
    method.toLowerCase()
  ] as OpenApiOperation | undefined;
  expect(operation, `缺少接口定义：${method} ${path}`).toBeTruthy();
  expect(arrayStringValues(operation?.tags)[0]).toBe(expected.tag);
  expect(stringValue(operation?.summary)).toBe(expected.summary);
}

function stringValue(value: unknown) {
  return typeof value === 'string' ? value.trim() : '';
}

function arrayStringValues(value: unknown) {
  return Array.isArray(value)
    ? value.filter((item): item is string => typeof item === 'string')
    : [];
}

function containsCjk(value: string) {
  return /[\u3400-\u9fff]/u.test(value);
}

function formatOpenApiTitles(titles: OpenApiTitle[]) {
  return titles
    .slice(0, 50)
    .map((item) => `${item.method} ${item.path} ${item.kind}: ${item.value}`)
    .join('\n');
}
