## ADDED Requirements

### Requirement: plugin.yaml 必填 scope_nature 字段
所有源码插件与动态插件的 `plugin.yaml` SHALL 包含 `scope_nature` 字段,值仅可为 `platform_only` 或 `tenant_aware`;缺失或非法值导致插件安装失败。

#### Scenario: 缺失 scope_nature
- **WHEN** 平台管理员尝试安装一个 `plugin.yaml` 不含 `scope_nature` 的插件
- **THEN** 返回 `bizerr.CodePluginScopeNatureMissing`
- **AND** 安装被拒,提示插件作者补全 manifest

#### Scenario: 非法 scope_nature 值
- **WHEN** 插件声明 `scope_nature: both`
- **THEN** 安装期校验失败,返回 `bizerr.CodePluginScopeNatureInvalid`
- **AND** 校验提示"仅支持 platform_only 或 tenant_aware"

### Requirement: scope_nature 语义
`scope_nature` 字段值 SHALL 按以下语义解释:`platform_only` 表示插件仅在平台层面运行影响所有租户,租户管理员 MUST NOT 看到该插件也 MUST NOT 控制其启用,`install_mode` 强制为 `global`;`tenant_aware` 表示插件支持平台级或租户级运行,`install_mode` 由平台管理员安装时选择 `global` 或 `tenant_scoped`。

#### Scenario: platform_only 插件对租户管理员不可见
- **WHEN** 租户管理员查询 `GET /tenant/plugins`
- **THEN** 返回结果不包含 `scope_nature = platform_only` 的插件
- **AND** 即使该插件已 enabled,也不展示在租户视图

### Requirement: plugin.yaml 多租户支持声明
所有源码插件与动态插件的 `plugin.yaml` SHALL 包含 `supports_multi_tenant` 布尔字段。该字段为 `true` 时表示插件支持租户级安装、租户级启用与新租户开通策略治理;该字段为 `false` 时,平台插件管理页 SHALL 展示“支持多租户=否”,并禁用“新租户启用”开关。

#### Scenario: 多租户支持列展示
- **WHEN** 平台管理员打开插件管理页面
- **THEN** 列表在“新租户启用”左侧展示“支持多租户”列
- **AND** 支持租户级治理的插件显示“是”
- **AND** 不支持租户级治理的插件显示“否”

#### Scenario: 不支持多租户的插件禁用新租户启用
- **WHEN** 插件 `supports_multi_tenant=false`
- **THEN** 插件管理页面中的“新租户启用”开关处于禁用状态
- **AND** 平台管理员不能通过该开关开启或关闭新租户开通策略

### Requirement: 插件详情治理元数据展示
平台插件管理页的插件详情弹窗 SHALL 展示与列表和安装治理一致的插件元数据快照,至少包括示例数据、支持多租户、新租户启用、作用域性质与安装模式,避免平台管理员必须回到列表页或安装弹窗才能判断插件治理边界。

#### Scenario: 详情弹窗展示新增插件治理元数据
- **WHEN** 平台管理员打开插件详情弹窗
- **THEN** 弹窗展示“示例数据”字段,按 `hasMockData` 显示“是/否”
- **AND** 弹窗展示“支持多租户”字段,按 `supportsMultiTenant` 显示“是/否”
- **AND** 弹窗展示“新租户启用”字段,按 `autoEnableForNewTenants` 显示“是/否”
- **AND** 弹窗展示“作用域性质”字段,按 `scopeNature` 显示 `platform_only` 或 `tenant_aware` 的本地化标签
- **AND** 弹窗展示“安装模式”字段,按 `installMode` 显示 `global` 或 `tenant_scoped` 的本地化标签

### Requirement: scope_nature 不可变
插件一旦安装,其 `scope_nature` SHALL 不可在运行时修改;只能通过插件升级到新版本(且新版本 manifest 中 scope_nature 不同)时才允许变更,且必须通过迁移脚本处理旧状态。

#### Scenario: 升级时 scope_nature 变更
- **WHEN** 插件从 v1(`platform_only`)升级到 v2(`tenant_aware`)
- **THEN** 升级流程检测到变更,要求平台管理员确认
- **AND** 提供迁移说明(如何从全局 enable 状态迁移到 per-tenant)
- **AND** 升级后 `sys_plugin.scope_nature` 与 `install_mode` 同步更新

### Requirement: 启动期一致性校验
框架启动期 SHALL 校验 `sys_plugin.scope_nature` 与 `sys_plugin.install_mode` 的组合合法:
- `platform_only` 必须 `install_mode = global`,否则 panic 启动失败。
- `tenant_aware` `install_mode` 可为 `global` 或 `tenant_scoped`。

#### Scenario: 不一致状态导致启动失败
- **WHEN** 由于 SQL 直接修改导致 `(scope_nature=platform_only, install_mode=tenant_scoped)` 这种非法组合
- **THEN** 启动期检查报错,服务拒绝启动
- **AND** 错误日志明确指出具体插件 id 与建议修复方法

### Requirement: 安装期 scope_nature 拷贝
插件安装时,系统 SHALL 从 `plugin.yaml` 拷贝 `scope_nature` 写入 `sys_plugin.scope_nature`;运行时该列只读(除非升级版本)。

#### Scenario: 安装期写入 sys_plugin
- **WHEN** 平台管理员安装插件 `org-center`(plugin.yaml 声明 `scope_nature: tenant_aware`)
- **THEN** `sys_plugin` 新增一行 `scope_nature='tenant_aware'`
- **AND** 该列后续运行时不允许 UPDATE
