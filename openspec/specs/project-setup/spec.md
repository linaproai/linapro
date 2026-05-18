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
- **THEN** 服务在配置的端口（默认 8080）上监听 HTTP 请求

### Requirement: 前端项目初始化
系统 SHALL 提供基于 Vben5 最新版 + Ant Design Vue 的前端项目，使用 pnpm monorepo 结构。

#### Scenario: 前端项目可构建
- **WHEN** 在 `apps/lina-vben/` 目录下执行 `pnpm install && pnpm build`
- **THEN** 项目成功构建产出 dist 产物

#### Scenario: 前端开发服务器启动
- **WHEN** 启动前端开发服务器
- **THEN** 服务在配置的端口上启动，可通过浏览器访问

### Requirement: 数据库配置

系统 SHALL 使用 PostgreSQL 14+ 作为默认数据库，通过 GoFrame 官方 PG 驱动 `gogf/gf/contrib/drivers/pgsql/v2` 连接。系统 SHALL 同时支持 SQLite 作为开发/演示方言，通过 GoFrame 内置的 SQLite driver 连接。系统 MUST NOT 支持 MySQL 作为运行时数据库。所有 SQL 源文件 MUST 使用 PostgreSQL 14+ 语法子集编写，可被 SQLite 方言翻译执行。PostgreSQL 默认路径 SHALL 使用数据库默认 deterministic collation，不创建或依赖自定义排序规则；业务文本键默认大小写敏感。

#### Scenario: PostgreSQL 默认数据库连接
- **WHEN** 后端服务启动且 `database.default.link` 以 `pgsql:` 开头
- **THEN** 后端通过 GoFrame PG 驱动连接到 PostgreSQL 数据库
- **AND** 服务启动不创建、删除或重建数据库
- **AND** 数据库创建、重建和 SQL 加载仅由 `make init confirm=init` / `make init confirm=init rebuild=true` 等运维初始化命令触发
- **AND** 业务文本键的唯一约束和等值匹配按 PostgreSQL 默认大小写敏感语义工作

#### Scenario: SQLite 开发演示模式
- **WHEN** 开发者将 `database.default.link` 改为 `sqlite::@file(./temp/sqlite/linapro.db)`
- **THEN** 后端通过 GoFrame SQLite 驱动连接到 SQLite 数据库文件
- **AND** SQLite 数据库文件父目录由 `make init` 的方言准备逻辑创建
- **AND** 启动日志输出"不得用于生产"的明确警示

#### Scenario: MySQL 链接被显式拒绝
- **WHEN** 配置文件 `database.default.link` 以 `mysql:` 开头
- **THEN** 后端启动失败并返回明确错误
- **AND** 错误消息说明 MySQL 不再支持，并列出当前支持的方言（`pgsql:`、`sqlite:`）
- **AND** 不静默回退到任何默认方言

#### Scenario: SQL 语法兼容性
- **WHEN** 编写 SQL schema 和查询
- **THEN** 所有 SQL 语句 MUST 使用 PostgreSQL 14+ 语法子集
- **AND** 不使用 PG 高级特性（JSONB、CREATE TRIGGER、CREATE FUNCTION、SERIAL/BIGSERIAL 简写、MERGE、WITH RECURSIVE 等）
- **AND** 可通过方言层 `TranslateDDL` 翻译为 SQLite 兼容语句
- **AND** MUST NOT 包含 MySQL 特有语法（AUTO_INCREMENT、UNSIGNED、ENGINE=、INSERT IGNORE、ON DUPLICATE KEY UPDATE 等）
- **AND** MUST NOT 创建或依赖自定义 collation；需要大小写不敏感语义的具体字段必须单独通过 OpenSpec 设计

### Requirement: API 代理配置
前端开发环境 SHALL 配置 API 代理，将 `/api` 前缀的请求转发到后端服务。

#### Scenario: API 请求代理
- **WHEN** 前端发起 `/api/v1/*` 请求
- **THEN** 请求被代理到后端服务地址（默认 `http://localhost:8080`）

### Requirement: 开发环境一键启动
系统 SHALL 提供 Makefile 命令，支持一键启动前后端开发环境。

#### Scenario: 启动开发环境
- **WHEN** 在项目根目录执行 `make dev`
- **THEN** 前端和后端服务同时启动

#### Scenario: 停止开发环境
- **WHEN** 在项目根目录执行 `make stop`
- **THEN** 前端和后端服务同时停止

