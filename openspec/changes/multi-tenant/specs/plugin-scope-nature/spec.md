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
