<script setup lang="ts">
import type { PropType } from 'vue';

import type { DeptTree } from '#/api/system/user';

import { onMounted, ref } from 'vue';

import { IconifyIcon } from '@vben/icons';

import { Empty, InputSearch, Skeleton, Tree } from 'ant-design-vue';

import { getDeptTree } from '#/api/system/user';

defineOptions({ inheritAttrs: false });

const props = withDefaults(defineProps<Props>(), {
  showSearch: true,
  api: getDeptTree,
});

const emit = defineEmits<{
  /** 点击刷新按钮的事件 */
  reload: [];
  /** 点击节点的事件 */
  select: [];
}>();

interface Props {
  /** 调用的接口 */
  api?: () => Promise<DeptTree[]>;
  /** 是否显示搜索框 */
  showSearch?: boolean;
}

const selectDeptId = defineModel('selectDeptId', {
  required: true,
  type: Array as PropType<string[]>,
});

const searchValue = defineModel('searchValue', {
  type: String,
  default: '',
});

/** 部门数据源 */
type DeptTreeArray = DeptTree[];
const deptTreeArray = ref<DeptTreeArray>([]);
const expandedDeptIds = ref<Array<number | string>>([]);
/** 骨架屏加载 */
const showTreeSkeleton = ref<boolean>(true);

function collectDeptIds(nodes: DeptTreeArray): Array<number | string> {
  const ids: Array<number | string> = [];
  const walk = (items: DeptTreeArray) => {
    for (const item of items) {
      ids.push(item.id);
      if (item.children?.length) {
        walk(item.children);
      }
    }
  };
  walk(nodes);
  return ids;
}

async function loadTree() {
  showTreeSkeleton.value = true;
  searchValue.value = '';
  selectDeptId.value = [];

  const ret = await props.api();
  deptTreeArray.value = ret;
  expandedDeptIds.value = collectDeptIds(ret);
  showTreeSkeleton.value = false;
}

async function handleReload() {
  await loadTree();
  emit('reload');
}

/** 静默刷新树数据（不重置选中状态和搜索框） */
async function refreshTree() {
  const ret = await props.api();
  deptTreeArray.value = ret;
  expandedDeptIds.value = collectDeptIds(ret);
}

onMounted(loadTree);

defineExpose({ refreshTree });
</script>

<template>
  <div :class="$attrs.class as string">
    <Skeleton
      :loading="showTreeSkeleton"
      :paragraph="{ rows: 8 }"
      active
      class="p-[8px]"
    >
      <div
        class="flex h-full flex-col overflow-y-auto rounded-lg bg-background"
      >
        <!-- 固定在顶部 必须加上bg-background背景色 否则会产生'穿透'效果 -->
        <div
          v-if="showSearch"
          class="sticky top-0 left-0 z-100 bg-background p-[8px]"
        >
          <InputSearch
            v-model:value="searchValue"
            :placeholder="$t('pages.common.search')"
            size="small"
            allow-clear
          >
            <template #enterButton>
              <a-button @click="handleReload">
                <IconifyIcon
                  class="text-primary"
                  icon="ant-design:sync-outlined"
                />
              </a-button>
            </template>
          </InputSearch>
        </div>
        <div class="h-full overflow-x-hidden px-[8px]">
          <Tree
            v-bind="$attrs"
            v-if="deptTreeArray.length > 0"
            v-model:expanded-keys="expandedDeptIds"
            v-model:selected-keys="selectDeptId"
            :class="$attrs.class as string"
            :field-names="{ title: 'label', key: 'id' }"
            :show-line="{ showLeafIcon: false }"
            :tree-data="deptTreeArray as any"
            :virtual="false"
            default-expand-all
            @select="$emit('select')"
          >
            <template #title="{ label }">
              <span v-if="label.includes(searchValue)">
                {{ label.substring(0, label.indexOf(searchValue)) }}
                <span class="text-primary">{{ searchValue }}</span>
                {{
                  label.substring(
                    label.indexOf(searchValue) + searchValue.length,
                  )
                }}
              </span>
              <span v-else>{{ label }}</span>
            </template>
          </Tree>
          <!-- 无部门数据时显示空状态 -->
          <div v-else class="mt-5">
            <Empty
              :image="Empty.PRESENTED_IMAGE_SIMPLE"
              :description="$t('pages.system.user.messages.noDepartments')"
            />
          </div>
        </div>
      </div>
    </Skeleton>
  </div>
</template>
