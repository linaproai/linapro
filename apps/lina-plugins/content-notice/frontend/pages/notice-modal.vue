<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Form, FormItem, Input, message, RadioGroup } from 'ant-design-vue';

import { noticeAdd, noticeInfo, noticeUpdate } from './notice-client';
import { FileUpload } from '#/components/upload';
import { $t } from '#/locales';
import { TiptapEditor } from '#/components/tiptap';
import { useDictStore } from '#/store/dict';

const emit = defineEmits<{ reload: [] }>();

interface FormData {
  id?: number;
  title: string;
  status: number;
  type: number;
  content: string;
  fileIds: string | string[];
}

const defaultValues: FormData = {
  id: undefined,
  title: '',
  status: 0,
  type: 1,
  content: '',
  fileIds: [],
};

const isEdit = computed(() => !!formData.value.id);
const formData = ref<FormData>({ ...defaultValues });
const title = computed(() =>
  isEdit.value
    ? $t('plugin.content-notice.drawer.editTitle')
    : $t('plugin.content-notice.drawer.createTitle'),
);

const formRules = reactive({
  title: [{ message: $t('plugin.content-notice.validation.title'), required: true }],
  status: [{ message: $t('plugin.content-notice.validation.status'), required: true }],
  type: [{ message: $t('plugin.content-notice.validation.type'), required: true }],
  content: [{ message: $t('plugin.content-notice.validation.content'), required: true }],
});

const { validate, validateInfos, resetFields } = Form.useForm(
  formData,
  formRules,
);

const dictStore = useDictStore();
const noticeTypeOptions = ref<{ label: string; value: number }[]>([]);

onMounted(async () => {
  const dicts = await dictStore.getDictOptionsAsync('sys_notice_type');
  noticeTypeOptions.value = dicts.map((item: any) => ({
    label: item.label,
    value: Number(item.value),
  }));
});

const [Modal, modalApi] = useVbenModal({
  class: 'w-[800px]',
  fullscreenButton: true,
  onConfirm: handleConfirm,
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) return;
    const data = modalApi.getData();
    if (data?.id) {
      modalApi.setState({ confirmLoading: true });
      try {
        const record = await noticeInfo(data.id);
        const fileIds = record.fileIds
          ? record.fileIds.split(',').filter(Boolean)
          : [];
        formData.value = {
          id: record.id,
          title: record.title,
          type: record.type,
          status: record.status,
          content: record.content || '',
          fileIds,
        };
      } finally {
        modalApi.setState({ confirmLoading: false });
      }
    } else {
      formData.value = { ...defaultValues, fileIds: [] };
      resetFields();
    }
  },
});

async function handleConfirm() {
  try {
    modalApi.lock(true);
    await validate();

    const { id, fileIds, ...values } = formData.value;
    const submitData = {
      ...values,
      fileIds: Array.isArray(fileIds) ? fileIds.join(',') : fileIds,
    };
    if (isEdit.value && id) {
      await noticeUpdate(id, submitData);
      message.success($t('pages.common.updateSuccess'));
    } else {
      await noticeAdd(submitData);
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
</script>

<template>
  <Modal :title="title">
    <Form layout="vertical">
      <FormItem :label="$t('plugin.content-notice.fields.title')" v-bind="validateInfos.title">
        <Input
          v-model:value="formData.title"
          :placeholder="$t('plugin.content-notice.placeholders.title')"
        />
      </FormItem>
      <div class="grid lg:grid-cols-2 sm:grid-cols-1">
        <FormItem :label="$t('plugin.content-notice.fields.status')" v-bind="validateInfos.status">
          <RadioGroup
            v-model:value="formData.status"
            button-style="solid"
            option-type="button"
            :options="[
              { label: $t('plugin.content-notice.status.draft'), value: 0 },
              { label: $t('plugin.content-notice.status.published'), value: 1 },
            ]"
          />
        </FormItem>
        <FormItem :label="$t('plugin.content-notice.fields.type')" v-bind="validateInfos.type">
          <RadioGroup
            v-model:value="formData.type"
            button-style="solid"
            option-type="button"
            :options="noticeTypeOptions"
          />
        </FormItem>
      </div>
      <FormItem :label="$t('plugin.content-notice.fields.content')" v-bind="validateInfos.content">
        <TiptapEditor v-model="formData.content" :height="300" scene="notice_image" />
      </FormItem>
      <FormItem :label="$t('plugin.content-notice.fields.attachments')">
        <FileUpload
          v-model:value="formData.fileIds"
          :max-count="5"
          :max-size="10"
          :enable-drag-upload="true"
          scene="notice_attachment"
        />
      </FormItem>
    </Form>
  </Modal>
</template>
