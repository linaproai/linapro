# Project Setup

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
系统 SHALL 使用 SQLite 作为数据库，通过 GoFrame 内置的 SQLite driver 连接。SQL 语法 MUST 兼容 MySQL。

#### Scenario: SQLite 数据库自动初始化
- **WHEN** 后端服务首次启动
- **THEN** 自动创建 SQLite 数据库文件并执行初始化 SQL

#### Scenario: SQL 语法兼容性
- **WHEN** 编写 SQL schema 和查询
- **THEN** 所有 SQL 语句不使用 SQLite 特有语法，可在 MySQL 上执行

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
