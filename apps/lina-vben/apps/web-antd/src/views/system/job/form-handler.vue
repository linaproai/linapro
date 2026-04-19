<script setup lang="ts">
import type {
  JobHandlerOption,
  JobRecord,
} from '#/api/system/job/model';
import type { VbenFormSchema } from '#/adapter/form';

import { nextTick, ref, watch } from 'vue';

import { Alert, Empty, Spin } from 'ant-design-vue';

import {
  buildJobHandlerSchemaFields,
  useVbenForm,
} from '#/adapter/form';
import { jobHandlerDetail } from '#/api/system/jobHandler';

interface HandlerPayload {
  handlerRef: string;
  params: Record<string, any>;
}

interface DynamicSchemaField {
  component: VbenFormSchema['component'];
  defaultValue?: any;
  description?: string;
  fieldName: string;
  format?: string;
  label: string;
  options?: Array<{ label: string; value: boolean | number | string }>;
  required?: boolean;
}

const props = withDefaults(
  defineProps<{
    builtin?: boolean;
    handlerOptions: JobHandlerOption[];
    loading?: boolean;
  }>(),
  {
    builtin: false,
    loading: false,
  },
);

const currentSchemaText = ref('');
const dynamicFields = ref<DynamicSchemaField[]>([]);
const schemaError = ref('');
const schemaLoading = ref(false);

const [Form, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    formItemClass: 'col-span-1',
    labelWidth: 96,
  },
  schema: buildFormSchema([], props.handlerOptions, props.builtin, props.loading),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

watch(
  () => [props.builtin, props.handlerOptions, props.loading] as const,
  async () => {
    await rebuildSchema(undefined);
  },
  { deep: true },
);

function buildFormSchema(
  fields: DynamicSchemaField[],
  options: JobHandlerOption[],
  builtin: boolean,
  loading: boolean,
): VbenFormSchema[] {
  const baseSchema: VbenFormSchema[] = [
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
        disabled: builtin,
        loading,
        onChange: (value: string) => {
          void handleHandlerRefChange(value);
        },
        options: options.map((item) => ({
          label: `${item.displayName} (${item.ref})`,
          value: item.ref,
        })),
        placeholder: '请选择任务处理器',
        showSearch: true,
      },
      fieldName: 'handlerRef',
      help: '内置任务的处理器引用不可修改',
      label: '任务处理器',
      rules: 'required',
    },
  ];

  const dynamicSchema = fields.map((field) => {
    const schema: VbenFormSchema = {
      component: field.component,
      defaultValue: field.defaultValue,
      fieldName: `params__${field.fieldName}`,
      help: field.description,
      label: field.label,
    };
    if (field.required) {
      schema.rules = 'required';
    }

    switch (field.component) {
      case 'InputNumber': {
        schema.componentProps = {
          disabled: builtin,
          precision: field.format === 'integer' ? 0 : undefined,
          style: { width: '100%' },
        };
        break;
      }
      case 'Select': {
        schema.componentProps = {
          allowClear: true,
          disabled: builtin,
          options: field.options,
          placeholder: `请选择${field.label}`,
        };
        break;
      }
      case 'Switch': {
        schema.componentProps = {
          checkedChildren: '是',
          disabled: builtin,
          unCheckedChildren: '否',
        };
        break;
      }
      case 'Textarea': {
        schema.componentProps = {
          disabled: builtin,
          placeholder: `请输入${field.label}`,
          rows: 4,
        };
        schema.formItemClass = 'col-span-2';
        break;
      }
      case 'DatePicker': {
        schema.componentProps = {
          disabled: builtin,
          style: { width: '100%' },
          valueFormat: field.format === 'date' ? 'YYYY-MM-DD' : 'YYYY-MM-DD HH:mm:ss',
        };
        break;
      }
      default: {
        schema.componentProps = {
          disabled: builtin,
          placeholder: `请输入${field.label}`,
        };
      }
    }
    return schema;
  });

  return [...baseSchema, ...dynamicSchema];
}

async function getCurrentValues() {
  try {
    return await formApi.getValues<Record<string, any>>();
  } catch {
    return {};
  }
}

async function rebuildSchema(paramValues?: Record<string, any>) {
  const currentValues = await getCurrentValues();
  formApi.setState({
    schema: buildFormSchema(
      dynamicFields.value,
      props.handlerOptions,
      props.builtin,
      props.loading,
    ),
  });
  await nextTick();

  const mergedValues: Record<string, any> = {
    ...currentValues,
  };
  for (const field of dynamicFields.value) {
    const dynamicKey = `params__${field.fieldName}`;
    mergedValues[dynamicKey] =
      paramValues?.[field.fieldName] ??
      currentValues[dynamicKey] ??
      field.defaultValue;
  }
  await formApi.setValues(mergedValues);
}

function parseJobParams(rawParams?: string) {
  if (!rawParams) {
    return {};
  }
  try {
    const parsed = JSON.parse(rawParams);
    return parsed && typeof parsed === 'object' ? parsed : {};
  } catch {
    return {};
  }
}

async function applyHandlerSchema(
  handlerRef: string,
  paramValues?: Record<string, any>,
) {
  schemaLoading.value = true;
  schemaError.value = '';
  try {
    currentSchemaText.value = '';
    dynamicFields.value = [];
    if (handlerRef) {
      const detail = await jobHandlerDetail(handlerRef);
      currentSchemaText.value = detail.paramsSchema || '';
      dynamicFields.value = buildJobHandlerSchemaFields(
        currentSchemaText.value,
      ) as DynamicSchemaField[];
    }
    await rebuildSchema(paramValues);
  } catch (error) {
    schemaError.value =
      error instanceof Error ? error.message : '加载处理器参数定义失败';
    dynamicFields.value = [];
    await rebuildSchema({});
  } finally {
    schemaLoading.value = false;
  }
}

async function handleHandlerRefChange(handlerRef: string) {
  await applyHandlerSchema(handlerRef, {});
}

async function load(record?: JobRecord | null) {
  const params = parseJobParams(record?.params);
  await applyHandlerSchema(record?.handlerRef || '', params);
  await formApi.setValues({
    handlerRef: record?.handlerRef || '',
  });
}

async function reset() {
  currentSchemaText.value = '';
  dynamicFields.value = [];
  schemaError.value = '';
  await rebuildSchema({});
  await formApi.setValues({
    handlerRef: '',
  });
}

async function validateAndBuild() {
  const { valid } = await formApi.validate();
  if (!valid) {
    throw new Error('请完善 Handler 配置');
  }
  const values = await formApi.getValues<Record<string, any>>();
  const params: Record<string, any> = {};
  for (const field of dynamicFields.value) {
    const dynamicKey = `params__${field.fieldName}`;
    const value = values[dynamicKey];
    if (value === '' || value === undefined || value === null) {
      if (field.component === 'Switch' && value === false) {
        params[field.fieldName] = false;
      }
      continue;
    }
    params[field.fieldName] = value;
  }
  return {
    handlerRef: values.handlerRef,
    params,
  } satisfies HandlerPayload;
}

defineExpose({
  load,
  reset,
  validateAndBuild,
});
</script>

<template>
  <div class="space-y-4" data-testid="job-form-handler">
    <Alert
      v-if="props.builtin"
      message="系统内置任务的处理器引用和参数已锁定，只允许调整公共调度字段。"
      type="info"
      show-icon
    />
    <Spin :spinning="props.loading || schemaLoading">
      <Form />
    </Spin>
    <Alert
      v-if="schemaError"
      :message="schemaError"
      show-icon
      type="error"
    />
    <Empty
      v-else-if="!props.loading && props.handlerOptions.length === 0"
      description="当前没有可用的任务处理器"
      image="simple"
    />
  </div>
</template>
