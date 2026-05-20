export interface SysConfig {
  id: number;
  name: string;
  key: string;
  value: string;
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

export interface ConfigListParams {
  pageNum?: number;
  pageSize?: number;
  name?: string;
  key?: string;
  beginTime?: string;
  endTime?: string;
}
