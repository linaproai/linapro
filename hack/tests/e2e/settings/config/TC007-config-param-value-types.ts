import { test, expect } from '../../../fixtures/auth';
import { ConfigPage } from '../../../pages/ConfigPage';

test.describe('TC007 参数值类型化编辑', () => {
  const selectName = `类型化下拉_${Date.now()}`;
  const selectKey = `test.typed.select.${Date.now()}`;

  test('TC-7a: 编辑忘记密码入口使用布尔组件并可保存', async ({
    adminPage,
  }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();
    await configPage.openEditByKey('sys.auth.forgetPasswordEnabled');

    await expect(
      adminPage.locator('[role="dialog"]').getByRole('radio', { name: /是|True/i }),
    ).toBeVisible();
    await expect(
      adminPage.locator('[role="dialog"]').getByRole('radio', { name: /否|False/i }),
    ).toBeVisible();

    // Toggle to the opposite of current selection by clicking both safely.
    await configPage.chooseBooleanValue('false');
    await configPage.confirmDialog();
    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });

    // Restore enabled state for later suite stability.
    await configPage.openEditByKey('sys.auth.forgetPasswordEnabled');
    await configPage.chooseBooleanValue('true');
    await configPage.confirmDialog();
    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC-7b: 编辑登录框位置通过下拉选择合法布局值', async ({
    adminPage,
  }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();
    await configPage.openEditByKey('sys.auth.loginPanelLayout');

    await configPage.selectDialogOption('参数键值', /左侧|Left|panel-left/i);
    await configPage.confirmDialog();
    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });

    // Restore default right layout.
    await configPage.openEditByKey('sys.auth.loginPanelLayout');
    await configPage.selectDialogOption('参数键值', /右侧|Right|panel-right/i);
    await configPage.confirmDialog();
    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC-7c: 自定义 select 参数创建后编辑可从 options 选择', async ({
    adminPage,
  }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    const optionsText = '选项甲=opt-a\n选项乙=opt-b';
    await configPage.createSelect(
      selectName,
      selectKey,
      optionsText,
      '选项甲',
      'typed select e2e',
    );
    await expect(adminPage.getByText(/创建成功|success/i)).toBeVisible({
      timeout: 5000,
    });

    await configPage.openEditByKey(selectKey);
    await configPage.selectDialogOption('参数键值', /选项乙|opt-b/i);
    await configPage.confirmDialog();
    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });

    await configPage.delete(selectName);
  });
});
