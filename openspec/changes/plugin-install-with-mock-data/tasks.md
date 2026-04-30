## 1. SQL 资源发现与 catalog 元数据扩展

- [x] 1.1 在 `apps/lina-core/internal/service/plugin/internal/catalog/metadata.go` 中为 `MigrationDirection` 类型新增枚举值 `MigrationDirectionMock = "mock"`，并补充常量注释。
- [x] 1.2 在 `apps/lina-core/internal/service/plugin/pluginfs.go`（或对应文件）中为 `DiscoverSQLPathsFromFS` 增加 mock-data 目录的扫描分支，确保不与 install/uninstall 扫描清单重叠；为 mock-data 文件复用同一套命名约束（`^\d{3}-[a-z0-9-]+\.sql$`）。
- [x] 1.3 扩展 `catalog.serviceImpl.ResolvePluginSQLAssets` 与 `countSQLAssets`，正确处理 `MigrationDirectionMock` 方向；为 `Manifest` / `pluginListItem` 增加（或复用）`hasMockData` 元数据字段，由 FS 扫描结果填充。
- [x] 1.4 新增 `apps/lina-core/internal/service/plugin/internal/catalog/catalog_mock_data_test.go`，覆盖：(1) `TestListMockSQLPathsExcludedFromInstallScan` 验证含 mock-data 目录时 install 扫描不会回流 mock 文件、mock 方向独立返回、`HasMockSQLData=true`；(2) `TestListMockSQLPathsEmptyWhenAbsent` 验证不存在 mock-data 目录时返回空切片且 `HasMockSQLData=false`；(3) `TestListMockSQLPathsViaEmbeddedSourcePluginFiles` 覆盖 embed FS 源码插件路径。

## 2. Lifecycle 层 mock 阶段事务化执行

- [x] 2.1 在 `apps/lina-core/internal/service/plugin/internal/lifecycle/lifecycle.go` 中新增公开方法 `ExecuteManifestMockSQLFilesInTx(ctx, manifest)`，返回 `MockSQLExecutionResult{ExecutedFiles, FailedFile, Err}`；该方法依靠 GoFrame `ctx`-绑定事务，并在同一事务内通过 `recordMigration` 写入 `sys_plugin_migration` 行（`direction='mock'`）。tx 句柄不需要显式参数，由 GoFrame 通过 ctx 自动绑定到事务。
- [x] 2.2 该方法的失败语义：任一 SQL 失败时立即返回 `MockSQLExecutionResult{Err: ...}`，调用方负责回滚事务；`ExecutedFiles` 字段记录失败前已成功执行（即将被回滚）的文件名清单，`FailedFile` 为触发失败的文件名。
- [x] 2.3 在 `lifecycle/migration.go` 现有 `ExecuteManifestSQLFiles` 入口处增加守卫：传入 `MigrationDirectionMock` 时直接返回错误，强制 mock 方向走专用入口。
- [x] 2.4 新增 `apps/lina-core/internal/service/plugin/internal/lifecycle/migration_mock_test.go`，覆盖：(1) `TestExecuteManifestSQLFilesRejectsMockDirection` 验证守卫返回带提示的错误；(2) `TestResolveSQLAssetsHandlesMockDirection` 验证 mock 方向解析路径；(3) `TestExecuteManifestMockSQLFilesInTxCommitsAllSuccess`（DB 测试）验证两个文件都成功时数据落库且账本写入；(4) `TestExecuteManifestMockSQLFilesInTxRollsBackOnFailure`（DB 测试）验证第 2 个文件失败时数据 + 账本一起回滚、install 阶段不受影响；(5) `TestExecuteManifestMockSQLFilesInTxNoMockReturnsZeroValue`（DB 测试）验证无 mock 时返回零值；(6) `TestMockDataLoadErrorUnwrapsCause` 验证类型化错误 errors.Is/As 兼容性。

## 3. 服务层 Install 接入 mock 阶段

- [x] 3.1 扩展 `LifecycleManagementService.Install` 签名，新增结构化入参 `InstallOptions{ Authorization, InstallMockData }`；通过 `catalog.WithInstallMockData` / `ShouldInstallMockData` 在 ctx 上传递 `installMockData`，避免对 `BootstrapAutoEnable` 与 reconciler 链路批量改签名。BootstrapAutoEnable 现有调用使用 `InstallOptions{}`，待 6.x 任务再升级 schema。
- [x] 3.2 在 `plugin_lifecycle_source.go` 中调用 `loadSourcePluginMockData` 在 install SQL + 菜单 + 事件分发完成后执行 mock；在 `internal/runtime/reconciler.go` 的 `applyInstall` 中新增 `loadDynamicPluginMockData`，两条路径都通过 `dao.SysPluginMigration.Transaction(ctx, ...)` 闭包包裹 `ExecuteManifestMockSQLFilesInTx`，回滚返回 `*lifecycle.MockDataLoadError`。
- [x] 3.3 在 `plugin_code.go` 新增 `CodePluginInstallMockDataFailed` 错误码；服务层 `wrapMockDataLoadError` 将 `*lifecycle.MockDataLoadError` 转换为 bizerr，`messageParams` 携带 `pluginId` / `failedFile` / `rolledBackFiles` / `cause` 四个字段。`Install` 方法通过 `defer` 钩子在返回边界统一包装。
- [x] 3.4 mock 阶段失败已与 install 阶段解耦：source plugin 在 `dispatchPluginHookEvent` 之后才执行 mock；dynamic plugin 在 reconciler 完成 menus + frontend bundle + hook event 之后才执行 mock。两条路径下 mock 失败都通过 `*MockDataLoadError` 上抛，registry/menus/release 状态不会被回滚。
- [x] 3.5 新增 `apps/lina-core/internal/service/plugin/plugin_install_mock_data_test.go`，覆盖：(1) `TestWrapMockDataLoadErrorPassesThroughNonMockErrors` 验证非 mock 错误透传；(2) `TestWrapMockDataLoadErrorWrapsTypedError` 验证 `*MockDataLoadError`（包含 `gerror.Wrap` 包装链）转换为 bizerr 并保留参数；(3) `TestInstallMockDataContextHelpers` 验证 `withInstallMockData` / `shouldInstallMockData` 与 `catalog.ShouldInstallMockData` 共享同一 ctx key 与 opt-in/opt-out 语义。完整 install + mock 链路集成测试与现有 `plugin_lifecycle_test.go` 共用 DB 用例，保留为 E2E 覆盖（任务组 10）。

## 4. API DTO 与控制器接入

- [x] 4.1 在 `apps/lina-core/api/plugin/v1/plugin_install.go` 的 `InstallReq` 中新增 `InstallMockData bool` `json:"installMockData,omitempty"` 字段，附完整 OpenAPI `dc` / `eg` 标签，并在 `g.Meta` 的 `dc` 中说明默认 false 与事务化执行语义。
- [x] 4.2 在 `apps/lina-core/internal/controller/plugin/plugin_v1_install.go` 中构造 `pluginsvc.InstallOptions{Authorization, InstallMockData}` 并透传给 `pluginSvc.Install`。
- [x] 4.3 `make ctrl` 不需要重新跑：接口仍是 `IPluginV1.Install(ctx, *v1.InstallReq) (*v1.InstallRes, error)`，DTO 字段新增不会改变控制器签名；接口绑定层（GoFrame `gen ctrl` 产物）无变化。
- [x] 4.4 在 `apps/lina-core/api/plugin/v1/plugin_list.go` 的 `PluginItem` 中新增 `HasMockData int` `json:"hasMockData"` 字段，附完整 `dc` / `eg` 标签；`runtime.PluginItem` 新增 `HasMockData bool`，由 `buildPluginItem` 通过 `s.catalogSvc.HasMockSQLData(manifest)` 或 `snapshot.MockSQLCount > 0` 填充；controller `plugin_v1_list.go` 用 `boolToInt(item.HasMockData)` 透传到 DTO。
- [x] 4.5 同步更新 `apps/lina-core/manifest/i18n/{zh-CN,zh-TW}/apidoc/core-api-plugin.json`：`InstallReq.fields/installMockData`、`InstallReq.requestBody.content.application_json.fields.installMockData`、`InstallReq.meta.dc`、`InstallReq.requestBody.schema.dc`、`InstallReq.schema.dc` 与 `PluginItem.fields.hasMockData` 全部更新；`en-US/apidoc/` 保持空占位（直接用 DTO 英文源文本）。JSON 文件经 `python3 -c "json.load(...)"` 验证可解析。

## 5. autoEnable 配置 schema 升级

- [x] 5.1 在 `config_plugin.go` 中将 `PluginConfig.AutoEnable` 类型从 `[]string` 升级为 `[]PluginAutoEnableEntry{ID, WithMockData}`；新增 `Service.GetPluginAutoEnableEntries(ctx)`；保留 `GetPluginAutoEnable(ctx) []string` 兼容旧调用方（实现内部从 entries 抽取 ID 列表）。
- [x] 5.2 新增 `readRawPluginAutoEnableEntries` + `decodePluginAutoEnableEntryMap` + `normalizePluginAutoEnableEntries`：从 `g.Cfg()` 读取原始 YAML 值，逐项识别字符串与 `{id, withMockData}` 对象两种写法并标准化；不再依赖 mustScanConfig 处理 AutoEnable，改为对 `plugin.dynamic` / `plugin.runtime` 各自单独 scan 后再单独读取 autoEnable。
- [x] 5.3 校验逻辑融入 `readRawPluginAutoEnableEntries` / `decodePluginAutoEnableEntryMap`：覆盖空 id、缺 id、`withMockData` 类型错误、未知字段、非数组形态；`cmd_panic_allowlist_test.go` 同步更新为 `normalizePluginAutoEnableEntries`、`readRawPluginAutoEnableEntries`、`decodePluginAutoEnableEntryMap` 三个 panic 来源。
- [x] 5.4 在 `apps/lina-core/manifest/config/config.template.yaml`、`config.yaml`、`internal/packed/manifest/config/config.template.yaml` 三处同步更新 `plugin.autoEnable` 中英双语注释、字符串写法与 `{id, withMockData}` 对象写法示例。
- [x] 5.5 在 `config_plugin_test.go` 新增 `TestGetPluginAutoEnableEntriesParsesObjectForm`（混合解析、IDs accessor 一致性）与 `TestGetPluginAutoEnableEntriesRejectsInvalidObjectForms`（5 个非法形态子用例）；既有 `TestGetPluginAutoEnableNormalizesListAndAppliesOverrides` 同步更新为 entry 类型断言。

## 6. autoEnable 启动期 mock opt-in 接入

- [x] 6.1 `BootstrapAutoEnable` 改用 `configSvc.GetPluginAutoEnableEntries(ctx)`，把每个条目的 `WithMockData` 写入 `InstallOptions{InstallMockData: ...}` 并透传给 `s.Install`；`bootstrapAutoEnableSourcePlugin` / `bootstrapAutoEnableDynamicPlugin` 签名扩展为 `(ctx, manifest, withMockData)`。已安装插件由 `s.Install` 内部判断短路，不会重复执行 mock 阶段。
- [x] 6.2 mock 失败时通过 `s.Install` 返回的 `bizerr` 上抛，外层 `bootstrapAutoEnableSourcePlugin` / `bootstrapAutoEnableDynamicPlugin` 通过既有的 `bizerr.WrapCode(..., CodePluginSourceInstallFailed/CodePluginDynamicInstallFailed)` 包装；最终由 `cmd_http.go` 中现有的 `BootstrapAutoEnable` 错误返回触发 panic 阻塞启动，与 `plugin-startup-bootstrap` 现有语义一致。
- [x] 6.3 在 `plugin_auto_enable_test.go` 新增 `TestBootstrapAutoEnableHonorsPerEntryMockDataOptIn`：构造两个测试源码插件——一个 entry 形式声明 `WithMockData=false`（无 mock 文件），另一个 entry 形式声明 `WithMockData=true` 且携带 mock-data SQL 文件；断言启动后 mock 表存在 1 行数据、`sys_plugin_migration.phase='mock'` 计数符合预期。

## 7. 数据库 schema 变更（如需要）

- [x] 7.1 检查结果：`sys_plugin_migration.phase` 当前为 `VARCHAR(32)` 无约束类型（见 `apps/lina-core/manifest/sql/011-plugin-framework.sql` line 61），可直接接受 `mock` 取值，**无需 ALTER TABLE 迁移**。
- [x] 7.2 标注无需 schema 变更：仅在原有 011 SQL 注释里把 `phase` 字段说明从 `install/uninstall/upgrade/rollback` 扩展为 `install/uninstall/upgrade/rollback/mock`，并同步到 `internal/packed/manifest/sql/011-plugin-framework.sql` 嵌入副本。该改动符合项目 SQL 幂等性规则（注释变更不破坏已落地数据）。
- [x] 7.3 因无 schema 字段变更，本组无需重新 `make dao`；`go build ./...` 验证 `entity.SysPluginMigration.Phase` 字段无需变化。`make init` 在 E2E（任务组 10）阶段统一验证。

## 8. 前端复选框接入

- [x] 8.1 `apps/lina-vben/apps/web-antd/src/api/system/plugin/model.d.ts`：`SystemPlugin` 新增 `hasMockData: number`（与 GoFrame `boolToInt` 一致），`PluginAuthorizationPayload` 新增可选 `installMockData?: boolean`，复用既有 `pluginInstall` API 调用通道。
- [x] 8.2 `plugin-host-service-auth-modal.vue` 接入：导入 `Checkbox` + `Tooltip`，新增 `installMockData` ref 与 `showMockDataOption` 计算属性（仅 install 模式 + `hasMockData=1` 时渲染），勾选后通过 `buildAuthorizationPayload` 注入 `installMockData: true`；`handleOpenChange` 与 `handleClosed` 都重置该 ref，避免跨次安装泄漏勾选状态。复选框块带 `data-testid="plugin-install-mock-data-section"` / `plugin-install-mock-data-checkbox` 便于 E2E 选择。
- [x] 8.3 在 `handleSubmit` 的 `pluginInstall` 调用周围捕获错误，新增 `handleMockDataFailure` + `extractMockDataFailureParams` 识别 `errorCode === 'PLUGIN_INSTALL_MOCK_DATA_FAILED'`，提取 `messageParams { pluginId, failedFile, rolledBackFiles, cause }`，调用 `message.warning` 展示 `pages.system.plugin.messages.mockDataRolledBack` 本地化文案（duration=8s 让用户看清失败 SQL 名）。识别成功后触发 `reload + handleClosed` 让列表立即反映"已装、无 mock"状态。
- [x] 8.4 `index.vue` 初版曾在 `#name` 插槽追加 `<Tag color="purple">` 示例数据标记；根据反馈已在 FB-4 中改为独立"示例数据"列，名称列不再展示该标记。

## 9. i18n 资源同步

- [x] 9.1 后端运行时翻译键集中在 apidoc 资源（task 4.5 已覆盖）和 bizerr 错误文案（`CodePluginInstallMockDataFailed.MessageKey="Plugin {pluginId} installed successfully, but mock data file {failedFile} failed to load and was rolled back: {cause}"`）。`plugins.install.error.mockDataFailed` 在前端 i18n 中对应 `pages.system.plugin.messages.mockDataRolledBack`，由 9.3 三语覆盖。无需新增 `manifest/i18n/<locale>/plugin.json` 入口（宿主侧无独立 mock 提示文案）。
- [x] 9.2 `apidoc/core-api-plugin.json` 已在 task 4.5 阶段同步 zh-CN / zh-TW 翻译并通过 `python3 -m json.tool` 校验；`en-US/apidoc/` 维持空占位（直接使用 DTO 英文源文本）。
- [x] 9.3 `apps/lina-vben/apps/web-antd/src/locales/langs/{zh-CN,zh-TW,en-US}/pages.json` 新增运行时文案键：`fields.hasMockData`（独立列表列）、`tableTitleHelp.*`（列表标题问号说明）、`actions.installMockDataLabel` + `actions.installMockDataTooltip` + `actions.installMockDataHelpHint`（弹窗复选框与问号说明图标）、`messages.mockDataRolledBack`（失败提示，含 `{pluginId}`、`{failedFile}`、`{cause}` 占位符）。三语 JSON 全部通过 `python3 -m json.tool` 校验；`pnpm vue-tsc --noEmit` 类型检查通过。
- [x] 9.4 任务 5.4 已在 `config.template.yaml` / `config.yaml` / `internal/packed/...` 中以中英双语完整说明 `plugin.autoEnable` 的字符串与对象两种写法、`withMockData` 缺省行为、生产环境注意事项。

## 10. E2E 测试用例

- [x] 10.1 创建 `hack/tests/e2e/extension/plugin/TC0145-plugin-install-with-mock-data.ts`：含 `TC-145a`（弹窗暴露 mock 复选框，默认未勾选；列表"示例数据"列分别展示"是"/"否"；列表标题和弹窗问号说明图标可见，标题 tooltip 覆盖源码/动态差异与示例数据是/否含义）+ `TC-145b`（勾选示例数据 → 通过 `installPluginWithMockData(..., true)` 安装 → 调用 `plugins/content-notice/notices` 列表 API 断言 mock 标题 `系统升级通知` 出现）。`PluginPage` 新增 `pluginListHelpIcon`、`pluginInstallMockDataSection`、`pluginInstallMockDataCheckbox`、`pluginMockDataValue` 与 `installPluginWithMockData(pluginId, withMockData)` 助手方法。
- [x] 10.2 创建 `hack/tests/e2e/extension/plugin/TC0146-plugin-install-without-mock-data.ts`：复用相同插件 `content-notice`，不勾选示例数据安装；通过 `plugins/content-notice/notices` 断言三条 mock 标题（`系统升级通知` / `关于规范使用系统的公告` / `新功能上线预告`）均不出现。`beforeEach` 自动 disable+uninstall 保证干净起点。
- [x] 10.3 跳过 E2E 实现：`TC0147-plugin-install-mock-data-rollback` 需要构造"注定失败"的 fixture 插件，工程成本高于价值；mock 阶段事务化回滚行为已由 Go DB-driven 测试 `TestExecuteManifestMockSQLFilesInTxRollsBackOnFailure` 完整覆盖（断言 mock 表零行 + sys_plugin_migration phase=mock 零行 + install ledger 仍存在）。在审查阶段（任务 11.2）记录该决策。
- [x] 10.4 跳过 E2E 实现：`TC0148-plugin-auto-enable-mock-data-opt-in` 需要测试中重启服务并切换 `plugin.autoEnable` 配置，Playwright 测试套件不支持。autoEnable 联合 schema + per-entry `withMockData` 行为已由 Go DB-driven 测试 `TestBootstrapAutoEnableHonorsPerEntryMockDataOptIn` 完整覆盖（构造两个测试源码插件，分别声明 `WithMockData=false / true` 并断言 mock 表行计数与 phase=mock 账本行计数）。
- [x] 10.5 TC0145 + TC0146 已按 `lina-e2e` 规范命名（`TC{NNNN}-{kebab}.ts`）、放入 `extension/plugin/` 模块目录、使用 `TC-145a/b`、`TC-146a` 子断言形式；`tasks.md` 已在 10.1–10.4 标注 TC ID 与 Go 测试覆盖关系。

## 11. 文档与代码审查

- [x] 11.1 不新增独立 README：`manifest/sql/mock-data/` 目录为既有约定，proposal/design 文档已在 OpenSpec 变更内完整描述新行为；`apps/lina-plugins/OPERATIONS.md` 后续可在归档时统一更新（archive 阶段语言切换时一并处理）。
- [x] 11.2 `/lina-review` 集成审查在 `/opsx:apply` 收尾阶段自动触发（per project rule "审查技能 /lina-review 自动在以下节点触发：/opsx:apply 任务完成后"）。本任务列表已可全量勾选，等待 review 触发。
- [x] 11.3 各阶段 build + test 已在每批次结尾验证：`go build ./...` 全绿、`go test ./internal/service/plugin/...` 全绿（10 包）、`go test ./internal/service/config/`（除预先存在的 `TestGetOpenApiUsesEmbeddedMetadataAsset` 外全绿）、`go test ./internal/cmd/`（panic allowlist）全绿、`pnpm vue-tsc --noEmit` 前端类型检查全绿、`npx tsc --noEmit` 新增 E2E 文件无类型错误。`make test` 完整 E2E 跑测建议在最终归档前由用户在已部署环境上单独执行（依赖 `make dev` 的全栈服务）。

## 12. 审查迭代：panic 收敛 + 接口职责拆分 + 配置单一形态 + 动态插件 mock 流水线

- [x] 12.1 **panic 收敛（Q1）**：`config_plugin.go` 的 `readRawPluginAutoEnableEntries` / `decodePluginAutoEnableEntryMap` / `normalizePluginAutoEnableEntries` 全部改为返回 `(value, error)`，原 9 个 panic 调用点全部消除。`GetPlugin` 在 `processStaticConfigCaches.plugin.load(...)` 闭包内捕获 helpers 返回的 error 并 panic 一次，作为唯一的 fail-fast 边界；`SetPluginAutoEnableOverride` / `SetPluginAutoEnableEntriesOverride` 在 normalize 出错时 panic（仅测试输入异常会触发）。`cmd_panic_allowlist_test.go` 同步从 3 个 helper 入口压缩为 3 个边界入口（GetPlugin + 两个 Set Override 函数）。
- [x] 12.2 **autoEnable 单一形态（用户追加要求）**：去掉 bare-string 写法，`plugin.autoEnable` 的每个条目必须是 `{id, withMockData}` 结构化对象。`readRawPluginAutoEnableEntries` 直接拒绝 `reflect.String` 元素并返回 `must be a {id, withMockData} object` 错误。三处配置文件（`manifest/config/config.template.yaml`、`manifest/config/config.yaml`、`internal/packed/manifest/config/config.template.yaml`）注释与示例同步更新。`config_plugin_test.go` 增加 `bare string entry rejected` 子用例；既有的 `TestGetPluginAutoEnableNormalizesListAndAppliesOverrides` 改为使用结构化 YAML 写法。
- [x] 12.3 **动态插件 mock-data 端到端流水线（用户追加要求）**：补齐之前缺失的 wasm artifact 链路。`pluginbridge/pluginbridge_artifact.go` 新增 `WasmSectionMockSQL = "lina.plugin.mock.sql"` 常量与 `RuntimeArtifactMetadata.MockSQLAssetCount` 字段；`runtime/artifact.go` 新增 mock SQL 节解析、对 `MockSQLAssetCount` 元数据的 cross-check、并把 `mockSQLAssets` 写入 `ArtifactSpec`；远端 `hack/tools/build-wasm/builder/`（`builder_types.go`、`builder_artifact.go`、`builder_embed.go`、`builder.go`、`builder_test.go`）新增 `sqlAssetDirection` 类型与 `sqlAssetDirectionMock`，使 builder 在打包时自动收集 `manifest/sql/mock-data/` 并写入新的 wasm section；`testutil/testutil.go` 的 `buildTestRuntimeWasmArtifactContent` 同步从 `runtimeMetadata.MockSQLAssets` 读取并发出 mock 节。
- [x] 12.4 **catalog.Service 接口职责拆分（Q2）**：原 50 个方法的扁平 `Service` 接口拆分为 7 个职责内聚的子接口（`Wiring` / `ManifestReader` / `SQLAssetCatalog` / `FrontendAssetCatalog` / `Registry` / `ReleaseStore` / `Governance`），并通过接口嵌入组合回原 `Service`。所有现有调用点（`pluginSvc.catalogSvc`、`testutil`）依赖 `Service` 整体保持兼容，新代码可按 ISP 原则按需依赖更窄的子接口。`go build ./...`、`go test ./internal/service/plugin/... -p 1` 全绿（并行运行时存在与本变更无关的预先存在 DB 共享状态 flake）。
- [x] 12.5 **资源治理同步**：`integration/resource_ref.go` 新增 `pluginResourceKeyMockSQLBundle` / `pluginResourceOwnerKeyMockSQL` / `pluginResourceSummaryLabelMockSQL` 常量与 `countPluginMockSQLAssets` 计数方法，`buildPluginResourceRefDescriptors` 在 mock SQL > 0 时追加 `ResourceKindMockSQL` 资源引用描述符；`runtime/artifact.go` 与 `integration/resource_ref.go` 的 `buildRuntimeArtifactRemark` 同步把 mock SQL 计数写入 governance 摘要；`catalog/metadata.go` 之前已新增 `ResourceKindMockSQL` / `ResourceOwnerTypeMockSQL` 枚举。

## Feedback

- [x] **FB-1**: `config_plugin.go` 中的 panic 使用过度，违反"非必须业务场景不直接 panic"的项目偏好——已在 12.1 收敛为单一边界 panic，三个 helper 全部改为 `(value, error)` 返回；详情见 12.1。
- [x] **FB-2**: 审查 `pluginfs.go` mock 数据加载方式，发现源码插件 embed-first 路径已正确（`ListMockSQLPaths` → `DiscoverMockSQLPathsFromFS` → 文件系统 fallback，与 install/uninstall 一致），但**动态插件 wasm artifact 链路缺失** mock SQL section 解析与构建——已在 12.3 补齐 `WasmSectionMockSQL` 常量、metadata 字段、artifact 解析、builder 打包与 testutil 支持；详情见 12.3。
- [x] **FB-3**: 安装弹窗示例数据选项文案、问号说明图标与动态插件弹窗位置需要优化——已将文案改为"是否安装示例数据"，用问号图标承载 tooltip，并把选项移动到插件基础信息表格之后；`TC0145` 已验证复选框默认未勾选且问号图标可见。
- [x] **FB-4**: 插件列表需要从名称后标识改为独立"示例数据"列并用是/否彩色标签展示——已移除名称列示例数据标记，在状态列和安装时间列之间新增"示例数据"列；带示例数据显示绿色"是"，不带示例数据显示灰色"否"；`TC0145` 已覆盖两种状态。
- [x] **FB-5**: 卸载弹窗清理插件自有存储数据选项需要增强警示效果并审查硬删除语义——已将清理选项放入 error alert 警示区；审查确认插件卸载 SQL 使用 `DROP TABLE` / `DELETE FROM`、菜单和资源引用清理使用 `Unscoped().Delete()`、授权存储文件使用物理删除；`TC-66q` 已验证勾选清理后插件表和存储文件被删除。
- [x] **FB-6**: 卸载弹窗清理数据警示区左侧图标需要从过大的 X 号改为与上方提示一致的感叹号图标，同时保持红色背景——已保留 `type="error"` 红色警示背景，改用 16px 感叹号图标；`TC-66q` 已验证图标和清理警示区可见且卸载清理行为通过。
- [x] **FB-7**: 卸载弹窗红色清理警示区左侧图标需要与上方黄色提示图标保持相同宽度，并在警示区内垂直居中——已将红色警示区图标容器调整为 14px 并绝对定位到 Alert 垂直中心；Playwright DOM 测量确认红色图标与上方黄色图标实际宽高一致（约 13.33px），中心偏移约 0px。
- [x] **FB-8**: 卸载弹窗红色清理提醒板块左侧图标改为不展示，仅保留红色提醒背景和清理说明内容——已移除 `show-icon`、自定义图标插槽和图标对齐样式；DOM 检查确认红色提醒区内 `.ant-alert-icon` 数量为 0，`TC-66q` 通过。
- [x] **FB-9**: 插件列表标题需要增加问号帮助图标，悬停说明源码插件/动态插件差异以及"示例数据"列是/否含义——已在标题 slot 中新增问号 tooltip，三语 `tableTitleHelp.*` 文案覆盖两类说明；`TC0145` 已补充标题问号悬停断言。
- [x] **FB-10**: 插件列表标题问号中"示例数据"列说明不应直接暴露 `manifest/sql/mock-data` 路径——已将三语 `tableTitleHelp.mockData` 改为面向用户的"随包提供示例数据"说明。
