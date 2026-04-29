## ADDED Requirements

### Requirement: 插件安装请求必须支持可选的示例数据加载选项

宿主在手动安装路径（`POST /plugins/{id}/install`）的请求 DTO 中 SHALL 提供 `installMockData bool` 字段，默认值为 `false`。当请求体显式设置 `installMockData=true` 时，宿主 SHALL 在 install SQL 全部执行成功后，扫描该插件 `manifest/sql/mock-data/` 目录中的 SQL 文件并按文件名顺序加载执行；当字段未设置或为 `false` 时，宿主 MUST 完全跳过 mock-data 目录的扫描与执行。前端 plugin 安装弹窗 SHALL 提供"安装示例数据"复选框，仅在该插件存在 `manifest/sql/mock-data/` 目录（即 `hasMockData=true`）时显示，复选框默认未勾选；勾选行为本身即视为用户对加载示例数据的知情同意，前端与后端 MUST NOT 引入额外的环境栅栏、独立权限或二次确认步骤。

#### Scenario: 用户勾选复选框成功安装示例数据
- **WHEN** 用户在插件安装弹窗中勾选"安装示例数据"复选框并提交
- **AND** 该插件 `manifest/sql/mock-data/` 目录下存在合法的 SQL 文件
- **AND** install SQL 全部执行成功
- **THEN** 宿主依次执行 mock-data 目录下的 SQL 文件
- **AND** 全部成功后插件被标记为已安装，请求返回成功响应

#### Scenario: 用户未勾选复选框时跳过示例数据
- **WHEN** 用户在插件安装弹窗中未勾选"安装示例数据"复选框并提交
- **AND** 该插件 `manifest/sql/mock-data/` 目录下存在合法的 SQL 文件
- **THEN** 宿主完全不扫描、不执行该目录下的任何 SQL 文件
- **AND** 数据库中不会出现任何 mock 数据行
- **AND** 请求返回成功响应且响应不含 mock 相关字段

#### Scenario: 插件不存在 mock-data 目录时复选框不可见
- **WHEN** 前端打开某插件的安装弹窗
- **AND** 插件元数据接口返回 `hasMockData=false`
- **THEN** "安装示例数据"复选框 MUST 不渲染到 DOM 中
- **AND** 即便用户在请求体里手动添加 `installMockData=true`，后端也不会找到 mock-data 目录而执行任何 SQL 文件

### Requirement: Mock SQL 与对应迁移账本写入必须包裹在同一数据库事务中

当 `installMockData=true` 触发 mock 阶段时，宿主 SHALL 通过 `dao.SysPluginMigration.Transaction(ctx, ...)` 闭包在**单一数据库事务**内顺序执行所有 mock SQL 文件，并在闭包内完成对 `sys_plugin_migration` 表的执行结果写入（每个文件一行，`direction='mock'`）。任意一个 mock SQL 文件执行失败时，宿主 MUST 让闭包返回错误以触发整个事务回滚，使**已经执行的 mock SQL 与已经写入的账本行同时回滚**，最终数据库中没有任何由本次 mock 阶段产生的数据行或账本行。Install SQL 阶段、菜单同步阶段、插件注册阶段 MUST NOT 与 mock 事务共享同一事务边界。

#### Scenario: Mock 阶段任一 SQL 失败导致全部回滚
- **WHEN** 用户在安装弹窗中勾选"安装示例数据"
- **AND** install SQL 全部执行成功
- **AND** mock-data 目录下有 3 个 SQL 文件，前 2 个执行成功，第 3 个执行时数据库返回错误
- **THEN** 宿主整体回滚 mock 事务
- **AND** 数据库中不存在前 2 个 SQL 已经写入的任何 mock 数据行
- **AND** `sys_plugin_migration` 表中不存在任何 `direction='mock'` 的记录
- **AND** 插件仍处于"已安装、无 mock"状态（install SQL 与菜单同步等成果保留）

#### Scenario: Mock 阶段全部成功的事务提交
- **WHEN** 用户勾选"安装示例数据"
- **AND** mock-data 目录下的所有 SQL 文件都执行成功
- **THEN** 事务提交，对应 mock 数据行落库
- **AND** `sys_plugin_migration` 表中每个 mock SQL 文件对应一行 `direction='mock'` 的成功记录

### Requirement: Mock 阶段失败时的响应必须携带可操作的失败信息

当 mock 阶段失败导致整体回滚时，宿主 SHALL 通过 `bizerr` 封装返回错误响应，错误内容 MUST 包含：错误码 `plugin.install.mockDataFailed`、i18n key `plugins.install.error.mockDataFailed`、以及包含 `pluginId`（插件标识）、`failedFile`（触发失败的 SQL 文件名）、`rolledBackFiles`（已回滚的文件名列表，包含失败前已执行成功的文件）、`cause`（数据库返回的原始错误信息片段）四个字段的 `messageParams`。前端 SHALL 通过 errorCode 识别该错误并展示包含上述参数的本地化提示，文案 MUST 明确告知用户"插件安装本身已成功，仅 mock 数据加载失败且已自动回滚，可选择接受当前状态或修复后卸载重装"。

#### Scenario: 前端展示 mock 失败的本地化提示
- **WHEN** 后端返回 `plugin.install.mockDataFailed` 错误响应
- **AND** 当前会话语言为 `zh-CN`
- **THEN** 前端在安装弹窗内展示中文提示，包含插件 ID、失败的 SQL 文件名与失败原因
- **AND** 提示明确包含"插件已安装，示例数据已自动回滚"等关键描述

#### Scenario: 错误响应同时承载 install 成功的事实
- **WHEN** install SQL 已成功落库，仅 mock 阶段失败
- **THEN** HTTP 响应使用错误码（非 200）但服务端已完成插件注册、菜单同步等动作
- **AND** 用户调用插件列表接口时该插件出现在 `installed=true` 状态

### Requirement: 迁移账本必须能区分 install / uninstall / mock 三个阶段

宿主 SHALL 在 `apps/lina-core/internal/service/plugin/internal/catalog/metadata.go` 的 `MigrationDirection` 类型上新增枚举值 `MigrationDirectionMock = "mock"`。`sys_plugin_migration` 表的 `direction` 字段 MUST 接受三种取值：`install` / `uninstall` / `mock`。每条 mock SQL 文件执行成功时 SHALL 在该表写入一行 `direction='mock'` 的记录，包含执行结果、错误信息（若有）、checksum、时间戳等与 install 行相同的元数据；事务回滚时这些行 MUST 同步回滚不留残记。

#### Scenario: 账本可按方向查询区分阶段
- **WHEN** 运维通过 SQL 查询 `SELECT direction, COUNT(*) FROM sys_plugin_migration WHERE plugin_id = 'content-notice' GROUP BY direction`
- **AND** 该插件曾经成功安装且包含 mock 数据
- **THEN** 查询结果至少出现 `install` 与 `mock` 两个方向的统计行
- **AND** 数量分别等于 install SQL 与 mock SQL 的文件数

### Requirement: 源码插件与动态插件必须共享同一套 mock 数据加载机制

宿主 SHALL 在源码插件（embed FS）与动态插件（运行时上传的制品 FS）两个加载路径中复用同一份 mock-data 目录扫描与事务化执行逻辑。两类插件 MUST 使用相同的 `manifest/sql/mock-data/` 目录约定、相同的 SQL 文件命名约束（`^\d{3}-[a-z0-9-]+\.sql$`）、相同的事务执行入口、相同的失败响应格式。前端复选框、配置 opt-in、错误码与 i18n 文案 MUST NOT 区分插件类型。

#### Scenario: 动态插件支持示例数据加载
- **WHEN** 一个动态插件的制品包内包含 `manifest/sql/mock-data/001-*.sql`
- **AND** 用户在前端弹窗中勾选"安装示例数据"安装该动态插件
- **THEN** 宿主使用与源码插件相同的事务化执行流程加载 mock 数据
- **AND** 失败时返回相同 errorCode 的错误响应

#### Scenario: 源码插件与动态插件复选框 UX 一致
- **WHEN** 用户分别打开一个源码插件与一个动态插件的安装弹窗
- **AND** 两个插件都包含 mock-data 目录
- **THEN** 复选框文案、tooltip、默认值、勾选后的提交行为完全一致
