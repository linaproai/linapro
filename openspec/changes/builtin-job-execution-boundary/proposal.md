## Why

当前定时任务启动链路把 `sys_job` 同时作为内置任务的治理投影和执行注册来源，导致内置任务先由代码/插件声明同步到表，再被持久化调度器从表中扫描并重复注册。需要明确内置任务与用户自定义任务的执行边界，减少启动期数据库操作和重复注册补丁，并让系统行为更符合“代码声明驱动内置能力、数据表承载治理展示”的架构定位。

## What Changes

- 将内置任务的执行源限定为宿主代码定义或插件 cron 声明，`sys_job.is_builtin=1` 仅作为管理台展示、日志关联、审计快照和治理投影。
- 持久化调度器启动加载仅扫描并注册 `is_builtin=0 AND status=enabled` 的用户自定义任务。
- 内置任务同步流程负责根据代码/插件声明注册运行期调度 entry，并同步或更新 `sys_job` 投影记录。
- 内置任务继续允许手动触发，但执行定义以代码/插件声明为准，不以管理表中可变字段作为真实执行源。
- 管理台与后端继续禁止内置任务的启停、编辑、重置和删除等会改变执行定义的操作。
- 插件禁用或卸载后，插件内置任务的治理投影可保留，但调度 entry 必须注销，且任务状态应投影为不可运行状态。
- 移除因内置任务重复注册而加入的启动期重复注册补丁与相应测试定位。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `cron-jobs`: 调整启动期调度注册规则，区分用户自定义任务和内置任务的真实执行源。
- `cron-job-management`: 明确内置任务投影、手动触发、状态展示和不可编辑边界。
- `cron-handler-registry`: 明确插件 cron 声明作为插件内置任务的执行源，并定义插件禁用/卸载后的调度注销与投影状态。

## Impact

- 后端调度器：`apps/lina-core/internal/service/jobmgmt/internal/scheduler/` 的启动注册、刷新、触发和测试。
- 内置任务同步：`apps/lina-core/internal/service/cron/` 的宿主内置任务、插件内置任务投影与注册链路。
- 任务管理服务：`apps/lina-core/internal/service/jobmgmt/` 的内置任务状态保护、手动触发、日志关联与 i18n 投影。
- 插件 cron 集成：源码插件与动态插件声明的 managed cron job 生命周期处理。
- 前端管理台：`apps/lina-vben/apps/web-antd/src/views/system/job/` 的内置任务操作可见性、手动触发入口和只读展示。
- 测试：后端调度器、任务管理、插件生命周期，以及必要的 E2E 启动与管理台行为验证。
