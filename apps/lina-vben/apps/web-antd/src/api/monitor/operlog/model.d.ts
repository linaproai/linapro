export interface OperLog {
  id: number;
  title: string;
  operSummary: string;
  operType: number;
  method: string;
  requestMethod: string;
  operName: string;
  operUrl: string;
  operIp: string;
  operParam: string;
  jsonResult: string;
  status: number;
  errorMsg: string;
  costTime: number;
  operTime: string;
}

export interface OperLogListParams {
  pageNum?: number;
  pageSize?: number;
  title?: string;
  operName?: string;
  operType?: number;
  status?: number;
  beginTime?: string;
  endTime?: string;
  orderBy?: string;
  orderDirection?: string;
}

export interface OperLogListResult {
  items: OperLog[];
  total: number;
}
