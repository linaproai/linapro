## ADDED Requirements

### Requirement: 角色批量删除
系统 SHALL 提供 `DELETE /api/v1/role?ids=...` 批量删除接口，复用单条删除的全部保护策略并保证整体事务原子性。

#### Scenario: 批量删除多个未分配用户的角色
- **WHEN** 拥有 `role:remove` 权限的调用方调用 `DELETE /api/v1/role?ids=2,3,4`
- **AND** 目标角色均未分配给任何用户且不是超级管理员
- **THEN** 系统在单事务内软删除全部角色
- **AND** 同步删除 `sys_role_menu` 和 `sys_user_role` 中对应关联记录
- **AND** 提交后触发一次访问拓扑变更通知

#### Scenario: 批量包含超级管理员角色被拒绝
- **WHEN** 调用方调用 `DELETE /api/v1/role?ids=1,2,3`
- **AND** id `1` 是超级管理员角色
- **THEN** 整个批次 MUST 被拒绝
- **AND** 没有任何角色被删除

#### Scenario: 空 id 列表在校验阶段被拒绝
- **WHEN** 调用方调用 `DELETE /api/v1/role?ids=`
- **THEN** 系统返回参数校验错误
- **AND** 不会进入事务

## MODIFIED Requirements

### Requirement: 删除角色

系统 SHALL 支持删除角色，删除过程 MUST 通过事务保证原子性，事务内任何关联记录删除失败都 MUST 触发整体回滚，禁止仅记录 warning 后继续删除角色本身。访问拓扑变更通知 MUST 在事务提交后再发出。

#### Scenario: 删除未分配用户的角色
- **WHEN** 用户删除一个未分配给任何用户的角色
- **THEN** 系统在事务内软删除该角色（设置 deleted_at）
- **AND** 同步删除 `sys_role_menu` 和 `sys_user_role` 中的关联记录
- **AND** 事务提交后再通知访问拓扑变更

#### Scenario: 删除已分配用户的角色
- **WHEN** 用户尝试删除一个已分配给用户的角色
- **THEN** 系统提示"该角色已分配给X个用户，是否确认删除？"
- **AND** 用户确认后在事务内删除角色，同时取消这些用户的角色分配

#### Scenario: 删除超级管理员角色
- **WHEN** 用户尝试删除 admin 角色
- **THEN** 系统返回错误信息"不能删除超级管理员角色"

#### Scenario: 关联清理失败触发整体回滚
- **WHEN** 用户删除一个角色
- **AND** 事务内 `sys_role_menu` 或 `sys_user_role` 关联清理返回错误
- **THEN** 角色软删除 MUST 一并回滚
- **AND** 操作返回底层错误而非仅记录 warning
- **AND** 不发出访问拓扑变更通知

### Requirement: 角色用户分配

系统 SHALL 支持为角色分配用户（"分配"功能）。批量授权与取消授权 MUST 使用单事务整体提交，所有插入或删除操作 MUST 整体成功或整体回滚，禁止逐条插入失败仅记录 warning 后继续处理后续条目。

#### Scenario: 查看角色的用户列表
- **WHEN** 用户点击角色的"分配"按钮
- **THEN** 系统跳转到角色用户管理页面
- **AND** 显示已分配该角色的用户列表
- **AND** 用户列表支持分页和搜索

#### Scenario: 取消用户的角色授权
- **WHEN** 用户在角色用户列表中点击"取消授权"
- **THEN** 系统在事务内删除 sys_user_role 中对应的关联记录
- **AND** 用户列表自动刷新

#### Scenario: 批量授权用户原子提交
- **WHEN** 用户选择多个未分配该角色的用户并点击"授权"
- **THEN** 系统在单事务内批量插入 `sys_user_role` 关联记录
- **AND** 任意一条插入失败 MUST 触发整体回滚
- **AND** 这些用户重新登录后可获得该角色的菜单权限
