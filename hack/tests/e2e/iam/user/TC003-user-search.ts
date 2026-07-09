import { test, expect } from '../../../fixtures/auth';
import { createAdminApiContext } from '../../../fixtures/plugin';
import { UserPage } from '../../../pages/UserPage';

type AdminApi = Awaited<ReturnType<typeof createAdminApiContext>>;

type CreatedUser = {
  id: number;
};

type SearchFixtureUser = {
  id: number;
  nickname: string;
  phone: string;
  username: string;
};

function unwrapApiData(payload: any) {
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return payload.data;
  }
  return payload;
}

function uniqueSuffix() {
  return `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
}

async function createSearchFixtureUser(
  adminApi: AdminApi,
): Promise<SearchFixtureUser> {
  const suffix = uniqueSuffix();
  const fixture = {
    nickname: `E2E Search ${suffix}`,
    phone: `139${Date.now().toString().slice(-8)}`,
    username: `e2e_search_${suffix}`,
  };
  const response = await adminApi.post('user', {
    data: {
      ...fixture,
      password: 'test123456',
      status: 1,
    },
  });
  expect(response.ok(), '创建搜索测试用户失败').toBeTruthy();
  const created = unwrapApiData(await response.json()) as CreatedUser;
  expect(created?.id, '创建搜索测试用户未返回 id').toBeTruthy();
  return { ...fixture, id: created.id };
}

async function deleteSearchFixtureUser(
  adminApi: AdminApi,
  fixture?: SearchFixtureUser,
) {
  if (!fixture?.id) {
    return;
  }
  await adminApi.delete(`user?ids=${fixture.id}`).catch(() => {});
}

test.describe('TC003 用户列表搜索', () => {
  let adminApi: AdminApi;
  let fixtureUser: SearchFixtureUser;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    fixtureUser = await createSearchFixtureUser(adminApi);
  });

  test.afterAll(async () => {
    await deleteSearchFixtureUser(adminApi, fixtureUser);
    await adminApi?.dispose();
  });

  test('TC003a: 按用户名搜索', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.fillSearchField('用户账号', fixtureUser.username);
    await userPage.clickSearch();

    const hasUser = await userPage.hasUser(fixtureUser.username);
    expect(hasUser).toBeTruthy();

    const rowCount = await userPage.getVisibleRowCount();
    expect(rowCount).toBeGreaterThan(0);
  });

  test('TC003b: 按用户昵称搜索', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.fillSearchField('用户昵称', fixtureUser.nickname);
    await userPage.clickSearch();

    await expect(userPage.getUserRow(fixtureUser.username)).toBeVisible();
    const rowCount = await userPage.getVisibleRowCount();
    expect(rowCount).toBeGreaterThan(0);
  });

  test('TC003c: 按手机号搜索', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.fillSearchField('手机号码', fixtureUser.phone);
    await userPage.clickSearch();

    await expect(userPage.getUserRow(fixtureUser.username)).toBeVisible();
    const rowCount = await userPage.getVisibleRowCount();
    expect(rowCount).toBeGreaterThan(0);
  });

  test('TC003d: 搜索后重置恢复全部数据', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Get initial row count
    const initialCount = await userPage.getVisibleRowCount();
    expect(initialCount).toBeGreaterThan(0);

    await userPage.fillSearchField('用户账号', fixtureUser.username);
    await userPage.clickSearch();

    await expect(userPage.getUserRow(fixtureUser.username)).toBeVisible();
    const filteredCount = await userPage.getVisibleRowCount();
    expect(filteredCount).toBeLessThanOrEqual(initialCount);
    expect(filteredCount).toBeGreaterThan(0);

    await userPage.clickReset();

    await userPage.expectSearchFieldValue('用户账号', '');
    const resetCount = await userPage.getVisibleRowCount();
    expect(resetCount).toBeGreaterThanOrEqual(filteredCount);
  });

  test('TC003e: 搜索请求包含正确的筛选参数', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.fillSearchField('用户昵称', fixtureUser.nickname);

    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/user') &&
        req.method() === 'GET' &&
        req.url().includes('nickname'),
      { timeout: 10000 },
    );

    await userPage.clickSearch();
    const request = await requestPromise;

    const url = request.url();
    expect(url).toContain('nickname=');
  });
});
