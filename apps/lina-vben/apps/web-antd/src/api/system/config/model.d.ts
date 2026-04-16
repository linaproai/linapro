export interface SysConfig {
  id: number;
  name: string;
  key: string;
  value: string;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface ConfigListParams {
  pageNum?: number;
  pageSize?: number;
  name?: string;
  key?: string;
  beginTime?: string;
  endTime?: string;
}
