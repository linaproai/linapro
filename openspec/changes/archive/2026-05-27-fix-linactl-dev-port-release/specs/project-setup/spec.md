## MODIFIED Requirements

### Requirement: 开发环境一键启动
系统 SHALL 提供 Makefile 命令，支持一键启动前后端开发环境，并提供独立的本地开发环境检查与初始化入口。`make dev`在重启由`linactl`管理的旧后端或前端服务时，SHALL 等待被停止服务的监听端口在有限时间内释放；如果端口仍被未知外部进程占用，系统 SHALL 拒绝自动终止未知进程，并提示开发者停止占用者或通过`BACKEND_PORT`、`FRONTEND_PORT`选择其他端口。

#### Scenario: 启动开发环境
- **WHEN** 在项目根目录执行`make dev`
- **THEN** 前端和后端服务同时启动

#### Scenario: 重启已管理开发环境
- **WHEN** 在项目根目录执行`make dev`
- **AND** 旧后端或前端服务由`linactl`通过`PID`文件管理
- **AND** 旧服务收到停止请求后短暂占用原监听端口
- **THEN** 系统 SHALL 等待端口释放后继续构建并启动新的后端和前端服务

#### Scenario: 拒绝未知端口占用者
- **WHEN** 在项目根目录执行`make dev`
- **AND** 后端或前端端口被非`linactl`管理的未知进程占用
- **THEN** 系统 SHALL 不自动终止未知进程
- **AND** 命令 SHALL 返回端口占用错误，并提示停止占用者或通过`BACKEND_PORT`、`FRONTEND_PORT`选择其他端口

#### Scenario: 停止开发环境
- **WHEN** 在项目根目录执行`make stop`
- **THEN** 前端和后端服务同时停止

#### Scenario: 检查开发环境
- **WHEN** 在项目根目录执行`make env.check`
- **THEN** 系统 SHALL 以表格展示本地 Go、Node.js、pnpm、Vite、Playwright 和 PostgreSQL 的名称、当前版本、要求版本、是否满足和备注
- **AND** 该检查 SHALL 不启动开发服务、不连接业务数据库、不修改本地依赖或配置文件

#### Scenario: 初始化开发环境
- **WHEN** 在项目根目录执行`make env.setup`
- **THEN** 系统 SHALL 安装或确认前端依赖，并安装 Playwright Chromium 浏览器及其所需系统依赖
- **AND** 该命令 SHALL 承接原`make dev.setup`的功能语义

#### Scenario: 旧初始化入口已移除
- **WHEN** 开发者执行`make dev.setup`或`linactl dev.setup`
- **THEN** 系统 SHALL 不再将其作为受支持开发环境初始化入口
- **AND** 帮助信息和修复提示 SHALL 指向`make env.setup`
