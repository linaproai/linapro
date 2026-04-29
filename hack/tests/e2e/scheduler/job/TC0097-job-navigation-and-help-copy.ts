import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { JobPage } from '../../../pages/JobPage';

import {
  buildShellJobPayload,
  createAdminApiContext,
  createJob,
  deleteJob,
  expectSuccess,
  getAccessibleMenus,
  getConfigByKey,
  getDefaultGroup,
  listLogs,
  setCronShellEnabled,
  type AccessibleMenuNode,
  updateConfigValue,
} from '../../../support/api/job';

function findMenuByTitle(
  list: AccessibleMenuNode[],
  title: string,
): AccessibleMenuNode | null {
  for (const item of list) {
    if (item.meta?.title === title) {
      return item;
    }
    const nested = findMenuByTitle(item.children ?? [], title);
    if (nested) {
      return nested;
    }
  }
  return null;
}

function formatRetentionSummary(mode: string, value: number) {
  switch (mode) {
    case 'count':
      return `按条数保留最近 ${value} 条日志`;
    case 'none':
      return '不自动清理日志';
    case 'days':
    default:
      return `按天保留最近 ${value} 天日志`;
  }
}

function parseRgbColor(value: string) {
  const matched = value.match(
    /rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*[\d.]+)?\)/,
  );
  expect(matched, `Expected an rgb/rgba color, got: ${value}`).toBeTruthy();
  return [matched![1], matched![2], matched![3]].map((item) =>
    Number.parseInt(item, 10),
  ) as [number, number, number];
}

function relativeLuminance([red, green, blue]: [number, number, number]) {
  const normalize = (channel: number) => {
    const value = channel / 255;
    return value <= 0.039_28
      ? value / 12.92
      : ((value + 0.055) / 1.055) ** 2.4;
  };
  const r = normalize(red);
  const g = normalize(green);
  const b = normalize(blue);
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrastRatio(foreground: string, background: string) {
  const foregroundLuminance = relativeLuminance(parseRgbColor(foreground));
  const backgroundLuminance = relativeLuminance(parseRgbColor(background));
  const lighter = Math.max(foregroundLuminance, backgroundLuminance);
  const darker = Math.min(foregroundLuminance, backgroundLuminance);
  return (lighter + 0.05) / (darker + 0.05);
}

test.describe('TC-97 定时任务导航与帮助文案', () => {
  const jobName = `e2e_job_navigation_copy_${Date.now()}`;

  let api: APIRequestContext;
  let jobId = 0;
  let originalShellSwitch: { id: number; value: string } | null = null;
  let originalThemeMode: { id: number; value: string } | null = null;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    originalShellSwitch = await getConfigByKey(api, 'cron.shell.enabled');
    originalThemeMode = await getConfigByKey(api, 'sys.ui.theme.mode');
    await setCronShellEnabled(api, true);
    await updateConfigValue(api, originalThemeMode.id, 'dark');

    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(
      api,
      buildShellJobPayload({
        concurrency: 'parallel',
        env: {},
        groupId: defaultGroup.id,
        maxConcurrency: 2,
        name: jobName,
        scope: 'all_node',
        shellCmd: "printf 'navigation help copy'",
        status: 'disabled',
      }),
    );
    jobId = created.id;
  });

  test.afterAll(async () => {
    if (jobId) {
      await deleteJob(api, jobId);
    }
    if (originalShellSwitch) {
      await updateConfigValue(
        api,
        originalShellSwitch.id,
        originalShellSwitch.value,
      );
    }
    if (originalThemeMode) {
      await updateConfigValue(
        api,
        originalThemeMode.id,
        originalThemeMode.value,
      );
    }
    await api.dispose();
  });

  test('TC-97a~m: 菜单分组、深色主题代码化表达式、帮助提示、表单校验、Shell 提示间距与停用任务立即执行应清晰可理解', async ({
    adminPage,
  }) => {
    const accessibleMenus = await getAccessibleMenus(api);
    const jobCatalog = findMenuByTitle(accessibleMenus.list, '任务调度');

    expect(jobCatalog).toBeTruthy();
    expect(
      (jobCatalog?.children ?? []).map((item) => item.meta?.title),
    ).toEqual(['任务管理', '分组管理', '执行日志']);

    await adminPage.goto('/system/scheduled-job');
    await adminPage.waitForLoadState('networkidle');
    await expect(adminPage).toHaveURL(/\/system\/scheduled-job(?:\/)?$/);

    const jobPage = new JobPage(adminPage);
    await jobPage.goto();
    await expect(adminPage).toHaveURL(/\/system\/job(?:\/)?$/);
    await expect
      .poll(async () =>
        adminPage.evaluate(() =>
          document.documentElement.classList.contains('dark'),
        ),
      )
      .toBe(true);

    await jobPage.fillSearchKeyword(jobName);
    await jobPage.clickSearch();
    await expect(await jobPage.hasSearchedJobMoreButton()).toBe(true);

    const rowText = await jobPage.getJobRowText(jobName);
    expect(rowText).toContain('所有节点执行');
    expect(rowText).toContain('允许并行执行');

    const cronDisplay = await jobPage.getCronDisplayMetrics(jobId);
    expect(cronDisplay.text).toBe('0 0 1 1 *');
    expect(cronDisplay.fieldCount).toBeLessThanOrEqual(1);
    expect(cronDisplay.fontFamily.toLowerCase()).toContain('monospace');
    expect(cronDisplay.backgroundColor).not.toBe('rgba(0, 0, 0, 0)');
    expect(cronDisplay.borderColor).not.toBe('rgba(0, 0, 0, 0)');
    expect(
      contrastRatio(cronDisplay.color, cronDisplay.backgroundColor),
      `Cron expression contrast should stay readable in dark theme, got color=${cronDisplay.color}, background=${cronDisplay.backgroundColor}`,
    ).toBeGreaterThanOrEqual(4.5);

    await jobPage.triggerSearchedJob();
    await expect
      .poll(
        async () => {
          const logs = await listLogs(api, jobId);
          return logs.total;
        },
        {
          timeout: 30000,
          message: '停用状态任务也应支持立即执行并生成日志',
        },
      )
      .toBeGreaterThan(0);

    const publicConfig = await expectSuccess<{
      cron: {
        logRetention: {
          mode: string;
          value: number;
        };
      };
    }>(await api.get('config/public/frontend'));
    const retentionSummary = formatRetentionSummary(
      publicConfig.cron.logRetention.mode,
      publicConfig.cron.logRetention.value,
    );

    await jobPage.openCreate();
    await expect(
      adminPage.getByRole('tab', { name: 'Handler', exact: true }),
    ).toHaveCount(0);
    await expect(
      adminPage.getByRole('tab', { name: 'Shell', exact: true }),
    ).toHaveCount(0);
    await expect(
      adminPage.getByLabel('任务处理器', { exact: true }),
    ).toHaveCount(0);

    await adminPage
      .getByTestId('job-form-shell')
      .locator('textarea')
      .first()
      .fill("printf 'validation shell'");

    await jobPage.hoverFieldHelp('定时表达式');
    await expect(await jobPage.isTooltipVisible('支持 5 段或 6 段 Cron')).toBe(
      true,
    );

    await jobPage.hoverFieldHelp('调度范围');
    await expect(
      await jobPage.isTooltipVisible('仅主节点执行：只有当前主节点会执行'),
    ).toBe(true);

    await jobPage.hoverFieldHelp('并发策略');
    await expect(
      await jobPage.isTooltipVisible('单例执行：本节点已有实例运行时'),
    ).toBe(true);

    await jobPage.hoverFieldHelp('日志保留');
    await expect(
      await jobPage.isTooltipVisible(`当前系统策略：${retentionSummary}`),
    ).toBe(true);

    await jobPage.hoverFieldHelp('超时时间（秒）');
    await expect(
      await jobPage.isTooltipVisible('任务实例单次运行允许的最长时长'),
    ).toBe(true);

    await jobPage.hoverFieldHelp('最大执行次数');
    await expect(
      await jobPage.isTooltipVisible('设置为 0 表示不限制执行次数'),
    ).toBe(true);

    const cronEditor = await jobPage.getCronEditorMetrics();
    expect(cronEditor.fontFamily.toLowerCase()).toContain('monospace');
    expect(cronEditor.backgroundColor).not.toBe('rgba(0, 0, 0, 0)');
    expect(cronEditor.borderRadius).not.toBe('0px');

    await jobPage.fillCommonFields({
      cronExpr: '* * * *',
      name: `${jobName}_invalid_cron`,
    });
    await jobPage.save();
    await expect(
      jobPage.messageNotice('Cron 表达式必须为 5 段或 6 段'),
    ).toBeVisible();

    await jobPage.fillCommonFields({
      cronExpr: '17 3 * * *',
      timezone: 'Mars/Phobos',
    });
    await jobPage.save();
    await expect(jobPage.messageNotice('请输入有效的时区')).toBeVisible();

    const shellWarningPadding = await jobPage.getElementVerticalPadding(
      'job-shell-warning-alert',
    );
    expect(shellWarningPadding.paddingTop).toBeGreaterThanOrEqual(5);
    expect(shellWarningPadding.paddingBottom).toBeGreaterThanOrEqual(5);

    await jobPage.closeDialog();

    await jobPage.fillSearchKeyword('任务日志清理');
    await jobPage.clickSearch();
    await expect(await jobPage.hasSearchedJobMoreButton()).toBe(false);
    await jobPage.openEditSearchedJob();

    const commonLockPadding = await jobPage.getElementVerticalPadding(
      'job-builtin-common-lock-alert',
    );
    expect(commonLockPadding.paddingTop).toBeGreaterThanOrEqual(5);
    expect(commonLockPadding.paddingBottom).toBeGreaterThanOrEqual(5);

    await expect(adminPage.getByTestId('job-builtin-detail-card')).toBeVisible();
    await expect(
      adminPage.getByRole('button', { name: /确\s*认/ }),
    ).toHaveCount(0);
  });
});
