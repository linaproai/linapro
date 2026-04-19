<script setup lang="ts">
import type { JobRecord } from '#/api/system/job/model';

import { ref } from 'vue';

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
        placeholder: '请输入 Shell 脚本内容',
        rows: 6,
      },
      fieldName: 'shellCmd',
      formItemClass: 'col-span-2',
      label: '执行命令',
      rules: 'required',
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: '请输入工作目录，不填则使用宿主当前目录',
      },
      fieldName: 'workDir',
      formItemClass: 'col-span-2',
      label: '工作目录',
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
      throw new Error('环境变量键不能为空');
    }
    if (seenKeys.has(key)) {
      duplicatedKeys.add(key);
      continue;
    }
    seenKeys.add(key);
    env[key] = value;
  }

  if (duplicatedKeys.size > 0) {
    throw new Error(`环境变量键重复: ${Array.from(duplicatedKeys).join(', ')}`);
  }

  return env;
}

async function validateAndBuild() {
  const { valid } = await formApi.validate();
  if (!valid) {
    throw new Error('请完善 Shell 配置');
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
        message="Shell 任务会在宿主节点直接执行，请严格控制命令内容和环境变量。"
        show-icon
        type="warning"
      />
    </div>
    <Form />

    <div class="rounded-md border border-border px-4 py-4">
      <div class="mb-3 flex items-center justify-between">
        <div>
          <div class="text-sm font-medium">环境变量</div>
          <div class="text-xs text-foreground/60">
            编辑已有值时默认遮罩，留空表示保持原值不变。
          </div>
        </div>
        <a-button
          data-testid="job-shell-env-add"
          type="dashed"
          @click="addEnvRow"
        >
          新增变量
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
            placeholder="变量名，如 APP_ENV"
          />
          <InputPassword
            v-model:value="row.value"
            :data-testid="`job-shell-env-value-${index}`"
            :placeholder="
              row.masked ? '已隐藏，留空表示保持原值' : '变量值'
            "
            visibility-toggle
          />
          <Space>
            <Tooltip v-if="row.masked" title="当前值已遮罩显示">
              <span class="text-xs text-warning">已隐藏</span>
            </Tooltip>
            <a-button
              danger
              type="link"
              :data-testid="`job-shell-env-remove-${index}`"
              @click="removeEnvRow(index)"
            >
              删除
            </a-button>
          </Space>
        </div>
      </div>
    </div>
  </div>
</template>
