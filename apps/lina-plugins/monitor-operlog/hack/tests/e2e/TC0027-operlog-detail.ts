import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { waitForDialogReady, waitForTableReady } from '@host-tests/support/ui';

test.describe('TC0027 操作日志详情查看', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'monitor-operlog');
  });

  test('TC0027a: 点击详情按钮打开详情抽屉', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await waitForTableReady(adminPage);

    // If there are rows, click the first detail button
    const rows = adminPage.locator('.vxe-body--row');
    const rowCount = await rows.count();
    if (rowCount === 0) {
      test.skip(true, 'No operation logs to test');
      return;
    }

    // Click the last matching detail button (which is in the fixed overlay, visible to user)
    const detailBtns = adminPage.getByRole('button', { name: /详\s*情/ });
    const count = await detailBtns.count();
    await detailBtns.nth(count > 1 ? count - Math.ceil(count / 2) : 0).click();

    // Verify detail drawer content is visible
    const drawer = await waitForDialogReady(adminPage.getByLabel('操作日志详情'));
    await expect(drawer.getByText('日志编号')).toBeVisible();
    await expect(drawer.getByText('操作结果').first()).toBeVisible();
    await expect(drawer.getByText('模块名称').first()).toBeVisible();
    await expect(drawer.getByText('操作摘要').first()).toBeVisible();
    await expect(drawer.getByText('操作类型').first()).toBeVisible();
    // "方法"字段应已移除
    await expect(drawer.getByText('方法')).not.toBeVisible();
  });
});
