import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';

test.describe('TC0059 字典类型级联删除', () => {
  const testTypeName = `级联测试_${Date.now()}`;
  const testTypeCode = `cascade_test_${Date.now()}`;
  const testDataLabel1 = `选项A_${Date.now()}`;
  const testDataValue1 = 'option_a';
  const testDataLabel2 = `选项B_${Date.now()}`;
  const testDataValue2 = 'option_b';

  test('TC0059a: 删除字典类型时显示级联删除提示', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Create a new dict type for testing
    await dictPage.createType(testTypeName, testTypeCode);
    await adminPage.waitForTimeout(500);

    // Create dict data for this type
    await dictPage.clickTypeRow(testTypeName);
    await dictPage.createData(testDataLabel1, testDataValue1);
    await adminPage.waitForTimeout(500);

    // Search for the type
    await dictPage.fillTypeSearchField('字典名称', testTypeName);
    await dictPage.clickTypeSearch();

    // Click delete button
    await dictPage.typePanel.locator('.ant-btn-sm').filter({ hasText: /删\s*除/ }).first().click();

    // Wait for confirmation modal
    const modal = adminPage.locator('.ant-modal-confirm');
    await modal.waitFor({ state: 'visible', timeout: 3000 });

    // Verify the cascade delete warning message
    await expect(modal.getByText(/同时删除.*字典数据/)).toBeVisible();

    // Click cancel to close (cleanup will happen in afterAll)
    await modal.getByRole('button', { name: /取\s*消/ }).click();
    await modal.waitFor({ state: 'hidden', timeout: 3000 });
  });

  test('TC0059b: 删除字典类型时级联删除关联的字典数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Create a new dict type for testing
    const typeName = `级联删除_${Date.now()}`;
    const typeCode = `cascade_del_${Date.now()}`;
    await dictPage.createType(typeName, typeCode);
    await adminPage.waitForTimeout(500);

    // Create dict data for this type
    await dictPage.clickTypeRow(typeName);
    const dataLabelA = `数据A_${Date.now()}`;
    await dictPage.createData(dataLabelA, 'data_a');
    await adminPage.waitForTimeout(500);

    // Verify data exists
    await dictPage.fillDataSearchField('字典标签', dataLabelA.split('_')[0]);
    await dictPage.clickDataSearch();
    expect(await dictPage.hasData(dataLabelA.split('_')[0])).toBeTruthy();

    // Now delete the type - search first
    await dictPage.fillTypeSearchField('字典名称', typeName);
    await dictPage.clickTypeSearch();
    await adminPage.waitForTimeout(500);

    // Click single row delete button (ghost button in action column)
    await dictPage.typePanel.locator('.ant-btn-sm').filter({ hasText: /删\s*除/ }).first().click();

    // Wait for modal and verify cascade delete warning
    const modal = adminPage.locator('.ant-modal-confirm');
    await modal.waitFor({ state: 'visible', timeout: 3000 });
    await expect(modal.getByText(/同时删除.*字典数据/)).toBeVisible();

    // Set up promise to wait for the delete API call
    const deletePromise = adminPage.waitForResponse(
      resp => resp.url().includes('/dict/type/') && resp.request().method() === 'DELETE',
      { timeout: 10000 }
    );

    // Click confirm button
    await modal.getByRole('button', { name: /确\s*定/ }).click();

    // Wait for delete API to complete
    const response = await deletePromise;
    console.log('Delete API response status:', response.status());

    // Wait for modal to close
    await modal.waitFor({ state: 'hidden', timeout: 5000 });

    // Wait for delete success message (filter to avoid strict mode with multiple messages)
    await adminPage.locator('.ant-message-success').filter({ hasText: /删\s*除成功/ }).waitFor({ state: 'visible', timeout: 5000 });

    // Verify the type is deleted
    await dictPage.fillTypeSearchField('字典名称', typeName);
    await dictPage.clickTypeSearch();
    expect(await dictPage.hasType(typeName)).toBeFalsy();

    // Verify the associated data is also deleted
    await dictPage.clickTypeReset();
    await adminPage.waitForTimeout(500);

    // Try to search for the deleted data - it should not exist
    await dictPage.fillDataSearchField('字典标签', dataLabelA.split('_')[0]);
    await dictPage.clickDataSearch();
    expect(await dictPage.hasData(dataLabelA.split('_')[0])).toBeFalsy();
  });

  test('TC0059c: 批量删除字典类型时显示级联删除提示', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Create two dict types
    const typeName1 = `批量删除1_${Date.now()}`;
    const typeCode1 = `batch_del_1_${Date.now()}`;
    const typeName2 = `批量删除2_${Date.now()}`;
    const typeCode2 = `batch_del_2_${Date.now()}`;

    await dictPage.createType(typeName1, typeCode1);
    await adminPage.waitForTimeout(300);
    await dictPage.createType(typeName2, typeCode2);
    await adminPage.waitForTimeout(500);

    // Search and select both types
    await dictPage.fillTypeSearchField('字典名称', '批量删除');
    await dictPage.clickTypeSearch();

    // Wait for search results to appear
    await adminPage.waitForTimeout(500);

    // Select first two rows
    await dictPage.selectTypeRow(0);
    await adminPage.waitForTimeout(200);
    await dictPage.selectTypeRow(1);
    await adminPage.waitForTimeout(300);

    // Click batch delete button - it's the danger button in toolbar (not the small action column buttons)
    // The toolbar button does NOT have the .ant-btn-sm class - use .first() to select the toolbar button
    await adminPage.locator('#dict-type').getByRole('button', { name: /删\s*除/ }).first().click();

    // Wait for confirmation modal
    const modal = adminPage.locator('.ant-modal-confirm');
    await modal.waitFor({ state: 'visible', timeout: 3000 });

    // Verify the cascade delete warning message
    await expect(modal.getByText(/同时删除.*字典数据/)).toBeVisible();

    // Set up promise to wait for the delete API call
    const deletePromise = adminPage.waitForResponse(
      resp => resp.url().includes('/dict/type/') && resp.request().method() === 'DELETE',
      { timeout: 15000 }
    );

    // Click confirm button
    await modal.getByRole('button', { name: /确\s*定/ }).click();

    // Wait for the delete API to complete
    await deletePromise;

    // Verify both types are deleted
    await dictPage.clickTypeReset();
    await dictPage.fillTypeSearchField('字典名称', typeName1);
    await dictPage.clickTypeSearch();
    expect(await dictPage.hasType(typeName1)).toBeFalsy();
  });
});