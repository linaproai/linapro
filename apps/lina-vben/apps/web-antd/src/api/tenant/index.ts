import type {
  LoginTenant,
  TenantPlugin,
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
  return requestClient.post<{ accessToken: string }>('/auth/select-tenant', {
    preToken,
    tenantId,
  });
}

export function authSwitchTenant(targetTenantId: number) {
  return requestClient.post<{ accessToken: string }>('/auth/switch-tenant', {
    tenantId: targetTenantId,
  });
}

export function tenantMembershipMe() {
  return requestClient.get<LoginTenant[]>('/tenant/members/me');
}

export async function tenantPluginList() {
  const res = await requestClient.get<{ list: TenantPlugin[]; total: number }>(
    '/tenant/plugins',
  );
  return { items: res.list, total: res.total };
}

export function tenantPluginEnable(pluginId: string) {
  return requestClient.post(`/tenant/plugins/${pluginId}/enable`);
}

export function tenantPluginDisable(pluginId: string) {
  return requestClient.post(`/tenant/plugins/${pluginId}/disable`);
}
