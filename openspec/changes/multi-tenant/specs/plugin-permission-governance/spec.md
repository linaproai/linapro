## ADDED Requirements

### Requirement: 平台与租户权限点分层
权限点字符串 SHALL 通过前缀区分作用域:
- `platform:*`:平台权限,仅可挂在 `is_platform_role=true` 角色上。
- 其他前缀(如 `system:*`、`monitor:*`、`<plugin>:*`):租户权限,仅可挂在租户角色上。

#### Scenario: 错误挂载被拒
- **WHEN** 创建租户角色时挂 `platform:*` 权限
- **THEN** 返回 `bizerr.CodePlatformPermissionOnTenantRole`

### Requirement: 租户管理员看不到平台权限
权限点选择 UI 与 API 在租户管理员视图中 SHALL 排除 `platform:*` 前缀;在平台管理员视图中显示全部并按前缀分组。

#### Scenario: 权限点选择器
- **WHEN** 租户管理员打开角色编辑页选择权限点
- **THEN** 列表中无 `platform:*` 选项

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
