import { Page, Locator, expect } from "@playwright/test";

import { waitForUploadReady } from "../support/ui";

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

  pluginDemoDynamicRecordGrid(): Locator {
    return this.page.getByTestId("plugin-demo-dynamic-record-grid").first();
  }

  pluginDemoDynamicRecordAddButton(): Locator {
    return this.page.getByTestId("plugin-demo-dynamic-record-add").first();
  }

  // Pagination locators keep the runtime demo list assertions readable across
  // the pagination regression scenarios.
  pluginDemoDynamicRecordPagination(): Locator {
    return this.page
      .getByTestId("plugin-demo-dynamic-record-pagination")
      .first();
  }

  pluginDemoDynamicPaginationSummary(): Locator {
    return this.page
      .getByTestId("plugin-demo-dynamic-pagination-summary")
      .first();
  }

  pluginDemoDynamicPaginationPage(pageNumber: number): Locator {
    return this.page
      .getByTestId(`plugin-demo-dynamic-pagination-page-${pageNumber}`)
      .first();
  }

  pluginDemoDynamicPaginationPrevButton(): Locator {
    return this.page
      .getByTestId("plugin-demo-dynamic-pagination-prev")
      .first();
  }

  pluginDemoDynamicPaginationNextButton(): Locator {
    return this.page
      .getByTestId("plugin-demo-dynamic-pagination-next")
      .first();
  }

  pluginDemoDynamicRecordModal(): Locator {
    return this.page.getByTestId("plugin-demo-dynamic-record-modal").last();
  }

  pluginDemoDynamicRecordTitleInput(): Locator {
    return this.page
      .getByTestId("plugin-demo-dynamic-record-title-input")
      .last();
  }

  pluginDemoDynamicRecordContentInput(): Locator {
    return this.page
      .getByTestId("plugin-demo-dynamic-record-content-input")
      .last();
  }

  pluginDemoDynamicRecordFileInput(): Locator {
    return this.page
      .getByTestId("plugin-demo-dynamic-record-file-input")
      .last();
  }

  pluginDemoDynamicRecordRemoveAttachment(): Locator {
    return this.page
      .getByTestId("plugin-demo-dynamic-record-remove-attachment")
      .last();
  }

  pluginDemoDynamicRecordSubmitButton(): Locator {
    return this.page.getByTestId("plugin-demo-dynamic-record-submit").last();
  }

  pluginDemoDynamicRecordRow(title: string): Locator {
    return this.pluginDemoDynamicRecordGrid()
      .locator("tbody tr", { hasText: title })
      .first();
  }

  pluginDemoDynamicEditButton(title: string): Locator {
    return this.pluginDemoDynamicRecordRow(title)
      .getByRole("button", { name: "编辑" })
      .first();
  }

  pluginDemoDynamicDeleteButton(title: string): Locator {
    return this.pluginDemoDynamicRecordRow(title)
      .getByRole("button", { name: "删除" })
      .first();
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

  dynamicUploadListItem(): Locator {
    return this.dynamicUploadDialog().locator(".ant-upload-list-item").last();
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
    return this.dynamicUploadDialog().locator(".ant-modal-close").last();
  }

  uploadSuccessDialog(): Locator {
    return this.dynamicUploadDialog()
      .getByTestId("plugin-dynamic-upload-success")
      .first();
  }

  messageNotice(text: string): Locator {
    return this.page
      .locator(".ant-message-notice")
      .filter({ hasText: text })
      .last();
  }

  tableColumn(title: string): Locator {
    return this.page
      .locator(".vxe-table--header .vxe-cell--title", { hasText: title })
      .first();
  }

  pluginMainRows(): Locator {
    return this.page.locator(".vxe-table--main-wrapper .vxe-body--row");
  }

  pluginRow(pluginId: string): Locator {
    return this.pluginMainRows().filter({ hasText: pluginId }).first();
  }

  hostServiceAuthModal(): Locator {
    return this.page.getByTestId("plugin-host-service-auth-modal").last();
  }

  hostServiceAuthDialog(): Locator {
    return this.page
      .getByRole("dialog", {
        name: /安装插件(?:并确认授权)?|启用插件(?:并确认授权)?/,
      })
      .last();
  }

  hostServiceAuthConfirmButton(): Locator {
    return this.hostServiceAuthDialog()
      .getByRole("button", { name: /确\s*认|确\s*定/i })
      .last();
  }

  hostServiceAuthInstallAndEnableButton(): Locator {
    return this.hostServiceAuthDialog()
      .getByTestId("plugin-install-enable-button")
      .last();
  }

  uninstallDialog(): Locator {
    return this.page.getByRole("dialog", { name: "卸载插件" }).last();
  }

  pluginDetailDialog(): Locator {
    return this.page.getByRole("dialog", { name: "插件详情" }).last();
  }

  pluginDetailModal(): Locator {
    return this.page.getByTestId("plugin-detail-modal").last();
  }

  pluginDetailEmptyHostServices(): Locator {
    return this.page.getByTestId("plugin-detail-empty-host-services").last();
  }

  uninstallPurgeCheckbox(): Locator {
    return this.page.getByTestId("plugin-uninstall-purge-checkbox").last();
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

  pluginSidebarIntroTitle(): Locator {
    return this.page
      .getByRole("heading", { name: "源码插件示例已生效" })
      .first();
  }

  pluginSidebarIntroSummary(): Locator {
    return this.page.getByText(
      "这是一条来自 plugin-demo-source 接口的简要介绍，用于验证源码插件菜单页可读取插件后端数据。",
    );
  }

  pluginSourceRecordGridTitle(): Locator {
    return this.page.getByText("示例记录").first();
  }

  pluginSourceRecordAddButton(): Locator {
    return this.page.getByTestId("plugin-demo-source-record-add").first();
  }

  pluginSourceRecordModal(): Locator {
    return this.page
      .getByRole("dialog", { name: /新增示例记录|编辑示例记录/ })
      .last();
  }

  pluginSourceRecordAttachmentAlert(): Locator {
    return this.page
      .getByTestId("plugin-demo-source-record-attachment-alert")
      .last();
  }

  pluginSourceRecordUploadSection(): Locator {
    return this.page
      .getByTestId("plugin-demo-source-record-upload-section")
      .last();
  }

  pluginSourceRecordExistingAttachment(): Locator {
    return this.page
      .getByTestId("plugin-demo-source-record-existing-attachment")
      .last();
  }

  pluginSourceRecordRemoveAttachmentOption(): Locator {
    return this.page
      .getByTestId("plugin-demo-source-record-remove-attachment-option")
      .last();
  }

  pluginSourceRecordDragger(): Locator {
    return this.page.getByTestId("plugin-demo-source-record-dragger").last();
  }

  pluginSourceRecordTitleInput(): Locator {
    return this.page
      .getByTestId("plugin-demo-source-record-title-input")
      .last();
  }

  pluginSourceRecordContentInput(): Locator {
    return this.page
      .getByTestId("plugin-demo-source-record-content-input")
      .last();
  }

  pluginSourceRecordRow(title: string): Locator {
    return this.page.locator(".vxe-body--row", { hasText: title }).first();
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
    await waitForUploadReady(this.dynamicUploadDialog());

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
    const installButton = await this.pluginActionButton(pluginId, /安\s*装/);
    await expect(installButton).toBeVisible();
    await installButton.click();
    await expect(this.hostServiceAuthDialog()).toBeVisible();
    await this.hostServiceAuthConfirmButton().click();
    await expect(this.hostServiceAuthDialog()).toHaveCount(0);
    await expect(
      await this.pluginActionButton(pluginId, /卸\s*载/),
    ).toBeVisible();
  }

  async installAndEnablePlugin(pluginId: string) {
    const installButton = await this.pluginActionButton(pluginId, /安\s*装/);
    await expect(installButton).toBeVisible();
    await installButton.click();
    await this.confirmInstallAndEnable();
  }

  async ensurePluginInstalled(pluginId: string) {
    const installButton = await this.pluginActionButton(pluginId, /安\s*装/);
    const installVisible = await installButton
      .isVisible({ timeout: 1500 })
      .catch(() => false);
    if (!installVisible) {
      return false;
    }
    await this.installPlugin(pluginId);
    return true;
  }

  async openInstallAuthorization(pluginId: string) {
    const installButton = await this.pluginActionButton(pluginId, /安\s*装/);
    await expect(installButton).toBeVisible();
    await installButton.click();
    await expect(this.hostServiceAuthModal()).toBeVisible();
  }

  async uninstallPlugin(pluginId: string) {
    await this.uninstallPluginWithOptions(pluginId, true);
  }

  async ensurePluginUninstalled(pluginId: string) {
    const uninstallButton = await this.pluginActionButton(pluginId, /卸\s*载/);
    const uninstallVisible = await uninstallButton
      .isVisible({ timeout: 1500 })
      .catch(() => false);
    if (!uninstallVisible) {
      return false;
    }
    await this.uninstallPlugin(pluginId);
    return true;
  }

  async openPluginDetail(pluginId: string) {
    const detailButton = await this.pluginActionButton(pluginId, /详\s*情/);
    await expect(detailButton).toBeVisible();
    await detailButton.click();
    await expect(this.pluginDetailDialog()).toBeVisible();
  }

  async uninstallPluginWithOptions(
    pluginId: string,
    purgeStorageData: boolean,
  ) {
    const uninstallButton = await this.pluginActionButton(pluginId, /卸\s*载/);
    await expect(uninstallButton).toBeVisible();
    await uninstallButton.click();
    await expect(this.uninstallDialog()).toBeVisible();
    const checkboxVisible = await this.uninstallPurgeCheckbox()
      .isVisible({ timeout: 1500 })
      .catch(() => false);
    if (checkboxVisible) {
      const isChecked = await this.uninstallPurgeCheckbox().isChecked();
      if (isChecked !== purgeStorageData) {
        await this.uninstallPurgeCheckbox().click();
      }
    }
    await this.uninstallDialog()
      .getByRole("button", { name: /确\s*认|确\s*定/i })
      .last()
      .click();
    await expect(this.uninstallDialog()).toHaveCount(0);
    await expect(
      await this.pluginActionButton(pluginId, /安\s*装/),
    ).toBeVisible();
  }

  async createPluginDemoDynamicRecord(input: {
    attachmentPath?: string;
    content: string;
    title: string;
  }) {
    await this.pluginDemoDynamicRecordAddButton().click();
    await expect(this.pluginDemoDynamicRecordModal()).toBeVisible();
    await this.pluginDemoDynamicRecordTitleInput().fill(input.title);
    await this.pluginDemoDynamicRecordContentInput().fill(input.content);
    if (input.attachmentPath) {
      await this.pluginDemoDynamicRecordFileInput().setInputFiles(
        input.attachmentPath,
      );
    }
    await this.pluginDemoDynamicRecordSubmitButton().click();
    await expect(this.pluginDemoDynamicRecordModal()).toHaveAttribute(
      "data-open",
      "false",
    );
    await expect(this.pluginDemoDynamicRecordRow(input.title)).toBeVisible();
  }

  async updatePluginDemoDynamicRecord(
    currentTitle: string,
    input: {
      attachmentPath?: string;
      content: string;
      removeAttachment?: boolean;
      title: string;
    },
  ) {
    await this.pluginDemoDynamicEditButton(currentTitle).click();
    await expect(this.pluginDemoDynamicRecordModal()).toBeVisible();
    await this.pluginDemoDynamicRecordTitleInput().fill(input.title);
    await this.pluginDemoDynamicRecordContentInput().fill(input.content);
    if (input.removeAttachment) {
      const checkbox = this.pluginDemoDynamicRecordRemoveAttachment().locator(
        'input[type="checkbox"]',
      );
      if ((await checkbox.isChecked()) !== true) {
        await checkbox.click();
      }
    }
    if (input.attachmentPath) {
      await this.pluginDemoDynamicRecordFileInput().setInputFiles(
        input.attachmentPath,
      );
    }
    await this.pluginDemoDynamicRecordSubmitButton().click();
    await expect(this.pluginDemoDynamicRecordModal()).toHaveAttribute(
      "data-open",
      "false",
    );
    await expect(this.pluginDemoDynamicRecordRow(input.title)).toBeVisible();
  }

  async deletePluginDemoDynamicRecord(title: string) {
    this.page.once("dialog", async (dialog) => {
      await dialog.accept();
    });
    await this.pluginDemoDynamicDeleteButton(title).click();
    await expect(this.pluginDemoDynamicRecordRow(title)).toHaveCount(0);
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

  async expectInstallActionVisible(pluginId: string) {
    await expect(
      await this.pluginActionButton(pluginId, /安\s*装/),
    ).toBeVisible();
  }

  async expectInstallActionHidden(pluginId: string) {
    await expect(
      await this.pluginActionButton(pluginId, /安\s*装/),
    ).toHaveCount(0);
  }

  async expectUninstallActionVisible(pluginId: string) {
    await expect(
      await this.pluginActionButton(pluginId, /卸\s*载/),
    ).toBeVisible();
  }

  async expectUninstallActionHidden(pluginId: string) {
    await expect(
      await this.pluginActionButton(pluginId, /卸\s*载/),
    ).toHaveCount(0);
  }

  async expectPluginSwitchDisabled(pluginId: string) {
    await expect(this.pluginEnabledSwitch(pluginId)).toHaveClass(
      /ant-switch-disabled/,
    );
  }

  async openEnableAuthorization(pluginId: string) {
    const switcher = this.pluginEnabledSwitch(pluginId);
    await expect(switcher).toBeVisible();
    await switcher.click();
    await expect(this.hostServiceAuthModal()).toBeVisible();
  }

  async confirmHostServiceAuthorization() {
    await this.hostServiceAuthConfirmButton().click();
    await expect(this.hostServiceAuthDialog()).toHaveCount(0);
  }

  async confirmInstallAndEnable() {
    await expect(this.hostServiceAuthDialog()).toBeVisible();
    await expect(this.hostServiceAuthInstallAndEnableButton()).toBeVisible();
    await this.hostServiceAuthInstallAndEnableButton().click();
    await expect(this.hostServiceAuthDialog()).toHaveCount(0);
  }

  private async pluginActionButton(pluginId: string, name: RegExp) {
    const row = this.pluginRow(pluginId);
    await expect(row, `未找到插件行: ${pluginId}`).toBeVisible();

    const rowID = await row.getAttribute("rowid");
    expect(rowID, `未找到插件行 rowid: ${pluginId}`).toBeTruthy();
    return this.page
      .locator(`.vxe-table--fixed-right-wrapper .vxe-body--row[rowid=\"${rowID}\"]`)
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
    await expect(this.page.getByTestId("dashboard-workspace-page")).toBeVisible();
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
    const [vxeTooltipAppeared, antTooltipAppeared] = await Promise.all([
      this.vxeTooltip()
        .waitFor({ state: "visible", timeout: 5000 })
        .then(() => true)
        .catch(() => false),
      this.antTooltip()
        .waitFor({ state: "visible", timeout: 5000 })
        .then(() => true)
        .catch(() => false),
    ]);
    expect(
      vxeTooltipAppeared,
      "描述列悬浮后不应回退到 VXE 浮层提示",
    ).toBeFalsy();
    expect(
      antTooltipAppeared,
      "描述列悬浮后不应额外弹出 Ant Design Tooltip",
    ).toBeFalsy();
    const delayedTitleCount = await this.page
      .locator("[title]")
      .evaluateAll((elements, text) => {
        return elements.filter((element) =>
          (element.getAttribute("title") || "").includes(text),
        ).length;
      }, descriptionText);
    expect(delayedTitleCount, "描述列应只保留单一系统默认提示来源").toBe(1);
  }

  async openSidebarExampleFromMenu() {
    await this.clickSidebarMenuItem("源码插件示例");
    await expect(this.pluginSidebarIntroTitle()).toHaveCount(0);
    await expect(this.pluginSidebarIntroSummary()).toHaveCount(0);
    await expect(this.pluginSourceRecordGridTitle()).toBeVisible();
  }

  async createSourceDemoRecord(
    title: string,
    content: string,
    filePath?: string,
  ) {
    await expect(this.pluginSourceRecordAddButton()).toBeVisible();
    await this.pluginSourceRecordAddButton().click();
    await expect(this.pluginSourceRecordModal()).toBeVisible();
    await this.expectSourceRecordModalCompactLayout();
    await this.pluginSourceRecordTitleInput().fill(title);
    await this.pluginSourceRecordContentInput().fill(content);
    if (filePath) {
      const [fileChooser] = await Promise.all([
        this.page.waitForEvent("filechooser"),
        this.pluginSourceRecordDragger().click(),
      ]);
      await fileChooser.setFiles(filePath);
      await expect(
        this.pluginSourceRecordModal().locator(".ant-upload-list-item"),
      ).toBeVisible();
    }
    await this.pluginSourceRecordModal()
      .getByRole("button", { name: /确\s*认|确\s*定/i })
      .last()
      .click();
    await expect(this.pluginSourceRecordModal()).toHaveCount(0);
    await expect(this.pluginSourceRecordRow(title)).toBeVisible();
  }

  async editSourceDemoRecord(
    currentTitle: string,
    nextTitle: string,
    nextContent: string,
  ) {
    const editButton = await this.pluginSourceRecordActionButton(
      currentTitle,
      /编\s*辑/,
    );
    await expect(editButton).toBeVisible();
    await editButton.click();
    await expect(this.pluginSourceRecordModal()).toBeVisible();
    await expect(this.pluginSourceRecordTitleInput()).toHaveValue(currentTitle);
    await this.expectSourceRecordModalCompactLayout();
    await this.pluginSourceRecordTitleInput().fill(nextTitle);
    await this.pluginSourceRecordContentInput().fill(nextContent);
    await this.pluginSourceRecordModal()
      .getByRole("button", { name: /确\s*认|确\s*定/i })
      .last()
      .click();
    await expect(this.pluginSourceRecordModal()).toHaveCount(0);
    await expect(this.pluginSourceRecordRow(nextTitle)).toBeVisible();
  }

  async deleteSourceDemoRecord(title: string) {
    const deleteButton = await this.pluginSourceRecordActionButton(
      title,
      /删\s*除/,
    );
    await expect(deleteButton).toBeVisible();
    await deleteButton.click();
    const confirmPopover = this.page.locator(".ant-popover:visible").last();
    await expect(confirmPopover).toBeVisible();
    await confirmPopover
      .getByRole("button", { name: /确\s*定|确\s*认/i })
      .click();
    await expect(this.pluginSourceRecordRow(title)).toHaveCount(0);
  }

  async downloadSourceDemoAttachment(fileName: string) {
    const downloadPromise = this.page.waitForEvent("download");
    await this.page.getByRole("button", { name: fileName }).first().click();
    return await downloadPromise;
  }

  private async pluginSourceRecordActionButton(title: string, name: RegExp) {
    const row = this.pluginSourceRecordRow(title);
    await expect(row, `未找到示例记录行: ${title}`).toBeVisible();
    return row.getByRole("button", { name }).first();
  }

  private async expectSourceRecordModalCompactLayout() {
    const modal = this.pluginSourceRecordModal();
    const alert = this.pluginSourceRecordAttachmentAlert();
    const uploadSection = this.pluginSourceRecordUploadSection();

    await expect(alert).toBeVisible();
    await expect(uploadSection).toBeVisible();

    const modalWidth = await modal.evaluate((element) => {
      return Math.round(element.getBoundingClientRect().width);
    });
    expect(
      modalWidth,
      "源码插件记录弹窗宽度应收敛，避免继续维持过宽布局",
    ).toBeLessThanOrEqual(620);

    const alertBox = await alert.boundingBox();
    const uploadSectionBox = await uploadSection.boundingBox();
    expect(alertBox, "附件提示块应可见").toBeTruthy();
    expect(uploadSectionBox, "上传区域应可见").toBeTruthy();

    const verticalGap = uploadSectionBox!.y - (alertBox!.y + alertBox!.height);
    expect(
      verticalGap,
      "附件提示块与上传区域之间应保留至少 16px 的垂直间距",
    ).toBeGreaterThanOrEqual(16);

    const existingAttachment = this.pluginSourceRecordExistingAttachment();
    const removeAttachmentOption =
      this.pluginSourceRecordRemoveAttachmentOption();
    const editSpacingVisible = await existingAttachment
      .isVisible({ timeout: 1000 })
      .catch(() => false);
    if (!editSpacingVisible) {
      return;
    }

    await expect(removeAttachmentOption).toBeVisible();
    const existingAttachmentBox = await existingAttachment.boundingBox();
    const removeAttachmentBox = await removeAttachmentOption.boundingBox();
    const draggerBox = await this.pluginSourceRecordDragger().boundingBox();
    expect(existingAttachmentBox, "当前附件信息块应可见").toBeTruthy();
    expect(removeAttachmentBox, "移除附件选项块应可见").toBeTruthy();
    expect(draggerBox, "附件上传区应可见").toBeTruthy();

    const removeOptionGapAbove =
      removeAttachmentBox!.y -
      (existingAttachmentBox!.y + existingAttachmentBox!.height);
    const removeOptionGapBelow =
      draggerBox!.y - (removeAttachmentBox!.y + removeAttachmentBox!.height);
    expect(
      removeOptionGapAbove,
      "“提交时移除当前附件”选项与当前附件信息块之间应保留足够间距",
    ).toBeGreaterThanOrEqual(12);
    expect(
      removeOptionGapBelow,
      "“提交时移除当前附件”选项与上传区之间应保留足够间距",
    ).toBeGreaterThanOrEqual(12);
  }
}
