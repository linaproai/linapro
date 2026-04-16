import { test, expect } from '../../fixtures/auth';
import { NoticePage } from '../../pages/NoticePage';

test.describe('TC0038 通知公告搜索筛选', () => {
  test('TC0038a: 按标题搜索', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();
    await noticePage.fillSearchField('公告标题', '系统升级');
    await noticePage.clickSearch();

    const hasNotice = await noticePage.hasNotice('系统升级通知');
    expect(hasNotice).toBeTruthy();
  });

  test('TC0038b: 按类型筛选', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();

    // Select type filter
    await adminPage.getByLabel('公告类型', { exact: true }).first().click();
    await adminPage.waitForTimeout(300);
    await adminPage.getByTitle('通知').click();
    await noticePage.clickSearch();

    // Results should only contain type=通知
    const total = await noticePage.getTotalCount();
    expect(total).toBeGreaterThan(0);
  });

  test('TC0038c: 重置搜索条件', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();

    const totalBefore = await noticePage.getTotalCount();

    await noticePage.fillSearchField('公告标题', 'NONEXISTENT');
    await noticePage.clickSearch();

    const totalAfter = await noticePage.getTotalCount();
    expect(totalAfter).toBeLessThanOrEqual(totalBefore);

    await noticePage.clickReset();
    const totalReset = await noticePage.getTotalCount();
    expect(totalReset).toBe(totalBefore);
  });
});
