<script setup lang="ts">
import type { JobRecord } from '#/api/system/job/model';

import { ref } from 'vue';

import { $t } from '#/locales';
import { Alert, Input, InputPassword, Space, Tooltip } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';

interface ShellPayload {
  env: Record<string, string>;
  shellCmd: string;
  workDir: string;
}

interface EnvRow {
  key: string;
  masked: boolean;
  originalValue: string;
  value: string;
}

const [Form, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    formItemClass: 'col-span-1',
    labelWidth: 96,
  },
  schema: [
    {
      component: 'Textarea',
      componentProps: {
        placeholder: $t('pages.system.job.shell.placeholders.command'),
        rows: 6,
      },
      fieldName: 'shellCmd',
      formItemClass: 'col-span-2',
      label: $t('pages.system.job.shell.fields.command'),
      rules: 'required',
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('pages.system.job.shell.placeholders.workDir'),
      },
      fieldName: 'workDir',
      formItemClass: 'col-span-2',
      label: $t('pages.system.job.shell.fields.workDir'),
    },
  ],
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

const envRows = ref<EnvRow[]>([]);

function createEnvRow(): EnvRow {
  return {
    key: '',
    masked: false,
    originalValue: '',
    value: '',
  };
}

function parseEnv(rawEnv?: string) {
  if (!rawEnv) {
    return {};
  }
  try {
    const parsed = JSON.parse(rawEnv);
    return parsed && typeof parsed === 'object' ? parsed : {};
  } catch {
    return {};
  }
}

function ensureEnvRows() {
  if (envRows.value.length === 0) {
    envRows.value = [createEnvRow()];
  }
}

function addEnvRow() {
  envRows.value.push(createEnvRow());
}

function removeEnvRow(index: number) {
  envRows.value.splice(index, 1);
  ensureEnvRows();
}

async function load(record?: JobRecord | null) {
  const env = parseEnv(record?.env) as Record<string, string>;
  const rows = Object.entries(env).map(([key, value]) => ({
    key,
    masked: !!record,
    originalValue: String(value ?? ''),
    value: '',
  }));
  envRows.value = rows.length > 0 ? rows : [createEnvRow()];
  await formApi.setValues({
    shellCmd: record?.shellCmd || '',
    workDir: record?.workDir || '',
  });
}

async function reset() {
  envRows.value = [createEnvRow()];
  await formApi.setValues({
    shellCmd: '',
    workDir: '',
  });
}

function resolveEnvValue(row: EnvRow) {
  if (row.masked && row.value === '') {
    return row.originalValue;
  }
  return row.value;
}

function buildEnvPayload() {
  const env: Record<string, string> = {};
  const duplicatedKeys = new Set<string>();
  const seenKeys = new Set<string>();

  for (const row of envRows.value) {
    const key = row.key.trim();
    const value = resolveEnvValue(row);
    if (key === '' && value === '') {
      continue;
    }
    if (key === '') {
      throw new Error($t('pages.system.job.shell.validation.envKeyRequired'));
    }
    if (seenKeys.has(key)) {
      duplicatedKeys.add(key);
      continue;
    }
    seenKeys.add(key);
    env[key] = value;
  }

  if (duplicatedKeys.size > 0) {
    throw new Error(
      $t('pages.system.job.shell.validation.envKeyDuplicated', {
        keys: Array.from(duplicatedKeys).join(', '),
      }),
    );
  }

  return env;
}

async function validateAndBuild() {
  const { valid } = await formApi.validate();
  if (!valid) {
    throw new Error($t('pages.system.job.messages.shellConfigInvalid'));
  }
  const values = await formApi.getValues<Record<string, any>>();
  return {
    env: buildEnvPayload(),
    shellCmd: values.shellCmd,
    workDir: values.workDir || '',
  } satisfies ShellPayload;
}

defineExpose({
  load,
  reset,
  validateAndBuild,
});
</script>

<template>
  <div class="space-y-4" data-testid="job-form-shell">
    <div class="py-[5px]" data-testid="job-shell-warning-alert">
      <Alert
        :message="$t('pages.system.job.shell.warning')"
        show-icon
        type="warning"
      />
    </div>
    <Form />

    <div class="rounded-md border border-border px-4 py-4">
      <div class="mb-3 flex items-center justify-between">
        <div>
          <div class="text-sm font-medium">
            {{ $t('pages.system.job.shell.fields.env') }}
          </div>
          <div class="text-xs text-foreground/60">
            {{ $t('pages.system.job.shell.envHelp') }}
          </div>
        </div>
        <a-button
          data-testid="job-shell-env-add"
          type="dashed"
          @click="addEnvRow"
        >
          {{ $t('pages.system.job.shell.actions.addEnv') }}
        </a-button>
      </div>

      <div class="space-y-3">
        <div
          v-for="(row, index) in envRows"
          :key="`env-${index}`"
          class="grid grid-cols-[1fr_1fr_auto] gap-3"
        >
          <Input
            v-model:value="row.key"
            :data-testid="`job-shell-env-key-${index}`"
            :placeholder="$t('pages.system.job.shell.placeholders.envKey')"
          />
          <InputPassword
            v-model:value="row.value"
            :data-testid="`job-shell-env-value-${index}`"
            :placeholder="
              row.masked
                ? $t('pages.system.job.shell.placeholders.envValueMasked')
                : $t('pages.system.job.shell.placeholders.envValue')
            "
            visibility-toggle
          />
          <Space>
            <Tooltip
              v-if="row.masked"
              :title="$t('pages.system.job.shell.messages.currentValueMasked')"
            >
              <span class="text-xs text-warning">
                {{ $t('pages.system.job.shell.messages.masked') }}
              </span>
            </Tooltip>
            <a-button
              danger
              type="link"
              :data-testid="`job-shell-env-remove-${index}`"
              @click="removeEnvRow(index)"
            >
              {{ $t('pages.common.delete') }}
            </a-button>
          </Space>
        </div>
      </div>
    </div>
  </div>
</template>
