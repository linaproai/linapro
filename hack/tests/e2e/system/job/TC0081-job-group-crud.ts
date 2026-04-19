import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { JobGroupPage } from '../../../pages/JobGroupPage';

import {
  createAdminApiContext,
  deleteGroup,
  listGroups,
} from './helpers';

test.describe('TC-81 定时任务分组 CRUD', () => {
  const code = `e2e_job_group_${Date.now()}`;
  const name = `E2E任务分组_${Date.now()}`;
  let api: APIRequestContext;
  let createdGroupId = 0;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
  });

  test.afterAll(async () => {
    if (createdGroupId) {
      await api.delete(`job-group/${createdGroupId}`);
    }
    await api.dispose();
  });

  test('TC-81a~e: 分组支持新增、查询、编辑、删除，默认分组不可删', async ({ adminPage }) => {
    const groupPage = new JobGroupPage(adminPage);
    await groupPage.goto();

    await expect(await groupPage.isDefaultDeleteDisabled()).toBe(true);

    await groupPage.createGroup({
      code,
      name,
      remark: 'E2E group',
      sortOrder: 10,
    });

    await groupPage.fillSearchField('分组编码', code);
    await groupPage.clickSearch();
    await expect(await groupPage.hasGroup(name)).toBe(true);

    const createdList = await listGroups(api, code);
    const created = createdList.list.find((item) => item.code === code);
    expect(created).toBeTruthy();
    createdGroupId = created!.id;

    await groupPage.editSearchedGroup({
      name: `${name}_已修改`,
      remark: 'E2E group updated',
      sortOrder: 20,
    });
    await expect(await groupPage.hasGroup(`${name}_已修改`)).toBe(true);

    await groupPage.deleteSearchedGroup();
    createdGroupId = 0;

    await groupPage.fillSearchField('分组编码', code);
    await groupPage.clickSearch();
    await expect(await groupPage.hasGroup(code)).toBe(false);
  });
});
