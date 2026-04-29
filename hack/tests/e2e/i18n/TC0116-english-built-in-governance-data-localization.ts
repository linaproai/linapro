import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';
import { JobGroupPage } from '../../pages/JobGroupPage';
import { JobLogPage } from '../../pages/JobLogPage';
import { JobPage } from '../../pages/JobPage';
import {
  createAdminApiContext,
  expectSuccess,
  getLog,
  triggerJob,
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
    ).toContainText('Unavailable');

    const dataText = await dataPanel.locator('.vxe-table--body').innerText();
    expect(dataText).not.toContain('启用');
    expect(dataText).not.toContain('停用');
    expect(dataText).not.toContain('插件处理器不可用');
    expect(dataText).not.toContain('不可用');
  });

  test('TC-116b: 调度中心任务、分组和执行日志列表按英文展示内置调度数据', async ({
    adminPage,
    mainLayout,
  }) => {
    const jobGroupPage = new JobGroupPage(adminPage);
    const jobPage = new JobPage(adminPage);
    const jobLogPage = new JobLogPage(adminPage);

    await mainLayout.switchLanguage('English');

    const api = await createAdminApiContext();
    try {
      const groupData = await expectSuccess<{
        list: Array<{ code: string; name: string; remark: string }>;
        total: number;
      }>(
        await api.get('job-group?pageNum=1&pageSize=100', {
          headers: { 'Accept-Language': 'en-US' },
        }),
      );
      const defaultGroup = groupData.list.find((item) => item.code === 'default');
      expect(defaultGroup?.name).toBe('Default Group');
      expect(defaultGroup?.remark).toContain('system default job group');

      const jobData = await expectSuccess<{
        list: Array<{
          description: string;
          groupCode: string;
          groupName: string;
          handlerRef: string;
          id: number;
          name: string;
        }>;
        total: number;
      }>(
        await api.get('job?pageNum=1&pageSize=100', {
          headers: { 'Accept-Language': 'en-US' },
        }),
      );
      const cleanupJob = jobData.list.find(
        (item) => item.handlerRef === 'host:cleanup-job-logs',
      );
      expect(cleanupJob?.name).toBe('Job Log Cleanup');
      expect(cleanupJob?.description).toContain('scheduled-job execution logs');
      expect(cleanupJob?.groupName).toBe('Default Group');
      expect(cleanupJob?.id).toBeGreaterThan(0);
      const triggeredCleanup = await triggerJob(api, cleanupJob!.id);
      await expect
        .poll(
          async () => {
            const detail = await getLog(api, triggeredCleanup.logId);
            return detail.status;
          },
          {
            message: 'built-in cleanup job log should be ready before UI assertion',
            timeout: 10_000,
          },
        )
        .toBe('success');

      const handlerData = await expectSuccess<{
        list: Array<{ displayName: string; ref: string }>;
      }>(
        await api.get('job/handler', {
          headers: { 'Accept-Language': 'en-US' },
        }),
      );
      const cleanupHandler = handlerData.list.find(
        (item) => item.ref === 'host:cleanup-job-logs',
      );
      expect(cleanupHandler?.displayName).toBe('Job Log Cleanup');
    } finally {
      await api.dispose();
    }

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
      const loginLogData = await expectSuccess<{
        items: Array<{ msg: string; status: number }>;
        total: number;
      }>(
        await api.get('loginlog?pageNum=1&pageSize=10', {
          headers: { 'Accept-Language': 'en-US' },
        }),
      );
      expect(loginLogData.items.some((item) => item.msg === 'Login successful')).toBe(
        true,
      );
      const response = await api.get('dict/type/export?pageNum=1&pageSize=1');
      expect(response.ok()).toBeTruthy();
      await expectSuccess(await api.post('auth/logout'));
    } finally {
      await api.dispose();
    }

    const auditApi = await createAdminApiContext();
    try {
      const operLogData = await expectSuccess<{
        items: Array<{
          operSummary: string;
          operType: string;
          status: number;
          title: string;
        }>;
      }>(
        await auditApi.get('operlog?pageNum=1&pageSize=20', {
          headers: { 'Accept-Language': 'en-US' },
        }),
      );
      const logoutLog = operLogData.items.find(
        (item) => item.title === 'Authentication' && item.operSummary === 'User logout',
      );
      expect(logoutLog?.operType).toBe('create');
      expect(logoutLog?.status).toBe(0);

      const exportLog = operLogData.items.find(
        (item) =>
          item.title === 'Dictionary Management' &&
          item.operSummary === 'Export dictionary type',
      );
      expect(exportLog?.operType).toBe('export');
      expect(exportLog?.status).toBe(0);
    } finally {
      await auditApi.dispose();
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
      .toContain('Export dictionary type');
    const operLogText = await adminPage.locator('body').innerText();
    expect(operLogText).toContain('Authentication');
    expect(operLogText).toContain('User logout');
    expect(operLogText).toContain('Dictionary Management');
    expect(operLogText).toContain('Export');
    expect(operLogText).not.toContain('认证管理');
    expect(operLogText).not.toContain('用户登出');
    expect(operLogText).not.toContain('导出字典类型');
    expect(operLogText).not.toContain('字典管理');
  });
});
