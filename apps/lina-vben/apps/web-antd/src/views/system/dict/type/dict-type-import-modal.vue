<script setup lang="ts">
import type { UploadFile } from 'ant-design-vue/es/upload/interface';

import { h, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { $t } from '@vben/locales';

import { Modal, Switch, Upload } from 'ant-design-vue';

import { dictImport, dictImportTemplate } from '#/api/system/dict/dict-type';
import { downloadBlob } from '#/utils/download';

const emit = defineEmits<{ reload: [] }>();

const UploadDragger = Upload.Dragger;

const [BasicModal, modalApi] = useVbenModal({
  onCancel: handleCancel,
  onConfirm: handleSubmit,
});

const fileList = ref<UploadFile[]>([]);
const updateSupport = ref(false);

async function handleSubmit() {
  try {
    modalApi.setState({ loading: true });
    if (fileList.value.length !== 1) {
      handleCancel();
      return;
    }
    const file = fileList.value[0]!.originFileObj as File;
    const result = await dictImport(file, updateSupport.value);
    const res = result as any;
    let modal = Modal.success;
    if (res.typeFail > 0 || res.dataFail > 0) {
      modal = Modal.error;
    }
    emit('reload');
    handleCancel();
    const content =
      res.typeFail > 0 || res.dataFail > 0
        ? `${$t('pages.system.dict.type.import.summary', {
            dataFail: res.dataFail,
            dataSuccess: res.dataSuccess,
            typeFail: res.typeFail,
            typeSuccess: res.typeSuccess,
          })}\n${res.failList
            .slice(0, 5)
            .map((item: any) =>
              $t('pages.system.dict.type.import.rowReason', {
                reason: item.reason,
                row: item.row,
                sheet: item.sheet,
              }),
            )
            .join('\n')}${res.failList.length > 5 ? '\n...' : ''}`
        : $t('pages.system.dict.type.import.success', {
            dataCount: res.dataSuccess,
            typeCount: res.typeSuccess,
          });
    modal({
      content: h('div', {
        class: 'max-h-[260px] overflow-y-auto whitespace-pre-wrap',
        innerHTML: content,
      }),
      title: $t('pages.system.dict.type.import.resultTitle'),
    });
  } catch (error) {
    console.warn(error);
    modalApi.close();
  } finally {
    modalApi.setState({ loading: false });
  }
}

function handleCancel() {
  modalApi.close();
  fileList.value = [];
  updateSupport.value = false;
}

async function handleDownloadTemplate() {
  try {
    const data = await dictImportTemplate();
    downloadBlob(data, 'dict-management-import-template.xlsx');
  } catch {
    Modal.error({ title: $t('pages.system.dict.type.import.downloadTemplateFailed') });
  }
}
</script>

<template>
  <BasicModal
    :close-on-click-modal="false"
    :fullscreen-button="false"
    :title="$t('pages.system.dict.type.import.title')"
  >
    <UploadDragger
      v-model:file-list="fileList"
      :before-upload="() => false"
      :max-count="1"
      :show-upload-list="true"
      accept="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet, application/vnd.ms-excel"
    >
      <p class="ant-upload-drag-icon flex items-center justify-center">
        <IconifyIcon class="text-primary text-5xl" icon="ant-design:inbox-outlined" />
      </p>
      <p class="ant-upload-text">{{ $t('pages.system.dict.type.import.uploadHint') }}</p>
    </UploadDragger>
    <div class="mt-2 flex flex-col gap-2">
      <div class="flex items-center gap-2">
        <span>{{ $t('pages.system.dict.type.import.fileHint') }}</span>
        <a-button type="link" @click="handleDownloadTemplate">
          <div class="flex items-center gap-[4px]">
            <IconifyIcon icon="ant-design:file-excel-outlined" />
            <span>{{ $t('pages.system.dict.type.import.downloadTemplate') }}</span>
          </div>
        </a-button>
      </div>
      <div class="flex items-center gap-2">
        <span :class="{ 'text-red-500': updateSupport }">
          {{ $t('pages.system.dict.type.import.overwrite') }}
        </span>
        <Switch v-model:checked="updateSupport" />
      </div>
    </div>
  </BasicModal>
</template>
