# Design

## Initial Workspace Foundation

LinaPro 初始形态提供一个可登录、可导航、可管理用户的默认管理工作台。前端采用经典管理后台结构：左侧菜单、顶部导航和右侧内容区；未登录用户进入登录页，登录成功后进入工作台首页；路由守卫在缺少或失效 token 时回到登录流程。

初始菜单使用静态系统管理分组和用户管理入口，后续演进中由动态菜单、插件挂载和工作台适配取代。该历史语义只保留为 foundation 背景，不再作为当前菜单契约事实来源。

## Project Setup Baseline

项目基础启动路径以“面向可持续交付的`AI`原生全栈框架”为定位，默认后端服务、前端工作台、数据库初始化和 mock 数据加载共同构成新开发者的首个可运行路径。后续默认数据库切换到 PostgreSQL 后，基础启动要求变为：先准备 PostgreSQL 14+，再运行`make db.init confirm=init`、`make db.mock confirm=mock`和开发服务；SQLite 只作为开发/演示方言，启动时必须提示不得用于生产。

MySQL 支持被彻底移除，`mysql:`链接必须显式失败，不允许静默回退。数据库创建、重建和 SQL 加载由运维初始化命令触发，运行时服务不负责低权限建库兜底。

## Cross-Domain Impacts

- `database-dialect-abstraction`承载 PostgreSQL/SQLite 方言抽象、`pgsql:`/`sqlite:`分发、MySQL 拒绝、错误分类、数据库版本和表元数据查询，当前契约由`openspec/specs/database-dialect-abstraction/spec.md`承载，历史 owner 为`archive/database-engine`。
- `database-bootstrap-commands`承载`init`/`mock`确认门禁、PG 系统库准备、SQLite 文件准备、逐句执行和首错即停，当前契约由`openspec/specs/database-bootstrap-commands/spec.md`承载，历史 owner 为`archive/database-engine`。
- `sql-source-syntax`承载 PostgreSQL 14+ SQL 源子集、双引号列标识符、`IDENTITY`、`COMMENT ON`、独立索引、`ON CONFLICT DO NOTHING`和禁止 PG 高级特性，当前契约由`openspec/specs/sql-source-syntax/spec.md`承载，历史 owner 为`archive/database-engine`。
- `volatile-table-bootstrap`承载`sys_online_session`、`sys_locker`、`sys_kv_cache`持久表与自然过期语义，当前契约由`openspec/specs/volatile-table-bootstrap/spec.md`承载，历史 owner 为`archive/database-engine`。
- `cluster-deployment-mode`承载 SQLite 强制单机、PostgreSQL 可启用集群和相关启动提示，当前契约由`openspec/specs/cluster-deployment-mode/spec.md`承载，历史 owner 为`archive/distributed-infra`。
- `user-auth`和`user-management`承载登录、JWT、认证中间件、用户 CRUD、导入导出和安全约束，当前契约由`openspec/specs/<capability>/spec.md`承载，历史 owner 分别为`archive/user-auth`和`archive/user-management`。
- `readme-localization-governance`承载目录级 README 英文主版与中文镜像规则，当前契约由`openspec/specs/readme-localization-governance/spec.md`承载，历史 owner 为`archive/i18n`。

## Risks And Boundaries

- foundation 历史中的早期静态菜单和用户管理页面只说明项目起步形态，不作为当前菜单、权限或用户管理契约。
- 数据库切换历史仍对启动体验有价值，但完整 SQL 方言和引导细节应读取`database-engine`，避免 foundation 与数据库 owner 重复保存同一事实。
- 当前能力契约以`openspec/specs`为准，foundation 仅用于理解项目基础形态如何演进。
