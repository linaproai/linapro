import { expect, test } from '../../fixtures/auth';
import { DeptPage } from '../../pages/DeptPage';

test.describe('TC0025 部门编码字段', () => {
  const suffix = Date.now();
  const deptName = `编码测试部_${suffix}`;
  const deptCode = `code_${suffix}`;
  const deptName2 = `编码测试二_${suffix}`;

  test('TC0025a: 新增部门时填写编码，列表中可见', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    await deptPage.goto();

    await deptPage.createSubDept('Lina科技', deptName, { code: deptCode });
    await expect(adminPage.getByText(/创建成功|success/i)).toBeVisible({
      timeout: 5000,
    });

    // Verify the code is displayed in the table
    const hasCode = await deptPage.hasDeptWithCode(deptName, deptCode);
    expect(hasCode).toBeTruthy();
  });

  test('TC0025b: 编辑部门编码', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    await deptPage.goto();

    const newCode = `edited_${suffix}`;
    await deptPage.editDept(deptName, deptName, { code: newCode });
    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });

    // Verify the updated code is displayed
    const hasNewCode = await deptPage.hasDeptWithCode(deptName, newCode);
    expect(hasNewCode).toBeTruthy();
  });

  test('TC0025c: 重复编码校验', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    await deptPage.goto();

    // Try to create another dept with the same code (edited_${suffix})
    const duplicateCode = `edited_${suffix}`;
    await deptPage.createSubDept('Lina科技', deptName2, {
      code: duplicateCode,
    });

    // Should see error message about duplicate code
    await expect(adminPage.getByText(/部门编码已存在/)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0025d: 清理测试数据', async ({ adminPage }) => {
    const deptPage = new DeptPage(adminPage);
    await deptPage.goto();

    // Delete test depts
    await deptPage.deleteDept(deptName);
    await adminPage.waitForTimeout(1000);

    const hasDept = await deptPage.hasDept(deptName);
    expect(hasDept).toBeFalsy();
  });
});
