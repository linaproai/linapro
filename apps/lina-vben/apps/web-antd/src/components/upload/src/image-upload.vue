<script setup lang="ts">
import type { UploadChangeParam, UploadFile } from 'ant-design-vue';

import type { UploadApiFn, UploadResult } from '#/api';

import { Image, ImagePreviewGroup, Upload } from 'ant-design-vue';

import { uploadApi } from '#/api';

import { defaultImageAcceptExts } from './helper';
import { useImagePreview, useUpload } from './hook';

const props = withDefaults(
  defineProps<{
    accept?: string;
    api?: UploadApiFn;
    disabled?: boolean;
    listType?: 'picture' | 'picture-card' | 'text';
    maxCount?: number;
    maxSize?: number;
    showSuccessMsg?: boolean;
    /** 使用场景标识（必填） */
    scene?: string;
  }>(),
  {
    api: () => uploadApi,
    accept: () => defaultImageAcceptExts.join(','),
    maxCount: 1,
    maxSize: 5,
    disabled: false,
    listType: 'picture-card',
    showSuccessMsg: true,
    scene: 'other',
  },
);

const emit = defineEmits<{
  change: [info: UploadChangeParam];
  success: [file: File, result: UploadResult];
}>();

const ossIdList = defineModel<string | string[]>('value', {
  default: () => [],
});

const { customRequest, acceptStr, handleChange, handleRemove, beforeUpload, innerFileList } =
  useUpload(props, emit, ossIdList, 'image');

const { previewVisible, previewImage, handleCancel, handlePreview } =
  useImagePreview();

function onPreview(file: UploadFile) {
  handlePreview(file);
}
</script>

<template>
  <div>
    <Upload
      v-model:file-list="innerFileList"
      :list-type="listType"
      :accept="accept"
      :disabled="disabled"
      :max-count="maxCount"
      :progress="{ showInfo: true }"
      :before-upload="beforeUpload"
      :custom-request="customRequest"
      @preview="onPreview"
      @change="handleChange"
      @remove="handleRemove"
    >
      <div
        v-if="innerFileList?.length < maxCount && listType === 'picture-card'"
      >
        <span class="icon-[ant-design--plus-outlined]"></span>
        <div class="mt-[8px]">上传</div>
      </div>
      <a-button
        v-if="innerFileList?.length < maxCount && listType !== 'picture-card'"
        :disabled="disabled"
      >
        <span class="icon-[ant-design--upload-outlined] mr-1"></span>
        上传
      </a-button>
    </Upload>
    <div
      class="text-[14px] leading-[1.5] text-black/45 dark:text-white/45"
      :class="{ 'mt-2': listType !== 'picture-card' }"
    >
      支持
      <span class="text-primary mx-1 font-medium">{{ acceptStr }}</span>
      格式，单个文件不超过
      <span class="text-primary mx-1 font-medium">{{ maxSize }}MB</span>
    </div>

    <ImagePreviewGroup
      :preview="{
        visible: previewVisible,
        onVisibleChange: handleCancel,
      }"
    >
      <Image class="hidden" :src="previewImage" />
    </ImagePreviewGroup>
  </div>
</template>
