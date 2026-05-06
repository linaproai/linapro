## ADDED Requirements

### Requirement: 插件 SQL 资源执行前必须经当前方言转译

系统 SHALL 在执行插件 `manifest/sql/` 安装资产、`manifest/sql/uninstall/` 卸载资产、`manifest/sql/mock-data/` mock 资产中的任一 SQL 文件之前，先调用当前方言的 `TranslateDDL` 将单一 MySQL 方言来源的 SQL 内容转换为目标方言可执行内容。该规则同时适用于源码插件与动态插件、安装阶段与卸载阶段、运行时嵌入式 SQL 与开发时本地 SQL。插件源码侧 SHALL 仅维护单一 MySQL 方言来源的 SQL 文件，不得为不同数据库引擎维护多份 SQL 文件。

#### Scenario: 源码插件安装时 SQL 走方言转译
- **当** 源码插件 `monitor-loginlog` 在 SQLite 模式下首次启用并执行 `manifest/sql/001-monitor-loginlog-schema.sql` 时
- **则** 插件安装管线先调用 `SQLiteDialect.TranslateDDL` 将 MySQL 方言 DDL 转译为 SQLite 兼容语句
- **且** 转译后的语句在 SQLite 数据库上成功执行
- **且** 插件源码 `manifest/sql/` 目录下保持单一 MySQL 方言 SQL 文件

#### Scenario: 动态插件升级时 SQL 走方言转译
- **当** 动态插件升级到新版本且新版本携带新的 `manifest/sql/*.sql` 文件
- **且** 当前宿主以 SQLite 方言运行
- **则** 插件升级管线对每个新 SQL 文件调用 `SQLiteDialect.TranslateDDL`
- **且** 转译后的语句逐一执行
- **且** 任一文件转译或执行失败时升级管线返回失败状态

#### Scenario: 插件卸载时 uninstall SQL 走方言转译
- **当** 源码插件或动态插件被卸载且 `manifest/sql/uninstall/` 下存在卸载 SQL 时
- **则** 卸载管线对每个 uninstall SQL 文件调用当前方言的 `TranslateDDL`
- **且** 转译后的 `DROP TABLE IF EXISTS` 等语句在当前数据库上成功执行
- **且** 卸载流程不依赖原 MySQL 方言专属语法

#### Scenario: 插件 mock-data 加载时 SQL 走方言转译
- **当** 运维人员运行 `make mock confirm=mock` 且某插件提供 `manifest/sql/mock-data/*.sql` 时
- **则** mock 加载管线对每个 mock SQL 文件调用当前方言的 `TranslateDDL`
- **且** 转译后的 `INSERT IGNORE INTO` / `INSERT INTO` 等语句在当前数据库上成功执行

#### Scenario: 插件 SQL 转译失败时安装管线快速失败
- **当** 插件某 SQL 文件包含当前方言转译器未覆盖的 MySQL 语法时
- **则** 插件安装 / 升级 / 卸载 / mock 加载管线立即停止后续 SQL 执行
- **且** 错误日志包含失败的插件标识、SQL 资产类型与失败文件名
- **且** 管线向上层返回失败状态，便于调用方明确定位待修复的 SQL 文件

### Requirement: 插件 SQL 文件必须能被默认方言转译器处理

源码插件与动态插件提交到仓库或发布产物中的 SQL 文件 SHALL 限定在默认方言转译器（当前为 `SQLiteDialect.TranslateDDL`）的覆盖范围内。具体而言，插件 SQL 不得使用以下 MySQL 特性：`FULLTEXT INDEX` / `SPATIAL INDEX` / `GENERATED ALWAYS AS` / 分区子句 / `ON DUPLICATE KEY UPDATE` / 数据库专用函数（`FIND_IN_SET` / `GROUP_CONCAT` / `IF()` 等）；这些语法本就由项目编码规范禁止，新增本规则强化了"在 SQLite 模式下也必须能跑通"的边界要求。

#### Scenario: 插件 SQL 不使用 FULLTEXT 等未覆盖语法
- **当** 任一源码插件或动态插件提交的 `manifest/sql/*.sql` 文件被默认方言转译器处理时
- **则** 转译过程不返回"未覆盖语法"错误
- **且** 转译结果可在 SQLite 上成功执行

#### Scenario: 违规 SQL 在审查与测试阶段被发现
- **当** 插件作者提交包含 `FULLTEXT INDEX` 等未覆盖语法的 SQL 文件时
- **则** 方言转译器单元测试或 E2E 测试在 SQLite 模式下报错
- **且** 错误消息明确指向违规文件与未覆盖关键字
- **且** 该 SQL 在 PR / 变更审查阶段必须被修复，方可进入主干
