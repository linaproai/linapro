import { test, expect } from '../../fixtures/auth';
import { NoticePage } from '../../pages/NoticePage';

test.describe('TC0037 通知公告 CRUD', () => {
  const testTitle = `测试通知_${Date.now()}`;
  const testTitleRenamed = `${testTitle}_修改`;

  test('TC0037a: 创建新通知公告', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();
    await noticePage.createNotice(testTitle, '通知', '草稿', '这是测试内容');

    await expect(
      adminPage.getByText(/新增成功|创建成功|success/i),
    ).toBeVisible({ timeout: 5000 });
  });

  test('TC0037b: 通知公告列表中可见新创建的记录', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();

    const hasNotice = await noticePage.hasNotice(testTitle);
    expect(hasNotice).toBeTruthy();
  });

  test('TC0037c: 编辑通知公告', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();
    await noticePage.editNotice(testTitle, testTitleRenamed);

    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0037d: 删除通知公告', async ({ adminPage }) => {
    const noticePage = new NoticePage(adminPage);
    await noticePage.goto();
    await noticePage.deleteNotice(testTitleRenamed);

    await expect(adminPage.getByText(/删除成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });
});
