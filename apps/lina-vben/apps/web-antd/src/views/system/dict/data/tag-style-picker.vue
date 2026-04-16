<script setup lang="ts">
import type { RadioChangeEvent } from 'ant-design-vue';

import { computed } from 'vue';

import { Input, RadioGroup, Select } from 'ant-design-vue';

import { tagSelectOptions } from '#/components/dict';

/**
 * 需要禁止透传
 * 不禁止会有奇怪的bug 会绑定到selectType上
 */
defineOptions({ inheritAttrs: false });

defineEmits<{ deselect: [] }>();

const options = [
  { label: '默认颜色', value: 'default' },
  { label: '自定义颜色', value: 'custom' },
] as const;

const computedOptions = computed(
  () => options as unknown as { label: string; value: string }[],
);

type SelectType = (typeof options)[number]['value'];

const selectType = defineModel<SelectType>('selectType', {
  default: 'default',
});

/**
 * color必须为hex颜色或者undefined
 */
const color = defineModel<string | undefined>('value', {
  default: undefined,
});

function handleSelectTypeChange(e: RadioChangeEvent) {
  // 必须给默认hex颜色 不能为空字符串
  color.value = e.target.value === 'custom' ? '#1677ff' : undefined;
}

function handleColorChange(e: Event) {
  color.value = (e.target as HTMLInputElement).value;
}
</script>

<template>
  <div class="flex flex-1 items-center gap-[6px]">
    <RadioGroup
      v-model:value="selectType"
      :options="computedOptions"
      button-style="solid"
      option-type="button"
      @change="handleSelectTypeChange"
    />
    <Select
      v-if="selectType === 'default'"
      v-model:value="color"
      :allow-clear="true"
      :options="tagSelectOptions()"
      class="flex-1"
      placeholder="请选择标签样式"
      @deselect="$emit('deselect')"
    />
    <Input
      v-if="selectType === 'custom'"
      :value="color"
      class="flex-1"
      placeholder="输入十六进制颜色值，如 #1677ff"
      @change="handleColorChange"
    />
  </div>
</template>
