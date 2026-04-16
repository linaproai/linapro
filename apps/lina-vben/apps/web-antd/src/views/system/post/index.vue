<script setup lang="ts">
import type { Post } from '#/api/system/post/model';

import { computed, onMounted, ref } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';

import { message, Modal, Popconfirm, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { postDelete, postDeptTree, postList } from '#/api/system/post';
import { useDictStore } from '#/store/dict';
import DeptTree from '#/views/system/user/dept-tree.vue';

import { columns, querySchema } from './data';
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
    schema: querySchema,
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
    columns,
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
  message.success('删除成功');
  await gridApi.query();
  deptTreeRef.value?.refreshTree();
}

function handleMultiDelete() {
  const rows = gridApi.grid.getCheckboxRecords() as Post[];
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的${ids.length}条记录吗？`,
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
  <Page :auto-content-height="true" content-class="flex gap-[8px] w-full">
    <DeptTree
      ref="deptTreeRef"
      :api="postDeptTree"
      v-model:select-dept-id="selectDeptId"
      class="w-[260px]"
      @reload="() => gridApi.reload()"
      @select="() => gridApi.reload()"
    />
    <Grid class="flex-1 overflow-hidden" table-title="岗位列表">
      <template #toolbar-tools>
        <Space>
          <a-button
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            删 除
          </a-button>
          <a-button type="primary" @click="handleAdd">新 增</a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleEdit(row)">编辑</ghost-button>
          <Popconfirm
            placement="left"
            title="确认删除？"
            @confirm="handleDelete(row)"
          >
            <ghost-button danger @click.stop="">删除</ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>

    <PostDrawerRef @reload="onReload" />
  </Page>
</template>
