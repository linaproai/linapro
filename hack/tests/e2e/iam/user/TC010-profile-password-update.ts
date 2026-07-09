import type { APIRequestContext, Page } from '@playwright/test';

import { mkdir } from 'node:fs/promises';
import path from 'node:path';

import { test, expect } from '../../../fixtures/auth';
import { LoginPage } from '../../../pages/LoginPage';
import { ProfilePage } from '../../../pages/ProfilePage';
import { createAdminApiContext, expectSuccess } from '../../../support/api/job';

type CreateUserResult = {
  id: number;
};

type ApiEnvelope<T> = {
  code: number;
  data: T;
  message?: string;
};

type JsonResponse = {
  json(): Promise<unknown>;
  ok(): boolean;
  url(): string;
};

function testUsername() {
  return `e2e_profile_pwd_${Date.now()}`;
}

async function createTemporaryUser(api: APIRequestContext, username: string) {
  const password = 'test123456';
  const result = await expectSuccess<CreateUserResult>(
    await api.post('user', {
      data: {
        username,
        password,
        nickname: '个人中心密码测试',
        status: 1,
      },
    }),
  );
  return { id: result.id, password };
}

async function deleteTemporaryUser(api: APIRequestContext, userID: number) {
  await expectSuccess(await api.delete(`user?ids=${userID}`));
}

async function expectProfileUpdateSuccess(response: JsonResponse) {
  expect(response.ok()).toBeTruthy();
  const payload = (await response.json()) as ApiEnvelope<unknown>;
  expect(
    payload.code,
    `expected profile update success for ${response.url()}: ${payload.message ?? ''}`,
  ).toBe(0);
}

async function captureEvidence(page: Page, name: string) {
  const now = new Date();
  const day = new Intl.DateTimeFormat('en-CA', {
    day: '2-digit',
    month: '2-digit',
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
  })
    .format(now)
    .replaceAll('-', '');
  const time = new Intl.DateTimeFormat('en-GB', {
    hour: '2-digit',
    hour12: false,
    minute: '2-digit',
    second: '2-digit',
    timeZone: 'Asia/Shanghai',
  })
    .format(now)
    .replaceAll(':', '');
  const dir = path.resolve(process.cwd(), '..', '..', 'temp', day);
  await mkdir(dir, { recursive: true });
  await page.screenshot({
    fullPage: true,
    path: path.join(dir, `${time}-${name}.png`),
  });
}

test.describe('TC-10 个人中心修改密码', () => {
  test('TC-10a: 仅提交 password 字段时修改密码成功', async ({ page }) => {
    const adminApi = await createAdminApiContext();
    const username = testUsername();
    const newPassword = 'PatchPwd12345';
    let userID = 0;

    try {
      const user = await createTemporaryUser(adminApi, username);
      userID = user.id;

      const loginPage = new LoginPage(page);
      await loginPage.goto();
      await loginPage.loginAndWaitForRedirect(username, user.password);

      const profilePage = new ProfilePage(page);
      await profilePage.goto();
      await profilePage.openPasswordTab();
      await expect(profilePage.passwordForm).toBeVisible();
      await captureEvidence(page, 'profile-password-form');

      const requestPromise = page.waitForRequest(
        (request) =>
          request.url().includes('/api/v1/user/profile') &&
          request.method() === 'PUT',
      );
      const responsePromise = page.waitForResponse(
        (response) =>
          response.url().includes('/api/v1/user/profile') &&
          response.request().method() === 'PUT',
      );

      await profilePage.submitPasswordChange(user.password, newPassword);

      const request = await requestPromise;
      expect(request.postDataJSON()).toEqual({ password: newPassword });

      await expectProfileUpdateSuccess(await responsePromise);
      await captureEvidence(page, 'profile-password-success');
    } finally {
      if (userID > 0) {
        await deleteTemporaryUser(adminApi, userID);
      }
      await adminApi.dispose();
    }
  });
});
