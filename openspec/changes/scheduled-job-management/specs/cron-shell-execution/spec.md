## ADDED Requirements

### Requirement: Shell 任务全局开关

系统 SHALL 通过系统参数 `cron.shell.enabled` 控制是否允许 Shell 类型任务的创建、修改与执行,默认关闭。

#### Scenario: 开关关闭时拒绝创建

- **WHEN** `cron.shell.enabled=false` 且用户尝试创建 `task_type=shell` 任务
- **THEN** 系统 SHALL 拒绝创建并返回明确错误

#### Scenario: 开关关闭时拒绝修改

- **WHEN** `cron.shell.enabled=false` 且用户尝试修改已存在 Shell 任务的任一字段
- **THEN** 系统 SHALL 拒绝修改

#### Scenario: 开关关闭时拒绝执行

- **WHEN** `cron.shell.enabled=false` 且 Shell 任务到达 tick 或被手动触发
- **THEN** 系统 SHALL 拒绝本次执行
- **AND** 日志 `status=failed`、`err_msg` 指明 shell 开关已关闭

#### Scenario: UI 隐藏 shell 选项

- **WHEN** 前端从系统参数接口读取到 `cron.shell.enabled=false`
- **THEN** UI SHALL 在任务类型选择中隐藏 `shell` 选项
- **AND** 对已存在的 shell 任务仅允许只读查看

#### Scenario: Windows 运行时强制关闭

- **WHEN** 宿主运行在 Windows 平台
- **THEN** 系统 SHALL 视 `cron.shell.enabled` 为 false
- **AND** UI 在页面上显示"当前平台不支持 shell 模式"

### Requirement: Shell 任务独立权限点

系统 SHALL 为 Shell 任务的管理与触发提供独立权限点 `system:job:shell`,与普通任务 CRUD 权限分离。

#### Scenario: 创建 Shell 任务权限

- **WHEN** 用户不具备 `system:job:shell` 权限而尝试创建 `task_type=shell` 任务
- **THEN** 系统 SHALL 返回无权限错误
- **AND** UI 应在任务类型选择中隐藏 shell 选项

#### Scenario: 修改/手动触发 Shell 任务权限

- **WHEN** 用户不具备 `system:job:shell` 权限而尝试修改或手动触发已存在 Shell 任务
- **THEN** 系统 SHALL 拒绝操作

#### Scenario: 默认权限分配

- **WHEN** 系统初始化 seed
- **THEN** `system:job:shell` 权限 SHALL 仅默认分配给内置 `admin` 角色

### Requirement: Shell 执行上下文

系统 SHALL 为 Shell 任务提供 `shell_cmd / work_dir / env / timeout_seconds` 四个维度的执行上下文,并保证行为可预测。

#### Scenario: 环境变量存储边界

- **WHEN** Shell 任务持久化 `env`
- **THEN** 系统 SHALL 在 `sys_job.env` 中以明文 JSON 保存该 KV 集合
- **AND** UI 在编辑既有 Shell 任务时 SHALL 对已保存值做遮罩显示
- **AND** 审计日志 SHALL 不记录原始 `env` 载荷

#### Scenario: 默认 Shell 解释器

- **WHEN** 系统执行 Shell 任务
- **THEN** 系统 SHALL 固定使用 `/bin/sh -c <shell_cmd>` 作为启动命令
- **AND** 支持多行脚本(UI 使用多行文本框录入)

#### Scenario: 工作目录

- **WHEN** 任务 `work_dir` 非空
- **THEN** 系统 SHALL 在启动子进程前校验该目录存在且进程有权限访问
- **AND** 子进程以该目录作为 CWD
- **AND** `work_dir` 为空时使用宿主进程当前 CWD

#### Scenario: 环境变量

- **WHEN** 任务 `env` 非空(KV JSON)
- **THEN** 系统 SHALL 将宿主进程环境变量作为基底
- **AND** 任务级 `env` 覆盖同名进程级变量
- **AND** 最终合并结果传入子进程

#### Scenario: 超时必填

- **WHEN** 创建或修改 Shell 任务
- **THEN** 系统 SHALL 要求 `timeout_seconds ∈ [1, 86400]`
- **AND** 缺失或越界时拒绝保存

### Requirement: Shell 进程生命周期

系统 SHALL 保证 Shell 任务的进程组被正确管理,支持超时与手动终止,且不遗留孤儿进程。

#### Scenario: 独立进程组启动

- **WHEN** 启动 Shell 子进程
- **THEN** 系统 SHALL 在 Unix 平台设置 `SysProcAttr.Setpgid=true`
- **AND** 记录 PGID 用于后续终止

#### Scenario: 超时强制终止

- **WHEN** 执行时间超过 `timeout_seconds`
- **THEN** 系统 SHALL `kill -- -<pgid>` 终止整个进程组
- **AND** 日志 `status=timeout`
- **AND** `err_msg` 记录超时时长

#### Scenario: 手动终止

- **WHEN** 管理员调用 `POST /job/log/{logId}/cancel` 终止 running 的 Shell 实例
- **THEN** 系统 SHALL 向进程组发送 SIGTERM,若在 5 秒内未退出则发送 SIGKILL
- **AND** 日志 `status=cancelled`

#### Scenario: 正常退出

- **WHEN** Shell 子进程正常结束
- **THEN** 系统 SHALL 根据 `exit_code` 判定结果:0 = `success`,非 0 = `failed`
- **AND** `err_msg` 包含 stderr 尾部摘要(失败时)

### Requirement: Shell 输出捕获与截断

系统 SHALL 捕获 Shell 子进程的 stdout 与 stderr,并在超出上限时截断保留前缀。

#### Scenario: 截断策略

- **WHEN** 子进程产生 stdout 或 stderr
- **THEN** 系统 SHALL 各自保留前 64KB
- **AND** 超出部分被丢弃并在尾部追加 `...[truncated]` 标记

#### Scenario: 日志存储

- **WHEN** Shell 任务执行结束
- **THEN** 系统 SHALL 将 `stdout / stderr / exit_code` 写入 `sys_job_log.result_json`
- **AND** 不创建额外日志表

### Requirement: Shell 操作审计

系统 SHALL 将 Shell 任务的敏感操作记录到操作日志,支持事后追责。

#### Scenario: 复用宿主审计中间件

- **WHEN** 创建、修改、手动触发或手动终止 Shell 任务通过宿主 HTTP API 执行
- **THEN** 系统 SHALL 复用现有宿主 `OperLog` 中间件写入 `oper_log`
- **AND** 对同一请求 SHALL 仅写入一条语义等价的审计记录
- **AND** 实现 SHALL NOT 额外手写第二条重复的 `oper_log` 记录

#### Scenario: 创建/修改审计

- **WHEN** 创建或修改 Shell 任务
- **THEN** 系统 SHALL 在 `oper_log` 写入一条记录
- **AND** 记录操作人、IP、操作类型、shell_cmd 快照、work_dir、timeout_seconds

#### Scenario: 手动触发审计

- **WHEN** 管理员手动触发 Shell 任务执行
- **THEN** 系统 SHALL 在 `oper_log` 写入触发记录
- **AND** 关联到生成的 `sys_job_log` 记录 ID

#### Scenario: 手动终止审计

- **WHEN** 管理员终止运行中的 Shell 实例
- **THEN** 系统 SHALL 在 `oper_log` 写入终止记录
- **AND** 记录被终止的 log_id 与目标任务 ID
