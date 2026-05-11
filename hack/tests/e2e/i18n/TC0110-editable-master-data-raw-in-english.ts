import type { APIRequestContext, Page } from '@playwright/test';

import { test, expect } from '../../fixtures/auth';
import { ensureSourcePluginEnabled } from '../../fixtures/plugin';
import { ConfigPage } from '../../pages/ConfigPage';
import { DeptPage } from '../../../../apps/lina-plugins/org-center/hack/tests/pages/DeptPage';
import { NoticePage } from '../../../../apps/lina-plugins/content-notice/hack/tests/pages/NoticePage';
import { PostPage } from '../../../../apps/lina-plugins/org-center/hack/tests/pages/PostPage';
import { RolePage } from '../../pages/RolePage';
import { UserPage } from '../../pages/UserPage';
import { createAdminApiContext, expectSuccess } from '../../support/api/job';

test.describe('TC0110 可编辑主数据退出 i18n 投影专项回归', () => {
  test('TC-110a: 英文环境下用户与组织管理页面中的可编辑主数据保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const userPage = new UserPage(adminPage);
    const deptPage = new DeptPage(adminPage);
    const postPage = new PostPage(adminPage);
    const rolePage = new RolePage(adminPage);

    await ensureOrgRawData(adminPage);
    await mainLayout.switchLanguage('English');

    await userPage.goto();
    await expect(await userPage.hasDeptTreeNode('研发部门')).toBe(true);

    await deptPage.goto();
    await expect(await deptPage.hasDeptInExpandedTree('研发部门')).toBe(true);

    await postPage.goto();
    await expect(await postPage.hasPostName('总经理')).toBe(true);

    await rolePage.goto();
    await expect(await rolePage.hasRole('User')).toBe(true);
  });

  test('TC-110b: 英文环境下参数管理页中的可编辑主数据保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const configPage = new ConfigPage(adminPage);

    await mainLayout.switchLanguage('English');

    await configPage.goto();
    await configPage.fillSearchField('参数键名', 'demo.notice.banner');
    await configPage.clickSearch();
    const configRow = configPage.findRowByExactKey('demo.notice.banner');
    await expect(configRow).toBeVisible();
    await expect(configRow).toContainText('demo.notice.banner');
    await expect(configRow).toContainText('欢迎使用 LinaPro');
  });

  test('TC-110c: 英文环境下通知管理页中的可编辑业务记录保持数据库原值', async ({
    adminPage,
    mainLayout,
  }) => {
    const noticePage = new NoticePage(adminPage);

    await ensureNoticeRawData(adminPage);
    await mainLayout.switchLanguage('English');

    await noticePage.goto();
    await expect(await noticePage.hasNotice('系统升级通知')).toBe(true);
  });
});

async function ensureOrgRawData(page: Page) {
  await ensureSourcePluginEnabled(page, 'org-center');
  const api = await createAdminApiContext();
  try {
    const dept = await ensureDept(api);
    await ensurePost(api, dept.id);
  } finally {
    await api.dispose();
  }
}

async function ensureNoticeRawData(page: Page) {
  await ensureSourcePluginEnabled(page, 'content-notice');
  const api = await createAdminApiContext();
  try {
    await ensureNotice(api);
  } finally {
    await api.dispose();
  }
}

async function ensureDept(api: APIRequestContext) {
  const existing = await expectSuccess<{ list: Array<{ id: number; name: string }> }>(
    await api.get(`dept?name=${encodeURIComponent('研发部门')}`),
  );
  const dept = existing.list.find((item) => item.name === '研发部门');
  if (dept) {
    return dept;
  }
  return expectSuccess<{ id: number }>(
    await api.post('dept', {
      data: {
        code: 'e2e-raw-dev',
        name: '研发部门',
        orderNum: 1,
        parentId: 0,
        status: 1,
      },
    }),
  );
}

async function ensurePost(api: APIRequestContext, deptId: number) {
  const existing = await expectSuccess<{
    list: Array<{ id: number; name: string }>;
  }>(await api.get(`post?pageNum=1&pageSize=100&name=${encodeURIComponent('总经理')}`));
  if (existing.list.some((item) => item.name === '总经理')) {
    return;
  }
  await expectSuccess(
    await api.post('post', {
      data: {
        code: 'E2E_RAW_CEO',
        deptId,
        name: '总经理',
        sort: 1,
        status: 1,
      },
    }),
  );
}

async function ensureNotice(api: APIRequestContext) {
  const existing = await expectSuccess<{
    list: Array<{ id: number; title: string }>;
  }>(
    await api.get(
      `notice?pageNum=1&pageSize=100&title=${encodeURIComponent('系统升级通知')}`,
    ),
  );
  if (existing.list.some((item) => item.title === '系统升级通知')) {
    return;
  }
  await expectSuccess(
    await api.post('notice', {
      data: {
        content:
          '<p>系统将于本周六凌晨2:00-4:00进行升级维护，届时系统将暂停服务。</p>',
        status: 1,
        title: '系统升级通知',
        type: 1,
      },
    }),
  );
}
