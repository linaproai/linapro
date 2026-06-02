## ADDED Requirements

### Requirement: 宿主插件设置写入必须为数据库原子 upsert

系统 SHALL 将宿主 `PluginSettingsService.SetString` 与 `SetSecret` 的写入实现为 PostgreSQL 原生 `INSERT ... ON CONFLICT (tenant_id, key) DO UPDATE SET ...`，确保对同一 `(pluginID, key)` 的并发写入不会因为应用层 count-then-insert/update 分支造成唯一键冲突或重复行。系统 MUST NOT 在写入路径中先 `Count()` 再 `Insert()`/`Update()`，也 MUST NOT 在写入数据里手写 `created_at` 或 `updated_at`，时间字段必须由 GoFrame 自动时间策略维护。

#### Scenario: 并发写入同一插件设置键

- **WHEN** 两个 admin 进程几乎同时对同一 `pluginID` 和同一 `key` 调用 `SetString`
- **THEN** 两次写入只产生一条 `sys_config` 行
- **AND** 写入路径不会因为应用层 count-then-insert/update 顺序出现唯一键冲突
- **AND** 写入路径不会手写 `created_at` 或 `updated_at`

#### Scenario: 写入路径不覆盖宿主治理位

- **WHEN** 插件通过 `SetString` 写入一条新值
- **AND** 这条 `(tenant_id, key)` 已经存在 `is_builtin=1` 的内置行
- **THEN** 系统不得在 `ON CONFLICT DO UPDATE` 时覆盖 `is_builtin`
- **AND** 系统不得通过插件 KV 写入路径绕过宿主内置保护

### Requirement: 宿主插件设置 upsert 必须自动恢复软删除残留

系统 SHALL 在宿主 `PluginSettingsService` 的 upsert 路径中将 `deleted_at` 列纳入 `ON CONFLICT DO UPDATE` 列表，使新值落在任何此前被软删除的 `(tenant_id, key)` 行上时，行的 `deleted_at` 同步重置回 `NULL`，并对 `GetString`、`List` 与 `GetMaskedSecret` 重新可见。系统 MUST NOT 让一条软删除残留行长期阻塞同一 `(tenant_id, key)` 的后续写入对调用方可见，也 MUST NOT 要求运维通过手动 SQL 才能恢复被软删除的插件设置。

#### Scenario: 软删除残留行被 upsert 自愈

- **WHEN** `sys_config` 中已存在一条 `(tenant_id=0, key='linapro-oidc-google.backendRedirects')` 且 `deleted_at IS NOT NULL` 的残留行
- **AND** 调用 `SetString(ctx, 'linapro-oidc-google', 'backendRedirects', '{"dashboard":"/x"}')`
- **THEN** `sys_config` 中该行的 `deleted_at` 被重置为 `NULL`
- **AND** 该行 `value` 被更新为新值
- **AND** 同上下文 `GetString(ctx, 'linapro-oidc-google', 'backendRedirects', '')` 返回新值
- **AND** 同上下文 `List(ctx, 'linapro-oidc-google')` 包含新值

#### Scenario: 同 `(tenant_id, key)` 无残留行时的常规写入

- **WHEN** `sys_config` 中不存在 `(tenant_id=0, key='linapro-oidc-google.backendRedirects')` 行
- **AND** 调用 `SetString(ctx, 'linapro-oidc-google', 'backendRedirects', '{"dashboard":"/x"}')`
- **THEN** 系统 INSERT 一条新行
- **AND** 新行的 `deleted_at` 为 `NULL`
- **AND** 后续 `GetString` 与 `List` 立即可见

### Requirement: 宿主插件设置清空必须物理删除底层行

系统 SHALL 将 `SetString(ctx, pluginID, key, "")` 与 `Delete(ctx, pluginID, key)` 路径实现为 `Unscoped()` 物理删除，使 `sys_config` 中对应行被从数据库表中移除而不是仅写入 `deleted_at`。系统 MUST NOT 在插件 KV 设置的清空路径上保留 GoFrame 默认的软删除行为，因为软删除的残留行会与 `(tenant_id, key)` 唯一索引共同导致后续 upsert 在不可见的行上原地更新。

#### Scenario: SetString 空值清空

- **WHEN** 调用 `SetString(ctx, 'linapro-oidc-google', 'backendRedirects', '')`
- **THEN** `sys_config` 中 `(tenant_id=0, key='linapro-oidc-google.backendRedirects')` 行被物理删除
- **AND** 即使绕过软删除过滤的 `Unscoped()` 查询也查不到这行

#### Scenario: 清空后再次写入重新可见

- **WHEN** 调用 `SetString(ctx, 'linapro-oidc-google', 'backendRedirects', '')`
- **AND** 调用 `SetString(ctx, 'linapro-oidc-google', 'backendRedirects', '{"profile":"/p"}')`
- **THEN** `GetString(ctx, 'linapro-oidc-google', 'backendRedirects', '')` 返回新值
- **AND** `List(ctx, 'linapro-oidc-google')` 中包含新值
