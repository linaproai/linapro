export type ConfigValueType =
  | 'text'
  | 'textarea'
  | 'number'
  | 'boolean'
  | 'select'
  | 'radio'
  | 'multi_select'
  | 'richtext';

export interface ConfigValueOption {
  label: string;
  value: string;
}

export interface SysConfig {
  id: number;
  name: string;
  key: string;
  value: string;
  valueType: ConfigValueType;
  options: ConfigValueOption[];
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
