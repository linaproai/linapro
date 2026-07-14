## ADDED Requirements

### Requirement: 目标插件必须支持 BeforeEnable 与 AfterEnable

系统 SHALL 为源码插件生命周期提供目标插件作用域的 `BeforeEnable` 与 `AfterEnable` 回调注册能力。`BeforeEnable` MUST 可在插件状态变为启用前否决该操作；`AfterEnable` MUST 仅在启用成功后以 best-effort 方式执行，不得用于否决。

#### Scenario: BeforeEnable 否决启用

- **WHEN** 管理端请求启用某源码插件且该插件注册了 `BeforeEnable` 并返回否决
- **THEN** 系统 MUST 保持该插件为未启用状态
- **AND** 系统 MUST 向调用方返回生命周期前置条件否决错误及可本地化的 reason

#### Scenario: AfterEnable 在成功后执行

- **WHEN** 某源码插件成功变为启用状态且注册了 `AfterEnable`
- **THEN** 系统 MUST 调用该回调
- **AND** 该回调失败 MUST NOT 回滚已成功的启用状态（best-effort 语义，与现有 After* 一致）

### Requirement: 系统必须支持全局前置生命周期 Hook

系统 SHALL 允许源码插件显式注册全局前置生命周期回调，至少包括 `GlobalBeforeInstall` 与 `GlobalBeforeEnable`。全局回调的输入 MUST 标识被操作的目标插件 ID 与操作类型。全局回调 MUST 可否决对目标插件的对应操作。宿主插件管理实现 MUST NOT 根据业务领域（例如邮件协议 kind）硬编码冲突规则。

#### Scenario: 全局 BeforeEnable 否决其他插件启用

- **WHEN** 插件 A 已注册 `GlobalBeforeEnable`，且管理端请求启用目标插件 B
- **THEN** 系统 MUST 在变更 B 的启用状态前调用 A 的全局回调，并传入目标插件 ID 为 B
- **AND** 若 A 返回否决，系统 MUST 不启用 B 并返回 veto 错误

#### Scenario: 未注册全局 Hook 的插件不参与

- **WHEN** 某源码插件未注册任何全局前置 Hook
- **THEN** 系统 MUST NOT 因安装或启用其他插件而调用该插件的空全局路径
- **AND** 参与者集合 MUST 仅包含显式注册了对应全局 Hook 的插件

### Requirement: 安装与启用编排必须聚合目标与全局前置 Hook

系统 SHALL 在源码插件安装与启用编排中，于产生持久化副作用之前依次或并发聚合：目标插件自管 `Before*`（若存在）与全部适用的全局 `GlobalBefore*`。任一否决、超时、panic 或错误 MUST 使该操作为失败（fail-closed），并聚合可诊断的 reason。

#### Scenario: 启用路径接入 precondition

- **WHEN** 管理端启用已安装的源码插件
- **THEN** 系统 MUST 在写入启用状态前执行目标 `BeforeEnable`（若已注册）与全部 `GlobalBeforeEnable`
- **AND** 仅当全部前置回调允许时才持久化启用状态并发布启用成功副作用

#### Scenario: 安装路径接入全局前置

- **WHEN** 管理端安装源码插件
- **THEN** 系统 MUST 在安装副作用前执行目标 `BeforeInstall`（若已注册）与全部 `GlobalBeforeInstall`
- **AND** 既有 force、依赖检查、authorization 与 SQL 语义 MUST NOT 因全局 Hook 引入而改变顺序义务

### Requirement: 全局 Hook 不得复用自管 BeforeInstall 语义

系统 SHALL 将全局前置 Hook 建模为独立注册面与独立输入类型，MUST NOT 通过「对所有插件广播目标插件的自管 `BeforeInstall`/`BeforeEnable`」实现全局拦截，以免误触发插件「处理自身安装」逻辑。

#### Scenario: 自管与全局输入区分

- **WHEN** 插件 C 同时注册了自管 `BeforeEnable` 与 `GlobalBeforeEnable`
- **THEN** 启用 C 时自管回调的语义 MUST 针对 C 自身
- **AND** 启用其他插件 D 时 MUST 仅调用 C 的全局回调（输入目标为 D），MUST NOT 以「正在启用 C」的语义调用 C 的自管 `BeforeEnable`

### Requirement: 可选提供全局 BeforeDisable 与 BeforeUninstall 注册面

系统 SHALL 提供 `GlobalBeforeDisable` 与 `GlobalBeforeUninstall` 的注册与编排能力，供 owner 插件在目标插件禁用或卸载前表达依赖保护。未注册则无行为变化。

#### Scenario: 全局 BeforeUninstall 可否决卸载

- **WHEN** 某插件注册了 `GlobalBeforeUninstall` 并对目标插件卸载返回否决
- **THEN** 系统 MUST 阻止该卸载（除非既有 force 卸载策略显式允许且规范要求可绕过时按既有 force 语义处理）
