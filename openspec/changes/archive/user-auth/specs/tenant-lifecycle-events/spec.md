## ADDED Requirements

### Requirement: 租户生命周期领域编排
`multi-tenant` 插件 SHALL 在租户创建、暂停、恢复、删除时执行当前版本明确需要的领域编排,但 SHALL NOT 创建 `plugin_multi_tenant_event_outbox` 表或声明未实现的跨插件事件总线。

#### Scenario: 创建租户执行内置副作用
- **WHEN** 平台管理员创建租户 T
- **THEN** `multi-tenant` 插件先持久化 `plugin_multi_tenant_tenant`
- **AND** 然后直接执行新租户插件启用策略的领域服务
- **AND** 不写入 outbox 事件,也不承诺异步重投递

### Requirement: 显式插件治理编排
新租户默认启用 tenant-scoped 插件的策略 SHALL 由 `multi-tenant` 插件在租户创建流程中显式调用插件治理服务完成,而不是通过生命周期事件订阅完成。

#### Scenario: 新租户默认启用策略
- **WHEN** 租户 T 创建成功
- **THEN** 系统查询已安装、已启用、`tenant_aware`、`tenant_scoped` 且平台策略允许新租户默认启用的插件
- **AND** 为租户 T 写入对应租户插件状态

### Requirement: 删除 outbox 占位表
系统 SHALL NOT 使用只有写入、缺少订阅注册、消费者认领、重试、死信和 per-subscriber 投递状态的 `plugin_multi_tenant_event_outbox` 占位表表达生命周期事件能力。

#### Scenario: 初始化 SQL 不创建 outbox
- **WHEN** 初始化或安装 multi-tenant 插件 schema
- **THEN** 数据库中不存在 `plugin_multi_tenant_event_outbox`
- **AND** Go 代码中不存在对该表的 DAO、DO、Entity 或手写表名访问

### Requirement: 删除前置否决与清理顺序
租户删除 SHALL 在所有插件 `CanTenantDelete` 钩子全部通过后才执行软删除。当前版本不通过 `tenant.deleted` 事件触发跨插件清理;需要跨插件删除清理时,必须先设计完整可靠的事件或生命周期编排机制。

#### Scenario: 多插件级联清理
- **WHEN** 租户 T 被删除
- **THEN** 系统先调用所有插件的 `CanTenantDelete(ctx, T)`,任意 `false` 即拒绝
- **AND** 全部通过后在 `plugin_multi_tenant_tenant` 中标记 `deleted_at`

### Requirement: 暂停/恢复事件不触发数据清理
租户暂停与恢复 SHALL 仅修改租户状态,不触发任何数据清理或事件分发;所有租户数据必须保留以便恢复。

#### Scenario: 暂停后恢复
- **WHEN** 租户 T 被暂停后又恢复
- **THEN** T 的所有数据完整保留
- **AND** 各插件的运行时状态恢复到暂停前等价状态
