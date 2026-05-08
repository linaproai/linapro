## Why

当前服务启动时会在短时间内输出大量 SQL 调试日志，也会执行多轮重复的插件注册表、发布快照、菜单和资源引用查询。最近一次审查中，`temp/lina-core.log` 在 `2026-05-07 17:40:55` 启动后 10 秒内记录到约 98 条 SQL，前 4 秒约 57 条属于启动或启动后立即触发链路，影响启动日志可读性，也增加了本地开发和演示场景的启动成本。

本变更将启动期 SQL 数量、启动日志噪音和启动同步链路作为独立治理对象，先从插件启动同步、运行时预热、内置任务同步和 SQL debug 配置入手，建立可量化、可回归的启动效率基线。

## What Changes

- 新增启动 SQL 效率治理能力，定义启动期 SQL 统计口径、预算、日志摘要和回归校验方式。
- 调整本地默认配置，使普通启动默认不输出每条 SQL 调试日志；需要排查 SQL 时仍可通过 `database.default.debug=true` 显式开启。
- 合并插件启动链路中的启动快照上下文，避免 `BootstrapAutoEnable`、插件 HTTP 路由注册、运行时前端包预热、动态运行时只读投影等阶段重复读取 `sys_plugin`、`sys_plugin_release`、`sys_menu` 和 `sys_plugin_resource_ref`。
- 优化插件清单同步的 no-op fast path：清单、发布快照、菜单、权限和资源引用均无变化时，不开启事务、不写库、不做写后回读。
- 优化启动期内置定时任务投影和调度注册：复用声明派生快照，避免在同一次启动中既按声明注册又重复从 `sys_job` 读取同一批内置任务。
- 为启动 SQL 统计和关键阶段耗时增加结构化摘要日志，保留可观测性而不是依赖 ORM SQL 明细刷屏。
- 增加后端启动 smoke 或单元测试，约束默认配置下不得输出 SQL 明细，并约束插件启动同步在 no-op 场景下的查询/写入次数。

## Capabilities

### New Capabilities

- `startup-sql-efficiency`: 定义宿主启动期 SQL 数量、启动日志噪音、插件启动同步快照复用、no-op 同步路径和启动效率回归测试要求。

### Modified Capabilities

- `plugin-startup-bootstrap`: 启动引导阶段必须复用同一轮插件启动快照，不得在后续插件接线和预热阶段重复构造等价快照。
- `plugin-manifest-lifecycle`: 插件清单同步在无差异时必须保持无副作用，不得开启空事务或执行写后回读。
- `cron-job-management`: 内置定时任务启动投影必须使用声明派生快照注册，避免重复持久化扫描同一批内置任务。

## Impact

- 受影响后端代码：
  - `apps/lina-core/internal/cmd/cmd_http.go`
  - `apps/lina-core/internal/cmd/cmd_http_runtime.go`
  - `apps/lina-core/internal/cmd/cmd_http_routes.go`
  - `apps/lina-core/internal/service/plugin/`
  - `apps/lina-core/internal/service/plugin/internal/catalog/`
  - `apps/lina-core/internal/service/plugin/internal/integration/`
  - `apps/lina-core/internal/service/plugin/internal/frontend/`
  - `apps/lina-core/internal/service/cron/`
  - `apps/lina-core/internal/service/jobmgmt/`
  - `apps/lina-core/manifest/config/config.yaml`
  - `apps/lina-core/manifest/config/config.template.yaml`

- 受影响测试：
  - 新增启动 SQL 统计/日志 smoke 测试。
  - 新增插件同步 no-op fast path 单元或集成测试。
  - 新增内置任务启动注册去重测试。

- i18n 影响：
  - 本变更不新增前端页面、菜单、按钮、表单、表格或用户可见业务文案；不修改 API DTO 文档源文本。预计不需要新增、修改或删除运行时 i18n、插件 manifest i18n 或 apidoc i18n 资源。

- 缓存一致性影响：
  - 本变更不新增业务缓存，不改变缓存权威数据源。启动快照仅限单次启动调用链内使用，不跨请求、不跨进程、不作为持久缓存。集群模式下仍以数据库和现有集群拓扑/修订机制为权威。
