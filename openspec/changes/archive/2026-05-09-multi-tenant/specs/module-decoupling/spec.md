## ADDED Requirements

### Requirement: multi-tenant 插件禁用时的联动隐藏
当 `multi-tenant` 插件未启用时,前端 SHALL 完全隐藏:
- 登录页的租户挑选器与租户输入框。
- 工作台头部的"当前租户"标识与切换器。
- 用户管理页面的"租户"列与筛选项。
- 所有 `/platform/*` 与 `/tenant/*` 入口菜单。

#### Scenario: 未启用时干净 UI
- **WHEN** multi-tenant 插件未启用
- **THEN** 工作台无任何"租户"相关 UI
- **AND** 等价于单租户系统的完整体验

### Requirement: 启用时联动显示
启用后 SHALL 联动显示对应 UI;状态切换通过 `IsEnabled(ctx, 'multi-tenant')` 判断,与现有"模块禁用"机制一致。

#### Scenario: 启用后立即生效
- **WHEN** 平台管理员安装并启用 multi-tenant
- **THEN** 平台管理员页面立即出现租户管理菜单
- **AND** 登录页流程切换为两阶段

### Requirement: 跨租户 UI 元素的双向隐藏
租户管理员视图 SHALL 隐藏所有平台级 UI(平台管理员菜单、impersonation 入口、跨租户报表);平台管理员视图 SHALL 同样隐藏租户级私有 UI(避免误操作)。

#### Scenario: 双向隔离
- **WHEN** 租户管理员登录工作台
- **THEN** 不可见 `/platform/*` 入口与 impersonation 按钮
- **AND** 平台管理员登录(tenant_id=0)时不可见仅租户内的菜单
