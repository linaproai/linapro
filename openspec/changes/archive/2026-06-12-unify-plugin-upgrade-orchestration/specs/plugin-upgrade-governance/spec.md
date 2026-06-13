## ADDED Requirements

### Requirement: 插件升级必须由统一升级编排组件执行

系统 SHALL 将源码插件升级和动态插件升级的 preview、execute、失败记账、release 提升、治理资源同步和缓存发布纳入同一升级编排模型。source 与 dynamic 插件可以保留不同执行策略，但共享依赖校验、反向依赖保护、失败诊断、治理守卫边界和缓存发布骨架。

#### Scenario: 管理员执行源码插件升级

- **WHEN** 管理员确认升级 source 插件
- **THEN** 根插件服务执行平台治理守卫后委托统一升级组件
- **AND** 统一升级组件执行 source 策略、升级 SQL、治理资源同步、release 提升和缓存发布
- **AND** 不通过再次调用根服务公开升级方法完成 source 升级

#### Scenario: 管理员执行动态插件升级

- **WHEN** 管理员确认升级 dynamic 插件
- **THEN** 根插件服务执行平台治理守卫后委托统一升级组件
- **AND** 统一升级组件执行 dynamic 策略、授权快照持久化、runtime upgrade request、release 切换和缓存发布
- **AND** 动态插件失败时仍保留旧有效 release 和可诊断失败状态

### Requirement: 插件升级失败诊断必须使用单一账本约定

系统 SHALL 使用一套`sys_plugin_migration`升级失败诊断约定表达 source 与 dynamic 插件升级失败。失败 phase、error code、message key、fallback、目标 release 和原始错误信息 MUST 由统一升级模型归一化，不得在 source 与 dynamic 路径分别维护互不一致的读写语义。

#### Scenario: source 插件升级 SQL 失败

- **WHEN** source 插件升级 SQL 执行失败
- **THEN** 系统记录统一 upgrade phase 的失败账本
- **AND** 插件运行时升级状态展示同一套 phase、error code、message key 和 fallback 语义
- **AND** 有效版本和有效 release 保持升级前状态

#### Scenario: dynamic 插件 release 切换失败

- **WHEN** dynamic 插件目标 release 切换或后续缓存发布失败
- **THEN** 系统使用统一失败诊断约定记录失败阶段
- **AND** 失败目标 release 不得成为有效 release 或派生缓存权威来源

### Requirement: 插件升级治理守卫必须只在公开入口执行一次

系统 SHALL 在公开插件升级入口执行平台治理守卫，并禁止统一升级组件通过再入公开插件服务方法重复执行守卫或重复发布缓存。内部 source/dynamic 策略必须通过窄契约调用所需能力，并保持启动期或内部查询路径的租户上下文语义。

#### Scenario: 统一入口分派到 source 策略

- **WHEN** 公开 runtime upgrade execute 入口判断目标插件类型为 source
- **THEN** 根门面已执行一次平台治理守卫
- **AND** 内部 source 策略直接执行升级流程
- **AND** 系统不再次调用公开`UpgradeSourcePlugin`入口

#### Scenario: 内部升级状态查询

- **WHEN** 启动期或管理读模型查询插件升级状态
- **THEN** 系统调用无副作用的统一升级状态查询能力
- **AND** 查询不得绕过租户上下文或修改治理数据

### Requirement: 插件升级缓存发布必须复用插件变化发布入口

系统 SHALL 在 source 和 dynamic 插件升级成功、失败或失败诊断变化后，通过统一插件变化发布入口发布作用域化变化。发布必须包含插件 ID、插件类型和 reason，并继续复用`plugin-runtime`revision controller、管理读模型失效、runtime 派生缓存失效、frontend bundle、i18n runtime bundle 和 WASM 派生缓存失效机制。

#### Scenario: source 插件升级成功后发布变化

- **WHEN** source 插件升级成功并切换有效 release
- **THEN** 统一升级组件通过插件变化发布入口发布 source 插件变化
- **AND** 管理读模型、runtime 派生缓存、frontend bundle 和 i18n runtime bundle 观察同一 revision 失效

#### Scenario: dynamic 插件升级失败后发布失败状态

- **WHEN** dynamic 插件升级失败并写入失败诊断
- **THEN** 统一升级组件通过插件变化发布入口发布 dynamic 插件变化
- **AND** 其他节点不得把失败目标 release 作为有效缓存来源

