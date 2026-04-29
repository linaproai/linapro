## Why

插件作者在 `manifest/sql/mock-data/` 目录下随插件交付了一份用于演示和功能验证的示例数据，但目前**安装流程完全不会执行这些 SQL 文件**——它们只能依赖宿主的 `make mock` 命令加载，而这个命令针对的是宿主自身的 mock 数据，并不会跑插件目录里的 mock SQL。结果就是：用户安装一个新插件后，UI 上只有空表，无法直观地理解和试用该插件的功能；测试同学也缺少一组开箱即用的演示数据来快速验证插件是否正常运行。

## What Changes

- **插件安装请求新增 `installMockData` 选项**：手动安装路径（`POST /plugins/{id}/install`）的请求 DTO 增加 `installMockData bool` 字段，由用户在前端"安装"按钮的弹窗中通过复选框勾选；勾选即视为"知情同意"加载示例数据，无须额外栅栏。
- **Mock SQL 全成全败语义**：当 `installMockData=true` 时，宿主会在 install SQL 全部执行完成后，将插件 `manifest/sql/mock-data/*.sql` 的所有文件以及对应的迁移账本写入操作**包裹在同一个数据库事务中**执行；任一文件失败时整体回滚，保证插件不会停留在"加载到一半"的脏状态。Install SQL 自身因含 DDL 在 MySQL 上无法事务化，仍按现状逐文件执行并依赖幂等保证。
- **失败响应携带可操作信息**：mock 阶段失败时，响应里返回失败 SQL 文件名、失败原因，以及"已自动回滚 mock 数据"的明确提示；用户可以选择接受当前"已装、无 mock"状态，或修复 SQL 后卸载重装。
- **覆盖源码插件与动态插件**：源码插件直接读 embed FS；动态插件通过既有的 SQL 资源发现机制读取制品包内的 `manifest/sql/mock-data/`，两类插件 UX 一致。
- **`autoEnable` 配置升级为联合类型**：`plugin.autoEnable` 兼容两种写法——纯字符串 `"plugin-id"` 与对象 `{id: "plugin-id", withMockData: true}`；默认（仅字符串、或 `withMockData` 缺省）**不**加载示例数据，仅在显式声明 `withMockData: true` 时才加载。已有配置全部使用字符串形式，无需改动即可平滑升级。
- **迁移账本扩展 `phase=mock` 阶段值**：`sys_plugin_migration` 现有的迁移方向枚举新增第三个取值 `mock`，记录每个 mock SQL 文件的执行结果；事务回滚时账本同步回滚，运维侧可据此追溯"哪些插件加载了示例数据"。
- **i18n 三语完整覆盖**：新增的复选框文案、说明性提示、失败错误码消息、配置文件注释，按治理规范同步 `zh-CN`、`zh-TW`、`en-US` 三套 i18n 资源。
- **E2E 用例覆盖**：新增 `hack/tests/e2e/system/plugin/` 下的测试用例，覆盖"勾选安装 mock"、"不勾选不安装 mock"、"mock SQL 失败时全部回滚"、"`autoEnable` opt-in 启动后看到 mock 数据"四类场景。

## Capabilities

### New Capabilities

- `plugin-mock-data-installation`：定义插件安装时可选加载示例数据的能力——包括 SQL 资源发现规则、单一事务执行语义、失败回滚行为、迁移账本写入与失败响应载荷。

### Modified Capabilities

- `plugin-manifest-lifecycle`：插件 SQL 资源发现规则扩展，明确 `manifest/sql/mock-data/` 在 SQL 资产层级中的位置和命名约束；动态插件制品打包必须沿用相同目录结构。
- `plugin-startup-bootstrap`：`plugin.autoEnable` 配置 schema 升级为字符串与对象两种写法的联合类型；启动期自动安装时，仅当显式配置 `withMockData: true` 才加载 mock 数据，且 mock 阶段失败被视为启动失败（与现有"启动期任意失败 panic"语义一致）。

## Impact

- **后端**
  - `apps/lina-core/api/plugin/v1/`：`InstallReq` 新增 `installMockData` 字段（含完整 OpenAPI 标签）。
  - `apps/lina-core/internal/service/plugin/`：`Install()` 方法签名扩展、`plugin_lifecycle_source.go` 与 `plugin_lifecycle.go` 接入新阶段、`BootstrapAutoEnable()` 解析配置 opt-in。
  - `apps/lina-core/internal/service/plugin/lifecycle/`：`MigrationDirection` 枚举新增 `mock`、`ResolveSQLAssets` / `ExecuteManifestSQLFiles` 增加 mock 扫描入口与事务化执行入口。
  - `apps/lina-core/internal/service/plugin/pluginfs.go`：`DiscoverSQLPathsFromFS` 暴露 mock-data 目录的扫描方法。
  - `apps/lina-core/internal/service/config/config_plugin.go`：`PluginConfig.AutoEnable` 升级为联合类型并补 `validatePluginAutoEnableRawValue` 校验分支。
  - `apps/lina-core/pkg/bizerr/`：在插件模块的 `*_code.go` 中新增 `plugin.install.mockDataFailed` 错误码。
- **前端**
  - `apps/lina-vben/apps/web-antd/src/views/system/plugin/plugin-host-service-auth-modal.vue`：增加"安装示例数据"复选框（带说明 tooltip），勾选后随安装请求发送 `installMockData: true`。
  - `apps/lina-vben/apps/web-antd/src/api/system/plugin/`：`pluginInstall` API 类型签名扩展。
- **配置文件**
  - `apps/lina-core/manifest/config/config.template.yaml`：`plugin.autoEnable` 注释新增对象写法的示例与说明。
- **i18n 资源**
  - 宿主 `apps/lina-core/manifest/i18n/{zh-CN,zh-TW,en-US}/menu.json`、`framework.json`、错误码消息文件等，新增复选框文案、错误提示等键值。
  - 宿主 `apps/lina-core/manifest/i18n/{zh-CN,zh-TW}/apidoc/**` 同步新增字段翻译；`en-US/apidoc/` 维持空占位（直接用 DTO 英文源文本）。
- **数据库迁移**
  - `apps/lina-core/manifest/sql/`：新增本迭代 SQL 文件，确保 `sys_plugin_migration.direction` 字段允许 `mock` 取值（如使用 ENUM/CHAR 限制则需扩展约束）。
- **E2E 测试**
  - `hack/tests/e2e/system/plugin/`：4 个新的 TC 用例。
- **不影响**：插件运行时 host service、菜单/规范同步、动态插件 wasm 加载流程不需要改动。
