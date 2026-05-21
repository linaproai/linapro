## Why

LinaPro 的基础设施层存在三类长期负担，共同制约了开发效率、部署灵活性和工程可维护性：

1. **数据库引擎选型不是最优解**：LinaPro 原以 MySQL 为默认与单一 SQL 源方言、SQLite 作为开发演示方言。MySQL 的 DDL 子集（反引号标识符、`ENGINE=`/`CHARSET=`/`COLLATE=` 子句、`UNSIGNED` 类型族、`AUTO_INCREMENT` 关键字、`INSERT IGNORE` 与 `ON DUPLICATE KEY UPDATE`、`ON UPDATE CURRENT_TIMESTAMP` 内联约束）与 ANSI SQL 偏离最大，导致 SQLite 翻译器需要 438 行正则改写并仍存在语义鸿沟。企业用户和外部部署偏好 PostgreSQL，作为面向可持续交付的 AI 原生全栈框架，LinaPro 的默认数据库需要更贴近现代企业生产环境。

2. **启动效率低下**：服务启动时在短时间内输出大量 SQL 调试日志，执行多轮重复的插件注册表、发布快照、菜单和资源引用查询。启动后 10 秒内记录到约 98 条 SQL，前 4 秒约 57 条属于启动或启动后立即触发链路，影响启动日志可读性，也增加了本地开发和演示场景的启动成本。

3. **开发命令不跨平台**：项目开发与交付入口主要依赖 `make`、POSIX Shell 语法以及 `lsof`、`awk`、`sed`、`nohup`、`kill` 等 Linux/macOS 常见工具。Windows 默认不内置 GNU Make，也不保证存在 Bash/POSIX 工具链，导致 Windows 用户即使安装了 Go、Node.js、Docker 等核心依赖，仍难以顺利执行项目常用命令。

本变更将这三个基础设施问题作为统一的基础能力治理对象，通过一次系统性改造建立更清洁的数据库引擎基础、更高效的启动路径和更包容的跨平台开发体验。

## What Changes

### Database Engine Migration

- **新增**：PostgreSQL 14+ 作为默认数据库与单一 SQL 源方言
- **保留**：SQLite 作为开发/演示方言（启动期警告"do not use in production"）
- **BREAKING**：完全移除 MySQL 支持（不再注册 `mysql:` link 前缀，删除 `pkg/dialect/internal/mysql/` 包，移除 GoFrame MySQL 驱动依赖）
- 新增 `pkg/dialect/internal/postgres/` 子包，实现完整的 `Dialect` 接口（`dialect.go`、`translate.go`、`error.go`、`prepare.go`、`metadata.go`）
- 重写 SQLite 翻译器为"PG → SQLite"翻译路径
- 所有宿主与插件 SQL 从 MySQL 语法改写为 PostgreSQL 语法
- 原 MEMORY 表（`sys_online_session`、`sys_locker`、`sys_kv_cache`）改造为普通持久表 + 自然过期语义
- 数据库派生 `uint64` 类型迁移为 `int64`
- 默认配置、本地 PG 启动说明、CI service container 配置和文档全量切换到 PG

### Startup SQL Efficiency

- 调整本地默认配置，使普通启动默认不输出每条 SQL 调试日志；需要排查 SQL 时仍可通过 `database.default.debug=true` 显式开启
- 引入一次启动链路内共享的 `StartupContext`，合并插件启动同步、运行时预热和内置任务同步中的重复快照构造
- 优化插件清单同步的 no-op fast path：清单、发布快照、菜单、权限和资源引用均无变化时，不开启事务、不写库、不做写后回读
- 优化启动期内置定时任务投影和调度注册：复用声明派生快照，避免重复从 `sys_job` 读取同一批内置任务
- 为启动 SQL 统计和关键阶段耗时增加结构化摘要日志
- 增加后端启动 smoke 或单元测试，约束默认配置下不得输出 SQL 明细

### Cross-Platform Dev Commands

- 新增跨平台开发命令能力，以 Go CLI（`hack/tools/linactl`）承载常用任务编排逻辑
- 在项目根目录提供 Windows `make.cmd` 薄包装入口
- 保留现有 `Makefile` 作为兼容层，逐步将复杂目标改为调用跨平台 Go CLI
- 将 `dev`、`stop`、`status`、`build`、`wasm`、`init`、`mock`、`test`、`test-go`、`help`、`cli.install` 等常用目标纳入跨平台入口
- 更新 `.github/workflows/` 增加 Windows runner 基本命令验证
- 更新根目录文档中的命令说明，明确跨平台推荐入口

## Capabilities

### New Capabilities

- `sql-source-syntax`：定义 PostgreSQL 作为单一 SQL 源语法的子集约定，包括允许的语法、禁止的 PG 高级特性、保留字处理、索引/注释拆分规则
- `volatile-table-bootstrap`：定义原 MEMORY 表在 PG/SQLite 下改为普通持久表后的自然过期契约
- `startup-sql-efficiency`：定义宿主启动期 SQL 数量、启动日志噪音、插件启动同步快照复用、no-op 同步路径和启动效率回归测试要求
- `cross-platform-dev-commands`：定义项目跨平台开发命令入口、Windows `make.cmd` 兼容入口、命令参数兼容、外部工具调用边界、测试与文档要求

### Modified Capabilities

- `database-dialect-abstraction`：从"MySQL/SQLite 双方言抽象"修改为"PostgreSQL/SQLite 双方言抽象"。新增 PostgreSQL 方言相关需求与场景；新增 `Dialect.QueryTableMetadata` 接口方法
- `database-bootstrap-commands`：修改 `PrepareDatabase` 方言分发场景（MySQL 替换为 PostgreSQL）；修改 `TranslateDDL` 方言分发场景
- `cluster-deployment-mode`：修改方言枚举（MySQL → PostgreSQL）；保持"SQLite 启动期自动锁定 `cluster.enabled=false`"语义
- `project-setup`：修改"项目使用 SQLite 作为数据库"为"项目使用 PostgreSQL 作为默认数据库，SQLite 作为开发演示方言"
- `plugin-startup-bootstrap`：启动引导阶段必须复用同一轮插件治理启动快照
- `plugin-manifest-lifecycle`：插件清单同步在无差异时必须保持无副作用
- `cron-job-management`：内置定时任务启动投影必须使用声明派生快照注册

## Impact

### Database Engine Migration Impact

- **方言层**：`apps/lina-core/pkg/dialect/` 全面改造（删除 `internal/mysql/`，新建 `internal/postgres/`，重写 `internal/sqlite/translate.go`）
- **SQL 源**：宿主 12 个 SQL 文件 + mock-data + 8 个包含实际 SQL 资源的插件 SQL 全量改写为 PG 语法
- **配置 / 工具链**：`config.yaml`、`config.template.yaml`、`hack/config.yaml`、`go.mod`、`main.go` 切换到 PG
- **业务代码**：数据库派生 `uint64` → `int64`；`plugin_data_table_comment.go` 重构为 `dialect.QueryTableMetadata` 调用
- **容器 / CI**：本地 PG 启动配置、GitHub Actions `services.postgres`、Dockerfile 清理
- **文档**：README / README.zh-CN / CLAUDE.md 技术栈描述更新

### Startup SQL Efficiency Impact

- **后端代码**：`cmd_http.go`、`cmd_http_runtime.go`、`cmd_http_routes.go`、`service/plugin/`、`service/cron/`、`service/jobmgmt/`
- **配置**：`config.yaml` 的 `database.default.debug` 默认值调整为 `false`
- **测试**：新增启动 SQL 统计/日志 smoke 测试、插件同步 no-op fast path 测试、内置任务启动注册去重测试

### Cross-Platform Dev Commands Impact

- **开发工具**：新增 `hack/tools/linactl` Go CLI、根目录 `make.cmd`
- **Makefile**：根目录与子模块 Makefile 收敛为薄包装
- **CI**：GitHub Actions 增加 Windows runner 验证
- **文档**：根目录 README/README.zh-CN 跨平台命令说明

### API / Data / Dependency Impact

- **HTTP API**：无变化（数据库类型不暴露到对外接口）
- **i18n**：有限影响。系统信息、framework 运行时文案和 apidoc 元数据中暴露的数据库名称/技术栈描述需要从 MySQL 更新为 PostgreSQL
- **缓存一致性**：MEMORY 表改持久表后，不新增启动期清空或跨实例协调策略；`sys_online_session`/`sys_locker`/`sys_kv_cache` 分别依赖既有 TTL 清理路径自然收敛。启动快照仅限单次启动调用链内使用，不跨请求、不跨进程、不作为持久缓存
- **数据迁移**：不提供数据迁移工具，全新项目重建数据库即可
- **回退**：`git revert` 即可，无数据风险
