import type { Recordable } from '@vben/types';

import type { AppUserInfo } from '#/api/core/user';
import type { LoginTenant } from '#/api/tenant/model';

import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { LOGIN_PATH } from '@vben/constants';
import { preferences } from '@vben/preferences';
import { resetAllStores, useAccessStore, useUserStore } from '@vben/stores';

import { notification } from 'ant-design-vue';
import { defineStore } from 'pinia';

import { getUserInfoApi, loginApi, logoutApi } from '#/api';
import { authSelectTenant } from '#/api/tenant';
import { $t } from '#/locales';
import { useTenantStore } from '#/store/tenant';

type UserMenuNode = {
  children?: UserMenuNode[];
  name?: string;
  path?: string;
};

function normalizeMenuPath(path: string) {
  return path.replace(/^\/+/u, '').replace(/\/+$/u, '');
}

function isMultiTenantMenuNode(item: UserMenuNode): boolean {
  const path = normalizeMenuPath(item.path || '');
  const name = item.name || '';

  // The host platform group is always present; only concrete tenant pages
  // should enable tenant-aware frontend behavior.
  return (
    path === 'platform/tenants' ||
    path.startsWith('platform/tenants/') ||
    path === 'tenant' ||
    path.startsWith('tenant/') ||
    name.startsWith('PlatformTenant') ||
    name.startsWith('Tenant')
  );
}

function hasMultiTenantMenu(items: UserMenuNode[] = []): boolean {
  return items.some((item) => {
    return (
      isMultiTenantMenuNode(item) ||
      hasMultiTenantMenu(item.children)
    );
  });
}

function resolveTenantEnabled(
  tenants: LoginTenant[],
  userInfo: AppUserInfo | null,
  currentTenant: LoginTenant | null,
) {
  return (
    tenants.length > 0 ||
    !!currentTenant ||
    hasMultiTenantMenu((userInfo?.menus ?? []) as UserMenuNode[])
  );
}

export const useAuthStore = defineStore('auth', () => {
  const accessStore = useAccessStore();
  const userStore = useUserStore();
  const router = useRouter();
  const tenantStore = useTenantStore();

  const loginLoading = ref(false);
  const pendingPreToken = ref('');
  const tenantLoginTransitioning = ref(false);

  /**
   * 异步处理登录操作
   * Asynchronously handle the login process
   * @param params 登录表单数据
   */
  async function authLogin(
    params: Recordable<any>,
    onSuccess?: () => Promise<void> | void,
  ) {
    // 异步处理用户登录操作并获取 accessToken
    let userInfo: AppUserInfo | null = null;
    try {
      loginLoading.value = true;
      const loginResult = await loginApi(params);
      const { accessToken, preToken, refreshToken } = loginResult;
      const tenants = Array.isArray(loginResult.tenants)
        ? loginResult.tenants
        : [];

      if (preToken && tenants.length > 1 && !accessToken) {
        pendingPreToken.value = preToken;
        tenantStore.setTenantContext({
          currentTenant: null,
          enabled: true,
          tenants,
        });
        return { requiresTenantSelection: true, tenants, userInfo };
      }

      // 如果成功获取到 accessToken
      if (accessToken) {
        accessStore.setAccessToken(accessToken);
        accessStore.setRefreshToken(refreshToken ?? null);

        // 获取用户信息并存储到 accessStore 中
        userInfo = await fetchUserInfo();
        userStore.setUserInfo(userInfo);
        tenantStore.setTenantContext({
          currentTenant: tenants.length === 1 ? tenants[0] : null,
          enabled: resolveTenantEnabled(tenants, userInfo, tenants[0] ?? null),
          tenants,
        });

        if (accessStore.loginExpired) {
          accessStore.setLoginExpired(false);
        } else {
          onSuccess
            ? await onSuccess?.()
            : await router.push(
                tenantStore.resolveFallbackPath(
                  userInfo.homePath || preferences.app.defaultHomePath,
                ),
              );
        }

        if (userInfo?.realName) {
          notification.success({
            description: `${$t('authentication.loginSuccessDesc')}: ${userInfo.realName}`,
            duration: 3,
            message: $t('authentication.loginSuccess'),
          });
        }
      }
    } finally {
      loginLoading.value = false;
    }

    return {
      requiresTenantSelection: false,
      userInfo,
    };
  }

  async function selectTenant(tenantId: number) {
    if (!pendingPreToken.value) {
      return;
    }
    try {
      loginLoading.value = true;
      tenantLoginTransitioning.value = true;
      const { accessToken, refreshToken } = await authSelectTenant(
        pendingPreToken.value,
        tenantId,
      );
      pendingPreToken.value = '';
      accessStore.setAccessToken(accessToken);
      accessStore.setRefreshToken(refreshToken ?? null);
      const selectedTenant =
        tenantStore.tenants.find((item) => item.id === tenantId) ?? null;
      tenantStore.setTenantContext({
        currentTenant: selectedTenant,
        enabled: true,
      });
      const userInfo = await fetchUserInfo();
      userStore.setUserInfo(userInfo);
      await router.push(
        tenantStore.resolveFallbackPath(
          userInfo.homePath || preferences.app.defaultHomePath,
        ),
      );
      notification.success({
        description: selectedTenant?.name || '',
        duration: 3,
        message: $t('pages.multiTenant.messages.tenantSelected'),
      });
    } finally {
      tenantLoginTransitioning.value = false;
      loginLoading.value = false;
    }
  }

  /**
   * 消费源码插件 OAuth 回调投递的登录结果。
   *
   * 回调通过 /oauth-handoff 路由的 query 参数把宿主 LoginByExternal 的结果带过来：
   * - 单租户用户携带 accessToken/refreshToken，可直接进入工作台；
   * - 多租户用户携带 preToken + tenants，需要继续走租户选择流程。
   */
  async function completeOAuthHandoff(payload: {
    accessToken?: string;
    preToken?: string;
    redirect?: string;
    refreshToken?: string;
    tenants?: LoginTenant[];
  }) {
    const tenants = Array.isArray(payload.tenants) ? payload.tenants : [];

    if (payload.accessToken) {
      try {
        loginLoading.value = true;
        accessStore.setAccessToken(payload.accessToken);
        accessStore.setRefreshToken(payload.refreshToken ?? null);
        const userInfo = await fetchUserInfo();
        userStore.setUserInfo(userInfo);
        tenantStore.setTenantContext({
          currentTenant: tenants.length === 1 ? tenants[0] : null,
          enabled: resolveTenantEnabled(tenants, userInfo, tenants[0] ?? null),
          tenants,
        });
        accessStore.setLoginExpired(false);
        await router.replace(
          tenantStore.resolveFallbackPath(
            payload.redirect ||
              userInfo.homePath ||
              preferences.app.defaultHomePath,
          ),
        );
        if (userInfo?.realName) {
          notification.success({
            description: `${$t('authentication.loginSuccessDesc')}: ${userInfo.realName}`,
            duration: 3,
            message: $t('authentication.loginSuccess'),
          });
        }
      } finally {
        loginLoading.value = false;
      }
      return { requiresTenantSelection: false };
    }

    if (payload.preToken && tenants.length > 0) {
      pendingPreToken.value = payload.preToken;
      tenantStore.setTenantContext({
        currentTenant: null,
        enabled: true,
        tenants,
      });
      await router.replace({
        path: LOGIN_PATH,
        query: payload.redirect
          ? { redirect: encodeURIComponent(payload.redirect) }
          : {},
      });
      return { requiresTenantSelection: true, tenants };
    }

    throw new Error('OAuth handoff payload is missing tokens or pre-login token');
  }

  async function clearSession(redirect: boolean = true) {
    resetAllStores();
    tenantStore.$reset();
    accessStore.setLoginExpired(false);

    // 回登录页带上当前路由地址
    await router.replace({
      path: LOGIN_PATH,
      query: redirect
        ? {
            redirect: encodeURIComponent(router.currentRoute.value.fullPath),
          }
        : {},
    });
  }

  async function logout(redirect: boolean = true) {
    try {
      await logoutApi();
    } catch {
      // 不做任何处理
    }
    await clearSession(redirect);
  }

  async function fetchUserInfo() {
    const userInfo = await getUserInfoApi();
    userStore.setUserInfo(userInfo);

    // Set access codes (permissions) for button-level access control
    if (userInfo.permissions) {
      accessStore.setAccessCodes(userInfo.permissions);
    }
    tenantStore.setTenantContext({
      enabled: resolveTenantEnabled(
        tenantStore.tenants,
        userInfo,
        tenantStore.currentTenant,
      ),
    });

    return userInfo;
  }

  function $reset() {
    loginLoading.value = false;
    pendingPreToken.value = '';
    tenantLoginTransitioning.value = false;
  }

  return {
    $reset,
    authLogin,
    clearSession,
    completeOAuthHandoff,
    fetchUserInfo,
    loginLoading,
    logout,
    pendingPreToken,
    selectTenant,
    tenantLoginTransitioning,
  };
});
