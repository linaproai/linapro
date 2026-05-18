<script setup lang="ts">
import type { JobLogRecord, JobRecord } from '#/api/system/job/model';

import { useAccess } from '@vben/access';
import { Page, useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { computed, onMounted, ref } from 'vue';

import { message, Modal, Popconfirm, Space } from 'ant-design-vue';

import { buildJobLogColumns, useVbenVxeGrid } from '#/adapter/vxe-table';
import { jobList } from '#/api/system/job';
import { jobLogCancel, jobLogClear, jobLogDelete } from '#/api/system/jobLog';

import JobLogDetail from './detail.vue';

const accessCodes = {
  cancel: 'system:joblog:cancel',
  list: 'system:joblog:list',
  remove: 'system:joblog:remove',
  shell: 'system:job:shell',
} as const;

const { hasAccessByCodes } = useAccess();

const [DetailModal, detailModalApi] = useVbenModal({
  connectedComponent: JobLogDetail,
});

const jobOptions = ref<Array<{ label: string; value: number }>>([]);
const checkedRows = ref<JobLogRecord[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);

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
        component: 'Select',
        componentProps: {
          options: [],
          placeholder: $t('pages.system.jobLog.placeholders.selectJob'),
        },
        fieldName: 'jobId',
        label: $t('pages.system.jobLog.fields.jobName'),
      },
      {
        component: 'Select',
        componentProps: {
          options: [
            { label: $t('pages.system.jobLog.status.running'), value: 'running' },
            { label: $t('pages.system.jobLog.status.success'), value: 'success' },
            { label: $t('pages.system.jobLog.status.failed'), value: 'failed' },
            { label: $t('pages.system.jobLog.status.cancelled'), value: 'cancelled' },
            { label: $t('pages.system.jobLog.status.timeout'), value: 'timeout' },
          ],
          placeholder: $t('pages.system.jobLog.placeholders.selectStatus'),
        },
        fieldName: 'status',
        label: $t('pages.system.jobLog.fields.status'),
      },
      {
        component: 'Input',
        fieldName: 'nodeId',
        label: $t('pages.system.jobLog.fields.nodeId'),
      },
      {
        component: 'RangePicker',
        componentProps: {
          valueFormat: 'YYYY-MM-DD HH:mm:ss',
        },
        fieldName: 'startTime',
        label: $t('pages.system.jobLog.fields.startAt'),
      },
    ],
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
  },
  gridOptions: {
    checkboxConfig: {
      highlight: true,
      reserve: true,
    },
    columns: buildJobLogColumns(),
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) => {
          const params: Record<string, any> = {
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          };
          if (Array.isArray(params.startTime)) {
            params.beginTime = params.startTime[0];
            params.endTime = params.startTime[1];
            delete params.startTime;
          }
          return await import('#/api/system/jobLog').then(({ jobLogList }) =>
            jobLogList(params),
          );
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    id: 'system-job-log-index',
  },
  gridEvents: {
    checkboxAll: syncCheckedRows,
    checkboxChange: syncCheckedRows,
  },
});

onMounted(async () => {
  const result = await jobList({
    pageNum: 1,
    pageSize: 100,
  });
  jobOptions.value = (result.items || []).map((item: JobRecord) => ({
    label: item.name,
    value: item.id,
  }));
  gridApi.formApi.updateSchema([
    {
      componentProps: {
        options: jobOptions.value,
        placeholder: $t('pages.system.jobLog.placeholders.selectJob'),
      },
      fieldName: 'jobId',
    },
  ]);
});

function isShellLog(row: JobLogRecord) {
  if (!row.jobSnapshot) {
    return false;
  }
  try {
    const snapshot = JSON.parse(row.jobSnapshot);
    return snapshot?.taskType === 'shell';
  } catch {
    return false;
  }
}

function syncCheckedRows() {
  checkedRows.value = (gridApi.grid?.getCheckboxRecords() ||
    []) as JobLogRecord[];
}

function canCancelRow(row: JobLogRecord) {
  if (row.status !== 'running') {
    return false;
  }
  if (!hasAccessByCodes([accessCodes.cancel])) {
    return false;
  }
  if (isShellLog(row) && !hasAccessByCodes([accessCodes.shell])) {
    return false;
  }
  return true;
}

function handleOpenDetail(row: JobLogRecord) {
  detailModalApi.setData({ id: row.id });
  detailModalApi.open();
}

async function handleCancel(row: JobLogRecord) {
  await jobLogCancel(row.id);
  message.success($t('pages.system.jobLog.messages.cancelSent'));
  await gridApi.query();
}

function handleDelete() {
  const ids = checkedRows.value.map((row) => row.id);
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('pages.system.jobLog.messages.deleteSelectedConfirm', {
      count: ids.length,
    }),
    onOk: async () => {
      await jobLogDelete(ids);
      checkedRows.value = [];
      message.success($t('pages.common.deleteSuccess'));
      await gridApi.query();
    },
  });
}

function handleClear() {
  const formValues = gridApi.formApi.form.values as Record<string, any>;
  const jobId =
    typeof formValues.jobId === 'number' ? formValues.jobId : undefined;
  const content = jobId
    ? $t('pages.system.jobLog.messages.clearCurrentConfirm')
    : $t('pages.system.jobLog.messages.clearAllConfirm');
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content,
    onOk: async () => {
      await jobLogClear(jobId);
      checkedRows.value = [];
      message.success($t('pages.system.jobLog.messages.clearSuccess'));
      await gridApi.query();
    },
  });
}

function handleReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true" data-testid="job-log-page">
    <Grid :table-title="$t('pages.system.jobLog.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button
            v-if="hasAccessByCodes([accessCodes.remove])"
            :disabled="!hasChecked"
            danger
            type="primary"
            data-testid="job-log-delete"
            @click="handleDelete"
          >
            {{ $t('pages.common.delete') }}
          </a-button>
          <a-button
            v-if="hasAccessByCodes([accessCodes.remove])"
            data-testid="job-log-clear"
            @click="handleClear"
          >
            {{ $t('pages.common.clear') }}
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button
            :data-testid="`job-log-detail-${row.id}`"
            @click.stop="handleOpenDetail(row)"
          >
            {{ $t('pages.common.detail') }}
          </ghost-button>
          <Popconfirm
            v-if="canCancelRow(row)"
            placement="left"
            :title="$t('pages.system.jobLog.messages.terminateConfirm')"
            @confirm="handleCancel(row)"
          >
            <ghost-button
              danger
              :data-testid="`job-log-cancel-${row.id}`"
              @click.stop=""
            >
              {{ $t('pages.system.jobLog.actions.terminate') }}
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>

    <DetailModal @reload="handleReload" />
  </Page>
</template>
