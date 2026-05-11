import type { APIRequestContext } from '@host-tests/support/playwright';

import { expect, test } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import {
  createAdminApiContext,
  expectSuccess,
} from '@host-tests/support/api/job';

interface MessageItem {
  id: number;
  type: number;
  typeLabel: string;
}

test.describe('TC0122 宿主消息中心 typeLabel 由后端按当前语言返回', () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'content-notice');
  });

  test('TC0122a: 中文 Accept-Language 下消息列表 typeLabel 为中文', async () => {
    const data = await expectSuccess<{ list: MessageItem[] }>(
      await adminApi.get('user/message?pageSize=20', {
        headers: { 'Accept-Language': 'zh-CN' },
      }),
    );
    expect(Array.isArray(data.list)).toBeTruthy();
    if (data.list.length === 0) {
      test.info().annotations.push({
        type: 'note',
        description: 'No messages available; schema-only check',
      });
      return;
    }
    for (const item of data.list) {
      expect(typeof item.typeLabel).toBe('string');
      expect(item.typeLabel.length).toBeGreaterThan(0);
      if (item.type === 1) expect(item.typeLabel).toBe('通知');
      if (item.type === 2) expect(item.typeLabel).toBe('公告');
    }
  });

  test('TC0122b: 英文 Accept-Language 下消息列表 typeLabel 为英文', async () => {
    const data = await expectSuccess<{ list: MessageItem[] }>(
      await adminApi.get('user/message?pageSize=20', {
        headers: { 'Accept-Language': 'en-US' },
      }),
    );
    expect(Array.isArray(data.list)).toBeTruthy();
    if (data.list.length === 0) {
      test.info().annotations.push({
        type: 'note',
        description: 'No messages available; schema-only check',
      });
      return;
    }
    for (const item of data.list) {
      expect(typeof item.typeLabel).toBe('string');
      if (item.type === 1) expect(item.typeLabel).toBe('Notice');
      if (item.type === 2) expect(item.typeLabel).toBe('Announcement');
    }
  });

  test('TC0122c: 通知详情接口 GetRes 同样返回本地化 typeLabel', async () => {
    const list = await expectSuccess<{ list: MessageItem[] }>(
      await adminApi.get('user/message?pageSize=1', {
        headers: { 'Accept-Language': 'en-US' },
      }),
    );
    if (list.list.length === 0) {
      test.info().annotations.push({
        type: 'note',
        description: 'No message available to fetch detail',
      });
      return;
    }
    const id = list.list[0]?.id;
    const detail = await expectSuccess<{ id: number; type: number; typeLabel: string }>(
      await adminApi.get(`user/message/${id}`, {
        headers: { 'Accept-Language': 'en-US' },
      }),
    );
    expect(typeof detail.typeLabel).toBe('string');
    if (detail.type === 1) expect(detail.typeLabel).toBe('Notice');
    if (detail.type === 2) expect(detail.typeLabel).toBe('Announcement');
  });

  test('TC0122d: 消息中心页面不再渲染 pages.status 原始 i18n key', async ({
    adminPage,
  }) => {
    await adminPage.goto('/system/message');
    await adminPage.waitForLoadState('networkidle');
    const html = await adminPage.content();
    expect(html).not.toContain('pages.status.notice');
    expect(html).not.toContain('pages.status.announcement');
  });
});
