export interface OnlineUser {
  tokenId: string;
  username: string;
  deptName: string;
  ip: string;
  browser: string;
  os: string;
  loginTime: string;
}

export interface OnlineListResult {
  items: OnlineUser[];
  total: number;
}

export interface OnlineListParams {
  pageNum?: number;
  pageSize?: number;
  username?: string;
  ip?: string;
}
