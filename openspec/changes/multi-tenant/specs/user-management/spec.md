## ADDED Requirements

### Requirement: sys_user 增加租户身份字段
`sys_user` SHALL 新增 `tenant_id INT NOT NULL DEFAULT 0` 字段:
- `tenant_id = 0` 表示平台用户(平台管理员、单租户模式下的所有用户)。
- `tenant_id > 0` 表示该用户的"主租户/默认登录租户"。
- 与 1:N membership 模型并存:`tenant_id` 决定登录默认进入的租户,membership 决定可访问哪些租户。

#### Scenario: 平台管理员的 tenant_id
- **WHEN** 平台管理员 admin 用户存在
- **THEN** `sys_user.tenant_id = 0`
- **AND** 不允许在 `plugin_multi_tenant_user_membership` 中出现 admin 的行

#### Scenario: 租户用户的主租户
- **WHEN** 用户 U 的 1:N membership 包含 [A, B, C]
- **AND** `sys_user.tenant_id = A`
- **THEN** U 登录无 hint 时默认进入租户 A
- **AND** U 仍可通过挑选器或 switch-tenant 进入 B 或 C

### Requirement: 用户查询按租户隔离
多租户启用时,用户列表/详情查询 SHALL 以 `plugin_multi_tenant_user_membership` 作为租户可见性的权威边界;`sys_user.tenant_id` 仅表示主租户/默认登录租户,不得作为用户列表的唯一过滤条件。租户 A 管理员仅可见 `membership.tenant_id = A AND status = active` 的用户;平台管理员通过系统用户管理页查询全量用户,并可按租户归属筛选。

#### Scenario: 租户管理员查询用户列表
- **WHEN** 租户 A 管理员调用 `GET /user/list`
- **THEN** 返回 `plugin_multi_tenant_user_membership.tenant_id = A AND status = active` 关联的所有用户
- **AND** 不返回任何与租户 A 无关的用户

#### Scenario: 主租户不同但 membership 命中
- **WHEN** 用户 U 的 `sys_user.tenant_id = A`
- **AND** U 同时拥有租户 B 的 active membership
- **AND** 租户 B 管理员调用 `GET /user/list`
- **THEN** 返回结果包含 U
- **AND** U 在租户 B 内的角色、数据权限与状态均按 B 的 membership 与 `sys_user_role.tenant_id=B` 解析

#### Scenario: 平台管理员查询全量用户
- **WHEN** 平台管理员调用 `GET /user`
- **THEN** 返回全租户用户列表,带 `tenantIds` 与 `tenantNames` 字段展示用户所属租户

#### Scenario: 平台管理员按租户筛选用户列表
- **WHEN** 平台管理员在用户管理页面选择租户 A 作为筛选条件
- **THEN** 前端调用 `GET /user?tenantId=A`
- **AND** 后端按 `plugin_multi_tenant_user_membership.tenant_id = A AND status = active` 筛选用户
- **AND** 不返回仅属于其他租户的用户

#### Scenario: 用户新增与编辑抽屉展示租户归属
- **WHEN** 多租户启用且用户打开系统用户管理的新增或编辑抽屉
- **THEN** 抽屉 SHALL 展示"所属租户"字段
- **AND** 平台上下文 SHALL 允许维护用户的 `tenantIds` membership 列表
- **AND** 租户上下文 SHALL 只展示并锁定当前租户归属,不得允许通过请求参数跨租户写入

#### Scenario: 操作用户绑定租户时租户候选项收敛
- **WHEN** 当前操作用户存在 active tenant membership
- **THEN** 用户管理页面中所有租户筛选、展示为可选列表的租户控件 SHALL 只展示该操作用户所属的 active 租户
- **AND** 新增/编辑用户时提交的 `tenantIds` SHALL 只能包含该操作用户所属的 active 租户
- **AND** 若请求包含其他租户,后端 SHALL 拒绝写入并返回跨租户错误
- **AND** 只有没有 active tenant membership 的平台全局管理员才可以在用户管理页面中查看并选择全量租户列表

### Requirement: 用户创建/导入按租户写入
租户管理员通过 `POST /user` 与 `POST /user/import` 创建用户 SHALL 自动写入 `tenant_id = bizctx.TenantId`;同时自动创建一条该租户 membership;不允许跨租户写入。

#### Scenario: 租户内创建用户
- **WHEN** 租户 A 管理员创建用户 U
- **THEN** `sys_user.tenant_id = A`
- **AND** 自动创建 `plugin_multi_tenant_user_membership(user_id=U, tenant_id=A, status=active)` 一行

#### Scenario: 平台管理员创建租户用户
- **WHEN** 平台管理员在系统用户管理新增抽屉选择租户 A 和租户 B 后创建用户 U
- **THEN** 后端写入 `plugin_multi_tenant_user_membership(user_id=U, tenant_id=A, status=active)` 与 `tenant_id=B` 两行
- **AND** `sys_user.tenant_id` SHALL 使用所选租户列表中的第一个租户作为默认登录租户
- **AND** 若平台管理员不选择任何租户,则创建平台用户且不写入 membership

#### Scenario: 跨租户创建被拒
- **WHEN** 租户 A 管理员尝试在请求中显式指定 `tenant_id = B`
- **THEN** 返回 `bizerr.CodeCrossTenantNotAllowed`
- **AND** 不写入数据

#### Scenario: 平台管理员编辑用户租户归属
- **WHEN** 平台管理员在系统用户管理编辑抽屉将用户 U 的所属租户改为租户 B
- **THEN** 后端 SHALL 在事务中以请求的 `tenantIds` 替换 U 的 active membership
- **AND** 同步更新 `sys_user.tenant_id` 为请求列表的第一个租户或 `0`
- **AND** 详情接口 SHALL 返回最新 `tenantIds` 与 `tenantNames` 供编辑抽屉回显

### Requirement: 用户批量编辑状态、角色和租户归属
用户管理页面 SHALL 支持对表格中已选中的多个用户执行批量编辑,可选择性更新用户状态、用户角色和所属租户。未选择更新的字段 SHALL 保持原值不变。后端 SHALL 使用单事务处理整个批量编辑请求,任一目标用户不可见、越权、包含当前登录用户或包含非法角色/租户时 SHALL 整体拒绝并不得部分写入。

#### Scenario: 批量编辑选中用户状态
- **WHEN** 管理员在用户管理页面选中多个非当前登录用户
- **AND** 打开批量编辑弹窗并选择更新用户状态为禁用
- **THEN** 前端 SHALL 调用 `PUT /user`
- **AND** 后端 SHALL 在数据权限和租户可见性校验通过后批量更新这些用户的 `sys_user.status`
- **AND** 若选中用户包含当前登录用户,后端 SHALL 返回 `CodeUserCurrentEditDenied` 并不更新任何用户

#### Scenario: 批量分配用户角色
- **WHEN** 管理员在批量编辑弹窗选择一个或多个角色
- **THEN** 后端 SHALL 以请求的 `roleIds` 替换每个目标用户在当前租户边界内的 `sys_user_role` 关系
- **AND** 角色候选项 SHALL 只包含当前租户/平台上下文下启用且可见的角色
- **AND** 若请求包含当前上下文不可见或跨租户角色,后端 SHALL 整体拒绝并不更新任何用户
- **AND** 若同一批量请求同时选择更新用户角色和所属租户,前端与后端 SHALL 拒绝该组合,要求分两次批量操作完成

#### Scenario: 批量编辑用户所属租户
- **WHEN** 多租户启用且平台管理员在批量编辑弹窗选择所属租户
- **THEN** 后端 SHALL 在事务中以请求的 `tenantIds` 替换每个目标用户的 active membership
- **AND** 同步更新每个目标用户的 `sys_user.tenant_id` 为请求列表中的第一个租户或 `0`
- **AND** 租户候选项 SHALL 继续遵循当前操作用户 active tenant membership 收敛规则
- **AND** 租户上下文 SHALL 只允许提交当前租户归属,不得通过批量编辑跨租户写入

### Requirement: 用户名全局唯一
`sys_user.username` SHALL 保持全局唯一(与单租户时代一致),不按租户分片;不同租户不能有相同 username。

#### Scenario: 用户名冲突跨租户
- **WHEN** 租户 A 中已有 username=`alice`
- **AND** 租户 B 管理员尝试创建 username=`alice`
- **THEN** 返回 `bizerr.CodeUserUsernameDuplicated`
- **AND** 不写入

### Requirement: 邀请已有用户加入租户
`POST /tenant/members/invite {username | user_id}` SHALL 允许租户管理员邀请已存在的全局用户加入本租户(仅创建 membership,不创建新 sys_user 行)。

#### Scenario: 邀请已存在用户
- **WHEN** 租户 B 管理员邀请已属租户 A 的用户 U(默认 `multi` 策略)
- **THEN** 创建 `(U, B, status=active)` membership
- **AND** U 登录后挑选器中出现租户 B
- **AND** sys_user 表不创建新行

#### Scenario: single 策略下邀请被拒
- **WHEN** `single` 策略启用且用户 U 已有 active membership 在租户 A
- **AND** 租户 B 管理员邀请 U
- **THEN** 返回 `bizerr.CodeMembershipExceedsCardinality`
