import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';
import type { ConfigValueOption, ConfigValueType } from '#/api/system/config/model';

import { $t } from '#/locales';
import { formatTimestamp } from '#/utils/time';

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'name',
    label: $t('pages.system.config.fields.name'),
  },
  {
    component: 'Input',
    fieldName: 'key',
    label: $t('pages.system.config.fields.key'),
  },
  {
    component: 'RangePicker',
    fieldName: 'createTime',
    label: $t('pages.common.createdAt'),
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    title: $t('pages.system.config.fields.name'),
    field: 'name',
    // Long display names read more naturally left-aligned (global grid default is center).
    align: 'left',
  },
  {
    title: $t('pages.system.config.fields.key'),
    field: 'key',
    // Technical keys (e.g. sys.auth.pageTitle) are left-aligned for scanability.
    align: 'left',
  },
  {
    title: $t('pages.system.config.fields.value'),
    field: 'value',
  },
  {
    title: $t('pages.common.remark'),
    field: 'remark',
  },
  {
    title: $t('pages.common.updatedAt'),
    field: 'updatedAt',
    formatter: ({ cellValue }) => formatTimestamp(cellValue),
  },
  {
    field: 'action',
    fixed: 'right',
    slots: { default: 'action' },
    title: $t('pages.common.actions'),
    resizable: false,
    width: 'auto',
  },
];

export function getValueTypeOptions() {
  return [
    { label: $t('pages.system.config.valueTypes.text'), value: 'text' },
    { label: $t('pages.system.config.valueTypes.textarea'), value: 'textarea' },
    { label: $t('pages.system.config.valueTypes.number'), value: 'number' },
    { label: $t('pages.system.config.valueTypes.boolean'), value: 'boolean' },
    { label: $t('pages.system.config.valueTypes.select'), value: 'select' },
    { label: $t('pages.system.config.valueTypes.radio'), value: 'radio' },
    {
      label: $t('pages.system.config.valueTypes.multi_select'),
      value: 'multi_select',
    },
    { label: $t('pages.system.config.valueTypes.richtext'), value: 'richtext' },
  ] as const;
}

export function isEnumValueType(valueType?: string) {
  return valueType === 'select' || valueType === 'radio' || valueType === 'multi_select';
}

/**
 * Modal density tiers for the parameter create/edit dialog.
 * Compact keeps the default mid-size form; spacious expands width and editing
 * surface for long content (richtext / long textarea).
 */
export type ConfigModalDensity = 'compact' | 'spacious';

/**
 * Layout chrome + value-editor sizing policy driven by `valueType`.
 * Keep this as the single source of truth so new long-form types only need a
 * new branch here instead of scattered pixel constants in the modal.
 */
export interface ConfigModalLayout {
  density: ConfigModalDensity;
  /** Merged into Vben Modal `class` (overrides default w-[520px] when set). */
  modalClass: string;
  /** Merged into Vben Modal `contentClass`. */
  contentClass: string;
  /** Whether the modal title bar shows the fullscreen control. */
  fullscreenButton: boolean;
  /**
   * CSS length for richtext editor min/max height (viewport-relative).
   * Empty when the active type is not richtext.
   */
  richtextEditorHeight: string;
  /** Textarea min rows when valueType is textarea. */
  textareaMinRows: number;
}

/** Default compact dialog matches Vben Modal base width (520px). */
const COMPACT_MODAL_LAYOUT: ConfigModalLayout = {
  density: 'compact',
  modalClass: '',
  contentClass: '',
  fullscreenButton: false,
  richtextEditorHeight: '',
  textareaMinRows: 3,
};

/**
 * Resolve modal + editor layout for one parameter value type.
 * Spacious types get a wider dialog, fullscreen entry, and taller editors.
 */
export function resolveConfigModalLayout(
  valueType?: ConfigValueType | string,
): ConfigModalLayout {
  switch (valueType) {
    case 'richtext':
      return {
        density: 'spacious',
        // Wider than the default 520px short form; capped by viewport.
        modalClass: 'w-[min(960px,96vw)] max-w-[96vw]',
        contentClass:
          'config-param-modal-content config-param-modal-content--richtext',
        fullscreenButton: true,
        // Floor 360px, prefer ~half viewport, ceiling 640px for large screens.
        richtextEditorHeight: 'clamp(360px, 52vh, 640px)',
        textareaMinRows: 3,
      };
    case 'textarea':
      return {
        density: 'spacious',
        modalClass: 'w-[min(720px,94vw)] max-w-[94vw]',
        contentClass: 'config-param-modal-content',
        fullscreenButton: true,
        richtextEditorHeight: '',
        textareaMinRows: 8,
      };
    default:
      return { ...COMPACT_MODAL_LAYOUT };
  }
}

/**
 * Parse admin-friendly option lines into structured options.
 * Supported lines:
 * - label=value
 * - label|value
 * - value  (label defaults to value)
 * Also accepts legacy JSON arrays for compatibility.
 */
export function parseOptionsText(raw: unknown): ConfigValueOption[] {
  if (typeof raw !== 'string' || !raw.trim()) {
    return [];
  }
  const trimmed = raw.trim();
  if (trimmed.startsWith('[')) {
    try {
      const parsed = JSON.parse(trimmed) as ConfigValueOption[];
      if (!Array.isArray(parsed)) {
        return [];
      }
      return parsed
        .filter((item) => item && typeof item.value === 'string' && item.value.trim())
        .map((item) => ({
          label: (item.label || item.value).trim(),
          value: item.value.trim(),
        }));
    } catch {
      return [];
    }
  }

  const options: ConfigValueOption[] = [];
  for (const line of trimmed.split('\n')) {
    const text = line.trim();
    if (!text) {
      continue;
    }
    const eq = text.indexOf('=');
    const pipe = text.indexOf('|');
    let sep = -1;
    if (eq >= 0 && pipe >= 0) {
      sep = Math.min(eq, pipe);
    } else if (eq >= 0) {
      sep = eq;
    } else if (pipe >= 0) {
      sep = pipe;
    }
    if (sep < 0) {
      options.push({ label: text, value: text });
      continue;
    }
    const label = text.slice(0, sep).trim();
    const value = text.slice(sep + 1).trim();
    if (!value) {
      continue;
    }
    options.push({ label: label || value, value });
  }
  return options;
}

/** Format structured options as simple one-option-per-line text. */
export function formatOptionsText(options: ConfigValueOption[] = []): string {
  return options
    .filter((item) => item?.value)
    .map((item) => {
      const label = (item.label || item.value).trim();
      const value = item.value.trim();
      if (!label || label === value) {
        return value;
      }
      return `${label}=${value}`;
    })
    .join('\n');
}

export function localizeOptionLabel(
  configKey: string,
  option: ConfigValueOption,
): string {
  const i18nKey = `config.${configKey}.option.${option.value}`;
  const translated = $t(i18nKey);
  if (translated && translated !== i18nKey) {
    return translated;
  }
  return option.label || option.value;
}

export function buildSelectOptions(
  configKey: string,
  options: ConfigValueOption[] = [],
) {
  return options.map((option) => ({
    label: localizeOptionLabel(configKey, option),
    value: option.value,
  }));
}

export function parseMultiSelectValue(value: string | undefined): string[] {
  if (!value) {
    return [];
  }
  return value
    .split(';')
    .map((part) => part.trim())
    .filter(Boolean);
}

export function joinMultiSelectValue(values: string[] | string | undefined): string {
  if (Array.isArray(values)) {
    return values.filter(Boolean).join(';');
  }
  return values || '';
}

export function normalizeFormValue(
  valueType: ConfigValueType | string | undefined,
  value: unknown,
): string {
  if (valueType === 'boolean') {
    if (value === true || value === 'true' || value === 1 || value === '1') {
      return 'true';
    }
    if (value === false || value === 'false' || value === 0 || value === '0') {
      return 'false';
    }
    // Unselected boolean stays empty so required validation can catch it.
    return '';
  }
  if (valueType === 'multi_select') {
    return joinMultiSelectValue(value as string[] | string | undefined);
  }
  if (valueType === 'number') {
    if (value === null || value === undefined || value === '') {
      return '';
    }
    return String(value);
  }
  return value == null ? '' : String(value);
}

/** 新增/编辑表单基础 schema；value 组件由 modal 按类型动态更新 */
export function buildModalSchema(params: {
  isEdit: boolean;
  isBuiltin: boolean;
}): VbenFormSchema[] {
  const lockType = params.isEdit && params.isBuiltin;
  // Built-in name/remark are i18n-owned display metadata; keep them read-only
  // so operators edit the parameter value without rewriting localized labels.
  const lockBuiltinMetadata = params.isEdit && params.isBuiltin;
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('pages.system.config.fields.name'),
      rules: 'required',
      componentProps: {
        disabled: lockBuiltinMetadata,
      },
    },
    {
      component: 'Input',
      fieldName: 'key',
      label: $t('pages.system.config.fields.key'),
      rules: 'required',
      componentProps: {
        disabled: params.isEdit && params.isBuiltin,
      },
    },
    {
      component: 'Select',
      fieldName: 'valueType',
      label: $t('pages.system.config.fields.valueType'),
      defaultValue: 'text',
      rules: 'required',
      componentProps: {
        options: [...getValueTypeOptions()],
        disabled: lockType,
        allowClear: false,
        class: 'w-full',
        // Keep dropdown wider than short selected labels so longer option text stays readable.
        dropdownMatchSelectWidth: false,
        dropdownStyle: { minWidth: '220px' },
      },
    },
    // Options must sit under valueType and above value so enum metadata is set
    // before choosing the parameter value.
    {
      component: 'Textarea',
      fieldName: 'optionsText',
      label: $t('pages.system.config.fields.options'),
      help: $t('pages.system.config.help.optionsSimple'),
      componentProps: {
        disabled: lockType,
        rows: 4,
        placeholder: $t('pages.system.config.placeholders.optionsSimple'),
      },
      dependencies: {
        show: (values) => isEnumValueType(values.valueType),
        triggerFields: ['valueType'],
      },
      formItemClass: 'items-start',
    },
    {
      component: 'Input',
      fieldName: 'value',
      label: $t('pages.system.config.fields.value'),
      rules: 'required',
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: $t('pages.common.remark'),
      componentProps: {
        rows: 3,
        disabled: lockBuiltinMetadata,
      },
      formItemClass: 'items-start',
    },
  ];
}

/**
 * Coerce the current form value into the shape expected by the active value type.
 * multi_select must use a string array; an empty string would render as a blank tag.
 */
export function coerceValueForType(
  valueType: ConfigValueType | string | undefined,
  value: unknown,
): string | string[] | number {
  if (valueType === 'multi_select') {
    if (Array.isArray(value)) {
      return value
        .map((item) => String(item ?? '').trim())
        .filter(Boolean);
    }
    return parseMultiSelectValue(
      value == null || value === '' ? '' : String(value),
    );
  }
  if (Array.isArray(value)) {
    return value
      .map((item) => String(item ?? '').trim())
      .filter(Boolean)
      .join(';');
  }
  if (valueType === 'number') {
    if (value === null || value === undefined || value === '') {
      return '';
    }
    const numeric = Number(value);
    return Number.isFinite(numeric) ? numeric : '';
  }
  if (valueType === 'boolean') {
    if (value === true || value === 'true' || value === 1 || value === '1') {
      return 'true';
    }
    if (value === false || value === 'false' || value === 0 || value === '0') {
      return 'false';
    }
    // New boolean fields default to no selection (neither true nor false).
    return '';
  }
  return value == null ? '' : String(value);
}

export function valueFieldSchema(params: {
  valueType: ConfigValueType | string;
  configKey: string;
  options: ConfigValueOption[];
}): Partial<VbenFormSchema> {
  const { valueType, configKey, options } = params;
  const selectOptions = buildSelectOptions(configKey, options);

  const layout = resolveConfigModalLayout(valueType);

  switch (valueType) {
    case 'textarea':
      return {
        component: 'Textarea',
        componentProps: {
          autoSize: {
            minRows: layout.textareaMinRows,
            maxRows: 20,
          },
          rows: layout.textareaMinRows,
          class: 'w-full',
        },
        formItemClass: 'items-start',
      };
    case 'richtext':
      return {
        component: 'RichText',
        componentProps: {
          // Viewport-relative pane: fixed band with internal scroll (not a short 280px strip).
          height: layout.richtextEditorHeight,
          maxHeight: layout.richtextEditorHeight,
          class: 'w-full',
          placeholder: $t('pages.system.config.placeholders.richtext'),
        },
        formItemClass: 'items-start w-full',
      };
    case 'number':
      return {
        component: 'InputNumber',
        componentProps: {
          class: 'w-full',
        },
      };
    case 'boolean':
      return {
        component: 'RadioGroup',
        // Empty default: do not preselect true/false for new boolean params.
        defaultValue: '',
        componentProps: {
          optionType: 'button',
          buttonStyle: 'solid',
          options: [
            { label: $t('pages.system.config.boolean.true'), value: 'true' },
            { label: $t('pages.system.config.boolean.false'), value: 'false' },
          ],
        },
      };
    case 'select':
      return {
        component: 'Select',
        componentProps: {
          options: selectOptions,
          allowClear: false,
          class: 'w-full',
          dropdownMatchSelectWidth: false,
          dropdownStyle: { minWidth: '220px' },
        },
      };
    case 'radio':
      return {
        component: 'RadioGroup',
        componentProps: {
          options: selectOptions,
        },
        formItemClass: 'items-start',
      };
    case 'multi_select':
      return {
        component: 'Select',
        // Keep controlled multi value as string[] only; never leave "" as model.
        defaultValue: [],
        componentProps: {
          mode: 'multiple',
          options: selectOptions,
          allowClear: true,
          class: 'w-full',
          dropdownMatchSelectWidth: false,
          dropdownStyle: { minWidth: '220px' },
          // Avoid tokenizing free text into empty tags when options are incomplete.
          maxTagCount: 'responsive',
        },
      };
    case 'text':
    default:
      return {
        component: 'Input',
        componentProps: {},
      };
  }
}
