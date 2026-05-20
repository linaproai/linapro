export interface JobGroupRecord {
  id: number;
  code: string;
  name: string;
  remark: string;
  sortOrder: number;
  isDefault: number;
  createdAt: number | null;
  updatedAt: number | null;
  jobCount: number;
}

export interface JobGroupListParams {
  pageNum?: number;
  pageSize?: number;
  code?: string;
  name?: string;
  orderBy?: string;
  orderDirection?: string;
}

export interface JobGroupPayload {
  code: string;
  name: string;
  remark?: string;
  sortOrder?: number;
}
