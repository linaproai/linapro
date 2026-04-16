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

/** Get file info by IDs (comma-separated) */
export function fileInfoByIds(ids: number | string) {
  return requestClient.get<{ list: FileInfo[] }>(`/file/info/${ids}`);
}

/** Delete files by IDs (comma-separated) */
export function fileRemove(ids: number[]) {
  return requestClient.delete(`/file/${ids.join(',')}`);
}

/** Download file by ID */
export function fileDownloadUrl(id: number) {
  return `/file/download/${id}`;
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
