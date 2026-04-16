import type { AxiosRequestConfig } from '@vben/request';

import { requestClient } from '#/api/request';

/**
 * Axios upload progress event type
 */
export type AxiosProgressEvent = AxiosRequestConfig['onUploadProgress'];

/**
 * Upload result returned by the server
 */
export interface UploadResult {
  id: number;
  name: string;
  original: string;
  url: string;
  suffix: string;
  size: number;
}

/**
 * Upload options
 */
export interface UploadOptions {
  onUploadProgress?: AxiosProgressEvent;
  signal?: AbortSignal;
  /** 使用场景标识（必填）：avatar=用户头像 notice_image=通知公告图片 notice_attachment=通知公告附件 other=其他 */
  scene: string;
}

/**
 * Upload a single file via the unified file upload API
 * @param file File to upload
 * @param options Upload options (scene is required for usage tracking)
 */
export function uploadApi(file: Blob | File, options: UploadOptions) {
  const { onUploadProgress, signal, scene } = options;
  const formData = new FormData();
  formData.append('file', file);
  formData.append('scene', scene);
  return requestClient.post<UploadResult>('/file/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress,
    signal,
    timeout: 60_000,
  });
}

/**
 * Upload API function type
 */
export type UploadApiFn = typeof uploadApi;
