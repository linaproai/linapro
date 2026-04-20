<script setup lang="ts">
import type { JobRecord } from '#/api/system/job/model';
import type { JobGroupRecord } from '#/api/system/jobGroup/model';

import { useAccess } from '@vben/access';
import { Page, useVbenModal } from '@vben/common-ui';

import { computed, onMounted, ref } from 'vue';

import {
  Dropdown,
  Menu,
  MenuItem,
  message,
  Modal,
  Popconfirm,
  Space,
  Tooltip,
} from 'ant-design-vue';

import { buildJobColumns, useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  jobDelete,
  jobList,
  jobReset,
  jobTrigger,
} from '#/api/system/job';
import {
  JOB_PLUGIN_PAUSED_TOOLTIP,
  JOB_STATUS_FILTER_OPTIONS,
  getJobSourceKind,
} from '#/api/system/job/meta';
import { jobGroupList } from '#/api/system/jobGroup';
import { publicFrontendSettings } from '#/runtime/public-frontend';

import JobForm from './form.vue';

const accessCodes = {
  add: 'system:job:add',
  edit: 'system:job:edit',
  remove: 'system:job:remove',
  reset: 'system:job:reset',
  shell: 'system:job:shell',
  trigger: 'system:job:trigger',
} as const;

const { hasAccessByCodes } = useAccess();

const [JobFormModal, jobFormApi] = useVbenModal({
  connectedComponent: JobForm,
});

const groupOptions = ref<Array<{ label: string; value: number }>>([]);

const shellCapability = computed(() => publicFrontendSettings.cron.shell);
const hasShellPermission = computed(() =>
  hasAccessByCodes([accessCodes.shell]),
);
const canCreateShellJob = computed(() => {
  return hasAccessByCodes([accessCodes.add]) && shellBlockedReason() === '';
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
        component: 'Select',
        componentProps: {
          options: [],
          placeholder: '请选择分组',
        },
        fieldName: 'groupId',
        label: '任务分组',
      },
      {
        component: 'Select',
        componentProps: {
          options: JOB_STATUS_FILTER_OPTIONS,
          placeholder: '请选择状态',
        },
        fieldName: 'status',
        label: '任务状态',
      },
      {
        component: 'Input',
        fieldName: 'keyword',
        label: '关键字',
      },
    ],
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
  },
  gridOptions: {
    checkboxConfig: {
      checkMethod: ({ row }: { row: JobRecord }) => row.isBuiltin !== 1,
      highlight: true,
      reserve: true,
    },
    columns: buildJobColumns(),
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues = {},
        ) => {
          return await jobList({
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
    id: 'system-job-index',
  },
  gridEvents: {
    checkboxAll: syncCheckedRows,
    checkboxChange: syncCheckedRows,
  },
});

const checkedRows = ref<JobRecord[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);

onMounted(async () => {
  await loadGroupOptions();
});

async function loadGroupOptions() {
  const result = await jobGroupList({ pageNum: 1, pageSize: 100 });
  groupOptions.value = (result.items || []).map((item: JobGroupRecord) => ({
    label: item.name,
    value: item.id,
  }));
  gridApi.formApi.updateSchema([
    {
      componentProps: {
        options: groupOptions.value,
        placeholder: '请选择分组',
      },
      fieldName: 'groupId',
    },
  ]);
}

function syncCheckedRows() {
  checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as JobRecord[];
}

function shellBlockedReason() {
  if (!hasShellPermission.value) {
    return '当前账号缺少 Shell 任务权限';
  }
  if (!shellCapability.value.supported) {
    return shellCapability.value.disabledReason || '当前平台不支持 Shell 任务';
  }
  if (!shellCapability.value.enabled) {
    return shellCapability.value.disabledReason || '当前环境未启用 Shell 任务';
  }
  return '';
}

function isShellRow(row: JobRecord) {
  return row.taskType === 'shell';
}

function isShellBlocked(row: JobRecord) {
  return isShellRow(row) && shellBlockedReason() !== '';
}

function isPluginPaused(row: JobRecord) {
  return row.status === 'paused_by_plugin';
}

function openCreateModal() {
  jobFormApi.setData({});
  jobFormApi.open();
}

function openEditModal(row: JobRecord) {
  if (row.isBuiltin === 1) {
    jobFormApi.setData({ id: row.id });
    jobFormApi.open();
    return;
  }
  if (isShellBlocked(row)) {
    message.warning(shellBlockedReason());
    return;
  }
  if (isPluginPaused(row)) {
    message.warning(JOB_PLUGIN_PAUSED_TOOLTIP);
    return;
  }
  jobFormApi.setData({ id: row.id });
  jobFormApi.open();
}

async function handleDelete(ids: Array<number>) {
  await jobDelete(ids);
  message.success('删除成功');
  checkedRows.value = [];
  await gridApi.query();
}

function handleMultiDelete() {
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的 ${checkedRows.value.length} 个任务吗？`,
    onOk: async () => {
      await handleDelete(checkedRows.value.map((row) => row.id));
    },
  });
}

async function handleTrigger(row: JobRecord) {
  if (isShellBlocked(row)) {
    message.warning(shellBlockedReason());
    return;
  }
  const result = await jobTrigger(row.id);
  message.success(`已触发执行，日志编号 ${result.logId}`);
  await gridApi.query();
}

async function handleReset(row: JobRecord) {
  await jobReset(row.id);
  message.success('执行次数已重置');
  await gridApi.query();
}

function canEditRow(row: JobRecord) {
  return (
    row.isBuiltin !== 1 &&
    hasAccessByCodes([accessCodes.edit]) &&
    !isPluginPaused(row)
  );
}

function canDeleteRow(row: JobRecord) {
  return (
    getJobSourceKind(row) === 'user_created' &&
    hasAccessByCodes([accessCodes.remove])
  );
}

function hasMoreActions(row: JobRecord) {
  return canResetRow(row) || canDeleteRow(row);
}

function canTriggerRow(row: JobRecord) {
  return (
    hasAccessByCodes([accessCodes.trigger]) &&
    (row.status === 'enabled' || row.status === 'disabled') &&
    !isPluginPaused(row)
  );
}

function showPausedTriggerDisabled(row: JobRecord) {
  return (
    hasAccessByCodes([accessCodes.trigger]) && row.status === 'paused_by_plugin'
  );
}

function canResetRow(row: JobRecord) {
  return row.isBuiltin !== 1 && hasAccessByCodes([accessCodes.reset]);
}

function canViewRow(row: JobRecord) {
  return row.isBuiltin === 1;
}

function handleReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true" data-testid="job-page">
    <Grid table-title="定时任务列表">
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
            v-if="canCreateShellJob"
            data-testid="job-add"
            type="primary"
            @click="openCreateModal"
          >
            新 增
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <Tooltip
            v-if="showPausedTriggerDisabled(row)"
            :title="JOB_PLUGIN_PAUSED_TOOLTIP"
          >
            <ghost-button disabled :data-testid="`job-trigger-${row.id}`">
              立即执行
            </ghost-button>
          </Tooltip>
          <Tooltip
            v-if="canTriggerRow(row) && isShellBlocked(row)"
            :title="shellBlockedReason()"
          >
            <ghost-button disabled :data-testid="`job-trigger-${row.id}`">
              立即执行
            </ghost-button>
          </Tooltip>
          <ghost-button
            v-else-if="canTriggerRow(row)"
            :disabled="isPluginPaused(row)"
            :data-testid="`job-trigger-${row.id}`"
            @click.stop="handleTrigger(row)"
          >
            立即执行
          </ghost-button>

          <Tooltip
            v-if="canEditRow(row) && isShellBlocked(row)"
            :title="shellBlockedReason()"
          >
            <ghost-button disabled :data-testid="`job-edit-${row.id}`">
              编辑
            </ghost-button>
          </Tooltip>
          <ghost-button
            v-if="canViewRow(row)"
            :data-testid="`job-edit-${row.id}`"
            @click.stop="openEditModal(row)"
          >
            详情
          </ghost-button>

          <ghost-button
            v-else-if="canEditRow(row)"
            :data-testid="`job-edit-${row.id}`"
            @click.stop="openEditModal(row)"
          >
            编辑
          </ghost-button>

          <Dropdown
            v-if="hasMoreActions(row)"
            placement="bottomRight"
          >
            <template #overlay>
              <Menu>
                <MenuItem
                  v-if="canResetRow(row)"
                  :key="`reset-${row.id}`"
                  @click="handleReset(row)"
                >
                  <span :data-testid="`job-reset-${row.id}`">重置计数</span>
                </MenuItem>
                <MenuItem v-if="canDeleteRow(row)" :key="`delete-${row.id}`">
                  <Popconfirm
                    placement="left"
                    title="确认删除该任务吗？"
                    @confirm="handleDelete([row.id])"
                  >
                    <span :data-testid="`job-delete-${row.id}`">删除</span>
                  </Popconfirm>
                </MenuItem>
              </Menu>
            </template>
            <a-button
              :data-testid="`job-more-${row.id}`"
              size="small"
              type="link"
            >
              更多
            </a-button>
          </Dropdown>
        </Space>
      </template>
    </Grid>

    <JobFormModal @reload="handleReload" />
  </Page>
</template>
