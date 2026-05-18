import type { DictType, DictTypeListParams } from './dict-type-model';

import { requestClient } from '#/api/request';

/** 字典类型列表 */
export async function dictTypeList(params?: DictTypeListParams) {
  const res = await requestClient.get<{ list: DictType[]; total: number }>(
    '/dict/type',
    { params },
  );
  // VXE-Grid proxy expects { items, total } format
  return { items: res.list, total: res.total };
}

/** 新增字典类型 */
export function dictTypeAdd(data: Partial<DictType>) {
  return requestClient.post('/dict/type', data);
}

/** 更新字典类型 */
export function dictTypeUpdate(id: number, data: Partial<DictType>) {
  return requestClient.put(`/dict/type/${id}`, data);
}

/** 删除字典类型 */
export function dictTypeDelete(id: number) {
  return requestClient.delete(`/dict/type/${id}`);
}

/** 获取字典类型详情 */
export function dictTypeInfo(id: number) {
  return requestClient.get<DictType>(`/dict/type/${id}`);
}

/** 导出字典类型 */
export function dictTypeExport(params?: DictTypeListParams) {
  return requestClient.download<Blob>('/dict/type/export', { params });
}

/** 获取字典类型选项列表 */
export async function dictTypeOptions() {
  const res = await requestClient.get<{ list: DictType[] }>(
    '/dict/type/options',
  );
  return res.list;
}

/** 导入字典类型 */
export function dictTypeImport(file: File, updateSupport?: boolean) {
  const formData = new FormData();
  formData.append('file', file);
  if (updateSupport) {
    formData.append('updateSupport', '1');
  }
  return requestClient.post<{
    success: number;
    fail: number;
    failList: Array<{ row: number; reason: string }>;
  }>('/dict/type/import', formData);
}

/** 下载字典类型导入模板 */
export function dictTypeImportTemplate() {
  return requestClient.download<Blob>('/dict/type/import-template');
}

/** 导出字典管理数据（合并导出：类型+数据） */
export function dictExport(params?: DictTypeListParams) {
  return requestClient.download<Blob>('/dict/export', { params });
}

/** 导入字典管理数据（合并导入：类型+数据） */
export function dictImport(file: File, updateSupport?: boolean) {
  const formData = new FormData();
  formData.append('file', file);
  if (updateSupport) {
    formData.append('updateSupport', '1');
  }
  return requestClient.post<{
    typeSuccess: number;
    typeFail: number;
    dataSuccess: number;
    dataFail: number;
    failList: Array<{ sheet: string; row: number; reason: string }>;
  }>('/dict/import', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}

/** 下载字典管理导入模板（合并模板：类型+数据） */
export function dictImportTemplate() {
  return requestClient.download<Blob>('/dict/import-template');
}
