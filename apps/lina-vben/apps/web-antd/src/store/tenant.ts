import type { LoginTenant, TenantState } from '#/api/tenant/model';
import type { PlatformTenant, TenantStatus } from '#/api/platform/tenant/model';
import type { Router } from 'vue-router';

import { computed, ref } from 'vue';

import { preferences } from '@vben/preferences';
import { useAccessStore } from '@vben/stores';

import { message } from 'ant-design-vue';
import { defineStore } from 'pinia';

import {
  platformTenantEndImpersonate,
  platformTenantImpersonate,
  platformTenantList,
} from '#/api/platform/tenant';
import { authLoginTenants, authSwitchTenant } from '#/api/tenant';
import { $t } from '#/locales';
import { refreshAccessibleState } from '#/router/access-refresh';

const storageKey = 'linapro:tenant-state';
const impersonationOriginalTokenKey =
  'linapro:tenant-impersonation-original-token';
const impersonationOriginalRefreshTokenKey =
  'linapro:tenant-impersonation-original-refresh-token';

function normalizeTenantOptions(items: LoginTenant[]) {
  return items.filter(
    (item) => item.status !== 'suspended' && item.status !== 'deleted',
  );
}

function toPlatformTenant(item: LoginTenant): PlatformTenant {
  const status = (item.status || 'active') as TenantStatus;
  return {
    code: item.code,
    id: item.id,
    name: item.name,
    status,
  };
}

function readOriginalAccessToken() {
  if (typeof localStorage === 'undefined') {
    return '';
  }
  return localStorage.getItem(impersonationOriginalTokenKey) || '';
}

function setOriginalAccessToken(token?: string) {
  if (typeof localStorage === 'undefined') {
    return;
  }
  if (token) {
    localStorage.setItem(impersonationOriginalTokenKey, token);
    return;
  }
  localStorage.removeItem(impersonationOriginalTokenKey);
}

function readOriginalRefreshToken() {
  if (typeof localStorage === 'undefined') {
    return '';
  }
  return localStorage.getItem(impersonationOriginalRefreshTokenKey) || '';
}

function setOriginalRefreshToken(token?: string | null) {
  if (typeof localStorage === 'undefined') {
    return;
  }
  if (token) {
    localStorage.setItem(impersonationOriginalRefreshTokenKey, token);
    return;
  }
  localStorage.removeItem(impersonationOriginalRefreshTokenKey);
}

function readInitialState(): TenantState {
  const fallbackState: TenantState = {
    currentTenant: null,
    enabled: false,
    impersonation: { active: false },
    tenants: [],
  };
  if (typeof localStorage === 'undefined') {
    return fallbackState;
  }
  try {
    const parsed = JSON.parse(
      localStorage.getItem(storageKey) || '',
    ) as Partial<TenantState>;
    const parsedImpersonation = parsed.impersonation ?? { active: false };
    return {
      currentTenant: parsed.currentTenant ?? null,
      enabled: parsed.enabled === true,
      impersonation: parsedImpersonation.active
        ? parsedImpersonation
        : { active: false },
      tenants: Array.isArray(parsed.tenants) ? parsed.tenants : [],
    };
  } catch {
    return fallbackState;
  }
}

export const useTenantStore = defineStore('tenant', () => {
  const accessStore = useAccessStore();
  const initialState = readInitialState();

  const enabled = ref(initialState.enabled);
  const tenants = ref<LoginTenant[]>(initialState.tenants);
  const currentTenant = ref<LoginTenant | null>(initialState.currentTenant);
  const impersonation = ref<TenantState['impersonation']>(
    initialState.impersonation || { active: false },
  );
  const loadingTenants = ref(false);
  const switching = ref(false);

  const isPlatform = computed(
    () =>
      !enabled.value || !currentTenant.value || currentTenant.value.id === 0,
  );
  const isImpersonation = computed(() => impersonation.value.active);
  const fallbackPath = computed(() =>
    enabled.value && isPlatform.value ? '/platform/tenants' : '/system/user',
  );

  function persist() {
    if (typeof localStorage === 'undefined') {
      return;
    }
    localStorage.setItem(
      storageKey,
      JSON.stringify({
        currentTenant: currentTenant.value,
        enabled: enabled.value,
        impersonation: impersonation.value,
        tenants: tenants.value,
      }),
    );
  }

  function setTenantContext(payload: {
    currentTenant?: LoginTenant | null;
    enabled?: boolean;
    tenants?: LoginTenant[];
  }) {
    if (typeof payload.enabled === 'boolean') {
      enabled.value = payload.enabled;
    }
    if (payload.tenants) {
      tenants.value = payload.tenants;
    }
    if ('currentTenant' in payload) {
      currentTenant.value = payload.currentTenant ?? null;
    }
    persist();
  }

  async function ensureTenantOptions(payload?: {
    force?: boolean;
    isPlatform?: boolean;
    userId?: number;
  }) {
    if (!enabled.value || isImpersonation.value) {
      return tenants.value;
    }
    if (!payload?.force && tenants.value.length > 0) {
      return tenants.value;
    }

    loadingTenants.value = true;
    try {
      if (payload?.isPlatform ?? isPlatform.value) {
        const result = await platformTenantList({
          pageNum: 1,
          pageSize: 100,
          status: 'active',
        });
        tenants.value = normalizeTenantOptions(
          result.items.map((tenant) => ({
            code: tenant.code,
            id: tenant.id,
            name: tenant.name,
            status: tenant.status,
          })),
        );
        persist();
        return tenants.value;
      }

      if (!payload?.userId) {
        return tenants.value;
      }
      tenants.value = normalizeTenantOptions(
        await authLoginTenants(payload.userId),
      );
      persist();
      return tenants.value;
    } finally {
      loadingTenants.value = false;
    }
  }

  async function refreshAccessAndEnterDefaultRoute(router: Router) {
    await refreshAccessibleState(router, {
      forceDefaultRoute: true,
      showLoadingToast: false,
    });
    const defaultPath = resolveFallbackPath();
    if (router.currentRoute.value.path !== defaultPath) {
      await router.replace({ force: true, path: defaultPath });
    }
  }

  async function switchTenant(tenantId: number, router: Router) {
    switching.value = true;
    try {
      if (isPlatform.value) {
        const result = await platformTenantImpersonate(tenantId);
        const tenant = result.tenant;
        const fallbackTenant = tenants.value.find((item) => item.id === tenantId);
        if (tenant) {
          startImpersonation(
            result.accessToken || result.token || '',
            tenant,
          );
          await refreshAccessAndEnterDefaultRoute(router);
          return;
        }
        if (fallbackTenant) {
          startImpersonation(
            result.accessToken || result.token || '',
            toPlatformTenant(fallbackTenant),
          );
          await refreshAccessAndEnterDefaultRoute(router);
        }
        return;
      }

      const result = await authSwitchTenant(tenantId);
      accessStore.setAccessToken(result.accessToken);
      accessStore.setRefreshToken(result.refreshToken ?? null);
      const nextTenant =
        tenants.value.find((item) => item.id === tenantId) ?? null;
      currentTenant.value = nextTenant;
      impersonation.value = { active: false };
      setOriginalAccessToken();
      setOriginalRefreshToken();
      persist();
      await refreshAccessAndEnterDefaultRoute(router);
      message.success($t('pages.multiTenant.messages.switchSuccess'));
    } finally {
      switching.value = false;
    }
  }

  function startImpersonation(
    token: string,
    tenant: NonNullable<TenantState['impersonation']['tenant']>,
  ) {
    if (!token) {
      return;
    }
    setOriginalAccessToken(accessStore.accessToken ?? undefined);
    setOriginalRefreshToken(accessStore.refreshToken);
    accessStore.setAccessToken(token);
    accessStore.setRefreshToken(null);
    enabled.value = true;
    currentTenant.value = {
      code: tenant.code,
      id: tenant.id,
      name: tenant.name,
    };
    impersonation.value = { active: true, tenant };
    persist();
  }

  async function exitImpersonation(router: Router) {
    const activeImpersonation = impersonation.value;
    if (activeImpersonation.active && activeImpersonation.tenant?.id) {
      await platformTenantEndImpersonate(activeImpersonation.tenant.id);
    }
    const originalAccessToken = readOriginalAccessToken();
    if (originalAccessToken) {
      accessStore.setAccessToken(originalAccessToken);
    }
    const originalRefreshToken = readOriginalRefreshToken();
    accessStore.setRefreshToken(originalRefreshToken || null);
    setOriginalAccessToken();
    setOriginalRefreshToken();
    impersonation.value = { active: false };
    currentTenant.value = null;
    persist();
    await refreshAccessAndEnterDefaultRoute(router);
  }

  function resolveFallbackPath(candidate?: string) {
    if (!enabled.value) {
      return candidate || preferences.app.defaultHomePath || '/';
    }
    if (isPlatform.value) {
      return '/platform/tenants';
    }
    if (
      candidate &&
      !candidate.startsWith('/platform') &&
      !candidate.startsWith('/tenant')
    ) {
      return candidate;
    }
    return '/system/user';
  }

  function $reset() {
    enabled.value = false;
    tenants.value = [];
    currentTenant.value = null;
    impersonation.value = { active: false };
    setOriginalAccessToken();
    setOriginalRefreshToken();
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem(storageKey);
    }
  }

  return {
    $reset,
    currentTenant,
    enabled,
    ensureTenantOptions,
    exitImpersonation,
    fallbackPath,
    impersonation,
    isImpersonation,
    isPlatform,
    loadingTenants,
    resolveFallbackPath,
    setTenantContext,
    startImpersonation,
    switching,
    switchTenant,
    tenants,
  };
});
