import { test, expect } from '../../../fixtures/auth';
import { ConfigPage } from '../../../pages/ConfigPage';

/**
 * Regression: horizontal Vben form validation errors must align under the
 * control column, not under the label column (left-0 against FormItem).
 */
test.describe('TC008 参数设置表单校验错误布局', () => {
  test('TC-8a: 空表单提交后校验错误对齐输入控件左侧', async ({
    adminPage,
  }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();
    await configPage.openCreateDialog();

    await configPage.clickDialogConfirm();

    const error = configPage.dialogFieldError('参数名称');
    await expect(error).toBeVisible({ timeout: 5000 });
    // i18n: zh-CN "请输入参数名称" / en-US "Please enter Parameter Name"
    await expect(error).toHaveText(/请输入参数名称|Please enter Parameter Name/i);

    const control = configPage.dialogFieldControl('参数名称');
    await expect(control).toBeVisible();

    const errorBox = await error.boundingBox();
    const controlBox = await control.boundingBox();
    expect(errorBox).toBeTruthy();
    expect(controlBox).toBeTruthy();

    // Bug: error.left pinned to FormItem left (under labels) → much smaller x.
    // Fix: error.left is relative to the control column ≈ control.left.
    const leftDelta = (errorBox!.x ?? 0) - (controlBox!.x ?? 0);
    expect(leftDelta).toBeGreaterThanOrEqual(-8);
    expect(leftDelta).toBeLessThan(24);

    // Project-root temp/ per testing.md E2E screenshot rules.
    const day = new Date().toISOString().slice(0, 10).replace(/-/g, '');
    const stamp = new Date()
      .toISOString()
      .replace(/[:.]/g, '-')
      .slice(0, 19);
    await adminPage.screenshot({
      path: `../../temp/${day}/${stamp}-config-form-validation-error-layout.png`,
      fullPage: false,
    });
  });

  test('TC-8b: 多个必填字段校验错误均不偏到标签列', async ({
    adminPage,
  }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();
    await configPage.openCreateDialog();
    await configPage.clickDialogConfirm();

    for (const label of ['参数名称', '参数键名', '参数键值'] as const) {
      const error = configPage.dialogFieldError(label);
      await expect(error).toBeVisible({ timeout: 5000 });
      const control = configPage.dialogFieldControl(label);
      const errorBox = await error.boundingBox();
      const controlBox = await control.boundingBox();
      expect(errorBox).toBeTruthy();
      expect(controlBox).toBeTruthy();
      const leftDelta = (errorBox!.x ?? 0) - (controlBox!.x ?? 0);
      expect(
        leftDelta,
        `${label} validation error should align under control, delta=${leftDelta}`,
      ).toBeGreaterThanOrEqual(-8);
    }
  });
});
