<script setup lang="ts">
import type { JobGroupRecord } from '#/api/system/jobGroup/model';

import { useAccess } from '@vben/access';
import { Page, useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { computed, h, ref } from 'vue';

import {
  message,
  Modal,
  Popconfirm,
  Space,
  Tag,
  Tooltip,
} from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { jobGroupDelete, jobGroupList } from '#/api/system/jobGroup';
import {
  localizeSeedJobGroupName,
  localizeSeedJobGroupRemark,
} from '#/utils/display-l10n';

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
        label: $t('pages.system.jobGroup.fields.code'),
      },
      {
        component: 'Input',
        fieldName: 'name',
        label: $t('pages.system.jobGroup.fields.name'),
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
      {
        field: 'code',
        title: $t('pages.system.jobGroup.fields.code'),
        minWidth: 160,
      },
      {
        field: 'name',
        title: $t('pages.system.jobGroup.fields.name'),
        minWidth: 180,
        formatter: ({ row }: { row: JobGroupRecord }) =>
          localizeSeedJobGroupName(row.code, row.name),
      },
      { field: 'sortOrder', title: $t('pages.fields.sort'), width: 90 },
      {
        field: 'jobCount',
        title: $t('pages.system.jobGroup.fields.jobCount'),
        width: 90,
      },
      {
        field: 'isDefault',
        title: $t('pages.system.jobGroup.fields.defaultGroup'),
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
                  () => $t('pages.system.jobGroup.fields.defaultGroup'),
                )
              : '-',
        },
      },
      {
        field: 'remark',
        title: $t('pages.common.remark'),
        minWidth: 220,
        formatter: ({ row }: { row: JobGroupRecord }) =>
          localizeSeedJobGroupRemark(row.code, row.remark),
      },
      {
        field: 'updatedAt',
        title: $t('pages.common.updatedAt'),
        minWidth: 180,
      },
      {
        field: 'action',
        fixed: 'right',
        title: $t('pages.common.actions'),
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
  checkedRows.value = (gridApi.grid?.getCheckboxRecords() ||
    []) as JobGroupRecord[];
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
  message.success($t('pages.common.deleteSuccess'));
  checkedRows.value = [];
  await gridApi.query();
}

function handleMultiDelete() {
  const rows = checkedRows.value;
  if (rows.length === 0) {
    return;
  }
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('pages.system.jobGroup.messages.deleteSelectedConfirm', {
      count: rows.length,
    }),
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
    <Grid :table-title="$t('pages.system.jobGroup.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button
            v-if="hasAccessByCodes([accessCodes.remove])"
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            {{ $t('pages.common.delete') }}
          </a-button>
          <a-button
            v-if="hasAccessByCodes([accessCodes.add])"
            data-testid="job-group-add"
            type="primary"
            @click="openCreateModal"
          >
            {{ $t('pages.common.add') }}
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
            {{ $t('pages.common.edit') }}
          </ghost-button>
          <template v-if="row.isDefault === 1">
            <Tooltip
              :title="
                $t('pages.system.jobGroup.messages.defaultDeleteDisabled')
              "
            >
              <ghost-button
                danger
                disabled
                :data-testid="`job-group-delete-${row.id}`"
              >
                {{ $t('pages.common.delete') }}
              </ghost-button>
            </Tooltip>
          </template>
          <Popconfirm
            v-else-if="canDeleteRow(row)"
            placement="left"
            :title="$t('pages.system.jobGroup.messages.deleteConfirm')"
            @confirm="handleDelete([row.id])"
          >
            <ghost-button
              danger
              :data-testid="`job-group-delete-${row.id}`"
              @click.stop=""
            >
              {{ $t('pages.common.delete') }}
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>

    <GroupModal @reload="handleReload" />
  </Page>
</template>
