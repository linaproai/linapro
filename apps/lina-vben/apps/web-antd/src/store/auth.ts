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

function hasMultiTenantMenu(items: UserMenuNode[] = []): boolean {
  return items.some((item) => {
    const path = item.path || '';
    const name = item.name || '';
    return (
      path.startsWith('/platform') ||
      path.startsWith('/tenant') ||
      name.startsWith('Platform') ||
      name.startsWith('Tenant') ||
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
      loginLoading.value = false;
    }
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
      enabled:
        tenantStore.enabled ||
        resolveTenantEnabled(
          tenantStore.tenants,
          userInfo,
          tenantStore.currentTenant,
        ),
    });

    return userInfo;
  }

  function $reset() {
    loginLoading.value = false;
  }

  return {
    $reset,
    authLogin,
    clearSession,
    fetchUserInfo,
    loginLoading,
    logout,
    pendingPreToken,
    selectTenant,
  };
});
