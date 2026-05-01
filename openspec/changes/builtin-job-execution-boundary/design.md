## Context

当前 `service/cron` 启动时先同步内置任务投影，再由 `jobmgmt/internal/scheduler.LoadAndRegister` 扫描 `sys_job where status=enabled` 并注册所有任务。这个模型把 `sys_job` 同时作为内置任务的治理投影和执行来源，导致宿主/插件内置任务存在重复注册路径，也迫使调度器用幂等补丁处理同名 gcron entry。

新的边界是：内置任务由宿主代码或插件声明驱动执行，`sys_job.is_builtin=1` 只承载管理台展示、日志关联、审计快照和治理状态；用户自定义任务仍由 `sys_job.is_builtin=0` 持久化记录驱动执行。

## Goals / Non-Goals

**Goals:**

- 让内置任务执行来源回到代码/插件声明，避免从 `sys_job` 反向驱动内置任务注册。
- 让持久化调度器启动加载只处理用户自定义任务。
- 保留内置任务在管理台可见、可手动触发、可关联日志的治理能力。
- 保留插件禁用/卸载时的任务不可用状态投影，确保禁用插件后不再执行其任务。
- 删除或重定位仅为重复注册服务的调度器补丁和测试。
- 维持 SQL debug 场景下启动数据库操作可解释、可度量。

**Non-Goals:**

- 不允许管理台修改内置任务的 cron 表达式、超时、并发策略、启停状态或执行定义。
- 不引入新的数据库表或外部调度依赖。
- 不改变用户自定义任务的 CRUD、启停、手动触发和 Shell 任务权限语义。
- 不改变 `sys_job_log` 作为统一执行日志表的定位。

## Decisions

### Decision 1: 内置任务由代码/插件声明直接注册

`cron.syncBuiltinScheduledJobs` SHALL 在同步 `sys_job` 投影后，使用同步结果中的 `sys_job.id` 注册对应 gcron entry。调度 entry 的 cron、scope、concurrency、timeout、handler_ref 等执行定义 SHALL 来源于当前代码/插件声明转换后的内存模型，而不是再次扫描并信任 `sys_job` 中的内置记录。

备选方案是继续让 `LoadAndRegister` 扫描全部 `enabled` 任务，并通过 `gcron.Remove` 或同名覆盖保持幂等。该方案保留了错误边界：内置任务执行仍被数据表反向驱动，且启动期需要额外查询和重复注册保护，因此不采用。

### Decision 2: 持久化调度器只加载用户自定义任务

`jobmgmt/internal/scheduler.LoadAndRegister` SHALL 查询 `is_builtin=0 AND status=enabled`。CRUD、启停和编辑刷新路径继续服务用户自定义任务；内置任务的运行期注册由 cron 组件的内置任务同步路径负责。

这样可以让“持久化任务调度器”保持清晰职责：从用户持久化数据恢复用户任务；内置任务则属于宿主/插件运行时能力。

### Decision 3: 内置任务手动触发保留，但执行定义来自声明快照

手动触发仍通过 `sys_job.id` 入口，以便复用权限、确认弹窗、日志关联和审计。触发内置任务时，后端 MUST 使用当前注册的 handler 和内置任务投影来校验可执行性；若插件 handler 不可用或任务处于 `paused_by_plugin`，触发 MUST 返回 handler unavailable 语义。

实现上可以继续读取 `sys_job` 作为日志快照来源，但不得允许管理台写入内置任务执行定义。若需要更严格的声明一致性，触发前可按 `handler_ref` 从当前内置任务声明索引重建执行参数。

### Decision 4: 插件生命周期控制插件内置任务调度 entry

插件启用时，系统 SHALL 发现并注册插件声明的 cron handler 和调度 entry；插件禁用或卸载时，系统 SHALL 注销该插件所有 cron handler 和调度 entry，并将关联 `sys_job` 投影为 `paused_by_plugin` / `plugin_unavailable`。历史日志和任务投影保留，便于管理员审计。

### Decision 5: 管理台对内置任务保持只读定义和可触发操作

前端继续隐藏或禁用内置任务的编辑、删除、启停、重置等定义变更操作；“立即触发”对可执行的内置任务保持可用。对于 `paused_by_plugin` 或 handler 不可用的内置任务，前端不得展示可点击的立即触发入口。

## Risks / Trade-offs

- [Risk] 内置任务注册需要拿到 `sys_job.id` 才能写日志关联。→ Mitigation: 内置任务同步先 upsert 投影并返回或重查投影，再注册运行期 entry。
- [Risk] 插件生命周期和 cron 组件之间的注册顺序可能产生短暂不可用状态。→ Mitigation: 保持现有同步生命周期回调，在同一请求链路内完成 handler 注册、投影同步和调度刷新。
- [Risk] 手动触发内置任务若继续读取 `sys_job`，可能被误解为表驱动执行。→ Mitigation: 规范和代码注释明确 `sys_job` 仅提供治理投影和日志快照，执行可用性以 handler registry 和声明索引为准。
- [Risk] 删除 `gcron.Remove` 幂等补丁后，真实重复注册会重新暴露。→ Mitigation: 将重复注册测试改为验证内置任务不进入 `LoadAndRegister`，并保留 CRUD `Refresh` 的显式 remove-then-register 行为。

## Migration Plan

1. 调整启动加载过滤条件，使 `LoadAndRegister` 只加载用户自定义任务。
2. 扩展内置任务同步返回投影记录或注册输入，确保注册时拥有稳定 `sys_job.id`。
3. 在内置任务同步路径注册宿主和插件内置任务调度 entry。
4. 调整插件禁用/卸载路径，确保注销插件内置任务调度 entry 并投影不可用状态。
5. 更新前后端测试，验证内置任务不能编辑/启停但可以手动触发。
6. 验证启动日志，确保不再通过 `LoadAndRegister` 重复注册内置任务。

回滚策略：若内置任务声明注册路径出现问题，可临时恢复 `LoadAndRegister` 对内置任务的加载过滤前行为，但必须保留重复注册保护并重新评估启动查询数量。

## Open Questions

无。用户已确认内置任务允许手动触发，其余边界按本设计执行。
