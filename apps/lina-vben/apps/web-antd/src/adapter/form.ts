import type {
  VbenFormSchema as FormSchema,
  VbenFormProps,
} from '@vben/common-ui';

import type { ComponentType } from './component';

import { setupVbenForm, useVbenForm as useForm, z } from '@vben/common-ui';
import { $t } from '@vben/locales';
import { preferences } from '@vben/preferences';

async function initSetupVbenForm() {
  setupVbenForm<ComponentType>({
    config: {
      // ant design vue组件库默认都是 v-model:value
      baseModelPropName: 'value',

      // 一些组件是 v-model:checked 或者 v-model:fileList
      modelPropNameMap: {
        Checkbox: 'checked',
        Radio: 'checked',
        Switch: 'checked',
        Upload: 'fileList',
      },
    },
    defineRules: {
      // 输入项目必填国际化适配
      required: (value, _params, ctx) => {
        if (value === undefined || value === null || value.length === 0) {
          return $t('ui.formRules.required', [ctx.label]);
        }
        return true;
      },
      // 选择项目必填国际化适配
      selectRequired: (value, _params, ctx) => {
        if (value === undefined || value === null) {
          return $t('ui.formRules.selectRequired', [ctx.label]);
        }
        return true;
      },
    },
  });
}

const ENGLISH_DIALOG_LABEL_WIDTH = 128;

function useEnglishDialogLabelWidth(options: VbenFormProps<ComponentType>) {
  if (
    preferences.app.locale !== 'en-US' ||
    options.showDefaultActions !== false
  ) {
    return options;
  }

  options.commonConfig ??= {};
  options.commonConfig.labelWidth = Math.max(
    options.commonConfig.labelWidth ?? 0,
    ENGLISH_DIALOG_LABEL_WIDTH,
  );
  return options;
}

function useVbenForm(options: VbenFormProps<ComponentType>) {
  return useForm<ComponentType>(useEnglishDialogLabelWidth(options));
}

export { initSetupVbenForm, useVbenForm, z };

export type VbenFormSchema = FormSchema<ComponentType>;
export type { VbenFormProps };

/**
 * 表单 Schema 获取器类型（用于动态获取表单配置）
 */
export type FormSchemaGetter = () =>
  | Promise<VbenFormSchema[]>
  | VbenFormSchema[];

export interface JobHandlerSchemaFieldOption {
  label: string;
  value: boolean | number | string;
}

export interface JobHandlerSchemaField {
  component: ComponentType;
  fieldName: string;
  label: string;
  description?: string;
  defaultValue?: any;
  format?: string;
  options?: JobHandlerSchemaFieldOption[];
  required?: boolean;
}

interface RawJsonSchemaProperty {
  default?: any;
  description?: string;
  enum?: any[];
  format?: string;
  type?: string;
}

interface RawJsonSchemaRoot {
  properties?: Record<string, RawJsonSchemaProperty>;
  required?: string[];
  type?: string;
}

const supportedJsonSchemaKeywords = new Set([
  'default',
  'description',
  'enum',
  'format',
  'properties',
  'required',
  'type',
]);

function validateSchemaKeywords(
  schemaText: string,
  payload: unknown,
  withinProperties = false,
) {
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) {
    throw new Error(
      $t('pages.system.job.handler.messages.schemaInvalidObject', {
        schema: schemaText,
      }),
    );
  }
  for (const [key, value] of Object.entries(
    payload as Record<string, unknown>,
  )) {
    if (withinProperties) {
      validateSchemaKeywords(schemaText, value, false);
      continue;
    }
    if (!supportedJsonSchemaKeywords.has(key)) {
      throw new Error(
        $t('pages.system.job.handler.messages.schemaUnsupportedKeyword', {
          key,
        }),
      );
    }
    if (key === 'properties') {
      validateSchemaKeywords(schemaText, value, true);
      continue;
    }
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      validateSchemaKeywords(schemaText, value);
    }
  }
}

function resolveComponent(property: RawJsonSchemaProperty): ComponentType {
  if (property.enum && property.enum.length > 0) {
    return 'Select';
  }
  switch (property.type) {
    case 'boolean': {
      return 'Switch';
    }
    case 'integer':
    case 'number': {
      return 'InputNumber';
    }
    case 'string': {
      if (property.format === 'textarea') {
        return 'Textarea';
      }
      if (property.format === 'date') {
        return 'DatePicker';
      }
      if (property.format === 'date-time') {
        return 'Input';
      }
      return 'Input';
    }
    default: {
      return 'Input';
    }
  }
}

function normalizeDefaultValue(property: RawJsonSchemaProperty) {
  if (property.default !== undefined) {
    return property.default;
  }
  if (property.type === 'boolean') {
    return false;
  }
  return undefined;
}

function normalizeOptions(
  values: any[] | undefined,
): JobHandlerSchemaFieldOption[] | undefined {
  if (!Array.isArray(values) || values.length === 0) {
    return undefined;
  }
  return values.map((value) => ({
    label: String(value),
    value,
  }));
}

/**
 * 将受限 JSON Schema draft-07 子集映射为任务处理器动态参数字段描述。
 */
export function buildJobHandlerSchemaFields(
  schemaText?: string,
): JobHandlerSchemaField[] {
  if (!schemaText) {
    return [];
  }
  const payload = JSON.parse(schemaText) as RawJsonSchemaRoot;
  validateSchemaKeywords(schemaText, payload);
  if (payload.type !== 'object') {
    throw new Error($t('pages.system.job.handler.messages.schemaRootObject'));
  }
  const properties = payload.properties ?? {};
  const requiredSet = payload.required
    ? new Set(payload.required)
    : new Set<string>();
  return Object.entries(properties).map(([fieldName, property]) => ({
    component: resolveComponent(property),
    fieldName,
    label: property.description || fieldName,
    description: property.description,
    defaultValue: normalizeDefaultValue(property),
    format: property.format,
    options: normalizeOptions(property.enum),
    required: requiredSet.has(fieldName),
  }));
}
