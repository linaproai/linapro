## Why

Lina 管理后台系统需要从零开始搭建。作为第一个迭代（Change 1），需要完成项目基础架构初始化、用户认证体系和用户管理功能，为后续所有业务模块提供可运行的基础平台。

## What Changes

- 初始化前端项目：基于 Vben5 最新版官方模板，使用 Vue 3 + Ant Design Vue 技术栈
- 初始化后端项目：基于 GoFrame v2 框架，配置 SQLite 数据库（兼容 MySQL 语法）
- 实现 JWT 认证体系：用户名 + 密码登录、登出、Token 刷新
- 实现用户管理模块：用户 CRUD（轻量用户表，仅包含基础字段）
- 搭建管理后台基础布局：侧边栏菜单框架、顶部导航
- E2E 测试覆盖：登录/登出流程、用户管理 CRUD 操作

## Capabilities

### New Capabilities

- `project-setup`: 前后端项目初始化、目录结构、构建配置、开发环境搭建
- `user-auth`: JWT 认证体系，包含登录、登出、Token 管理、认证中间件
- `user-management`: 用户管理 CRUD，轻量用户表设计（账号、密码、昵称、状态、时间戳）
- `base-layout`: 管理后台基础布局，侧边栏菜单框架、顶部导航栏、路由配置

### Modified Capabilities

（无，全新项目）

## Impact

- 新增前端项目 `apps/lina-vben/`（Vben5 monorepo 结构）
- 新增后端项目 `apps/lina-core/`（GoFrame v2 标准结构）
- 新增数据库表：`sys_user`
- 新增 API 端点：认证相关（登录/登出）、用户 CRUD
- 新增 E2E 测试：Playwright 测试用例
- 依赖：GoFrame v2、SQLite driver、JWT 库、Vben5、Ant Design Vue、Playwright
