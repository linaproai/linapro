import { Page, Locator, expect } from "@playwright/test";

export class PluginPage {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  get tableTitle(): Locator {
    return this.page.getByText("插件列表").first();
  }

  get dynamicUploadTrigger(): Locator {
    return this.page.getByTestId("plugin-dynamic-upload-trigger").first();
  }

  get dynamicUploadDragger(): Locator {
    return this.page.getByTestId("plugin-dynamic-upload-dragger").first();
  }

  get dynamicOverwriteSwitch(): Locator {
    return this.page.getByTestId("plugin-dynamic-overwrite-switch").first();
  }

  get sidebarMenu(): Locator {
    return this.page.getByRole("menu").first();
  }

  sidebarMenuItem(menuName: string): Locator {
    return this.sidebarMenu.getByText(menuName, { exact: true }).first();
  }

  async clickSidebarMenuItem(menuName: string) {
    await this.expectSidebarMenuVisible(menuName);
    await this.sidebarMenuItem(menuName).click();
  }

  pluginIframeFrame() {
    return this.page.frameLocator("iframe:visible");
  }

  pluginIframe(): Locator {
    return this.page.locator("iframe:visible").first();
  }

  pluginPageRefreshNotice(): Locator {
    return this.page
      .locator(".ant-notification-notice", { hasText: "插件已更新" })
      .last();
  }

  pluginPageRefreshButton(): Locator {
    return this.pluginPageRefreshNotice()
      .getByRole("button", { name: "刷新当前页面" })
      .first();
  }

  pluginDynamicEmbeddedHost(): Locator {
    return this.page.getByTestId("plugin-dynamic-embedded-host").first();
  }

  pluginDemoDynamicTitle(): Locator {
    return this.page
      .getByRole("heading", { name: "动态插件示例已生效" })
      .first();
  }

  pluginDemoDynamicDescription(): Locator {
    return this.page.getByText(
      "该页面来自 plugin-demo-dynamic 的动态挂载入口，用于验证宿主主内容区展示与独立静态页面跳转。",
    );
  }

  pluginDemoDynamicOpenStandaloneButton(): Locator {
    return this.page.getByTestId("plugin-demo-dynamic-open-standalone").first();
  }

  dynamicUploadDialog(): Locator {
    return this.page.getByRole("dialog", { name: "上传插件" }).last();
  }

  dynamicUploadTriggerLabel(): Locator {
    return this.dynamicUploadTrigger.getByText("上传插件", { exact: true });
  }

  dynamicUploadHint(): Locator {
    return this.dynamicUploadDialog().getByText(
      "仅支持单个 .wasm 文件，上传后可在列表中继续安装并启用。",
      { exact: true },
    );
  }

  dynamicOverwriteHint(): Locator {
    return this.dynamicUploadDialog().getByText(
      "允许覆盖同 ID 且未安装的插件工作区文件",
      { exact: true },
    );
  }

  dynamicUploadConfirmButton(): Locator {
    return this.dynamicUploadDialog()
      .getByRole("button", { name: /确\s*认|知\s*道了|知\s*道|ok/i })
      .last();
  }

  dynamicUploadCancelButton(): Locator {
    return this.dynamicUploadDialog()
      .getByRole("button", { name: /取\s*消|cancel/i })
      .last();
  }

  dynamicUploadCloseButton(): Locator {
    return this.dynamicUploadDialog()
      .locator(".ant-modal-close")
      .last();
  }

  uploadSuccessDialog(): Locator {
    return this.dynamicUploadDialog()
      .getByTestId("plugin-dynamic-upload-success")
      .first();
  }

  tableColumn(title: string): Locator {
    return this.page
      .locator(".vxe-table--header .vxe-cell--title", { hasText: title })
      .first();
  }

  pluginRow(pluginId: string): Locator {
    return this.page.locator(".vxe-body--row", { hasText: pluginId }).first();
  }

  pluginInstallButton(pluginId: string): Locator {
    return this.pluginRow(pluginId)
      .getByRole("button", { name: /安\s*装/ })
      .first();
  }

  pluginUninstallButton(pluginId: string): Locator {
    return this.pluginRow(pluginId)
      .getByRole("button", { name: /卸\s*载/ })
      .first();
  }

  hostServiceAuthModal(): Locator {
    return this.page.getByTestId("plugin-host-service-auth-modal").last();
  }

  hostServiceAuthDialog(): Locator {
    return this.page
      .getByRole("dialog", { name: /安装插件并确认权限|启用插件并确认权限/ })
      .last();
  }

  hostServiceAuthCheckbox(
    pluginId: string,
    service: string,
    resourceRef: string,
  ): Locator {
    void pluginId;
    void service;
    return this.hostServiceAuthModal()
      .getByRole("checkbox", { name: resourceRef })
      .first();
  }

  pluginSourceDisabledUninstallTrigger(pluginId: string): Locator {
    return this.page.getByTestId(
      `plugin-source-uninstall-disabled-${pluginId}`,
    );
  }

  pluginEnabledSwitch(pluginId: string): Locator {
    return this.pluginRow(pluginId).locator(".ant-switch").first();
  }

  pluginDescriptionCell(pluginId: string): Locator {
    return this.pluginRow(pluginId)
      .getByTestId(`plugin-description-${pluginId}`)
      .first();
  }

  antTooltip(): Locator {
    return this.page.locator(".ant-tooltip:visible");
  }

  vxeTooltip(): Locator {
    return this.page.locator(".vxe-table--tooltip-wrapper:visible");
  }

  headerActionBeforeSlot(): Locator {
    return this.page.getByText("plugin-demo-source 头部前置扩展").first();
  }

  headerActionAfterSlot(): Locator {
    return this.page.getByText("plugin-demo-source 头部后置扩展").first();
  }

  pluginSidebarSimpleTitle(): Locator {
    return this.page
      .getByRole("heading", { name: "源码插件示例已生效" })
      .first();
  }

  pluginSidebarBriefDescription(): Locator {
    return this.page.getByText(
      "这是一条来自 plugin-demo-source 接口的简要介绍，用于验证源码插件菜单页可读取插件后端数据。",
    );
  }

  workspaceBeforeSlot(): Locator {
    return this.page.getByText(
      "plugin-demo-source 正在通过 `dashboard.workspace.before` 在工作台顶部插入横幅内容。",
    );
  }

  workspaceAfterSlot(): Locator {
    return this.page.getByText("源码插件示例工作台卡片").first();
  }

  crudToolbarSlot(): Locator {
    return this.page.getByText("plugin-demo-source CRUD 扩展").first();
  }

  async gotoManage() {
    await this.page.goto("/system/plugin");
    await expect(this.tableTitle).toBeVisible();
  }

  async searchByPluginId(pluginId: string) {
    const input = this.page.getByRole("textbox", { name: "插件标识" }).first();
    await expect(input).toBeVisible();
    await input.fill(pluginId);
    await this.page.getByRole("button", { name: "搜 索" }).click();
    await expect(this.pluginRow(pluginId)).toBeVisible();
  }

  async syncPlugins() {
    await this.page.getByRole("button", { name: "同步插件" }).click();
    await this.page.waitForLoadState("networkidle");
  }

  async uploadDynamicPlugin(
    filePath: string,
    overwrite = false,
    expectedSuccessText?: string,
  ) {
    await this.dynamicUploadTrigger.click();
    await expect(this.dynamicUploadDialog()).toBeVisible();
    await expect(this.dynamicUploadDragger).toBeVisible();
    if (overwrite) {
      const isChecked =
        (await this.dynamicOverwriteSwitch.getAttribute("aria-checked")) ===
        "true";
      if (!isChecked) {
        await this.dynamicOverwriteSwitch.click();
      }
    }
    const [fileChooser] = await Promise.all([
      this.page.waitForEvent("filechooser"),
      this.dynamicUploadDragger.click(),
    ]);
    await fileChooser.setFiles(filePath);

    // Ant Design Upload updates the modal state asynchronously after the file
    // chooser closes. Waiting for the rendered upload item avoids clicking the
    // confirm button before the file is committed into the reactive file list.
    await expect(
      this.dynamicUploadDialog().locator(".ant-upload-list-item"),
    ).toBeVisible();
    await this.page.waitForTimeout(1500);

    const uploadResponsePromise = this.page.waitForResponse(
      (response) =>
        response.url().includes("/plugins/dynamic/package") &&
        response.request().method() === "POST",
      { timeout: 30000 },
    );

    await this.dynamicUploadConfirmButton().click();

    const uploadResponse = await uploadResponsePromise;
    expect(uploadResponse.status()).toBe(200);

    await expect(this.uploadSuccessDialog()).toBeVisible();
    await expect(this.uploadSuccessDialog()).toContainText(
      expectedSuccessText ?? "上传成功，请在插件列表中继续安装并启用。",
    );
    await expect(this.dynamicUploadConfirmButton()).toContainText("知道了");
    await expect(this.dynamicUploadCancelButton()).toHaveCount(0);
    await expect(this.dynamicUploadCloseButton()).toHaveCount(0);
    await this.dynamicUploadConfirmButton().click();
    await expect(this.dynamicUploadDialog()).not.toBeVisible();

    // The Vite dev server keeps HMR-related requests alive, so waiting for
    // `networkidle` here can hang even after the upload flow already finished.
    // Use stable UI signals instead of transport-level idleness.
    await expect(this.dynamicUploadTrigger).toBeVisible();
    await expect(this.tableTitle).toBeVisible();
  }

  async installPlugin(pluginId: string) {
    const row = this.pluginRow(pluginId);
    await expect(row).toBeVisible();
    await this.page.getByRole("button", { name: /安\s*装/ }).last().click();
    const confirmPopover = this.page.locator(".ant-popover:visible").last();
    await expect(confirmPopover).toBeVisible();
    await confirmPopover
      .getByRole("button", { name: /确\s*定|确\s*认/i })
      .click();
    await expect(await this.pluginActionButton(pluginId, /卸\s*载/)).toBeVisible();
  }

  async openInstallAuthorization(pluginId: string) {
    const row = this.pluginRow(pluginId);
    await expect(row).toBeVisible();
    await this.page.getByRole("button", { name: /安\s*装/ }).last().click();
    const confirmPopover = this.page.locator(".ant-popover:visible").last();
    await expect(confirmPopover).toBeVisible();
    await confirmPopover
      .getByRole("button", { name: /确\s*定|确\s*认/i })
      .click();
    await expect(this.hostServiceAuthModal()).toBeVisible();
  }

  async uninstallPlugin(pluginId: string) {
    const row = this.pluginRow(pluginId);
    await expect(row).toBeVisible();
    await this.page.getByRole("button", { name: /卸\s*载/ }).last().click();
    const confirmPopover = this.page.locator(".ant-popover:visible").last();
    await expect(confirmPopover).toBeVisible();
    await confirmPopover
      .getByRole("button", { name: /确\s*定|确\s*认/i })
      .click();
    await expect(await this.pluginActionButton(pluginId, /安\s*装/)).toBeVisible();
  }

  async setPluginEnabled(pluginId: string, enabled: boolean) {
    const row = this.pluginRow(pluginId);
    await expect(row).toBeVisible();
    const switcher = row.locator(".ant-switch").first();
    const isChecked = (await switcher.getAttribute("aria-checked")) === "true";
    if (isChecked !== enabled) {
      await switcher.click();
      if (enabled) {
        const authDialogVisible = await this.hostServiceAuthDialog()
          .isVisible({ timeout: 1500 })
          .catch(() => false);
        if (authDialogVisible) {
          await this.confirmHostServiceAuthorization();
        }
      }
      await expect(switcher).toHaveAttribute(
        "aria-checked",
        enabled ? "true" : "false",
      );
      await expect(
        this.page.getByText(enabled ? "插件已启用" : "插件已禁用").last(),
      ).toBeVisible();
    }
  }

  async openEnableAuthorization(pluginId: string) {
    const switcher = this.pluginEnabledSwitch(pluginId);
    await expect(switcher).toBeVisible();
    await switcher.click();
    await expect(this.hostServiceAuthModal()).toBeVisible();
  }

  async setHostServiceAuthorization(
    pluginId: string,
    service: string,
    resourceRef: string,
    checked: boolean,
  ) {
    const checkbox = this.hostServiceAuthCheckbox(pluginId, service, resourceRef);
    await expect(checkbox).toBeVisible();
    const isChecked = await checkbox.isChecked();
    if (isChecked !== checked) {
      await checkbox.click();
    }
  }

  async confirmHostServiceAuthorization() {
    await this.hostServiceAuthDialog()
      .getByRole("button", { name: /确\s*认|确\s*定/i })
      .last()
      .click();
    await expect(this.hostServiceAuthDialog()).toHaveCount(0);
  }

  private async pluginActionButton(pluginId: string, name: RegExp) {
    const rows = this.page.locator(".vxe-table--main-wrapper .vxe-body--row");
    const rowCount = await rows.count();
    let rowIndex = -1;

    for (let index = 0; index < rowCount; index++) {
      const row = rows.nth(index);
      const text = (await row.textContent()) ?? "";
      if (text.includes(pluginId)) {
        rowIndex = index;
        break;
      }
    }

    expect(rowIndex, `未找到插件行: ${pluginId}`).toBeGreaterThanOrEqual(0);
    return this.page
      .locator(".vxe-table--fixed-right-wrapper .vxe-body--row")
      .nth(rowIndex)
      .getByRole("button", { name })
      .first();
  }

  async expectSidebarMenuVisible(menuName: string) {
    const menuItem = this.sidebarMenuItem(menuName);
    const visible = await menuItem.isVisible().catch(() => false);
    if (!visible) {
      await this.sidebarMenuItem("插件管理").click();
    }
    await expect(menuItem).toBeVisible();
  }

  async expectSidebarMenuHidden(menuName: string) {
    const visible = await this.sidebarMenu
      .getByText(menuName, { exact: true })
      .first()
      .isVisible({ timeout: 1500 })
      .catch(() => false);
    expect(visible).toBeFalsy();
  }

  async gotoWorkspace() {
    await this.page.goto("/dashboard/workspace");
    await expect(
      this.page.getByText("开始您一天的工作吧！").first(),
    ).toBeVisible();
  }

  async expectWorkspaceSlotHidden() {
    await expect(this.workspaceBeforeSlot()).toHaveCount(0);
    await expect(this.workspaceAfterSlot()).toHaveCount(0);
  }

  async expectHeaderSlotsHidden() {
    await expect(this.headerActionBeforeSlot()).toHaveCount(0);
    await expect(this.headerActionAfterSlot()).toHaveCount(0);
  }

  async expectCrudSlotsHidden() {
    await expect(this.crudToolbarSlot()).toHaveCount(0);
  }

  async expectTableColumnVisible(title: string) {
    await expect(this.tableColumn(title)).toBeVisible();
  }

  async expectTableColumnHidden(title: string) {
    await expect(this.tableColumn(title)).toHaveCount(0);
  }

  async expectTableColumnBetween(
    targetTitle: string,
    previousTitle: string,
    nextTitle: string,
  ) {
    const headerTitles = (
      await this.page
        .locator(".vxe-table--header .vxe-cell--title")
        .allTextContents()
    )
      .map((title) => title.trim())
      .filter(Boolean);

    const targetIndex = headerTitles.indexOf(targetTitle);
    const previousIndex = headerTitles.indexOf(previousTitle);
    const nextIndex = headerTitles.indexOf(nextTitle);

    expect(targetIndex, `未找到列表列: ${targetTitle}`).toBeGreaterThanOrEqual(
      0,
    );
    expect(
      previousIndex,
      `未找到列表列: ${previousTitle}`,
    ).toBeGreaterThanOrEqual(0);
    expect(nextIndex, `未找到列表列: ${nextTitle}`).toBeGreaterThanOrEqual(0);
    expect(
      targetIndex,
      `${targetTitle} 应位于 ${previousTitle} 之后`,
    ).toBeGreaterThan(previousIndex);
    expect(targetIndex, `${targetTitle} 应位于 ${nextTitle} 之前`).toBeLessThan(
      nextIndex,
    );
  }

  async expectDescriptionUsesNativeTooltip(pluginId: string) {
    const descriptionTestId = `plugin-description-${pluginId}`;
    const descriptionCell = this.pluginDescriptionCell(pluginId);
    const descriptionText =
      ((await descriptionCell.textContent()) || "").trim() || "-";
    await expect(descriptionCell).toBeVisible();
    await expect(this.page.getByTestId(descriptionTestId)).toHaveCount(1);
    await expect(descriptionCell).toHaveAttribute("title", descriptionText);
    await descriptionCell.hover();
    await expect(this.vxeTooltip()).toHaveCount(0);
    await expect(this.antTooltip()).toHaveCount(0);
    await this.page.waitForTimeout(5000);
    await expect(this.vxeTooltip()).toHaveCount(0);
    await expect(this.antTooltip()).toHaveCount(0);
    const delayedTitleCount = await this.page
      .locator("[title]")
      .evaluateAll((elements, text) => {
        return elements.filter((element) =>
          (element.getAttribute("title") || "").includes(text),
        ).length;
      }, descriptionText);
    expect(delayedTitleCount, "描述列应只保留单一系统默认提示来源").toBe(1);
  }

  async expectSourcePluginDisabledUninstall(pluginId: string) {
    const uninstallButton = this.pluginSourceDisabledUninstallTrigger(pluginId);
    const tooltipText =
      "源码插件不支持页面动态卸载，如需移除请在源码中取消注册后重新构建宿主。";

    const hasVisibleDisabledButton = await uninstallButton.evaluateAll(
      (elements, expectedTitle) => {
        return elements.some((element) => {
          if (!(element instanceof HTMLButtonElement)) {
            return false;
          }
          const style = window.getComputedStyle(element);
          const rect = element.getBoundingClientRect();
          const isVisible =
            style.display !== "none" &&
            style.visibility !== "hidden" &&
            rect.width > 0 &&
            rect.height > 0;
          return (
            isVisible &&
            element.disabled &&
            element.getAttribute("title") === expectedTitle
          );
        });
      },
      tooltipText,
    );

    expect(
      hasVisibleDisabledButton,
      "源码插件应显示一个可见的灰态卸载按钮，并携带动态卸载提示",
    ).toBeTruthy();
  }

  async openSidebarExampleFromMenu() {
    await this.clickSidebarMenuItem("源码插件示例");
    await expect(this.pluginSidebarSimpleTitle()).toBeVisible();
    await expect(this.pluginSidebarBriefDescription()).toBeVisible();
  }
}
