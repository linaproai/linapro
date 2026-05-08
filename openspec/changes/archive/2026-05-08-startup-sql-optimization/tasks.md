## 1. 基线与配置

- [x] 1.1 记录当前 MySQL 默认启动基线：采集从后端进程启动到路由绑定完成的 SQL 明细数量、启动耗时、插件数量、内置任务数量，并保存到变更说明或任务备注中
- [x] 1.2 将 `apps/lina-core/manifest/config/config.yaml` 的 `database.default.debug` 默认值调整为 `false`，保持 `config.template.yaml` 为 `false`，并补充注释说明需要排查 SQL 时显式开启
- [x] 1.3 增加默认启动日志 smoke，断言默认配置下启动日志不包含 `SHOW FULL COLUMNS`、`SELECT ... FROM`、`INSERT INTO` 等 ORM SQL 明细
- [x] 1.4 增加显式 debug 配置测试或手工验证记录，确认 `database.default.debug=true` 时仍可输出 SQL 明细

## 2. 启动共享上下文

- [x] 2.1 设计并实现一次 HTTP 启动编排内共享的 `StartupContext` 或等价结构，承载 catalog、integration、job 启动快照和统计采集器
- [x] 2.2 调整 `startHTTPRuntime` / `bindSourcePluginHTTPRoutes` 等启动编排函数签名，确保 `BootstrapAutoEnable`、插件路由注册、runtime frontend prewarm、cron startup 能复用同一启动上下文
- [x] 2.3 将 `catalog.WithStartupDataSnapshot` 接入共享上下文，避免同一启动链路重复读取 `sys_plugin` 与 `sys_plugin_release`
- [x] 2.4 将 `integration.WithStartupDataSnapshot` 接入共享上下文，避免同一启动链路重复读取 `sys_menu` 与 `sys_plugin_resource_ref`
- [x] 2.5 将 `jobmgmt.withStartupDataSnapshot` 或等价能力接入共享上下文，避免同一启动链路重复读取 `sys_job_group` 与内置 `sys_job`
- [x] 2.6 增加启动快照复用测试，断言一次启动编排中 catalog、integration、job 快照构造次数分别在预算内

## 3. 插件同步 no-op fast path

- [x] 3.1 为插件 registry 和 release metadata 同步补齐写后 snapshot 更新能力，优先用 `InsertAndGetId` 与 `existing + data` 合成启动投影，减少不必要回读
- [x] 3.2 为 manifest menu 和 dynamic route permission menu 增加差异比较函数；无差异时直接返回，不进入 `dao.SysMenu.Transaction`
- [x] 3.3 为 plugin resource ref 同步前置差异判断；无差异时不执行写入和不更新软删除状态
- [x] 3.4 调整 `SyncManifest` / `syncMetadata` 编排，确保 registry、release、menu、resource ref 均无差异时不写库、不回读、不产生空事务
- [x] 3.5 增加源码插件 no-op 同步测试，覆盖无差异时无 `INSERT`、`UPDATE`、`DELETE`、空事务
- [x] 3.6 增加差异同步测试，覆盖 manifest 描述、release snapshot、菜单、route permission、resource ref 变更时仍能正确写入并更新启动快照

## 4. Cron 启动注册去重

- [x] 4.1 审查 `cron.Start`、`syncBuiltinScheduledJobs`、`persistentScheduler.LoadAndRegister` 的查询边界，确认内置任务注册仅使用声明派生 projection snapshot
- [x] 4.2 若仍存在重复读取内置任务的路径，调整调度器启动扫描条件，确保 `LoadAndRegister` 排除 `is_builtin=1`
- [x] 4.3 增加或更新调度器测试，断言内置任务由 `RegisterJobSnapshot` 注册，持久化扫描只加载非内置启用任务
- [x] 4.4 评估 monitor-server 首轮采集任务是否需要延迟到首个 interval 后执行；若调整，补充不影响监控可用性的测试或说明

## 5. 启动摘要与可观测性

- [x] 5.1 增加启动统计采集器，记录插件扫描数量、插件同步变更数量、no-op 插件数量、快照构造次数、内置任务投影数量、启动阶段耗时
- [x] 5.2 在启动完成后输出一条或少量结构化摘要日志，禁止包含完整 SQL 文本
- [x] 5.3 区分宿主启动编排 SQL 和启动后浏览器请求、首轮定时任务 SQL，避免统计口径混淆
- [x] 5.4 增加启动摘要日志测试，断言摘要包含关键字段且默认不依赖 ORM SQL 明细

## 6. 回归验证

- [x] 6.1 运行 `gofmt` 和相关后端单元测试，至少覆盖 `apps/lina-core/internal/service/plugin/...`、`cron`、`jobmgmt`、`cmd` 相关包
- [x] 6.2 在 MySQL 配置下执行 `make init`、`make mock`、后端启动 smoke，确认 admin/admin123 登录链路不回归
- [x] 6.3 在 SQLite 配置下执行后端启动 smoke，确认 SQLite 警告、单节点模式、admin/admin123 登录链路不回归
- [x] 6.4 对比优化前后启动 SQL 基线，记录默认 debug=false 下日志行数变化，以及 debug=true 诊断模式下项目可控 SQL 数量变化
- [x] 6.5 明确记录 i18n 影响判断：本变更不新增、修改或删除运行时语言包、插件 manifest i18n 或 apidoc i18n 资源
- [x] 6.6 明确记录缓存一致性判断：启动快照仅限单次启动编排，不跨请求、不跨进程、不作为业务缓存；集群模式仍以数据库和现有拓扑/修订机制为权威
- [x] 6.7 调用 `lina-review` 完成代码与规范审查

## 实施记录

- 基线来源：`temp/lina-core.log`，启动点为 `2026-05-07T17:40:55.369+08:00` 的 `SHOW TABLES`，HTTP 监听完成为 `2026-05-07T17:40:57.392+08:00`。
- 启动至 HTTP 监听完成耗时约 `2.023s`，窗口内日志 `60` 行，其中 ORM SQL 明细 `55` 条、事务明细 `14` 条、累计 DB 日志耗时约 `206ms`。
- 启动后 `4s` 窗口内日志 `65` 行，其中 ORM SQL 明细 `57` 条、事务明细 `14` 条、累计 DB 日志耗时约 `241ms`。
- 启动后 `10s` 窗口内日志 `116` 行，其中 ORM SQL 明细 `97` 条、事务明细 `14` 条、累计 DB 日志耗时约 `319ms`。
- 主要重复来源：默认 SQL debug、插件 catalog/integration 快照重复构造、插件菜单 no-op 空事务、registry/release 写后回读、runtime frontend prewarm 重复读取插件治理表、monitor-server 首轮定时任务与启动日志混杂。
- `apps/lina-core/manifest/config/config.yaml` 为 `.gitignore` 忽略的本地配置；当前工作区本地值已调整为 `debug: false`。版本化交付模板 `manifest/config/config.template.yaml` 与嵌入模板 `internal/packed/manifest/config/config.template.yaml` 均保持 `debug: false` 并补充诊断模式注释。
- 默认 debug=false 的日志行为通过配置文件测试和启动摘要日志单元测试覆盖；显式 debug=true 通过 `TestDatabaseDebugCanBeEnabledExplicitly` 覆盖配置可读性。2026-05-08 已补充 MySQL 与 SQLite 后端启动 smoke。
- 插件 no-op 菜单和 resource ref 同步通过 `gdb.CatchSQL` 与临时 ORM debug logger 捕获验证，第二次同步不包含 `INSERT`、`UPDATE`、`DELETE`、`BEGIN`、`COMMIT`、`ROLLBACK`。
- Cron 审查结论：`persistentScheduler.LoadAndRegister` 已按 `is_builtin=0` 扫描启用持久化任务，内置任务由 `SyncBuiltinJobs` 返回的 projection snapshot 调用 `RegisterJobSnapshot` 注册；monitor-server 首轮采集属于启动后首轮定时任务，不纳入宿主启动摘要统计，本轮不调整首轮执行语义。
- 3.4 已完成：`SyncManifest` 先准备 release metadata，再绑定 registry 的 `release_id`，最后同步 release 依赖的 resource ref 与 node state，避免首次发现后第二次启动再补写 `sys_plugin_node_state.release_id`。集群模式下针对 `PluginNodeStateMessageManifestSynchronized` 的无差异节点投影新增 no-op 判断，重复清单同步不再刷新 `last_heartbeat_at`；动态运行时收敛和真实生命周期变更仍保留节点心跳写入语义。
- MySQL smoke 已完成：在 `database.default.link=mysql:root:12345678@tcp(127.0.0.1:3306)/linapro?...`、`database.default.debug=false` 下执行 `make init confirm=init rebuild=true` 与 `make mock confirm=mock` 通过；随后构建 `temp/bin/lina-mysql-smoke` 并启动后端，`GET /api/v1/health` 返回 `code=0,status=ok,mode=single`，`POST /api/v1/auth/login` 使用 `admin/admin123` 返回 `code=0` 且 `accessToken` 非空。该轮启动摘要为 `elapsed=2.044s catalogSnapshots=1 integrationSnapshots=1 jobSnapshots=1 pluginScans=1 pluginItems=9 pluginChanged=0 pluginNoop=18 builtinNoop=7 persistentJobs=0`。
- SQLite smoke 已完成：执行 `bash hack/tests/scripts/run-sqlite-smoke.sh` 通过；脚本临时写入 `sqlite::@file(./temp/sqlite/linapro.db)` 配置，执行 `make init confirm=init rebuild=true` 与 `make mock confirm=mock`，启动后端并断言日志包含 `SQLite mode is active`、`SQLite mode only supports single-node deployment`、`do not use SQLite mode in production`，同时确认 health 返回单节点模式且 `admin/admin123` 登录返回非空 token。脚本退出后已恢复 MySQL 本地配置。
- `lina-review` 审查结论：未修改 `dao` / `do` / `entity` 生成文件；本变更无 API DTO、SQL 文件、前端 UI、运行时 i18n 或 apidoc i18n 资源变更；新增启动快照只挂载在单次启动 context，不跨请求和进程。审查中已修复两点：忽略文件 `internal/packed/manifest/config/config.template.yaml` 在配置测试中改为存在才校验；`ReadOnlyList` 仅构造 catalog 快照，避免管理页只读 GET 额外读取 integration 表。
- 补充验证后复审结论：`git status --short` 与 `git ls-files --others --exclude-standard` 显示仅 `tasks.md` 有验证记录变更，无未跟踪文件；`openspec status --change startup-sql-optimization --json` 显示变更 complete，`openspec validate startup-sql-optimization --strict` 通过，`git diff --check` 通过。补充内容未引入代码、API、SQL、前端、i18n 或缓存行为变更，无阻塞问题。
- i18n 影响判断：本变更不新增、修改或删除用户可见前端文案、菜单、路由、按钮、表单、表格、API DTO 文档源文本、插件 manifest i18n 或 apidoc i18n 资源；无需同步运行时语言包。
- 缓存一致性判断：启动快照仅挂载在一次 HTTP 启动编排 context 上，不跨请求、不跨进程、不作为业务缓存；权威数据源仍为数据库。单机模式只使用进程内启动快照；集群模式下插件生命周期和任务治理仍复用现有 cluster/topology、数据库投影和节点状态机制，启动摘要不承担跨实例一致性职责。

## Feedback

- [x] **FB-1**: 将启动阶段观测 key 从字符串字面量抽象为命名类型和常量，避免 `plugin_bootstrap_auto_enable` 等 phase 名称在调用点硬编码
