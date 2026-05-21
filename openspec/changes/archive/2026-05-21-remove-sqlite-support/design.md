## Context

LinaPro 当前代码同时维护 PostgreSQL 与 SQLite 两条运行路径：`pkg/dbdriver` 注册 PG/SQLite 驱动，`pkg/dialect` 按 `database.default.link` 分发 PostgreSQL 或 SQLite 方言，初始化命令通过 `Dialect.TranslateDDL` 把 PostgreSQL 源 SQL 转译到 SQLite，启动钩子在 SQLite 下强制关闭集群模式，CI 和 E2E 还维护 SQLite smoke 通道。

这与当前项目以 PostgreSQL 14+ 为默认数据库的方向不一致。继续支持 SQLite 会让 SQL 源治理、插件 SQL 生命周期、缓存/集群语义和 CI 门禁持续承担一条非生产数据库路径。

## Goals / Non-Goals

**Goals:**

- 运行时、初始化命令、mock 命令、CI 和 E2E 不再支持 SQLite。
- `sqlite:` 链接必须明确失败，不能静默回退到 PostgreSQL 或进入半初始化状态。
- 移除 SQLite 驱动依赖、方言实现、DDL 转译器、启动降级钩子、SQLite 专属测试和脚本。
- 现有 PostgreSQL 初始化、插件 SQL 生命周期、系统信息和缓存服务继续可编译、可测试。
- 明确记录 i18n 与缓存影响：无运行时翻译资源变更；不新增缓存，只删除 SQLite 特殊一致性语义。

**Non-Goals:**

- 不引入 MySQL、Redis-only 或其他数据库支持。
- 不重写现有 SQL schema 到 PostgreSQL 高级特性；本次只解除 SQLite 支持约束。
- 不改变业务 API、权限模型、数据权限边界、插件 manifest 结构或前端页面交互。
- 不自动迁移已有 SQLite 数据文件；使用方需迁移到 PostgreSQL。

## Decisions

1. SQLite 链接快速失败，而不是保留只读兼容层。

   备选方案是保留 SQLite 驱动注册但不承诺测试覆盖。该方案会制造“看似可用”的隐性支持，后续变更仍会被 SQLite 约束拖住。因此选择在 `dialect.From` 和驱动注册边界明确拒绝 SQLite。

2. 保留 `pkg/dialect` 抽象，但只保留 PostgreSQL 具体实现。

   备选方案是完全删除方言层，直接在初始化命令里调用 PostgreSQL 实现。该方案会扩大调用点改动，并削弱未来数据库边界治理。保留抽象能让 `PrepareDatabase`、`DatabaseVersion`、`QueryTableMetadata` 和错误分类继续集中管理。

3. 删除 SQLite 专属测试与 E2E，而不是改成跳过。

   SQLite 不再是受支持目标后，保留跳过测试会造成维护噪音和误导。对应验证改为 PostgreSQL 单元测试、SQL 资产 smoke（需要 `LINA_TEST_PGSQL_LINK` 时显式启用）和静态扫描。

4. 默认镜像运行配置改用 PostgreSQL 链接。

   SQLite 文件型数据库不再适合作为容器默认运行配置。镜像配置必须显式指向 PostgreSQL 服务；部署方可通过环境或挂载配置覆盖主机、账号和库名。

## Risks / Trade-offs

- [Risk] 现有本地开发者使用 SQLite 文件启动会失败。  
  Mitigation：配置模板、README 和错误信息统一指向 PostgreSQL 链接与初始化命令。

- [Risk] 移除 SQLite 转译后，部分旧规范仍提到跨方言子集。  
  Mitigation：本变更同步修改相关 OpenSpec 基线，并通过 `rg` 扫描当前非归档路径的 SQLite 引用。

- [Risk] Go module 依赖可能仍通过间接路径保留 SQLite 包。  
  Mitigation：运行 `go mod tidy`，并用 `rg` 检查 `apps/lina-core/go.mod` / `go.sum` 中 SQLite 驱动残留。

- [Risk] CI workflow 输入删除后调用方未同步。  
  Mitigation：同步修改 main、nightly、release 与 reusable suite，删除 SQLite reusable workflow 和脚本入口。

## Migration Plan

1. 将 `database.default.link` 配置改为 PostgreSQL，例如 `pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable`。
2. 运行 `make init confirm=init rebuild=true` 初始化 PostgreSQL 数据库。
3. 如需演示数据，运行 `make mock confirm=mock`。
4. 原 SQLite 数据文件不再由 LinaPro 读取；需要保留历史数据时，先由使用方自行导出并迁移到 PostgreSQL。

回滚策略：如必须恢复 SQLite，需要恢复本变更删除的驱动注册、方言实现、转译器、测试脚本和 CI workflow，并重新建立完整测试门禁；不建议只恢复依赖或链接前缀。

## Open Questions

- 暂无。用户需求明确为移除 SQLite 支持，本次按破坏性变更处理。
