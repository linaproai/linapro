import type { APIRequestContext, Page } from '@playwright/test';

import { request as playwrightRequest } from '@playwright/test';

import { test, expect } from '../../fixtures/auth';
import { config } from '../../fixtures/config';

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? 'http://127.0.0.1:8080/api/v1/';

type CurrentUserProfile = {
  avatar: string;
  email: string;
  realName: string;
  username: string;
};

function unwrapApiData(payload: any) {
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return payload.data;
  }
  return payload;
}

async function createAdminApiContext(): Promise<APIRequestContext> {
  const loginApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await loginApi.post('auth/login', {
    data: {
      password: config.adminPass,
      username: config.adminUser,
    },
  });
  expect(
    loginResponse.ok(),
    `管理员登录 API 失败, status=${loginResponse.status()}`,
  ).toBeTruthy();

  const loginResult = unwrapApiData(await loginResponse.json());
  const accessToken = loginResult?.accessToken;
  expect(accessToken, '未获取到 accessToken').toBeTruthy();
  await loginApi.dispose();

  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

async function fetchCurrentUserProfile(
  adminApi: APIRequestContext,
): Promise<CurrentUserProfile> {
  const response = await adminApi.get('user/info');
  expect(response.ok(), `查询当前用户信息失败, status=${response.status()}`).toBe(
    true,
  );
  const payload = unwrapApiData(await response.json());
  return {
    avatar: payload?.avatar ?? '',
    email: payload?.email ?? '',
    realName: payload?.realName ?? '',
    username: payload?.username ?? '',
  };
}

function getAvatarFallbackText(profile: CurrentUserProfile) {
  const text = (profile.realName || profile.username).trim();
  return text ? text.slice(-2).toUpperCase() : '';
}

function getUserDropdownLocators(adminPage: Page) {
  return {
    menu: adminPage.getByTestId('layout-user-dropdown-menu'),
    profile: adminPage.getByTestId('layout-user-dropdown-profile'),
    trigger: adminPage.getByTestId('layout-user-dropdown-trigger'),
    triggerAvatar: adminPage.getByTestId('layout-user-dropdown-trigger-avatar'),
  };
}

test.describe('TC0046 用户头像下拉菜单', () => {
  let adminApi: APIRequestContext;
  let currentUserProfile: CurrentUserProfile;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    currentUserProfile = await fetchCurrentUserProfile(adminApi);
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test('TC0046a: 下拉菜单不显示文档、Github、问题&帮助', async ({
    adminPage,
  }) => {
    await adminPage.goto('/');
    await adminPage.waitForLoadState('networkidle');

    const { menu: dropdownContent, trigger: avatarTrigger } =
      getUserDropdownLocators(adminPage);
    await avatarTrigger.click();

    await expect(dropdownContent).toBeVisible();
    const menuItems = dropdownContent.locator('[role="menuitem"]');
    const count = await menuItems.count();
    const menuTexts: string[] = [];
    for (let i = 0; i < count; i++) {
      const text = await menuItems.nth(i).textContent();
      if (text) menuTexts.push(text.trim());
    }

    // Verify removed menu items do NOT exist
    expect(menuTexts.join(',')).not.toContain('文档');
    expect(menuTexts.join(',')).not.toContain('GitHub');
    expect(menuTexts.join(',')).not.toMatch(/问题/);

    // Verify "个人中心" still exists (with possible ant-design spacing)
    expect(menuTexts.join(',')).toMatch(/个\s*人\s*中\s*心/);
  });

  test('TC0046b: 下拉菜单显示正确的用户昵称和邮箱', async ({
    adminPage,
  }) => {
    await adminPage.goto('/');
    await adminPage.waitForLoadState('networkidle');

    const { menu: dropdownContent, profile, trigger: avatarTrigger } =
      getUserDropdownLocators(adminPage);
    await avatarTrigger.click();

    await expect(
      adminPage.getByText('ann.vben@gmail.com'),
    ).toHaveCount(0);

    await expect(dropdownContent).toBeVisible();
    await expect(profile.getByTestId('layout-user-dropdown-name')).toHaveText(
      currentUserProfile.realName,
    );
    if (currentUserProfile.username) {
      await expect(profile.getByTestId('layout-user-dropdown-tag')).toHaveText(
        currentUserProfile.username,
      );
    } else {
      await expect(profile.getByTestId('layout-user-dropdown-tag')).toHaveCount(
        0,
      );
    }
    if (currentUserProfile.email) {
      await expect(
        profile.getByTestId('layout-user-dropdown-description'),
      ).toHaveText(currentUserProfile.email);
    } else {
      await expect(
        profile.getByTestId('layout-user-dropdown-description'),
      ).toHaveText('');
    }
  });

  test('TC0046c: 页面右上角应展示用户头像或默认头像', async ({
    adminPage,
  }) => {
    await adminPage.goto('/');
    await adminPage.waitForLoadState('networkidle');

    const {
      trigger: avatarTrigger,
      triggerAvatar,
    } = getUserDropdownLocators(adminPage);
    await expect(avatarTrigger).toBeVisible();
    await expect(triggerAvatar).toBeVisible();

    const avatarImages = triggerAvatar.locator('img[alt]');
    const hasVisibleAvatarImage =
      (await avatarImages.count()) > 0 &&
      (await avatarImages.first().isVisible());
    const avatarFallbackText = getAvatarFallbackText(currentUserProfile);
    const hasVisibleAvatarFallback = avatarFallbackText
      ? await triggerAvatar
          .getByText(avatarFallbackText, { exact: true })
          .first()
          .isVisible()
          .catch(() => false)
      : false;

    expect(hasVisibleAvatarImage || hasVisibleAvatarFallback).toBeTruthy();

    if (hasVisibleAvatarImage) {
      await expect(avatarImages.first()).toHaveAttribute(
        'alt',
        currentUserProfile.realName,
      );
      const src = await avatarImages.first().getAttribute('src');
      expect(src).toBeTruthy();
      expect(src!.length).toBeGreaterThan(0);
    }
  });
});
