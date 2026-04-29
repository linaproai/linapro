## 1. SQL 资源发现与 catalog 元数据扩展

- [ ] 1.1 在 `apps/lina-core/internal/service/plugin/internal/catalog/metadata.go` 中为 `MigrationDirection` 类型新增枚举值 `MigrationDirectionMock = "mock"`，并补充常量注释。
- [ ] 1.2 在 `apps/lina-core/internal/service/plugin/pluginfs.go`（或对应文件）中为 `DiscoverSQLPathsFromFS` 增加 mock-data 目录的扫描分支，确保不与 install/uninstall 扫描清单重叠；为 mock-data 文件复用同一套命名约束（`^\d{3}-[a-z0-9-]+\.sql$`）。
- [ ] 1.3 扩展 `catalog.serviceImpl.ResolvePluginSQLAssets` 与 `countSQLAssets`，正确处理 `MigrationDirectionMock` 方向；为 `Manifest` / `pluginListItem` 增加（或复用）`hasMockData` 元数据字段，由 FS 扫描结果填充。
- [ ] 1.4 为 1.1–1.3 的所有改动补 catalog 单元测试（`catalog_test.go` 中现有 `installAssets` / `uninstallAssets` 测试附近），覆盖："只有 install 文件无 mock 时返回空 mock 资产"、"含 mock-data 目录时只在 mock 方向返回该目录文件"、"hasMockData 元数据正确"等场景。

## 2. Lifecycle 层 mock 阶段事务化执行

- [ ] 2.1 在 `apps/lina-core/internal/service/plugin/internal/lifecycle/lifecycle.go` 中新增公开方法 `ExecuteManifestMockSQLFilesInTx(ctx, tx gdb.TX, manifest *catalog.Manifest) (rolledBackFiles []string, failedFile string, err error)`；该方法仅在调用方传入的事务句柄内顺序执行 mock 资产，并在同一事务内写入 `sys_plugin_migration` 行（`direction='mock'`）。
- [ ] 2.2 该方法的失败语义：任一 SQL 失败时立即返回错误，调用方负责回滚事务；返回值 `rolledBackFiles` 包含失败前已成功执行（即将被回滚）的文件名清单，`failedFile` 为触发失败的文件名。
- [ ] 2.3 在 `lifecycle/migration.go`（或对应文件）的现有 `ExecuteManifestSQLFiles` 入口处明确不接受 `MigrationDirectionMock` 参数（mock 方向必须走专用入口），返回明确错误以避免误用。
- [ ] 2.4 为 2.1–2.3 增加单元测试（参考 `migration_test.go`），覆盖：mock SQL 全部成功提交、第二个文件失败时整体回滚（数据库无残留）、空 mock-data 目录下方法返回零值且不写账本。

## 3. 服务层 Install 接入 mock 阶段

- [ ] 3.1 在 `apps/lina-core/internal/service/plugin/plugin.go` 中扩展 `LifecycleManagementService.Install` 接口签名，增加结构化入参（如 `InstallOptions{ InstallMockData bool }`）；同时调整 `BootstrapAutoEnable` 内部实现以支持按条目传入 `withMockData`。
- [ ] 3.2 在 `apps/lina-core/internal/service/plugin/plugin_lifecycle_source.go` 与 `plugin_lifecycle.go` 两类插件分支中接入：install SQL 阶段保持不变；install 成功后若 `installMockData=true` 则调用 `dao.SysPluginMigration.Transaction(ctx, ...)` 闭包包裹 `ExecuteManifestMockSQLFilesInTx`，事务回滚后将 `rolledBackFiles` / `failedFile` / cause 透传给 `bizerr` 包装层。
- [ ] 3.3 在插件模块的 `*_code.go` 中新增错误码常量 `CodeMockDataFailed`（错误码字符串 `plugin.install.mockDataFailed`，i18n key `plugins.install.error.mockDataFailed`），并在服务层使用 `bizerr.WrapCode` 包装 mock 阶段错误，`messageParams` 携带 `pluginId` / `failedFile` / `rolledBackFiles` / `cause` 四个字段。
- [ ] 3.4 在服务层（或调用层）确保 mock 阶段失败时插件已完成的 install/菜单同步等阶段成果不被回滚，即响应失败但插件仍处于"已安装、无 mock"状态——通过单元测试断言此行为。
- [ ] 3.5 为 `Install` 链路的 mock 接入新增/补充服务层单元测试（参考 `plugin_lifecycle_test.go`），覆盖：`installMockData=false` 不扫描 mock-data；`installMockData=true` 全部成功；mock 阶段失败时 install 状态保留 + 错误码正确 + `rolledBackFiles` 非空。

## 4. API DTO 与控制器接入

- [ ] 4.1 在 `apps/lina-core/api/plugin/v1/install.go`（或同等文件）的 `InstallReq` 中新增字段 `InstallMockData bool` `json:"installMockData"`，附完整 OpenAPI 标签（`dc` 英文说明 + `eg` 示例值），并明确"默认 false，仅在用户显式勾选时为 true"的语义。
- [ ] 4.2 在 `apps/lina-core/internal/controller/plugin/plugin_v1_install.go` 中将 `req.InstallMockData` 透传给 `pluginSvc.Install`；保持其他字段（如 `Authorization`）行为不变。
- [ ] 4.3 通过 `make ctrl` 重新生成接口骨架；手工合入业务逻辑（已经在 4.2 完成）。
- [ ] 4.4 在 `pluginListItem` / 插件详情接口响应 DTO 中增加 `HasMockData bool` `json:"hasMockData"` 字段（在 1.3 已经准备好元数据），附 `dc` / `eg` 标签；用于前端决定是否展示复选框。
- [ ] 4.5 同步更新对应宿主 apidoc 翻译资源（`apps/lina-core/manifest/i18n/{zh-CN,zh-TW}/apidoc/plugin/**`），按治理规范确保英文源文本变化后非英文翻译不缺失；`en-US/apidoc/` 维持空占位。

## 5. autoEnable 配置 schema 升级

- [ ] 5.1 在 `apps/lina-core/internal/service/config/config_plugin.go` 中将 `PluginConfig.AutoEnable` 字段类型从 `[]string` 升级为可承载混合条目的内部结构（如 `[]PluginAutoEnableEntry{ ID string; WithMockData bool }`）；继续提供 `GetPluginAutoEnable(ctx) []string` 兼容旧接口供其他模块（如 controller `BuildAutoEnableManagedSet`）使用，并新增 `GetPluginAutoEnableEntries(ctx)` 返回完整结构。
- [ ] 5.2 重写或扩展 `normalizePluginAutoEnableIDs` 为 `normalizePluginAutoEnableEntries`，在配置加载阶段把 YAML 中的字符串与对象元素统一标准化为内部结构；缺省 `withMockData` 取 `false`。
- [ ] 5.3 扩展 `validatePluginAutoEnableRawValue`，覆盖联合类型的合法/非法条目用例；同步更新 `cmd_panic_allowlist_test.go` 中的 panic allowlist。
- [ ] 5.4 在 `apps/lina-core/manifest/config/config.template.yaml` 与 `config.yaml` 中更新 `plugin.autoEnable` 注释与示例条目，展示字符串与对象两种写法。
- [ ] 5.5 为 5.1–5.3 增加单元测试，覆盖：纯字符串列表解析、纯对象列表解析、混合解析、`{withMockData=true}` 条目、非法条目（空 id / 错误类型 / 缺 id）等。

## 6. autoEnable 启动期 mock opt-in 接入

- [ ] 6.1 在 `apps/lina-core/internal/service/plugin/plugin_auto_enable.go` 的 `BootstrapAutoEnable` 实现中读取新结构条目，对每个条目调用 3.1 扩展后的 `Install` 时按 `withMockData` 设置 `InstallMockData`；已安装插件不重复执行 mock。
- [ ] 6.2 mock 阶段失败时遵循 `plugin-startup-bootstrap` 现有的"任一失败 panic 阻塞启动"语义：通过 `panic` 抛出包含 pluginID、failedFile、cause 的错误信息。
- [ ] 6.3 为 `BootstrapAutoEnable` 增加针对 mock opt-in 的单元测试（在 `plugin_auto_enable_test.go` 中扩展），覆盖：字符串条目不加载 mock；`{withMockData=true}` 条目加载 mock 并成功；mock 失败时 `BootstrapAutoEnable` 返回错误并包含失败 SQL 文件信息。

## 7. 数据库 schema 变更（如需要）

- [ ] 7.1 检查 `sys_plugin_migration.direction` 字段当前定义类型；如为受约束类型（ENUM/CHECK）则在新增 `apps/lina-core/manifest/sql/<NNN>-plugin-install-with-mock-data.sql` 中通过 `ALTER TABLE` 扩展约束允许 `mock` 取值，确保 SQL 幂等可重入。
- [ ] 7.2 如果当前为 `VARCHAR` 等无约束类型，则 7.1 不需要 SQL 改动；在 tasks.md 进度中明确标注"无需 schema 变更"。
- [ ] 7.3 运行 `make init` 验证新 SQL 文件可重复执行；若有 schema 变更则同步运行 `make dao` 更新 `dao/do/entity`。

## 8. 前端复选框接入

- [ ] 8.1 在 `apps/lina-vben/apps/web-antd/src/api/system/plugin/` 中扩展 `pluginInstall` 的请求类型，新增可选字段 `installMockData?: boolean`；扩展 plugin 列表/详情类型新增 `hasMockData: boolean`。
- [ ] 8.2 在 `apps/lina-vben/apps/web-antd/src/views/system/plugin/plugin-host-service-auth-modal.vue` 中：增加"安装示例数据"复选框（含 i18n label 与 tooltip），仅当 `props.row.hasMockData === true` 时渲染；勾选后随提交携带 `installMockData: true`。
- [ ] 8.3 实现 mock 阶段失败的前端错误处理：识别 errorCode `plugin.install.mockDataFailed`，在弹窗内（或全局提示）展示包含 `pluginId`、`failedFile`、`rolledBackFiles`、`cause` 参数的本地化提示；提示明确告知"插件已安装，示例数据已自动回滚"。
- [ ] 8.4 在 `apps/lina-vben/apps/web-antd/src/views/system/plugin/index.vue` 的列表渲染中按需展示 mock 数据相关辅助信息（如 hasMockData 标识），避免影响现有 UI 节奏。

## 9. i18n 资源同步

- [ ] 9.1 在 `apps/lina-core/manifest/i18n/{zh-CN,zh-TW,en-US}/menu.json`、`framework.json`、`plugin.json` 等相应文件中新增/更新键值：复选框 label `plugins.install.mockDataLabel`、tooltip `plugins.install.mockDataTooltip`、错误提示 `plugins.install.error.mockDataFailed` 等，三语完整覆盖。
- [ ] 9.2 在 `apps/lina-core/manifest/i18n/{zh-CN,zh-TW}/apidoc/plugin/**.json` 中同步 4.1/4.4 中新增的 DTO 字段中文/繁中翻译；`en-US/apidoc/` 不写英文映射，保持空占位。
- [ ] 9.3 检查并更新 `apps/lina-vben` 前端运行时 i18n 资源（如有，相应位置 `packages/locales` 或对应 web-antd 资源），确保前端 UI 文案三语一致。
- [ ] 9.4 在 `apps/lina-core/manifest/config/config.template.yaml` 的注释中以中英双语形式说明 `plugin.autoEnable` 的两种写法与 `withMockData` 缺省行为。

## 10. E2E 测试用例

- [ ] 10.1 创建 `hack/tests/e2e/extension/plugin/TC0145-plugin-install-with-mock-data.ts`：登录 → 打开内容公告插件安装弹窗 → 勾选"安装示例数据" → 安装成功 → 进入对应插件页面验证 mock 数据可见。
- [ ] 10.2 创建 `hack/tests/e2e/extension/plugin/TC0146-plugin-install-without-mock-data.ts`：相同插件 → 不勾选示例数据 → 安装成功 → 进入对应插件页面验证表为空（无 mock 数据）。
- [ ] 10.3 创建 `hack/tests/e2e/extension/plugin/TC0147-plugin-install-mock-data-rollback.ts`：构造一个 mock SQL 注定失败的测试插件（或在测试 fixture 中注入失败注入），勾选示例数据安装 → 断言响应错误码 `plugin.install.mockDataFailed`、断言 mock 数据全部回滚（DB 中插件 mock 表为空）、断言插件仍处于已安装状态可正常使用。
- [ ] 10.4 创建 `hack/tests/e2e/extension/plugin/TC0148-plugin-auto-enable-mock-data-opt-in.ts`：在测试 fixture 中将某插件以 `{id, withMockData: true}` 配置写入 autoEnable → 重启或重置测试环境 → 启动后断言对应 mock 数据已加载；同时验证另一字符串条目对应的插件 mock 表为空。
- [ ] 10.5 在 `tasks.md` 进度中关联以上 4 个 TC ID（TC0145–TC0148），按 `lina-e2e` 技能规范登记到对应模块目录与文件结构中。

## 11. 文档与代码审查

- [ ] 11.1 在 `apps/lina-plugins/<plugin-id>/manifest/sql/mock-data/` 的 README（如不存在则不强行创建）或 `OPERATIONS.md` 中说明新行为，必要时补 zh/en 双语版本。
- [ ] 11.2 通过 `lina-review` 技能进行实现层与规范层的最终审查（`/opsx:apply` 完成后自动触发）。
- [ ] 11.3 完成后视情况运行 `make test`（前后端 E2E）确认相关用例与回归通过。
