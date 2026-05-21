## Why

当前仓库已经把 PostgreSQL 作为默认数据库，但仍保留 SQLite 方言、驱动、启动降级、CI smoke 和 E2E 通道，导致数据库支持矩阵复杂、测试成本高，并持续限制 SQL 源只能使用 SQLite 可翻译子集。

本次变更将运行时数据库支持收敛到 PostgreSQL 14+，移除 SQLite 作为开发、演示和测试目标的支持，降低后端、插件生命周期、CI 和文档维护成本。

## What Changes

- **BREAKING**：`database.default.link` 不再支持 `sqlite:` 前缀；使用 SQLite 链接时启动、初始化和 mock 加载必须明确失败。
- **BREAKING**：移除 GoFrame SQLite 驱动注册、SQLite 方言实现、SQLite DDL 转译器和 SQLite 专属启动钩子。
- **BREAKING**：移除 SQLite 专属后端测试、E2E 用例、CI smoke workflow 和 `pnpm test:sqlite*` 入口。
- 将默认镜像运行配置、配置模板、README、测试文档和现行 OpenSpec 基线统一更新为 PostgreSQL-only 支持口径。
- 保留 PostgreSQL 方言抽象作为数据库准备、SQL 执行、版本查询、表元数据查询和错误分类的统一边界。

## Capabilities

### New Capabilities

- `postgresql-only-database-support`：定义 LinaPro 运行时、初始化、测试和交付链路只支持 PostgreSQL 14+ 数据库的能力边界。

### Modified Capabilities

- `project-setup`：数据库配置从 PostgreSQL + SQLite 开发演示模式调整为 PostgreSQL-only。
- `database-dialect-abstraction`：方言抽象移除 SQLite 方言、SQLite 转译和 SQLite 错误分类，只保留 PostgreSQL 实现与明确的不支持错误。
- `database-bootstrap-commands`：初始化和 mock 命令不再为 SQLite 准备数据库或转译 SQL。
- `cluster-deployment-mode`：移除 SQLite 专属单机锁定与启动警告，集群开关仅由配置和 PostgreSQL + Redis coordination 约束决定。
- `cluster-coordination-config`：移除 SQLite 禁止集群 coordination 的特殊规则。
- `sql-source-syntax`：SQL 源不再受 SQLite 转译能力约束，可围绕 PostgreSQL 14+ 语法子集治理。
- `plugin-cache-service`：缓存服务规范移除 SQLite 持久化和 SQLite 锁冲突语义。
- `volatile-table-bootstrap`：易失性表引导规范移除 SQLite 场景。
- `release-image-build`：共享 CI 与 release 简要门禁移除 SQLite smoke。

## Impact

- 后端：`apps/lina-core/pkg/dialect`、`apps/lina-core/pkg/dbdriver`、启动初始化命令、系统信息、插件数据服务和缓存相关测试。
- 工具与 CI：`.github/workflows/*`、`.github/image/config.runtime.yaml`、`hack/tests`、`hack/makefiles/database.mk`、`hack/tools/linactl` 测试。
- 依赖：移除 `github.com/gogf/gf/contrib/drivers/sqlite/v2` 以及 SQLite 驱动链路的间接依赖（以 `go mod tidy` 结果为准）。
- 文档：同步英文与中文 README、测试 README、配置模板和现行 OpenSpec 基线。
- i18n：本次不新增、修改或删除用户可见运行时文案翻译键；仅调整配置、文档、测试和后端错误 fallback。
- 缓存一致性：本次不新增缓存；移除 SQLite 单机特殊分支后，单机仍使用进程内/SQL table 策略，集群仍要求 PostgreSQL + Redis coordination。
