/**
 * 渲染工具函数
 * 用于表格列中渲染各种格式化内容
 */
import { h } from 'vue';

import { Tag } from 'ant-design-vue';

import type { DictEnumKey } from '@vben/constants';

import { DictTag } from '#/components/dict';
import { getDictOptions } from './dict';

/**
 * 根据 tagStyle 解析 Tag 颜色
 * tagStyle 格式示例: "success" 或 "background-color: #f50; color: #fff"
 */
function parseTagStyle(tagStyle?: string): string {
  if (!tagStyle) {
    return 'default';
  }
  // 如果是简单的颜色名称（如 success, error, warning 等）
  if (!tagStyle.includes(':')) {
    return tagStyle;
  }
  // 否则返回 default，由 style 属性处理
  return 'default';
}

/**
 * 渲染字典标签
 * 用于表格列中根据字典值渲染对应的标签
 * @param value 字典值
 * @param dictType 字典类型
 * @returns VNode
 */
export function renderDict(
  value: number | string,
  dictType: DictEnumKey | string,
) {
  const options = getDictOptions(dictType);
  return h(DictTag, { dicts: options, value });
}

/**
 * 渲染简单标签（不使用字典）
 * 用于已知标签样式的情况
 * @param value 显示值
 * @param color 标签颜色
 * @param tagStyle 自定义样式
 */
export function renderTag(
  value: number | string,
  color?: string,
  tagStyle?: string,
) {
  const resolvedColor = color || parseTagStyle(tagStyle);
  return h(
    Tag,
    {
      color: resolvedColor,
      style: tagStyle?.includes(':') ? tagStyle : undefined,
    },
    () => String(value ?? '-'),
  );
}
