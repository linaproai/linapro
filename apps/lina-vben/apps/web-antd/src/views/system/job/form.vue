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
import {
  getJobRetentionFieldHelp,
  JOB_CONCURRENCY_FIELD_HELP,
  JOB_CONCURRENCY_OPTIONS,
  JOB_CRON_FIELD_HELP,
  JOB_MAX_EXECUTIONS_FIELD_HELP,
  JOB_SCOPE_FIELD_HELP,
  JOB_SCOPE_OPTIONS,
  JOB_TIMEOUT_FIELD_HELP,
} from '#/api/system/job/meta';
import { jobHandlerList } from '#/api/system/jobHandler';
import { jobGroupList } from '#/api/system/jobGroup';
import { publicFrontendSettings } from '#/runtime/public-frontend';

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

const COMMON_TIMEZONE_OPTIONS = [
  'Asia/Shanghai',
  'Asia/Hong_Kong',
  'Asia/Singapore',
  'Asia/Tokyo',
  'UTC',
  'Europe/London',
  'Europe/Berlin',
  'America/New_York',
  'America/Los_Angeles',
] as const;
const JOB_FORM_STATUS_OPTIONS = ['enabled', 'disabled'] as const;

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
const retentionHelp = computed(() =>
  getJobRetentionFieldHelp(publicFrontendSettings.cron.logRetention),
);
const currentSystemTimezone = computed(
  () => publicFrontendSettings.cron.timezone.current || 'Asia/Shanghai',
);

const [CommonForm, commonFormApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    formItemClass: 'col-span-1',
    labelWidth: 112,
  },
  schema: buildCommonSchema(
    [],
    false,
    retentionHelp.value,
    currentSystemTimezone.value,
  ),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

function buildTimezoneOptions(currentTimezone: string) {
  const values = new Set<string>(COMMON_TIMEZONE_OPTIONS);
  const trimmedCurrentTimezone = currentTimezone.trim();
  if (trimmedCurrentTimezone) {
    values.add(trimmedCurrentTimezone);
  }

  return Array.from(values).map((value) => ({
    label: value === trimmedCurrentTimezone ? `${value}（当前系统）` : value,
    value,
  }));
}

function buildCommonSchema(
  options: GroupOption[],
  builtin: boolean,
  retentionHint: ReturnType<typeof getJobRetentionFieldHelp>,
  currentTimezone: string,
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
        'aria-label': '定时表达式',
        'data-testid': 'job-cron-editor',
        autocomplete: 'off',
        class: 'job-cron-code-input',
        placeholder: '支持 5 段或 6 段，如 17 3 * * *',
        spellcheck: false,
      },
      fieldName: 'cronExpr',
      help: JOB_CRON_FIELD_HELP,
      label: '定时表达式',
      rules: 'required',
    },
    {
      component: 'AutoComplete',
      componentProps: {
        options: buildTimezoneOptions(currentTimezone),
        placeholder: '可选择常用时区或输入自定义 IANA 时区',
      },
      defaultValue: currentTimezone,
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
        options: JOB_SCOPE_OPTIONS,
      },
      defaultValue: 'master_only',
      fieldName: 'scope',
      help: JOB_SCOPE_FIELD_HELP,
      label: '调度范围',
      rules: 'required',
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        disabled: builtin,
        optionType: 'button',
        options: JOB_CONCURRENCY_OPTIONS,
      },
      defaultValue: 'singleton',
      fieldName: 'concurrency',
      help: JOB_CONCURRENCY_FIELD_HELP,
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
      help: JOB_TIMEOUT_FIELD_HELP,
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
      help: JOB_MAX_EXECUTIONS_FIELD_HELP,
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
      help: retentionHint,
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
    schema: buildCommonSchema(
      groupOptions.value,
      isBuiltin.value,
      retentionHelp.value,
      currentSystemTimezone.value,
    ),
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
      timezone: currentSystemTimezone.value,
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
    const cronExpr = ensureCronExpressionValue(values.cronExpr);
    const timezone = ensureTimezoneValue(values.timezone);
    const preview = await jobCronPreview(cronExpr, timezone);
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

function splitCronExpressionFields(value: string) {
  const trimmedValue = value.trim();
  if (!trimmedValue) {
    return [];
  }
  return trimmedValue.split(/\s+/);
}

function ensureCronExpressionValue(value: unknown) {
  const cronExpr = String(value || '').trim();
  if (!cronExpr) {
    throw new Error('定时表达式不能为空');
  }
  if (cronExpr.length > 128) {
    throw new Error('定时表达式长度不能超过 128 个字符');
  }

  const fields = splitCronExpressionFields(cronExpr);
  if (fields.length !== 5 && fields.length !== 6) {
    throw new Error('定时表达式仅支持 5 段或 6 段');
  }
  if (fields.length === 6 && fields[0] === '#') {
    throw new Error(
      '6 段定时表达式的秒位必须填写具体值，5 段表达式无需手工填写 #',
    );
  }
  return cronExpr;
}

function isValidTimezoneValue(value: string) {
  const timezone = value.trim();
  if (!timezone) {
    return false;
  }
  try {
    new Intl.DateTimeFormat('zh-CN', { timeZone: timezone });
    return true;
  } catch {
    return false;
  }
}

function ensureTimezoneValue(value: unknown) {
  const timezone = String(value || '').trim();
  if (!timezone) {
    throw new Error('任务时区不能为空');
  }
  if (!isValidTimezoneValue(timezone)) {
    throw new Error('任务时区不合法');
  }
  return timezone;
}

function ensureIntegerRangeValue(
  label: string,
  value: unknown,
  options: {
    allowZero?: boolean;
    max?: number;
    min: number;
  },
) {
  const numericValue = Number(value);
  if (!Number.isInteger(numericValue)) {
    throw new Error(`${label}必须为整数`);
  }
  if (options.allowZero && numericValue === 0) {
    return numericValue;
  }
  if (numericValue < options.min) {
    throw new Error(`${label}不能小于 ${options.min}`);
  }
  if (
    typeof options.max === 'number' &&
    Number.isFinite(options.max) &&
    numericValue > options.max
  ) {
    throw new Error(`${label}不能大于 ${options.max}`);
  }
  return numericValue;
}

function ensureSelectValue(
  label: string,
  value: unknown,
  supportedValues: readonly string[],
) {
  const text = String(value || '').trim();
  if (!supportedValues.includes(text)) {
    throw new Error(`${label}配置不合法`);
  }
  return text;
}

function validateCommonFormValues(values: Record<string, any>) {
  const groupId = Number(values.groupId);
  if (!Number.isInteger(groupId) || groupId <= 0) {
    throw new Error('请选择任务分组');
  }
  const jobName = String(values.name || '').trim();
  if (jobName === '') {
    throw new Error('任务名称不能为空');
  }
  if (jobName.length > 128) {
    throw new Error('任务名称长度不能超过 128 个字符');
  }

  ensureCronExpressionValue(values.cronExpr);
  ensureTimezoneValue(values.timezone);
  ensureSelectValue(
    '调度范围',
    values.scope,
    JOB_SCOPE_OPTIONS.map((item) => String(item.value)),
  );
  ensureSelectValue(
    '并发策略',
    values.concurrency,
    JOB_CONCURRENCY_OPTIONS.map((item) => String(item.value)),
  );
  ensureSelectValue('任务状态', values.status, JOB_FORM_STATUS_OPTIONS);
  ensureIntegerRangeValue('超时时间(秒)', values.timeoutSeconds, {
    max: 86400,
    min: 1,
  });
  ensureIntegerRangeValue('最大执行次数', values.maxExecutions ?? 0, {
    allowZero: true,
    min: 0,
  });
  if (values.concurrency === 'parallel') {
    ensureIntegerRangeValue('最大并发', values.maxConcurrency, {
      max: 100,
      min: 1,
    });
  }
}

async function buildPayload() {
  const { errors, valid } = await commonFormApi.validate();
  if (!valid) {
    const firstError = Object.values(errors || {}).find((value) => !!value);
    throw new Error(
      typeof firstError === 'string' && firstError
        ? firstError
        : '请完善公共调度配置',
    );
  }
  const values = await commonFormApi.getValues<Record<string, any>>();
  validateCommonFormValues(values);

  const cronExpr = ensureCronExpressionValue(values.cronExpr);
  const timezone = ensureTimezoneValue(values.timezone);

  const commonPayload = {
    concurrency: values.concurrency,
    cronExpr,
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
    timezone,
    scope: values.scope,
  };

  if (activeTaskType.value === 'handler') {
    const handlerPayload = await handlerFormRef.value?.validateAndBuild();
    if (!handlerPayload) {
      throw new Error('请完善 Handler 配置');
    }
    return {
      ...commonPayload,
      env: {},
      handlerRef: handlerPayload.handlerRef,
      params: handlerPayload.params,
      shellCmd: '',
      workDir: '',
    } satisfies JobPayload;
  }

  const shellPayload = await shellFormRef.value?.validateAndBuild();
  if (!shellPayload) {
    throw new Error('请完善 Shell 配置');
  }
  return {
    ...commonPayload,
    env: shellPayload.env,
    handlerRef: '',
    params: {},
    shellCmd: shellPayload.shellCmd,
    workDir: shellPayload.workDir,
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
    const msg = error instanceof Error ? error.message : '保存定时任务失败';
    message.error(msg);
  } finally {
    modalApi.lock(false);
  }
}
</script>

<template>
  <Modal :title="title" data-testid="job-form-modal">
    <div class="space-y-4">
      <div
        v-if="isBuiltin"
        class="py-[5px]"
        data-testid="job-builtin-common-lock-alert"
      >
        <Alert
          message="系统内置任务的分组、名称、任务类型、处理器引用、处理器参数、调度范围和并发策略已锁定。"
          show-icon
          type="info"
        />
      </div>

      <CommonForm />

      <div class="rounded-md border border-border px-4 py-4">
        <div class="mb-3 flex items-center justify-between">
          <div>
            <div class="text-sm font-medium">Cron 预览</div>
            <div class="text-xs text-foreground/60">
              支持 5 段或 6 段 Cron；5 段表达式会在运行时自动补 # 秒占位。
            </div>
          </div>
          <a-button data-testid="job-cron-preview" @click="handlePreviewCron">
            预 览
          </a-button>
        </div>

        <Alert
          v-if="cronPreviewError"
          :message="cronPreviewError"
          show-icon
          type="error"
        />
        <ul v-else-if="cronPreviewTimes.length > 0" class="space-y-1 text-sm">
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

<style scoped>
:deep(.job-cron-code-input) {
  background: var(--ant-color-fill-tertiary, #f5f5f5);
  border: 1px solid var(--ant-color-border-secondary, #f0f0f0);
  border-radius: 8px;
  box-shadow: inset 0 0 0 1px rgb(0 0 0 / 2%);
  font-family:
    ui-monospace, 'SFMono-Regular', SFMono-Regular, Menlo, Monaco, Consolas,
    'Liberation Mono', 'Courier New', monospace;
  letter-spacing: 0.02em;
}
</style>
