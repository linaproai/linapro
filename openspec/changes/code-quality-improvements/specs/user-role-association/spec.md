## ADDED Requirements

### Requirement: sys_user_role 必须包含 role_id 反向索引
系统 SHALL 在 `sys_user_role` 表上维护 `KEY idx_role_id (role_id)` 反向索引，以支撑"按角色查询用户"、"按角色批量删除关联"等访问路径，避免在仅有 `(user_id, role_id)` 复合主键场景下产生全表扫描。

#### Scenario: sys_user_role 反向索引存在
- **WHEN** `make init` 完成数据库初始化
- **THEN** `SHOW INDEX FROM sys_user_role` 结果中 MUST 出现 `idx_role_id`，索引列为 `role_id`

#### Scenario: 按角色查询用户走索引
- **WHEN** 服务执行 `WHERE role_id = ?` 类型查询
- **THEN** 数据库 MUST 选中 `idx_role_id` 索引，避免全表扫描

## MODIFIED Requirements

### Requirement: 用户删除时清理角色关联

系统 SHALL 在删除用户时清理角色关联数据，并 MUST 把用户软删除与 `sys_user_role` 关联清理放入同一事务，任意失败 MUST 触发整体回滚。

#### Scenario: 删除用户事务化清理角色关联
- **WHEN** 用户删除一个用户
- **AND** 全部清理步骤成功
- **THEN** 系统在事务内软删除用户记录
- **AND** 系统删除 sys_user_role 中该用户的所有关联记录
- **AND** 事务提交后才通知访问拓扑变更

#### Scenario: 关联清理失败触发整体回滚
- **WHEN** 用户删除一个用户
- **AND** `sys_user_role` 关联清理在事务内返回错误
- **THEN** 用户软删除 MUST 一并回滚
- **AND** 操作返回底层错误
