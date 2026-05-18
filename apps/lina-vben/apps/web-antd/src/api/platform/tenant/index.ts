import type {
  PlatformTenant,
  PlatformTenantListParams,
  TenantImpersonationResult,
} from './model';

import { requestClient } from '#/api/request';

export async function platformTenantList(params?: PlatformTenantListParams) {
  const res = await requestClient.get<{
    list: PlatformTenant[];
    total: number;
  }>('/platform/tenants', { params });
  return { items: res.list, total: res.total };
}

export function platformTenantImpersonate(id: number) {
  return requestClient.post<TenantImpersonationResult>(
    `/platform/tenants/${id}/impersonate`,
  );
}

export function platformTenantEndImpersonate(id: number) {
  return requestClient.post(`/platform/tenants/${id}/end-impersonate`);
}
