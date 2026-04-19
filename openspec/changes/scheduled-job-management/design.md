## Context

### 现状

- `apps/lina-core/internal/service/cron` 已经存在,承载宿主侧代码内置任务:`session-cleanup`、`servermon-collector`、`servermon-cleanup`、`access-topology-sync`、`runtime-param-sync`,以及 `pluginSvc.RegisterCrons` 注入的插件级 cron。
- 底层调度器是 GoFrame 的 `gcron`;选主能力来自 `service/cluster`;分布式锁能力来自 `service/locker` 与 `service/hostlock`。
- 所有内置任务在进程启动时由代码直接注册,缺少任何用户可编辑入口与持久化。

### 需求概述

新增面向运维用户的"定时任务管理"子系统,覆盖 CRUD、分组、启停、执行日志、次数/并发策略、手动触发与终止,并支持:
- **Handler 模式**: 调用宿主或插件注册的具名 handler;参数经 JSON Schema 校验。
- **Shell 模式**: 执行 `/bin/sh -c` 多行脚本,带独立权限与全局开关。

### 约束

- 必须复用现有 `service/cron` 与 `service/cluster`,不能新起第二套调度器。
- 后端遵循 `GoFrame v2` 约定;DAO/DO/Entity 由 `gf gen dao` 自动生成。
- 前端严格遵循 `Vben5 + Ant Design Vue`,且必须与 `ruoyi-plus-vben5` 的交互风格保持一致。
- 枚举值走字典模块,不在代码中硬编码字面量。
- 数据库操作使用跨库通用 SQL,避免 MySQL 专有函数。
- 时间长度统一 `time.Duration`;配置文件用 `"10s"/"5m"/"1h"` 字符串。

## Goals / Non-Goals

**Goals:**

1. 提供完整的用户可管理定时任务体系(CRUD、分组、启停、日志、清理策略)。
2. 同时支持 handler 模式与 shell 模式,两种任务类型共享同一张任务表。
3. 在现有 `service/cron` 之上复用 gcron,通过动态注册支持运行时 CRUD。
4. 插件 handler 在插件启/停事件中自动上/下线,关联任务自动级联状态而不丢数据。
5. Shell 模式实现多层安全防护:独立权限点 + 全局开关 + 审计日志 + 进程组 kill + 输出截断。
6. 集群模式下正确区分 `master-only` 与 `all-node` 调度范围,沿用 `cluster.IsPrimary()` 守卫。
7. 内置任务通过 seed SQL 注入,允许运维修改 `cron_expr / timezone / status / max_executions / log_retention_override`,其他字段锁死且 seed 不覆盖用户修改。
8. UI 遵循 Vben5 组件规范,参数区依据 handler 的 JSON Schema 动态渲染。

**Non-Goals:**

1. 不做"全局单例"第三种调度模式(本迭代仅 master-only / all-node)。
2. 不做排队模式(parallel 超过 `max_concurrency` 时跳过而非排队)。
3. 不做 Windows 下的 shell 执行(生产环境为 Linux;开发环境 Windows 默认禁用 shell 模式,仅提示)。
4. 不做 cron 表达式可视化编辑器(本迭代仅做文本输入 + 下一次执行时间预览)。
5. 不提供 handler 参数的前端自定义组件库,只从 JSON Schema 映射到 Vben 内置控件(输入框、数字、开关、下拉、日期等)。
6. 不在本迭代实现"任务依赖 / DAG"语义。
7. 不做 handler 脚本沙箱(Wasm/JS)——插件 handler 沿用已有 Wasm 插件体系。
8. 不支持"一次性任务"(即给定绝对时间点执行一次);如有需要通过 `max_executions=1` + cron 近似实现。

## Decisions

### D1. 任务类型:单表 + `task_type` 字段,而非两张表

**决定**: `sys_job` 一张表,用 `task_type ∈ {handler, shell}` 区分,类型相关字段按需填写(handler 用 `handler_ref` + `params`,shell 用 `shell_cmd` + `work_dir` + `env`)。

**理由**: 两种类型共享大量字段(分组、名称、cron、scope、concurrency、executed_count、status、is_builtin、log_retention),分两张表会导致查询/列表场景反复联表;单表设计使调度器统一处理。缺点是部分列对某一类型恒为空,属于可接受的稀疏性。

**备选方案**: 主表只存公共字段,`sys_job_handler` / `sys_job_shell` 存差异字段——被拒,查询复杂度远高于收益。

### D2. 调度器复用 gcron,持久化状态从数据库加载

**决定**:
- 启动时 `service/cron/cron.go` 新增一次性操作:`loadPersistentJobs(ctx)` 从 `sys_job where status=enabled` 读取并注册到 gcron。
- 任务 CRUD 在 controller/service 层统一经过 `JobScheduler` 组件,它负责 `gcron.Remove + gcron.Add` 的原子刷新。
- 调度回调是一个薄 wrapper `runJob(jobID)`,每次 tick 进入 wrapper 后从数据库读取最新任务配置(命中小容量 LRU 缓存,避免高频 DB 压力),再决定分发到 handler / shell 执行器。

**理由**: 不自研调度器;保留 gcron 的成熟度。wrapper 层从 DB 读配置解决了"CRUD 修改参数后调度仍使用旧快照"的问题;缓存失效时机为 CRUD 完成时主动 invalidate。

**备选方案**:
- 使用全内存注册表,CRUD 时直接改内存——被拒,进程重启后容易与 DB 漂移,且多节点间无共识。
- 每次 tick 直接读 DB 无缓存——被拒,高频任务可能压 DB,缓存 + 主动失效更合理。

### D3. 集群调度范围:沿用已有两档,不引入全局单例

**决定**: `scope ∈ {master_only, all_node}`。
- `master_only`: 所有节点都注册 gcron,触发时先检查 `cluster.IsPrimary()`,非主节点直接返回。
- `all_node`: 每节点各自注册、各自执行。

**理由**: 已有 `service/cron` 就是这两档实现,迁移成本最低;全局单例对用户而言理解成本更高,且典型用户任务(清理、同步、健康探测)master-only 已经够用。

**风险**: Primary 切换瞬间的 tick 可能被跳过(新主还未认领完成选主);日志会记录 `status=skipped_not_primary`,运维可感知。

### D4. 并发策略:单例 / 并行 + max_concurrency

**决定**: `concurrency ∈ {singleton, parallel}`。
- `singleton`: 本节点单例(`gcron.SetSingleton(true)` 或等价语义),已在跑时跳过新 tick,日志 `status=skipped_singleton`。
- `parallel`: 本节点按 `max_concurrency` 软上限;超限时跳过新 tick,日志 `status=skipped_max_concurrency`。

**理由**: 简单可观察。"全局单例"需要走分布式锁,增加复杂度,本迭代 Non-Goal。

**备选方案**: 超限时排队——被拒,队列会产生次生复杂度(持久化、重启丢失、饥饿)。

### D5. Handler 注册:宿主内置 + 插件订阅,统一走 `HandlerRegistry`

**决定**: 新增 `service/jobhandler` 组件,定义 `Registry` 接口:

```go
type Registry interface {
    Register(ref string, def HandlerDef) error  // ref 形如 host:xxx 或 plugin:<pluginID>/<name>
    Unregister(ref string)
    Lookup(ref string) (HandlerDef, bool)
    List(ctx context.Context) []HandlerInfo  // UI 下拉用
}

type HandlerDef struct {
    Ref          string                     // 唯一标识
    DisplayName  string                     // UI 展示名
    Description  string                     // 介绍
    ParamsSchema string                     // JSON Schema 文本
    Source       HandlerSource              // host | plugin
    PluginID     string                     // 插件来源时填
    Invoke       func(ctx context.Context, params json.RawMessage) (result any, err error)
}
```

- 宿主启动阶段调用 `Registry.Register("host:clean-session-logs", ...)` 注册内置 handler。
- `service/plugin` 在插件启用事件中遍历清单并调 `Registry.Register`,禁用/卸载时调 `Unregister`。
- `Registry.Unregister(ref)` 会触发 `JobScheduler` 级联:所有 `handler_ref=<ref>` 且 `status=enabled` 的任务置为 `paused_by_plugin`,UI 标红;重新 `Register` 时自动恢复为 `enabled`(仅恢复那些 `stop_reason=plugin_unavailable` 的任务)。

**理由**: Registry 是唯一的真理来源,任务调度、UI 下拉、参数校验都走它;级联逻辑集中,避免散落在各调用方。

### D6. Shell 安全分层

**决定**: Shell 任务引入三层防护:

```
L1 权限        system:job:shell 权限点,默认仅 admin 拥有
L2 全局开关    cron.shell.enabled = true/false,运维随时接管
L3 审计        shell 任务的创建/修改/手动触发全部写 oper_log
```

**执行时**:
- 固定 `/bin/sh -c <shell_cmd>` 启动 `exec.CommandContext`。
- `SysProcAttr.Setpgid = true`(Unix),超时/手动终止时 `kill -- -pgid`。
- `work_dir` 为空时使用宿主进程工作目录;非空时先校验目录存在 + 非根路径。
- `env` 为任务级 KV map,叠加在宿主进程环境之上(任务级覆盖进程级)。
- `stdout` / `stderr` 各自 `io.LimitReader` 截留前 64KB,超出追加 `...[truncated]` 标记。
- `timeout` 为必填字段(秒),范围 `[1, 86400]`,`ctx, cancel := context.WithTimeout(parentCtx, timeout)`。

**Windows 下的处理**: 构建期不禁编,运行期如 `runtime.GOOS=="windows"` 则 `cron.shell.enabled` 强制视为 false,UI 提示"当前平台不支持 shell 模式"。

### D7. 系统内置任务的"部分只读"

**决定**: 锁定 `is_builtin=1` 任务的 `task_type / handler_ref / params / scope / concurrency / group_id / name`;开放 `cron_expr / timezone / status / max_executions / log_retention_override` 可改。

**seed 策略**: seed SQL 使用 `INSERT ... ON DUPLICATE KEY UPDATE` 的变体——只有 `last_seeded_at` 小于代码内置版本号时才更新锁定字段;开放字段只在首次插入时设定默认值,之后 seed 不覆盖。通过 `sys_job.seed_version` 列记录每次 seed 版本。

**理由**: RuoYi 风格的"seed 完全覆盖"常导致升级时吞掉用户改过的 cron 表达式;版本号 gate 保证升级只推新内置任务与强制字段。

### D8. 日志清理:全局默认 + 任务级覆盖 + 系统内置清理任务

**决定**:
- 系统参数新增 `cron.log.retention`(默认 `{mode: days, value: 30}`)。
- 任务级 `log_retention_override`(JSON,`{mode: days|count|none, value: N}`),null 表示跟随全局。
- 新增系统内置任务 `host:cleanup-job-logs`,每日凌晨扫描 `sys_job_log` 按策略清理。

**理由**: 清理本身是定时任务,用本子系统自举更干净;`mode=none` 表示不清理(用户愿意承担存储成本的场景)。

### D9. 手动触发与手动终止的语义

**决定**:
- **手动触发**: UI "立即执行"按钮 → POST `/job/{id}/trigger` → scheduler 启动一次新执行,`trigger=manual`,**不计入 `executed_count`**。
- **手动终止**: 正在 running 的日志行上 → POST `/job/log/{id}/cancel` → scheduler 查找对应 goroutine,调 `cancel()`。Handler 必须响应 `ctx.Done()`;Shell 执行 kill 进程组。
- 终止后日志 status = `cancelled`,附上 `end_at`。

### D10. 手动终止与 handler 规范的契约

**决定**: 明文约束所有 handler(包括插件)实现必须满足:
1. 接受 `ctx context.Context` 作为首参;长阻塞操作必须把 `ctx` 透传到下游(HTTP / DB / 锁等)。
2. 周期性检查 `ctx.Done()`,收到取消信号后尽快退出并返回 `ctx.Err()`。
3. 清理逻辑使用 `defer`,确保取消路径也能跑清理。

违反上述契约的 handler 在"手动终止"时会表现为"日志已标记 cancelled 但 goroutine 继续占用资源",运维可观察后排查对应 handler 代码。

### D11. RESTful 接口设计

**决定**:

```
GET    /job                 列表(支持分组/状态/关键字/类型过滤)
GET    /job/{id}            详情
POST   /job                 创建
PUT    /job/{id}            更新
DELETE /job                 批量删除
POST   /job/{id}/trigger    手动触发一次
POST   /job/log/{id}/cancel 手动终止运行中实例
PUT    /job/{id}/status     启用/禁用
POST   /job/{id}/reset      重置 executed_count

GET    /job-group           列表
POST   /job-group           创建
PUT    /job-group/{id}      更新
DELETE /job-group           批量删除(默认分组拒绝)

GET    /job/log             列表(按任务/时间/状态过滤)
GET    /job/log/{id}        详情(含 result_json 全文)
DELETE /job/log             清空(按任务或全部)

GET    /job/handler         handler 注册表列表(下拉用)
GET    /job/handler/{ref}   handler 详情(含 ParamsSchema)
```

### D12. 数据模型关键字段

```
sys_job_group
  id, code, name, remark, sort_order, is_default (0|1)
  created_at, updated_at, deleted_at

sys_job
  id, group_id, name, description
  task_type (handler|shell)
  handler_ref (nullable, shell 任务为 null)
  params (json, handler 任务用)
  shell_cmd (nullable), work_dir (nullable), env (json, nullable), timeout_seconds
  cron_expr, timezone
  scope (master_only|all_node)
  concurrency (singleton|parallel)
  max_concurrency (parallel 时 >=1,singleton 恒为 1)
  max_executions (>=0, 0=无限)
  executed_count (运行时累加)
  stop_reason (nullable: max_executions_reached | plugin_unavailable | manual | ...)
  log_retention_override (json, nullable)
  status (enabled|disabled|paused_by_plugin)
  is_builtin (0|1), seed_version (int)
  created_by, created_at, updated_by, updated_at, deleted_at
  唯一约束: (group_id, name)  — 组内唯一

sys_job_log
  id, job_id, job_snapshot (json, 执行时的任务配置快照)
  node_id
  trigger (cron|manual)
  params_snapshot (json,handler 类型时的参数)
  start_at, end_at, duration_ms
  status (running|success|failed|cancelled|timeout|skipped_not_primary|skipped_singleton|skipped_max_concurrency)
  err_msg (nullable)
  result_json (json, nullable; shell 任务含 stdout/stderr/exit_code)
  created_at
  索引: (job_id, start_at DESC), (status), (start_at DESC)
```

## Risks / Trade-offs

- **gcron 动态注册的竞态**: CRUD 与调度 tick 并发时,`Remove + Add` 中间窗口若刚好命中 tick,可能多触发或漏触发一次。→ 在 `JobScheduler` 对 `Remove/Add` 加单任务互斥锁(`sync.Map` per-job mutex),且在 add 前清空 LRU 缓存。
- **Primary 切换瞬间 master-only 任务的 tick 丢失**: 新主认领前 tick 可能被跳过。→ 日志明确记录 `status=skipped_not_primary`,运维可见;需要强保证的任务用 handler 自己做幂等(外部状态记录)。
- **Handler goroutine 泄漏**: Handler 未响应 `ctx.Done()` 时,"终止"只是标记日志,goroutine 仍在。→ Spec 写清契约,review 时把关;未来可考虑超时后无法终止则 panic 对应 goroutine(本迭代不做)。
- **Shell 模式滥用风险**: 即便有权限点和开关,管理员仍可执行任意命令。→ 全局开关默认关闭,需运维主动开;权限点默认只给 admin;所有 shell 操作进审计日志;shell_cmd 与 work_dir 字段展示告警色。
- **插件禁用引发大量任务级联**: 插件下有 100 个任务时,一次禁用会产生 100 条 `paused_by_plugin` 记录和日志。→ `Unregister` 级联使用批量 UPDATE + 单条审计日志记录影响范围。
- **Seed 升级与用户改动的冲突**: 新版本内置任务改了 scope,旧任务已存在但 scope 是锁定字段。→ `seed_version` 低时强制覆盖锁定字段,开放字段永不覆盖;运维可通过"重置为默认"显式恢复。
- **result_json 过大导致行过宽**: shell 输出 64KB * 两路,极端情况一行 128KB。→ MySQL 使用 `longtext`;前端查询列表时不带 `result_json`,详情接口再返回。
- **任务在删除分组时的迁移**: 删除分组会把任务迁到默认分组 → 需要明确不可逆;UI 弹确认框并说明影响范围。
- **时区漂移**: 多节点服务器 TZ 不一致时,不带 timezone 的任务触发点不一致。→ `timezone` 必填,seed 时默认 `Asia/Shanghai`,任务表单默认取服务器 TZ。

## Migration Plan

### 部署步骤

1. 合并 PR 后执行 `make init` 触发 `manifest/sql/<NNN>-scheduled-job-management.sql` 迁移:建三张新表、插入默认分组、插入内置任务 seed、插入 `cron.shell.enabled=false` 与 `cron.log.retention` 两个系统参数、注册权限点与菜单项。
2. 前端构建部署,灰度开启 `system:job:*` 菜单项。
3. 后端重启时,`service/cron/cron.go` 的 `Start(ctx)` 会调用 `jobmgmt.Scheduler.LoadAndRegister(ctx)`,将 `sys_job where status=enabled` 全量注册到 gcron;已有代码内置 cron 继续由旧路径注册,互不影响。
4. 运维在系统参数页开启 `cron.shell.enabled=true`(若需 shell 能力),并为需要执行 shell 的角色授予 `system:job:shell` 权限。

### 回滚策略

- 后端回滚:还原代码版本;数据表保留(三张新表不会被旧版本引用),重启后旧版本不再加载任务。
- 前端回滚:隐藏菜单入口。
- 数据回滚:本次变更只新增表与 seed,无破坏性 DDL;无需数据回滚。
- 紧急停用:在系统参数页设 `cron.shell.enabled=false`,shell 任务立即被拦截。

## Open Questions

1. Shell 任务的 `env` 是否需要支持加密存储(数据库密码等敏感值)?本迭代暂按明文存储,在 UI 表单上用 password 控件遮罩显示。→ 待用户反馈决定是否引入配置加密组件。
2. Handler 执行超时字段要不要做成 handler 级别的默认值(由 Registry 提供),任务可以覆盖?本迭代任务表自带 `timeout_seconds`,handler 可以不声明默认。
3. `max_concurrency` 在 `all_node` 模式下是"每节点上限"还是"全局上限"?本迭代按"每节点上限"实现(简单一致),如有全局需求再迭代。
4. 是否需要暴露 "下次执行时间" 的预览(前端 cron 表达式实时解析)?本迭代在表单下方显示"最近 5 次下次执行时刻",基于后端接口 `GET /job/cron-preview?expr=...&tz=...`。
