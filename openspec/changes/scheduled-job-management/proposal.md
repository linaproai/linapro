## Why

当前 `apps/lina-core/internal/service/cron` 只承载代码内置的治理任务(会话清理、监控采集、运行参数同步等),没有面向运维用户的定时任务管理入口。现实场景中,管理员需要按需创建业务定时任务(数据清理、数据对账、外部系统健康探测等),并具备以下能力:可视化 CRUD、分组管理、启停控制、执行日志、次数策略、集群调度范围选择、并发策略、手动触发与终止,以及对自定义 shell 命令的支持。

本次变更在不影响现有代码内置 cron 能力的前提下,新增一整套"用户可管理的定时任务"子系统,同时扩展 handler 注册机制以复用宿主与插件的已有代码能力。

## What Changes

- **数据模型**: 新增 `sys_job_group`、`sys_job`、`sys_job_log` 三张表,保存用户定义的定时任务、分组与执行日志。
- **任务类型**: 支持两种执行形态
  - `handler` 类型:调用宿主或插件注册的具名 handler,参数通过 handler 声明的 `JSON Schema draft-07` 标量子集校验。
  - `shell` 类型:执行用户自定义的 `/bin/sh -c` 多行脚本,带强制 timeout、工作目录、环境变量、进程组 kill、stdout/stderr 截断。
- **调度语义**: 每个任务独立配置 `scope`(master-only / all-node)、`concurrency`(singleton / parallel + max_concurrency)、`timezone`、`cron_expr`、`timeout_seconds`、`max_executions`、`log_retention`。
- **Handler 注册表**: 新增 host-side handler 注册与插件 handler 订阅,插件禁用/卸载时关联任务自动置 `paused_by_plugin` 并在 UI 标红。
- **系统内置任务**: 通过 seed SQL 入库的内置任务(`is_builtin=1`)允许用户修改 `cron_expr / timezone / status / timeout_seconds / max_executions / log_retention_override`,其余字段(handler_ref / params / scope / concurrency / task_type)锁死;升级时 seed 不覆盖用户改过的字段。
- **Shell 安全治理**: 新增权限点 `system:job:shell`、系统参数 `cron.shell.enabled` 全局开关,并复用宿主现有 `oper_log` 中间件对 shell 任务创建/修改/手动触发/手动终止各写入且仅写入一条审计日志;审计中保留 `shell_cmd / work_dir / timeout_seconds` 快照,但不得落原始 `env` 值。
- **人机交互**: UI 提供手动触发一次(`trigger=manual`,不计入 `executed_count`)和手动终止运行中实例(`ctx.Cancel` + 进程组 kill)。
- **执行次数策略**: `max_executions>0` 时,达到上限自动禁用并在 `stop_reason` 记录原因;手动重置后可重新启用。
- **日志清理**: 全局默认清理策略(保留条数或天数)+ 任务级覆盖;由系统内置定时任务负责执行清理。
- **前端**: 基于 Vben5(`useVbenForm / useVbenModal / useVbenVxeGrid / GhostButton + Popconfirm / IconifyIcon`)构建任务管理、分组管理、执行日志三个页面;handler 参数区根据 schema 动态渲染表单。
- **E2E**: 新增针对任务 CRUD、分组、手动触发/终止、shell 开关、集群调度范围的 Playwright 测试用例。

## Capabilities

### New Capabilities
- `cron-job-management`: 用户可管理的定时任务 CRUD、分组、启停、执行日志、次数策略、手动触发与终止、日志清理策略。
- `cron-handler-registry`: Handler 注册表契约,涵盖宿主注册、插件订阅、JSON Schema 参数声明、插件生命周期对任务状态的级联影响。
- `cron-shell-execution`: Shell 类型任务的安全边界、执行上下文、进程生命周期与输出截留规则。

### Modified Capabilities
- `cron-jobs`: 将现有"主节点/全节点"调度语义扩展到用户可管理任务;现有内置任务继续满足,但语义在新增 `scope` 字段维度下复用。

## Impact

- **数据库**: 新增 `manifest/sql/<序号>-scheduled-job-management.sql`,包含三张新表 DDL、seed(默认分组、内置任务记录、全局清理参数、权限点)、幂等保护。
- **后端新增模块**:
  - `internal/service/cron` 内现有 `Service` 继续保留并扩展注册能力。
  - 新增 `internal/service/jobmgmt`(任务/分组/日志领域服务)与 `internal/service/jobhandler`(handler 注册表),按规范分子包与主文件。
  - 按资源拆分新增 `api/job/v1/*.go`、`api/jobgroup/v1/*.go`、`api/joblog/v1/*.go`、`api/jobhandler/v1/*.go` REST 接口。
  - 对应新增 `internal/controller/job/*`、`internal/controller/jobgroup/*`、`internal/controller/joblog/*`、`internal/controller/jobhandler/*` 控制器骨架并分别实现。
- **后端改造**:
  - `service/cron/cron.go` 在启动阶段拉起新增子系统(从数据库加载启用任务 → 注册 gcron);CRUD 动态刷新注册表。
  - `service/plugin` 在启用/禁用/卸载成功路径中通过显式生命周期回调同步 handler 可用性,供任务状态级联,不引入独立事件总线。
  - `service/config` 新增 `cron.shell.enabled` 与 `cron.log.retention` 的类型化运行时读取入口,`sysconfig` 继续负责 `sys_config` 数据管理与页面 CRUD。
- **前端**:`apps/lina-vben/apps/web-antd` 新增路由、视图、API、adapter;菜单项与按钮权限遵循模块解耦设计原则,在宿主未启用相关特性时隐藏。
- **权限**: 新增菜单权限 `system:job:list / system:jobgroup:list / system:joblog:list`、按钮权限 `system:job:add/edit/remove/status/trigger/reset`、`system:jobgroup:add/edit/remove`、`system:joblog:remove/cancel`,以及 shell 附加权限 `system:job:shell`。
- **E2E**: `hack/tests/e2e/system/job/TC*` 新增任务/分组/日志的核心测试用例。
- **字典**: 任务状态、触发方式、结果状态、shell 模式启用状态、清理策略类型等枚举统一进字典模块(`sys_dict_type` + `sys_dict_data`)。
