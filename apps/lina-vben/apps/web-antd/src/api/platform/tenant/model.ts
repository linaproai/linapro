export type TenantStatus = 'active' | 'deleted' | 'suspended';

export interface PlatformTenant {
  id: number;
  code: string;
  name: string;
  status: TenantStatus;
  remark?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface PlatformTenantListParams {
  pageNum?: number;
  pageSize?: number;
  code?: string;
  name?: string;
  status?: TenantStatus | '';
}

export interface PlatformTenantMember {
  id: number;
  userId: number;
  tenantId: number;
  username: string;
  nickname?: string;
  status: number;
}

export interface PlatformTenantMemberListParams {
  pageNum?: number;
  pageSize?: number;
  status?: number;
  tenantId?: number;
  userId?: number;
}

export interface TenantImpersonationResult {
  accessToken?: string;
  tenant?: PlatformTenant;
  token?: string;
}
