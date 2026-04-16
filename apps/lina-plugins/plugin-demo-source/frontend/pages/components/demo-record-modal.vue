<script setup lang="ts">
import type { UploadFile } from "ant-design-vue";

import { computed, reactive, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";

import {
  Alert,
  Checkbox,
  Form,
  FormItem,
  Input,
  Upload,
  message,
} from "ant-design-vue";

import {
  createDemoRecord,
  getDemoRecord,
  updateDemoRecord,
} from "../demo-record-client";

const UploadDragger = Upload.Dragger;

const emit = defineEmits<{ reload: [] }>();

interface FormData {
  id?: number;
  title: string;
  content: string;
  removeAttachment: boolean;
}

const defaultValues: FormData = {
  id: undefined,
  title: "",
  content: "",
  removeAttachment: false,
};

const formData = ref<FormData>({ ...defaultValues });
const existingAttachmentName = ref("");
const selectedFile = ref<File | null>(null);
const fileList = ref<UploadFile[]>([]);

const isEdit = computed(() => !!formData.value.id);
const modalTitle = computed(() =>
  isEdit.value ? "编辑示例记录" : "新增示例记录",
);

const formRules = reactive({
  title: [{ message: "请输入记录标题", required: true }],
});

const { resetFields, validate, validateInfos } = Form.useForm(
  formData,
  formRules,
);

const [Modal, modalApi] = useVbenModal({
  class: "w-[600px] max-w-[calc(100vw-32px)]",
  onConfirm: handleConfirm,
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) return;
    const data = modalApi.getData<{ id?: number }>();
    resetModalState();
    if (!data?.id) {
      return;
    }

    modalApi.setState({ confirmLoading: true });
    try {
      const record = await getDemoRecord(data.id);
      formData.value = {
        id: record.id,
        title: record.title,
        content: record.content || "",
        removeAttachment: false,
      };
      existingAttachmentName.value = record.attachmentName || "";
    } finally {
      modalApi.setState({ confirmLoading: false });
    }
  },
});

function resetModalState() {
  formData.value = { ...defaultValues };
  existingAttachmentName.value = "";
  selectedFile.value = null;
  fileList.value = [];
  resetFields();
}

function handleBeforeUpload(file: File) {
  selectedFile.value = file;
  fileList.value = [
    {
      name: file.name,
      originFileObj: file,
      size: file.size,
      status: "done",
      uid: `${Date.now()}`,
    },
  ];
  formData.value.removeAttachment = false;
  return false;
}

function handleRemoveFile() {
  selectedFile.value = null;
  fileList.value = [];
}

async function handleConfirm() {
  try {
    modalApi.lock(true);
    await validate();

    const payload = {
      title: formData.value.title.trim(),
      content: formData.value.content.trim(),
      removeAttachment: !selectedFile.value && formData.value.removeAttachment,
    };

    if (isEdit.value && formData.value.id) {
      await updateDemoRecord(formData.value.id, payload, selectedFile.value);
      message.success("更新成功");
    } else {
      await createDemoRecord(payload, selectedFile.value);
      message.success("创建成功");
    }

    emit("reload");
    modalApi.close();
  } finally {
    modalApi.lock(false);
  }
}
</script>

<template>
  <Modal :title="modalTitle">
    <Form
      layout="vertical"
      data-testid="plugin-demo-source-record-form"
    >
      <FormItem label="记录标题" v-bind="validateInfos.title">
        <Input
          v-model:value="formData.title"
          data-testid="plugin-demo-source-record-title-input"
          maxlength="128"
          placeholder="请输入记录标题"
        />
      </FormItem>
      <FormItem label="记录内容">
        <Input.TextArea
          v-model:value="formData.content"
          :maxlength="1000"
          :rows="5"
          data-testid="plugin-demo-source-record-content-input"
          placeholder="请输入记录内容"
          show-count
        />
      </FormItem>
      <FormItem label="附件">
        <div>
          <Alert
            data-testid="plugin-demo-source-record-attachment-alert"
            message="该附件会跟随源码插件示例记录一起保存；卸载插件时若勾选清理存储数据，附件文件也会一并删除。"
            show-icon
            type="info"
          />
          <div
            data-testid="plugin-demo-source-record-upload-section"
            class="mt-5 space-y-4"
          >
            <div
              v-if="existingAttachmentName && !selectedFile"
              data-testid="plugin-demo-source-record-existing-attachment"
              class="rounded border border-dashed border-slate-300 bg-slate-50 px-3 py-2 text-sm text-slate-600"
            >
              当前附件：{{ existingAttachmentName }}
            </div>
            <div
              v-if="existingAttachmentName && !selectedFile"
              data-testid="plugin-demo-source-record-remove-attachment-option"
              class="py-1"
            >
              <Checkbox v-model:checked="formData.removeAttachment">
                提交时移除当前附件
              </Checkbox>
            </div>
            <UploadDragger
              :before-upload="handleBeforeUpload"
              :file-list="fileList"
              :max-count="1"
              data-testid="plugin-demo-source-record-dragger"
              @remove="handleRemoveFile"
            >
              <p class="ant-upload-text">点击或拖拽上传一个附件文件</p>
              <p class="ant-upload-hint">
                不上传则保留当前状态；上传新文件会自动替换旧附件。
              </p>
            </UploadDragger>
          </div>
        </div>
      </FormItem>
    </Form>
  </Modal>
</template>
