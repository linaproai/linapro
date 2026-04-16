import type { ConfigListParams, SysConfig } from './model';

import { requestClient } from '#/api/request';

/** 参数设置列表 */
export async function configList(params?: ConfigListParams) {
  const res = await requestClient.get<{ list: SysConfig[]; total: number }>(
    '/config',
    { params },
  );
  // VXE-Grid proxy expects { items, total } format
  return { items: res.list, total: res.total };
}

/** 新增参数设置 */
export function configAdd(data: Partial<SysConfig>) {
  return requestClient.post('/config', data);
}

/** 更新参数设置 */
export function configUpdate(id: number, data: Partial<SysConfig>) {
  return requestClient.put(`/config/${id}`, data);
}

/** 删除参数设置 */
export function configDelete(id: number) {
  return requestClient.delete(`/config/${id}`);
}

/** 获取参数设置详情 */
export function configInfo(id: number) {
  return requestClient.get<SysConfig>(`/config/${id}`);
}

/** 导出参数设置 */
export function configExport(params?: ConfigListParams) {
  return requestClient.download<Blob>('/config/export', { params });
}

/** 导入参数设置 */
export function configImport(file: File, updateSupport?: boolean) {
  const formData = new FormData();
  formData.append('file', file);
  if (updateSupport) {
    formData.append('updateSupport', '1');
  }
  return requestClient.post<{
    success: number;
    fail: number;
    failList: Array<{ row: number; reason: string }>;
  }>('/config/import', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}

/** 下载参数设置导入模板 */
export function configImportTemplate() {
  return requestClient.download<Blob>('/config/import-template');
}
