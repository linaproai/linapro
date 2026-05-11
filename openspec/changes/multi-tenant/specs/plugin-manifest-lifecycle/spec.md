## ADDED Requirements

### Requirement: plugin.yaml 多租户字段
plugin.yaml SHALL 增加以下字段:
- `scope_nature`(必填):`platform_only` 或 `tenant_aware`,详见 `plugin-scope-nature` 能力。
- `default_install_mode`(可选,scope_nature=tenant_aware 时):`global` 或 `tenant_scoped`,默认 `tenant_scoped`。

#### Scenario: 完整 manifest 字段
- **WHEN** 一个 tenant_aware 插件 `plugin.yaml` 含 `scope_nature, default_install_mode`
- **THEN** 安装期解析并写入 `sys_plugin` 对应字段
- **AND** 缺失可选字段按默认值处理

#### Scenario: 必填字段缺失
- **WHEN** plugin.yaml 缺 `scope_nature`
- **THEN** 安装失败,返回 `bizerr.CodePluginScopeNatureMissing`

### Requirement: 安装期一致性校验
安装时 SHALL 校验:
1. `scope_nature=platform_only` 仅可 `install_mode=global`。
2. `default_install_mode` 必须与 scope_nature 兼容(platform_only 不支持 tenant_scoped 默认)。

#### Scenario: 非法组合被拒
- **WHEN** 平台管理员尝试安装 `scope_nature=platform_only` 插件并指定 `install_mode=tenant_scoped`
- **THEN** 返回 `bizerr.CodePluginInstallModeInvalidForScopeNature`

### Requirement: sys_plugin 增加治理列
`sys_plugin` SHALL 增加 `scope_nature VARCHAR(32)`、`install_mode VARCHAR(32)` 与平台策略列 `auto_enable_for_new_tenants BOOL`,并作为平台全局插件注册目录保持不携带 `tenant_id`;`auto_enable_for_new_tenants` SHALL 由平台插件系统维护,不得由 plugin.yaml 同步覆盖;`sys_plugin_state` SHALL 保留 `id` 自增技术主键,增加 `tenant_id INT NOT NULL DEFAULT 0`,并使用 `(plugin_id, tenant_id, state_key)` 唯一索引表达业务唯一性,插件启用状态使用稳定 `state_key='__tenant_enabled__'` 行表达。

#### Scenario: 列存在性
- **WHEN** SQL 迁移完成
- **THEN** `sys_plugin` 包含 `scope_nature`、`install_mode` 与 `auto_enable_for_new_tenants` 列
- **AND** `sys_plugin` 不包含 `tenant_id` 列
- **AND** `sys_plugin_state` 包含 `id` 自增主键
- **AND** `sys_plugin_state` 包含 `(plugin_id, tenant_id, state_key)` 唯一索引

### Requirement: 卸载受 LifecycleGuard 否决保护
卸载流程 SHALL 在执行实际卸载步骤前调用所有插件的 `CanUninstall` 钩子;详见 `plugin-lifecycle-guard` 能力。

#### Scenario: 否决卸载
- **WHEN** 任意插件 `CanUninstall` 返回 false
- **THEN** 卸载被拒,聚合 reason 返回给平台管理员
