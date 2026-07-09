/**
 * 字典工具函数
 */
import { reactive } from 'vue';

import { defineStore } from 'pinia';

import { dictDataByType } from '#/api/system/dict/dict-data';
import type { DictData } from '#/api/system/dict/dict-data-model';

import type { DictEnumKey } from '@vben/constants';

import { DictEnum } from '@vben/constants';

interface DictOption {
  label: string;
  value: string;
  tagStyle?: string;
  cssClass?: string;
}

export const useDictStore = defineStore('dict', () => {
  /**
   * select radio checkbox等使用 只能为固定格式{label, value}
   */
  const dictOptionsMap = reactive(new Map<string, DictOption[]>());
  /**
   * 添加一个字典请求状态的缓存
   * 主要解决多次请求重复api的问题
   */
  const dictRequestCache = new Map<string, Promise<DictOption[]>>();

  function normalizeDictOptions(list: DictData[] = []): DictOption[] {
    return list.map((item) => ({
      label: item.label,
      value: item.value,
      tagStyle: item.tagStyle,
      cssClass: item.cssClass,
    }));
  }

  function ensureDictOptions(dictName: string): DictOption[] {
    if (!dictOptionsMap.has(dictName)) {
      dictOptionsMap.set(dictName, []);
    }
    return dictOptionsMap.get(dictName)!;
  }

  function loadDictOptions(dictName: string, options: DictOption[]) {
    const promise = dictDataByType(dictName)
      .then((list) => {
        if (dictRequestCache.get(dictName) !== promise) {
          return options;
        }
        const items = normalizeDictOptions(list || []);
        options.splice(0, options.length, ...items);
        return options;
      })
      .catch(() => options)
      .finally(() => {
        if (dictRequestCache.get(dictName) === promise) {
          dictRequestCache.delete(dictName);
        }
      });
    dictRequestCache.set(dictName, promise);
    return promise;
  }

  /**
   * 获取字典选项（同步返回响应式数组，异步加载数据）
   * @param dictName 字典类型
   * @returns 响应式字典选项数组
   */
  function getDictOptions(dictName: string): DictOption[] {
    if (!dictName) return [];
    const options = ensureDictOptions(dictName);

    // 如果数组为空且没有正在进行的请求，触发异步加载
    if (options.length === 0 && !dictRequestCache.has(dictName)) {
      loadDictOptions(dictName, options);
    }

    return options;
  }

  /**
   * 异步获取字典选项（等待数据加载完成）
   * @param dictName 字典类型
   * @returns Promise<字典选项数组>
   */
  async function getDictOptionsAsync(dictName: string): Promise<DictOption[]> {
    // 先触发同步加载
    const options = getDictOptions(dictName);
    // 如果已有数据或没有正在进行的请求，直接返回
    if (options.length > 0 || !dictRequestCache.has(dictName)) {
      return options;
    }
    // 等待正在进行的请求完成
    await dictRequestCache.get(dictName)?.catch(() => options);
    return options;
  }

  async function refreshDictOptions(dictName?: string): Promise<void> {
    if (!dictName) {
      if (dictOptionsMap.size === 0) {
        dictRequestCache.clear();
        return;
      }
      await Promise.all(
        [...dictOptionsMap.keys()].map((key) => refreshDictOptions(key)),
      );
      return;
    }
    const options = ensureDictOptions(dictName);
    dictRequestCache.delete(dictName);
    await loadDictOptions(dictName, options);
  }

  function resetCache() {
    dictOptionsMap.clear();
    dictRequestCache.clear();
  }

  function $reset() {
    resetCache();
  }

  return {
    dictOptionsMap,
    getDictOptions,
    getDictOptionsAsync,
    refreshDictOptions,
    resetCache,
    $reset,
  };
});

/**
 * 获取字典选项列表（同步返回响应式数组）
 * 与参考项目保持一致的API设计
 * @param dictType 字典类型
 * @returns 字典选项数组
 */
export function getDictOptions(dictType: DictEnumKey | string): DictOption[] {
  const dictStore = useDictStore();
  return dictStore.getDictOptions(dictType);
}

/**
 * 同步获取字典选项（从缓存中获取，如缓存不存在返回空数组）
 * @param dictType 字典类型
 * @returns 字典选项数组
 */
export function getDictOptionsSync(dictType: DictEnumKey | string): DictOption[] {
  const dictStore = useDictStore();
  return dictStore.dictOptionsMap.get(dictType) || [];
}

export { DictEnum };
