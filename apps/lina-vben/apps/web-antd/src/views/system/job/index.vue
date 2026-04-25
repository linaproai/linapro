<script setup lang="ts">
import type { JobRecord } from '#/api/system/job/model';
import type { JobGroupRecord } from '#/api/system/jobGroup/model';

import { useAccess } from '@vben/access';
import { Page, useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';

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
  JOB_STATUS_FILTER_OPTIONS,
  getJobPluginPausedTooltip,
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
          placeholder: $t('pages.system.job.placeholders.selectGroup'),
        },
        fieldName: 'groupId',
        label: $t('pages.system.job.fields.group'),
      },
      {
        component: 'Select',
        componentProps: {
          options: JOB_STATUS_FILTER_OPTIONS,
          placeholder: $t('pages.system.job.placeholders.selectStatus'),
        },
        fieldName: 'status',
        label: $t('pages.system.job.fields.status'),
      },
      {
        component: 'Input',
        fieldName: 'keyword',
        label: $t('pages.system.job.fields.keyword'),
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
        placeholder: $t('pages.system.job.placeholders.selectGroup'),
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
    return $t('pages.system.job.messages.noShellPermission');
  }
  if (!shellCapability.value.supported) {
    return (
      shellCapability.value.disabledReason ||
      $t('pages.system.job.messages.platformUnsupported')
    );
  }
  if (!shellCapability.value.enabled) {
    return (
      shellCapability.value.disabledReason ||
      $t('pages.system.job.messages.environmentDisabled')
    );
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
    message.warning(getJobPluginPausedTooltip());
    return;
  }
  jobFormApi.setData({ id: row.id });
  jobFormApi.open();
}

async function handleDelete(ids: Array<number>) {
  await jobDelete(ids);
  message.success($t('pages.common.deleteSuccess'));
  checkedRows.value = [];
  await gridApi.query();
}

function handleMultiDelete() {
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('pages.system.job.messages.deleteSelectedConfirm', {
      count: checkedRows.value.length,
    }),
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
  message.success($t('pages.system.job.messages.triggerSuccess', { logId: result.logId }));
  await gridApi.query();
}

async function handleReset(row: JobRecord) {
  await jobReset(row.id);
  message.success($t('pages.system.job.messages.resetSuccess'));
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
    <Grid :table-title="$t('pages.system.job.tableTitle')">
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
            v-if="canCreateShellJob"
            data-testid="job-add"
            type="primary"
            @click="openCreateModal"
          >
            {{ $t('pages.common.add') }}
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <Space>
          <Tooltip
            v-if="showPausedTriggerDisabled(row)"
            :title="getJobPluginPausedTooltip()"
          >
            <ghost-button disabled :data-testid="`job-trigger-${row.id}`">
              {{ $t('pages.system.job.actions.runNow') }}
            </ghost-button>
          </Tooltip>
          <Tooltip
            v-if="canTriggerRow(row) && isShellBlocked(row)"
            :title="shellBlockedReason()"
          >
            <ghost-button disabled :data-testid="`job-trigger-${row.id}`">
              {{ $t('pages.system.job.actions.runNow') }}
            </ghost-button>
          </Tooltip>
          <ghost-button
            v-else-if="canTriggerRow(row)"
            :disabled="isPluginPaused(row)"
            :data-testid="`job-trigger-${row.id}`"
            @click.stop="handleTrigger(row)"
          >
            {{ $t('pages.system.job.actions.runNow') }}
          </ghost-button>

          <Tooltip
            v-if="canEditRow(row) && isShellBlocked(row)"
            :title="shellBlockedReason()"
          >
            <ghost-button disabled :data-testid="`job-edit-${row.id}`">
              {{ $t('pages.common.edit') }}
            </ghost-button>
          </Tooltip>
          <ghost-button
            v-if="canViewRow(row)"
            :data-testid="`job-edit-${row.id}`"
            @click.stop="openEditModal(row)"
          >
            {{ $t('pages.common.detail') }}
          </ghost-button>

          <ghost-button
            v-else-if="canEditRow(row)"
            :data-testid="`job-edit-${row.id}`"
            @click.stop="openEditModal(row)"
          >
            {{ $t('pages.common.edit') }}
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
                  <span :data-testid="`job-reset-${row.id}`">{{ $t('pages.system.job.actions.resetCount') }}</span>
                </MenuItem>
                <MenuItem v-if="canDeleteRow(row)" :key="`delete-${row.id}`">
                  <Popconfirm
                    placement="left"
                    :title="$t('pages.system.job.messages.deleteConfirm')"
                    @confirm="handleDelete([row.id])"
                  >
                    <span :data-testid="`job-delete-${row.id}`">{{ $t('pages.common.delete') }}</span>
                  </Popconfirm>
                </MenuItem>
              </Menu>
            </template>
            <a-button
              :data-testid="`job-more-${row.id}`"
              size="small"
              type="link"
            >
              {{ $t('pages.common.more') }}
            </a-button>
          </Dropdown>
        </Space>
      </template>
    </Grid>

    <JobFormModal @reload="handleReload" />
  </Page>
</template>
