## ADDED Requirements

### Requirement: LifecycleGuard 接口族
宿主 SHALL 定义以下独立接口供插件按需实现,每个接口仅声明一个否决钩子方法:
- `CanUninstaller.CanUninstall(ctx) (ok bool, reason string, err error)`
- `CanDisabler.CanDisable(ctx) (ok, reason, err)`
- `CanTenantDisabler.CanTenantDisable(ctx, tenantID int) (ok, reason, err)`
- `CanTenantDeleter.CanTenantDelete(ctx, tenantID int) (ok, reason, err)`
- `CanInstallModeChanger.CanChangeInstallMode(ctx, from, to string) (ok, reason, err)`

#### Scenario: 插件按需实现
- **WHEN** `multi-tenant` 插件实现 `CanUninstall` 与 `CanTenantDelete`,不实现其他
- **THEN** 宿主在执行卸载与租户删除前调用对应钩子
- **AND** 调用 `CanDisable` 时由于未实现,默认 `ok=true`

### Requirement: 否决聚合与展示
宿主 SHALL 在执行受保护动作前并发调用所有相关插件的对应钩子(超时 5s/钩子);任何一个返回 `ok=false` 即拒绝;**不短路**,所有 reason 聚合返回。

#### Scenario: 多插件多否决
- **WHEN** 平台管理员尝试卸载 `multi-tenant`
- **AND** 钩子 A 返回 `(false, "tenants_exist", nil)`
- **AND** 钩子 B 返回 `(false, "billing_pending", nil)`
- **THEN** 系统拒绝卸载
- **AND** UI 同时展示 A 与 B 的 reason

#### Scenario: 钩子超时
- **WHEN** 某插件 `CanUninstall` 5s 内未返回
- **THEN** 系统视为 `ok=false`
- **AND** reason = `"plugin.<id>.guard.timeout"`(i18n key)
- **AND** 操作日志记录超时事件

#### Scenario: 钩子 panic
- **WHEN** 某插件钩子 panic
- **THEN** 系统 recover 并视为 `ok=false`
- **AND** reason = `"plugin.<id>.guard.panic"`
- **AND** 错误日志记录 panic 栈

### Requirement: reason 必须是 i18n key
所有否决理由 SHALL 返回 i18n key 字符串(如 `"plugin.multi-tenant.uninstall_blocked.tenants_exist"`),禁止返回硬编码中文/英文文案;前端按 `bizctx.Locale` 渲染对应翻译。

#### Scenario: 后台展示本地化 reason
- **WHEN** 中文环境的平台管理员触发否决
- **THEN** 后台前端按 `zh-CN` locale 加载对应翻译并展示
- **AND** 同一 reason 在 `en-US` 环境展示英文翻译

#### Scenario: 缺失翻译时的降级
- **WHEN** 某 reason key 在当前 locale 下未注册
- **THEN** fallback 到 `en-US` 翻译
- **AND** 同时在错误日志中记录翻译缺失,以触发后续补全

### Requirement: 平台管理员 --force 通道
平台管理员 SHALL 可通过 `?force=true` query 参数与 UI 二次确认绕过否决执行受保护动作;`--force` 操作必须满足:
1. UI 强制要求文本输入插件 ID 二次确认。
2. 操作日志写入 `oper_type='other'`,并在摘要或载荷中标记 `platform_force_action`,完整记录被绕过的所有 reason 与触发用户。
3. 配置 `plugin.allow_force_uninstall: false` 时 `--force` 通道关闭(严格合规模式)。

#### Scenario: 平台管理员 force-uninstall
- **WHEN** 卸载被否决,平台管理员选择"强制卸载"
- **AND** 二次输入插件 ID 确认
- **THEN** 卸载继续执行
- **AND** 操作日志双轨记录(`oper_type='other'` + `platform_force_action` 摘要或载荷 + 完整 reason 列表)

#### Scenario: 严格合规模式下禁用 force
- **WHEN** `plugin.allow_force_uninstall = false`
- **AND** 平台管理员尝试 force
- **THEN** 返回 `bizerr.CodeForceActionForbidden`
- **AND** 卸载不执行

### Requirement: 钩子并发与超时
所有否决钩子 SHALL 并发执行(避免串行慢),单个钩子超时 5s,总超时 10s;超过总超时仍未返回的整体视为否决。

#### Scenario: 大量插件钩子
- **WHEN** 系统已安装 30 个插件,触发 `CanTenantDelete`
- **THEN** 30 个钩子并发调用
- **AND** 总耗时不超过 10s(若任意钩子未返回则按超时处理)

### Requirement: 钩子调用的可观察性
所有否决钩子调用 SHALL 在 `monitor-operlog` 中以 `oper_type='other'` 记录,并在摘要或载荷中标记 `lifecycle_guard`,包括:`plugin_id`、`hook_method`、`ok`、`reason`、`elapsed_ms`、是否 force、是否超时、是否 panic。

#### Scenario: 审计可追溯
- **WHEN** 运维事后排查"租户 T 为何被拒绝删除"
- **THEN** 在 operlog 中可按 `tenant_id=T, oper_type='other'` 并结合 `lifecycle_guard` 摘要或载荷检索到所有调用记录
- **AND** 看到每个插件的具体 ok/reason
