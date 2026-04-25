<script setup lang="ts">
import type {
  JobPayload,
  JobRecord,
} from '#/api/system/job/model';
import type { JobGroupRecord } from '#/api/system/jobGroup/model';
import type { VbenFormSchema } from '#/adapter/form';

import { useAccess } from '@vben/access';
import { useVbenModal } from '@vben/common-ui';

import { computed, ref } from 'vue';

import { $t } from '#/locales';
import { Alert, Descriptions, DescriptionsItem, Tag, message } from 'ant-design-vue';

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
  getJobSourceColor,
  getJobSourceKind,
  getJobSourceLabel,
  parsePluginIdFromHandlerRef,
} from '#/api/system/job/meta';
import { jobGroupList } from '#/api/system/jobGroup';
import { publicFrontendSettings } from '#/runtime/public-frontend';

import JobFormShell from './form-shell.vue';

const emit = defineEmits<{ reload: [] }>();

const accessCodes = {
  shell: 'system:job:shell',
} as const;

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
const cronPreviewError = ref('');
const cronPreviewTimes = ref<string[]>([]);
const groupOptions = ref<GroupOption[]>([]);

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
});

const isBuiltin = computed(() => currentRecord.value?.isBuiltin === 1);
const title = computed(() => {
  if (isBuiltin.value) {
    return $t('pages.system.job.drawer.detailTitle');
  }
  return currentRecord.value
    ? $t('pages.system.job.drawer.editTitle')
    : $t('pages.system.job.drawer.createTitle');
});
const retentionHelp = computed(() =>
  getJobRetentionFieldHelp(publicFrontendSettings.cron.logRetention),
);
const currentSystemTimezone = computed(
  () => publicFrontendSettings.cron.timezone.current || 'Asia/Shanghai',
);
const currentSourceKind = computed(() => getJobSourceKind(currentRecord.value));
const currentSourceLabel = computed(() => getJobSourceLabel(currentSourceKind.value));
const currentPluginId = computed(() =>
  parsePluginIdFromHandlerRef(currentRecord.value?.handlerRef),
);
const readonlyParamsText = computed(() => {
  const raw = currentRecord.value?.params || '{}';
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw || '{}';
  }
});

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
    label:
      value === trimmedCurrentTimezone
        ? $t('pages.system.job.placeholders.currentTimezone', { value })
        : value,
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
        placeholder: $t('pages.system.job.placeholders.selectGroup'),
      },
      fieldName: 'groupId',
      label: $t('pages.system.job.fields.group'),
      rules: 'required',
    },
    {
      component: 'Input',
      componentProps: {
        disabled: builtin,
        placeholder: $t('pages.system.job.placeholders.name'),
      },
      fieldName: 'name',
      label: $t('pages.system.job.fields.name'),
      rules: 'required',
    },
    {
      component: 'Textarea',
      componentProps: {
        disabled: builtin,
        placeholder: $t('pages.system.job.placeholders.description'),
        rows: 3,
      },
      fieldName: 'description',
      formItemClass: 'col-span-2',
      label: $t('pages.fields.description'),
    },
    {
      component: 'Input',
      componentProps: {
        'aria-label': $t('pages.system.job.fields.cronExpr'),
        'data-testid': 'job-cron-editor',
        autocomplete: 'off',
        class: 'job-cron-code-input',
        disabled: builtin,
        placeholder: $t('pages.system.job.placeholders.cronExpr'),
        spellcheck: false,
      },
      fieldName: 'cronExpr',
      help: JOB_CRON_FIELD_HELP,
      label: $t('pages.system.job.fields.cronExpr'),
      rules: 'required',
    },
    {
      component: 'AutoComplete',
      componentProps: {
        disabled: builtin,
        options: buildTimezoneOptions(currentTimezone),
        placeholder: $t('pages.system.job.placeholders.timezone'),
      },
      defaultValue: currentTimezone,
      fieldName: 'timezone',
      label: $t('pages.system.job.fields.timezone'),
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
      label: $t('pages.system.job.fields.scope'),
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
      label: $t('pages.system.job.fields.concurrency'),
      rules: 'required',
    },
    {
      component: 'InputNumber',
      componentProps: {
        disabled: builtin,
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
      label: $t('pages.system.job.fields.maxConcurrency'),
      rules: 'required',
    },
    {
      component: 'InputNumber',
      componentProps: {
        disabled: builtin,
        max: 86400,
        min: 1,
        precision: 0,
        style: { width: '100%' },
      },
      defaultValue: 300,
      fieldName: 'timeoutSeconds',
      help: JOB_TIMEOUT_FIELD_HELP,
      label: $t('pages.system.job.fields.timeoutSeconds'),
      rules: 'required',
    },
    {
      component: 'InputNumber',
      componentProps: {
        disabled: builtin,
        min: 0,
        precision: 0,
        style: { width: '100%' },
      },
      defaultValue: 0,
      fieldName: 'maxExecutions',
      help: JOB_MAX_EXECUTIONS_FIELD_HELP,
      label: $t('pages.system.job.fields.maxExecutions'),
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        disabled: builtin,
        optionType: 'button',
        options: builtin
          ? [
              { label: $t('pages.system.job.status.enabled'), value: 'enabled' },
              { label: $t('pages.system.job.status.disabled'), value: 'disabled' },
              {
                label: $t('pages.system.job.status.pluginUnavailable'),
                value: 'paused_by_plugin',
              },
            ]
          : [
              { label: $t('pages.system.job.status.enabled'), value: 'enabled' },
              { label: $t('pages.system.job.status.disabled'), value: 'disabled' },
            ],
      },
      defaultValue: 'disabled',
      fieldName: 'status',
      label: $t('pages.system.job.fields.status'),
      rules: 'required',
    },
    {
      component: 'Select',
      componentProps: {
        disabled: builtin,
        options: [
          { label: $t('pages.system.job.retention.followSystem'), value: '' },
          { label: $t('pages.system.job.retention.days'), value: 'days' },
          { label: $t('pages.system.job.retention.count'), value: 'count' },
          { label: $t('pages.system.job.retention.none'), value: 'none' },
        ],
        placeholder: $t('pages.system.job.placeholders.selectRetention'),
      },
      defaultValue: '',
      fieldName: 'retentionMode',
      help: retentionHint,
      label: $t('pages.system.job.fields.retention'),
    },
    {
      component: 'InputNumber',
      componentProps: {
        disabled: builtin,
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
      label: $t('pages.system.job.fields.retentionValue'),
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
    status:
      record.status === 'paused_by_plugin'
        ? 'paused_by_plugin'
        : record.status === 'enabled'
          ? 'enabled'
          : 'disabled',
    timeoutSeconds: record.timeoutSeconds,
    timezone: record.timezone,
  });
}

async function loadModalData(id?: number) {
  const [groupResult, record] = await Promise.all([
    jobGroupList({ pageNum: 1, pageSize: 100 }),
    typeof id === 'number' ? jobDetail(id) : Promise.resolve(null),
  ]);

  groupOptions.value = mapGroupOptions(groupResult.items || []);
  currentRecord.value = record;
  rebuildCommonSchema();
  await fillCommonForm(record);
  await shellFormRef.value?.load(record);
}

const [Modal, modalApi] = useVbenModal({
  class: 'w-[980px]',
  fullscreenButton: true,
  onClosed: async () => {
    currentRecord.value = null;
    cronPreviewError.value = '';
    cronPreviewTimes.value = [];
    await commonFormApi.resetForm();
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
      error instanceof Error
        ? error.message
        : $t('pages.system.job.messages.previewFailed');
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
    throw new Error($t('pages.system.job.validation.retentionValuePositive'));
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
    throw new Error($t('pages.system.job.validation.cronRequired'));
  }
  if (cronExpr.length > 128) {
    throw new Error($t('pages.system.job.validation.cronTooLong'));
  }

  const fields = splitCronExpressionFields(cronExpr);
  if (fields.length !== 5 && fields.length !== 6) {
    throw new Error($t('pages.system.job.validation.cronInvalidCount'));
  }
  if (fields.length === 6 && fields[0] === '#') {
    throw new Error($t('pages.system.job.validation.cronSecondInvalid'));
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
    throw new Error($t('pages.system.job.validation.timezoneRequired'));
  }
  if (!isValidTimezoneValue(timezone)) {
    throw new Error($t('pages.system.job.validation.timezoneInvalid'));
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
    throw new Error($t('pages.system.job.validation.integerRequired', { label }));
  }
  if (options.allowZero && numericValue === 0) {
    return numericValue;
  }
  if (numericValue < options.min) {
    throw new Error(
      $t('pages.system.job.validation.minValue', {
        label,
        value: options.min,
      }),
    );
  }
  if (
    typeof options.max === 'number' &&
    Number.isFinite(options.max) &&
    numericValue > options.max
  ) {
    throw new Error(
      $t('pages.system.job.validation.maxValue', {
        label,
        value: options.max,
      }),
    );
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
    throw new Error($t('pages.system.job.validation.invalidOption', { label }));
  }
  return text;
}

function validateCommonFormValues(values: Record<string, any>) {
  const groupId = Number(values.groupId);
  if (!Number.isInteger(groupId) || groupId <= 0) {
    throw new Error($t('pages.system.job.validation.groupRequired'));
  }
  const jobName = String(values.name || '').trim();
  if (jobName === '') {
    throw new Error($t('pages.system.job.validation.nameRequired'));
  }
  if (jobName.length > 128) {
    throw new Error($t('pages.system.job.validation.nameTooLong'));
  }

  ensureCronExpressionValue(values.cronExpr);
  ensureTimezoneValue(values.timezone);
  ensureSelectValue(
    $t('pages.system.job.fields.scope'),
    values.scope,
    JOB_SCOPE_OPTIONS.map((item) => String(item.value)),
  );
  ensureSelectValue(
    $t('pages.system.job.fields.concurrency'),
    values.concurrency,
    JOB_CONCURRENCY_OPTIONS.map((item) => String(item.value)),
  );
  ensureSelectValue(
    $t('pages.system.job.fields.status'),
    values.status,
    JOB_FORM_STATUS_OPTIONS,
  );
  ensureIntegerRangeValue(
    $t('pages.system.job.fields.timeoutSeconds'),
    values.timeoutSeconds,
    {
      max: 86400,
      min: 1,
    },
  );
  ensureIntegerRangeValue(
    $t('pages.system.job.fields.maxExecutions'),
    values.maxExecutions ?? 0,
    {
      allowZero: true,
      min: 0,
    },
  );
  if (values.concurrency === 'parallel') {
    ensureIntegerRangeValue($t('pages.system.job.fields.maxConcurrency'), values.maxConcurrency, {
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
        : $t('pages.system.job.messages.formInvalid'),
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
    taskType: 'shell',
    timeoutSeconds: Number(values.timeoutSeconds),
    timezone,
    scope: values.scope,
  };

  const shellPayload = await shellFormRef.value?.validateAndBuild();
  if (!shellPayload) {
    throw new Error($t('pages.system.job.messages.shellConfigInvalid'));
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
  if (isBuiltin.value) {
    modalApi.close();
    return;
  }
  try {
    modalApi.lock(true);
    const payload = await buildPayload();
    if (currentRecord.value) {
      await jobUpdate(currentRecord.value.id, payload);
      message.success($t('pages.common.updateSuccess'));
    } else {
      await jobCreate(payload);
      message.success($t('pages.common.createSuccess'));
    }
    emit('reload');
    modalApi.close();
  } catch (error) {
    const msg = error instanceof Error ? error.message : $t('pages.system.job.messages.saveFailed');
    message.error(msg);
  } finally {
    modalApi.lock(false);
  }
}
</script>

<template>
  <Modal
    :footer="isBuiltin ? false : undefined"
    :title="title"
    data-testid="job-form-modal"
  >
    <div class="space-y-4">
      <div
        v-if="isBuiltin"
        class="py-[5px]"
        data-testid="job-builtin-common-lock-alert"
      >
        <Alert
          :message="$t('pages.system.job.messages.builtinReadonly')"
          show-icon
          type="info"
        />
      </div>

      <CommonForm />

      <div
        v-if="!isBuiltin"
        class="rounded-md border border-border px-4 py-4"
      >
        <div class="mb-3 flex items-center justify-between">
          <div>
            <div class="text-sm font-medium">
              {{ $t('pages.system.job.preview.title') }}
            </div>
            <div class="text-xs text-foreground/60">
              {{ $t('pages.system.job.preview.description') }}
            </div>
          </div>
          <a-button data-testid="job-cron-preview" @click="handlePreviewCron">
            {{ $t('pages.system.job.actions.preview') }}
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
          {{ $t('pages.system.job.preview.empty') }}
        </div>
      </div>

      <div
        v-if="isBuiltin"
        class="rounded-md border border-border px-4 py-4"
        data-testid="job-builtin-detail-card"
      >
        <div class="mb-3 flex items-center justify-between">
          <div>
            <div class="text-sm font-medium">
              {{ $t('pages.system.job.builtin.title') }}
            </div>
            <div class="text-xs text-foreground/60">
              {{ $t('pages.system.job.builtin.description') }}
            </div>
          </div>
        </div>

        <Descriptions
          :column="1"
          bordered
          class="job-builtin-descriptions"
          size="small"
        >
          <DescriptionsItem :label="$t('pages.system.job.fields.source')">
            <Tag :color="getJobSourceColor(currentSourceKind)">
              {{ currentSourceLabel }}
            </Tag>
            <span
              v-if="currentPluginId"
              class="ml-2 text-xs text-foreground/60"
            >
              {{ currentPluginId }}
            </span>
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.job.builtin.taskType')">
            {{ currentRecord?.taskType === 'shell' ? 'Shell' : 'Handler' }}
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.job.builtin.handlerRef')">
            <code>{{ currentRecord?.handlerRef || '-' }}</code>
          </DescriptionsItem>
          <DescriptionsItem
            :label="$t('pages.system.job.builtin.handlerParams')"
          >
            <pre class="job-json-code">{{ readonlyParamsText }}</pre>
          </DescriptionsItem>
        </Descriptions>
      </div>

      <div
        v-else
        class="rounded-md border border-border px-4 py-4"
      >
        <div class="mb-3 flex items-center justify-between">
          <div>
            <div class="text-sm font-medium">
              {{ $t('pages.system.job.shell.title') }}
            </div>
            <div class="text-xs text-foreground/60">
              {{ $t('pages.system.job.shell.description') }}
            </div>
          </div>
          <Alert
            v-if="!shellVisible"
            :message="shellUnavailableReason"
            type="warning"
            show-icon
          />
        </div>

        <JobFormShell ref="shellFormRef" />
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

.job-json-code {
  margin: 0;
  overflow-x: auto;
  padding: 6px 0;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
