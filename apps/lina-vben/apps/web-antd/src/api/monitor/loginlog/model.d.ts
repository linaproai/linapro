export interface LoginLog {
  id: number;
  userName: string;
  status: number;
  ip: string;
  browser: string;
  os: string;
  msg: string;
  loginTime: string;
}

export interface LoginLogListParams {
  pageNum?: number;
  pageSize?: number;
  userName?: string;
  ip?: string;
  status?: number;
  beginTime?: string;
  endTime?: string;
  orderBy?: string;
  orderDirection?: string;
}

export interface LoginLogListResult {
  items: LoginLog[];
  total: number;
}
