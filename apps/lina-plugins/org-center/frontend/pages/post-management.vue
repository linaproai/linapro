<script lang="ts">
export const pluginPageMeta = {
  capabilities: ['organization.management'],
  routePath: '/system/post',
  title: 'Positions',
};
</script>

<script setup lang="ts">
import type { Post } from './post-client';

import { computed, onMounted, ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';

import { message, Modal, Popconfirm, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { $t } from '#/locales';
import { postDelete, postDeptTree, postList } from './post-client';
import { useDictStore } from '#/store/dict';
import DeptTree from '#/views/system/user/dept-tree.vue';

import { buildColumns, buildQuerySchema } from './post-data';
import PostDrawer from './post-drawer.vue';

const selectDeptId = ref<string[]>([]);
const deptTreeRef = ref<InstanceType<typeof DeptTree>>();

// 加载字典数据
const dictStore = useDictStore();

onMounted(async () => {
  const statusOptions = await dictStore.getDictOptionsAsync('sys_normal_disable');
  gridApi.formApi.updateSchema([
    {
      fieldName: 'status',
      componentProps: {
        options: statusOptions.map((d) => ({
          label: d.label,
          value: Number(d.value),
        })),
      },
    },
  ]);
});

const [PostDrawerRef, postDrawerApi] = useVbenDrawer({
  connectedComponent: PostDrawer,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: buildQuerySchema(),
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
    handleReset: async () => {
      selectDeptId.value = [];
      const { formApi, reload } = gridApi;
      await formApi.resetForm();
      const formValues = formApi.form.values;
      formApi.setLatestSubmissionValues(formValues);
      await reload(formValues);
    },
  },
  gridOptions: {
    checkboxConfig: {
      highlight: true,
      reserve: true,
    },
    columns: buildColumns(),
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async ({ page }: { page: { currentPage: number; pageSize: number } }, formValues: Record<string, any> = {}) => {
          if (selectDeptId.value.length === 1) {
            formValues.deptId = selectDeptId.value[0];
          } else {
            Reflect.deleteProperty(formValues, 'deptId');
          }
          return await postList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    id: 'system-post-index',
  },
  gridEvents: {
    checkboxChange: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as Post[];
    },
    checkboxAll: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as Post[];
    },
  },
});

const checkedRows = ref<Post[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);

function handleAdd() {
  postDrawerApi.setData({});
  postDrawerApi.open();
}

function handleEdit(row: Post) {
  postDrawerApi.setData({ id: row.id });
  postDrawerApi.open();
}

async function handleDelete(row: Post) {
  await postDelete(String(row.id));
  message.success($t('pages.common.deleteSuccess'));
  await gridApi.query();
  deptTreeRef.value?.refreshTree();
}

function handleMultiDelete() {
  const rows = gridApi.grid.getCheckboxRecords() as Post[];
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('plugin.org-center.post.messages.deleteSelectedConfirm', {
      count: ids.length,
    }),
    onOk: async () => {
      await postDelete(ids.join(','));
      checkedRows.value = [];
      await gridApi.query();
      deptTreeRef.value?.refreshTree();
    },
  });
}

function onReload() {
  gridApi.query();
  deptTreeRef.value?.refreshTree();
}
</script>

<template>
  <Page
    :auto-content-height="true"
    content-class="flex flex-col gap-[8px] w-full 2xl:flex-row"
  >
    <DeptTree
      ref="deptTreeRef"
      :api="postDeptTree"
      v-model:select-dept-id="selectDeptId"
      class="w-full 2xl:w-[240px]"
      @reload="() => gridApi.reload()"
      @select="() => gridApi.reload()"
    />
    <Grid
      class="flex-1 overflow-hidden"
      :table-title="$t('plugin.org-center.post.tableTitle')"
    >
      <template #toolbar-tools>
        <Space>
          <a-button
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            {{ $t('pages.common.delete') }}
          </a-button>
          <a-button type="primary" @click="handleAdd">
            {{ $t('pages.common.add') }}
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleEdit(row)">
            {{ $t('pages.common.edit') }}
          </ghost-button>
          <Popconfirm
            placement="left"
            :title="$t('pages.common.deleteConfirm')"
            @confirm="handleDelete(row)"
          >
            <ghost-button danger @click.stop="">
              {{ $t('pages.common.delete') }}
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>

    <PostDrawerRef @reload="onReload" />
  </Page>
</template>
