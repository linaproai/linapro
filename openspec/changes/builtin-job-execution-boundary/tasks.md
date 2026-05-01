## 1. 后端调度边界调整

- [x] 1.1 修改 `apps/lina-core/internal/service/jobmgmt/internal/scheduler`：`LoadAndRegister` 仅加载 `is_builtin=0 AND status=enabled` 的用户自定义任务
- [x] 1.2 保留用户自定义任务 CRUD/启停路径中的 `Refresh` remove-then-register 语义，确认不影响普通任务动态刷新
- [x] 1.3 移除或重定位 `registerJob` 中仅为启动期内置任务重复注册服务的 `gcron.Remove(jobEntryName(job.Id))` 补丁
- [x] 1.4 调整缺失插件 handler 的启动降级逻辑，确保仅对用户自定义插件 handler 任务通过持久化加载降级；插件内置任务由插件生命周期路径投影为不可用

## 2. 内置任务声明驱动注册

- [x] 2.1 修改 `apps/lina-core/internal/service/jobmgmt` 的内置任务同步返回值或查询能力，使同步后可获得每个内置任务的稳定 `sys_job.id`
- [x] 2.2 修改 `apps/lina-core/internal/service/cron`：宿主内置任务同步后直接按代码定义注册 gcron entry，并使用投影 `sys_job.id` 写日志关联
- [x] 2.3 修改插件内置任务同步：源码插件和动态插件 cron 声明启用后直接注册调度 entry，不依赖 `LoadAndRegister` 扫描 `sys_job`
- [x] 2.4 插件禁用或卸载时注销该插件所有内置任务调度 entry，并将对应 `sys_job` 投影为 `paused_by_plugin` / `plugin_unavailable`
- [x] 2.5 确认内置任务投影保留日志关联、列表展示、详情展示、i18n 投影和 source 标识

## 3. 手动触发与管理保护

- [x] 3.1 确认 `TriggerJob` 对 `is_builtin=1` 任务允许手动触发，并通过当前 handler registry / 内置声明校验可执行性
- [x] 3.2 确认 `paused_by_plugin` 或 handler 不可用的内置任务手动触发返回稳定业务错误
- [x] 3.3 确认后端继续拒绝内置任务编辑、删除、启停、重置等执行定义变更
- [x] 3.4 检查前端 `system/job` 页面：内置任务隐藏或禁用编辑、删除、启停、重置；可运行内置任务保留“立即触发”；不可用插件内置任务隐藏或禁用“立即触发”
- [x] 3.5 评估 i18n 影响；本次未新增或修改运行时 UI/API 文案、菜单、按钮、状态标签或后端可见错误码，无需更新 i18n JSON 资源

## 4. 后端测试

- [x] 4.1 更新 scheduler 单测：`LoadAndRegister` 不注册 `is_builtin=1` 任务，只注册用户自定义 enabled 任务
- [x] 4.2 更新或删除重复注册测试：验证内置任务不再通过持久化扫描重复注册，而不是依赖 `registerJob` 覆盖同名 entry
- [x] 4.3 增加宿主内置任务同步测试：同步后注册调度 entry，并能通过投影 `sys_job.id` 写入执行日志
- [x] 4.4 增加插件生命周期测试：插件启用注册插件内置任务 entry，禁用/卸载注销 entry 并投影 `paused_by_plugin`
- [x] 4.5 增加手动触发测试：可运行内置任务允许触发，不可用插件内置任务拒绝触发，用户自定义任务行为不回归

## 5. E2E 与启动日志验证

- [x] 5.1 创建 `hack/tests/e2e/scheduler/job/TC0158-builtin-job-execution-boundary.ts`，验证内置任务只读操作、可运行内置任务“立即触发”和不可用插件内置任务触发入口状态
- [x] 5.2 运行 `npx playwright test e2e/scheduler/job/TC0158-builtin-job-execution-boundary.ts`，2 个子测试通过
- [x] 5.3 运行 `go test -p 1 ./...`，确认后端全量测试通过
- [x] 5.4 运行 `openspec validate builtin-job-execution-boundary --strict`
- [x] 5.5 启动服务并分析 `temp/lina-core.log` 中 `http server started` 前的 SQL debug，确认启动期无内置任务重复注册写库，且 `LoadAndRegister` 不扫描注册 `is_builtin=1` 任务

## 6. 审查收尾

- [x] 6.1 调用 `/lina-review` 审查本变更，覆盖 GoFrame 约束、OpenSpec 符合性、i18n 影响、前端交互和测试结果
- [x] 6.2 根据审查结果修复问题并重新运行受影响测试
- [x] 6.3 更新最终验证记录，准备进入归档前确认

## Verification

- `go test ./internal/service/jobhandler ./internal/service/jobmgmt/... ./internal/service/cron` passed.
- `go test -p 1 ./...` passed under `apps/lina-core`.
- `openspec validate builtin-job-execution-boundary --strict` passed.
- `npx playwright test e2e/scheduler/job/TC0158-builtin-job-execution-boundary.ts` passed: 2 tests.
- Startup SQL before `http server started`: `startup_sql_statements=48`, `startup_select=26`, `startup_show=8`, `startup_insert=0`, `startup_update=0`, `startup_delete=0`, `startup_sys_job_writes=0`.
- Startup evidence: persistent scheduler query uses `FROM sys_job WHERE status='enabled' AND is_builtin=0`; built-in projection snapshot query remains `WHERE is_builtin=1` for display/log linkage only.
