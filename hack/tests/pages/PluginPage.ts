import { Page, Locator, expect } from "@playwright/test";

import { workspacePath } from "../fixtures/config";
import { waitForRouteReady, waitForUploadReady } from "../support/ui";

const pluginManageMenuPattern = /插件管理|Plugin Management/iu;
const extensionCenterMenuPattern = /扩展中心|Extension Center|Extensions/iu;
const pluginTableTitlePattern = /插件列表|Plugin List/iu;
const pluginInstallActionPattern = /安\s*装|Install/iu;
const pluginUninstallActionPattern = /卸\s*载|Uninstall/iu;
const pluginDetailActionPattern = /详\s*情|Detail(?:s)?/iu;
const pluginUpgradeActionPattern = /升\s*级|重试升级|Upgrade|Retry Upgrade/iu;
const confirmActionPattern = /确\s*认|确\s*定|confirm|ok/iu;
const cancelActionPattern = /取\s*消|cancel/iu;
const pluginLifecycleActionTimeout = 120_000;
const pluginLifecycleRefreshProbeTimeout = 5_000;

type SidebarMenuName = RegExp | string;

type PluginColumnHelpName =
  | "mockData"
  | "runtimeState"
  | "supportsMultiTenant"
  | "tenantProvisioning"
  | "type";

export class PluginPage {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  get tableTitle(): Locator {
    // Prefer the visible grid title; keep-alive or secondary headers may leave
    // a hidden matching node that would break `.first()`.
    return this.page
      .getByText(pluginTableTitlePattern)
      .filter({ visible: true })
      .first();
  }

  pluginListHelpIcon(): Locator {
    return this.page.getByTestId("plugin-list-help-icon").first();
  }

  pluginColumnHelpIcon(name: PluginColumnHelpName): Locator {
    const testIds = {
      mockData: "plugin-mock-data-column-help-icon",
      runtimeState: "plugin-runtime-state-column-help-icon",
      supportsMultiTenant: "plugin-supports-multi-tenant-column-help-icon",
      tenantProvisioning: "plugin-tenant-provisioning-column-help-icon",
      type: "plugin-type-column-help-icon",
    } as const;
    return this.page.getByTestId(testIds[name]).first();
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
    return this.page
      .locator("aside")
      .filter({ has: this.page.locator('ul[role="menu"]') })
      .first();
  }

  sidebarMenuItem(menuName: SidebarMenuName): Locator {
    if (typeof menuName !== "string") {
      return this.sidebarMenu.getByRole("menuitem", { name: menuName }).first();
    }
    return this.sidebarMenu
      .getByRole("menuitem", { name: menuName, exact: true })
      .first();
  }

  private sidebarSubmenuForMenuItem(menuName: SidebarMenuName): Locator {
    return this.sidebarMenu
      .locator(".ant-menu-submenu, .vben-sub-menu")
      .filter({ hasText: menuName })
      .locator(".ant-menu-submenu-title, .vben-sub-menu-content")
      .first();
  }

  private sidebarSubmenuForPattern(pattern: RegExp): Locator {
    return this.sidebarMenu
      .locator(".ant-menu-submenu-title, .vben-sub-menu-content")
      .filter({ hasText: pattern })
      .first();
  }

  private async expandSidebarSubmenu(submenuTitle: Locator) {
    const submenu = submenuTitle
      .locator(
        "xpath=ancestor::*[contains(concat(' ', normalize-space(@class), ' '), ' ant-menu-submenu ') or contains(concat(' ', normalize-space(@class), ' '), ' vben-sub-menu ')][1]",
      )
      .first();
    const className = (await submenu.getAttribute("class").catch(() => "")) ?? "";
    if (
      !className.includes("ant-menu-submenu-open") &&
      !className.includes("is-opened")
    ) {
      await submenuTitle.click();
    }
  }

  private async expandExtensionCenterIfVisible() {
    const submenu = this.sidebarSubmenuForPattern(extensionCenterMenuPattern);
    const visible = await submenu
      .isVisible({ timeout: 1500 })
      .catch(() => false);
    if (!visible) {
      return false;
    }
    await this.expandSidebarSubmenu(submenu);
    return true;
  }

  async clickSidebarMenuItem(menuName: SidebarMenuName) {
    const menuItem = await this.expectSidebarMenuVisible(menuName);
    await menuItem.click();
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

  dynamicUploadDialog(): Locator {
    return this.page
      .getByRole("dialog", { name: /上传动态插件|Upload Dynamic Plugin/iu })
      .last();
  }

  dynamicUploadTriggerLabel(): Locator {
    return this.dynamicUploadTrigger.getByText(/上传插件|Upload Plugin/iu);
  }

  dynamicUploadHint(): Locator {
    return this.dynamicUploadDialog().getByText(
      /上传单个 `?\.wasm`? 动态插件包|Upload a single `?\.wasm`? artifact/iu,
    );
  }

  dynamicUploadListItem(): Locator {
    return this.dynamicUploadDialog().locator(".ant-upload-list-item").last();
  }

  dynamicOverwriteHint(): Locator {
    return this.dynamicUploadDialog().getByText(
      /允许覆盖相同插件标识和版本的已有插件包|Allow upload to overwrite an existing plugin package/iu,
    );
  }

  dynamicUploadConfirmButton(): Locator {
    return this.dynamicUploadDialog()
      .getByRole("button", {
        name: /确\s*认|确\s*定|知\s*道了|知\s*道|confirm|got it|ok/iu,
      })
      .last();
  }

  dynamicUploadCancelButton(): Locator {
    return this.dynamicUploadDialog()
      .getByRole("button", { name: cancelActionPattern })
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
    return this.messageNotices(text).last();
  }

  messageNotices(text: string): Locator {
    return this.page.locator(".ant-message-notice").filter({ hasText: text });
  }

  tableColumn(title: string): Locator {
    return this.page
      .locator(".vxe-table--header .vxe-cell--title", { hasText: title })
      .first();
  }

  tableHeaderCell(title: string): Locator {
    return this.page
      .locator(".vxe-table--header .vxe-header--column")
      .filter({ hasText: title })
      .first();
  }

  pluginMainRows(): Locator {
    return this.page.locator(".vxe-table--main-wrapper .vxe-body--row");
  }

  pluginRow(pluginId: string): Locator {
    return this.pluginMainRows().filter({ hasText: pluginId }).first();
  }

  private pluginLifecycleSwitch(row: Locator): Locator {
    return row
      .locator(
        '.ant-switch:not(.ant-switch-small):not([data-testid^="plugin-tenant-provisioning-"])',
      )
      .first();
  }

  private pluginIdSearchInput(): Locator {
    return this.page
      .getByRole("textbox", { name: /插件标识|Plugin ID/iu })
      .first();
  }

  private async filterByPluginId(pluginId: string) {
    const input = this.pluginIdSearchInput();
    await expect(input).toBeVisible();
    await input.fill(pluginId);

    const listResponse = this.page
      .waitForResponse(
        (response) => {
          const request = response.request();
          return (
            request.method() === "GET" &&
            new URL(response.url()).pathname.endsWith("/plugins")
          );
        },
        { timeout: 30_000 },
      )
      .catch(() => null);

    await this.page.getByRole("button", { name: /搜\s*索|Search/iu }).click();
    await listResponse;
  }

  private async ensurePluginRowVisible(pluginId: string) {
    const row = this.pluginRow(pluginId);
    if (await row.isVisible({ timeout: 1500 }).catch(() => false)) {
      return row;
    }

    await this.filterByPluginId(pluginId);
    const filteredRow = this.pluginRow(pluginId);
    await expect(filteredRow, `未找到插件行: ${pluginId}`).toBeVisible();
    return filteredRow;
  }

  hostServiceAuthModal(): Locator {
    return this.page.getByTestId("plugin-host-service-auth-modal").last();
  }

  pluginInstallDescriptions(): Locator {
    return this.page.getByTestId("plugin-install-descriptions").last();
  }

  pluginInstallDescriptionLabels(): Locator {
    return this.pluginInstallDescriptions().locator(
      ".ant-descriptions-item-label",
    );
  }

  hostServiceAuthDialog(): Locator {
    return this.page
      .getByRole("dialog", {
        name: /安装前审核插件授权|启用前审核插件授权|安装插件(?:并确认授权)?|启用插件(?:并确认授权)?|Install Plugin|Enable Plugin/i,
      })
      .last();
  }

  hostServiceAuthConfirmButton(): Locator {
    return this.hostServiceAuthDialog()
      .getByRole("button", { name: confirmActionPattern })
      .last();
  }

  hostServiceAuthInstallAndEnableButton(): Locator {
    return this.hostServiceAuthDialog()
      .getByTestId("plugin-install-enable-button")
      .last();
  }

  pluginInstallMockDataSection(): Locator {
    return this.page.getByTestId("plugin-install-mock-data-section").last();
  }

  pluginInstallMockDataCheckbox(): Locator {
    return this.hostServiceAuthDialog()
      .getByRole("checkbox", {
        name: /是否安装示例数据|是否安裝示例資料|Install mock data\?/iu,
      })
      .last();
  }

  pluginInstallMockDataHelpIcon(): Locator {
    return this.page.getByTestId("plugin-install-mock-data-help-icon").last();
  }

  pluginInstallModeSection(): Locator {
    return this.page.getByTestId("plugin-install-mode-section").last();
  }

  pluginInstallModeRow(): Locator {
    return this.page.getByTestId("plugin-install-mode-row").last();
  }

  pluginInstallModeSelect(): Locator {
    return this.hostServiceAuthDialog()
      .getByTestId("plugin-install-mode-select")
      .last();
  }

  pluginInstallModeDescription(): Locator {
    return this.page.getByTestId("plugin-install-mode-description").last();
  }

  pluginDependencySummary(): Locator {
    return this.page.getByTestId("plugin-dependency-summary").last();
  }

  pluginDependencyFrameworkBlocker(): Locator {
    return this.page
      .getByTestId("plugin-dependency-framework-blocker")
      .last();
  }

  pluginDependencyBlockers(): Locator {
    return this.page.getByTestId("plugin-dependency-blockers").last();
  }

  pluginDependencyReverseBlockers(): Locator {
    return this.page.getByTestId("plugin-dependency-reverse-blockers").last();
  }

  installModeStandaloneSelector(): Locator {
    return this.page.getByTestId("install-mode-selector").last();
  }

  pluginMockDataValue(pluginId: string): Locator {
    return this.page.getByTestId(`plugin-mock-data-value-${pluginId}`).first();
  }

  pluginSupportsMultiTenantValue(pluginId: string): Locator {
    return this.page
      .getByTestId(`plugin-supports-multi-tenant-${pluginId}`)
      .first();
  }

  pluginTenantProvisioningSwitch(pluginId: string): Locator {
    return this.page
      .getByTestId(`plugin-tenant-provisioning-${pluginId}`)
      .first();
  }

  pluginTenantProvisioningSwitches(pluginId: string): Locator {
    return this.page.getByTestId(`plugin-tenant-provisioning-${pluginId}`);
  }

  pluginRuntimeState(pluginId: string): Locator {
    return this.page.getByTestId(`plugin-runtime-state-${pluginId}`).first();
  }

  pluginVersionValue(pluginId: string): Locator {
    return this.page.getByTestId(`plugin-version-${pluginId}`).first();
  }

  pluginDetailAction(pluginId: string): Locator {
    return this.page.getByTestId(`plugin-detail-button-${pluginId}`).last();
  }

  pluginManageAction(pluginId: string): Locator {
    return this.page.getByTestId(`plugin-manage-button-${pluginId}`).last();
  }

  pluginManageActionWrapper(pluginId: string): Locator {
    return this.page.getByTestId(`plugin-manage-wrapper-${pluginId}`).last();
  }

  pluginManualRepairAction(pluginId: string): Locator {
    return this.page.getByTestId(`plugin-abnormal-repair-${pluginId}`).last();
  }

  async openPluginManagement(pluginId: string) {
    const manageAction = this.pluginManageAction(pluginId);
    await expect(manageAction).toBeVisible();
    await expect(manageAction).toBeEnabled();
    await manageAction.click();
  }

  async expectManualRepairActionMatchesDetailStyle(pluginId: string) {
    const detailAction = this.pluginDetailAction(pluginId);
    const manualRepairAction = this.pluginManualRepairAction(pluginId);

    await expect(detailAction).toBeVisible();
    await expect(manualRepairAction).toBeVisible();
    await expect(manualRepairAction).toHaveClass(/ant-btn/u);

    const [detailMetrics, manualRepairMetrics] = await Promise.all([
      detailAction.evaluate((node) => {
        const style = getComputedStyle(node);
        const rect = node.getBoundingClientRect();
        return {
          fontSize: style.fontSize,
          height: rect.height,
        };
      }),
      manualRepairAction.evaluate((node) => {
        const style = getComputedStyle(node);
        const rect = node.getBoundingClientRect();
        return {
          fontSize: style.fontSize,
          height: rect.height,
        };
      }),
    ]);

    expect(manualRepairMetrics.fontSize).toBe(detailMetrics.fontSize);
    expect(
      Math.abs(manualRepairMetrics.height - detailMetrics.height),
    ).toBeLessThanOrEqual(1);
  }

  pluginUpgradeModal(): Locator {
    return this.page.getByTestId("plugin-upgrade-modal").last();
  }

  pluginUpgradeDialog(): Locator {
    return this.page
      .getByRole("dialog", { name: /升级插件|Upgrade Plugin/iu })
      .last();
  }

  pluginUpgradeConfirmButton(): Locator {
    return this.pluginUpgradeDialog()
      .getByRole("button", {
        name: /确认升级|Confirm Upgrade|确\s*认|confirm/iu,
      })
      .last();
  }

  pluginUpgradeFromManifest(): Locator {
    return this.page.getByTestId("plugin-upgrade-from-manifest").last();
  }

  pluginUpgradeToManifest(): Locator {
    return this.page.getByTestId("plugin-upgrade-to-manifest").last();
  }

  pluginUpgradeSqlSummary(): Locator {
    return this.page.getByTestId("plugin-upgrade-sql-summary").last();
  }

  pluginUpgradeRiskSectionTitle(): Locator {
    return this.page.getByTestId("plugin-upgrade-risk-section-title").last();
  }

  uninstallDialog(): Locator {
    return this.page
      .getByRole("dialog", { name: /卸载插件|Uninstall Plugin/iu })
      .last();
  }

  uninstallConfirmButton(): Locator {
    return this.uninstallDialog()
      .getByRole("button", { name: confirmActionPattern })
      .last();
  }

  pluginDetailDialog(): Locator {
    return this.page
      .getByRole("dialog", { name: /插件详情|Plugin Details/iu })
      .last();
  }

  pluginDetailModal(): Locator {
    return this.page.getByTestId("plugin-detail-modal").last();
  }

  pluginDetailDescriptions(): Locator {
    return this.page.getByTestId("plugin-detail-descriptions").last();
  }

  pluginDetailDescriptionLabels(): Locator {
    return this.pluginDetailDescriptions().locator(
      ".ant-descriptions-item-label",
    );
  }

  pluginRouteReviewToggle(): Locator {
    return this.page.getByTestId("plugin-route-review-toggle").last();
  }

  pluginDetailDescriptionRow(): Locator {
    return this.page.getByTestId("plugin-detail-description-row").last();
  }

  pluginDetailEmptyHostServices(): Locator {
    return this.page.getByTestId("plugin-detail-empty-host-services").last();
  }

  pluginDetailHasMockData(): Locator {
    return this.page.getByTestId("plugin-detail-has-mock-data").last();
  }

  pluginDetailSupportsMultiTenant(): Locator {
    return this.page
      .getByTestId("plugin-detail-supports-multi-tenant")
      .last();
  }

  pluginDetailTenantProvisioning(): Locator {
    return this.page.getByTestId("plugin-detail-tenant-provisioning").last();
  }

  pluginDetailScopeNature(): Locator {
    return this.page.getByTestId("plugin-detail-scope-nature").last();
  }

  pluginDetailInstallMode(): Locator {
    return this.page.getByTestId("plugin-detail-install-mode").last();
  }

  pluginDetailPluginScope(): Locator {
    return this.page.getByTestId("plugin-detail-plugin-scope").last();
  }

  pluginDetailPluginStatus(): Locator {
    return this.page.getByTestId("plugin-detail-plugin-status").last();
  }

  pluginAutoEnableTag(pluginId: string): Locator {
    return this.page.getByTestId(`plugin-auto-enable-tag-${pluginId}`).first();
  }

  pluginBuiltinTag(pluginId: string): Locator {
    return this.page.getByTestId(`plugin-builtin-tag-${pluginId}`).first();
  }

  pluginBuiltinDetailAlert(): Locator {
    return this.page.getByTestId("plugin-builtin-detail-alert").last();
  }

  pluginDetailDistribution(): Locator {
    return this.page.getByTestId("plugin-detail-distribution").last();
  }

  pluginNameCell(pluginId: string): Locator {
    return this.page.getByTestId(`plugin-name-cell-${pluginId}`).first();
  }

  pluginAutoEnableDetailAlert(): Locator {
    return this.page.getByTestId("plugin-auto-enable-detail-alert").last();
  }

  pluginAutoEnableUninstallAlert(): Locator {
    return this.page.getByTestId("plugin-auto-enable-uninstall-alert").last();
  }

  pluginManagedActionDialog(): Locator {
    return this.page.locator(".ant-modal-confirm").last();
  }

  uninstallPurgeCheckbox(): Locator {
    return this.uninstallDialog()
      .getByRole("checkbox", {
        name: /同时清理插件自有存储数据|同時清理插件自有存儲數據|Also clear plugin-owned storage data/iu,
      })
      .last();
  }

  uninstallPurgeCheckboxWrapper(): Locator {
    return this.page.getByTestId("plugin-uninstall-purge-checkbox").last();
  }

  uninstallPurgeWarning(): Locator {
    return this.page.getByTestId("plugin-uninstall-purge-warning").last();
  }

  lifecyclePreconditionDialog(): Locator {
    return this.page.getByTestId("lifecycle-precondition-dialog").last();
  }

  lifecyclePreconditionReasonAlert(): Locator {
    return this.page.getByTestId("lifecycle-precondition-reason-alert").last();
  }

  lifecyclePreconditionForceAlert(): Locator {
    return this.page.getByTestId("lifecycle-precondition-force-alert").last();
  }

  lifecyclePreconditionForcePluginIdInput(): Locator {
    return this.page.getByTestId("lifecycle-precondition-force-plugin-id").last();
  }

  lifecyclePreconditionReasonText(): Locator {
    return this.page.getByTestId("lifecycle-precondition-reason");
  }

  lifecyclePreconditionConfirmButton(): Locator {
    return this.page
      .getByRole("dialog", { name: /生命周期前置条件|Lifecycle Precondition/iu })
      .last()
      .getByRole("button", { name: confirmActionPattern })
      .last();
  }

  pluginEnabledSwitch(pluginId: string): Locator {
    return this.pluginLifecycleSwitch(this.pluginRow(pluginId));
  }

  pluginEnabledSwitches(pluginId: string): Locator {
    const row = this.pluginRow(pluginId);
    return row.locator(
      '.ant-switch:not(.ant-switch-small):not([data-testid^="plugin-tenant-provisioning-"])',
    );
  }

  pluginDescriptionCell(pluginId: string): Locator {
    return this.pluginRow(pluginId)
      .getByTestId(`plugin-description-${pluginId}`)
      .first();
  }

  antTooltip(): Locator {
    return this.page.locator(".ant-tooltip:visible");
  }

  async expectColumnHelpTooltip(
    name: PluginColumnHelpName,
    text: string | RegExp,
  ) {
    await this.pluginColumnHelpIcon(name).hover();
    await expect(this.antTooltip().filter({ hasText: text }).last()).toBeVisible();
  }

  vxeTooltip(): Locator {
    return this.page.locator(".vxe-table--tooltip-wrapper:visible");
  }

  async gotoManage() {
    await this.page.goto("/system/plugin");
    await expect(this.tableTitle).toBeVisible();
  }

  async searchByPluginId(pluginId: string) {
    await this.filterByPluginId(pluginId);
    await expect(this.pluginRow(pluginId)).toBeVisible();
  }

  async syncPlugins() {
    await this.page
      .getByRole("button", { name: /同步插件|Synchronize Plugins/iu })
      .click();
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
    const uploadPayload = (await uploadResponse.json().catch(() => null)) as {
      code?: number;
      message?: string;
    } | null;
    expect(
      uploadPayload?.code,
      `动态插件上传接口应返回成功: ${uploadPayload?.message ?? ""}`,
    ).toBe(0);

    await expect(this.uploadSuccessDialog()).toBeVisible();
    const successPattern =
      expectedSuccessText ??
      /插件包上传成功|Plugin package uploaded successfully/iu;
    await expect(this.uploadSuccessDialog()).toContainText(successPattern);
    await expect(this.dynamicUploadConfirmButton()).toContainText(
      /知道了|Got It/iu,
    );
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
    const installButton = await this.pluginActionButton(
      pluginId,
      pluginInstallActionPattern,
    );
    await expect(installButton).toBeVisible();
    await installButton.click();
    await expect(this.hostServiceAuthDialog()).toBeVisible();
    await this.confirmHostServiceAuthorization();
    await expect(
      await this.pluginActionButton(pluginId, /卸\s*载/),
    ).toBeVisible();
  }

  async installAndEnablePlugin(pluginId: string) {
    const installButton = await this.pluginActionButton(
      pluginId,
      pluginInstallActionPattern,
    );
    await expect(installButton).toBeVisible();
    await installButton.click();
    await this.confirmInstallAndEnable();
  }

  async ensurePluginInstalled(pluginId: string) {
    const installButton = await this.pluginActionButton(
      pluginId,
      pluginInstallActionPattern,
    );
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
    const installButton = await this.pluginActionButton(
      pluginId,
      pluginInstallActionPattern,
    );
    await expect(installButton).toBeVisible();
    await installButton.click();
    await expect(this.hostServiceAuthModal()).toBeVisible();
  }

  async selectInstallMode(modeLabel: string | RegExp) {
    await this.pluginInstallModeSelect().locator(".ant-select-selector").click();
    const option = this.page
      .locator(".ant-select-dropdown:visible .ant-select-item-option")
      .filter({ hasText: modeLabel })
      .last();
    await expect(option).toBeVisible();
    await option.click();
    await expect(this.pluginInstallModeSelect()).toContainText(modeLabel);
  }

  async expectInstallModeDescriptionAfterSelect() {
    const selectBox = await this.pluginInstallModeSelect().boundingBox();
    const descriptionBox =
      await this.pluginInstallModeDescription().boundingBox();
    expect(selectBox).not.toBeNull();
    expect(descriptionBox).not.toBeNull();
    expect(descriptionBox!.x).toBeGreaterThan(selectBox!.x + selectBox!.width);
  }

  async expectInstallModeSectionDashedBorder() {
    await expect(this.pluginInstallModeSection()).toHaveCSS(
      "border-top-style",
      "dashed",
    );
  }

  async expectInstallModeDescriptionWithoutBorder() {
    await expect(this.pluginInstallModeDescription()).toHaveCSS(
      "border-top-width",
      "0px",
    );
  }

  async expectInstallModePlatformOnlyHintGap() {
    await expect(this.pluginInstallModeSection()).toHaveCSS("gap", "12px");
  }

  async installPluginWithMockData(pluginId: string, withMockData: boolean) {
    const installButton = await this.pluginActionButton(
      pluginId,
      pluginInstallActionPattern,
    );
    await expect(installButton).toBeVisible();
    await installButton.click();
    await expect(this.hostServiceAuthDialog()).toBeVisible();
    if (withMockData) {
      await expect(this.pluginInstallMockDataSection()).toBeVisible();
      const checkbox = this.pluginInstallMockDataCheckbox();
      const isChecked = await checkbox.isChecked();
      if (!isChecked) {
        await checkbox.check();
      }
      await expect(checkbox).toBeChecked();
    } else {
      // Even when the plugin ships mock data, the checkbox should default to
      // unchecked so a forgetful click does not bring demo rows into the table.
      const sectionVisible = await this.pluginInstallMockDataSection()
        .isVisible({ timeout: 1500 })
        .catch(() => false);
      if (sectionVisible) {
        const checkbox = this.pluginInstallMockDataCheckbox();
        const isChecked = await checkbox.isChecked();
        if (isChecked) {
          await checkbox.uncheck();
        }
        await expect(checkbox).not.toBeChecked();
      }
    }
    await this.confirmHostServiceAuthorization();
    await expect(
      await this.pluginActionButton(pluginId, /卸\s*载/),
    ).toBeVisible();
  }

  async uninstallPlugin(pluginId: string) {
    await this.uninstallPluginWithOptions(pluginId, true);
  }

  async openUninstallDialog(pluginId: string) {
    const uninstallButton = await this.pluginActionButton(
      pluginId,
      pluginUninstallActionPattern,
    );
    await expect(uninstallButton).toBeVisible();
    await uninstallButton.click();
    await expect(this.uninstallDialog()).toBeVisible();
  }

  async cancelUninstallDialog() {
    await this.uninstallDialog()
      .getByRole("button", { name: cancelActionPattern })
      .last()
      .click();
    await expect(this.uninstallDialog()).toHaveCount(0);
  }

  async ensurePluginUninstalled(pluginId: string) {
    const uninstallButton = await this.pluginActionButton(
      pluginId,
      pluginUninstallActionPattern,
    );
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
    // Prefer stable test id when present; fall back to action-column text match.
    const byTestId = this.pluginDetailAction(pluginId);
    const detailButton = (await byTestId.count()) > 0
      ? byTestId
      : await this.pluginActionButton(pluginId, pluginDetailActionPattern);
    await expect(detailButton).toBeVisible();
    await detailButton.click();
    await expect(this.pluginDetailModal()).toBeVisible();
  }

  /**
   * Opens the plugin detail modal by clicking a non-interactive cell in the row
   * (name cell), not the action-column Detail button.
   */
  async openPluginDetailByRowClick(pluginId: string) {
    await this.ensurePluginRowVisible(pluginId);
    const nameCell = this.pluginNameCell(pluginId);
    await expect(nameCell).toBeVisible();
    await nameCell.click();
    await expect(this.pluginDetailModal()).toBeVisible();
  }

  async expectPluginRowClickableCursor(pluginId: string) {
    const row = this.pluginRow(pluginId);
    await expect(row).toBeVisible();
    await expect(row).toHaveCSS("cursor", "pointer");
  }

  /**
   * Asserts every label cell in a plugin Descriptions table stays on one line
   * (no multi-line wrap of field names such as "Authorization Status").
   */
  private async expectDescriptionsLabelsNoWrap(
    labels: Locator,
    contextLabel: string,
  ) {
    await expect(labels.first()).toBeVisible();
    const count = await labels.count();
    expect(count, `${contextLabel}标签列应至少包含一个字段`).toBeGreaterThan(0);

    for (let index = 0; index < count; index += 1) {
      const label = labels.nth(index);
      await expect(
        label,
        `${contextLabel}第 ${index + 1} 个标签应禁止换行`,
      ).toHaveCSS("white-space", "nowrap");
      await expect
        .poll(
          async () =>
            await label.evaluate((element) => {
              // With wrap, multi-line labels grow scrollHeight; with nowrap they stay one line.
              return element.scrollHeight <= element.clientHeight + 1;
            }),
          {
            message: `${contextLabel}第 ${index + 1} 个标签内容应保持单行`,
          },
        )
        .toBe(true);
    }
  }

  async expectPluginDetailLabelsNoWrap() {
    await this.expectDescriptionsLabelsNoWrap(
      this.pluginDetailDescriptionLabels(),
      "插件详情",
    );
  }

  async expectPluginInstallLabelsNoWrap() {
    await this.expectDescriptionsLabelsNoWrap(
      this.pluginInstallDescriptionLabels(),
      "插件安装",
    );
  }

  async openRuntimeUpgradeDialog(pluginId: string) {
    const upgradeButton = await this.pluginActionButton(
      pluginId,
      pluginUpgradeActionPattern,
    );
    await expect(upgradeButton).toBeVisible();
    await upgradeButton.click();
    await expect(this.pluginUpgradeModal()).toBeVisible();
  }

  async confirmRuntimeUpgrade() {
    await expect(this.pluginUpgradeConfirmButton()).toBeEnabled();
    await this.pluginUpgradeConfirmButton().click();
    await expect(this.pluginUpgradeModal()).toHaveCount(0);
  }

  async uninstallPluginWithOptions(
    pluginId: string,
    purgeStorageData: boolean,
  ) {
    const uninstallButton = await this.pluginActionButton(
      pluginId,
      pluginUninstallActionPattern,
    );
    await expect(uninstallButton).toBeVisible();
    await uninstallButton.click();
    await expect(this.uninstallDialog()).toBeVisible();
    const checkboxVisible = await this.uninstallPurgeCheckboxWrapper()
      .isVisible({ timeout: 1500 })
      .catch(() => false);
    if (checkboxVisible) {
      await expect(this.uninstallPurgeWarning()).toBeVisible();
      const isChecked = await this.uninstallPurgeCheckbox().isChecked();
      if (isChecked !== purgeStorageData) {
        await this.uninstallPurgeCheckbox().click();
      }
    }
    const uninstallResponse = this.page.waitForResponse(
      (response) => {
        const request = response.request();
        return (
          request.method() === "DELETE" &&
          new URL(response.url()).pathname.endsWith(`/plugins/${pluginId}`)
        );
      },
      { timeout: pluginLifecycleActionTimeout },
    );
    await this.uninstallDialog()
      .getByRole("button", { name: confirmActionPattern })
      .last()
      .click();
    const response = await uninstallResponse;
    expect(response.ok(), `uninstall ${pluginId} should return 2xx`).toBe(true);
    await expect(this.uninstallDialog()).toHaveCount(0, {
      timeout: pluginLifecycleActionTimeout,
    });
    await expect(
      await this.pluginActionButton(pluginId, pluginInstallActionPattern),
    ).toBeVisible({ timeout: pluginLifecycleActionTimeout });
  }

  async openUninstallDialogAndConfirm(pluginId: string) {
    await this.openUninstallDialog(pluginId);
    await this.uninstallDialog()
      .getByRole("button", { name: confirmActionPattern })
      .last()
      .click();
  }

  async setPluginEnabled(pluginId: string, enabled: boolean) {
    for (let attempt = 1; attempt <= 4; attempt += 1) {
      const row = await this.ensurePluginRowVisible(pluginId);
      const switcher = this.pluginLifecycleSwitch(row);
      await expect(
        switcher,
        `未找到插件启用状态开关: ${pluginId}`,
      ).toBeVisible();
      const isChecked = (await switcher.getAttribute("aria-checked")) === "true";
      if (isChecked === enabled) {
        return;
      }

      const className = (await switcher.getAttribute("class")) ?? "";
      if (className.includes("ant-switch-disabled")) {
        if (attempt < 4) {
          await this.filterByPluginId(pluginId);
          await waitForRouteReady(this.page, 15000);
          continue;
        }
        await expect(switcher).not.toHaveClass(/ant-switch-disabled/, {
          timeout: pluginLifecycleActionTimeout,
        });
      }

      const actionPath = enabled ? "enable" : "disable";
      const statusResponse = this.page.waitForResponse(
        (response) => {
          const request = response.request();
          return (
            request.method() === "PUT" &&
            new URL(response.url()).pathname.endsWith(
              `/plugins/${pluginId}/${actionPath}`,
            )
          );
        },
        { timeout: pluginLifecycleActionTimeout },
      );
      const pluginStateRefresh = this.page
        .waitForResponse(
          (response) => {
            const request = response.request();
            return (
              request.method() === "GET" &&
              new URL(response.url()).pathname.endsWith(
                "/plugins/dynamic/state",
              )
            );
          },
          { timeout: pluginLifecycleRefreshProbeTimeout },
        )
        .catch(() => null);
      const menuRefresh = this.page
        .waitForResponse(
          (response) => {
            const request = response.request();
            return (
              request.method() === "GET" &&
              new URL(response.url()).pathname.endsWith("/menus/all")
            );
          },
          { timeout: pluginLifecycleRefreshProbeTimeout },
        )
        .catch(() => null);
      await switcher.click();
      if (enabled) {
        const authDialogVisible = await this.hostServiceAuthDialog()
          .isVisible({ timeout: 1500 })
          .catch(() => false);
        if (authDialogVisible) {
          await this.confirmHostServiceAuthorization();
        }
      }
      const response = await statusResponse;
      expect(
        response.ok(),
        `${actionPath} ${pluginId} should return 2xx`,
      ).toBe(true);
      await Promise.all([pluginStateRefresh, menuRefresh]);
      await this.filterByPluginId(pluginId);
      const refreshedSwitcher = this.pluginLifecycleSwitch(
        this.pluginRow(pluginId),
      );
      await expect(refreshedSwitcher).toHaveAttribute(
        "aria-checked",
        enabled ? "true" : "false",
        { timeout: pluginLifecycleActionTimeout },
      );
      await this.page
        .getByText(
          enabled
            ? /插件已启用|Plugin enabled/i
            : /插件已禁用|Plugin disabled/i,
        )
        .last()
        .waitFor({ state: "visible", timeout: 3000 })
        .catch(() => undefined);
      await this.page
        .getByText(/加载菜单中|Loading Menu/i)
        .last()
        .waitFor({ state: "hidden", timeout: pluginLifecycleActionTimeout })
        .catch(() => undefined);
      return;
    }
  }

  async cancelManagedActionWarning() {
    await expect(this.pluginManagedActionDialog()).toBeVisible();
    await this.pluginManagedActionDialog()
      .getByRole("button", { name: cancelActionPattern })
      .last()
      .click();
    await expect(this.pluginManagedActionDialog()).toHaveCount(0);
  }

  async confirmManagedActionWarning() {
    await expect(this.pluginManagedActionDialog()).toBeVisible();
    await this.pluginManagedActionDialog()
      .getByRole("button", {
        name: /继续停用|继续禁用|继续卸载|Continue|confirm|ok|确\s*认|确\s*定/iu,
      })
      .last()
      .click();
    await expect(this.pluginManagedActionDialog()).toHaveCount(0);
  }

  async expectInstallActionVisible(pluginId: string) {
    await expect(
      await this.pluginActionButton(pluginId, pluginInstallActionPattern),
    ).toBeVisible();
  }

  async expectInstallActionHidden(pluginId: string) {
    await expect(
      await this.pluginActionButton(pluginId, pluginInstallActionPattern),
    ).toHaveCount(0);
  }

  async expectUpgradeActionHidden(pluginId: string) {
    await expect(
      await this.pluginActionButton(pluginId, pluginUpgradeActionPattern),
    ).toHaveCount(0);
  }

  async expectUninstallActionVisible(pluginId: string) {
    await expect(
      await this.pluginActionButton(pluginId, pluginUninstallActionPattern),
    ).toBeVisible();
  }

  async expectUninstallActionHidden(pluginId: string) {
    await expect(
      await this.pluginActionButton(pluginId, pluginUninstallActionPattern),
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
    const confirmResponse = this.waitForLifecycleActionResponse();
    await this.hostServiceAuthConfirmButton().click();
    await confirmResponse;
    await expect(this.hostServiceAuthDialog()).toHaveCount(0);
  }

  async confirmInstallAndEnable() {
    await expect(this.hostServiceAuthDialog()).toBeVisible();
    await expect(this.hostServiceAuthInstallAndEnableButton()).toBeVisible();
    const confirmResponse = this.waitForLifecycleActionResponse();
    await this.hostServiceAuthInstallAndEnableButton().click();
    await confirmResponse;
    await expect(this.hostServiceAuthDialog()).toHaveCount(0);
  }

  private waitForLifecycleActionResponse() {
    return this.page
      .waitForResponse(
        (response) => {
          const request = response.request();
          const method = request.method();
          const pathname = new URL(response.url()).pathname;
          return (
            (method === "POST" || method === "PUT") &&
            (/\/plugins\/[^/]+\/(install|enable)$/.test(pathname) ||
              /\/plugins\/install$/.test(pathname))
          );
        },
        { timeout: pluginLifecycleActionTimeout },
      )
      .catch(() => null);
  }

  private async pluginActionButton(pluginId: string, name: RegExp) {
    const row = await this.ensurePluginRowVisible(pluginId);

    const rowID = await row.getAttribute("rowid");
    expect(rowID, `未找到插件行 rowid: ${pluginId}`).toBeTruthy();
    return this.page
      .locator(
        `.vxe-table--fixed-right-wrapper .vxe-body--row[rowid=\"${rowID}\"]`,
      )
      .getByRole("button", { name })
      .first();
  }

  async expectSidebarMenuVisible(menuName: SidebarMenuName) {
    for (let attempt = 1; attempt <= 3; attempt += 1) {
      const menuItem = this.sidebarMenuItem(menuName);
      if (await menuItem.isVisible().catch(() => false)) {
        await expect(menuItem).toBeVisible();
        return menuItem;
      }

      const parentSubmenu = this.sidebarSubmenuForMenuItem(menuName);
      if (await parentSubmenu.isVisible({ timeout: 1000 }).catch(() => false)) {
        await this.expandSidebarSubmenu(parentSubmenu);
      }

      if (!(await menuItem.isVisible({ timeout: 1000 }).catch(() => false))) {
        await this.expandExtensionCenterIfVisible();
      }

      if (!(await menuItem.isVisible({ timeout: 1000 }).catch(() => false))) {
        const refreshedParentSubmenu = this.sidebarSubmenuForMenuItem(menuName);
        if (
          await refreshedParentSubmenu
            .isVisible({ timeout: 1000 })
            .catch(() => false)
        ) {
          await this.expandSidebarSubmenu(refreshedParentSubmenu);
        }
      }

      if (!(await menuItem.isVisible({ timeout: 1000 }).catch(() => false))) {
        const pluginManageMenu = this.sidebarMenu
          .getByText(pluginManageMenuPattern, { exact: true })
          .first();
        if (
          await pluginManageMenu.isVisible({ timeout: 1000 }).catch(() => false)
        ) {
          await pluginManageMenu.click();
        }
      }

      if (await menuItem.isVisible({ timeout: 1000 }).catch(() => false)) {
        await expect(menuItem).toBeVisible();
        return menuItem;
      }

      if (attempt < 3) {
        await this.page.reload({ waitUntil: "domcontentloaded" });
        await waitForRouteReady(this.page, 15000);
      }
    }

    const menuItem = this.sidebarMenuItem(menuName);
    await expect(menuItem).toBeVisible();
    return menuItem;
  }

  async expectSidebarMenuHidden(menuName: SidebarMenuName) {
    await this.expandExtensionCenterIfVisible();
    const visible = await this.sidebarMenuItem(menuName)
      .isVisible({ timeout: 1500 })
      .catch(() => false);
    expect(visible).toBeFalsy();
  }

  async gotoWorkspace() {
    await this.page.goto(workspacePath("/dashboard/workspace"));
    await expect(
      this.page.getByTestId("dashboard-workspace-page"),
    ).toBeVisible();
  }

  async expectTableColumnVisible(title: string) {
    await expect(this.tableColumn(title)).toBeVisible();
  }

  async expectTableColumnHidden(title: string) {
    await expect(this.tableColumn(title)).toHaveCount(0);
  }

  async expectTableColumnBetween(
    targetTitle: string | string[],
    previousTitle: string | string[],
    nextTitle: string | string[],
  ) {
    const headerTitles = (
      await this.page
        .locator(".vxe-table--header .vxe-cell--title")
        .allTextContents()
    )
      .map((title) => title.trim())
      .filter(Boolean);

    const findHeaderIndex = (titleOrTitles: string | string[]) => {
      const titles = Array.isArray(titleOrTitles) ? titleOrTitles : [titleOrTitles];
      return headerTitles.findIndex((title) => titles.includes(title));
    };
    const formatTitle = (titleOrTitles: string | string[]) =>
      Array.isArray(titleOrTitles) ? titleOrTitles.join(" / ") : titleOrTitles;

    const targetIndex = findHeaderIndex(targetTitle);
    const previousIndex = findHeaderIndex(previousTitle);
    const nextIndex = findHeaderIndex(nextTitle);
    const targetLabel = formatTitle(targetTitle);
    const previousLabel = formatTitle(previousTitle);
    const nextLabel = formatTitle(nextTitle);

    expect(targetIndex, `未找到列表列: ${targetLabel}`).toBeGreaterThanOrEqual(
      0,
    );
    expect(
      previousIndex,
      `未找到列表列: ${previousLabel}`,
    ).toBeGreaterThanOrEqual(0);
    expect(nextIndex, `未找到列表列: ${nextLabel}`).toBeGreaterThanOrEqual(0);
    expect(
      targetIndex,
      `${targetLabel} 应位于 ${previousLabel} 之后`,
    ).toBeGreaterThan(previousIndex);
    expect(targetIndex, `${targetLabel} 应位于 ${nextLabel} 之前`).toBeLessThan(
      nextIndex,
    );
  }

  async expectTableColumnAfter(targetTitle: string, previousTitle: string) {
    const headerTitles = (
      await this.page
        .locator(".vxe-table--header .vxe-cell--title")
        .allTextContents()
    )
      .map((title) => title.trim())
      .filter(Boolean);

    const targetIndex = headerTitles.indexOf(targetTitle);
    const previousIndex = headerTitles.indexOf(previousTitle);

    expect(targetIndex, `未找到列表列: ${targetTitle}`).toBeGreaterThanOrEqual(
      0,
    );
    expect(
      previousIndex,
      `未找到列表列: ${previousTitle}`,
    ).toBeGreaterThanOrEqual(0);
    expect(
      targetIndex,
      `${targetTitle} 应位于 ${previousTitle} 之后`,
    ).toBeGreaterThan(previousIndex);
  }

  async expectTableColumnOrder(expectedTitles: string[]) {
    const headerTitles = (
      await this.page
        .locator(".vxe-table--header .vxe-cell--title")
        .allTextContents()
    )
      .map((title) => title.trim())
      .filter(Boolean);

    expect(
      headerTitles.slice(0, expectedTitles.length),
      "插件管理列表前置列顺序不符合预期",
    ).toEqual(expectedTitles);
  }

  async expectTableColumnAligned(title: string, expectedAlign: string) {
    const cell = this.tableHeaderCell(title);
    await expect(cell, `未找到列表列: ${title}`).toBeVisible();
    await expect
      .poll(
        async () =>
          await cell.evaluate(
            (element) => globalThis.getComputedStyle(element).textAlign,
          ),
        { message: `${title} 列标题应${expectedAlign}对齐` },
      )
      .toBe(expectedAlign);
  }

  async expectTableColumnBodyAligned(title: string, expectedAlign: string) {
    const headerTitles = (
      await this.page
        .locator(".vxe-table--header .vxe-cell--title")
        .allTextContents()
    )
      .map((headerTitle) => headerTitle.trim())
      .filter(Boolean);
    const columnIndex = headerTitles.indexOf(title);

    expect(columnIndex, `未找到列表列: ${title}`).toBeGreaterThanOrEqual(0);

    const bodyCell = this.page
      .locator(".vxe-table--main-wrapper .vxe-body--row")
      .first()
      .locator(".vxe-body--column")
      .nth(columnIndex);
    await expect(bodyCell, `未找到列表单元格: ${title}`).toBeVisible();
    await expect
      .poll(
        async () =>
          await bodyCell.evaluate(
            (element) => globalThis.getComputedStyle(element).textAlign,
          ),
        { message: `${title} 列内容应${expectedAlign}对齐` },
      )
      .toBe(expectedAlign);
  }

  async expectTableColumnLeftAligned(title: string) {
    await this.expectTableColumnBodyAligned(title, "left");
  }

  async expectTableColumnCentered(title: string) {
    await this.expectTableColumnAligned(title, "center");
  }

  async expectTableColumnWiderThan(
    widerTitle: string,
    narrowerTitles: string[],
  ) {
    const widerCell = this.tableHeaderCell(widerTitle);
    await expect(widerCell, `未找到列表列: ${widerTitle}`).toBeVisible();
    const widerWidth = await widerCell.evaluate(
      (element) => element.getBoundingClientRect().width,
    );

    for (const narrowerTitle of narrowerTitles) {
      const narrowerCell = this.tableHeaderCell(narrowerTitle);
      await expect(narrowerCell, `未找到列表列: ${narrowerTitle}`).toBeVisible();
      const narrowerWidth = await narrowerCell.evaluate(
        (element) => element.getBoundingClientRect().width,
      );
      expect(
        widerWidth,
        `${widerTitle} 列宽应大于 ${narrowerTitle}`,
      ).toBeGreaterThan(narrowerWidth);
    }
  }

  async expectTableColumnWidthAtMost(title: string, maxWidth: number) {
    const cell = this.tableHeaderCell(title);
    await expect(cell, `未找到列表列: ${title}`).toBeVisible();
    const width = await cell.evaluate(
      (element) => element.getBoundingClientRect().width,
    );
    expect(width, `${title} 列宽不应超过 ${maxWidth}px`).toBeLessThanOrEqual(
      maxWidth,
    );
  }

  /**
   * Measures the fixed action column via a body cell (header cells for
   * fixed-right columns are marked fixed--hidden in the main header table).
   */
  async expectPluginActionColumnWidthAtMost(
    pluginId: string,
    maxWidth: number,
  ) {
    const detail = this.pluginDetailAction(pluginId);
    await expect(detail).toBeVisible();
    const width = await detail.evaluate((element) => {
      const cell = element.closest(".vxe-body--column");
      return cell instanceof HTMLElement
        ? cell.getBoundingClientRect().width
        : 0;
    });
    expect(
      width,
      `插件 ${pluginId} 操作列宽不应超过 ${maxWidth}px（实际 ${width}px）`,
    ).toBeLessThanOrEqual(maxWidth);
  }

  /**
   * Asserts the action-column buttons for a plugin row stay on a single line
   * by comparing vertical baselines of the always-visible detail/manage
   * buttons and any sibling ghost buttons in the same cell.
   */
  async expectPluginActionButtonsSingleLine(pluginId: string) {
    const detail = this.pluginDetailAction(pluginId);
    const manage = this.pluginManageAction(pluginId);
    await expect(detail).toBeVisible();
    await expect(manage).toBeVisible();

    await expect
      .poll(
        async () =>
          await detail.evaluate((detailNode, manageTestId) => {
            const cell = detailNode.closest(".vxe-body--column");
            if (!(cell instanceof HTMLElement)) {
              return false;
            }
            const manageNode = cell.querySelector(
              `[data-testid="${manageTestId}"]`,
            );
            if (!(manageNode instanceof HTMLElement)) {
              return false;
            }
            const buttons = Array.from(
              cell.querySelectorAll("button.ant-btn"),
            ).filter((node): node is HTMLElement => node instanceof HTMLElement);
            if (buttons.length < 2) {
              return false;
            }
            const tops = buttons.map((button) => button.getBoundingClientRect().top);
            const baseTop = tops[0] ?? 0;
            // Wrapped buttons jump by roughly a full control height (~24px+).
            return tops.every((top) => Math.abs(top - baseTop) <= 2);
          }, `plugin-manage-button-${pluginId}`),
        {
          message: `插件 ${pluginId} 操作列按钮应保持单行不换行`,
        },
      )
      .toBe(true);
  }

  async expectPluginVersionNotClipped(pluginId: string) {
    const version = this.pluginVersionValue(pluginId);
    await expect(version).toBeVisible();
    await expect
      .poll(
        async () =>
          await version.evaluate((element) => {
            const cell = element.closest(".vxe-body--column");
            if (!cell) {
              return false;
            }
            const elementRect = element.getBoundingClientRect();
            const cellRect = cell.getBoundingClientRect();
            const horizontalPadding = 2;
            const contentFits = element.scrollWidth <= element.clientWidth + 1;
            return (
              contentFits &&
              elementRect.left >= cellRect.left - horizontalPadding &&
              elementRect.right <= cellRect.right + horizontalPadding
            );
          }),
        { message: `插件 ${pluginId} 的版本号内容不应被单元格裁剪` },
      )
      .toBe(true);
  }

  async expectBooleanTableCell(
    cell: Locator,
    expected: boolean,
  ) {
    await expect(cell).toBeVisible();
    await expect(cell).toContainText(expected ? /是|Yes/iu : /否|No/iu);
  }

  async expectTenantProvisioningDisabled(pluginId: string) {
    await expect(this.pluginTenantProvisioningSwitch(pluginId)).toHaveClass(
      /ant-switch-disabled/,
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
}
