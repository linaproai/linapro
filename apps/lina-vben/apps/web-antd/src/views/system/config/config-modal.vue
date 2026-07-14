<script setup lang="ts">
import type {
  ConfigValueOption,
  ConfigValueType,
  SysConfig,
} from '#/api/system/config/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { configAdd, configInfo, configUpdate } from '#/api/system/config';
import { syncPublicFrontendSettings } from '#/runtime/public-frontend';

import {
  buildModalSchema,
  coerceValueForType,
  formatOptionsText,
  isEnumValueType,
  normalizeFormValue,
  parseOptionsText,
  resolveConfigModalLayout,
  valueFieldSchema,
} from './data';

const emit = defineEmits<{ reload: [] }>();

const isEdit = ref(false);
const recordId = ref<number>(0);
const isBuiltin = ref(false);
const configKey = ref('');
const loadedOptions = ref<ConfigValueOption[]>([]);
const activeValueType = ref<ConfigValueType | string>('text');
const title = computed(() =>
  isEdit.value
    ? $t('pages.system.config.drawer.editTitle')
    : $t('pages.system.config.drawer.createTitle'),
);

const [BasicForm, formApi] = useVbenForm({
  schema: buildModalSchema({ isEdit: false, isBuiltin: false }),
  showDefaultActions: false,
  // Switching valueType remounts the value control; keep model-update validation off
  // so required rules do not flash while the user is only picking a type.
  commonConfig: {
    formFieldProps: {
      validateOnModelUpdate: false,
      validateOnChange: false,
    },
  },
  handleValuesChange: async (values, fieldsChanged) => {
    if (
      fieldsChanged.includes('valueType') ||
      fieldsChanged.includes('optionsText')
    ) {
      await refreshValueFieldSchema(values);
    }
    if (fieldsChanged.includes('valueType')) {
      // multi_select expects string[]; empty string becomes a blank tag in Ant Select.
      // shouldValidate=false: type switch must not surface required errors for empty value.
      const nextValue = coerceValueForType(values.valueType, values.value);
      await formApi.setFieldValue('value', nextValue, false);
      // updateSchema / component swap can leave a stale error on the value field.
      await formApi.resetValidate();
    }
  },
});

const [BasicModal, modalApi] = useVbenModal({
  fullscreenButton: false,
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  onOpenChange: async (isOpen) => {
    if (!isOpen) {
      return;
    }
    modalApi.setState({ loading: true });

    const { id } = modalApi.getData() as { id?: number };
    isEdit.value = !!id;
    recordId.value = id || 0;
    isBuiltin.value = false;
    configKey.value = '';
    loadedOptions.value = [];
    activeValueType.value = 'text';
    applyModalLayout('text');

    await formApi.setState({
      schema: buildModalSchema({
        isEdit: isEdit.value,
        isBuiltin: false,
      }),
    });
    await formApi.resetForm();

    if (isEdit.value && id) {
      const record = await configInfo(id);
      await applyRecord(record);
    } else {
      await formApi.setValues({
        valueType: 'text',
        value: '',
        optionsText: '',
      });
      await refreshValueFieldSchema({ valueType: 'text', optionsText: '' });
    }

    modalApi.setState({ loading: false });
  },
});

/**
 * Apply density policy for the active value type (width, fullscreen, content class).
 * Always clear fullscreen when the type changes so compact forms do not stay expanded.
 */
function applyModalLayout(valueType?: ConfigValueType | string) {
  const layout = resolveConfigModalLayout(valueType);
  activeValueType.value = valueType || 'text';
  modalApi.setState({
    class: layout.modalClass,
    contentClass: layout.contentClass,
    fullscreenButton: layout.fullscreenButton,
    fullscreen: false,
  });
}

async function applyRecord(record: SysConfig) {
  isBuiltin.value = record.isBuiltin === 1;
  configKey.value = record.key;
  loadedOptions.value = record.options || [];
  await formApi.setState({
    schema: buildModalSchema({
      isEdit: true,
      isBuiltin: isBuiltin.value,
    }),
  });

  const valueType = (record.valueType || 'text') as ConfigValueType;
  const optionsText = formatOptionsText(record.options || []);

  const formValue = coerceValueForType(valueType, record.value);

  await formApi.setValues({
    name: record.name,
    key: record.key,
    valueType,
    value: formValue,
    optionsText,
    remark: record.remark,
  });
  await refreshValueFieldSchema({
    valueType,
    optionsText,
    key: record.key,
  });
}

async function refreshValueFieldSchema(values: Record<string, any>) {
  const valueType = (values.valueType || 'text') as ConfigValueType;
  applyModalLayout(valueType);

  const parsedFromText = parseOptionsText(values.optionsText);
  const options =
    parsedFromText.length > 0 ? parsedFromText : loadedOptions.value;
  const schema = valueFieldSchema({
    valueType,
    configKey: values.key || configKey.value || '',
    options,
  });
  // multi_select is optional until the user picks choices; required would force
  // an empty array into invalid states while switching types.
  const valueRules = valueType === 'multi_select' ? undefined : 'required';
  await formApi.updateSchema([
    {
      fieldName: 'value',
      ...schema,
      label: $t('pages.system.config.fields.value'),
      rules: valueRules,
    },
  ]);
  // Schema/component swap must not leave a red required message until submit.
  await formApi.resetValidate();
}

async function handleConfirm() {
  try {
    modalApi.lock(true);
    const { valid } = await formApi.validate();
    if (!valid) {
      return;
    }
    const data = await formApi.getValues();
    const valueType = (data.valueType || 'text') as ConfigValueType;
    if (isEnumValueType(valueType)) {
      const options = parseOptionsText(data.optionsText);
      if (options.length === 0 && !isBuiltin.value) {
        message.error($t('pages.system.config.messages.optionsRequired'));
        return;
      }
      data.options = options;
    } else {
      data.options = [];
    }
    data.value = normalizeFormValue(valueType, data.value);
    delete data.optionsText;

    if (isEdit.value) {
      await configUpdate(recordId.value, data);
      await syncPublicFrontendSettings();
      message.success($t('pages.common.updateSuccess'));
    } else {
      await configAdd(data);
      await syncPublicFrontendSettings();
      message.success($t('pages.common.createSuccess'));
    }
    emit('reload');
    modalApi.close();
  } catch (error) {
    console.error(error);
  } finally {
    modalApi.lock(false);
  }
}

async function handleClosed() {
  await formApi.resetForm();
  isBuiltin.value = false;
  loadedOptions.value = [];
  configKey.value = '';
  activeValueType.value = 'text';
  // Restore compact chrome so the next open does not inherit spacious classes.
  applyModalLayout('text');
}
</script>

<template>
  <BasicModal :title="title">
    <div :data-testid="`config-modal-value-type-${activeValueType}`">
      <BasicForm />
    </div>
  </BasicModal>
</template>

<style scoped>
/*
 * Spacious content types: keep the value editor full width inside the form grid
 * so richtext/textarea use the expanded modal surface.
 */
:deep(.config-param-modal-content--richtext .tiptap-editor) {
  width: 100%;
}
</style>
