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
- [x] 11.6 `paused_by_plugin` 状态显式标红并展示"插件不可用"tooltip

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
