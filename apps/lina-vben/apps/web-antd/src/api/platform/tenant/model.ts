export type TenantStatus = 'active' | 'deleted' | 'suspended';

export interface PlatformTenant {
  id: number;
  code: string;
  name: string;
  status: TenantStatus;
  remark?: string;
  createdAt?: number | null;
  updatedAt?: number | null;
}

export interface PlatformTenantListParams {
  pageNum?: number;
  pageSize?: number;
  code?: string;
  name?: string;
  status?: TenantStatus | '';
}

export interface TenantImpersonationResult {
  accessToken?: string;
  tenant?: PlatformTenant;
  token?: string;
}
