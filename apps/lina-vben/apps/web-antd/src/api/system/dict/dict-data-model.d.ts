export interface DictData {
  id: number;
  dictType: string;
  label: string;
  value: string;
  sort: number;
  tagStyle: string;
  cssClass: string;
  status: number;
  isBuiltin: number;
  remark: string;
  sourceTenantId: number;
  isFallback: boolean;
  canEdit: boolean;
  canOverride: boolean;
  overrideMode: 'createTenantOverride' | 'none';
  createdAt: number | null;
  updatedAt: number | null;
}

export interface DictDataListParams {
  pageNum?: number;
  pageSize?: number;
  dictType?: string;
  label?: string;
  ids?: number[];
}

export interface DictDataListResult {
  items: DictData[];
  total: number;
}
