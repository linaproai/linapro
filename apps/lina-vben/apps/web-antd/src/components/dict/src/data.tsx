import type { VNode } from 'vue';

import { Tag } from 'ant-design-vue';

import { $t } from '#/locales';

interface TagType {
  [key: string]: { color: string; labelKey: string };
}

export const tagTypes: TagType = {
  cyan: { color: 'cyan', labelKey: 'pages.system.dict.data.tagStyle.presets.cyan' },
  danger: { color: 'error', labelKey: 'pages.system.dict.data.tagStyle.presets.danger' },
  /** 由于和elementUI不同 用于替换颜色 */
  default: { color: 'default', labelKey: 'pages.system.dict.data.tagStyle.presets.default' },
  green: { color: 'green', labelKey: 'pages.system.dict.data.tagStyle.presets.green' },
  info: { color: 'default', labelKey: 'pages.system.dict.data.tagStyle.presets.info' },
  orange: { color: 'orange', labelKey: 'pages.system.dict.data.tagStyle.presets.orange' },
  /** 自定义预设 color可以为16进制颜色 */
  pink: { color: 'pink', labelKey: 'pages.system.dict.data.tagStyle.presets.pink' },
  primary: { color: 'processing', labelKey: 'pages.system.dict.data.tagStyle.presets.primary' },
  purple: { color: 'purple', labelKey: 'pages.system.dict.data.tagStyle.presets.purple' },
  red: { color: 'red', labelKey: 'pages.system.dict.data.tagStyle.presets.red' },
  success: { color: 'success', labelKey: 'pages.system.dict.data.tagStyle.presets.success' },
  warning: { color: 'warning', labelKey: 'pages.system.dict.data.tagStyle.presets.warning' },
};

// 字典选择使用 { label: string; value: string }[]
interface Options {
  label: string | VNode;
  value: string;
}

export function tagSelectOptions() {
  const selectArray: Options[] = [];
  Object.keys(tagTypes).forEach((key) => {
    if (!tagTypes[key]) return;
    const label = $t(tagTypes[key].labelKey);
    const color = tagTypes[key].color;
    selectArray.push({
      label: <Tag color={color}>{label}</Tag>,
      value: key,
    });
  });
  return selectArray;
}
