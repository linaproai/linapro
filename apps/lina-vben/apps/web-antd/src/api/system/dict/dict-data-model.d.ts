export interface DictData {
  id: number;
  dictType: string;
  label: string;
  value: string;
  sort: number;
  tagStyle: string;
  cssClass: string;
  status: number;
  remark: string;
  createdAt: string;
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
