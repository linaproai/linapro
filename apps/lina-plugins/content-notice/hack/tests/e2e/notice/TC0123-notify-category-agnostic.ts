import type { APIRequestContext } from '@host-tests/support/playwright';

import { expect, test } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import {
  createAdminApiContext,
  expectSuccess,
} from '@host-tests/support/api/job';

interface MessageItem {
  id: number;
  type?: number;
  categoryCode: string;
  typeLabel: string;
  typeColor: string;
}

test.describe('TC0123 宿主消息中心 category-agnostic 设计', () => {
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

  test('TC0123a: 消息列表透出 categoryCode/typeLabel/typeColor，且不再返回 type 数值字段', async () => {
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
      expect(typeof item.categoryCode).toBe('string');
      expect(item.categoryCode.length).toBeGreaterThan(0);
      expect(typeof item.typeLabel).toBe('string');
      expect(typeof item.typeColor).toBe('string');
      // type 数值字段已被弃用：response 不再返回它
      expect(item.type).toBeUndefined();
    }
  });

  test('TC0123b: 插件类目（announcement / notice）的 typeLabel/typeColor 来自插件 manifest/i18n', async () => {
    const data = await expectSuccess<{ list: MessageItem[] }>(
      await adminApi.get('user/message?pageSize=50', {
        headers: { 'Accept-Language': 'zh-CN' },
      }),
    );
    if (data.list.length === 0) {
      test.info().annotations.push({
        type: 'note',
        description: 'No messages available to assert plugin category labels',
      });
      return;
    }
    for (const item of data.list) {
      switch (item.categoryCode) {
        case 'announcement': {
          expect(item.typeLabel).toBe('公告');
          expect(item.typeColor).toBe('green');
          break;
        }
        case 'notice': {
          expect(item.typeLabel).toBe('通知');
          expect(item.typeColor).toBe('blue');
          break;
        }
        // 其他类目（system / other / 未来插件）走 host 兜底，不在此用例断言
      }
    }
  });

  test('TC0123c: 英文环境下插件类目同样按插件 manifest/i18n 返回英文 typeLabel/typeColor', async () => {
    const data = await expectSuccess<{ list: MessageItem[] }>(
      await adminApi.get('user/message?pageSize=50', {
        headers: { 'Accept-Language': 'en-US' },
      }),
    );
    if (data.list.length === 0) return;
    for (const item of data.list) {
      switch (item.categoryCode) {
        case 'announcement': {
          expect(item.typeLabel).toBe('Announcement');
          expect(item.typeColor).toBe('green');
          break;
        }
        case 'notice': {
          expect(item.typeLabel).toBe('Notice');
          expect(item.typeColor).toBe('blue');
          break;
        }
      }
    }
  });

  test('TC0123d: 详情接口 GetRes 同样按 categoryCode 解析 typeLabel/typeColor，且不再返回 type', async () => {
    const list = await expectSuccess<{ list: MessageItem[] }>(
      await adminApi.get('user/message?pageSize=1', {
        headers: { 'Accept-Language': 'zh-CN' },
      }),
    );
    if (list.list.length === 0) return;
    const id = list.list[0]?.id;
    const detail = await expectSuccess<MessageItem>(
      await adminApi.get(`user/message/${id}`, {
        headers: { 'Accept-Language': 'zh-CN' },
      }),
    );
    expect(typeof detail.categoryCode).toBe('string');
    expect(detail.categoryCode.length).toBeGreaterThan(0);
    expect(typeof detail.typeLabel).toBe('string');
    expect(typeof detail.typeColor).toBe('string');
    expect(detail.type).toBeUndefined();
  });
});
