## ADDED Requirements

### Requirement: sys_role 增加租户归属与数据权限
`sys_role` SHALL 新增以下字段:
- `tenant_id INT NOT NULL DEFAULT 0`:平台上下文角色 = 0,租户上下文角色 = 所属租户。
- `data_scope SMALLINT NOT NULL DEFAULT 2`:角色数据权限范围,其中 `1=全部数据`、`2=本租户数据`、`3=本部门数据`、`4=本人数据`。
- `sys_role` SHALL NOT 维护平台角色布尔标识;平台/租户管理员能力由当前 `bizctx.TenantId`、角色 `data_scope` 与 `system:*` 功能权限共同推导。

#### Scenario: 平台上下文角色与租户上下文角色共存
- **WHEN** 系统初始化
- **THEN** seed 数据写入"超级管理员"`(tenant_id=0, data_scope=1)`
- **AND** 创建租户 T 时自动 provision "租户管理员"模板角色 `(tenant_id=T, data_scope=2)`

### Requirement: 角色查询按层级隔离
租户上下文调用 `GET /role/list` SHALL 仅返回 `tenant_id = current_tenant` 的角色;平台上下文可通过平台治理接口查全量。

#### Scenario: 租户管理员视图
- **WHEN** 租户 A 管理员查询角色列表
- **THEN** 返回租户 A 的角色,不含平台上下文角色

#### Scenario: 平台管理员视图
- **WHEN** 平台管理员调用 `GET /platform/roles`
- **THEN** 返回所有租户的所有角色,带 `tenant_id` 列展示

### Requirement: 权限点统一 system 命名空间
权限点字符串 SHALL 只表达"资源 + 动作",不表达平台/租户边界。宿主与多租户插件内置管理权限 SHALL 统一使用 `system:*` 形态,例如 `system:tenant:list`、`system:tenant:impersonate`、`system:tenant:member:list` 与 `system:tenant:plugin:enable`。平台/租户边界 SHALL 由 API 路由平面、当前 `bizctx.TenantId`、数据权限与菜单/插件启用状态共同约束。

#### Scenario: 相同权限在不同上下文下生效范围不同
- **WHEN** 平台上下文用户持有 `system:user:list` 且数据权限为全部数据
- **THEN** 平台 API 可返回跨租户用户聚合视图
- **WHEN** 租户 A 上下文用户持有 `system:user:list` 且数据权限为本租户数据
- **THEN** 普通系统 API 仅返回租户 A 内用户

### Requirement: 租户上下文禁止全部数据
`data_scope=1` SHALL 表示平台全局的全部数据,仅允许在 `tenant_id=0` 的平台上下文角色中配置。租户上下文角色 SHALL 只能配置 `data_scope` 为 `2`、`3` 或 `4`。

#### Scenario: 租户角色配置全部数据被拒
- **WHEN** 租户 A 管理员创建或更新角色并选择 `data_scope=1`
- **THEN** 返回 `bizerr.CodeTenantRoleAllDataScopeForbidden`
- **AND** 角色写入被拒

### Requirement: 角色数据权限选项按租户能力展示
角色管理页面 SHALL 仅在多租户插件已安装且启用时展示和允许选择 `data_scope=2` 的"本租户数据"选项。多租户插件未启用时,角色管理列表、创建抽屉与编辑抽屉 SHALL 不展示"本租户数据"文案或选项,并将创建默认值回退为"全部数据"。

#### Scenario: 单租户模式隐藏本租户数据
- **WHEN** 多租户插件未启用
- **AND** 用户打开角色管理页面或新增角色抽屉
- **THEN** 数据权限下拉框不包含"本租户数据"
- **AND** 新增角色数据权限默认选中"全部数据"

#### Scenario: 多租户启用时展示本租户数据
- **WHEN** 多租户插件已启用
- **AND** 用户打开角色管理页面或新增角色抽屉
- **THEN** 数据权限下拉框包含"本租户数据"
- **AND** 新增角色数据权限默认选中"本租户数据"

### Requirement: 角色名/编码作用域
角色 `name` 与 `code` SHALL 在 `(tenant_id, code)` 上唯一(平台上下文角色全局唯一,租户上下文角色租户内唯一);不同租户可重复使用同名角色。

#### Scenario: 同名角色跨租户允许
- **WHEN** 租户 A 创建角色 `code=manager`
- **AND** 租户 B 也创建角色 `code=manager`
- **THEN** 两次创建均成功
- **AND** 各自独立存在

### Requirement: 删除角色的租户校验
租户管理员 SHALL 仅可删除本租户角色;平台管理员 MAY 删除任意角色;删除前 MUST 校验该角色未绑定任何用户(否则需先解绑或显式 cascade)。

#### Scenario: 跨租户删除角色被拒
- **WHEN** 租户 A 管理员尝试删除租户 B 的角色
- **THEN** 返回 `bizerr.CodeRoleTenantForbidden`
