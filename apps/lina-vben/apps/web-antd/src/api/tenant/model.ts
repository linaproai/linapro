import type { PlatformTenant } from '#/api/platform/tenant/model';

export interface LoginTenant {
  id: number;
  code: string;
  name: string;
  status?: string;
}

export interface TenantAwareLoginResult {
  accessToken?: string;
  preToken?: string;
  refreshToken?: string;
  tenants?: LoginTenant[];
}

export interface TenantTokenResult {
  accessToken: string;
  refreshToken?: string;
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
