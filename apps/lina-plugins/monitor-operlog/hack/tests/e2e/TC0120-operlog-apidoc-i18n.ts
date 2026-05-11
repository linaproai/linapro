import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import {
  createAdminApiContext,
  expectSuccess,
} from '@host-tests/support/api/job';
import {
  xlsxRead,
  xlsxUtils,
} from '../support/xlsx';

type OperLog = {
  id: number;
  operSummary: string;
  requestMethod: string;
  title: string;
};

type OperLogList = {
  items: OperLog[];
  total: number;
};

async function findOperLogByLocalizedText(
  api: Awaited<ReturnType<typeof createAdminApiContext>>,
  locale: string,
  title: string,
  summary: string,
) {
  const result = await expectSuccess<OperLogList>(
    await api.get(
      `operlog?pageNum=1&pageSize=20&title=${encodeURIComponent(
        title,
      )}&orderBy=operTime&orderDirection=desc`,
      {
        headers: {
          'Accept-Language': locale,
        },
      },
    ),
  );
  return (
    result.items.find(
      (item) =>
        item.title === title &&
        item.operSummary === summary &&
        item.requestMethod === 'GET',
    ) ?? null
  );
}

test.describe('TC0120 操作日志路由文案复用 apidoc 国际化', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'monitor-operlog');
  });

  test('TC-120a~c: 列表、详情与导出按查看语言展示审计路由文案', async () => {
    const api = await createAdminApiContext();
    try {
      const triggerResponse = await api.get(
        'dict/type/export?pageNum=1&pageSize=1',
        {
          headers: {
            'Accept-Language': 'zh-CN',
          },
        },
      );
      expect(triggerResponse.ok()).toBeTruthy();

      let zhLog: OperLog | null = null;
      await expect
        .poll(async () => {
          zhLog = await findOperLogByLocalizedText(
            api,
            'zh-CN',
            '字典管理',
            '导出字典类型',
          );
          return Boolean(zhLog);
        }, { timeout: 15_000 })
        .toBe(true);
      expect(zhLog).toBeTruthy();

      const enLog = await findOperLogByLocalizedText(
        api,
        'en-US',
        'Dictionary Management',
        'Export dictionary type',
      );
      expect(enLog?.id).toBe(zhLog!.id);

      const detail = await expectSuccess<OperLog>(
        await api.get(`operlog/${zhLog!.id}`, {
          headers: {
            'Accept-Language': 'en-US',
          },
        }),
      );
      expect(detail.title).toBe('Dictionary Management');
      expect(detail.operSummary).toBe('Export dictionary type');

      const exportResponse = await api.get(`operlog/export?ids=${zhLog!.id}`, {
        headers: {
          'Accept-Language': 'en-US',
        },
      });
      expect(exportResponse.ok()).toBeTruthy();
      const workbook = xlsxRead(await exportResponse.body(), { type: 'buffer' });
      const sheet = workbook.Sheets[workbook.SheetNames[0]];
      const rows = xlsxUtils.sheet_to_json(sheet, { header: 1 }) as string[][];
      const tableText = rows.map((row) => row.join('|')).join('\n');

      expect(rows[0]).toContain('Module Name');
      expect(rows[0]).toContain('Operation Summary');
      expect(tableText).toContain('Dictionary Management');
      expect(tableText).toContain('Export dictionary type');
      expect(tableText).not.toContain('字典管理');
      expect(tableText).not.toContain('导出字典类型');
    } finally {
      await api.dispose();
    }
  });
});
