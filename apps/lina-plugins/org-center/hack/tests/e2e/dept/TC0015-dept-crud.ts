import { expect, test } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { DeptPage } from '../../pages/DeptPage';

test.describe('TC0015 部门管理 CRUD', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'org-center');
  });

  const suffix = Date.now();
  const testDeptName = `测试部门_${suffix}`;
  const subDeptName = `子部门A_${suffix}`;
  const subDeptRenamed = `子部门B_${suffix}`;

  test('TC0015a: 在根部门下创建子部门', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    await deptPage.goto();
    // 在已有的根组织下创建子部门
    await deptPage.createSubDept('LinaPro.AI', testDeptName);

    await expect(adminPage.getByText(/创建成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0015b: 新创建的部门在列表中可见', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    await deptPage.goto();

    const hasDept = await deptPage.hasDept(testDeptName);
    expect(hasDept).toBeTruthy();
  });

  test('TC0015c: 创建子部门', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    await deptPage.goto();
    await deptPage.createSubDept(testDeptName, subDeptName);

    await expect(adminPage.getByText(/创建成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0015d: 编辑部门', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    await deptPage.goto();
    await deptPage.editDept(subDeptName, subDeptRenamed);

    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0015e: 删除子部门后删除父部门', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    await deptPage.goto();

    // 先删子部门
    await deptPage.deleteDept(subDeptRenamed);

    // 再删父部门
    await deptPage.deleteDept(testDeptName);

    const hasDept = await deptPage.hasDept(testDeptName);
    expect(hasDept).toBeFalsy();
  });
});
