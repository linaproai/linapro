<script setup lang="ts">
import type { CheckboxChangeEvent } from 'ant-design-vue/es/checkbox/interface';
import type { DataNode } from 'ant-design-vue/es/tree';

import { computed, nextTick, onMounted, ref } from 'vue';

import { Checkbox, Tree } from 'ant-design-vue';

import { treeToList } from '#/utils/tree';

defineOptions({ inheritAttrs: false });

const props = withDefaults(defineProps<Props>(), {
  expandAllOnInit: false,
  fieldNames: () => ({ key: 'id', title: 'label' }),
  resetOnStrictlyChange: true,
  treeData: () => [],
});

interface Props {
  expandAllOnInit?: boolean;
  fieldNames?: { key: string; title: string };
  resetOnStrictlyChange?: boolean;
  treeData?: DataNode[];
}

const expandStatus = ref(false);
const selectAllStatus = ref(false);

const associationText = computed(() => {
  return checkStrictly.value ? '父子节点关联' : '父子节点独立';
});

const checkedKeys = defineModel<(number | string)[]>('value', {
  default: () => [],
});

const checkStrictly = defineModel<boolean>('checkStrictly', {
  default: () => true,
});

const computedCheckedKeys = computed<any>({
  get() {
    if (!checkStrictly.value) {
      return {
        checked: [...checkedKeys.value],
        halfChecked: [],
      };
    }
    return checkedKeys.value;
  },
  set(v) {
    if (!checkStrictly.value) {
      checkedKeys.value = [...v.checked, ...v.halfChecked];
      return;
    }
    checkedKeys.value = v;
  },
});

const allKeys = computed(() => {
  const idField = props.fieldNames.key;
  return treeToList(props.treeData as any[]).map((item: any) => item[idField]);
});

function handleCheckedAllChange(e: CheckboxChangeEvent) {
  checkedKeys.value = e.target.checked ? allKeys.value : [];
}

const expandedKeys = ref<string[]>([]);
function handleExpandOrCollapseAll() {
  expandStatus.value = !expandStatus.value;
  expandedKeys.value = expandStatus.value ? allKeys.value : [];
}

function handleCheckStrictlyChange() {
  if (props.resetOnStrictlyChange) {
    checkedKeys.value = [];
  }
}

onMounted(async () => {
  if (props.expandAllOnInit) {
    await nextTick();
    expandedKeys.value = allKeys.value;
  }
});
</script>

<template>
  <div class="bg-background w-full rounded-lg border-[1px] p-[12px]">
    <div class="flex items-center justify-between gap-2 border-b-[1px] pb-2">
      <div class="opacity-75">
        <span>节点状态: </span>
        <span :class="[checkStrictly ? 'text-primary' : 'text-red-500']">
          {{ associationText }}
        </span>
      </div>
    </div>
    <div
      class="flex flex-wrap items-center justify-between border-b-[1px] py-2"
    >
      <a-button size="small" @click="handleExpandOrCollapseAll">
        展开/折叠全部
      </a-button>
      <Checkbox
        v-model:checked="selectAllStatus"
        @change="handleCheckedAllChange"
      >
        全选/取消全选
      </Checkbox>
      <Checkbox
        v-model:checked="checkStrictly"
        @change="handleCheckStrictlyChange"
      >
        父子节点关联
      </Checkbox>
    </div>
    <div class="py-2">
      <Tree
        :check-strictly="!checkStrictly"
        v-model:checked-keys="computedCheckedKeys"
        v-model:expanded-keys="expandedKeys"
        :checkable="true"
        :field-names="fieldNames"
        :selectable="false"
        :tree-data="treeData"
      >
        <template
          v-for="slotName in Object.keys($slots)"
          :key="slotName"
          #[slotName]="data"
        >
          <slot :name="slotName" v-bind="data ?? {}"></slot>
        </template>
      </Tree>
    </div>
  </div>
</template>

<style lang="scss" scoped>
:deep(.ant-tree) {
  & .ant-tree-checkbox {
    margin: 0;
    margin-right: 6px;
  }

  & .ant-tree-switcher {
    display: flex;
    align-items: center;
    justify-content: center;
  }
}
</style>