## 1. 数据库与字典

- [x] 1.1 在 `apps/lina-core/manifest/sql/` 新增 `014-scheduled-job-management.sql`,包含 `sys_job_group / sys_job / sys_job_log` 三张表 DDL(带索引)、外键约束、幂等 `CREATE TABLE IF NOT EXISTS`
- [x] 1.2 在同一 SQL 中插入默认分组(`code=default, is_default=1`)、系统参数 `cron.shell.enabled=false`、`cron.log.retention={"mode":"days","value":30}`
- [x] 1.3 在同一 SQL 中插入菜单(任务管理 / 分组管理 / 执行日志)与按钮权限(`system:job:add/edit/remove/status/trigger/reset`、`system:jobgroup:add/edit/remove`、`system:joblog:remove/cancel`、`system:job:shell`),并注册菜单权限 `system:job:list / system:jobgroup:list / system:joblog:list`
- [x] 1.4 在同一 SQL 中插入字典类型与字典数据:`cron_job_status / cron_job_task_type / cron_job_scope / cron_job_concurrency / cron_job_trigger / cron_job_log_status / cron_log_retention_mode`
- [x] 1.5 在同一 SQL 中 seed 一个内置任务 `host:cleanup-job-logs`(master-only、singleton、每日 03:17、is_builtin=1、seed_version=1)
- [x] 1.6 执行 `make init` 验证 SQL 幂等可重入;确认重复执行无报错

## 2. 后端 DAO 与 API 骨架

- [x] 2.1 在 `apps/lina-core/internal/cmd/` 相关表配置处更新后运行 `make dao`,生成 `dao/sys_job.go`、`dao/sys_job_group.go`、`dao/sys_job_log.go` 及 DO/Entity
- [x] 2.2 在 `api/job/v1/` 按接口用途拆文件创建 DTO:`job_list.go / job_detail.go / job_create.go / job_update.go / job_delete.go / job_status.go / job_trigger.go / job_reset.go`
- [x] 2.3 在 `api/jobgroup/v1/` 创建分组 DTO:`group_list.go / group_create.go / group_update.go / group_delete.go`
- [x] 2.4 在 `api/joblog/v1/` 创建日志 DTO:`log_list.go / log_detail.go / log_cancel.go / log_clear.go`
- [x] 2.5 在 `api/jobhandler/v1/` 创建 handler 注册表查询 DTO:`handler_list.go / handler_detail.go`
- [x] 2.6 在 `api/job/v1/job_cron_preview.go` 添加 cron 表达式预览 DTO(入参 `expr / timezone`,出参最近 5 次触发时刻)
- [x] 2.7 所有 DTO `g.Meta` 带 `dc / permission` 标签、所有字段带 `dc / eg / json`,满足接口文档规范;修改后运行 `make ctrl` 生成控制器骨架

## 3. 后端 Handler 注册表

- [x] 3.1 创建 `internal/service/jobhandler/jobhandler.go`(主文件):定义 `Registry / HandlerDef / HandlerInfo / HandlerSource` 接口与结构体、`serviceImpl`、`New()`;主文件承载组件包注释
- [x] 3.2 创建 `internal/service/jobhandler/jobhandler_host.go`:宿主内置 handler 注册入口 `RegisterHostHandlers()`,按命名 `host:xxx` 注册第一批 handler(清理日志、清理会话日志、重新生成会话指纹等 seed 内置任务所需)
- [x] 3.3 创建 `internal/service/jobhandler/jobhandler_plugin.go`:定义插件生命周期观察器,在 `service/plugin` 的启用/禁用/卸载成功路径中同步调 `Register / Unregister`
- [x] 3.4 创建 `internal/service/jobhandler/jobhandler_schema.go`:基于 `JSON Schema draft-07` 标量子集兼容校验库封装 `ValidateParams(schemaText, paramsJSON) error`,并显式拒绝超出本迭代支持范围的关键字
- [x] 3.5 创建 `internal/service/jobhandler/jobhandler_test.go`:Register 冲突、Lookup 命中/未命中、Unregister 级联通知回调

## 4. 后端任务持久化与调度核心

- [x] 4.1 创建 `internal/service/jobmgmt/jobmgmt.go`(主文件):定义 `Service / JobMgmt`、`New()`、基础依赖(DAO、handler registry、scheduler)
- [x] 4.2 创建 `internal/service/jobmgmt/jobmgmt_group.go`:分组 CRUD(默认分组不可删、删除时任务迁到默认分组)
- [x] 4.3 创建 `internal/service/jobmgmt/jobmgmt_job_crud.go`:任务 CRUD(组内唯一、内置任务字段锁定校验、开关联动、任务变更后通知 scheduler 刷新)
- [x] 4.4 创建 `internal/service/jobmgmt/jobmgmt_job_status.go`:启用/禁用/重置计数;启用时校验 handler 是否可用
- [x] 4.5 创建 `internal/service/jobmgmt/jobmgmt_log.go`:日志查询、清空、清理策略计算(全局 + 任务级覆盖)
- [x] 4.6 创建 `internal/service/jobmgmt/jobmgmt_cron_preview.go`:基于 `robfig/cron` parser 计算下一次触发时刻(按 timezone 解析)
- [x] 4.7 创建 `internal/service/jobmgmt/jobmgmt_test.go`:CRUD 单测覆盖内置任务锁定、分组删除迁移、公共 `timeout_seconds` 校验、并发策略校验

## 5. 后端调度器组件(gcron 之上)

- [x] 5.1 创建 `internal/service/jobmgmt/scheduler/scheduler.go`(主文件):定义 `Scheduler` 接口、`New()`、启动时 `LoadAndRegister(ctx)`、per-job 互斥锁
- [x] 5.2 创建 `internal/service/jobmgmt/scheduler/scheduler_register.go`:`Add / Remove / Refresh` 对 gcron 的封装,含小容量 LRU 缓存与主动失效
- [x] 5.3 创建 `internal/service/jobmgmt/scheduler/scheduler_runner.go`:tick wrapper `runJob(jobID, trigger)` —— 判断 scope / concurrency / max_executions / timeout_seconds,分发到 handler 或 shell 执行器,捕获结果写日志
- [x] 5.4 创建 `internal/service/jobmgmt/scheduler/scheduler_cancel.go`:维护 `runningInstances map[logID]cancelFn`,支持 `CancelLog(logID)`
- [x] 5.5 创建 `internal/service/jobmgmt/scheduler/scheduler_test.go`:覆盖 Add/Remove 竞态、scope 守卫跳过、singleton/parallel 计数、max_executions 触发禁用

## 6. 后端 Shell 执行器

- [x] 6.1 创建 `internal/service/jobmgmt/shellexec/shellexec.go`(主文件):定义 `Executor` 接口、`New()`、`cron.shell.enabled` 与 Windows 平台守卫
- [x] 6.2 创建 `internal/service/jobmgmt/shellexec/shellexec_process.go`:`/bin/sh -c` 启动子进程、`Setpgid`、work_dir/env 合并、stdout/stderr `LimitReader` 64KB 截留
- [x] 6.3 创建 `internal/service/jobmgmt/shellexec/shellexec_lifecycle.go`:超时 `kill -- -<pgid>`、手动终止 SIGTERM → 5 秒 → SIGKILL 升级路径,并避免写入重复 `oper_log`
- [x] 6.4 创建 `internal/service/jobmgmt/shellexec/shellexec_test.go`:输出截断、超时终止、手动取消

## 7. 后端控制器实现

- [x] 7.1 `controller/job/v1_new.go` 初始化依赖字段(`jobmgmt.Service / jobhandler.Registry / scheduler.Scheduler`),`NewV1()` 一次性注入
- [x] 7.2 `controller/job/v1_*.go` 按接口实现业务:列表/详情/创建/更新/删除/启停/触发/重置/cron 预览
- [x] 7.3 `controller/jobgroup/v1_*.go` 实现分组 CRUD
- [x] 7.4 `controller/joblog/v1_*.go` 实现日志查询/详情/清空/取消
- [x] 7.5 `controller/jobhandler/v1_*.go` 实现 handler 列表/详情
- [x] 7.6 按资源声明 `g.Meta.permission`:任务接口使用 `system:job:*`,分组接口使用 `system:jobgroup:*`,日志接口使用 `system:joblog:*`;Shell 创建/修改/触发接口追加 `system:job:shell`,Shell 日志取消接口需组合 `system:joblog:cancel + system:job:shell`
- [x] 7.7 为 shell 创建/修改/触发/取消接口声明 `operLog` 元标签并返回必要关联标识(如 `log_id`),复用宿主 `OperLog` 中间件完成单条审计记录写入
- [x] 7.8 扩展宿主审计请求参数脱敏规则,对 Shell 相关接口中的 `env` 载荷做值级遮罩,确保 `oper_log` 不落原始环境变量值

## 8. 后端集成到启动流程

- [x] 8.1 修改 `internal/service/cron/cron.go`:在 `Start(ctx)` 末尾调用 `jobmgmtScheduler.LoadAndRegister(ctx)`;通过构造注入,不在 tick 内临时 `New`
- [x] 8.2 修改 `internal/service/plugin` 的启用/禁用/卸载成功路径,通过显式生命周期回调同步通知 `jobhandler` 观察器
- [x] 8.3 修改 `internal/service/config` 暴露 `cron.shell.enabled` 与 `cron.log.retention` 的类型化读取入口并复用现有 runtime param 刷新机制;`sysconfig` 继续负责参数数据管理
- [x] 8.4 在 `cmd` 启动装配处,完成 `jobhandler / jobmgmt / scheduler / shellexec` 的依赖注入

## 9. 前端 API 与适配器

- [x] 9.1 在 `apps/lina-vben/apps/web-antd/src/api/system/` 新增 `job.ts / jobGroup.ts / jobLog.ts / jobHandler.ts`,覆盖 CRUD/触发/取消/日志/cron 预览/handler 列表接口
- [x] 9.2 在 `src/adapter/form/` 新增 handler 动态参数子表单适配器:根据 `JSON Schema draft-07` 标量子集生成 Vben form `schema` 数组(支持 string/integer/number/boolean/enum/date/date-time/textarea)
- [x] 9.3 在 `src/adapter/vxe-table/` 注册任务列表、日志列表、分组列表列定义

## 10. 前端页面:分组管理

- [x] 10.1 创建 `src/views/system/job-group/index.vue`:`Page + useVbenVxeGrid` 列表 + `GhostButton + Popconfirm` 操作列
- [x] 10.2 创建 `src/views/system/job-group/modal.vue`:`useVbenModal + useVbenForm` 新增/编辑弹窗
- [x] 10.3 默认分组行禁用删除按钮、显示"默认分组"标签

## 11. 前端页面:任务管理

- [x] 11.1 创建 `src/views/system/job/index.vue`:列表页,顶部筛选含分组、状态、任务类型、关键字
- [x] 11.2 创建 `src/views/system/job/form.vue`:`useVbenModal` 主弹窗;顶部 Tab 切换 handler / shell 类型(shell tab 在 `cron.shell.enabled=false` 或无 `system:job:shell` 权限时隐藏)
- [x] 11.3 创建 `src/views/system/job/form-handler.vue`:handler 选择下拉 + 动态渲染参数子表单(基于 schema)+ cron 表达式 + timezone + scope + concurrency + max_concurrency + timeout_seconds + max_executions + log_retention_override
- [x] 11.4 创建 `src/views/system/job/form-shell.vue`:shell_cmd 多行 textarea + work_dir + env(KV 表格,编辑态遮罩既有值) + timeout_seconds + 告警色提示
- [x] 11.5 操作列:启用/禁用、立即执行、编辑、删除、重置计数;内置任务仅显示部分按钮
- [x] 11.6 `paused_by_plugin` 状态显式标红并展示"插件处理器不可用"tooltip

## 12. 前端页面:执行日志

- [x] 12.1 创建 `src/views/system/job-log/index.vue`:日志列表,顶部筛选含任务、状态、节点、时间范围
- [x] 12.2 创建 `src/views/system/job-log/detail.vue`:`useVbenModal` 详情弹窗,展示 trigger / params_snapshot / result_json(shell 任务含 stdout/stderr 代码高亮)
- [x] 12.3 正在运行的日志行显式"终止"按钮,确认后调 cancel 接口
- [x] 12.4 日志列表的批量清空功能(需 `system:joblog:remove` 权限),终止按钮需 `system:joblog:cancel`;终止 shell 实例时前端还需叠加 `system:job:shell`

## 13. 前端路由与菜单

- [x] 13.1 在 `src/router/routes/modules/system.ts` 或等价路由模块中注册 `/system/job / /system/job-group / /system/job-log`
- [x] 13.2 菜单图标使用 `IconifyIcon`(`ant-design:clock-circle-outlined` 或同类图标)
- [x] 13.3 按钮权限对应 `system:job:*`、`system:jobgroup:*`、`system:joblog:*`,Shell 创建/修改/触发额外叠加 `system:job:shell`

## 14. E2E 测试用例

- [x] 14.1 TC0081 定时任务分组 CRUD(新增、编辑、删除非默认组、默认组不可删)——`hack/tests/e2e/system/job/TC0081-job-group-crud.ts`
- [x] 14.2 TC0082 Handler 类型任务 CRUD + 参数动态表单渲染——`hack/tests/e2e/system/job/TC0082-job-handler-crud.ts`
- [x] 14.3 TC0083 Shell 类型任务创建(开启 `cron.shell.enabled` 前置)——`hack/tests/e2e/system/job/TC0083-job-shell-crud.ts`
- [x] 14.4 TC0084 Shell 全局开关关闭时前端隐藏 shell 类型选项、后端拒绝写入——`hack/tests/e2e/system/job/TC0084-job-shell-switch.ts`
- [x] 14.5 TC0085 任务启用/禁用、状态切换即时生效——`TC0085-job-enable-disable.ts`
- [x] 14.6 TC0086 手动触发一次执行 & 日志 trigger=manual、不计入 executed_count——`TC0086-job-manual-trigger.ts`
- [x] 14.7 TC0087 长任务手动终止 → 日志 status=cancelled——`hack/tests/e2e/system/job/TC0087-job-manual-cancel.ts`
- [x] 14.8 TC0088 `max_executions` 达上限自动禁用 + `stop_reason` 显示——`TC0088-job-max-executions.ts`
- [x] 14.9 TC0089 执行日志列表筛选、详情、清空——`TC0089-job-log-list.ts`
- [x] 14.10 TC0090 插件禁用导致任务 paused_by_plugin 标红、启用按钮禁用——`TC0090-job-plugin-cascade.ts`
- [x] 14.11 TC0091 系统内置任务 cron 可改、handler_ref 锁定、删除被拒——`TC0091-job-builtin-readonly.ts`
- [x] 14.12 TC0092 删除非默认分组时任务迁移到默认分组——`TC0092-job-group-migration.ts`
- [x] 14.13 TC0093 时区字段持久化与下一次执行时间预览——`TC0093-job-timezone-preview.ts`
- [x] 14.14 TC0094 shell 任务输出 stdout/stderr 截断后可查看——`hack/tests/e2e/system/job/TC0094-job-shell-output.ts`
- [x] 14.15 TC0095 Handler 任务超时后日志 `status=timeout`、`err_msg` 含超时时长——`hack/tests/e2e/system/job/TC0095-job-handler-timeout.ts`
- [x] 14.16 TC0096 无 `system:job:shell` 权限时禁止终止运行中的 shell 实例——`hack/tests/e2e/system/job/TC0096-job-shell-cancel-permission.ts`
- [x] 14.17 所有新增用例在执行过程中自动运行 `pnpm test -- TC008x` / `pnpm test -- TC009x` 验证通过;测试文件命名与 TC ID 严格对应

## 15. 文档与收尾

- [x] 15.1 更新 `apps/lina-core/README.md / README.zh_CN.md`:在模块列表中新增"定时任务管理"章节(中英同步)
- [x] 15.2 更新 `apps/lina-vben/apps/web-antd/README.md / README.zh_CN.md`:新增路由与页面入口说明
- [x] 15.3 验证 `make init / make dao / make ctrl / make dev / make test` 全流程通过
- [x] 15.4 调用 `/openspec-review` 技能进行代码与规范审查,处理所有严重问题
- [ ] 15.5 用户确认功能完成后,执行 `/opsx:archive scheduled-job-management`(归档时 proposal/design/tasks 与 specs 将被统一重写为英文,符合归档语言规范)

## Feedback

- [x] **FB-1**: 统一任务管理权限矩阵并移除未定义的导出权限
- [x] **FB-2**: 固化 `job / jobgroup / joblog / jobhandler` 资源拆分与控制器归属
- [x] **FB-3**: 明确 handler `ParamsSchema` 的 `JSON Schema draft-07` 标量子集与校验边界
- [x] **FB-4**: 统一 shell 审计复用宿主 `OperLog` 中间件并避免重复 `oper_log`
- [x] **FB-5**: 明确 `cron.shell.enabled` 与 `cron.log.retention` 的运行时读取归属
- [x] **FB-6**: 将插件 handler 联动从模糊事件总线改为显式生命周期回调
- [x] **FB-7**: 明确 `timeout_seconds` 为所有任务的公共字段并补齐测试规划
- [x] **FB-8**: 明确 shell `env` 审计脱敏边界并追加实现任务
- [x] **FB-9**: 明确终止 shell 实例时的组合权限与测试覆盖
- [x] **FB-10**: 将 admin 用户权限放行与菜单查询逻辑从角色特判调整为用户特判，并移除对 `sys_role_menu` 的超管依赖
- [x] **FB-11**: 统一 SQL seed 写法，移除 `ON DUPLICATE KEY UPDATE` 与显式自增主键写入；优先修正定时任务 SQL，并补充全仓库 SQL 审查任务
- [x] **FB-12**: 移除 admin 角色菜单种子与插件菜单同步中的冗余 `sys_role_menu` 写入
- [x] **FB-13**: 审查并同步其他 SQL 副本，移除残留的 admin 角色菜单绑定与旧版显式自增 `id` seed 写法
- [x] **FB-14**: 将内置 admin 账号策略从 `pkg` 收拢到 `service/user` 组件内部，并消除跨组件误导性公共依赖
- [x] **FB-15**: 补齐本次新增控制器骨架注释并修正审查遗漏，确保未跟踪控制器文件也纳入当前轮次检查
- [x] **FB-16**: 改进 `openspec-review` 技能范围识别流程，将未跟踪文件纳入规范审查范围并禁止仅依赖 `git diff`
- [x] **FB-17**: 将 SQL seed 禁止 `ON DUPLICATE KEY UPDATE` 与禁止显式写入自增 `id` 上升为项目规范，并同步审查技能与活跃设计文档
- [x] **FB-18**: 将定时任务调度能力视为当前迭代核心功能，补齐关键执行链路的单元测试与 E2E 覆盖并完成通过验证
- [x] **FB-19**: 在系统管理下增加“定时任务”目录菜单，并保持原有页面入口兼容可访问
- [x] **FB-20**: 明确 `paused_by_plugin` 状态含义，改进任务状态筛选与列表 tooltip 文案
- [x] **FB-21**: 明确 Cron 表达式同时支持 5 段与 6 段，并补充表单帮助文案
- [x] **FB-22**: 将任务列表中的调度范围、并发策略改为易理解的中文展示
- [x] **FB-23**: 为调度范围、并发策略、日志保留补充问号提示，并在“跟随系统”说明中展示当前系统日志保留策略
- [x] **FB-24**: 为定时任务导航、旧入口跳转与帮助文案改进补充 `TC0097` E2E 覆盖
- [x] **FB-25**: 修复执行日志“清空”在无 `jobId` 条件时触发 ORM 删除保护报错的问题
- [x] **FB-26**: 为执行日志列表补齐批量删除入口，并复用与操作日志一致的多选删除交互
- [x] **FB-27**: 将 5 段 Cron 表达式归一化后的秒位占位从 `0` 调整为 `#`，并同步校验、说明文案与测试
- [x] **FB-28**: 将新增/编辑弹窗中的 `Cron 表达式` 标签改为 `定时表达式`，并消除该字段标题换行问题
- [x] **FB-29**: 将定时表达式输入框改进为代码框风格，并评估当前组件能力下的高亮支持边界
- [x] **FB-30**: 为调度范围、并发策略、日志保留的帮助提示增加分段换行，提升长文案可读性
- [x] **FB-31**: 将 `all_node` 的中文展示文案统一调整为 `所有节点执行`
- [x] **FB-32**: 将任务时区字段改为支持常用下拉与自定义输入的组件，并默认选中宿主当前系统时区
- [x] **FB-33**: 允许停用状态的任务仍可从列表手动“立即执行”用于测试，同时保持 `paused_by_plugin` 的限制
- [x] **FB-34**: 将任务列表中的定时表达式列改为代码高亮展示，提升 Cron 可读性
- [x] **FB-35**: 为 `超时时间(秒)` 与 `最大执行次数` 增加帮助提示，并明确 `0` 表示不限制执行次数
- [x] **FB-36**: 调整系统内置任务编辑页两处锁定提示块与表单区域之间的上下间距，保证不少于 `5px`
- [x] **FB-37**: 在“全新项目无需兼容历史债务”前提下，移除 `014-scheduled-job-management.sql` 中不必要的升级回填式菜单/字典/角色关联 SQL
- [x] **FB-38**: 统一 `014-scheduled-job-management.sql` 中调度范围、并发策略等字典种子文案，确保与当前最终 UI 文案一致
- [x] **FB-39**: 将 `jobmgmt` 组件内部使用的调度器与 Shell 执行器子组件收拢到 `internal` 目录，避免被外部直接引用
- [x] **FB-40**: 在新增与编辑任务时严格校验定时表达式、时区、状态与公共调度字段，非法输入需返回明确错误
- [x] **FB-41**: 将新增与编辑页中的定时表达式输入改为代码样式输入框，无需分段代码高亮
- [x] **FB-42**: 调整 Shell 任务警告提示块与表单区域的垂直间距，保证不少于 `5px`
- [x] **FB-43**: 将宿主服务与插件源码注册的定时任务统一投影到 `sys_job` 并在任务管理中完整展示
- [x] **FB-44**: 将公共任务创建/编辑入口收敛为仅支持 Shell 任务，Handler 任务改为源码注册只读展示
- [x] **FB-45**: 在任务列表与详情中增加 `宿主内置 / 插件内置 / 用户创建` 来源展示，并对源码注册任务收紧为只读操作
- [x] **FB-46**: 调整系统管理下定时任务目录与子菜单图标，确保左侧菜单图标全局唯一且不重复
- [x] **FB-47**: 补充并更新定时任务 E2E 覆盖，验证源码注册任务可见、Handler 创建入口移除与只读行为
- [x] **FB-48**: 修复仓库中剩余的 E2E TypeScript 编译错误，确保相关测试页对象与用例可通过 `tsc --noEmit`
- [x] **FB-49**: 将任务列表中的定时表达式展示收敛为简单代码块样式，不再做分段高亮
- [x] **FB-50**: 修复定时触发执行后 `executed_count` 未累计的问题，并补齐回归测试
- [x] **FB-51**: 当任务列表某行没有二级操作时隐藏空的“更多”按钮
- [x] **FB-52**: 将任务启停入口从操作列迁移到状态列点击切换，并补齐交互回归测试
- [x] **FB-53**: 恢复任务列表状态列为只读展示，启停改回通过编辑弹窗中的任务状态字段完成
- [x] **FB-54**: 为动态插件补齐定时任务声明、投影与执行链路，确保插件内置任务可进入统一任务管理并支持维护
- [x] **FB-55**: 将 guest-side runtime host service SDK 整理为与 `DataHostService` 一致的独立接口对象入口，统一调用方式
- [x] **FB-56**: 明确定时任务治理契约应保持独立边界，不将插件定时任务声明直接并入 runtime host service
- [x] **FB-57**: 收敛动态插件定时任务中的枚举值与协议字符串硬编码，统一到公共 bridge 常量与辅助函数
- [x] **FB-58**: 动态插件定时任务改为通过独立 cron host service 代码注册，移除 YAML/custom section 声明链路
- [x] **FB-59**: 修复插件生命周期同步定时任务时误对源码插件执行动态 cron 发现，避免动态插件安装授权流程被无关插件错误打断
- [x] **FB-60**: 改进动态插件授权页的宿主服务文案与 cron 服务展示，统一使用“数据表名 / 存储路径 / 访问地址”等等长标签，并在“任务服务”卡片中展示已注册定时任务的名称、表达式、调度范围与并发策略
- [x] **FB-61**: 调整动态插件授权页宿主服务排序与 cron 卡片样式，将运行时服务移到最底部、移除“定时任务”摘要标签，并将任务属性标题改为粗体展示
- [x] **FB-62**: 统一插件授权页与详情页的宿主服务标签背景语义，并将详情页“当前生效范围”文案收敛为“生效范围”
