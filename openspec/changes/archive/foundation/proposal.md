## Why

foundation 分组记录 LinaPro 从初始项目脚手架进入可运行全栈框架的基础形态：后端、前端、默认管理工作台布局、登录页、用户管理入口、路由守卫和项目启动路径。后续数据库默认从 MySQL 切换到 PostgreSQL 的历史也曾落入本分组，因为它影响了 README、初始化命令、默认配置和本地启动体验。

当前归档治理中，数据库方言、SQL 源、易失性表、集群部署和 README 镜像已经分别有更明确的 owner。foundation 只保留框架初始布局、项目启动定位和数据库迁移对基础启动路径的摘要，避免继续保存数据库引擎完整规范全文。

## What Changes

- 建立默认管理工作台的初始结构：登录页、侧边栏、顶部导航栏、内容区、用户信息、退出登录、路由守卫和用户管理基础页面。
- 明确项目基础启动从“全新项目脚手架”演进为“PostgreSQL 默认数据库 + SQLite 开发演示”的框架启动路径。
- 记录数据库切换对 foundation 的影响：README 启动前置、默认数据库配置、`make db.init`/`make db.mock`门禁、SQLite 开发演示和 MySQL 显式拒绝。
- 将数据库方言抽象、SQL 源语法、易失性表自然过期、集群部署和 README 镜像的完整历史交由对应 owner 分组承载。

## Capabilities

### New Capabilities

- `base-layout`

### Modified Capabilities

- `project-setup`

## Impact

- 影响初始工作台布局、登录/退出体验、用户管理入口和项目快速启动说明。
- 交叉影响数据库引擎、SQL 初始化、集群模式、认证、用户管理和 README 镜像治理；这些能力的当前契约由`openspec/specs`和对应 owner 分组承载。
- 不再在 foundation 分组长期保留 PostgreSQL 切换的完整 SQL、方言和数据库引导规范全文。
