## ADDED Requirements

### Requirement: org-center 岗位表租户化
`plugin_org_center_post` 与 `plugin_org_center_user_post` SHALL 加 `tenant_id` 列;岗位 CRUD 与"用户岗位选项"查询按 `tenant_id = bizctx.TenantId` 隔离。

#### Scenario: 跨租户岗位不可见
- **WHEN** 租户 A 管理员查询岗位列表
- **THEN** 仅返回 `tenant_id=A` 岗位

### Requirement: 租户内岗位编码唯一
岗位 `code` 唯一性约束 SHALL 在 `(tenant_id, code)` 上;不同租户可重复使用同 code。

#### Scenario: 同 code 跨租户不冲突
- **WHEN** 租户 A 与租户 B 各自创建 `code='engineer'` 岗位
- **THEN** 两次创建均成功

### Requirement: 租户删除时不通过事件级联清理岗位
`org-center` SHALL NOT 依赖未实现的 `tenant.deleted` 事件总线清理本租户岗位与用户-岗位关联;需要清理时应通过显式管理流程或后续可靠生命周期编排设计实现。

#### Scenario: 不写岗位清理事件
- **WHEN** 租户 T 被删除
- **THEN** 系统不写入 `tenant.deleted` outbox
- **AND** 不触发岗位事件订阅者回调
