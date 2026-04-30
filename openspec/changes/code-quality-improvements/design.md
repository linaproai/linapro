## Context

`temp/codebase-review-report.md` 经逐条核验后形成的修复清单跨越后端、SQL、前端、运维四个维度。这些条目互相独立、影响面有限、但都已具备明确的"代码现状 + 期望行为"结论，适合通过一次专题迭代统一治理。

当前活跃迭代已分别认领：
- `backend-hardcoded-chinese-i18n-governance`：后端中文字面量与 i18n 治理；
- `plugin-api-query-performance`：插件列表查询副作用与会话活跃时间写入节流；
- `regression-feedback-localization-ui`：英文环境前端文案/布局/服务端默认间隔回归。

本变更必须避免与上述三者重叠，把焦点收敛到"宿主基础加固"——事务正确性、SQL 性能与一致性、批量接口与前端配套、宿主运行期可观测性与可运维性。

约束：
1. 项目无历史负担，所有 SQL 直接修改原文件，配合 `make init` 重新初始化数据库；
2. 所有 SQL 必须维持幂等；
3. 后端业务错误必须通过 `bizerr` 封装；
4. 后端时间长度必须使用 `time.Duration`，配置使用 `"30s"` 等带单位字符串；
5. 所有 API 文档源文本使用英文；
6. 文档变更使用中文（活跃迭代规则）。

## Goals / Non-Goals

**Goals:**
- 消除已识别的事务破坏点（用户/角色/菜单删除、`AssignUsers`），杜绝部分成功导致的孤儿数据。
- 关闭已识别的安全缺口：上传文件路由必须经过统一鉴权。
- 补齐已识别的 SQL 性能与一致性缺口：常用查询索引、字典软删除字段、调度任务外键改造。
- 用真正的批量接口替代前端"循环调用单条 API"的反模式。
- 引入 `/health` 与优雅关停以满足容器化部署的最小可运维性要求。
- 把硬编码的"上海时区"改为可配置；删除明显误导的空包。

**Non-Goals:**
- 审计日志能力的完整建模与数据持久化。
- API 限流、TraceID 中间件、请求取消、Vue 全局错误边界等横切设施。
- DI 容器化与 `cmd_http.go` 控制器装配重构。
- i18n 翻译键覆盖治理与中文硬编码清理。
- 字典管理规范层调整（规范已正确，仅需 SQL 实现对齐）。

## Decisions

### D1：删除事务统一收敛到 `dao.Xxx.Transaction`

把用户/角色/菜单删除及其关联清理操作放入一个事务闭包，事务内任何失败均 `return err` 触发回滚。`NotifyAccessTopologyChanged` 等通知放在事务提交后。

**为什么不只加重试**：重试无法保证原子性，且关联删除失败往往意味着外部状态异常，直接回滚比"反复 warning"更安全。

**替代方案**：补偿事务 / Saga。该方案在本项目规模下显著过度设计；`GoFrame` 的事务闭包已能覆盖。

### D2：`AssignUsers` 改为单事务 + 批量插入

收集需要新增的关联，构造 `[]do.SysUserRole` 后一次 `Insert(slice)`。失败整体回滚。

**为什么不分批 chunk**：分配一次写入的关联条数受 UI 选择限制，在合理量级（< 1000）下单次插入开销与多次小批次相当但事务更简单。

### D3：上传文件访问接口内聚到文件模块

`GET /api/v1/uploads/*` 由 `api/file/v1` 的标准 DTO 与 `internal/controller/file` 控制器维护，并随文件控制器一起挂载到受保护静态 API 分组。权限标签沿用文件下载权限 `system:file:download`，由统一 `Auth` 与 `Permission` 中间件处理。控制器不得直接拼接本地上传目录或调用 `ServeFile`，而是把 URL 中的相对存储路径交给文件 service 解析文件元数据，再通过当前配置的 storage backend 读取文件流，保证本地存储、分布式部署和对象存储接入时逻辑一致。

**为什么不改为按需 token 链接**：访问已上传文件属于业务行为，需要纳入统一鉴权与审计；做"带签名 URL"模式与本次范围不符。匿名访问保留口径仅在公开下载场景显式开启，目前没有此需求。

### D4：SQL 索引与外键改造

| 表 | 改动 | 理由 |
|---|---|---|
| `sys_user_role` | `KEY idx_role_id (role_id)` | "按角色查询用户"、"按角色删除关联"为常用路径 |
| `sys_role_menu` | `KEY idx_menu_id (menu_id)` | "按菜单删除角色关联"为级联删除路径 |
| `sys_user` | `KEY idx_status / idx_phone / idx_created_at` | 状态筛选、手机号搜索、时间范围筛选 |
| `sys_online_session` | `KEY idx_last_active_time (last_active_time)` | 超时会话清理 |
| `sys_dict_type / sys_dict_data` | `deleted_at DATETIME DEFAULT NULL` | 与其他业务表统一软删除策略，匹配现有 `dict-management` 规范 |
| `sys_job` | 移除 `fk_sys_job_group_id`，新增 `KEY idx_group_id` | 与仓库其他关联表保持一致；高并发任务调度下减少外键锁开销 |

**为什么用普通 `KEY` 而非 `INDEX`**：项目现有 SQL 风格统一使用 `KEY`，跟随仓库惯例。

**为什么字典表用软删除而非保持硬删除**：
1. `dict-management` 规范已在表结构中声明 `deleted_at`；
2. 字典类型/数据被业务广泛引用，硬删除可能导致历史日志/操作日志记录的字典 value 失去标签解释能力，软删除允许后续审计回溯。

### D5：菜单 `isDescendant` 内存判定

一次性 `dao.SysMenu.Ctx(ctx).Scan(&all)` 拉取全部菜单（菜单数据量小），构造 `parentChildren := map[int][]int`，做 BFS 判定 `targetId` 是否在 `parentId` 子树中。复杂度从 `O(depth × QPS)` 降至 `O(N)` 一次。

**为什么不引入显式的 path 字段**：方案需要写入侧维护，会扩散到所有创建/移动菜单接口，超出本次范围。当前数据规模不足以让"加 path 列 + 维护"比"内存判定"更划算。

### D6：批量删除 RESTful 设计

```
DELETE /api/v1/user?ids=1&ids=2&ids=3
DELETE /api/v1/role?ids=1&ids=2&ids=3
```

- 使用 `DELETE` 方法 + repeated Query 参数承载 `ids`（仓库 `cron-job-management` 的批量取消等接口已有先例）。
- DTO `BatchDeleteReq` 使用 `Ids []int json:"ids" v:"required|min-length:1" dc:"..." eg:"1"`。
- Service 层实现 `BatchDelete(ctx, ids)`：在单事务内复用现有单条 Delete 的全部保护策略（内置管理员、当前登录用户、角色被引用提示等）。任意一条触发 `bizerr` 即整体回滚。
- 前端 `userBatchDelete` / `roleBatchDelete` API 一次发起，UI 错误以 bizerr 统一格式展示。

**为什么不复用单条 API 做后端转发**：仍然是 N 次事务，且无法整体回滚。

**为什么 Query 参数而非 Body**：与项目既有"GET 列表用 Query"风格一致，DELETE 语义无需 body；浏览器/中间件均良好支持。

### D7：服务器监控前端自动刷新与可见性感知

使用 `@vueuse/core` 的 `useIntervalFn` + `useDocumentVisibility`：visible 时启动 30s 轮询，hidden 时 `pause`。组件 unmount 时显式 `stop`。

`store/message.ts` 的轮询同样改造为可见性感知：hidden 时 `pause`，visible 时立即触发一次 `fetchUnreadCount`。

**为什么不在路由切换时清理**：消息轮询是全局行为，与路由无关；改用可见性是行业惯例。

### D8：`loadedPaths` 改为有界 LRU

将 `Set<string>` 替换为简单 LRU（`Map` + 大小阈值，命中时 `delete + set` 重新插入）。阈值 50 条；超出时弹出最旧。

**为什么不用 `WeakMap`**：键是字符串路径，不是对象，`WeakMap` 不适用。

### D9：语言切换不触发完整权限重载

在 `bootstrap.ts` 的 `watch(preferences.app.locale, ...)` 中：
- 保留 `syncPublicFrontendSettings(locale)`（公共配置依赖语言）；
- 保留 `useDictStore().resetCache()`（字典缓存按语言键存）；
- **移除** `refreshAccessibleState(router)`；
- 后端菜单路由响应新增 `meta.i18nKey`，前端菜单模型保留该键，语言切换时通过运行时语言包本地重绘菜单标题，不请求 `/api/v1/user/info` 或 `/api/v1/menus/all`；
- 增加防御性扫描：菜单/路由 `meta.title` 必须使用 `$t(...)` 或 i18n key，不得使用启动时一次性求值的字符串。如发现 hardcode，将 title 改为响应式访问。

### D10：宿主运行期运维能力（新能力 `host-runtime-operations`）

#### `/api/v1/health` 健康探针
- 公开路由（无鉴权）。
- 实现：执行 `dao.SysUser.Ctx(ctx).Limit(1).Count()` 作为 DB 探活；超过 `health.timeout`（默认 `5s`）视为不可用。
- 响应：`200 {status:"ok"}` 或 `503 {status:"unavailable", reason:"database probe failed"}`，内部错误仅写入日志，不直接暴露给匿名调用方。

#### 优雅关停
- HTTP 入口使用 GoFrame `Server.Run()`，复用框架内置的 `SIGTERM` / `SIGINT` 等 shutdown 信号监听与 HTTP graceful shutdown。
- `cmd_http.go` 不再自行注册 `os/signal`，也不直接重复调用 HTTP Server `Shutdown()`。
- `Server.Run()` 返回后按顺序清理宿主自有运行期资源：
  1. 触发 `cronSvc.Stop(ctx)`；
  2. 停止集群服务；
  3. 关闭数据库连接池；
- 运行期资源清理受 `shutdown.timeout`（默认 `30s`）约束，超时返回错误并记录 warning。

#### 上传路由保护
见 D3。

#### 残留空包清理
删除 `apps/lina-core/pkg/auditi18n/` 与 `apps/lina-core/pkg/audittype/`。grep 当前调用确认无依赖，否则在拆除阶段补救。

#### 调度器默认时区可配置
`cron_managed_jobs.go` 的 `defaultManagedJobTimezone` 不再以常量形式硬编码，改为通过配置服务读取 `scheduler.defaultTimezone`，默认值为 `UTC`。`config.template.yaml` 增 `scheduler.defaultTimezone: "UTC"` 并在 README 中说明可改。

### D11：宿主基础服务接口职责拆分

`config.Service` 保持作为完整配置服务的组合接口，但不再直接平铺全部方法；按集群、鉴权、登录、前端、i18n、cron、宿主运行期、交付元数据、插件、上传与运行时参数同步拆分为小接口后由 `Service` embed。这样调用方可在后续重构和测试中依赖更窄的 reader/syncer。

`middleware.Service` 同样保持调用方兼容，但拆分为：
- `HTTPMiddleware`：只包含可直接安装到 GoFrame 路由组的 HTTP 中间件方法和中间件工厂；
- `RuntimeSupport`：只包含非中间件支撑方法，例如 `SessionStore()` 与 `PublishedRouteMiddlewares()`。

这些拆分不改变运行时行为，也不新增用户可见文案或 API 契约；i18n 资源无需变更。

## Risks / Trade-offs

- [SQL 重新初始化破坏开发数据库现状] → 在 tasks.md 中明确执行步骤为"修改 SQL → `make init`"；review 时检查没有遗留迁移路径假设。
- [上传路由加鉴权后第三方页面引用图片失效] → 当前页面/插件均通过登录态访问，未鉴权场景没有产品诉求；review 阶段抽查所有 `<img src="/api/v1/uploads/...">` 用法是否在登录态下加载。
- [删除事务变严格后报错率上升] → 这是预期行为，错误本来就被吞掉；E2E 增加"关联删除失败时整体回滚"用例验证 UI 行为。
- [菜单 `isDescendant` 改为一次性加载] → 菜单总数远小于 1k，单次内存量级可忽略；如未来超大规模再换 path 列方案。
- [移除 `sys_job` 外键改为应用层维护] → 调度任务在写入侧已有 `group_id` 校验路径；review 时确认所有写入路径都经过 service 层。
- [优雅关停超时处理] → HTTP graceful shutdown 由 GoFrame `Server.Run()` 负责；宿主自有运行期资源清理由 `shutdown.timeout`（默认 `30s`）约束，超时返回错误并打印 warning。
- [`/health` 添加 DB 探活带来基线 QPS] → 单条 `Limit(1).Count()` 极轻量，可忽略；K8s 探针周期通常 ≥ 10s。

## Migration Plan

1. 合并前在开发环境执行：
   - `make init` 重建数据库（验证幂等且包含新增索引、`deleted_at` 字段）。
   - `make dao` 重新生成实体（验证 dict 实体包含 `DeletedAt`）。
   - `make ctrl` 重新生成控制器骨架（用户、角色批量删除接口）。
2. 部署侧：因外键被移除，旧库需要执行 `ALTER TABLE sys_job DROP FOREIGN KEY fk_sys_job_group_id;`——**因项目无历史包袱，统一通过 `make init` 重新初始化即可**，无需补充迁移脚本。
3. 容器化部署需更新 deployment：
   - `livenessProbe` / `readinessProbe` 指向 `/api/v1/health`；
   - `terminationGracePeriodSeconds ≥ shutdown.timeout`（默认 30）。
4. 监控 cron 任务时区：在配置中显式声明 `scheduler.defaultTimezone`，否则按 UTC。

## Open Questions

- 上传路由的最终权限标签使用 `file:read` 还是更细粒度的 `file:download`？apply 阶段需对照 `file` 模块的菜单/按钮权限既有定义决定，避免与现有权限码冲突。
- `/health` 是否在 `cluster.enabled=true` 时需要返回主从角色信息？本次先返回 `{status, mode}` 中的 `mode`（`master` / `slave`），但不做主从切换探针，避免与 `leader-election` 能力交叉。
