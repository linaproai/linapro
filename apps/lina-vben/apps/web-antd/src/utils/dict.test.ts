import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useDictStore } from './dict';

const dictApiMock = vi.hoisted(() => ({
  dictDataByType: vi.fn(),
}));

vi.mock('#/api/system/dict/dict-data', () => dictApiMock);

function dictData(label: string, value: string, tagStyle = 'primary') {
  return {
    canEdit: true,
    canOverride: false,
    cssClass: '',
    createdAt: null,
    dictType: 'sys_normal_disable',
    id: Number(value),
    isBuiltin: 1,
    isFallback: false,
    label,
    overrideMode: 'none' as const,
    remark: '',
    sort: Number(value),
    sourceTenantId: 0,
    status: 1,
    tagStyle,
    updatedAt: null,
    value,
  };
}

describe('dict store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    dictApiMock.dictDataByType.mockReset();
  });

  it('refreshes cached dictionary options without replacing mounted array references', async () => {
    dictApiMock.dictDataByType
      .mockResolvedValueOnce([dictData('正常', '1')])
      .mockResolvedValueOnce([dictData('已同步', '1', 'success')]);

    const store = useDictStore();
    const options = store.getDictOptions('sys_normal_disable');

    await store.getDictOptionsAsync('sys_normal_disable');
    expect(options).toEqual([
      expect.objectContaining({
        label: '正常',
        tagStyle: 'primary',
        value: '1',
      }),
    ]);

    await store.refreshDictOptions('sys_normal_disable');

    expect(store.getDictOptions('sys_normal_disable')).toBe(options);
    expect(options).toEqual([
      expect.objectContaining({
        label: '已同步',
        tagStyle: 'success',
        value: '1',
      }),
    ]);
    expect(dictApiMock.dictDataByType).toHaveBeenCalledTimes(2);
  });

  it('reuses one in-flight request for concurrent dictionary option readers', async () => {
    let resolveRequest: (value: ReturnType<typeof dictData>[]) => void;
    dictApiMock.dictDataByType.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveRequest = resolve;
      }),
    );

    const store = useDictStore();
    const options = store.getDictOptions('sys_show_hide');
    const asyncOptions = store.getDictOptionsAsync('sys_show_hide');

    resolveRequest!([dictData('显示', '1')]);
    await asyncOptions;

    expect(store.getDictOptions('sys_show_hide')).toBe(options);
    expect(options[0]?.label).toBe('显示');
    expect(dictApiMock.dictDataByType).toHaveBeenCalledTimes(1);
  });
});
