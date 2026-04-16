<script setup lang="ts">
import { computed, reactive, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Form, FormItem, Input, message, RadioGroup } from 'ant-design-vue';

import { noticeAdd, noticeInfo, noticeUpdate } from '#/api/system/notice';
import { FileUpload } from '#/components/upload';
import { TiptapEditor } from '#/components/tiptap';

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
const title = computed(() => (isEdit.value ? '编辑通知公告' : '新增通知公告'));

const formRules = reactive({
  title: [{ message: '请输入公告标题', required: true }],
  status: [{ message: '请选择公告状态', required: true }],
  type: [{ message: '请选择公告类型', required: true }],
  content: [{ message: '请输入公告内容', required: true }],
});

const { validate, validateInfos, resetFields } = Form.useForm(
  formData,
  formRules,
);

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
      message.success('更新成功');
    } else {
      await noticeAdd(submitData);
      message.success('创建成功');
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
      <FormItem label="公告标题" v-bind="validateInfos.title">
        <Input
          v-model:value="formData.title"
          placeholder="请输入公告标题"
        />
      </FormItem>
      <div class="grid lg:grid-cols-2 sm:grid-cols-1">
        <FormItem label="公告状态" v-bind="validateInfos.status">
          <RadioGroup
            v-model:value="formData.status"
            button-style="solid"
            option-type="button"
            :options="[
              { label: '草稿', value: 0 },
              { label: '已发布', value: 1 },
            ]"
          />
        </FormItem>
        <FormItem label="公告类型" v-bind="validateInfos.type">
          <RadioGroup
            v-model:value="formData.type"
            button-style="solid"
            option-type="button"
            :options="[
              { label: '通知', value: 1 },
              { label: '公告', value: 2 },
            ]"
          />
        </FormItem>
      </div>
      <FormItem label="公告内容" v-bind="validateInfos.content">
        <TiptapEditor v-model="formData.content" :height="300" scene="notice_image" />
      </FormItem>
      <FormItem label="附件">
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
