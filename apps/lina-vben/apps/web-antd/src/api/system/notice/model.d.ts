export interface Notice {
  id: number;
  title: string;
  type: number;
  content: string;
  fileIds: string;
  status: number;
  remark: string;
  createdBy: number;
  createdByName: string;
  updatedBy: number;
  createdAt: string;
  updatedAt: string;
}

export interface NoticeListParams {
  pageNum?: number;
  pageSize?: number;
  title?: string;
  type?: number;
  createdBy?: string;
}
