## MODIFIED Requirements

### Requirement: 开发环境一键启动
系统 SHALL 提供 Makefile 命令，支持一键启动前后端开发环境，并提供独立的本地开发环境检查与初始化入口。

#### Scenario: 启动开发环境
- **WHEN** 在项目根目录执行 `make dev`
- **THEN** 前端和后端服务同时启动

#### Scenario: 停止开发环境
- **WHEN** 在项目根目录执行 `make stop`
- **THEN** 前端和后端服务同时停止

#### Scenario: 检查开发环境
- **WHEN** 在项目根目录执行 `make env.check`
- **THEN** 系统 SHALL 以表格展示本地 Go、Node.js、pnpm、Vite、Playwright 和 PostgreSQL 的名称、当前版本、要求版本、是否满足和备注
- **AND** 该检查 SHALL 不启动开发服务、不连接业务数据库、不修改本地依赖或配置文件

#### Scenario: 初始化开发环境
- **WHEN** 在项目根目录执行 `make env.setup`
- **THEN** 系统 SHALL 安装或确认前端依赖，并安装 Playwright Chromium 浏览器及其所需系统依赖
- **AND** 该命令 SHALL 承接原 `make dev.setup` 的功能语义

#### Scenario: 旧初始化入口已移除
- **WHEN** 开发者执行 `make dev.setup` 或 `linactl dev.setup`
- **THEN** 系统 SHALL 不再将其作为受支持开发环境初始化入口
- **AND** 帮助信息和修复提示 SHALL 指向 `make env.setup`
