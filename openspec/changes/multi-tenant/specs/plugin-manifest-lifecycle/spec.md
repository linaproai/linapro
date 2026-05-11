## ADDED Requirements

### Requirement: plugin.yaml 多租户字段
plugin.yaml SHALL 增加以下字段:
- `scope_nature`(必填):`platform_only` 或 `tenant_aware`,详见 `plugin-scope-nature` 能力。
- `supports_multi_tenant`(必填):布尔值,表示插件是否支持租户级安装、租户级启用与新租户开通策略治理;`scope_nature=platform_only` 时必须为 `false`。
- `default_install_mode`(可选,scope_nature=tenant_aware 时):`global` 或 `tenant_scoped`,默认 `tenant_scoped`。

#### Scenario: 完整 manifest 字段
- **WHEN** 一个 tenant_aware 插件 `plugin.yaml` 含 `scope_nature, supports_multi_tenant, default_install_mode`
- **THEN** 安装期解析并写入 `sys_plugin` 对应字段
- **AND** 缺失可选字段按默认值处理

#### Scenario: 必填字段缺失
- **WHEN** plugin.yaml 缺 `scope_nature`
- **THEN** 安装失败,返回 `bizerr.CodePluginScopeNatureMissing`

### Requirement: 安装期一致性校验
安装时 SHALL 校验:
1. `scope_nature=platform_only` 仅可 `install_mode=global`。
2. `scope_nature=platform_only` 时 `supports_multi_tenant` 必须为 `false`。
3. `supports_multi_tenant=false` 时 `default_install_mode` 强制为 `global`,不得启用租户级安装或新租户开通策略。
4. `default_install_mode` 必须与 scope_nature 和 supports_multi_tenant 兼容(platform_only 不支持 tenant_scoped 默认)。

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

### Requirement: 动态插件孤儿卸载
动态插件已安装但 staging artifact 与 active release artifact 均不可用时,宿主 SHALL 允许平台管理员通过 `force=true` 执行受限孤儿卸载;该路径 SHALL 清理宿主治理状态、菜单权限、资源引用、运行时缓存与节点投影,但 SHALL 明确跳过插件内嵌 uninstall SQL、Wasm 生命周期逻辑和插件授权存储清理。

#### Scenario: artifact 缺失时普通卸载失败
- **WHEN** 动态插件已安装
- **AND** staging artifact 与 active release artifact 均不存在或不可解析
- **AND** 平台管理员未使用 `force=true`
- **THEN** 卸载失败并提示 artifact 缺失

#### Scenario: force 孤儿卸载清理宿主治理状态
- **WHEN** 动态插件已安装
- **AND** staging artifact 与 active release artifact 均不存在或不可解析
- **AND** 平台管理员使用 `force=true`
- **THEN** 宿主将插件 registry 置为未安装/禁用并解除 active release 绑定
- **AND** 删除插件宿主菜单权限与 `sys_plugin_resource_ref`
- **AND** 将 release 状态标记为 `uninstalled`
- **AND** 记录节点投影消息说明已跳过插件内嵌 uninstall SQL 与插件侧存储清理
