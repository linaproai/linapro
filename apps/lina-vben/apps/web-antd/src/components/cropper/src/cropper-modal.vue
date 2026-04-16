<script lang="ts" setup>
import type { PropType } from 'vue';

import type { CropendResult, Cropper } from './typing';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Avatar, message, Space, Tooltip, Upload } from 'ant-design-vue';

import { dataURLtoBlob } from '#/utils/file/base64Conver';

import CropperImage from './cropper.vue';

type apiFunParams = { file: Blob; filename: string; name: string };

defineOptions({ name: 'CropperModal' });

const props = defineProps({
  circled: { default: true, type: Boolean },
  size: { default: 0, type: Number },
  src: { default: '', type: String },
  uploadApi: {
    required: true,
    type: Function as PropType<(params: apiFunParams) => Promise<any>>,
  },
});

const emit = defineEmits(['uploadSuccess', 'uploadError', 'register']);

let filename = '';
const src = ref(props.src || '');
const previewSource = ref('');
const cropper = ref<Cropper>();
let scaleX = 1;
let scaleY = 1;

const prefixCls = 'cropper-am';
const [BasicModal, modalApi] = useVbenModal({
  onConfirm: handleOk,
  onOpenChange(isOpen) {
    if (isOpen) {
      modalLoading(true);
    } else {
      previewSource.value = '';
      modalLoading(false);
    }
  },
});

function modalLoading(loading: boolean) {
  modalApi.setState({ confirmLoading: loading, loading });
}

function handleBeforeUpload(file: File) {
  if (props.size > 0 && file.size > 1024 * 1024 * props.size) {
    emit('uploadError', { msg: '图片太大' });
    return false;
  }
  const reader = new FileReader();
  reader.readAsDataURL(file);
  src.value = '';
  previewSource.value = '';
  reader.addEventListener('load', (e) => {
    src.value = (e.target?.result as string) ?? '';
    filename = file.name;
  });
  return false;
}

function handleCropend({ imgBase64 }: CropendResult) {
  previewSource.value = imgBase64;
}

function handleReady(cropperInstance: Cropper) {
  cropper.value = cropperInstance;
  modalLoading(false);
}

function handleReadyError() {
  modalLoading(false);
}

function handlerToolbar(event: string, arg?: number) {
  if (event === 'scaleX') {
    scaleX = arg = scaleX === -1 ? 1 : -1;
  }
  if (event === 'scaleY') {
    scaleY = arg = scaleY === -1 ? 1 : -1;
  }
  (cropper?.value as any)?.[event]?.(arg);
}

async function handleOk() {
  const uploadApi = props.uploadApi;
  if (uploadApi && typeof uploadApi === 'function') {
    if (!previewSource.value) {
      message.warn('未选择图片');
      return;
    }
    const blob = dataURLtoBlob(previewSource.value);
    try {
      modalLoading(true);
      const result = await uploadApi({ file: blob, filename, name: 'file' });
      emit('uploadSuccess', { data: result.url, source: previewSource.value });
      modalApi.close();
    } finally {
      modalLoading(false);
    }
  }
}
</script>
<template>
  <BasicModal
    v-bind="$attrs"
    confirm-text="确认"
    :fullscreen-button="false"
    title="头像裁剪"
    class="w-[800px]"
  >
    <div :class="prefixCls">
      <div :class="`${prefixCls}-left`" class="w-full">
        <div :class="`${prefixCls}-cropper`">
          <CropperImage
            v-if="src"
            :circled="circled"
            :src="src"
            crossorigin="anonymous"
            height="300px"
            @cropend="handleCropend"
            @ready="handleReady"
            @ready-error="handleReadyError"
          />
        </div>

        <div :class="`${prefixCls}-toolbar`">
          <Upload
            :before-upload="handleBeforeUpload"
            :file-list="[]"
            accept="image/*"
          >
            <Tooltip title="选择图片" placement="bottom">
              <a-button size="small" type="primary">
                <template #icon>
                  <div class="flex items-center justify-center">
                    <span class="icon-[ant-design--upload-outlined]"></span>
                  </div>
                </template>
              </a-button>
            </Tooltip>
          </Upload>
          <Space>
            <Tooltip title="重置" placement="bottom">
              <a-button
                :disabled="!src"
                size="small"
                type="primary"
                @click="handlerToolbar('reset')"
              >
                <template #icon>
                  <div class="flex items-center justify-center">
                    <span class="icon-[ant-design--reload-outlined]"></span>
                  </div>
                </template>
              </a-button>
            </Tooltip>
            <Tooltip title="向左旋转" placement="bottom">
              <a-button
                :disabled="!src"
                size="small"
                type="primary"
                @click="handlerToolbar('rotate', -45)"
              >
                <template #icon>
                  <div class="flex items-center justify-center">
                    <span
                      class="icon-[ant-design--rotate-left-outlined]"
                    ></span>
                  </div>
                </template>
              </a-button>
            </Tooltip>
            <Tooltip title="向右旋转" placement="bottom">
              <a-button
                :disabled="!src"
                size="small"
                type="primary"
                @click="handlerToolbar('rotate', 45)"
              >
                <template #icon>
                  <div class="flex items-center justify-center">
                    <span
                      class="icon-[ant-design--rotate-right-outlined]"
                    ></span>
                  </div>
                </template>
              </a-button>
            </Tooltip>
            <Tooltip title="水平翻转" placement="bottom">
              <a-button
                :disabled="!src"
                size="small"
                type="primary"
                @click="handlerToolbar('scaleX')"
              >
                <template #icon>
                  <div class="flex items-center justify-center">
                    <span class="icon-[vaadin--arrows-long-h]"></span>
                  </div>
                </template>
              </a-button>
            </Tooltip>
            <Tooltip title="垂直翻转" placement="bottom">
              <a-button
                :disabled="!src"
                size="small"
                type="primary"
                @click="handlerToolbar('scaleY')"
              >
                <template #icon>
                  <div class="flex items-center justify-center">
                    <span class="icon-[vaadin--arrows-long-v]"></span>
                  </div>
                </template>
              </a-button>
            </Tooltip>
            <Tooltip title="放大" placement="bottom">
              <a-button
                :disabled="!src"
                size="small"
                type="primary"
                @click="handlerToolbar('zoom', 0.1)"
              >
                <template #icon>
                  <div class="flex items-center justify-center">
                    <span class="icon-[ant-design--zoom-in-outlined]"></span>
                  </div>
                </template>
              </a-button>
            </Tooltip>
            <Tooltip title="缩小" placement="bottom">
              <a-button
                :disabled="!src"
                size="small"
                type="primary"
                @click="handlerToolbar('zoom', -0.1)"
              >
                <template #icon>
                  <div class="flex items-center justify-center">
                    <span class="icon-[ant-design--zoom-out-outlined]"></span>
                  </div>
                </template>
              </a-button>
            </Tooltip>
          </Space>
        </div>
      </div>
      <div :class="`${prefixCls}-right`">
        <div :class="`${prefixCls}-preview`">
          <img
            v-if="previewSource"
            alt="预览"
            :src="previewSource"
          />
        </div>
        <template v-if="previewSource">
          <div :class="`${prefixCls}-group`">
            <Avatar :src="previewSource" size="large" />
            <Avatar :size="48" :src="previewSource" />
            <Avatar :size="64" :src="previewSource" />
            <Avatar :size="80" :src="previewSource" />
          </div>
        </template>
      </div>
    </div>
  </BasicModal>
</template>

<style lang="scss">
.cropper-am {
  display: flex;

  &-left,
  &-right {
    height: 340px;
  }

  &-left {
    width: 55%;
  }

  &-right {
    width: 45%;
  }

  &-cropper {
    height: 300px;
    background: #eee;
    background-image:
      linear-gradient(
        45deg,
        rgb(0 0 0 / 25%) 25%,
        transparent 0,
        transparent 75%,
        rgb(0 0 0 / 25%) 0
      ),
      linear-gradient(
        45deg,
        rgb(0 0 0 / 25%) 25%,
        transparent 0,
        transparent 75%,
        rgb(0 0 0 / 25%) 0
      );
    background-position:
      0 0,
      12px 12px;
    background-size: 24px 24px;
  }

  &-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 10px;
  }

  &-preview {
    width: 220px;
    height: 220px;
    margin: 0 auto;
    overflow: hidden;
    border: 1px solid #eee;
    border-radius: 50%;

    img {
      width: 100%;
      height: 100%;
    }
  }

  &-group {
    display: flex;
    align-items: center;
    justify-content: space-around;
    padding-top: 8px;
    margin-top: 8px;
    border-top: 1px solid #eee;
  }
}
</style>
