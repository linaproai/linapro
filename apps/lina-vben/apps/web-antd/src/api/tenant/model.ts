import type { PlatformTenant } from '#/api/platform/tenant/model';
import type { SystemPlugin } from '#/api/system/plugin/model';

export interface LoginTenant {
  id: number;
  code: string;
  name: string;
  status?: string;
}

export interface TenantAwareLoginResult {
  accessToken?: string;
  preToken?: string;
  tenants?: LoginTenant[];
}

export interface TenantPlugin extends SystemPlugin {
  installMode?: 'global' | 'tenant_scoped' | string;
  scopeNature?: 'platform_only' | 'tenant_aware' | string;
  tenantEnabled?: number;
}

export interface TenantState {
  enabled: boolean;
  currentTenant: LoginTenant | null;
  tenants: LoginTenant[];
  impersonation: {
    actingUserId?: number;
    active: boolean;
    tenant?: PlatformTenant;
  };
}
