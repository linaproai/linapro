## ADDED Requirements

### Requirement: 权限点不编码平台/租户边界
权限点字符串 SHALL 只表达功能动作,不通过 `platform:*` 或 `tenant:*` 前缀编码平台/租户边界。宿主与官方多租户插件 SHALL 统一使用 `system:*` 权限点;平台/租户边界由路由平面、当前租户上下文、数据权限、插件启用状态与接口级校验共同决定。

#### Scenario: 统一权限点
- **WHEN** 平台租户管理 API 声明权限
- **THEN** 使用 `system:tenant:*` 权限点
- **AND** 不使用 `platform:tenant:*`
- **WHEN** 租户成员管理 API 声明权限
- **THEN** 使用 `system:tenant:member:*` 权限点
- **AND** 不使用 `tenant:member:*`

### Requirement: 权限点选择按上下文和资源可见性过滤
权限点选择 UI 与 API SHALL 按当前上下文、路由平面和插件启用状态过滤可见权限点,不得依赖权限字符串前缀区分平台/租户权限。

#### Scenario: 权限点选择器
- **WHEN** 租户管理员打开角色编辑页选择权限点
- **THEN** 列表仅包含当前租户上下文可配置的菜单与按钮权限
- **AND** 不通过 `platform:*` 前缀过滤

### Requirement: 权限解析按当前租户启用插件过滤
权限解析时 SHALL 排除"插件未在当前租户启用"的权限点(即使用户已被分配);避免显示用户实际无法操作的权限。

#### Scenario: 未启用插件的权限被过滤
- **WHEN** 用户在租户 A 持有 `monitor-online` 权限点
- **AND** `monitor-online` 在租户 A 中 disabled
- **THEN** 权限解析结果不含该权限点
- **AND** UI 上"在线用户"按钮不可见

### Requirement: 接口级权限校验保留
即便菜单/按钮已隐藏,后端接口级权限校验 SHALL 仍执行,确保安全;租户未启用插件 + 跨租户访问双重防护。

#### Scenario: 后端校验为最后一道防线
- **WHEN** 攻击者绕过前端菜单隐藏直接调用未启用插件的接口
- **THEN** 后端中间件先按 `(tenant, plugin)` enable 校验返回 404
- **AND** 即使 enable 后,权限校验仍按角色权限点过滤

### Requirement: 菜单按钮权限按页面实际按钮投影
插件菜单清单中的按钮权限 SHALL 只投影对应管理页面真实展示或可触发的按钮/入口。平台租户管理页面的按钮权限 SHALL 仅包含租户查询、新增、修改、删除、代操作和跳转系统用户管理的入口;租户解析、租户成员、租户插件等仅作为 API 能力存在时,不得作为租户管理页面的按钮权限挂载到该菜单下。

#### Scenario: 审查租户管理按钮权限
- **WHEN** 管理员在菜单管理页面展开"租户管理"
- **THEN** 按钮权限列表只展示页面实际按钮对应的权限点
- **AND** 不展示 `system:tenant:resolver:*`、`system:tenant:member:*` 或 `system:tenant:plugin:*`
