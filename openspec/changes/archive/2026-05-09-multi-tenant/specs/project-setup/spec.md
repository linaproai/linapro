## MODIFIED Requirements

### Requirement: 开发环境一键启动
系统 SHALL 提供 Makefile 命令,支持一键启动前后端开发环境。`make stop` 与由 `make dev` 触发的停止阶段 MUST 只终止能够证明为当前仓库 LinaPro 开发服务的进程,不得仅凭 `temp/pids/*.pid` 中的 PID 或端口占用结果终止进程。

#### Scenario: 启动开发环境
- **WHEN** 在项目根目录执行 `make dev`
- **THEN** 前端和后端服务同时启动

#### Scenario: 停止开发环境
- **WHEN** 在项目根目录执行 `make stop`
- **THEN** 前端和后端服务同时停止

#### Scenario: 陈旧 PID 指向非 LinaPro 进程
- **WHEN** `temp/pids/*.pid` 中的 PID 当前存在但进程身份不匹配当前仓库的 LinaPro 后端或前端开发服务
- **THEN** `make stop` SHALL 跳过该进程
- **AND** 清理对应的陈旧 PID 文件

#### Scenario: 开发端口被非 LinaPro 进程占用
- **WHEN** 默认开发端口被其他进程监听
- **THEN** `make stop` SHALL 不终止该进程
- **AND** `make dev` SHALL 在后续启动阶段由端口冲突自然失败并提示查看服务日志
