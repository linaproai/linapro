<script setup lang="ts">
import type { JobLogRecord, JobRecord } from '#/api/system/job/model';

import { useAccess } from '@vben/access';
import { Page, useVbenModal } from '@vben/common-ui';

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
          placeholder: '请选择任务',
        },
        fieldName: 'jobId',
        label: '任务名称',
      },
      {
        component: 'Select',
        componentProps: {
          options: [
            { label: '运行中', value: 'running' },
            { label: '成功', value: 'success' },
            { label: '失败', value: 'failed' },
            { label: '取消', value: 'cancelled' },
            { label: '超时', value: 'timeout' },
          ],
          placeholder: '请选择状态',
        },
        fieldName: 'status',
        label: '执行状态',
      },
      {
        component: 'Input',
        fieldName: 'nodeId',
        label: '执行节点',
      },
      {
        component: 'RangePicker',
        componentProps: {
          valueFormat: 'YYYY-MM-DD HH:mm:ss',
        },
        fieldName: 'startTime',
        label: '开始时间',
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
        placeholder: '请选择任务',
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
  message.success('终止指令已发送');
  await gridApi.query();
}

function handleDelete() {
  const ids = checkedRows.value.map((row) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的 ${ids.length} 条执行日志吗？`,
    onOk: async () => {
      await jobLogDelete(ids);
      checkedRows.value = [];
      message.success('删除成功');
      await gridApi.query();
    },
  });
}

function handleClear() {
  const formValues = gridApi.formApi.form.values as Record<string, any>;
  const jobId =
    typeof formValues.jobId === 'number' ? formValues.jobId : undefined;
  const content = jobId
    ? '确认清空当前任务的执行日志吗？'
    : '确认清空全部执行日志吗？';
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content,
    onOk: async () => {
      await jobLogClear(jobId);
      checkedRows.value = [];
      message.success('清空成功');
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
    <Grid table-title="执行日志列表">
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
            删 除
          </a-button>
          <a-button
            v-if="hasAccessByCodes([accessCodes.remove])"
            data-testid="job-log-clear"
            @click="handleClear"
          >
            清 空
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button
            :data-testid="`job-log-detail-${row.id}`"
            @click.stop="handleOpenDetail(row)"
          >
            详情
          </ghost-button>
          <Popconfirm
            v-if="canCancelRow(row)"
            placement="left"
            title="确认终止当前任务实例吗？"
            @confirm="handleCancel(row)"
          >
            <ghost-button
              danger
              :data-testid="`job-log-cancel-${row.id}`"
              @click.stop=""
            >
              终止
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>

    <DetailModal @reload="handleReload" />
  </Page>
</template>
