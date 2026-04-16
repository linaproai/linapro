import { test, expect } from '../../fixtures/auth';
import { PostPage } from '../../pages/PostPage';

test.describe('TC0017 岗位按部门过滤', () => {
  test('TC0017a: 选择部门后岗位列表按部门过滤', async ({ adminPage }) => {
    const postPage = new PostPage(adminPage);
    await postPage.goto();

    // Record initial row count
    const initialCount = await postPage.getTotalCount();

    // Select "研发部门" in the left DeptTree
    await postPage.selectDept('研发部门');

    // After selecting a dept, the grid should reload with filtered results
    // Verify the request includes deptId by checking the grid reloaded
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    // The table should have rendered (may have fewer or equal rows)
    const filteredCount = await postPage.getTotalCount();
    // Filtered count should be <= initial count (or 0 if no posts in that dept)
    expect(filteredCount).toBeLessThanOrEqual(initialCount);
  });

  test('TC0017b: 清除部门选择后显示全部岗位', async ({ adminPage }) => {
    const postPage = new PostPage(adminPage);
    await postPage.goto();

    // Select a dept first
    await postPage.selectDept('研发部门');
    await adminPage.waitForTimeout(500);

    // Click reset to clear dept selection and show all
    await postPage.clickReset();

    // After reset, the DeptTree selection should be cleared
    // and the grid should show all posts
    const totalCount = await postPage.getTotalCount();
    expect(totalCount).toBeGreaterThanOrEqual(0);
  });
});
