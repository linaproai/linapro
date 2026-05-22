## Why

用户在登录后进入后台首页时，`temp/login.log` 记录了大量 SQL。日志显示首屏加载期间有 10 个并发请求，共执行 152 条 SQL，其中 `sys_plugin_release`、`sys_plugin` 和 `sys_online_session` 的重复查询占主要比例。当前问题不是单条 SQL 慢，而是运行期重复读取和每请求鉴权状态校验放大了登录后首页的数据库压力。

## What Changes

- 新增登录后首页 SQL 执行效率规范，约束首屏运行期重复 SQL 的治理方式和验证口径。
- 优化在线会话校验流程，减少每个鉴权请求对 `sys_online_session` 的重复 `COUNT` 往返。
- 优化插件 catalog 运行期读取路径，减少同一请求或同一列表投影中对 `sys_plugin_release` 的重复按 ID 和按版本查询。
- 补充单元测试覆盖会话校验行为和插件 release 读取复用行为。
- 记录 i18n、缓存一致性、数据权限和开发工具跨平台影响评估。

## Capabilities

### New Capabilities

- `login-home-sql-efficiency`: 约束登录后首页运行期 SQL 数量、重复查询治理、会话校验往返、插件 catalog 读取复用和验证要求。

### Modified Capabilities

- 无。

## Impact

- 影响 `apps/lina-core/internal/service/session` 的在线会话校验实现。
- 影响 `apps/lina-core/internal/service/plugin/internal/catalog` 及相关插件列表/运行期读取路径。
- 不改变公开 HTTP API 契约、数据库 schema、前端页面结构或用户可见文案。
- 本变更不新增 i18n 文案；缓存改动必须明确单机与集群一致性边界。
