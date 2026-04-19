<script setup lang="ts">
import type {
  JobHandlerOption,
  JobPayload,
  JobRecord,
} from '#/api/system/job/model';
import type { JobGroupRecord } from '#/api/system/jobGroup/model';
import type { VbenFormSchema } from '#/adapter/form';

import { useAccess } from '@vben/access';
import { useVbenModal } from '@vben/common-ui';

import { computed, ref } from 'vue';

import { Alert, message, Tabs, TabPane } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import {
  jobCreate,
  jobCronPreview,
  jobDetail,
  jobUpdate,
} from '#/api/system/job';
import { jobHandlerList } from '#/api/system/jobHandler';
import { jobGroupList } from '#/api/system/jobGroup';
import {
  publicFrontendSettings,
} from '#/runtime/public-frontend';

import JobFormHandler from './form-handler.vue';
import JobFormShell from './form-shell.vue';

const emit = defineEmits<{ reload: [] }>();

const accessCodes = {
  shell: 'system:job:shell',
} as const;

type TaskTypeKey = 'handler' | 'shell';

interface GroupOption {
  isDefault: boolean;
  label: string;
  value: number;
}

const { hasAccessByCodes } = useAccess();

const currentRecord = ref<JobRecord | null>(null);
const activeTaskType = ref<TaskTypeKey>('handler');
const cronPreviewError = ref('');
const cronPreviewTimes = ref<string[]>([]);
const groupOptions = ref<GroupOption[]>([]);
const handlerOptions = ref<JobHandlerOption[]>([]);

const handlerFormRef = ref<InstanceType<typeof JobFormHandler>>();
const shellFormRef = ref<InstanceType<typeof JobFormShell>>();

const hasShellPermission = computed(() =>
  hasAccessByCodes([accessCodes.shell]),
);
const shellCapability = computed(() => publicFrontendSettings.cron.shell);
const shellVisible = computed(() => {
  return (
    hasShellPermission.value &&
    shellCapability.value.enabled &&
    shellCapability.value.supported
  );
});
const shellUnavailableReason = computed(() => {
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
});

const title = computed(() =>
  currentRecord.value ? '编辑定时任务' : '新增定时任务',
);
const isBuiltin = computed(() => currentRecord.value?.isBuiltin === 1);

const [CommonForm, commonFormApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    formItemClass: 'col-span-1',
    labelWidth: 96,
  },
  schema: buildCommonSchema([], false),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

function buildCommonSchema(
  options: GroupOption[],
  builtin: boolean,
): VbenFormSchema[] {
  return [
    {
      component: 'Select',
      componentProps: {
        disabled: builtin,
        options,
        placeholder: '请选择任务分组',
      },
      fieldName: 'groupId',
      label: '所属分组',
      rules: 'required',
    },
    {
      component: 'Input',
      componentProps: {
        disabled: builtin,
        placeholder: '请输入任务名称',
      },
      fieldName: 'name',
      label: '任务名称',
      rules: 'required',
    },
    {
      component: 'Textarea',
      componentProps: {
        placeholder: '请输入任务描述',
        rows: 3,
      },
      fieldName: 'description',
      formItemClass: 'col-span-2',
      label: '任务描述',
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: '请输入 Cron 表达式',
      },
      fieldName: 'cronExpr',
      label: 'Cron 表达式',
      rules: 'required',
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: '请输入 IANA 时区，如 Asia/Shanghai',
      },
      defaultValue: 'Asia/Shanghai',
      fieldName: 'timezone',
      label: '任务时区',
      rules: 'required',
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        disabled: builtin,
        optionType: 'button',
        options: [
          { label: '主节点', value: 'master_only' },
          { label: '所有节点', value: 'all_node' },
        ],
      },
      defaultValue: 'master_only',
      fieldName: 'scope',
      label: '调度范围',
      rules: 'required',
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        disabled: builtin,
        optionType: 'button',
        options: [
          { label: '单例执行', value: 'singleton' },
          { label: '并行执行', value: 'parallel' },
        ],
      },
      defaultValue: 'singleton',
      fieldName: 'concurrency',
      label: '并发策略',
      rules: 'required',
    },
    {
      component: 'InputNumber',
      componentProps: {
        min: 1,
        precision: 0,
        style: { width: '100%' },
      },
      defaultValue: 1,
      dependencies: {
        show: (values) => values.concurrency === 'parallel',
        triggerFields: ['concurrency'],
      },
      fieldName: 'maxConcurrency',
      label: '最大并发',
      rules: 'required',
    },
    {
      component: 'InputNumber',
      componentProps: {
        max: 86400,
        min: 1,
        precision: 0,
        style: { width: '100%' },
      },
      defaultValue: 300,
      fieldName: 'timeoutSeconds',
      label: '超时时间(秒)',
      rules: 'required',
    },
    {
      component: 'InputNumber',
      componentProps: {
        min: 0,
        precision: 0,
        style: { width: '100%' },
      },
      defaultValue: 0,
      fieldName: 'maxExecutions',
      label: '最大执行次数',
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
        options: [
          { label: '启用', value: 'enabled' },
          { label: '停用', value: 'disabled' },
        ],
      },
      defaultValue: 'disabled',
      fieldName: 'status',
      label: '任务状态',
      rules: 'required',
    },
    {
      component: 'Select',
      componentProps: {
        options: [
          { label: '跟随系统', value: '' },
          { label: '按天保留', value: 'days' },
          { label: '按条数保留', value: 'count' },
          { label: '不自动清理', value: 'none' },
        ],
        placeholder: '请选择日志保留策略',
      },
      defaultValue: '',
      fieldName: 'retentionMode',
      label: '日志保留',
    },
    {
      component: 'InputNumber',
      componentProps: {
        min: 1,
        precision: 0,
        style: { width: '100%' },
      },
      dependencies: {
        show: (values) =>
          values.retentionMode === 'days' || values.retentionMode === 'count',
        triggerFields: ['retentionMode'],
      },
      fieldName: 'retentionValue',
      label: '保留阈值',
    },
  ];
}

function parseRetention(raw: string) {
  if (!raw) {
    return {
      retentionMode: '',
      retentionValue: undefined,
    };
  }
  try {
    const parsed = JSON.parse(raw);
    return {
      retentionMode: parsed?.mode || '',
      retentionValue:
        typeof parsed?.value === 'number' && parsed.value > 0
          ? parsed.value
          : undefined,
    };
  } catch {
    return {
      retentionMode: '',
      retentionValue: undefined,
    };
  }
}

function mapGroupOptions(groups: JobGroupRecord[]) {
  return groups.map((item) => ({
    isDefault: item.isDefault === 1,
    label: item.name,
    value: item.id,
  }));
}

function getDefaultGroupId() {
  const defaultGroup = groupOptions.value.find((item) => item.isDefault);
  return defaultGroup?.value || groupOptions.value[0]?.value || 0;
}

function rebuildCommonSchema() {
  commonFormApi.setState({
    schema: buildCommonSchema(groupOptions.value, isBuiltin.value),
  });
}

async function fillCommonForm(record?: JobRecord | null) {
  if (!record) {
    await commonFormApi.setValues({
      concurrency: 'singleton',
      cronExpr: '0 0 1 1 *',
      description: '',
      groupId: getDefaultGroupId(),
      maxConcurrency: 1,
      maxExecutions: 0,
      name: '',
      retentionMode: '',
      retentionValue: undefined,
      scope: 'master_only',
      status: 'disabled',
      timeoutSeconds: 300,
      timezone: 'Asia/Shanghai',
    });
    return;
  }

  const retention = parseRetention(record.logRetentionOverride);
  await commonFormApi.setValues({
    concurrency: record.concurrency,
    cronExpr: record.cronExpr,
    description: record.description,
    groupId: record.groupId,
    maxConcurrency: record.maxConcurrency,
    maxExecutions: record.maxExecutions,
    name: record.name,
    retentionMode: retention.retentionMode,
    retentionValue: retention.retentionValue,
    scope: record.scope,
    status: record.status === 'enabled' ? 'enabled' : 'disabled',
    timeoutSeconds: record.timeoutSeconds,
    timezone: record.timezone,
  });
}

async function loadModalData(id?: number) {
  const [groupResult, handlerResult, record] = await Promise.all([
    jobGroupList({ pageNum: 1, pageSize: 100 }),
    jobHandlerList(),
    typeof id === 'number' ? jobDetail(id) : Promise.resolve(null),
  ]);

  groupOptions.value = mapGroupOptions(groupResult.items || []);
  handlerOptions.value = handlerResult.list || [];
  currentRecord.value = record;
  rebuildCommonSchema();
  await fillCommonForm(record);

  if (record?.taskType === 'shell') {
    activeTaskType.value = shellVisible.value ? 'shell' : 'handler';
  } else {
    activeTaskType.value = 'handler';
  }

  await handlerFormRef.value?.load(record);
  await shellFormRef.value?.load(record);
}

const [Modal, modalApi] = useVbenModal({
  class: 'w-[980px]',
  fullscreenButton: true,
  onClosed: async () => {
    currentRecord.value = null;
    cronPreviewError.value = '';
    cronPreviewTimes.value = [];
    activeTaskType.value = 'handler';
    await commonFormApi.resetForm();
    await handlerFormRef.value?.reset();
    await shellFormRef.value?.reset();
  },
  onConfirm: handleConfirm,
  onOpenChange: async (open) => {
    if (!open) {
      return;
    }
    modalApi.setState({ loading: true });
    const { id } = modalApi.getData<{ id?: number }>();
    await loadModalData(id);
    modalApi.setState({ loading: false });
  },
});

async function handlePreviewCron() {
  cronPreviewError.value = '';
  try {
    const values = await commonFormApi.getValues<Record<string, any>>();
    const preview = await jobCronPreview(values.cronExpr, values.timezone);
    cronPreviewTimes.value = preview.times || [];
  } catch (error) {
    cronPreviewTimes.value = [];
    cronPreviewError.value =
      error instanceof Error ? error.message : 'Cron 预览失败';
  }
}

function buildRetentionOverride(values: Record<string, any>) {
  const mode = values.retentionMode || '';
  if (!mode) {
    return null;
  }
  if (mode === 'none') {
    return {
      mode: 'none',
      value: 0,
    };
  }
  const value = Number(values.retentionValue || 0);
  if (value <= 0) {
    throw new Error('日志保留阈值必须大于 0');
  }
  return {
    mode,
    value,
  };
}

async function buildPayload() {
  const { valid } = await commonFormApi.validate();
  if (!valid) {
    throw new Error('请完善公共调度配置');
  }
  const values = await commonFormApi.getValues<Record<string, any>>();
  const specific =
    activeTaskType.value === 'handler'
      ? await handlerFormRef.value?.validateAndBuild()
      : await shellFormRef.value?.validateAndBuild();
  if (!specific) {
    throw new Error('请完善任务类型配置');
  }

  return {
    concurrency: values.concurrency,
    cronExpr: values.cronExpr,
    description: values.description || '',
    groupId: Number(values.groupId),
    logRetentionOverride: buildRetentionOverride(values),
    maxConcurrency:
      values.concurrency === 'parallel'
        ? Number(values.maxConcurrency || 1)
        : 1,
    maxExecutions: Number(values.maxExecutions || 0),
    name: values.name,
    status: values.status,
    taskType: activeTaskType.value,
    timeoutSeconds: Number(values.timeoutSeconds),
    timezone: values.timezone,
    scope: values.scope,
    ...(activeTaskType.value === 'handler'
      ? {
          env: {},
          handlerRef: specific.handlerRef,
          params: specific.params,
          shellCmd: '',
          workDir: '',
        }
      : {
          env: specific.env,
          handlerRef: '',
          params: {},
          shellCmd: specific.shellCmd,
          workDir: specific.workDir,
        }),
  } satisfies JobPayload;
}

async function handleConfirm() {
  try {
    modalApi.lock(true);
    const payload = await buildPayload();
    if (currentRecord.value) {
      await jobUpdate(currentRecord.value.id, payload);
      message.success('更新成功');
    } else {
      await jobCreate(payload);
      message.success('创建成功');
    }
    emit('reload');
    modalApi.close();
  } catch (error) {
    const msg =
      error instanceof Error ? error.message : '保存定时任务失败';
    message.error(msg);
  } finally {
    modalApi.lock(false);
  }
}
</script>

<template>
  <Modal
    :title="title"
    data-testid="job-form-modal"
  >
    <div class="space-y-4">
      <Alert
        v-if="isBuiltin"
        message="系统内置任务的分组、名称、任务类型、处理器引用、处理器参数、调度范围和并发策略已锁定。"
        show-icon
        type="info"
      />

      <CommonForm />

      <div class="rounded-md border border-border px-4 py-4">
        <div class="mb-3 flex items-center justify-between">
          <div>
            <div class="text-sm font-medium">Cron 预览</div>
            <div class="text-xs text-foreground/60">
              校验表达式与时区后，展示最近 5 次触发时间。
            </div>
          </div>
          <a-button
            data-testid="job-cron-preview"
            @click="handlePreviewCron"
          >
            预 览
          </a-button>
        </div>

        <Alert
          v-if="cronPreviewError"
          :message="cronPreviewError"
          show-icon
          type="error"
        />
        <ul
          v-else-if="cronPreviewTimes.length > 0"
          class="space-y-1 text-sm"
        >
          <li
            v-for="item in cronPreviewTimes"
            :key="item"
            class="rounded bg-accent px-3 py-2"
          >
            {{ item }}
          </li>
        </ul>
        <div v-else class="text-sm text-foreground/60">
          点击预览按钮查看下一次执行时间。
        </div>
      </div>

      <div class="rounded-md border border-border px-4 py-4">
        <div class="mb-3 flex items-center justify-between">
          <div>
            <div class="text-sm font-medium">任务类型配置</div>
            <div class="text-xs text-foreground/60">
              Handler 任务通过注册表执行，Shell 任务直接在宿主节点执行命令。
            </div>
          </div>
          <Alert
            v-if="!shellVisible"
            :message="shellUnavailableReason"
            type="warning"
            show-icon
          />
        </div>

        <Tabs
          v-model:activeKey="activeTaskType"
          data-testid="job-form-task-tabs"
        >
          <TabPane
            key="handler"
            tab="Handler"
            data-testid="job-form-tab-handler"
          >
            <JobFormHandler
              ref="handlerFormRef"
              :builtin="isBuiltin"
              :handler-options="handlerOptions"
            />
          </TabPane>

          <TabPane
            v-if="shellVisible && !isBuiltin"
            key="shell"
            tab="Shell"
            data-testid="job-form-tab-shell"
          >
            <JobFormShell ref="shellFormRef" />
          </TabPane>
        </Tabs>
      </div>
    </div>
  </Modal>
</template>
