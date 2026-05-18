## Context

当前启动 SQL 主要来自四类路径：

1. GoFrame 首次访问 DAO 表时执行表和列元数据探测，例如 `SHOW TABLES`、`SHOW FULL COLUMNS FROM ...`。这属于框架行为，短期内不应通过绕开 ORM 或手写 SQL 规避。
2. 插件启动同步在多个阶段读取相同治理表。`BootstrapAutoEnable` 会执行 `SyncSourcePlugins`，随后插件 HTTP 路由注册会刷新启用快照，运行时前端包预热又重新构造 catalog snapshot，动态运行时只读投影和页面请求也会重复读取 `sys_plugin` / `sys_plugin_release`。
3. 插件清单同步已有部分匹配判断，但菜单同步仍会进入事务，空事务在日志中表现为 `BEGIN` / `COMMIT`；写后还会回读 registry/release 刷新启动快照。
4. Cron 启动会同步内置任务并注册调度器，随后 monitor 插件内置任务可能马上执行一次，导致启动附近日志夹杂业务定时任务 SQL。

这次优化应优先减少项目自身可控的重复 SQL 和日志噪音，而不是和 GoFrame 的元数据探测机制对抗。项目是新项目，不需要为历史启动行为保留兼容路径，可以直接改启动编排和快照模型。

## Goals / Non-Goals

**Goals:**

- 默认启动日志不输出 ORM SQL 明细，普通开发和演示时日志聚焦关键启动阶段。
- 在插件、cron 和运行时预热无实际状态变更时，减少重复查询、空事务和写后回读。
- 建立启动 SQL 统计口径和自动化回归测试，避免后续迭代重新引入启动 SQL 膨胀。
- 保持插件生命周期、动态插件运行时、内置定时任务治理语义不变。
- 明确 i18n 和缓存一致性影响判断。

**Non-Goals:**

- 不修改 GoFrame ORM 元数据探测机制，不通过手写 SQL 替代 DAO 访问。
- 不移除必要的启动健康检查、插件授权校验、动态运行时收敛或 cron 注册。
- 不改变插件安装、启用、禁用、卸载的用户可见语义。
- 不改变集群模式下主从节点的生命周期副作用边界。
- 不优化启动后浏览器页面请求、定时任务周期执行或业务接口 SQL。

## Decisions

### 决策一：把 SQL 明细日志和真实 SQL 优化拆开处理

本地 `config.yaml` 当前设置 `database.default.debug=true`，导致 ORM 每条 SQL 都输出。模板文件默认是 `false`。本变更将交付型默认配置统一为 `false`，并保留注释说明如何临时开启。

理由：

- 启动日志刷屏首先是配置问题，关闭 debug 不影响真实行为。
- SQL 明细日志属于诊断工具，不适合作为默认开发体验。
- 真实 SQL 优化仍通过后续任务处理，避免把“看不到日志”误当作“减少执行”。

替代方案：保留 `debug=true`，仅过滤启动阶段 SQL 日志。该方案需要侵入 GoFrame 日志处理链，收益低且容易影响排查能力，因此不采用。

### 决策二：引入一次启动链路内共享的 `StartupContext`

在 HTTP 启动编排中创建一次启动上下文，包含：

- catalog startup snapshot：`sys_plugin`、`sys_plugin_release`
- integration startup snapshot：插件菜单和资源引用
- job startup snapshot：`sys_job_group`、内置 `sys_job`
- 可选的启动统计采集器：按阶段记录 SQL 预算、耗时和是否发生写入

`BootstrapAutoEnable`、插件 HTTP 路由注册、runtime frontend prewarm、cron builtin sync 等启动阶段复用该上下文。已有 `WithStartupDataSnapshot` 可以保留，但应避免每个阶段重复构造等价快照。

理由：

- 当前已有 snapshot 设计，问题在于生命周期太短、每个阶段重新构造。
- 将快照作用域提升到一次启动编排内，可以减少重复全表读取，同时保持快照不跨请求、不长期缓存。
- 单次启动上下文可以自然携带统计信息，便于最后输出摘要。

替代方案：把 catalog 和 integration snapshot 做成进程级缓存。该方案会引入跨请求一致性问题，需要复杂失效机制；启动同步只需要短生命周期快照，因此不采用。

### 决策三：插件同步 no-op 时不得进入事务或写后回读

插件 manifest 同步拆成两步：

1. 计算期望投影：registry、release snapshot、menu specs、permission menu specs、resource refs。
2. 与启动快照中的当前投影比较；只有存在差异时才进入事务、执行写入或回读。

菜单同步需要新增 `PluginMenusMatch` 或等价比较能力，避免无变化时仍开启 `dao.SysMenu.Transaction`。资源引用同步和 release metadata 已有字段匹配函数，应扩展到更早返回。

理由：

- 日志中多个空 `BEGIN` / `COMMIT` 来自无变化同步路径。
- 插件数量增长时，空事务会按插件数线性增长。
- 新项目可以直接把启动同步定义为“差异驱动”，无需兼容过去每次都跑事务的行为。

替代方案：保留事务但调低事务日志级别。该方案只降低可见噪音，未减少数据库操作，因此不采用。

### 决策四：写后优先更新启动快照，只有必要时回读数据库

对于 startup snapshot 内的 registry、release、menu、resource ref、builtin job 投影：

- 插入路径优先使用 `InsertAndGetId` 构造 entity 并写入 snapshot。
- 更新路径使用 `existing + data` 合成最新 entity 并写入 snapshot。
- 只有依赖数据库默认值、自动时间字段或复杂触发结果时才回读。

理由：

- 当前 `refreshStartupRegistry`、`refreshStartupRelease` 写后回读会制造额外查询。
- 启动同步使用的字段大多来自 manifest 或 DO projection，本地可确定。
- GoFrame 自动维护的 `created_at` / `updated_at` 对同一启动同步后续判断通常不是必需字段。

替代方案：继续写后回读确保实体完全等于数据库行。该方案最保守，但启动链路字段需求有限，不值得为完全等价付出每次额外查询。

### 决策五：内置定时任务注册只使用声明派生快照

Cron 启动顺序保留：先同步内置任务投影，再注册调度器。内置任务注册使用 `SyncBuiltinJobs` 返回的 projection 直接调用 `RegisterJobSnapshot`，持久化扫描只加载用户创建任务和非内置启用任务。

理由：

- 当前已有 `RegisterJobSnapshot` 和 `LoadAndRegisterSkipsBuiltinJobs` 的设计方向，应该强化这个边界。
- 内置任务执行定义权威来源是源码声明，`sys_job` 行只是治理投影。
- 避免启动期既 upsert 内置行又从 `sys_job` 读取同一批内置行。

替代方案：统一从 `sys_job` 加载所有任务。该方案简单，但违背内置任务“声明为权威”的现有规范，并增加重复扫描。

### 决策六：启动 SQL 统计用于回归门禁，不追求绝对零 SQL

新增测试不应断言精确 SQL 条数，而应断言：

- 默认配置不输出 SQL 明细。
- 插件同步 no-op 不写库、不产生空事务。
- 启动共享快照构造次数保持在预算内。
- 启动摘要日志包含阶段耗时和差异统计。

理由：

- GoFrame 版本、MySQL/SQLite 方言和测试环境会影响元数据探测条数。
- 精确 SQL 条数测试容易脆弱。
- 针对项目可控行为建立门禁更稳定。

## Risks / Trade-offs

- [Risk] 启动共享快照过期导致后续阶段读到旧状态。
  Mitigation：快照仅在同一启动编排内使用；所有写入路径必须同步更新快照。动态插件状态收敛和集群副作用仍以数据库为权威。

- [Risk] no-op 比较遗漏字段，导致应同步的菜单或资源未更新。
  Mitigation：比较函数必须覆盖持久化字段；测试覆盖 manifest 字段变更、菜单变更、route permission 变更、resource ref 变更四类差异。

- [Risk] 关闭默认 SQL debug 后排查问题不方便。
  Mitigation：配置注释保留开启方式；启动摘要日志保留关键阶段耗时、变更数量和错误上下文。

- [Risk] 统计测试在不同数据库或 GoFrame 版本下波动。
  Mitigation：测试约束项目可控行为，不断言框架元数据探测的绝对条数。

- [Risk] Cron 首轮 monitor 任务被误判为启动 SQL。
  Mitigation：统计口径区分 host startup phase 和 first scheduled job phase；启动摘要只统计宿主启动编排到路由绑定完成前的链路。

## Migration Plan

1. 先调整默认 SQL debug 配置与文档注释，降低日志噪音。
2. 引入启动共享上下文和统计摘要，不改变业务语义。
3. 将插件启动同步、运行时预热和 cron builtin sync 接入共享上下文。
4. 增加插件 no-op 比较和写后 snapshot 更新，移除空事务和不必要回读。
5. 补齐单元/集成/smoke 测试，记录优化前后数据。
6. 调用 `lina-review` 进行最终审查。

回滚策略：如发现启动状态同步异常，可先关闭共享启动上下文，恢复各阶段独立读取数据库；默认 SQL debug 配置可单独保留为 `false`。

## Open Questions

- 是否需要在开发模式下提供显式 `make dev sql_debug=1` 快捷开关，还是只依赖配置文件修改。
- 启动摘要日志是否需要输出到结构化 JSON 字段，还是先沿用当前文本日志格式。
- 动态 runtime reconciler 在单节点模式下当前不会启动；若后续改变单节点收敛策略，需要同步更新启动 SQL 统计口径。
