import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';
import { JobGroupPage } from '../../pages/JobGroupPage';
import { JobLogPage } from '../../pages/JobLogPage';
import { JobPage } from '../../pages/JobPage';
import {
  createAdminApiContext,
  expectSuccess,
} from '../../support/api/job';
import { waitForRouteReady } from '../../support/ui';

test.describe('TC0116 英文环境内置治理数据本地化回归', () => {
  test('TC-116a: 字典管理内置字典类型与数据列表按英文展示', async ({
    adminPage,
    mainLayout,
  }) => {
    const dictPage = new DictPage(adminPage);

    await mainLayout.switchLanguage('English');
    await dictPage.goto();
    await dictPage.fillTypeSearchField('字典类型', 'cron_job_status');
    await dictPage.clickTypeSearch();

    const typePanel = adminPage.locator('#dict-type');
    const typeRow = typePanel
      .locator('.vxe-body--row', {
        hasText: 'cron_job_status',
      })
      .first();
    await expect(typeRow).toContainText('Scheduled Job Status');
    await expect(typeRow).not.toContainText('定时任务状态');

    await dictPage.clickTypeRow('cron_job_status');

    const dataPanel = adminPage.locator('#dict-data');
    await expect(
      dataPanel.locator('.vxe-body--row', { hasText: 'enabled' }),
    ).toContainText('Enabled');
    await expect(
      dataPanel.locator('.vxe-body--row', { hasText: 'disabled' }),
    ).toContainText('Disabled');
    await expect(
      dataPanel.locator('.vxe-body--row', { hasText: 'paused_by_plugin' }),
    ).toContainText('Plugin Handler Unavailable');

    const dataText = await dataPanel.locator('.vxe-table--body').innerText();
    expect(dataText).not.toContain('启用');
    expect(dataText).not.toContain('停用');
    expect(dataText).not.toContain('插件处理器不可用');
  });

  test('TC-116b: 调度中心任务、分组和执行日志列表按英文展示内置调度数据', async ({
    adminPage,
    mainLayout,
  }) => {
    const jobGroupPage = new JobGroupPage(adminPage);
    const jobPage = new JobPage(adminPage);
    const jobLogPage = new JobLogPage(adminPage);

    await mainLayout.switchLanguage('English');

    await jobGroupPage.goto();
    await expect(await jobGroupPage.hasGroup('Default Group')).toBe(true);
    await expect(await jobGroupPage.hasGroup('默认分组')).toBe(false);

    await jobPage.goto();
    await expect(await jobPage.hasJob('Job Log Cleanup')).toBe(true);
    await expect(await jobPage.hasJob('任务日志清理')).toBe(false);

    await jobLogPage.goto();
    await expect
      .poll(async () =>
        adminPage
          .locator('[data-testid="job-log-page"] .vxe-table--body')
          .first()
          .innerText(),
      )
      .toMatch(
        /Job Log Cleanup|Online Session Cleanup|Server Monitor Collection|Server Monitor Cleanup/,
      );

    const jobLogText = await adminPage
      .locator('[data-testid="job-log-page"] .vxe-table--body')
      .first()
      .innerText();
    expect(jobLogText).not.toMatch(
      /默认分组|任务日志清理|在线会话清理|服务监控采集|服务监控清理/,
    );
  });

  test('TC-116c: 审计日志与登录日志列表按英文展示内置类型、状态和摘要', async ({
    adminPage,
    mainLayout,
  }) => {
    const api = await createAdminApiContext();
    try {
      const response = await api.get('dict/type/export?pageNum=1&pageSize=1');
      expect(response.ok()).toBeTruthy();
      await expectSuccess(await api.post('auth/logout'));
    } finally {
      await api.dispose();
    }

    await mainLayout.switchLanguage('English');

    await adminPage.goto('/monitor/loginlog');
    await waitForRouteReady(adminPage);
    await expect
      .poll(async () => adminPage.locator('body').innerText())
      .toContain('Login successful');
    const loginLogText = await adminPage.locator('body').innerText();
    expect(loginLogText).toContain('Success');
    expect(loginLogText).not.toContain('登录成功');

    await adminPage.goto('/monitor/operlog');
    await waitForRouteReady(adminPage);
    await expect
      .poll(async () => adminPage.locator('body').innerText(), {
        timeout: 10_000,
      })
      .toContain('Export Dictionary Types');
    const operLogText = await adminPage.locator('body').innerText();
    expect(operLogText).toContain('Authentication');
    expect(operLogText).toContain('User Logout');
    expect(operLogText).toContain('Dictionaries');
    expect(operLogText).toContain('Export');
    expect(operLogText).toContain('Success');
    expect(operLogText).not.toContain('认证管理');
    expect(operLogText).not.toContain('用户登出');
    expect(operLogText).not.toContain('导出字典类型');
    expect(operLogText).not.toContain('字典管理');
  });
});
