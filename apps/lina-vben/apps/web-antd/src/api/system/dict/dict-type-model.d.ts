export interface DictType {
  id: number;
  name: string;
  type: string;
  status: number;
  remark: string;
  createdAt: string;
}

export interface DictTypeListParams {
  pageNum?: number;
  pageSize?: number;
  name?: string;
  type?: string;
  ids?: number[];
}

export interface DictTypeListResult {
  items: DictType[];
  total: number;
}
