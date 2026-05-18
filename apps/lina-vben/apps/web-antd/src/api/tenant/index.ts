import type {
  LoginTenant,
  TenantTokenResult,
} from './model';

import { requestClient } from '#/api/request';

export async function authLoginTenants(userId: number) {
  const res = await requestClient.get<{ list: LoginTenant[] }>(
    '/auth/login-tenants',
    { params: { userId } },
  );
  return res.list;
}

export function authSelectTenant(preToken: string, tenantId: number) {
  return requestClient.post<TenantTokenResult>('/auth/select-tenant', {
    preToken,
    tenantId,
  });
}

export function authSwitchTenant(targetTenantId: number) {
  return requestClient.post<TenantTokenResult>('/auth/switch-tenant', {
    tenantId: targetTenantId,
  });
}
