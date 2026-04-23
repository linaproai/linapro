## ADDED Requirements

### Requirement: 源码插件必须区分当前生效版本与源码发现版本
系统 SHALL 将源码插件的当前生效版本与源码树中发现到的更高版本明确区分：`sys_plugin.version` / `release_id` 只表示当前生效版本；源码扫描发现的新版本只能以 release 形式写入治理表，在显式执行源码插件升级前不得直接覆盖当前生效版本。

#### Scenario: 已安装源码插件发现更高版本
- **WHEN** 源码插件 `plugin-demo` 当前生效版本为 `v0.1.0`，且源码树中的 `plugin.yaml` 已提升到 `v0.5.0`
- **THEN** `sys_plugin.version` 必须继续保持 `v0.1.0`
- **AND** 系统必须记录 `v0.5.0` 对应的源码插件 release 快照
- **AND** 该新 release 在显式升级前不得被视为当前生效版本

### Requirement: 源码插件升级必须在开发阶段显式完成
系统 SHALL 要求源码插件升级通过开发态统一命令显式执行，而不是在宿主启动期间自动补做升级；该命令必须支持单插件与批量源码插件升级。

#### Scenario: 升级指定源码插件
- **WHEN** 开发者执行 `make upgrade confirm=upgrade scope=source-plugin plugin=plugin-demo`
- **THEN** 系统必须只为该源码插件生成并执行升级计划
- **AND** 不得同时触发其他源码插件或动态插件的升级

#### Scenario: 批量升级全部源码插件
- **WHEN** 开发者执行 `make upgrade confirm=upgrade scope=source-plugin plugin=all`
- **THEN** 系统必须扫描全部源码插件并按确定性顺序处理待升级插件
- **AND** 对于未安装或无需升级的源码插件，必须明确输出跳过结果

### Requirement: 宿主启动前必须校验源码插件升级是否完成
宿主 SHALL 在启动阶段完成源码插件扫描后执行待升级校验；若存在已安装源码插件的源码发现版本高于当前生效版本，则宿主 MUST 阻断启动，并提示先执行对应的开发态升级命令。

#### Scenario: 启动时存在待升级源码插件
- **WHEN** 宿主启动时发现 `plugin-demo` 当前生效版本为 `v0.1.0`，源码发现版本为 `v0.5.0`
- **THEN** 启动流程必须失败
- **AND** 错误信息必须包含插件 ID、当前版本、发现版本和建议执行的 `make upgrade` 命令

### Requirement: 源码插件升级必须记录 upgrade phase 并同步治理资源
源码插件升级命令 SHALL 在版本比较通过后执行 `phase=upgrade` 的迁移流程，并同步菜单、权限和资源引用治理数据；升级成功后必须将当前生效版本切换到新 release。

#### Scenario: 源码插件升级成功
- **WHEN** 开发者对已安装源码插件执行升级命令且全部 SQL / 菜单 / 资源同步成功
- **THEN** `sys_plugin.version` 与 `release_id` 必须更新为新版本对应 release
- **AND** `sys_plugin_migration` 中必须记录 `phase=upgrade` 的执行结果
- **AND** 新版本 release 必须成为当前生效 release

#### Scenario: 源码插件升级失败
- **WHEN** 源码插件升级过程中某个升级 SQL 或治理同步步骤失败
- **THEN** 命令必须立即停止后续步骤
- **AND** 必须保留失败的 upgrade 记录与错误信息
- **AND** 本轮不得自动执行 rollback

### Requirement: 动态插件升级继续保留运行时模型
系统 SHALL 继续通过动态插件 upload + install/reconcile 的运行时模型处理动态插件升级；开发态 `make upgrade` 不得介入动态插件升级流程。

#### Scenario: 开发态升级命令不处理动态插件
- **WHEN** 开发者执行 `make upgrade confirm=upgrade scope=source-plugin plugin=plugin-demo`
- **THEN** 系统不得扫描或切换任何动态插件 release
- **AND** 动态插件仍然必须通过 upload + install/reconcile 进入升级流程
