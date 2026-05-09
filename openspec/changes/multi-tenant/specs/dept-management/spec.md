## ADDED Requirements

### Requirement: org-center 部门表租户化
`plugin_org_center_dept` 与 `plugin_org_center_user_dept` SHALL 加 `tenant_id` 列;部门树构建/查询/修改全部按 `tenant_id = bizctx.TenantId` 隔离。

#### Scenario: 跨租户部门不可见
- **WHEN** 租户 A 管理员查询部门树
- **THEN** 仅返回 `tenant_id=A` 的部门节点
- **AND** 不见租户 B 的任何部门

### Requirement: 租户创建时初始化默认部门
`org-center` 插件 SHALL 订阅 `tenant.created` 事件,在新租户创建时自动 insert 一行"租户根部门"(`name = tenant.name`,`tenant_id = 新租户 id`,`parent_id = 0`)。

#### Scenario: 租户创建后部门树
- **WHEN** 平台管理员创建租户 T,name="ACME"
- **THEN** `plugin_org_center_dept` 自动新增一行 `(tenant_id=T, name='ACME', parent_id=0)`
- **AND** 租户 T 管理员登录后部门树非空

### Requirement: 租户删除时级联清理部门
`org-center` 插件 SHALL 订阅 `tenant.deleted` 事件,清理该租户所有 dept、post、user_dept、user_post 记录(幂等)。

#### Scenario: 删除租户清理部门
- **WHEN** 租户 T 被删除
- **THEN** 所有 `plugin_org_center_*.tenant_id = T` 行被删除
- **AND** 不影响其他租户

### Requirement: orgcap.Provider 实现按租户视图
`org-center` 插件实现的 `orgcap.Provider` 接口方法 SHALL 内部读取 `bizctx.TenantId` 并过滤本租户数据;平台管理员 impersonation 时按目标租户视图返回。

#### Scenario: 部门 scope 注入
- **WHEN** 租户 A 内的查询走 `datascope.ApplyUserScope` 触发 dept-scope 注入
- **THEN** EXISTS 子查询的 dept 来自 `tenant_id=A`
- **AND** 不污染租户 B 的查询计划
