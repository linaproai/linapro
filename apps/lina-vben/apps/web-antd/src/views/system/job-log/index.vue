<script setup lang="ts">
import type { JobLogRecord, JobRecord } from '#/api/system/job/model';

import { useAccess } from '@vben/access';
import { Page, useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { onMounted, ref } from 'vue';

import {
  Alert,
  Checkbox,
  DatePicker,
  message,
  Modal,
  Popconfirm,
  Space,
} from 'ant-design-vue';

import { buildJobLogColumns, useVbenVxeGrid } from '#/adapter/vxe-table';
import { jobList } from '#/api/system/job';
import { jobLogCancel, jobLogClear } from '#/api/system/jobLog';

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

const RangePicker = DatePicker.RangePicker;
const jobOptions = ref<Array<{ label: string; value: number }>>([]);
const deleteRange = ref<[string, string]>();
const deleteAllLogs = ref(false);
const deleteRangeModalOpen = ref(false);
const deleteRangeSubmitting = ref(false);

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
  deleteRange.value = undefined;
  deleteAllLogs.value = false;
  deleteRangeModalOpen.value = true;
}

async function handleDeleteRangeConfirm() {
  const [beginTime, endTime] = deleteRange.value ?? [];
  if (!deleteAllLogs.value && (!beginTime || !endTime)) {
    message.warning($t('pages.system.jobLog.messages.deleteRangeRequired'));
    return;
  }

  deleteRangeSubmitting.value = true;
  try {
    const result = await jobLogClear(
      deleteAllLogs.value ? undefined : { beginTime, endTime },
    );
    message.success(
      $t('pages.system.jobLog.messages.deleteRangeSuccess', {
        count: result?.deleted ?? 0,
      }),
    );
    deleteRangeModalOpen.value = false;
    await gridApi.query();
  } finally {
    deleteRangeSubmitting.value = false;
  }
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
            danger
            type="primary"
            data-testid="job-log-delete"
            @click="handleDelete"
          >
            {{ $t('pages.common.delete') }}
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

    <Modal
      v-model:open="deleteRangeModalOpen"
      :destroy-on-close="true"
      :title="$t('pages.system.jobLog.messages.deleteRangeTitle')"
    >
      <div>
        <div data-testid="job-log-delete-alert">
          <Alert
            :message="$t('pages.system.jobLog.messages.deleteRangeDescription')"
            show-icon
            type="warning"
          />
        </div>
        <div data-testid="job-log-delete-all-option" style="margin-top: 16px">
          <Checkbox v-model:checked="deleteAllLogs">
            {{ $t('pages.system.jobLog.messages.deleteAllLabel') }}
          </Checkbox>
          <div class="text-xs text-gray-500" style="margin-top: 4px">
            {{ $t('pages.system.jobLog.messages.deleteAllHint') }}
          </div>
        </div>
        <div data-testid="job-log-delete-range-section" style="margin-top: 16px">
          <RangePicker
            v-model:value="deleteRange"
            :disabled="deleteAllLogs"
            class="w-full"
            value-format="YYYY-MM-DD"
          />
        </div>
      </div>
      <template #footer>
        <a-button @click="deleteRangeModalOpen = false">
          {{ $t('pages.common.cancel') }}
        </a-button>
        <a-button
          :loading="deleteRangeSubmitting"
          danger
          type="primary"
          @click="handleDeleteRangeConfirm"
        >
          {{ $t('pages.common.confirm') }}
        </a-button>
      </template>
    </Modal>
  </Page>
</template>
