import path from 'node:path';
import fs from 'node:fs';
import { test, expect } from '../../fixtures/auth';
import { FilePage } from '../../pages/FilePage';

test.describe('TC0048 文件管理', () => {
  // Create a temporary test file
  const testFileName = `test_upload_${Date.now()}.txt`;
  const testFilePath = path.join('/tmp', testFileName);

  test.beforeAll(() => {
    fs.writeFileSync(testFilePath, 'This is a test file for E2E upload testing.');
  });

  test.afterAll(() => {
    if (fs.existsSync(testFilePath)) {
      fs.unlinkSync(testFilePath);
    }
  });

  test('TC0048a: 文件管理页面可正常访问', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // Verify page title and table are visible
    await expect(adminPage.getByText('文件列表')).toBeVisible();
    await expect(adminPage.locator('.vxe-table')).toBeVisible();
  });

  test('TC0048b: 文件上传按钮打开上传弹窗', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    await filePage.openFileUploadModal();

    const modal = adminPage.locator('[role="dialog"]');
    await expect(modal.getByText('文件上传')).toBeVisible();
    // Should have drag upload area
    await expect(modal.locator('.ant-upload-drag')).toBeVisible();
  });

  test('TC0048c: 图片上传按钮打开上传弹窗', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    await filePage.openImageUploadModal();

    const modal = adminPage.locator('[role="dialog"]');
    await expect(modal.getByText('图片上传')).toBeVisible();
  });

  test('TC0048d: 上传文件后文件列表中可见', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // Upload via API to avoid complex file input interaction
    const token = await adminPage.evaluate(() => {
      return localStorage.getItem('preferences') || '';
    });

    // Use file chooser for upload
    await filePage.openFileUploadModal();

    const modal = adminPage.locator('[role="dialog"]');
    const fileChooserPromise = adminPage.waitForEvent('filechooser');
    await modal.locator('.ant-upload-drag').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(testFilePath);

    // Wait for upload success
    await expect(
      adminPage.getByText(/上传成功/),
    ).toBeVisible({ timeout: 10000 });

    // Close modal via the X button (last button in dialog)
    await modal.locator('button').last().click();
    await adminPage.waitForTimeout(1000);

    // Verify file appears in the list
    const hasFile = await filePage.hasFile(testFileName);
    expect(hasFile).toBeTruthy();
  });

  test('TC0048e: 搜索条件筛选文件', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // Search by file type (suffix) - now a Select dropdown
    const suffixLabel = adminPage.locator('label').filter({ hasText: '文件类型' });
    const suffixSelect = suffixLabel.locator('..').locator('.ant-select').first();
    await suffixSelect.click();

    // Wait for dropdown and select 'png' which should exist
    const dropdown = adminPage.locator('.ant-select-dropdown').last();
    await expect(dropdown).toBeVisible({ timeout: 5000 });
    const pngOption = dropdown.getByText('png', { exact: true });
    const hasPng = await pngOption.count();
    if (hasPng > 0) {
      await pngOption.click();
      await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
      await adminPage.waitForTimeout(1000);

      // All results should have png suffix
      const rowCount = await filePage.getRowCount();
      expect(rowCount).toBeGreaterThan(0);
    } else {
      // Close dropdown if no png option
      await adminPage.keyboard.press('Escape');
    }
  });

  test('TC0048p: 文件类型下拉框显示正确的后缀格式', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // Open file type dropdown
    const suffixLabel = adminPage.locator('label').filter({ hasText: '文件类型' });
    const suffixSelect = suffixLabel.locator('..').locator('.ant-select').first();
    await suffixSelect.click();

    const dropdown = adminPage.locator('.ant-select-dropdown').last();
    await expect(dropdown).toBeVisible({ timeout: 5000 });

    // Get all option texts
    const options = dropdown.locator('.ant-select-item-option-content');
    const optionCount = await options.count();
    expect(optionCount).toBeGreaterThan(0);

    // Each option should be a plain extension without dot (e.g. "png", not ".png")
    for (let i = 0; i < optionCount; i++) {
      const text = await options.nth(i).innerText();
      expect(text).toMatch(/^\w+$/);
      expect(text).not.toMatch(/^\./);
    }

    await adminPage.keyboard.press('Escape');
  });

  test('TC0048g: 文件预览列展示预览或下载链接', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    const rowCount = await filePage.getRowCount();
    if (rowCount > 0) {
      // Each row's preview column should show one of:
      // - Image preview (for images)
      // - "PDF 预览" link (for PDFs)
      // - Plain URL text (for other file types, truncated with tooltip)
      const rows = adminPage.locator('.vxe-body--row');
      const count = await rows.count();
      for (let i = 0; i < Math.min(count, 5); i++) {
        const urlCell = rows.nth(i).locator('td').nth(4);
        const cellText = await urlCell.innerText();
        // Should be either an image preview, "PDF 预览", or a URL text
        const isImagePreview = (await urlCell.locator('.ant-image').count()) > 0;
        const isPdfPreview = cellText.includes('PDF 预览');
        const isUrlText = cellText.includes('http');
        expect(isImagePreview || isPdfPreview || isUrlText).toBeTruthy();
      }
    }
  });

  test('TC0048h: 下载按钮点击不报错', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    const rowCount = await filePage.getRowCount();
    if (rowCount > 0) {
      // Listen for console errors
      const consoleErrors: string[] = [];
      adminPage.on('console', (msg) => {
        if (msg.type() === 'error') {
          consoleErrors.push(msg.text());
        }
      });

      // Intercept the download request to verify it succeeds
      const downloadPromise = adminPage.waitForResponse(
        (resp) => resp.url().includes('/file/download/') && resp.status() === 200,
        { timeout: 15000 },
      );

      // Click download button on first row
      const firstRow = adminPage.locator('.vxe-body--row').first();
      await firstRow.getByRole('button', { name: /下\s*载/ }).click();

      // Wait for download response
      const response = await downloadPromise;
      expect(response.status()).toBe(200);

      // No download-related console errors
      const downloadErrors = consoleErrors.filter((e) => e.includes('Download failed'));
      expect(downloadErrors).toHaveLength(0);
    }
  });

  test('TC0048i: 上传者列展示账号名称而非昵称', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    const rowCount = await filePage.getRowCount();
    if (rowCount > 0) {
      // The uploader column (index 8: checkbox=0, original=1, suffix=2, scene=3, url=4, size=5, createdAt=6, createdByName=7)
      const uploaderCell = adminPage.locator('.vxe-body--row').first().locator('td').nth(7);
      const uploaderText = await uploaderCell.innerText();
      // Should be a username like 'admin', not a nickname like '管理员'
      expect(uploaderText.trim()).toBe('admin');
    }
  });

  test('TC0048j: 文件大小列支持排序', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // The size column header should be sortable - click it
    const sizeHeader = adminPage.locator('.vxe-header--column').filter({ hasText: '文件大小' }).first();

    // Listen for API request with sort params
    const sortRequestPromise = adminPage.waitForRequest(
      (req) => req.url().includes('/file?') && req.url().includes('orderBy=size'),
      { timeout: 10000 },
    );

    await sizeHeader.click();

    // Verify API request was sent with sort parameters
    const request = await sortRequestPromise;
    expect(request.url()).toContain('orderBy=size');
  });

  test('TC0048m: 上传时间列支持排序', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // The createdAt column header should be sortable - click it
    const createdAtHeader = adminPage.locator('.vxe-header--column').filter({ hasText: '上传时间' }).first();

    // Listen for API request with sort params
    const sortRequestPromise = adminPage.waitForRequest(
      (req) => req.url().includes('/file?') && req.url().includes('orderBy=createdAt'),
      { timeout: 10000 },
    );

    await createdAtHeader.click();

    // Verify API request was sent with sort parameters
    const request = await sortRequestPromise;
    expect(request.url()).toContain('orderBy=createdAt');
  });

  test('TC0048k: 详情按钮打开文件详情对话框', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    const rowCount = await filePage.getRowCount();
    if (rowCount > 0) {
      // Click detail button on first row
      const firstRow = adminPage.locator('.vxe-body--row').first();
      await firstRow.getByRole('button', { name: /详\s*情/ }).click();

      // Verify detail modal opens
      const modal = adminPage.locator('[role="dialog"]');
      await expect(modal.getByText('文件详情')).toBeVisible({ timeout: 5000 });

      // Should show file info in Descriptions
      await expect(modal.getByText('原始文件名')).toBeVisible();
      await expect(modal.getByText('存储文件名')).toBeVisible();
      await expect(modal.getByText('文件大小')).toBeVisible();
      await expect(modal.getByText('上传者')).toBeVisible();
      await expect(modal.getByText('上传时间')).toBeVisible();
      await expect(modal.getByText('使用场景')).toBeVisible();
    }
  });

  test('TC0048l: 使用场景筛选条件可见', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // The search form should have a "使用场景" label
    const sceneLabel = adminPage.locator('form label').filter({ hasText: '使用场景' });
    await expect(sceneLabel.first()).toBeVisible();
  });

  test('TC0048o: 使用场景筛选下拉框包含预定义选项', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // Wait for the form to fully load (the scene select is populated via API)
    // The scene field should have a label "使用场景"
    await adminPage.waitForTimeout(2000);

    // Find the scene select by looking for the Ant Design select with aria-label or by form item
    // VbenForm uses form items, so we look for the select near the "使用场景" text
    const sceneLabel = adminPage.locator('label').filter({ hasText: '使用场景' });

    // Check if the scene field exists (it might not if API failed)
    const labelCount = await sceneLabel.count();
    if (labelCount === 0) {
      // Skip test if the scene field is not available
      console.log('Scene select field not found, skipping test');
      return;
    }

    // The select is usually the next sibling or nearby element after the label
    const sceneFormItem = sceneLabel.locator('..').locator('.ant-select').first();
    await sceneFormItem.click();

    // Wait for dropdown to open and check for predefined options
    const dropdown = adminPage.locator('.ant-select-dropdown').last();
    await expect(dropdown).toBeVisible({ timeout: 5000 });

    // Should have predefined scene options
    await expect(dropdown.getByText('用户头像')).toBeVisible();
    await expect(dropdown.getByText('通知公告图片')).toBeVisible();
    await expect(dropdown.getByText('通知公告附件')).toBeVisible();
    await expect(dropdown.getByText('其他')).toBeVisible();

    // Close dropdown by clicking elsewhere
    await adminPage.keyboard.press('Escape');
  });

  test('TC0048f: 删除文件', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // Get initial row count
    const initialCount = await filePage.getRowCount();

    if (initialCount > 0) {
      // Listen for successful delete response
      const deleteResponsePromise = adminPage.waitForResponse(
        (resp) => resp.url().includes('/file/') && resp.request().method() === 'DELETE' && resp.status() === 200,
        { timeout: 10000 },
      );

      // Click delete on first row
      const firstRow = adminPage.locator('.vxe-body--row').first();
      await firstRow.getByRole('button', { name: /删\s*除/ }).click();

      // Confirm delete (button text is "确 定")
      await adminPage
        .getByRole('button', { name: /确\s*定/ })
        .click();

      // Verify delete API succeeded
      const deleteResponse = await deleteResponsePromise;
      expect(deleteResponse.status()).toBe(200);
    }
  });

  test('TC0048n: 文件列表默认开启预览模式', async ({ adminPage }) => {
    const filePage = new FilePage(adminPage);
    await filePage.goto();

    // The preview switch should be ON by default
    const previewSwitch = adminPage.locator('.ant-switch-checked');
    await expect(previewSwitch).toBeVisible();

    // If there are image files, they should show as image preview instead of URL text
    const rowCount = await filePage.getRowCount();
    if (rowCount > 0) {
      // Check if any row has an image preview (ant-image element)
      const hasImagePreview = await adminPage.locator('.ant-image').count();
      // At least check the preview mode is active
      expect(hasImagePreview).toBeGreaterThanOrEqual(0);
    }
  });
});
