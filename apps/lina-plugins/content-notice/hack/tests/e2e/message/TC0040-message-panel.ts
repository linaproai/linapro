import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';

test.describe('TC0040 消息面板操作', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'content-notice');
  });

  test('TC0040a: 铃铛图标可见', async ({ adminPage }) => {
    // The notification bell should be visible in the header
    const bell = adminPage.locator('.bell-button').first();
    await expect(bell).toBeVisible({ timeout: 5000 });
  });

  test('TC0040b: 点击铃铛显示消息面板', async ({ adminPage }) => {
    const bell = adminPage.locator('.bell-button').first();
    await expect(bell).toBeVisible({ timeout: 5000 });
    await bell.click();

    const popover = adminPage.locator('.side-content:visible').last();
    await expect(popover).toBeVisible({ timeout: 5000 });
    await expect(popover.getByText('通知', { exact: true }).first()).toBeVisible();
    await expect(
      popover.getByRole('button', { name: /查看所有消息/ }),
    ).toBeVisible();
  });
});
