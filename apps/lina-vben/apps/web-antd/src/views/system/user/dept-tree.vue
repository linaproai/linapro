<script setup lang="ts">
import type { PropType } from 'vue';

import type { DeptTree } from '#/api/system/dept/model';

import { onMounted, ref } from 'vue';

import { IconifyIcon } from '@vben/icons';
import { Empty, InputSearch, Skeleton, Tree } from 'ant-design-vue';

import { deptTree as getDeptTree } from '#/api/system/dept';

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
/** 骨架屏加载 */
const showTreeSkeleton = ref<boolean>(true);

async function loadTree() {
  showTreeSkeleton.value = true;
  searchValue.value = '';
  selectDeptId.value = [];

  const ret = await props.api();

  deptTreeArray.value = ret;
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
        class="bg-background flex h-full flex-col overflow-y-auto rounded-lg"
      >
        <!-- 固定在顶部 必须加上bg-background背景色 否则会产生'穿透'效果 -->
        <div
          v-if="showSearch"
          class="bg-background z-100 sticky left-0 top-0 p-[8px]"
        >
          <InputSearch
            v-model:value="searchValue"
            :placeholder="$t('pages.common.search')"
            size="small"
            allow-clear
          >
            <template #enterButton>
              <a-button @click="handleReload">
                <IconifyIcon class="text-primary" icon="ant-design:sync-outlined" />
              </a-button>
            </template>
          </InputSearch>
        </div>
        <div class="h-full overflow-x-hidden px-[8px]">
          <Tree
            v-bind="$attrs"
            v-if="deptTreeArray.length > 0"
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
              description="无部门数据"
            />
          </div>
        </div>
      </div>
    </Skeleton>
  </div>
</template>
