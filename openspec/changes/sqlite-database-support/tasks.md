## 1. 依赖与脚手架

- [ ] 1.1 在 `apps/lina-core/go.mod` 中加入 `github.com/gogf/gf/contrib/drivers/sqlite/v2` 依赖；执行 `go mod tidy` 确认 `go.sum` 同步更新
- [ ] 1.2 在 `apps/lina-core/main.go`（或现有 driver 注册点）匿名 import SQLite driver，与现有 MySQL driver 注册并列
- [ ] 1.3 在仓库根 `.gitignore` 增加 `temp/sqlite/*.db` / `temp/sqlite/*.db-shm` / `temp/sqlite/*.db-wal` 三个条目，避免 SQLite 数据文件被误提交（若 `temp/` 已被忽略则无需变更，记录排查结论）
- [ ] 1.4 在 `apps/lina-core/manifest/config/config.template.yaml` 与 `config.yaml` 的 `database.default.link` 注释中追加 SQLite 链接示例（`sqlite::./temp/sqlite/linapro.db`），默认值仍为 MySQL 链接以保持向后兼容
- [ ] 1.5 创建 `apps/lina-core/pkg/dialect/` 目录骨架：`dialect.go`（接口与 `From()` 分发函数）、`dialect_mysql.go`、`dialect_sqlite.go`、`dialect_sqlite_translate.go`（SQLite DDL 转译器实现），含主文件包注释与各文件文件级注释

## 2. Dialect 接口与 MySQL 实现

- [ ] 2.1 在 `dialect.go` 定义 `Dialect` 接口：`Name() string` / `TranslateDDL(ctx, ddl) (string, error)` / `PrepareDatabase(ctx, link, rebuild) error` / `SupportsCluster() bool` / `OnStartup(ctx, configSvc) error`，并补齐每个方法的接口注释
- [ ] 2.2 在 `dialect.go` 实现 `From(link string) (Dialect, error)`：解析 link 协议头（`mysql:` → MySQL，`sqlite:` → SQLite），未识别前缀返回包含已支持前缀列表的明确 `bizerr` 错误
- [ ] 2.3 在 `dialect_mysql.go` 实现 `MySQLDialect`：`TranslateDDL` 直接返回输入，`PrepareDatabase` 迁移现有 `prepareInitDatabase` 中 MySQL 专属的链接解析、`DROP DATABASE` 与 `CREATE DATABASE` 逻辑（包括现有 `databaseNameFromMySQLLink` / `serverLinkFromMySQLLink` / `quoteMySQLIdentifier` 三个工具函数迁入），`SupportsCluster` 返回 `true`，`OnStartup` 为 no-op
- [ ] 2.4 为 `MySQLDialect` 编写单元测试：覆盖 `TranslateDDL` 输入输出字节相等、`PrepareDatabase` 链接解析、`SupportsCluster` 返回值

## 3. SQLite 方言与 DDL 转译器

- [ ] 3.1 在 `dialect_sqlite.go` 实现 `SQLiteDialect.PrepareDatabase`：解析 SQLite link 取得文件路径，自动 `mkdir -p` 父目录；`rebuild=true` 时删除主 db 文件以及可能存在的 `.db-shm` / `.db-wal` 附属文件；目录创建失败返回包含路径的明确错误
- [ ] 3.2 在 `dialect_sqlite.go` 实现 `SQLiteDialect.SupportsCluster() = false`，`Name() = "sqlite"`
- [ ] 3.3 在 `dialect_sqlite.go` 实现 `SQLiteDialect.OnStartup`：调用 `configSvc.OverrideClusterEnabledForDialect(false)`（步骤 5.1 新增方法），并通过 `logger.Warningf(ctx, ...)` 输出至少 4 行 `[WARNING]` 级别日志，覆盖：当前为 SQLite 模式、cluster 强制锁定原因、不得用于生产、切换回 MySQL 的指引
- [ ] 3.4 在 `dialect_sqlite_translate.go` 实现 SQLite DDL 转译器，覆盖以下转换规则：
  - 反引号标识符去除（含字符串内反引号识别避免误伤）
  - `INT/BIGINT [UNSIGNED] PRIMARY KEY AUTO_INCREMENT` → `INTEGER PRIMARY KEY AUTOINCREMENT`
  - `TINYINT / SMALLINT [UNSIGNED]` → `INTEGER`
  - `VARCHAR(N) / CHAR(N) / LONGTEXT / MEDIUMTEXT` → `TEXT`
  - `DECIMAL(M,N)` → `NUMERIC`
  - `INSERT IGNORE INTO` → `INSERT OR IGNORE INTO`
  - 删除 `ENGINE=InnoDB` / `ENGINE=MEMORY` / `DEFAULT CHARSET=utf8mb4` / `COLLATE=utf8mb4_general_ci` 子句
  - 删除列级 `COMMENT '...'` 与表级 `COMMENT='...'`（含 `COMMENT=` 与 `COMMENT '...'` 两种语法变体）
  - 仅删除 `ON UPDATE CURRENT_TIMESTAMP` 部分，保留 `DEFAULT CURRENT_TIMESTAMP`
  - 表内联 `KEY idx_xxx (col)` / `INDEX idx_xxx (col)` / `UNIQUE KEY uk_xxx (col)` 提取为表创建后的独立 `CREATE INDEX` / `CREATE UNIQUE INDEX` 语句
  - 整句删除 `CREATE DATABASE IF NOT EXISTS xxx;` 与 `USE xxx;`
- [ ] 3.5 转译器遇到未覆盖的 MySQL 特性（`FULLTEXT INDEX` / `SPATIAL INDEX` / `GENERATED ALWAYS AS` / 分区子句 / `ON DUPLICATE KEY UPDATE` / 数据库专用函数）时返回明确错误，错误消息包含未覆盖关键字
- [ ] 3.6 编写 SQLite DDL 转译器的单元测试：以现有 14 个宿主 SQL 文件 + 7 个插件 SQL 文件 + 7 个 mock SQL 文件作为 fixture，每个 fixture 同时断言"转译成功不报错"与"转译结果可在临时 SQLite 数据库上成功执行"
- [ ] 3.7 编写 SQLite DDL 转译器的负面测试：构造包含 `FULLTEXT` / `GENERATED` / `ON DUPLICATE KEY UPDATE` 的 DDL 输入，断言转译器返回包含未覆盖关键字的错误

## 4. cmd init / mock 接入方言层

- [ ] 4.1 重构 `apps/lina-core/internal/cmd/cmd_init_database.go`：`prepareInitDatabase` 改为 `dialect.From(link).PrepareDatabase(ctx, link, rebuild)`；将原 MySQL 专属逻辑全部迁入 `MySQLDialect`
- [ ] 4.2 修改 `apps/lina-core/internal/cmd/cmd_init.go`：在调用 `splitSQLStatements` 前先调用 `dialect.From(link).TranslateDDL(ctx, content)` 转译 SQL 文件内容；转译失败时使用现有快速失败路径并保留失败文件名
- [ ] 4.3 修改 `apps/lina-core/internal/cmd/cmd_mock.go`：与 4.2 同样的转译接入方式，覆盖 `manifest/sql/mock-data/` 资产
- [ ] 4.4 编写 cmd init / mock 集成测试：分别用 MySQL 与 SQLite 链接执行端到端的 init + mock 流程，断言所有 28 个 SQL 文件成功执行
- [ ] 4.5 验证 `init --rebuild=true` 在 SQLite 模式下正确删除数据库文件并重建

## 5. 集群锁定 + 启动期警告

- [ ] 5.1 在 `apps/lina-core/internal/service/config/config_cluster.go` 新增 `OverrideClusterEnabledForDialect(value bool)` 方法：方言层调用后将内存中的 `cluster.enabled` 锁定为指定值，后续所有 `IsClusterEnabled` 调用稳定返回该值，不再读取配置文件
- [ ] 5.2 在 `apps/lina-core/internal/cmd/cmd_http_runtime.go`（或等价启动 bootstrap 入口）找到 cluster 服务初始化前的位置，调用 `dialect.From(link).OnStartup(ctx, configSvc)`；该调用必须早于 `clusterSvc` 初始化与选举循环启动
- [ ] 5.3 编写单元测试：模拟 SQLite link 启动场景，断言 `IsClusterEnabled` 返回 `false` 且日志输出至少 4 行 `[WARNING]` 级别消息覆盖必要内容
- [ ] 5.4 编写单元测试：模拟 MySQL link + `cluster.enabled=true` 启动场景，断言 `IsClusterEnabled` 返回 `true` 且无 SQLite 相关警告日志

## 6. 插件 install / uninstall pipeline 接入方言层

- [ ] 6.1 在 `apps/lina-core/internal/service/plugin/` 找到执行插件 `manifest/sql/` 的 install 路径，在执行前插入 `dialect.From(link).TranslateDDL` 转译；转译失败时返回包含插件标识、资产类型与失败文件名的明确错误
- [ ] 6.2 在同模块的 uninstall 路径上做相同接入，覆盖 `manifest/sql/uninstall/` 资产
- [ ] 6.3 在 mock-data 加载路径上做相同接入（与步骤 4.3 一致），覆盖插件 `manifest/sql/mock-data/` 资产
- [ ] 6.4 编写插件生命周期集成测试：在 SQLite 模式下完整执行"安装 → 启用 → 加 mock → 禁用 → 卸载"流程，断言每一步插件 SQL 资产成功转译并执行
- [ ] 6.5 验证现有源码插件（`monitor-loginlog` / `monitor-operlog` / `monitor-server` / `org-center` / `content-notice` / `plugin-demo-source`）与动态插件（`plugin-demo-dynamic`）在 SQLite 模式下安装、启用、卸载行为正常

## 7. kvcache 后端重命名与原子自增重写

- [ ] 7.1 重命名 `apps/lina-core/internal/service/kvcache/internal/mysql-memory/` 目录为 `sqltable/`；更新所有导入路径与子文件包声明
- [ ] 7.2 在 `apps/lina-core/internal/service/kvcache/kvcache_backend.go` 中将常量 `BackendMySQLMemory` 重命名为 `BackendSQLTable`，字符串值由 `"mysql-memory"` 改为 `"sql-table"`；更新所有引用方
- [ ] 7.3 重写 `sqltable/sqltable_ops.go`（原 `mysql_memory_ops.go`）的 `Incr` 实现：去除 `gdb.Raw("LAST_INSERT_ID(value_int + ...)")` 与 `Raw("SELECT LAST_INSERT_ID()")` 调用，改为事务内 `dao.SysKvCache.Ctx(ctx).Where(...).Data(do.SysKvCache{ValueInt: ...增量更新}).Update()` + `dao.SysKvCache.Ctx(ctx).Where(...).Value(cols.ValueInt)` 的两步顺序；保证返回新值与原子性
- [ ] 7.4 检查 `sqltable/` 内其他方法是否仍使用 MySQL 专属 SQL，若有同样改写为方言中性形式
- [ ] 7.5 更新 `sqltable/` 的现有单元测试：原使用 MySQL 容器的测试改为参数化执行（同时跑 MySQL 与 SQLite），或新增对应的 SQLite 测试套件，断言 `Incr` 在两种引擎下行为一致
- [ ] 7.6 更新 `apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_cache.go`、`hostfn_service_cache_test.go`、`hostfn_service_lock_test.go` 中可能存在的 `BackendMySQLMemory` 字面量引用与测试 fixture 中的 `ENGINE=MEMORY` 文案

## 8. 启动期目录创建与 SQLite 路径治理

- [ ] 8.1 验证 `SQLiteDialect.PrepareDatabase` 在父目录不存在时自动创建（步骤 3.1 实现），编写覆盖路径不存在场景的单元测试
- [ ] 8.2 验证 SQLite link 中包含 `~`（HOME 展开）或绝对路径时的解析行为，明确文档约定（仅支持相对工作目录路径与绝对路径，不展开 `~`），编写测试覆盖
- [ ] 8.3 验证 `~/.linapro/data/linapro.db` 这种路径若用户写入会得到清晰的错误（不解析 `~`），错误消息提示用户改用绝对路径或工作目录相对路径

## 9. E2E 测试套件参数化运行

- [ ] 9.1 在 `hack/tests/` 增加 SQLite 模式参数化执行通道：通过环境变量（如 `LINAPRO_E2E_DIALECT=sqlite`）控制测试启动时使用的数据库链接，自动准备 `temp/sqlite/linapro.db`
- [ ] 9.2 新增测试用例 `hack/tests/e2e/dialect/TC0164-sqlite-mode-startup.ts`：启动 SQLite 模式后断言终端日志包含至少 4 行 SQLite 警告、`/api/v1/cluster/status`（或等价端点）返回单节点状态、登录 admin/admin123 成功
- [ ] 9.3 新增测试用例 `hack/tests/e2e/dialect/TC0165-sqlite-mode-business-zero-impact.ts`：在 SQLite 模式下完整跑通"用户列表 → 创建用户 → 修改用户 → 删除用户"与"插件列表 → 启用插件 → 禁用插件"两类核心业务场景，断言行为与 MySQL 模式一致
- [ ] 9.4 新增测试用例 `hack/tests/e2e/dialect/TC0166-sqlite-mode-rebuild-and-reseed.ts`：执行 `make init confirm=init rebuild=true` + `make mock confirm=mock` 验证 SQLite 模式下重建数据库 + 加载 mock 数据完整流程
- [ ] 9.5 在 CI 配置中并行触发 MySQL 与 SQLite 两条 E2E 通道（若 CI 资源紧张则至少在主线合并前手动跑过 SQLite 通道一次）

## 10. 文档与 README 同步

- [ ] 10.1 更新 `apps/lina-core/README.md` 与 `README.zh_CN.md`：在"数据库配置"部分追加 SQLite 链接示例与"演示 / 测试用途、不支持生产、不支持集群"的说明
- [ ] 10.2 更新仓库根 `README.md` 与 `README.zh_CN.md`：在"快速开始 / 环境要求"部分追加"可选 SQLite 模式无需 MySQL"的说明
- [ ] 10.3 更新 `Makefile` 顶部注释（不修改命令本身）：说明 `make init` / `make mock` 自动按 `database.default.link` 协议头分发到对应方言
- [ ] 10.4 检查 i18n 影响面：本变更不涉及前端 UI 文案、菜单、按钮、表单、表格、apidoc 文档等翻译资源调整，明确记录"i18n 资源不需要新增、修改或删除"的判断

## 11. 自我审查与验收

- [ ] 11.1 运行 `gofmt`、`go vet`、`golangci-lint` 通过；运行 `go test ./...` 后端单元测试全绿
- [ ] 11.2 在 MySQL 链接下完整执行 `make init confirm=init` → `make mock confirm=mock` → `make dev` → 浏览器访问管理工作台 → 登录 admin/admin123，确认零回归
- [ ] 11.3 在 SQLite 链接下完整执行 `make init confirm=init` → `make mock confirm=mock` → `make dev` → 浏览器访问管理工作台 → 登录 admin/admin123，确认终端有醒目 `[WARNING]` 输出且业务功能可用
- [ ] 11.4 在 SQLite 链接 + `cluster.enabled=true` 配置下启动，确认 `IsClusterEnabled` 实际为 `false` 且警告日志醒目
- [ ] 11.5 在 SQLite 链接下完整跑通插件管理"安装 → 启用 → 卸载"流程（至少覆盖 `monitor-loginlog` 一个源码插件与 `plugin-demo-dynamic` 一个动态插件）
- [ ] 11.6 调用 `/lina-review` 技能进行最终代码与规范审查
