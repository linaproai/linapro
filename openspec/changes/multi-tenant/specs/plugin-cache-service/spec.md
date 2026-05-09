## ADDED Requirements

### Requirement: 插件缓存 key 默认携带租户维度
插件通过 host service 调用的缓存读写接口 SHALL 默认在 cache key 上前缀 `tenant=<id>:`;插件可在管理平台模式下显式声明 `scope=platform` 来访问跨租户共享缓存(需平台管理员权限与审计),但默认安全选项是租户维度。impersonation 模式仍使用目标租户维度缓存。

#### Scenario: 默认租户维度
- **WHEN** 插件调用 `cache.Set("foo", value)` 在租户 A 上下文
- **THEN** 实际 key = `plugin-<id>:tenant=A:foo`
- **AND** 租户 B 上下文调 `cache.Get("foo")` 不命中

#### Scenario: 显式平台共享
- **WHEN** 插件调用 `cache.SetPlatform("foo", value)` 且具备 `platform:cache:write` 权限
- **THEN** 实际 key = `plugin-<id>:tenant=0:foo`
- **AND** 各租户共享读写

### Requirement: 失效广播按租户精细化
插件调用 `cache.Invalidate(scope, key)` SHALL 在 distributed-cache-coordination 失效消息中携带当前租户;`cache.InvalidatePlatform(...)` 才广播全租户级联失效。

#### Scenario: 单租户失效
- **WHEN** 租户 A 的插件调用 `cache.Invalidate("dict", "type1")`
- **THEN** 仅 `tenant=A` 缓存失效

### Requirement: 平台共享缓存的审计要求
任何对 `tenant=0` 的写入 SHALL 在 operlog 中记录 `action_kind='platform_cache_write'`;失效操作同样审计。

#### Scenario: 审计可追溯
- **WHEN** 插件 P 在租户 A 上下文中错误地写了 `tenant=0` 缓存(通过 force flag)
- **THEN** operlog 立即记录,运维可在事后排查
