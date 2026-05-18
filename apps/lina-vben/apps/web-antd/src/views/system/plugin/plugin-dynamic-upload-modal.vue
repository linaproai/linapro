<script setup lang="ts">
import type { UploadFile } from 'ant-design-vue/es/upload/interface';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';

import { Alert, Modal, Switch, Upload } from 'ant-design-vue';

import { pluginDynamicUpload } from '#/api/system/plugin';
import { $t } from '#/locales';

const emit = defineEmits<{ reload: [] }>();

const UploadDragger = Upload.Dragger;

const [BasicModal, modalApi] = useVbenModal({
  onCancel: handleCancel,
  onConfirm: handleSubmit,
});

const fileList = ref<UploadFile[]>([]);
const overwriteSupport = ref(false);
const successMessage = ref('');

async function handleSubmit() {
  if (successMessage.value) {
    emit('reload');
    handleCancel();
    return;
  }
  try {
    modalApi.setState({ loading: true });
    if (fileList.value.length !== 1) {
      Modal.warning({ title: $t('pages.system.plugin.upload.selectFile') });
      return;
    }

    const uploadItem = fileList.value[0]!;
    const rawFile = uploadItem.originFileObj as Blob | File;
    // Ant Design Upload may expose a Blob-like object here. Rebuilding a
    // concrete File preserves the original `.wasm` filename so the backend can
    // validate the extension and store the artifact under the expected name.
    const file =
      rawFile instanceof File
        ? rawFile
        : new File([rawFile], uploadItem.name || 'dynamic-plugin.wasm', {
            type: rawFile.type || 'application/wasm',
          });
    await pluginDynamicUpload(file, overwriteSupport.value);
    successMessage.value = $t('pages.system.plugin.upload.success');
  } catch (error) {
    console.warn(error);
  } finally {
    modalApi.setState({ loading: false });
  }
}

function handleCancel() {
  modalApi.close();
  fileList.value = [];
  overwriteSupport.value = false;
  successMessage.value = '';
}
</script>

<template>
  <BasicModal
    :close-on-click-modal="false"
    :close-on-press-escape="!successMessage"
    :closable="!successMessage"
    :fullscreen-button="false"
    :confirm-text="
      successMessage
        ? $t('pages.system.plugin.upload.acknowledge')
        : $t('pages.common.confirm')
    "
    :show-cancel-button="!successMessage"
    :title="$t('pages.system.plugin.upload.title')"
  >
    <template v-if="!successMessage">
      <UploadDragger
        v-model:file-list="fileList"
        :before-upload="() => false"
        :max-count="1"
        :show-upload-list="true"
        accept=".wasm,application/wasm"
        data-testid="plugin-dynamic-upload-dragger"
      >
        <p class="ant-upload-drag-icon flex items-center justify-center">
          <IconifyIcon
            class="text-primary text-5xl"
            icon="ant-design:inbox-outlined"
          />
        </p>
        <p class="ant-upload-text">
          {{ $t('pages.system.plugin.upload.dragText') }}
        </p>
        <p class="ant-upload-hint">
          {{ $t('pages.system.plugin.upload.dragHint') }}
        </p>
      </UploadDragger>
      <div class="mt-2 flex items-center gap-2">
        <span :class="{ 'text-red-500': overwriteSupport }">
          {{ $t('pages.system.plugin.upload.overwriteHint') }}
        </span>
        <div class="flex items-center gap-2">
          <Switch
            v-model:checked="overwriteSupport"
            data-testid="plugin-dynamic-overwrite-switch"
          />
        </div>
      </div>
    </template>
    <Alert
      v-else
      :message="successMessage"
      data-testid="plugin-dynamic-upload-success"
      show-icon
      type="success"
    />
  </BasicModal>
</template>
