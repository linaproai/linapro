import { cloneDeep, filterTree } from '@vben/utils';

import { useTenantStore } from '#/store/tenant';

type TenantAccessMode = 'disabled' | 'enabled' | 'platform' | 'tenant';

type TenantAccessRoute = {
  children?: TenantAccessRoute[];
  meta?: {
    tenantAccessMode?: TenantAccessMode;
  };
  name?: unknown;
  path?: string;
};

function inferTenantAccessMode(
  route: TenantAccessRoute,
): TenantAccessMode | undefined {
  const name = typeof route.name === 'string' ? route.name : '';
  const path = route.path || '';
  if (path.startsWith('/platform') || name.startsWith('Platform')) {
    return 'platform';
  }
  if (path.startsWith('/tenant') || name.startsWith('Tenant')) {
    return 'tenant';
  }
  return undefined;
}

function canAccessTenantRoute(route: TenantAccessRoute) {
  const tenantStore = useTenantStore();
  const mode = route.meta?.tenantAccessMode ?? inferTenantAccessMode(route);

  switch (mode) {
    case 'disabled': {
      return !tenantStore.enabled;
    }
    case 'enabled': {
      return tenantStore.enabled;
    }
    case 'platform': {
      return tenantStore.enabled && tenantStore.isPlatform;
    }
    case 'tenant': {
      return false;
    }
    default: {
      return true;
    }
  }
}

function canAccessTenantLocation(location: TenantAccessRoute) {
  const tenantStore = useTenantStore();
  const path = location.path || '';
  const mode = location.meta?.tenantAccessMode ?? inferTenantAccessMode(location);

  switch (mode) {
    case 'disabled': {
      return !tenantStore.enabled;
    }
    case 'enabled': {
      return tenantStore.enabled;
    }
    case 'platform': {
      return tenantStore.enabled && tenantStore.isPlatform;
    }
    case 'tenant': {
      return false;
    }
    default: {
      if (path.startsWith('/platform')) {
        return tenantStore.enabled && tenantStore.isPlatform;
      }
      if (path.startsWith('/tenant')) {
        return false;
      }
      return true;
    }
  }
}

function filterTenantAccessRoutes<T extends TenantAccessRoute>(
  routes: T[],
): T[] {
  return filterTree(cloneDeep(routes), canAccessTenantRoute);
}

export {
  canAccessTenantLocation,
  canAccessTenantRoute,
  filterTenantAccessRoutes,
};
