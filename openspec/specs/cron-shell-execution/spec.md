# 定时任务 Shell 执行规范

## 目的
待定 - 由归档变更 scheduled-job-management 创建。归档后更新目的。

## 需求
### 需求：Shell 任务全局开关

系统 SHALL 通过系统参数 `cron.shell.enabled` 控制是否允许 Shell 类型任务的创建、修改与执行，默认关闭。

#### 场景：开关关闭时拒绝创建

- **当** `cron.shell.enabled=false` 且用户尝试创建 `task_type=shell` 任务时
- **则** 系统 SHALL 拒绝创建并返回明确错误

#### 场景：开关关闭时拒绝修改

- **当** `cron.shell.enabled=false` 且用户尝试修改已存在 Shell 任务的任一字段时
- **则** 系统 SHALL 拒绝修改

#### 场景：开关关闭时拒绝执行

- **当** `cron.shell.enabled=false` 且 Shell 任务到达 tick 或被手动触发时
- **则** 系统 SHALL 拒绝本次执行
- **且** 日志 `status=failed`、`err_msg` 指明 shell 开关已关闭

#### 场景：UI 隐藏 shell 选项

- **当** 前端从系统参数接口读取到 `cron.shell.enabled=false` 时
- **则** UI SHALL 在任务类型选择中隐藏 `shell` 选项
- **且** 对已存在的 shell 任务仅允许只读查看

#### 场景：Windows 运行时强制关闭

- **当** 宿主运行在 Windows 平台时
- **则** 系统 SHALL 视 `cron.shell.enabled` 为 false
- **且** UI 在页面上显示"当前平台不支持 shell 模式"

### 需求：Shell 任务独立权限点

系统 SHALL 为 Shell 任务的管理与触发提供独立权限点 `system:job:shell`，与普通任务 CRUD 权限分离。

#### 场景：创建 Shell 任务权限

- **当** 用户不具备 `system:job:shell` 权限而尝试创建 `task_type=shell` 任务时
- **则** 系统 SHALL 返回无权限错误
- **且** UI 应在任务类型选择中隐藏 shell 选项

#### 场景：修改/手动触发 Shell 任务权限

- **当** 用户不具备 `system:job:shell` 权限而尝试修改或手动触发已存在 Shell 任务时
- **则** 系统 SHALL 拒绝操作

#### 场景：手动终止 Shell 任务权限

- **当** 用户不具备 `system:job:shell` 权限而尝试终止 running 的 Shell 实例时
- **则** 系统 SHALL 拒绝操作

#### 场景：默认权限分配

- **当** 系统初始化 seed 时
- **则** `system:job:shell` 权限 SHALL 仅默认分配给内置 `admin` 角色

### 需求：Shell 执行上下文

系统 SHALL 为 Shell 任务提供 `shell_cmd / work_dir / env / timeout_seconds` 四个维度的执行上下文，并保证行为可预测。

#### 场景：环境变量存储边界

- **当** Shell 任务持久化 `env` 时
- **则** 系统 SHALL 在 `sys_job.env` 中以明文 JSON 保存该 KV 集合
- **且** UI 在编辑既有 Shell 任务时 SHALL 对已保存值做遮罩显示
- **且** 审计日志 SHALL 不记录原始 `env` 载荷

#### 场景：默认 Shell 解释器

- **当** 系统执行 Shell 任务时
- **则** 系统 SHALL 固定使用 `/bin/sh -c <shell_cmd>` 作为启动命令
- **且** 支持多行脚本（UI 使用多行文本框录入）

#### 场景：工作目录

- **当** 任务 `work_dir` 非空时
- **则** 系统 SHALL 在启动子进程前校验该目录存在且进程有权限访问
- **且** 子进程以该目录作为 CWD
- **且** `work_dir` 为空时使用宿主进程当前 CWD

#### 场景：环境变量

- **当** 任务 `env` 非空（KV JSON）时
- **则** 系统 SHALL 将宿主进程环境变量作为基底
- **且** 任务级 `env` 覆盖同名进程级变量
- **且** 最终合并结果传入子进程

#### 场景：超时必填

- **当** 创建或修改 Shell 任务时
- **则** 系统 SHALL 要求 `timeout_seconds ∈ [1, 86400]`
- **且** 缺失或越界时拒绝保存

### 需求：Shell 进程生命周期

系统 SHALL 保证 Shell 任务的进程组被正确管理，支持超时与手动终止，且不遗留孤儿进程。

#### 场景：独立进程组启动

- **当** 启动 Shell 子进程时
- **则** 系统 SHALL 在 Unix 平台设置 `SysProcAttr.Setpgid=true`
- **且** 记录 PGID 用于后续终止

#### 场景：超时强制终止

- **当** 执行时间超过 `timeout_seconds` 时
- **则** 系统 SHALL `kill -- -<pgid>` 终止整个进程组
- **且** 日志 `status=timeout`
- **且** `err_msg` 记录超时时长

#### 场景：手动终止

- **当** 管理员调用 `POST /job/log/{logId}/cancel` 终止 running 的 Shell 实例时
- **则** 系统 SHALL 向进程组发送 SIGTERM，若在 5 秒内未退出则发送 SIGKILL
- **且** 日志 `status=cancelled`

#### 场景：正常退出

- **当** Shell 子进程正常结束时
- **则** 系统 SHALL 根据 `exit_code` 判定结果：0 = `success`，非 0 = `failed`
- **且** `err_msg` 包含 stderr 尾部摘要（失败时）

### 需求：Shell 输出捕获与截断

系统 SHALL 捕获 Shell 子进程的 stdout 与 stderr，并在超出上限时截断保留前缀。

#### 场景：截断策略

- **当** 子进程产生 stdout 或 stderr 时
- **则** 系统 SHALL 各自保留前 64KB
- **且** 超出部分被丢弃并在尾部追加 `...[truncated]` 标记

#### 场景：日志存储

- **当** Shell 任务执行结束时
- **则** 系统 SHALL 将 `stdout / stderr / exit_code` 写入 `sys_job_log.result_json`
- **且** 不创建额外日志表

### 需求：Shell 操作审计

系统 SHALL 将 Shell 任务的敏感操作记录到操作日志，支持事后追责。

#### 场景：复用宿主审计中间件

- **当** 创建、修改、手动触发或手动终止 Shell 任务通过宿主 HTTP API 执行时
- **则** 系统 SHALL 复用现有宿主 `OperLog` 中间件写入 `oper_log`
- **且** 对同一请求 SHALL 仅写入一条语义等价的审计记录
- **且** 实现 SHALL NOT 额外手写第二条重复的 `oper_log` 记录

#### 场景：创建/修改审计

- **当** 创建或修改 Shell 任务时
- **则** 系统 SHALL 在 `oper_log` 写入一条记录
- **且** 记录操作人、IP、操作类型、shell_cmd 快照、work_dir、timeout_seconds
- **且** 不记录原始 `env` 值

#### 场景：手动触发审计

- **当** 管理员手动触发 Shell 任务执行时
- **则** 系统 SHALL 在 `oper_log` 写入触发记录
- **且** 关联到生成的 `sys_job_log` 记录 ID

#### 场景：手动终止审计

- **当** 管理员终止运行中的 Shell 实例时
- **则** 系统 SHALL 在 `oper_log` 写入终止记录
- **且** 记录被终止的 log_id 与目标任务 ID
- **且** 不记录原始 `env` 值
