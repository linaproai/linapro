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

import { $t } from "#/locales";

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
  isEdit.value
    ? $t("plugin.plugin-demo-source.page.modal.editTitle")
    : $t("plugin.plugin-demo-source.page.modal.createTitle"),
);

const formRules = reactive({
  title: [
    {
      message: $t("plugin.plugin-demo-source.page.validation.title"),
      required: true,
    },
  ],
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
      message.success($t("pages.common.updateSuccess"));
    } else {
      await createDemoRecord(payload, selectedFile.value);
      message.success($t("pages.common.createSuccess"));
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
      <FormItem
        :label="$t('plugin.plugin-demo-source.page.fields.title')"
        v-bind="validateInfos.title"
      >
        <Input
          v-model:value="formData.title"
          data-testid="plugin-demo-source-record-title-input"
          maxlength="128"
          :placeholder="$t('plugin.plugin-demo-source.page.placeholders.title')"
        />
      </FormItem>
      <FormItem :label="$t('plugin.plugin-demo-source.page.fields.content')">
        <Input.TextArea
          v-model:value="formData.content"
          :maxlength="1000"
          :rows="5"
          data-testid="plugin-demo-source-record-content-input"
          :placeholder="$t('plugin.plugin-demo-source.page.placeholders.content')"
          show-count
        />
      </FormItem>
      <FormItem :label="$t('plugin.plugin-demo-source.page.fields.attachment')">
        <div>
          <Alert
            data-testid="plugin-demo-source-record-attachment-alert"
            :message="$t('plugin.plugin-demo-source.page.messages.attachmentHint')"
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
              {{
                $t("plugin.plugin-demo-source.page.messages.currentAttachment", {
                  name: existingAttachmentName,
                })
              }}
            </div>
            <div
              v-if="existingAttachmentName && !selectedFile"
              data-testid="plugin-demo-source-record-remove-attachment-option"
              class="py-1"
            >
              <Checkbox v-model:checked="formData.removeAttachment">
                {{ $t("plugin.plugin-demo-source.page.messages.removeAttachment") }}
              </Checkbox>
            </div>
            <UploadDragger
              :before-upload="handleBeforeUpload"
              :file-list="fileList"
              :max-count="1"
              data-testid="plugin-demo-source-record-dragger"
              @remove="handleRemoveFile"
            >
              <p class="ant-upload-text">
                {{ $t("plugin.plugin-demo-source.page.messages.uploadText") }}
              </p>
              <p class="ant-upload-hint">
                {{ $t("plugin.plugin-demo-source.page.messages.uploadHint") }}
              </p>
            </UploadDragger>
          </div>
        </div>
      </FormItem>
    </Form>
  </Modal>
</template>
