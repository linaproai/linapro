## Why

针对 `temp/codebase-review-report.md` 的源码审查结论，对照当前代码逐条核验后确认：宿主侧仍存在多个已被验证的正确性、安全、数据库性能与可观测性缺口。这些问题既不被 `backend-hardcoded-chinese-i18n-governance`、`plugin-api-query-performance`、`regression-feedback-localization-ui` 三个活跃迭代覆盖，也尚未沉淀到任何能力规范中。需要通过一次聚焦的"宿主基础加固"迭代将其统一治理，保证后续功能扩展和性能优化建立在稳定的宿主基线之上。

## What Changes

### 数据一致性与安全
- 用户删除（`internal/service/user/user.go` `Delete`）必须将用户软删除、组织关联清理、用户角色关联清理整体放入事务，吞错的 `Warningf` 改为返回错误并触发回滚；事务提交后再调用 `NotifyAccessTopologyChanged`。
- 角色删除（`internal/service/role/role.go` `Delete`）现有事务内角色菜单关联与用户角色关联删除失败仅 `Warningf` 后继续执行，必须改为 `return err` 让事务回滚，避免角色被删但关联记录残留。
- 角色用户分配（`internal/service/role/role.go` `AssignUsers`）必须改为整体事务 + 批量插入，禁止逐条插入并 `Warningf` 吞错。
- 菜单删除（`internal/service/menu/menu.go` Delete 流程）事务内角色菜单关联清理同样必须从 `Warningf` 改为返回错误。
- 上传文件路由（`internal/cmd/cmd_http.go` `bindUploadRoutes`）必须挂载到受保护路由分组下，由统一 Auth + Permission 中间件鉴权，禁止匿名访问已上传文件。

### 数据库结构与性能
项目无历史负担，按规范直接修改原有 SQL 文件并通过 `make init` 重新初始化数据库，所有改动保持幂等。
- `manifest/sql/008-menu-role-management.sql`：`sys_user_role` 增加 `KEY idx_role_id (role_id)`，`sys_role_menu` 增加 `KEY idx_menu_id (menu_id)`。
- `manifest/sql/001-project-init.sql`：`sys_user` 增加 `KEY idx_status (status)`、`KEY idx_phone (phone)`、`KEY idx_created_at (created_at)`。
- `sys_online_session` 所在 SQL 文件：增加 `KEY idx_last_active_time (last_active_time)` 支撑超时清理。
- `manifest/sql/002-dict-dept-post.sql`：`sys_dict_type` 与 `sys_dict_data` 增加 `deleted_at DATETIME DEFAULT NULL`，由 GoFrame 自动软删除接管，与其他业务表保持一致；执行 `make dao` 重新生成实体。
- `manifest/sql/014-scheduled-job-management.sql`：移除 `CONSTRAINT fk_sys_job_group_id`，改为 `KEY idx_group_id (group_id)`，由应用层维护一致性。

### 批量操作与前端性能
- 宿主新增 RESTful 批量删除接口 `DELETE /api/v1/user?ids=...` 与 `DELETE /api/v1/role?ids=...`，DTO 字段使用 `json` 标签和英文 `dc` / `eg`，`g.Meta` 携带模块对应权限标签；Service 层 `BatchDelete` 在单事务内复用现有保护策略（内置管理员、当前登录用户、角色被引用等）。
- 前端 `views/system/user/index.vue` 与 `views/system/role/index.vue` 的批量删除调用从 `for` 循环逐条删除改为单次批量 API。
- 菜单 `isDescendant`（`internal/service/menu/menu.go`）改为一次性加载父子映射后内存 BFS 判定，消除按层 SQL 往返。
- 服务器监控前端页面（`views/monitor/server/index.vue`）增加 30 秒自动刷新，使用 `useIntervalFn` + 页面可见性事件，标签页隐藏时暂停轮询。
- 用户消息轮询（`store/message.ts`）改造为可见性感知：隐藏时暂停 `setInterval`，可见时立即刷新一次。
- 路由守卫 `loadedPaths`（`router/guard.ts`）改为有界 LRU（默认 50 条），防止 SPA 长时间运行内存单调增长。
- 语言切换流程（`bootstrap.ts`）保留公共配置同步与字典缓存重置，但不再触发整套权限/菜单/路由重载；菜单标题须依赖响应式 `$t()` 自动更新。

### 宿主运行期可观测性与运维基础
- 新增公开 `GET /api/v1/health` 健康探针：执行一次轻量 DB 探活，正常返回 `{status:"ok"}`，DB 不可达返回 503，可被 K8s/反向代理直接消费。
- 在 HTTP 入口处增加 `SIGTERM` / `SIGINT` 信号处理，按顺序优雅关停 HTTP Server、Cron 调度器与数据库连接池；`shutdown.timeout` 默认 `30s` 可配。
- 删除空包 `apps/lina-core/pkg/auditi18n/` 与 `apps/lina-core/pkg/audittype/`，避免造成"已存在审计能力"的误导。审计日志能力的真实落地由独立迭代承接。
- `internal/service/cron/cron_managed_jobs.go` 的 `defaultManagedJobTimezone` 常量改为读取配置 `scheduler.defaultTimezone`，默认值改为 `UTC`，便于多区域部署。

### 不在本变更范围
- 后端中文硬编码清理由 `backend-hardcoded-chinese-i18n-governance` 承接。
- 英文环境页面文案/布局回归由 `regression-feedback-localization-ui` 承接。
- `GET /plugins` 副作用、`sys_online_session.last_active_time` 写入节流由 `plugin-api-query-performance` 承接。
- 审计日志真实实现、API 限流中间件、TraceID 中间件、Vue 全局错误边界、请求取消基础设施各自单独立迭代。
- `cmd_http.go` 内 `NewV1()` 自构造 service 引发的多实例与 DI 容器化重构按用户决定不并入本次，留作未来"宿主装配重构"迭代。

## Capabilities

### New Capabilities

- `host-runtime-operations`: 宿主运行期通用运维能力，覆盖健康探针、优雅关停、静态资源路由鉴权与运行期默认值的可配置化（含调度器默认时区与残留空包清理）。

### Modified Capabilities

- `user-management`: 删除事务化、批量删除接口与前端联动、`sys_user` 关键查询索引补齐。
- `role-management`: 删除事务内错误回滚、`AssignUsers` 整体事务化、批量删除接口与前端联动。
- `user-role-association`: `sys_user_role` 增加 `role_id` 反向索引以支撑按角色查询用户。
- `menu-management`: 菜单删除事务内错误回滚、`isDescendant` 改为内存判定、`sys_role_menu` 反向索引。
- `dict-management`: 字典类型与字典数据表补齐 `deleted_at`，与其他业务表统一软删除策略。
- `cron-job-management`: 默认时区可配置化、移除 `sys_job` 唯一外键约束并补反向索引。
- `server-monitor`: 监控页面前端轮询自动刷新与可见性感知。
- `user-message`: 用户消息轮询的页面可见性感知。
- `online-user`: `sys_online_session.last_active_time` 索引以支撑会话超时清理。
- `framework-i18n-runtime-performance`: 语言切换不再触发整套权限/菜单/路由重载。

字典管理（`dict-management`）现有规范已在 `字典数据表设计` 中声明 `deleted_at` 字段，本次仅需 SQL 落地以使实现与规范一致，不构成规范级行为变更，不在 Modified Capabilities 中列出。

## Impact

- **代码影响**：
  - 后端：`internal/service/{user,role,menu,cron}/`、`internal/cmd/cmd_http.go`、`internal/controller/{user,role}/`、`api/{user,role}/v1/`、`pkg/auditi18n`、`pkg/audittype`。
  - SQL：`manifest/sql/001-project-init.sql`、`002-dict-dept-post.sql`、`008-menu-role-management.sql`、`014-scheduled-job-management.sql`，以及 `sys_online_session` 所在 SQL 文件。
  - 配置：`apps/lina-core/manifest/config/config.yaml` 增加 `scheduler.defaultTimezone`、`shutdown.timeout`。
  - 前端：`views/system/user/index.vue`、`views/system/role/index.vue`、`views/monitor/server/index.vue`、`store/message.ts`、`router/guard.ts`、`bootstrap.ts`、`api/system/{user,role}/index.ts`。
- **运维影响**：补齐 `/health` 与优雅关停后，K8s/容器编排平台可以使用标准探针与回收流程；移除外键后高并发下任务调度锁开销下降。
- **测试影响**：新增/扩展 `apps/lina-core` 单元测试覆盖批量删除、删除事务回滚；扩展 `hack/tests/e2e/` 覆盖批量删除前端流程、健康探针匿名访问与监控页面自动刷新。
- **依赖影响**：无新增第三方依赖；`@vueuse/core` 已在仓库中可用。
- **不破坏 API 契约**：批量删除是新增接口；既有单条 `DELETE /api/v1/user/{id}` 与 `DELETE /api/v1/role/{id}` 保持不变。
