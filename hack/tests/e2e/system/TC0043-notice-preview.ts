import { test, expect } from '../../fixtures/auth';
import { NoticePage } from '../../pages/NoticePage';

test.describe('TC0043 通知公告预览', () => {
  test('TC0043a: 预览 Mock 通知公告内容', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();

    // Preview the first mock notice (系统升级通知)
    await noticePage.previewNotice('系统升级通知');

    // Modal should display the notice content
    const modal = adminPage.locator('[role="dialog"]');
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Should show the notice type in descriptions
    const descArea = modal.locator('.ant-descriptions');
    await expect(descArea).toBeVisible({ timeout: 5000 });

    // Should show the content
    await expect(
      modal.getByText('系统将于本周六凌晨'),
    ).toBeVisible({ timeout: 5000 });
  });
});
