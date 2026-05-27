# 项目初始化规范

## Purpose

定义项目初始化、前后端启动、数据库配置与开发环境代理等基础能力，确保仓库在本地开发和后续扩展过程中具备稳定的一致性基线。
## Requirements
### Requirement: 后端项目初始化
系统 SHALL 提供基于 GoFrame v2 框架的后端项目，项目结构遵循 GoFrame 标准分层架构（api / controller / service / dao / model）。

#### Scenario: 后端项目可编译运行
- **WHEN** 在 `apps/lina-core/` 目录下执行 `go build` 或 `make build`
- **THEN** 项目成功编译为可执行文件

#### Scenario: 后端服务启动并监听端口
- **WHEN** 启动后端服务
- **THEN** 服务在配置的端口（默认 9120）上监听 HTTP 请求

### Requirement: 前端项目初始化
系统 SHALL 提供基于 Vben5 最新版 + Ant Design Vue 的前端项目，使用 pnpm monorepo 结构。

#### Scenario: 前端项目可构建
- **WHEN** 在 `apps/lina-vben/` 目录下执行 `pnpm install && pnpm build`
- **THEN** 项目成功构建产出 dist 产物

#### Scenario: 前端开发服务器启动
- **WHEN** 启动前端开发服务器
- **THEN** 服务在配置的端口上启动，可通过浏览器访问

### Requirement: 数据库配置

系统 SHALL 使用 PostgreSQL 14+ 作为唯一运行时数据库，通过 GoFrame 官方 PG 驱动 `gogf/gf/contrib/drivers/pgsql/v2` 连接。系统 MUST NOT 支持 SQLite、MySQL 或其他数据库作为运行时数据库。所有 SQL 源文件 MUST 使用 PostgreSQL 14+ 语法编写，并在 PostgreSQL 数据库上直接执行。PostgreSQL 默认路径 SHALL 使用数据库默认 deterministic collation，不创建或依赖自定义排序规则；业务文本键默认大小写敏感。

#### Scenario: PostgreSQL 默认数据库连接

- **WHEN** 后端服务启动且 `database.default.link` 以 `pgsql:` 开头
- **THEN** 后端通过 GoFrame PG 驱动连接到 PostgreSQL 数据库
- **AND** 服务启动不创建、删除或重建数据库
- **AND** 数据库创建、重建和 SQL 加载仅由 `make init confirm=init` / `make init confirm=init rebuild=true` 等运维初始化命令触发
- **AND** 业务文本键的唯一约束和等值匹配按 PostgreSQL 默认大小写敏感语义工作

#### Scenario: SQLite 链接被显式拒绝

- **WHEN** 配置文件 `database.default.link` 以 `sqlite:` 开头
- **THEN** 后端启动失败并返回明确错误
- **AND** 错误消息说明 SQLite 不再支持，并列出当前支持的方言仅为 `pgsql:`
- **AND** 不静默回退到任何默认方言

#### Scenario: MySQL 链接被显式拒绝

- **WHEN** 配置文件 `database.default.link` 以 `mysql:` 开头
- **THEN** 后端启动失败并返回明确错误
- **AND** 错误消息说明 MySQL 不再支持，并列出当前支持的方言仅为 `pgsql:`
- **AND** 不静默回退到任何默认方言

#### Scenario: SQL 语法兼容性

- **WHEN** 编写 SQL schema 和查询
- **THEN** 所有 SQL 语句 MUST 使用 PostgreSQL 14+ 语法
- **AND** MUST NOT 包含 MySQL 特有语法（AUTO_INCREMENT、UNSIGNED、ENGINE=、INSERT IGNORE、ON DUPLICATE KEY UPDATE 等）
- **AND** MUST NOT 包含 SQLite 特有语法或依赖 SQLite 文件数据库行为
- **AND** MUST NOT 创建或依赖自定义 collation；需要大小写不敏感语义的具体字段必须单独通过 OpenSpec 设计

### Requirement: API 代理配置
前端开发环境 SHALL 配置 API 代理，将 `/api` 前缀的请求转发到后端服务。

#### Scenario: API 请求代理
- **WHEN** 前端发起 `/api/v1/*` 请求
- **THEN** 请求被代理到后端服务地址（默认 `http://localhost:9120`）

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
- **THEN** 系统 SHALL 安装或确认前端依赖，并安装 Playwright Chromium headless shell 及其所需系统依赖
- **AND** 该命令 SHALL 承接原 `make dev.setup` 的功能语义
- **AND** 该命令 SHALL 面向默认 headless E2E 测试初始化，不要求同时安装完整 Chromium 浏览器

#### Scenario: 完整浏览器调试
- **WHEN** 开发者需要运行 headed、UI 或 debug 模式的 Playwright 测试
- **THEN** 系统 SHALL 允许开发者通过 Playwright 原生命令额外安装完整 Chromium 浏览器
- **AND** `make env.setup` SHALL 不因该调试场景改变默认 headless shell 安装语义

#### Scenario: 旧初始化入口已移除
- **WHEN** 开发者执行 `make dev.setup` 或 `linactl dev.setup`
- **THEN** 系统 SHALL 不再将其作为受支持开发环境初始化入口
- **AND** 帮助信息和修复提示 SHALL 指向 `make env.setup`

### Requirement: 开发环境必须支持可配置管理工作台入口

系统 SHALL 在本地开发环境中支持默认管理工作台入口路径配置。前端开发服务器、后端代理、登录跳转、页面刷新和测试入口必须使用该配置后的工作台路径，而不是假定默认管理工作台部署在固定路径。开发模式 SHALL 以后端公开地址作为统一访问入口，后端在工作台入口路径下代理或转发到 Vite dev server；当工作台入口不是 `/` 时，Vite dev server 不得作为根路径公共入口吞掉源码插件公开路由。

#### Scenario: 开发环境访问管理工作台
- **WHEN** 工作台入口路径配置为 `/admin`
- **AND** 开发者启动本地开发环境
- **THEN** 开发者可以通过后端公开地址的 `/admin` 访问默认管理工作台
- **AND** 管理工作台内部页面刷新仍能返回工作台应用
- **AND** 后端在 `/admin` 及其子路径下代理或转发到 Vite dev server

#### Scenario: 开发代理不吞掉源码插件根路由
- **WHEN** 工作台入口路径配置为 `/admin`
- **AND** 源码插件在后端注册 `/` 路由
- **THEN** 开发环境访问后端公开地址的 `/` 时请求到达源码插件后端路由
- **AND** 前端开发服务器不得将 `/` 强制重写为管理工作台应用

#### Scenario: API 代理保持原有语义
- **WHEN** 工作台入口路径配置发生变化
- **THEN** `/api/v1/*` 请求仍代理到后端宿主 API
- **AND** `/x/*` 统一插件 API 请求仍代理到后端，并由后端按源码插件 GoFrame handler 或动态插件 runtime dispatcher 处理

#### Scenario: Vite 不作为公共根入口
- **WHEN** 开发模式启动 Vite dev server
- **THEN** Vite 只承接后端转发过来的管理工作台 `basePath` 请求或内部资源请求
- **AND** 开发文档、命令输出和 E2E 配置必须把后端公开地址作为默认访问入口
- **AND** 直接访问 Vite 根路径不得作为验证源码插件公开路由可用性的依据

### Requirement: 测试入口必须区分管理工作台路径与源码插件公开路径

系统 SHALL 在 E2E 和开发测试配置中区分管理工作台访问路径与后端公共访问根路径。测试不得再把 `E2E_BASE_URL` 的根路径等同于默认管理工作台首页。

#### Scenario: E2E 使用工作台入口登录
- **WHEN** E2E 需要访问默认管理工作台登录页
- **THEN** 测试必须使用配置后的工作台入口路径解析登录地址
- **AND** 不得硬编码根路径 `/` 作为管理工作台首页

#### Scenario: E2E 验证源码插件公开路由
- **WHEN** E2E 需要验证源码插件注册的公开 HTTP 路由
- **THEN** 测试必须直接访问后端公共路径
- **AND** 不得通过管理工作台路由断言插件公开页面是否可用

#### Scenario: E2E 使用统一插件 API 前缀
- **WHEN** E2E 需要验证插件 API 路径
- **THEN** 源码插件 API 用例 MUST 访问 `/x/{plugin-id}/api/v1/...`
- **AND** 动态插件 API 用例 MUST 访问 `/x/{plugin-id}/api/v1/...`
- **AND** 宿主控制面 API 用例继续访问 `/api/v1/...`

