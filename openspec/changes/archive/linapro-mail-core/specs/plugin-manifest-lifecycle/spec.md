## ADDED Requirements

### Requirement: 源码插件启用必须执行生命周期前置条件

系统 SHALL 在源码插件从非启用变为启用时，于持久化启用状态之前执行生命周期前置回调，至少包括目标插件 `BeforeEnable`（若已注册）以及全部已注册的 `GlobalBeforeEnable`。前置否决 MUST 阻止启用状态写入。既有依赖检查、平台上下文守卫与 authorization 语义 MUST 继续适用且不得被前置回调绕过。

#### Scenario: 启用被全局 Hook 否决

- **WHEN** 某 owner 插件的 `GlobalBeforeEnable` 对目标插件返回否决
- **THEN** lifecycle 编排 MUST 不将该目标插件标记为启用
- **AND** 调用方 MUST 收到生命周期前置否决错误

#### Scenario: 启用成功后的观察副作用

- **WHEN** 源码插件启用前置全部通过并成功写入启用状态
- **THEN** 系统 MUST 继续执行既有启用成功副作用（例如 runtime 快照同步与 `plugin.enabled` 观察扩展点）
- **AND** 目标 `AfterEnable`（若已注册）MUST 按 best-effort 语义执行

### Requirement: 源码插件安装必须聚合全局 BeforeInstall

系统 SHALL 在源码插件安装副作用前聚合目标 `BeforeInstall`（若已注册）与全部 `GlobalBeforeInstall`。全局 Hook 的引入 MUST NOT 改变 force 安装、依赖计划、SQL migration 账本与 rollback 诊断的既有义务，仅增加可选 veto 参与者。

#### Scenario: 安装被全局 Hook 否决

- **WHEN** 某插件注册的 `GlobalBeforeInstall` 否决目标插件安装
- **THEN** 系统 MUST 不完成该安装的持久化成功路径
- **AND** 否决 reason MUST 可被管理端展示

### Requirement: 禁用与卸载可接入全局前置且保留 force 语义

系统 SHALL 在源码插件禁用与卸载编排中支持调用已注册的 `GlobalBeforeDisable` 与 `GlobalBeforeUninstall`。全局 Hook MUST NOT 削弱既有 force/purge、依赖阻断与权限校验的既有语义；force 路径是否可绕过全局 veto 须与既有 force 约定一致。

#### Scenario: 无全局参与者时行为不变

- **WHEN** 系统中没有任何插件注册全局 BeforeDisable/BeforeUninstall
- **THEN** 禁用与卸载路径 MUST 与引入全局 Hook 前的行为一致
- **AND** MUST NOT 引入额外空参与者调用开销
