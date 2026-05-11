import { requestClient } from '#/api/request';

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

export interface PlatformTenantPayload {
  code: string;
  name: string;
  remark?: string;
}

export interface TenantImpersonationResult {
  accessToken?: string;
  tenant?: PlatformTenant;
  token?: string;
}

export async function platformTenantList(params?: PlatformTenantListParams) {
  const res = await requestClient.get<{
    list: PlatformTenant[];
    total: number;
  }>('/platform/tenants', { params });
  return { items: res.list, total: res.total };
}

export function platformTenantCreate(payload: PlatformTenantPayload) {
  return requestClient.post<PlatformTenant>('/platform/tenants', payload);
}

export function platformTenantUpdate(
  id: number,
  payload: Omit<PlatformTenantPayload, 'code'>,
) {
  return requestClient.put<PlatformTenant>(`/platform/tenants/${id}`, payload);
}

export function platformTenantChangeStatus(id: number, status: TenantStatus) {
  return requestClient.put<PlatformTenant>(`/platform/tenants/${id}/status`, {
    status,
  });
}

export function platformTenantDelete(id: number) {
  return requestClient.delete(`/platform/tenants/${id}`);
}
