import { requestClient } from '#/api/request';

export interface TenantPlugin {
  id: string;
  name: string;
  description: string;
  enabled: number;
  installMode?: 'global' | 'tenant_scoped' | string;
  scopeNature?: 'platform_only' | 'tenant_aware' | string;
  tenantEnabled?: number;
}

export async function tenantPluginList() {
  const res = await requestClient.get<{
    list: TenantPlugin[];
    total: number;
  }>('/tenant/plugins');
  return { items: res.list, total: res.total };
}

export function tenantPluginEnable(pluginId: string) {
  return requestClient.post(`/tenant/plugins/${pluginId}/enable`);
}

export function tenantPluginDisable(pluginId: string) {
  return requestClient.post(`/tenant/plugins/${pluginId}/disable`);
}
