import type { VNode } from 'vue';

import { Tag } from 'ant-design-vue';

import { $t } from '#/locales';

interface TagType {
  [key: string]: { color: string; label: string };
}

export const tagTypes: TagType = {
  cyan: { color: 'cyan', label: $t('pages.system.dict.data.tagStyle.presets.cyan') },
  danger: { color: 'error', label: $t('pages.system.dict.data.tagStyle.presets.danger') },
  /** 由于和elementUI不同 用于替换颜色 */
  default: { color: 'default', label: $t('pages.system.dict.data.tagStyle.presets.default') },
  green: { color: 'green', label: $t('pages.system.dict.data.tagStyle.presets.green') },
  info: { color: 'default', label: $t('pages.system.dict.data.tagStyle.presets.info') },
  orange: { color: 'orange', label: $t('pages.system.dict.data.tagStyle.presets.orange') },
  /** 自定义预设 color可以为16进制颜色 */
  pink: { color: 'pink', label: $t('pages.system.dict.data.tagStyle.presets.pink') },
  primary: { color: 'processing', label: $t('pages.system.dict.data.tagStyle.presets.primary') },
  purple: { color: 'purple', label: $t('pages.system.dict.data.tagStyle.presets.purple') },
  red: { color: 'red', label: $t('pages.system.dict.data.tagStyle.presets.red') },
  success: { color: 'success', label: $t('pages.system.dict.data.tagStyle.presets.success') },
  warning: { color: 'warning', label: $t('pages.system.dict.data.tagStyle.presets.warning') },
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
    const label = tagTypes[key].label;
    const color = tagTypes[key].color;
    selectArray.push({
      label: <Tag color={color}>{label}</Tag>,
      value: key,
    });
  });
  return selectArray;
}
