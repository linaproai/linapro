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
