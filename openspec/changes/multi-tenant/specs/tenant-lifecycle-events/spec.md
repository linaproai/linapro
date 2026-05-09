## ADDED Requirements

### Requirement: 租户生命周期事件总线
`multi-tenant` 插件 SHALL 在租户状态发生变更时发布事件 `tenant.created` / `tenant.suspended` / `tenant.resumed` / `tenant.archived` / `tenant.deleted`;其他插件可订阅这些事件以执行各自的初始化与清理。

#### Scenario: 创建租户触发事件
- **WHEN** 平台管理员创建租户 T
- **THEN** `multi-tenant` 插件先持久化 `plugin_multi_tenant_tenant`
- **AND** 然后向所有订阅 `tenant.created` 的插件发布事件 payload `{tenant_id: T, code, name, plan}`
- **AND** 事件投递失败时不回滚租户创建,但记录失败日志并支持重投递

### Requirement: 事件订阅机制
插件 SHALL 通过实现 `tenantcap.LifecycleSubscriber` 接口订阅事件;接口由 `multi-tenant` 插件以 host service 形式暴露,订阅注册在插件 `OnEnable` 钩子中完成。

#### Scenario: org-center 订阅 tenant.created
- **WHEN** `org-center` 插件 enable
- **THEN** 注册 `OnTenantCreated(ctx, evt)` 回调
- **AND** 收到事件后自动创建租户根部门 `name = evt.name`,`tenant_id = evt.tenant_id`

### Requirement: 事件投递语义(at-least-once)
事件投递 SHALL 保证 at-least-once 语义,订阅者必须自行实现幂等;事件持久化在 `plugin_multi_tenant_event_outbox` 表(初版直接同步分发,outbox 表先建好留扩展)。

#### Scenario: 订阅者幂等
- **WHEN** 同一 `tenant.created` 事件被投递两次
- **THEN** 订阅者基于 `(tenant_id, event_type, event_id)` 做去重
- **AND** 第二次投递不产生副作用

### Requirement: 删除前置否决与清理顺序
`tenant.deleted` 事件 SHALL 在所有插件 `CanTenantDelete` 钩子全部通过后才发布;事件分发顺序应反向(后注册的先清理),确保依赖关系。

#### Scenario: 多插件级联清理
- **WHEN** 租户 T 被删除
- **THEN** 系统先调用所有插件的 `CanTenantDelete(ctx, T)`,任意 `false` 即拒绝
- **AND** 全部通过后,按反注册顺序触发各插件的清理回调
- **AND** 最后才在 `plugin_multi_tenant_tenant` 中标记 `deleted_at`

### Requirement: 暂停/恢复事件不触发数据清理
`tenant.suspended` 与 `tenant.resumed` SHALL 仅作为状态变更通知,不触发任何数据清理;插件可据此调整运行时行为(例如关闭定时任务、暂停消费),但必须保留所有数据以便恢复。

#### Scenario: 暂停后恢复
- **WHEN** 租户 T 被暂停后又恢复
- **THEN** T 的所有数据完整保留
- **AND** 各插件的运行时状态恢复到暂停前等价状态
