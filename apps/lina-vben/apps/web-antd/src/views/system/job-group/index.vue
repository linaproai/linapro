<script setup lang="ts">
import type { JobGroupRecord } from '#/api/system/jobGroup/model';

import { useAccess } from '@vben/access';
import { Page, useVbenModal } from '@vben/common-ui';

import { computed, h, ref } from 'vue';

import { message, Modal, Popconfirm, Space, Tag, Tooltip } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  jobGroupDelete,
  jobGroupList,
} from '#/api/system/jobGroup';

import JobGroupModal from './modal.vue';

const accessCodes = {
  add: 'system:jobgroup:add',
  edit: 'system:jobgroup:edit',
  list: 'system:jobgroup:list',
  remove: 'system:jobgroup:remove',
} as const;

const { hasAccessByCodes } = useAccess();

const [GroupModal, groupModalApi] = useVbenModal({
  connectedComponent: JobGroupModal,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    schema: [
      {
        component: 'Input',
        fieldName: 'code',
        label: '分组编码',
      },
      {
        component: 'Input',
        fieldName: 'name',
        label: '分组名称',
      },
    ],
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
  },
  gridOptions: {
    checkboxConfig: {
      checkMethod: ({ row }: { row: JobGroupRecord }) => row.isDefault !== 1,
      highlight: true,
      reserve: true,
    },
    columns: [
      { type: 'checkbox', width: 56 },
      { field: 'code', title: '分组编码', minWidth: 160 },
      { field: 'name', title: '分组名称', minWidth: 180 },
      { field: 'sortOrder', title: '排序', width: 90 },
      { field: 'jobCount', title: '任务数', width: 90 },
      {
        field: 'isDefault',
        title: '默认分组',
        width: 120,
        slots: {
          default: ({ row }: { row: JobGroupRecord }) =>
            row.isDefault === 1
              ? h(
                  Tag,
                  {
                    color: 'gold',
                    'data-testid': `job-group-default-tag-${row.id}`,
                  },
                  () => '默认分组',
                )
              : '-',
        },
      },
      { field: 'remark', title: '备注', minWidth: 220 },
      { field: 'updatedAt', title: '更新时间', minWidth: 180 },
      {
        field: 'action',
        fixed: 'right',
        title: '操作',
        width: 220,
        slots: { default: 'action' },
      },
    ],
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues = {},
        ) => {
          return await jobGroupList({
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
    id: 'system-job-group-index',
  },
  gridEvents: {
    checkboxAll: syncCheckedRows,
    checkboxChange: syncCheckedRows,
  },
});

const checkedRows = ref<JobGroupRecord[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);

function syncCheckedRows() {
  checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as JobGroupRecord[];
}

function openCreateModal() {
  groupModalApi.setData({});
  groupModalApi.open();
}

function openEditModal(row: JobGroupRecord) {
  groupModalApi.setData({ record: row });
  groupModalApi.open();
}

function canDeleteRow(row: JobGroupRecord) {
  return row.isDefault !== 1 && hasAccessByCodes([accessCodes.remove]);
}

async function handleDelete(ids: Array<number>) {
  await jobGroupDelete(ids);
  message.success('删除成功');
  checkedRows.value = [];
  await gridApi.query();
}

function handleMultiDelete() {
  const rows = checkedRows.value;
  if (rows.length === 0) {
    return;
  }
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的 ${rows.length} 个任务分组吗？分组下的任务将自动迁移到默认分组。`,
    onOk: async () => {
      await handleDelete(rows.map((row) => row.id));
    },
  });
}

function handleReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true" data-testid="job-group-page">
    <Grid table-title="任务分组列表">
      <template #toolbar-tools>
        <Space>
          <a-button
            v-if="hasAccessByCodes([accessCodes.remove])"
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            删 除
          </a-button>
          <a-button
            v-if="hasAccessByCodes([accessCodes.add])"
            data-testid="job-group-add"
            type="primary"
            @click="openCreateModal"
          >
            新 增
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button
            v-if="hasAccessByCodes([accessCodes.edit])"
            :data-testid="`job-group-edit-${row.id}`"
            @click.stop="openEditModal(row)"
          >
            编辑
          </ghost-button>
          <template v-if="row.isDefault === 1">
            <Tooltip title="默认分组不可删除">
              <ghost-button
                danger
                disabled
                :data-testid="`job-group-delete-${row.id}`"
              >
                删除
              </ghost-button>
            </Tooltip>
          </template>
          <Popconfirm
            v-else-if="canDeleteRow(row)"
            placement="left"
            title="确认删除该分组吗？任务将自动迁移到默认分组。"
            @confirm="handleDelete([row.id])"
          >
            <ghost-button
              danger
              :data-testid="`job-group-delete-${row.id}`"
              @click.stop=""
            >
              删除
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>

    <GroupModal @reload="handleReload" />
  </Page>
</template>
