import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../fixtures/auth';
import { DashboardPage } from '../../pages/DashboardPage';
import { ProfilePage } from '../../pages/ProfilePage';
import { RoleAuthUserPage } from '../../pages/RoleAuthUserPage';
import { UserPage } from '../../pages/UserPage';
import { createAdminApiContext, expectSuccess } from '../../support/api/job';
import type { MainLayout } from '../../pages/MainLayout';

type CurrentUserInfo = {
  email: string;
  realName: string;
  username: string;
};

type CurrentUserProfile = {
  email: string;
  id: number;
  nickname: string;
  phone: string;
  sex: number;
  username: string;
};

type RoleListResult = {
  list: Array<{
    id: number;
    key: string;
    name: string;
  }>;
  total: number;
};

function expectedRealName(nickname: string, username: string) {
  return nickname || username;
}

async function getCurrentProfile(api: APIRequestContext) {
  return expectSuccess<CurrentUserProfile>(await api.get('user/profile'));
}

async function getCurrentUserInfo(api: APIRequestContext) {
  return expectSuccess<CurrentUserInfo>(await api.get('user/info'));
}

async function updateCurrentNickname(
  api: APIRequestContext,
  nickname: string,
) {
  await expectSuccess(await api.put('user/profile', { data: { nickname } }));
}

async function waitForCurrentRealName(
  api: APIRequestContext,
  expectedName: string,
) {
  await expect
    .poll(async () => {
      const currentUser = await getCurrentUserInfo(api);
      return currentUser.realName;
    }, { timeout: 10000 })
    .toBe(expectedName);
}

async function resolveAdminRoleId(api: APIRequestContext) {
  const roles = await expectSuccess<RoleListResult>(
    await api.get('role?key=admin&page=1&size=20'),
  );
  const adminRole = roles.list.find((item) => item.key === 'admin');
  expect(adminRole, '缺少 admin 角色').toBeTruthy();
  return adminRole!.id;
}

async function expectRawNicknameVisible(
  mainLayout: MainLayout,
  dashboardPage: DashboardPage,
  profilePage: ProfilePage,
  roleAuthUserPage: RoleAuthUserPage,
  userPage: UserPage,
  username: string,
  nickname: string,
  adminRoleId: number,
) {
  await mainLayout.switchLanguage('English');

  await dashboardPage.gotoWorkspace();
  await expect(
    dashboardPage.workspacePage.getByText(nickname, { exact: false }).first(),
  ).toBeVisible();

  await mainLayout.openUserDropdown();
  await expect(mainLayout.userDropdownName).toHaveText(nickname);

  await profilePage.goto();
  await expect(profilePage.panelDisplayName(nickname)).toBeVisible();
  await expect(profilePage.nicknameInput).toHaveValue(nickname);

  await userPage.goto();
  await userPage.searchByUsername(username);
  await expect(userPage.getUserRow(username)).toContainText(nickname);

  await roleAuthUserPage.goto(adminRoleId);
  await roleAuthUserPage.searchByUsername(username);
  await expect(roleAuthUserPage.userRow(username)).toContainText(nickname);
}

test.describe('TC0109 用户可编辑数据退出 i18n 投影专项回归', () => {
  test.describe.configure({ mode: 'serial' });

  let adminApi: APIRequestContext | null = null;
  let adminRoleId = 0;
  let originalProfile: CurrentUserProfile | null = null;
  let sentinelNickname = '';

  test.beforeAll(async () => {
    const api = await createAdminApiContext();
    const profile = await getCurrentProfile(api);

    adminApi = api;
    originalProfile = profile;
    adminRoleId = await resolveAdminRoleId(api);
    sentinelNickname = `昵称Raw${Date.now().toString().slice(-6)}`;

    await updateCurrentNickname(api, sentinelNickname);
    await waitForCurrentRealName(
      api,
      expectedRealName(sentinelNickname, profile.username),
    );
  });

  test.afterAll(async () => {
    if (!adminApi || !originalProfile) {
      return;
    }
    try {
      await updateCurrentNickname(adminApi, originalProfile.nickname);
      await waitForCurrentRealName(
        adminApi,
        expectedRealName(originalProfile.nickname, originalProfile.username),
      );
    } finally {
      await adminApi.dispose();
    }
  });

  test('TC-109a: 英文环境下昵称在用户可编辑数据展示位保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const dashboardPage = new DashboardPage(adminPage);
    const profilePage = new ProfilePage(adminPage);
    const roleAuthUserPage = new RoleAuthUserPage(adminPage);
    const userPage = new UserPage(adminPage);

    await expectRawNicknameVisible(
      mainLayout,
      dashboardPage,
      profilePage,
      roleAuthUserPage,
      userPage,
      originalProfile!.username,
      sentinelNickname,
      adminRoleId,
    );
  });
});
