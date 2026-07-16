import type {
  FileDetail,
  FileInfo,
  FileSuffixItem,
  FileUsageSceneItem,
} from './model';

import { requestClient } from '#/api/request';

/** File list query params */
export interface FileListParams {
  pageNum?: number;
  pageSize?: number;
  name?: string;
  original?: string;
  suffix?: string;
  scene?: string;
  beginTime?: string;
  endTime?: string;
}

/** File list result */
export interface FileListResult {
  list: FileInfo[];
  total: number;
}

/** Get file list with pagination */
export async function fileList(params?: FileListParams) {
  const res = await requestClient.get<FileListResult>('/file', { params });
  return { items: res.list, total: res.total };
}

/** Get file info by IDs (query array ids[]) */
export function fileInfoByIds(ids: number | number[]) {
  const list = Array.isArray(ids) ? ids : [ids];
  return requestClient.get<{ list: FileInfo[] }>('/file/info', {
    params: { ids: list },
  });
}

/** Delete files by IDs (query array ids[]) */
export function fileRemove(ids: number[]) {
  return requestClient.delete('/file', {
    params: { ids },
  });
}

/** Download file by ID */
export function fileDownloadUrl(id: number) {
  return `/file/download/${id}`;
}

/** Direct download access description for one file (proxy or presigned). */
export interface DirectDownloadResult {
  access?: {
    mode: string;
    operation?: string;
    method?: string;
    url?: string;
    headers?: Record<string, string>;
    expiresAt?: number;
  };
  proxyUrl?: string;
}

/** Issue client direct download access or proxy indication for one file. */
export function fileDirectDownload(id: number) {
  return requestClient.get<DirectDownloadResult>(`/file/${id}/direct-download`);
}

/** Get file usage scene options */
export async function fileUsageScenes() {
  const res = await requestClient.get<{ list: FileUsageSceneItem[] }>(
    '/file/scenes',
  );
  return res.list;
}

/** Get file detail with usage scenes */
export async function fileDetail(id: number) {
  return await requestClient.get<FileDetail>(`/file/detail/${id}`);
}

/** Get file suffix options from database */
export async function fileSuffixes() {
  const res = await requestClient.get<{ list: FileSuffixItem[] }>(
    '/file/suffixes',
  );
  return res.list;
}
