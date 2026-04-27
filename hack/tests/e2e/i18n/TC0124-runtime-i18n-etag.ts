import { test, expect } from '../../fixtures/auth';

const runtimeMessagesPath = '/api/v1/i18n/runtime/messages';

test.describe('TC0124 运行时翻译包 ETag 协商', () => {
  test('TC-124a: 首次请求返回 ETag 与 Cache-Control 头', async ({
    adminPage,
  }) => {
    const response = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: { 'Accept-Language': 'en-US' },
      },
    );
    expect(response.ok()).toBeTruthy();
    expect(response.status()).toBe(200);

    const etag = response.headers()['etag'];
    expect(etag, 'expected ETag header on first response').toBeTruthy();
    expect(etag).toMatch(/^"en-US-\d+"$/);

    const cacheControl = response.headers()['cache-control'];
    expect(cacheControl).toBe('private, must-revalidate');

    const payload = await response.json();
    const messages = payload?.data?.messages ?? payload?.messages;
    expect(messages, 'expected messages payload to be present on 200').toBeTruthy();
  });

  test('TC-124b: 携带匹配 If-None-Match 时返回 304 且无 body', async ({
    adminPage,
  }) => {
    const firstResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: { 'Accept-Language': 'en-US' },
      },
    );
    expect(firstResponse.ok()).toBeTruthy();
    const etag = firstResponse.headers()['etag'];
    expect(etag).toBeTruthy();

    const secondResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: {
          'Accept-Language': 'en-US',
          'If-None-Match': etag!,
        },
      },
    );
    expect(secondResponse.status()).toBe(304);
    expect(secondResponse.headers()['etag']).toBe(etag);

    const body = await secondResponse.body();
    expect(body.length, '304 response must not carry a body').toBe(0);
  });

  test('TC-124c: 不同语言独立 ETag,跨语言 If-None-Match 不命中', async ({
    adminPage,
  }) => {
    const enResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: { 'Accept-Language': 'en-US' },
      },
    );
    const enEtag = enResponse.headers()['etag'];
    expect(enEtag).toMatch(/^"en-US-\d+"$/);

    const zhResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=zh-CN`,
      {
        headers: { 'Accept-Language': 'zh-CN' },
      },
    );
    const zhEtag = zhResponse.headers()['etag'];
    expect(zhEtag).toMatch(/^"zh-CN-\d+"$/);
    expect(zhEtag).not.toBe(enEtag);

    // Sending the en-US ETag while requesting zh-CN must not produce a 304.
    const crossResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=zh-CN`,
      {
        headers: {
          'Accept-Language': 'zh-CN',
          'If-None-Match': enEtag!,
        },
      },
    );
    expect(crossResponse.status()).toBe(200);
  });

  test('TC-124d: If-None-Match 通配符 * 也命中 304', async ({
    adminPage,
  }) => {
    const firstResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: { 'Accept-Language': 'en-US' },
      },
    );
    const etag = firstResponse.headers()['etag'];
    expect(etag).toBeTruthy();

    const wildcardResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: {
          'Accept-Language': 'en-US',
          'If-None-Match': '*',
        },
      },
    );
    expect(wildcardResponse.status()).toBe(304);
    expect(wildcardResponse.headers()['etag']).toBe(etag);
  });
});
