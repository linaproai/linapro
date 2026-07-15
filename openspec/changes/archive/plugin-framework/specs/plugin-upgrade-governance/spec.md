## Requirements

### Requirement: 插件升级必须由统一升级编排组件执行

系统 SHALL 将源码插件升级和动态插件升级的 preview、execute、失败记账、release 提升、治理资源同步和缓存发布纳入同一升级编排模型。source 与 dynamic 插件可以保留不同执行策略，但共享依赖校验、反向依赖保护、失败诊断、治理守卫边界和缓存发布骨架。

### Requirement: 插件升级失败诊断必须使用单一账本约定

系统 SHALL 使用一套 `sys_plugin_migration` 升级失败诊断约定表达 source 与 dynamic 插件升级失败。失败 phase、error code、message key、fallback、目标 release 和原始错误信息 MUST 由统一升级模型归一化。

### Requirement: 插件升级治理守卫必须只在公开入口执行一次

系统 SHALL 在公开插件升级入口执行平台治理守卫，并禁止统一升级组件通过再入公开插件服务方法重复执行守卫或重复发布缓存。

### Requirement: 插件升级缓存发布必须复用插件变化发布入口

系统 SHALL 在 source 和 dynamic 插件升级成功、失败或失败诊断变化后，通过统一插件变化发布入口发布作用域化变化。发布必须包含插件 ID、插件类型和 reason，并继续复用 `plugin-runtime` revision controller。

### Requirement: 插件升级必须校验下游 Provider 插件依赖

插件升级 SHALL 校验升级后的 provider 插件状态不会破坏其他已安装插件通过既有`dependencies.plugins`声明的硬依赖。升级与发布切换 MUST 按安装轴评估：只要存在已安装下游硬依赖且候选版本不满足其版本范围，系统 MUST 拒绝该操作或进入规范明确的阻断状态，无论下游当前是否启用。

禁用 provider 插件时 MUST 按运行轴评估：仅当存在已启用下游硬依赖时阻断禁用；已安装但已禁用的下游 MUST NOT 单独阻断禁用。pluginservice capability 的可选消费仍通过运行时可用性降级表达，不引入独立 capability 依赖阻断模型。

#### Scenario: Provider 升级后不满足下游插件依赖版本

- **WHEN** 已安装插件`consumer`在`dependencies.plugins`中硬依赖`linapro-org-core`版本范围`>=1.0.0 <2.0.0`
- **AND** 管理员尝试将`linapro-org-core`升级为不满足该范围的新版本
- **THEN** 升级请求失败或要求先处理下游依赖
- **AND** 错误包含下游插件 ID、provider 插件 ID 和所需版本范围
- **AND** 即使`consumer`当前已禁用，升级仍 MUST 被阻断

#### Scenario: 禁用唯一 Provider 时保护已启用下游硬依赖

- **WHEN** 插件`consumer`已启用且通过`dependencies.plugins`硬依赖唯一 tenant provider 插件
- **AND** 管理员尝试禁用唯一 tenant provider 插件
- **THEN** 禁用请求失败
- **AND** 错误包含依赖该 provider 插件的已启用下游插件列表

#### Scenario: 下游仅禁用时允许禁用唯一 Provider

- **WHEN** 插件`consumer`已安装但已禁用，且通过`dependencies.plugins`硬依赖唯一 tenant provider 插件
- **AND** 管理员尝试禁用唯一 tenant provider 插件
- **THEN** 禁用请求 MUST 成功（在无其他已启用下游硬依赖时）
