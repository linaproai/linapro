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
 * Neutral upload strategy from direct-upload init / chunked init.
 */
export interface UploadStrategy {
  channel: 'direct' | 'proxy' | string;
  encoding: 'single' | 'multipart' | string;
}

/**
 * Multipart execution plan returned by the host.
 */
export interface UploadMultipartPlan {
  partSize: number;
  minPartSize?: number;
  maxParts?: number;
  maxConcurrency?: number;
}

/**
 * One completed multipart part for complete APIs.
 */
export interface MultipartPartItem {
  partNumber: number;
  etag: string;
}

/**
 * Direct upload init response.
 * Optional fields are conditional: file only on instant reuse, multipart only when
 * strategy.encoding is multipart, access/session absent when instant reuse succeeds.
 */
export interface DirectUploadInitResult {
  /** True when contentHash matched an existing file and no upload is required. */
  instantReuse: boolean;
  /** Session id for complete/abort; empty/absent on instant reuse. */
  uploadSessionId?: string;
  /** Neutral transfer description; mode proxy means host-mediated upload. */
  access?: DirectUploadAccess;
  /** Planned channel/encoding for non-instant uploads. */
  strategy?: UploadStrategy;
  /** Part size/concurrency when strategy.encoding is multipart. */
  multipart?: UploadMultipartPlan;
  /** Existing file metadata when instantReuse is true. */
  file?: UploadResult;
}

/**
 * Chunked upload init response (host-mediated proxy multipart).
 */
export interface ChunkedUploadInitResult {
  uploadSessionId: string;
  /** Planned channel/encoding (typically proxy + multipart). */
  strategy?: UploadStrategy;
  /** Part size/concurrency when encoding is multipart. */
  multipart?: UploadMultipartPlan;
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
  parts?: MultipartPartItem[];
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
 * Issue short-lived access for one direct multipart part.
 */
export function directUploadPartURL(body: {
  uploadSessionId: string;
  partNumber: number;
  size?: number;
}) {
  return requestClient.post<{ access: DirectUploadAccess }>(
    '/file/direct-upload/part-url',
    body,
  );
}

/**
 * Initialize host-mediated chunked upload.
 */
export function chunkedUploadInit(body: {
  scene: string;
  fileName: string;
  size: number;
  contentType?: string;
  contentHash?: string;
}) {
  return requestClient.post<ChunkedUploadInitResult>(
    '/file/upload/chunked/init',
    body,
  );
}

/**
 * Upload one host-mediated chunked part.
 */
export function chunkedUploadPart(body: {
  uploadSessionId: string;
  partNumber: number;
  file: Blob;
  onUploadProgress?: AxiosProgressEvent;
  signal?: AbortSignal;
}) {
  const formData = new FormData();
  formData.append('uploadSessionId', body.uploadSessionId);
  formData.append('partNumber', String(body.partNumber));
  formData.append('file', body.file);
  return requestClient.post<{
    partNumber: number;
    etag?: string;
    receivedBytes?: number;
  }>('/file/upload/chunked/part', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: body.onUploadProgress,
    signal: body.signal,
    timeout: 120_000,
  });
}

/**
 * Complete host-mediated chunked upload.
 */
export function chunkedUploadComplete(body: {
  uploadSessionId: string;
  parts?: MultipartPartItem[];
}) {
  return requestClient.post<UploadResult>('/file/upload/chunked/complete', body);
}

/**
 * Abort host-mediated chunked upload.
 */
export function chunkedUploadAbort(uploadSessionId: string) {
  return requestClient.post('/file/upload/chunked/abort', { uploadSessionId });
}

/**
 * Upload a single file via strategy returned by init:
 * direct/proxy × single/multipart. Callers never branch on vendor IDs.
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
    // Init endpoint unavailable or failed closed → classic whole-file path.
    return wholeFileUpload(file, options);
  }

  if (init.instantReuse && init.file) {
    reportProgress(onUploadProgress, file.size, file.size);
    return init.file;
  }

  const strategy = normalizeStrategy(init);
  if (strategy.channel === 'proxy' && strategy.encoding === 'multipart') {
    return proxyMultipartUpload(file, options, init.multipart);
  }
  if (strategy.channel === 'direct' && strategy.encoding === 'multipart') {
    if (!init.uploadSessionId) {
      return proxyMultipartUpload(file, options, init.multipart);
    }
    return directMultipartUpload(file, init.uploadSessionId, init.multipart, {
      onUploadProgress,
      signal,
    });
  }
  if (strategy.channel === 'direct' && strategy.encoding === 'single') {
    if (!init.uploadSessionId || !init.access) {
      return wholeFileUpload(file, options);
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

  // proxy + single (default)
  return wholeFileUpload(file, options);
}

function normalizeStrategy(init: DirectUploadInitResult): UploadStrategy {
  if (init.strategy?.channel && init.strategy?.encoding) {
    return {
      channel: String(init.strategy.channel).toLowerCase(),
      encoding: String(init.strategy.encoding).toLowerCase(),
    };
  }
  // Backward compatibility: access.mode proxy → proxy single; else direct single.
  const mode = (init.access?.mode || 'proxy').toLowerCase();
  if (mode === 'proxy' || !init.uploadSessionId || !init.access) {
    return { channel: 'proxy', encoding: 'single' };
  }
  return { channel: 'direct', encoding: 'single' };
}

async function wholeFileUpload(file: Blob | File, options: UploadOptions) {
  const { onUploadProgress, signal, scene } = options;
  const formData = new FormData();
  formData.append('file', file);
  formData.append('scene', scene);
  return requestClient.post<UploadResult>('/file/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress,
    signal,
    timeout: 120_000,
  });
}

async function proxyMultipartUpload(
  file: Blob | File,
  options: UploadOptions,
  plan?: UploadMultipartPlan,
) {
  const { onUploadProgress, signal, scene } = options;
  const fileName =
    file instanceof File && file.name ? file.name : 'upload.bin';
  const contentType =
    file instanceof File && file.type ? file.type : 'application/octet-stream';

  const init = await chunkedUploadInit({
    scene,
    fileName,
    size: file.size,
    contentType,
  });
  const partSize = Math.max(plan?.partSize || init.multipart?.partSize || 8 * 1024 * 1024, 5 * 1024 * 1024);
  const slices = sliceFile(file, partSize);
  const parts: MultipartPartItem[] = [];
  let uploaded = 0;

  try {
    for (let i = 0; i < slices.length; i++) {
      const part = slices[i]!;
      const partNumber = i + 1;
      const result = await chunkedUploadPart({
        uploadSessionId: init.uploadSessionId,
        partNumber,
        file: part,
        signal,
        onUploadProgress: (event) => {
          const loaded = uploaded + (event.loaded || 0);
          reportProgress(onUploadProgress, loaded, file.size);
        },
      });
      uploaded += part.size;
      reportProgress(onUploadProgress, uploaded, file.size);
      if (result.etag) {
        parts.push({ partNumber, etag: result.etag });
      }
    }
    return await chunkedUploadComplete({
      uploadSessionId: init.uploadSessionId,
      parts: parts.length > 0 ? parts : undefined,
    });
  } catch (error) {
    try {
      await chunkedUploadAbort(init.uploadSessionId);
    } catch {
      // best-effort abort
    }
    throw error;
  }
}

async function directMultipartUpload(
  file: Blob | File,
  uploadSessionId: string,
  plan: UploadMultipartPlan | undefined,
  options: {
    onUploadProgress?: AxiosProgressEvent;
    signal?: AbortSignal;
  },
) {
  const partSize = Math.max(plan?.partSize || 8 * 1024 * 1024, 5 * 1024 * 1024);
  const maxConcurrency = Math.max(plan?.maxConcurrency || 3, 1);
  const slices = sliceFile(file, partSize);
  const parts: MultipartPartItem[] = new Array(slices.length);
  let uploaded = 0;

  try {
    await runWithConcurrency(slices.length, maxConcurrency, async (index) => {
      const part = slices[index]!;
      const partNumber = index + 1;
      const { access } = await directUploadPartURL({
        uploadSessionId,
        partNumber,
        size: part.size,
      });
      const etag = await executeDirectTransfer(part, access, {
        signal: options.signal,
        onUploadProgress: (event) => {
          // Approximate progress from completed parts only for concurrent uploads.
          reportProgress(
            options.onUploadProgress,
            uploaded + (event.loaded || 0),
            file.size,
          );
        },
      });
      parts[index] = {
        partNumber,
        etag: etag || '',
      };
      uploaded += part.size;
      reportProgress(options.onUploadProgress, uploaded, file.size);
    });

    const completed = parts.filter((p) => p && p.etag);
    if (completed.length !== slices.length) {
      throw new Error('Direct multipart upload missing part etags');
    }
    return await directUploadComplete({
      uploadSessionId,
      parts: completed,
    });
  } catch (error) {
    try {
      await directUploadAbort(uploadSessionId);
    } catch {
      // best-effort abort
    }
    throw error;
  }
}

function sliceFile(file: Blob, partSize: number): Blob[] {
  const slices: Blob[] = [];
  let offset = 0;
  while (offset < file.size) {
    const end = Math.min(offset + partSize, file.size);
    slices.push(file.slice(offset, end));
    offset = end;
  }
  if (slices.length === 0) {
    slices.push(file.slice(0, 0));
  }
  return slices;
}

async function runWithConcurrency(
  total: number,
  concurrency: number,
  worker: (index: number) => Promise<void>,
) {
  let next = 0;
  const runners = Array.from(
    { length: Math.min(concurrency, total) },
    async () => {
      while (next < total) {
        const current = next;
        next += 1;
        await worker(current);
      }
    },
  );
  await Promise.all(runners);
}

function reportProgress(
  onUploadProgress: AxiosProgressEvent | undefined,
  loaded: number,
  total: number,
) {
  onUploadProgress?.({
    loaded,
    total,
    bytes: loaded,
    lengthComputable: true,
  } as any);
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
