<script setup lang="ts">
import type { UploadFile } from 'ant-design-vue/es/upload/interface';

import { h, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';

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
        ? `字典类型: 成功 ${res.typeSuccess} 条，失败 ${res.typeFail} 条\n字典数据: 成功 ${res.dataSuccess} 条，失败 ${res.dataFail} 条\n${res.failList
            .slice(0, 5)
            .map((item: any) => `[${item.sheet}] 第${item.row}行: ${item.reason}`)
            .join('\n')}${res.failList.length > 5 ? '\n...' : ''}`
        : `成功导入 ${res.typeSuccess} 条字典类型，${res.dataSuccess} 条字典数据`;
    modal({
      content: h('div', {
        class: 'max-h-[260px] overflow-y-auto whitespace-pre-wrap',
        innerHTML: content,
      }),
      title: '提示',
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
    downloadBlob(data, '字典管理导入模板.xlsx');
  } catch {
    Modal.error({ title: '下载模板失败' });
  }
}
</script>

<template>
  <BasicModal
    :close-on-click-modal="false"
    :fullscreen-button="false"
    title="字典管理导入"
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
      <p class="ant-upload-text">点击或者拖拽到此处上传文件</p>
    </UploadDragger>
    <div class="mt-2 flex flex-col gap-2">
      <div class="flex items-center gap-2">
        <span>允许导入xlsx, xls文件</span>
        <a-button type="link" @click="handleDownloadTemplate">
          <div class="flex items-center gap-[4px]">
            <IconifyIcon icon="ant-design:file-excel-outlined" />
            <span>下载模板</span>
          </div>
        </a-button>
      </div>
      <div class="flex items-center gap-2">
        <span :class="{ 'text-red-500': updateSupport }">
          是否更新/覆盖已存在的字典类型和字典数据
        </span>
        <Switch v-model:checked="updateSupport" />
      </div>
    </div>
  </BasicModal>
</template>