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
 * Neutral client direct-access description from the host.
 */
export interface DirectUploadAccess {
  mode: 'presigned_url' | 'form_post' | 'temporary_credentials' | 'proxy' | string;
  operation?: string;
  method?: string;
  url?: string;
  headers?: Record<string, string>;
  formFields?: Record<string, string>;
  accessKeyId?: string;
  secretAccessKey?: string;
  sessionToken?: string;
  expiresAt?: number;
}

/**
 * Direct upload init response.
 */
export interface DirectUploadInitResult {
  instantReuse: boolean;
  uploadSessionId?: string;
  access?: DirectUploadAccess;
  file?: UploadResult;
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
 * Initialize a host-owned direct upload session (or proxy/instant reuse).
 */
export function directUploadInit(body: {
  scene: string;
  fileName: string;
  size: number;
  contentType?: string;
  contentHash?: string;
}) {
  return requestClient.post<DirectUploadInitResult>('/file/direct-upload/init', body);
}

/**
 * Complete a host-owned direct upload session after client transfer.
 */
export function directUploadComplete(body: {
  uploadSessionId: string;
  etag?: string;
}) {
  return requestClient.post<UploadResult>('/file/direct-upload/complete', body);
}

/**
 * Abort a host-owned direct upload session.
 */
export function directUploadAbort(uploadSessionId: string) {
  return requestClient.post('/file/direct-upload/abort', { uploadSessionId });
}

/**
 * Upload a single file via direct cloud access when available, otherwise
 * fall back to host-mediated multipart upload. Callers never branch on vendor IDs.
 */
export async function uploadApi(file: Blob | File, options: UploadOptions) {
  const { onUploadProgress, signal, scene } = options;
  const fileName =
    file instanceof File && file.name ? file.name : 'upload.bin';
  const contentType =
    file instanceof File && file.type ? file.type : 'application/octet-stream';

  let init: DirectUploadInitResult | undefined;
  try {
    init = await directUploadInit({
      scene,
      fileName,
      size: file.size,
      contentType,
    });
  } catch {
    // Init endpoint unavailable or failed closed → classic multipart path.
    return multipartUpload(file, options);
  }

  if (init.instantReuse && init.file) {
    onUploadProgress?.({
      loaded: file.size,
      total: file.size,
      bytes: file.size,
      lengthComputable: true,
    } as any);
    return init.file;
  }

  const mode = (init.access?.mode || 'proxy').toLowerCase();
  if (mode === 'proxy' || !init.uploadSessionId || !init.access) {
    return multipartUpload(file, options);
  }

  try {
    const etag = await executeDirectTransfer(file, init.access, {
      onUploadProgress,
      signal,
    });
    return await directUploadComplete({
      uploadSessionId: init.uploadSessionId,
      etag,
    });
  } catch (error) {
    try {
      await directUploadAbort(init.uploadSessionId);
    } catch {
      // best-effort abort
    }
    throw error;
  }
}

async function multipartUpload(file: Blob | File, options: UploadOptions) {
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
 * Execute a vendor-neutral direct transfer description.
 * Does not fall back to multipart on CORS/network failure.
 */
async function executeDirectTransfer(
  file: Blob | File,
  access: DirectUploadAccess,
  options: {
    onUploadProgress?: AxiosProgressEvent;
    signal?: AbortSignal;
  },
): Promise<string | undefined> {
  const mode = (access.mode || '').toLowerCase();
  if (mode === 'presigned_url') {
    return putPresigned(file, access, options);
  }
  if (mode === 'form_post') {
    await postForm(file, access, options);
    return undefined;
  }
  throw new Error(`Unsupported direct upload mode: ${access.mode || 'unknown'}`);
}

async function putPresigned(
  file: Blob | File,
  access: DirectUploadAccess,
  options: {
    onUploadProgress?: AxiosProgressEvent;
    signal?: AbortSignal;
  },
): Promise<string | undefined> {
  if (!access.url) {
    throw new Error('Direct upload URL is missing');
  }
  const method = (access.method || 'PUT').toUpperCase();
  const headers = new Headers();
  if (access.headers) {
    for (const [key, value] of Object.entries(access.headers)) {
      if (value != null && value !== '') {
        headers.set(key, value);
      }
    }
  }
  // Prefer browser XMLHttpRequest for upload progress events.
  return await new Promise<string | undefined>((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open(method, access.url!, true);
    headers.forEach((value, key) => {
      xhr.setRequestHeader(key, value);
    });
    if (options.signal) {
      if (options.signal.aborted) {
        reject(new DOMException('Aborted', 'AbortError'));
        return;
      }
      options.signal.addEventListener(
        'abort',
        () => {
          xhr.abort();
          reject(new DOMException('Aborted', 'AbortError'));
        },
        { once: true },
      );
    }
    xhr.upload.onprogress = (event) => {
      if (!event.lengthComputable) return;
      options.onUploadProgress?.({
        loaded: event.loaded,
        total: event.total,
        bytes: event.loaded,
        lengthComputable: true,
      } as any);
    };
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(xhr.getResponseHeader('ETag') || undefined);
        return;
      }
      reject(new Error(`Direct upload failed with status ${xhr.status}`));
    };
    xhr.onerror = () => {
      reject(new Error('Direct upload network error (check bucket CORS)'));
    };
    xhr.send(file);
  });
}

async function postForm(
  file: Blob | File,
  access: DirectUploadAccess,
  options: {
    onUploadProgress?: AxiosProgressEvent;
    signal?: AbortSignal;
  },
): Promise<void> {
  if (!access.url) {
    throw new Error('Direct upload form URL is missing');
  }
  const formData = new FormData();
  if (access.formFields) {
    for (const [key, value] of Object.entries(access.formFields)) {
      formData.append(key, value);
    }
  }
  const fileName = file instanceof File && file.name ? file.name : 'file';
  formData.append('file', file, fileName);

  await new Promise<void>((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open('POST', access.url!, true);
    if (options.signal) {
      if (options.signal.aborted) {
        reject(new DOMException('Aborted', 'AbortError'));
        return;
      }
      options.signal.addEventListener(
        'abort',
        () => {
          xhr.abort();
          reject(new DOMException('Aborted', 'AbortError'));
        },
        { once: true },
      );
    }
    xhr.upload.onprogress = (event) => {
      if (!event.lengthComputable) return;
      options.onUploadProgress?.({
        loaded: event.loaded,
        total: event.total,
        bytes: event.loaded,
        lengthComputable: true,
      } as any);
    };
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve();
        return;
      }
      reject(new Error(`Direct form upload failed with status ${xhr.status}`));
    };
    xhr.onerror = () => {
      reject(new Error('Direct form upload network error (check bucket CORS)'));
    };
    xhr.send(formData);
  });
}

/**
 * Upload API function type
 */
export type UploadApiFn = typeof uploadApi;
