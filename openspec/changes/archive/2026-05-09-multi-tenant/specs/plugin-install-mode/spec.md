## ADDED Requirements

### Requirement: install_mode 字段语义
`sys_plugin.install_mode` SHALL 取值 `global` 或 `tenant_scoped`:
- `global`:插件对所有租户生效,启用/禁用由平台管理员控制(`sys_plugin_state.tenant_id=0 AND state_key='__tenant_enabled__'` 一行)。
- `tenant_scoped`:插件按租户独立启用,启用/禁用由各租户管理员控制(`sys_plugin_state.tenant_id>0 AND state_key='__tenant_enabled__'` 多行)。

#### Scenario: global 模式下的状态行
- **WHEN** 插件 P 安装为 `install_mode=global`
- **THEN** `sys_plugin_state` 仅一行 `(plugin_id=P, tenant_id=0, state_key='__tenant_enabled__', enabled=true|false)`
- **AND** 各租户的 `IsEnabled(ctx, P)` 都读取该单一状态

#### Scenario: tenant_scoped 模式下的状态行
- **WHEN** 插件 P 安装为 `install_mode=tenant_scoped`
- **THEN** `sys_plugin_state` 每个租户一行
- **AND** 租户 T 的 `IsEnabled(ctx, P)` 读 `(plugin_id=P, tenant_id=T, state_key='__tenant_enabled__')` 行;不存在则视为 `enabled=false`

### Requirement: 安装期 install_mode 选择
平台管理员安装 `tenant_aware` 插件时 SHALL 通过 UI 或 API 显式选择 `install_mode`;`platform_only` 插件 install_mode 强制为 `global` 不允许选择。

#### Scenario: 平台管理员选择 install_mode
- **WHEN** 平台管理员调用 `POST /platform/plugins {plugin_id, install_mode: tenant_scoped}`
- **AND** 该插件 `scope_nature = tenant_aware`
- **THEN** 安装成功,`sys_plugin.install_mode = tenant_scoped`
- **AND** 不为任何租户初始化 `sys_plugin_state` 行(等租户管理员主动 enable)

#### Scenario: 安装 platform_only 插件
- **WHEN** 平台管理员安装 `multi-tenant` 插件(`scope_nature = platform_only`)
- **THEN** 系统忽略请求中的 install_mode 字段
- **AND** 强制写入 `install_mode = global`

### Requirement: 新租户自动启用策略
`tenant_aware` 插件不得通过 plugin.yaml 声明新租户默认启用策略;平台插件系统 SHALL 通过 `sys_plugin.auto_enable_for_new_tenants` 维护该策略。当插件声明 `supports_multi_tenant=true`、已安装、宿主层已启用、`install_mode = tenant_scoped` 且 `auto_enable_for_new_tenants = true` 时,系统 SHALL 在新租户创建时自动初始化 `sys_plugin_state(tenant_id=新租户, state_key='__tenant_enabled__', enabled=true)`。

#### Scenario: 平台策略控制新租户默认启用
- **WHEN** `content-notice` 安装为 `tenant_scoped` 模式且宿主层已启用
- **AND** 平台管理员将 `auto_enable_for_new_tenants` 设置为 true
- **AND** 平台管理员创建新租户 T
- **THEN** 系统通过 `multi-tenant` 插件创建租户流程中的显式插件治理服务调用自动 insert `sys_plugin_state(plugin_id=content-notice, tenant_id=T, state_key='__tenant_enabled__', enabled=true)`
- **AND** 新租户管理员登录后立即可见该租户级插件菜单

#### Scenario: 插件同步不覆盖平台策略
- **WHEN** 平台管理员将插件 P 的 `auto_enable_for_new_tenants` 设置为 true
- **AND** 后续启动同步或手动同步 P 的 plugin.yaml
- **THEN** `sys_plugin.auto_enable_for_new_tenants` 保持平台管理员设置的值
- **AND** 不从 plugin.yaml 读取或覆盖该策略

### Requirement: install_mode 切换规则
install_mode 切换 SHALL 按以下规则执行:`global → tenant_scoped` 允许,需平台管理员确认,系统自动为所有 active 租户写入 `sys_plugin_state(enabled = 当前全局值)`;`tenant_scoped → global` 仅允许平台管理员执行,且 MUST 二次确认 + 强制审计,立即对所有租户强制启用;`platform_only ↔ tenant_aware` 运行时禁止切换,仅可通过插件升级版本变更。

#### Scenario: global 切 tenant_scoped 自动初始化
- **WHEN** 平台管理员将 `monitor-loginlog` 从 `global` 切到 `tenant_scoped`
- **AND** 该插件原本 `enabled=true`
- **AND** 系统当前有 5 个 active 租户
- **THEN** `sys_plugin_state` 新增 5 行 `(plugin_id=monitor-loginlog, tenant_id=Ti, state_key='__tenant_enabled__', enabled=true)`
- **AND** 原 `(tenant_id=0)` 行被删除或保留为参考(实现可选)

#### Scenario: tenant_scoped 切 global 二次确认
- **WHEN** 平台管理员将 `content-notice` 从 `tenant_scoped` 切到 `global`
- **THEN** UI 弹出二次确认对话框,提示"将对所有租户强制启用"
- **AND** 用户确认后写入 `sys_plugin_state(tenant_id=0, state_key='__tenant_enabled__', enabled=true)`
- **AND** 删除所有 `tenant_id>0` 的状态行
- **AND** 操作日志以 `oper_type='other'` 记录 install_mode 变更强审计,摘要或载荷包含 `plugin_install_mode_change`

#### Scenario: 租户管理员不能切换 install_mode
- **WHEN** 租户管理员尝试将某插件从 `tenant_scoped` 切换为 `global`
- **THEN** 返回 403 `bizerr.CodePlatformPermissionRequired`
- **AND** 不修改任何 `sys_plugin` 或 `sys_plugin_state` 行

### Requirement: install_mode 切换的前置否决
插件 SHALL 可通过实现 `LifecycleGuard.CanChangeInstallMode(ctx, from, to)` 钩子拒绝 install_mode 切换(可选);若返回 `ok=false`,平台管理员仍 MAY 走 `--force` 通道。

#### Scenario: 插件拒绝切换
- **WHEN** 某插件实现 `CanChangeInstallMode` 并在 `tenant_scoped → global` 时返回 `ok=false`
- **THEN** 平台管理员看到否决理由
- **AND** 必须显式 force 才能继续
