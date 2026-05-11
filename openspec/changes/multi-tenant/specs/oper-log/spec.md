## ADDED Requirements

### Requirement: 操作日志加租户与 impersonation 字段
`monitor-operlog` 表 SHALL 增加 `tenant_id`、`acting_user_id`、`on_behalf_of_tenant_id`、`is_impersonation` 字段(语义同 login-log)。所有受保护操作(包括 lifecycle guard、force action、tenant lifecycle)必须记录。

#### Scenario: 普通业务操作
- **WHEN** 租户 A 用户 U 修改自己资料
- **THEN** operlog `(tenant_id=A, user_id=U, oper_type='update')`

### Requirement: 强制操作单独审计类
平台管理员通过 `--force` 通道执行的操作 SHALL 写入 `oper_type='other'`;`oper_summary`、route metadata 或载荷包含 `platform_force_action`、被绕过的所有 reason key 与上下文(plugin_id 或 tenant_id)。

#### Scenario: force-uninstall 审计
- **WHEN** 平台管理员强制卸载某插件
- **THEN** operlog 写入 `oper_type='other', oper_summary='platform_force_action', payload={plugin_id, bypassed_reasons:[...]}`
- **AND** 该日志即使在普通查询接口也清晰可见

### Requirement: lifecycle_guard 调用单独审计类
所有 LifecycleGuard 钩子调用 SHALL 在 operlog 中以 `oper_type='other'` 记录,`oper_summary`、route metadata 或载荷标记 `lifecycle_guard`,载荷含 `plugin_id`、`hook_method`、`ok`、`reason`、`elapsed_ms`、`timeout`、`panic` 标志。

#### Scenario: 钩子调用记录
- **WHEN** 系统执行 `CanUninstall(multi-tenant)` 返回 false
- **THEN** operlog 写入 `(oper_type='other', oper_summary='lifecycle_guard', payload={plugin_id:'multi-tenant', hook:'CanUninstall', ok:false, reason:'tenants_exist'})`

### Requirement: 操作日志查询按租户隔离
操作日志查询接口 SHALL 经 `tenantcap.Apply` 过滤;租户管理员仅可见 `tenant_id=current` 的日志。平台管理员仅通过 `/platform/oper-log` 管理平台接口查看全量,且 MUST 支持按 `acting_user_id`、`tenant_id`、`oper_type`、`is_impersonation` 等组合筛选;impersonation 模式下仍仅可见目标租户日志。

#### Scenario: 平台运维视图
- **WHEN** 平台管理员调查"上周谁用了 force"
- **THEN** 在 `/platform/oper-log` 按 `oper_type='other' AND oper_summary='platform_force_action' AND created_at >= 上周一` 查询
- **AND** 返回所有 force 操作清单
