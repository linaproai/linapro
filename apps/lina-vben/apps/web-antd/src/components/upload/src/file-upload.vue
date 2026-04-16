<script setup lang="ts">
import type { UploadChangeParam } from 'ant-design-vue';

import type { UploadApiFn, UploadResult } from '#/api';

import { computed } from 'vue';

import { Upload } from 'ant-design-vue';

import { uploadApi } from '#/api';

import { defaultFileAcceptExts } from './helper';
import { useUpload } from './hook';

const props = withDefaults(
  defineProps<{
    accept?: string;
    api?: UploadApiFn;
    disabled?: boolean;
    enableDragUpload?: boolean;
    maxCount?: number;
    maxSize?: number;
    showSuccessMsg?: boolean;
    /** 使用场景标识（必填） */
    scene?: string;
  }>(),
  {
    api: () => uploadApi,
    accept: () => defaultFileAcceptExts.join(','),
    maxCount: 1,
    maxSize: 5,
    disabled: false,
    enableDragUpload: false,
    showSuccessMsg: true,
    scene: 'other',
  },
);

const emit = defineEmits<{
  change: [info: UploadChangeParam];
  success: [file: File, result: UploadResult];
}>();

const CurrentUploadComponent = computed(() => {
  return props.enableDragUpload ? Upload.Dragger : Upload;
});

const ossIdList = defineModel<string | string[]>('value', {
  default: () => [],
});

const { customRequest, acceptStr, handleChange, handleRemove, beforeUpload, innerFileList } =
  useUpload(props, emit, ossIdList, 'file');
</script>

<template>
  <div>
    <CurrentUploadComponent
      v-model:file-list="innerFileList"
      :accept="accept"
      :disabled="disabled"
      :max-count="maxCount"
      :progress="{ showInfo: true }"
      :before-upload="beforeUpload"
      :custom-request="customRequest"
      @change="handleChange"
      @remove="handleRemove"
    >
      <div v-if="!enableDragUpload && innerFileList?.length < maxCount">
        <a-button :disabled="disabled">
          <span class="icon-[ant-design--upload-outlined] mr-1"></span>
          上传
        </a-button>
      </div>
      <div v-if="enableDragUpload">
        <p class="ant-upload-drag-icon">
          <span class="icon-[ant-design--inbox-outlined] text-4xl text-primary"></span>
        </p>
        <p class="ant-upload-text">点击或拖拽文件到此区域上传</p>
      </div>
    </CurrentUploadComponent>
    <div
      v-if="true"
      class="mt-2 text-[14px] leading-[1.5] text-black/45 dark:text-white/45"
    >
      支持
      <span class="text-primary mx-1 font-medium">{{ acceptStr }}</span>
      格式，单个文件不超过
      <span class="text-primary mx-1 font-medium">{{ maxSize }}MB</span>
    </div>
  </div>
</template>
