## ADDED Requirements

### Requirement: 定时任务表租户化
`sys_job` 与 `sys_job_log` SHALL 加 `tenant_id`;`tenant_id=0` 表示平台级内置任务;`tenant_id>0` 表示租户业务任务。

#### Scenario: 租户业务任务
- **WHEN** 租户 A 管理员创建任务 `daily-report`
- **THEN** 写入 `sys_job(tenant_id=A, ...)`,执行时 `bizctx.TenantId=A`

#### Scenario: 平台级内置任务
- **WHEN** seed 内置任务"session 清理"
- **THEN** 写入 `sys_job(tenant_id=0, ...)`,执行时 `bizctx.TenantId=0`

### Requirement: 任务执行的租户上下文绑定
调度器 SHALL 在触发任务执行前,根据任务的 `tenant_id` 构造 `bizctx.Context` 并注入 `TenantId`;任务 handler 内通过 `bizctx.TenantId(ctx)` 读取。

#### Scenario: handler 见到正确租户
- **WHEN** `daily-report` 任务在租户 A 触发
- **THEN** handler 内 `bizctx.TenantId(ctx) = A`
- **AND** handler 中所有 DAO 读取自动按 A 过滤

### Requirement: 任务查询/创建/修改按租户隔离
租户管理员 SHALL 仅可见/改本租户任务,且 MUST NOT 创建 `tenant_id=0` 的平台任务;平台管理员创建或修改跨租户/平台任务必须通过 `/platform/jobs/*` 管理平台接口。

#### Scenario: 租户尝试创建平台任务
- **WHEN** 租户管理员请求中显式 `tenant_id=0`
- **THEN** 返回 `bizerr.CodeJobTenantForbidden`

### Requirement: 任务日志按租户隔离查询
`sys_job_log` 查询 SHALL 按租户过滤;租户内可见自己任务的执行日志。平台管理员仅通过 `/platform/job-logs` 管理平台接口查看全量;impersonation 模式下仍仅可见目标租户任务日志。

#### Scenario: 任务日志隔离
- **WHEN** 租户 A 管理员查询 `daily-report` 执行历史
- **THEN** 返回 `tenant_id=A` 且 `job_id=daily-report.id` 的日志
