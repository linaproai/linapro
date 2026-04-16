import type {
  UploadChangeParam,
  UploadFile,
} from 'ant-design-vue';
import type { FileType } from 'ant-design-vue/es/upload/interface';
import type { UploadRequestOption } from 'ant-design-vue/es/vc-upload/interface';

import type { ModelRef } from 'vue';

import type { AxiosProgressEvent, UploadApiFn, UploadResult } from '#/api';
import type { FileInfo } from '#/api/system/file/model';

import { onUnmounted, ref, watch } from 'vue';

import { message, Upload } from 'ant-design-vue';

import { fileInfoByIds } from '#/api/system/file';

/**
 * Image preview hook
 */
export function useImagePreview() {
  const previewVisible = ref(false);
  const previewImage = ref('');

  function handleCancel() {
    previewVisible.value = false;
  }

  async function handlePreview(file: UploadFile) {
    if (!file) return;
    if (!file.url && !file.preview && file.originFileObj) {
      file.preview = await getBase64(file.originFileObj);
    }
    previewImage.value = file.url || file.preview || '';
    previewVisible.value = true;
  }

  function getBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.readAsDataURL(file);
      reader.addEventListener('load', () => resolve(reader.result as string));
      reader.addEventListener('error', (error) => reject(error));
    });
  }

  return { previewVisible, previewImage, handleCancel, handlePreview };
}

/**
 * Common upload hook for both file and image upload components
 */
export function useUpload(
  props: {
    api: UploadApiFn;
    maxSize: number;
    maxCount: number;
    accept?: string;
    showSuccessMsg?: boolean;
    abortOnUnmounted?: boolean;
    /** 使用场景标识（必填） */
    scene?: string;
  },
  emit: {
    (e: 'change', info: UploadChangeParam): void;
    (e: 'success', file: File, result: UploadResult): void;
  },
  bindValue: ModelRef<string | string[]>,
  uploadType: 'file' | 'image',
) {
  const innerFileList = ref<UploadFile[]>([]);
  let isUpload = false;

  const acceptStr = props.accept
    ?.split(',')
    .map((item) => (item.startsWith('.') ? item.slice(1) : item))
    .join(', ');

  function handleChange(info: UploadChangeParam) {
    const { file: currentFile, fileList } = info;

    switch (currentFile.status) {
      case 'done': {
        if (!currentFile.response) return;
        const result = currentFile.response as UploadResult;
        currentFile.url = result.url;
        currentFile.uid = String(result.id);
        currentFile.name = result.original;
        if (uploadType === 'image') {
          currentFile.thumbUrl = result.url;
        }
        isUpload = true;
        if (props.maxCount === 1) {
          bindValue.value = String(result.id);
        } else {
          if (!Array.isArray(bindValue.value)) {
            bindValue.value = [];
          }
          bindValue.value = [...bindValue.value, String(result.id)];
        }
        break;
      }
      case 'error': {
        fileList.splice(fileList.indexOf(currentFile), 1);
        break;
      }
    }
    emit('change', info);
  }

  function handleRemove(currentFile: UploadFile) {
    if (props.maxCount === 1) {
      bindValue.value = '';
    } else {
      (bindValue.value as string[]).splice(
        bindValue.value.indexOf(currentFile.uid),
        1,
      );
    }
    return true;
  }

  function beforeUpload(file: FileType) {
    const isLtMax = file.size / 1024 / 1024 < props.maxSize;
    if (!isLtMax) {
      message.error(`文件大小不能超过${props.maxSize}MB`);
      return Upload.LIST_IGNORE;
    }
    return file;
  }

  const uploadAbort = new AbortController();

  async function customRequest(info: UploadRequestOption<any>) {
    try {
      const progressEvent: AxiosProgressEvent = (e) => {
        const percent = Math.trunc((e.loaded / e.total!) * 100);
        info.onProgress!({ percent });
      };
      const res = await props.api(info.file as File, {
        onUploadProgress: progressEvent,
        signal: uploadAbort.signal,
        scene: props.scene || 'other',
      });
      info.onSuccess!(res);
      if (props.showSuccessMsg !== false) {
        message.success('上传成功');
      }
      emit('success', info.file as File, res);
    } catch (error: any) {
      console.error(error);
      info.onError!(error);
    }
  }

  onUnmounted(() => {
    if (props.abortOnUnmounted !== false) {
      uploadAbort.abort();
    }
  });

  // Watch for external value changes (e.g., form initialization)
  watch(
    () => bindValue.value,
    async (value) => {
      if (!value || value.length === 0) {
        innerFileList.value = [];
        return;
      }
      if (isUpload) {
        isUpload = false;
        return;
      }
      // Fetch file info from server to populate the file list
      try {
        const ids = Array.isArray(value) ? value.join(',') : value;
        const resp = await fileInfoByIds(ids);
        innerFileList.value = (resp.list || []).map(
          (info: FileInfo): UploadFile => ({
            uid: String(info.id),
            name: info.original,
            fileName: info.original,
            url: info.url,
            thumbUrl: uploadType === 'image' ? info.url : undefined,
            status: 'done',
          }),
        );
      } catch {
        innerFileList.value = [];
      }
    },
    { immediate: true },
  );

  return {
    handleChange,
    handleRemove,
    beforeUpload,
    customRequest,
    innerFileList,
    acceptStr,
  };
}
