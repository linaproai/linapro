## ADDED Requirements

### Requirement: org-center 部门表租户化
`plugin_org_center_dept` 与 `plugin_org_center_user_dept` SHALL 加 `tenant_id` 列;部门树构建/查询/修改全部按 `tenant_id = bizctx.TenantId` 隔离。

#### Scenario: 跨租户部门不可见
- **WHEN** 租户 A 管理员查询部门树
- **THEN** 仅返回 `tenant_id=A` 的部门节点
- **AND** 不见租户 B 的任何部门

### Requirement: 租户创建默认部门暂不自动建模
`org-center` 插件 SHALL NOT 依赖未实现的 `tenant.created` 事件总线自动创建租户根部门;租户部门数据由 org-center 的租户化 CRUD 与 mock/业务初始化流程显式维护。

#### Scenario: 租户创建不写事件 outbox
- **WHEN** 平台管理员创建租户 T,name="ACME"
- **THEN** 系统不写入 `plugin_multi_tenant_event_outbox`
- **AND** org-center 部门表不会因为事件订阅自动新增根部门

### Requirement: 租户删除时不通过事件级联清理部门
`org-center` 插件 SHALL NOT 依赖未实现的 `tenant.deleted` 事件总线清理租户部门数据;需要清理时应通过显式管理流程或后续可靠生命周期编排设计实现。

#### Scenario: 删除租户清理部门
- **WHEN** 租户 T 被删除
- **THEN** 不触发 `tenant.deleted` outbox 或订阅者回调
- **AND** 不影响其他租户数据

### Requirement: orgcap.Provider 实现按租户视图
`org-center` 插件实现的 `orgcap.Provider` 接口方法 SHALL 内部读取 `bizctx.TenantId` 并过滤本租户数据;平台管理员 impersonation 时按目标租户视图返回。

#### Scenario: 部门 scope 注入
- **WHEN** 租户 A 内的查询走 `datascope.ApplyUserScope` 触发 dept-scope 注入
- **THEN** EXISTS 子查询的 dept 来自 `tenant_id=A`
- **AND** 不污染租户 B 的查询计划
