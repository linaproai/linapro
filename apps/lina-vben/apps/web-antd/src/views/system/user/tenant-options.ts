import type { PlatformTenant } from '#/api/platform/tenant/model';
import type { LoginTenant } from '#/api/tenant/model';

import { platformTenantList } from '#/api/platform/tenant';
import { authLoginTenants } from '#/api/tenant';
import {
  canListLoginTenants,
  canListPlatformTenants,
} from '#/store/tenant-permissions';

export interface UserTenantOption {
  label: string;
  value: number;
}

interface UserTenantOptionSource {
  accessCodes?: string[];
  currentTenant?: LoginTenant | null;
  isPlatform: boolean;
  tenants: LoginTenant[];
  userId?: number;
}

function activeLoginTenants(items: LoginTenant[]) {
  return items.filter(
    (item) =>
      item.id > 0 && item.status !== 'suspended' && item.status !== 'deleted',
  );
}

function toOptions(items: Array<LoginTenant | PlatformTenant>) {
  return items.map((item) => ({
    label: item.name,
    value: item.id,
  }));
}

export async function loadUserTenantOptions(
  source: UserTenantOptionSource,
): Promise<UserTenantOption[]> {
  if (!source.isPlatform) {
    return source.currentTenant && source.currentTenant.id > 0
      ? toOptions([source.currentTenant])
      : [];
  }

  if (
    source.userId &&
    source.userId > 0 &&
    canListLoginTenants(source.accessCodes)
  ) {
    const userTenants = activeLoginTenants(
      await authLoginTenants(source.userId),
    );
    if (userTenants.length > 0) {
      return toOptions(userTenants);
    }
  }

  const visibleTenants = activeLoginTenants(source.tenants);
  if (visibleTenants.length > 0) {
    return toOptions(visibleTenants);
  }

  if (!canListPlatformTenants(source.accessCodes)) {
    return [];
  }
  const result = await platformTenantList({
    pageNum: 1,
    pageSize: 100,
    status: 'active',
  });
  return toOptions(result.items);
}
