import type { DictData, DictDataListParams } from './dict-data-model';

import { requestClient } from '#/api/request';

/** 字典数据列表 */
export async function dictDataList(params?: DictDataListParams) {
  const res = await requestClient.get<{ list: DictData[]; total: number }>(
    '/dict/data',
    { params },
  );
  // VXE-Grid proxy expects { items, total } format
  return { items: res.list, total: res.total };
}

/** 新增字典数据 */
export function dictDataAdd(data: Partial<DictData>) {
  return requestClient.post('/dict/data', data);
}

/** 更新字典数据 */
export function dictDataUpdate(id: number, data: Partial<DictData>) {
  return requestClient.put(`/dict/data/${id}`, data);
}

/** 删除字典数据 */
export function dictDataDelete(id: number) {
  return requestClient.delete(`/dict/data/${id}`);
}

/** 获取字典数据详情 */
export function dictDataInfo(id: number) {
  return requestClient.get<DictData>(`/dict/data/${id}`);
}

/** 导出字典数据 */
export function dictDataExport(params?: DictDataListParams) {
  return requestClient.download<Blob>('/dict/data/export', { params });
}

/** 根据字典类型获取字典数据列表 */
export async function dictDataByType(dictType: string) {
  const res = await requestClient.get<{ list: DictData[] }>(
    `/dict/data/type/${dictType}`,
  );
  return res.list;
}

/** 导入字典数据 */
export function dictDataImport(file: File, updateSupport?: boolean) {
  const formData = new FormData();
  formData.append('file', file);
  if (updateSupport) {
    formData.append('updateSupport', '1');
  }
  return requestClient.post<{
    success: number;
    fail: number;
    failList: Array<{ row: number; reason: string }>;
  }>('/dict/data/import', formData);
}

/** 下载字典数据导入模板 */
export function dictDataImportTemplate() {
  return requestClient.download<Blob>('/dict/data/import-template');
}
