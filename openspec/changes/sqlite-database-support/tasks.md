## 1. 依赖与脚手架

- [x] 1.1 在 `apps/lina-core/go.mod` 中加入 `github.com/gogf/gf/contrib/drivers/sqlite/v2` 依赖；执行 `go mod tidy` 确认 `go.sum` 同步更新
- [x] 1.2 在 `apps/lina-core/main.go`（或现有 driver 注册点）匿名 import SQLite driver，与现有 MySQL driver 注册并列
- [x] 1.3 在仓库根 `.gitignore` 增加 `temp/sqlite/*.db` / `temp/sqlite/*.db-shm` / `temp/sqlite/*.db-wal` 三个条目，避免 SQLite 数据文件被误提交（若 `temp/` 已被忽略则无需变更，记录排查结论）
- [x] 1.4 在 `apps/lina-core/manifest/config/config.template.yaml` 与 `config.yaml` 的 `database.default.link` 注释中追加 SQLite 链接示例（`sqlite::@file(./temp/sqlite/linapro.db)`），默认值仍为 MySQL 链接以保持向后兼容
- [x] 1.5 创建 `apps/lina-core/pkg/dialect/` 目录骨架：`dialect.go`（接口与 `From()` 分发函数）、`dialect_mysql.go`、`dialect_sqlite.go`、`dialect_sqlite_translate.go`（SQLite DDL 转译器实现），含主文件包注释与各文件文件级注释

## 2. Dialect 接口与 MySQL 实现

- [x] 2.1 在 `dialect.go` 定义公共稳定 `Dialect` 接口：`Name() string` / `TranslateDDL(ctx, sourceName, ddl) (string, error)` / `PrepareDatabase(ctx, link, rebuild) error` / `SupportsCluster() bool` / `OnStartup(ctx, runtime) error`，并补齐每个方法的接口注释；公开签名不得引用宿主 `internal` 具体服务类型
- [x] 2.2 在 `dialect.go` 实现 `From(link string) (Dialect, error)`：解析 link 协议头（`mysql:` → MySQL，`sqlite:` → SQLite），未识别前缀返回包含已支持前缀列表的明确 `bizerr` 错误
- [x] 2.3 在 MySQL 方言实现中实现：`TranslateDDL` 直接返回输入，`PrepareDatabase` 迁移现有 `prepareInitDatabase` 中 MySQL 专属的链接解析、`DROP DATABASE` 与 `CREATE DATABASE` 逻辑（包括现有 `databaseNameFromMySQLLink` / `serverLinkFromMySQLLink` / `quoteMySQLIdentifier` 三个工具函数迁入），`SupportsCluster` 返回 `true`，`OnStartup` 为 no-op
- [x] 2.4 在 `dialect.go` 定义 `RuntimeConfig` 等窄接口，由宿主 config service 适配实现 `OverrideClusterEnabledForDialect(value bool)`；验证 `pkg/dialect` 可被 `internal` 外部代码导入且不反向依赖宿主私有包
- [x] 2.5 为 MySQL 方言编写单元测试：覆盖 `TranslateDDL(ctx, sourceName, ddl)` 输入输出字节相等、`sourceName` 不影响 no-op、`PrepareDatabase` 链接解析、`SupportsCluster` 返回值

## 3. SQLite 方言与 DDL 转译器

- [x] 3.1 在 SQLite 方言实现中实现 `PrepareDatabase`：解析 SQLite link 取得文件路径，自动 `mkdir -p` 父目录；`rebuild=true` 时删除主 db 文件以及可能存在的 `.db-shm` / `.db-wal` 附属文件；目录创建失败返回包含路径的明确错误
- [x] 3.2 在 SQLite 方言实现中实现 `SupportsCluster() = false`，`Name() = "sqlite"`
- [x] 3.3 在 SQLite 方言实现中实现 `OnStartup`：调用 `runtime.OverrideClusterEnabledForDialect(false)`（步骤 5.1 新增方法由宿主 config service 适配实现），并输出启动提示日志，覆盖：当前为 SQLite 模式、cluster 强制锁定原因、不得用于生产
- [x] 3.4 在 `dialect_sqlite_translate.go` 实现 SQLite DDL 转译器，覆盖以下转换规则：
  - 反引号标识符去除（含字符串内反引号识别避免误伤）
  - 当前 SQL 真实出现的 `INT/BIGINT [UNSIGNED] PRIMARY KEY AUTO_INCREMENT`、`AUTO_INCREMENT PRIMARY KEY`、`NOT NULL AUTO_INCREMENT` + 表级 `PRIMARY KEY(id)` 等写法 → `INTEGER PRIMARY KEY AUTOINCREMENT`
  - `TINYINT / SMALLINT [UNSIGNED]` → `INTEGER`
  - `VARCHAR(N) / CHAR(N) / LONGTEXT / MEDIUMTEXT` → `TEXT`
  - `DECIMAL(M,N)` → `NUMERIC`
  - `INSERT IGNORE INTO` → `INSERT OR IGNORE INTO`
  - 删除 `ENGINE=InnoDB` / `ENGINE=MEMORY` / `DEFAULT CHARSET=utf8mb4` / `COLLATE=utf8mb4_general_ci` 子句
  - 删除列级 `COMMENT '...'` 与表级 `COMMENT='...'`（含 `COMMENT=` 与 `COMMENT '...'` 两种语法变体）
  - 仅删除 `ON UPDATE CURRENT_TIMESTAMP` 部分，保留 `DEFAULT CURRENT_TIMESTAMP`
  - 表内联 `KEY idx_xxx (col)` / `INDEX idx_xxx (col)` / `UNIQUE KEY uk_xxx (col)` / `UNIQUE INDEX uk_xxx (col)` / 当前 SQL 已使用的表达式索引提取为表创建后的独立 `CREATE INDEX` / `CREATE UNIQUE INDEX` 语句
  - 整句删除 `CREATE DATABASE IF NOT EXISTS xxx;` 与 `USE xxx;`
- [x] 3.5 转译器遇到未覆盖的 MySQL 特性（`FULLTEXT INDEX` / `SPATIAL INDEX` / `GENERATED ALWAYS AS` / 分区子句 / `ON DUPLICATE KEY UPDATE` / 数据库专用函数）时返回明确错误，错误消息包含 `sourceName`、行号提示与未覆盖关键字
- [x] 3.6 编写 SQLite DDL 转译器的单元测试：自动扫描当前宿主安装 SQL、插件安装 SQL、宿主 mock SQL、插件 mock SQL 与插件卸载 SQL 作为 fixture，每个 fixture 同时断言"转译成功不报错"与"转译结果可在临时 SQLite 数据库上成功执行"
- [x] 3.7 编写 SQLite DDL 转译器的负面测试：构造包含 `FULLTEXT` / `GENERATED` / `ON DUPLICATE KEY UPDATE` 的 DDL 输入，断言转译器返回包含 `sourceName`、行号提示与未覆盖关键字的错误
- [x] 3.8 编写当前真实 SQL 形态覆盖测试：覆盖 `AUTO_INCREMENT PRIMARY KEY`、表级 `PRIMARY KEY(id)`、`UNIQUE INDEX`、表达式唯一索引 `NULLIF(code, '')` 等已存在写法，确保不靠规范化 SQL 文件绕过转译能力缺口

## 4. cmd init / mock 接入方言层

- [x] 4.1 重构 `apps/lina-core/internal/cmd/cmd_init_database.go`：`prepareInitDatabase` 改为 `dialect.From(link).PrepareDatabase(ctx, link, rebuild)`；将原 MySQL 专属逻辑全部迁入 MySQL 方言实现
- [x] 4.2 修改 `apps/lina-core/internal/cmd/` 的共享 SQL 执行路径：在调用 `splitSQLStatements` 前先调用 `dialect.From(link).TranslateDDL(ctx, sourceName, content)` 转译 SQL 文件内容；`sourceName` 使用源文件路径，转译失败时使用现有快速失败路径并保留失败文件名、行号提示与未覆盖关键字
- [x] 4.3 修改 `apps/lina-core/internal/cmd/cmd_mock.go`：与 4.2 同样的转译接入方式，覆盖 `manifest/sql/mock-data/` 资产；明确 `make mock` 不调用 `PrepareDatabase`，必须依赖 `make init` 已完成初始化
- [ ] 4.4 编写 cmd init / mock 集成测试：分别用 MySQL 与 SQLite 链接执行端到端的 init + mock 流程，断言当前所有宿主与插件 SQL 文件成功执行
- [x] 4.5 验证 `init --rebuild=true` 在 SQLite 模式下正确删除数据库文件并重建
- [x] 4.6 编写 `make mock` 依赖初始化的失败测试：SQLite 配置下未执行 init 或数据库表不存在时运行 mock 必须快速失败，不创建或重建数据库
- [x] 4.7 编写 SQLite 正向集成测试：使用 `database.default.link` 配置的 SQLite 临时文件执行 embedded 宿主 `init --rebuild=true` → `mock`，断言当前宿主初始化 SQL 与 mock SQL 成功执行并写入 `admin` / `user001`

## 5. 集群锁定 + 启动期警告

- [x] 5.1 在 `apps/lina-core/internal/service/config/config_cluster.go` 新增 `OverrideClusterEnabledForDialect(value bool)` 方法：方言层调用后将内存中的 `cluster.enabled` 锁定为指定值，后续所有 `IsClusterEnabled` 调用稳定返回该值，不再读取配置文件
- [x] 5.2 在 `apps/lina-core/internal/cmd/cmd_http_runtime.go`（或等价启动 bootstrap 入口）找到 cluster 服务初始化前的位置，调用 `dialect.From(link).OnStartup(ctx, runtime)`；`runtime` 由宿主 config service 适配，且该调用必须早于 `clusterSvc` 初始化与选举循环启动
- [x] 5.3 编写单元测试：模拟 SQLite link 启动场景，断言 `IsClusterEnabled` 返回 `false` 且日志输出 SQLite 模式、cluster 锁定和不得用于生产等必要内容
- [x] 5.4 编写单元测试：模拟 MySQL link + `cluster.enabled=true` 启动场景，断言 `IsClusterEnabled` 返回 `true` 且无 SQLite 相关警告日志

## 6. 插件 install / uninstall pipeline 接入方言层

- [x] 6.1 在 `apps/lina-core/internal/service/plugin/` 找到执行插件 `manifest/sql/` 的 install 路径，在执行前插入 `dialect.From(link).TranslateDDL(ctx, sourceName, content)` 转译；`sourceName` 包含插件标识、资产类型与失败文件名，转译失败时返回包含行号提示与未覆盖关键字的明确错误
- [x] 6.2 在同模块的 uninstall 路径上做相同接入，覆盖 `manifest/sql/uninstall/` 资产
- [x] 6.3 在 mock-data 加载路径上做相同接入（与步骤 4.3 一致），覆盖插件 `manifest/sql/mock-data/` 资产
- [x] 6.4 编写插件生命周期集成测试：在 SQLite 模式下完整执行"安装 → 启用 → 加 mock → 禁用 → 卸载"流程，断言每一步插件 SQL 资产成功转译并执行
- [x] 6.5 验证现有源码插件（`monitor-loginlog` / `monitor-operlog` / `monitor-server` / `org-center` / `content-notice` / `plugin-demo-source`）与动态插件（`plugin-demo-dynamic`）在 SQLite 模式下安装、启用、卸载行为正常

## 7. kvcache 后端重命名与原子自增重写

- [x] 7.1 重命名 `apps/lina-core/internal/service/kvcache/internal/mysql-memory/` 目录为 `sqltable/`；更新所有导入路径与子文件包声明
- [x] 7.2 在 `apps/lina-core/internal/service/kvcache/kvcache_backend.go` 中将常量 `BackendMySQLMemory` 重命名为 `BackendSQLTable`，字符串值由 `"mysql-memory"` 改为 `"sql-table"`；更新所有引用方
- [x] 7.3 重写 `sqltable/sqltable_ops.go`（原 `mysql_memory_ops.go`）的 `Incr` 实现：去除 `gdb.Raw("LAST_INSERT_ID(value_int + ...)")` 与 `Raw("SELECT LAST_INSERT_ID()")` 调用，改为方言中性的 CAS 重试流程：读取当前整数快照，缺失时 `INSERT IGNORE value_int=0` 初始化后重新读取，通过带 `value_int=<snapshot>` 条件的参数化 `UPDATE` 写入新值；对快照竞争和数据库建议重试的锁冲突执行有限退避重试；保证首次 `incr(delta)` 返回 `delta`，且不修改原始 MySQL `sys_kv_cache ENGINE=MEMORY` 交付 SQL
- [x] 7.4 检查 `sqltable/` 内其他方法是否仍使用 MySQL 专属 SQL，若有同样改写为方言中性形式
- [x] 7.5 更新 `sqltable/` 的现有单元测试：原使用 MySQL 容器的测试改为参数化执行（同时跑 MySQL 与 SQLite），或新增对应的 SQLite 测试套件，断言 `Incr` 在两种引擎下行为一致；覆盖并发首次递增缺失键、首次 `delta` 作为初始值、递增非整数值不变、并发递增不丢失更新
- [x] 7.6 更新 `apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_cache.go`、`hostfn_service_cache_test.go`、`hostfn_service_lock_test.go` 中可能存在的 `BackendMySQLMemory` 字面量引用与测试 fixture 中的 `ENGINE=MEMORY` 文案

## 8. 启动期目录创建与 SQLite 路径治理

- [x] 8.1 验证 SQLite 方言 `PrepareDatabase` 在父目录不存在时自动创建（步骤 3.1 实现），编写覆盖路径不存在场景的单元测试
- [x] 8.2 验证 SQLite link 中包含 `~`（HOME 展开）或绝对路径时的解析行为，明确文档约定（仅支持相对工作目录路径与绝对路径，不展开 `~`），编写测试覆盖
- [x] 8.3 验证 `~/.linapro/data/linapro.db` 这种路径若用户写入会得到清晰的错误（不解析 `~`），错误消息提示用户改用绝对路径或工作目录相对路径

## 9. E2E 测试套件参数化运行

- [x] 9.1 在 `hack/tests/` 增加 SQLite 模式执行通道：测试夹具必须在启动服务前写入测试配置文件中的 `database.default.link=sqlite::@file(./temp/sqlite/linapro.db)` 并自动准备 `temp/sqlite/linapro.db`；后端运行时不得读取命令行参数或环境变量作为数据库方言来源
- [x] 9.2 新增测试用例 `hack/tests/e2e/dialect/TC0164-sqlite-mode-startup.ts`：启动 SQLite 模式后断言终端日志包含 SQLite 启动提示、`/api/v1/cluster/status`（或等价端点）返回单节点状态、登录 admin/admin123 成功
- [x] 9.3 新增测试用例 `hack/tests/e2e/dialect/TC0165-sqlite-mode-business-zero-impact.ts`：在 SQLite 模式下完整跑通"用户列表 → 创建用户 → 修改用户 → 删除用户"与"插件列表 → 启用插件 → 禁用插件"两类核心业务场景，断言行为与 MySQL 模式一致
- [x] 9.4 新增测试用例 `hack/tests/e2e/dialect/TC0166-sqlite-mode-rebuild-and-reseed.ts`：执行 `make init confirm=init rebuild=true` + `make mock confirm=mock` 验证 SQLite 模式下重建数据库 + 加载 mock 数据完整流程
- [x] 9.5 在 CI 配置中保留 SQLite 后端 smoke 通道，覆盖 SQLite 启动警告、单节点 health 与管理员登录；完整 SQLite E2E 通道保留为手动验证入口，避免主 CI 每次运行完整浏览器与插件生命周期回归

## 10. 文档与 README 同步

- [x] 10.1 更新 `apps/lina-core/README.md` 与 `README.zh_CN.md`：在"数据库配置"部分追加 SQLite 链接示例与"演示 / 测试用途、不支持生产、不支持集群"的说明
- [x] 10.2 更新仓库根 `README.md` 与 `README.zh_CN.md`：在"快速开始 / 环境要求"部分追加"可选 SQLite 模式无需 MySQL"的说明
- [x] 10.3 更新 `Makefile` 顶部注释（不修改命令本身）：说明 `make init` / `make mock` 自动按 `database.default.link` 协议头分发到对应方言
- [x] 10.4 检查 i18n 影响面：本变更不涉及前端 UI 文案、菜单、按钮、表单、表格、apidoc 文档等翻译资源调整，明确记录"i18n 资源不需要新增、修改或删除"的判断

## 11. 自我审查与验收

- [x] 11.1 运行 `gofmt`、`go vet ./...` 通过；运行 `go test ./...` 后端单元测试全绿
- [ ] 11.2 在 MySQL 链接下完整执行 `make init confirm=init` → `make mock confirm=mock` → `make dev` → 浏览器访问管理工作台 → 登录 admin/admin123，确认零回归
- [x] 11.3 在 SQLite 链接下完整执行 `make init confirm=init` → `make mock confirm=mock` → `make dev` → 浏览器访问管理工作台 → 登录 admin/admin123，确认终端有清晰 SQLite 单机模式提示且业务功能可用
- [x] 11.4 在 SQLite 链接 + `cluster.enabled=true` 配置下启动，确认 `IsClusterEnabled` 实际为 `false` 且警告日志醒目
- [x] 11.5 在 SQLite 链接下完整跑通插件管理"安装 → 启用 → 卸载"流程（至少覆盖 `monitor-loginlog` 一个源码插件与 `plugin-demo-dynamic` 一个动态插件）
- [x] 11.6 调用 `/lina-review` 技能进行最终代码与规范审查
- [ ] 11.7 运行 `golangci-lint run` 通过（当前本机未安装 `golangci-lint`，待工具可用后执行）

## Feedback

- [x] **FB-1**: 不得为了 SQLite 支持修改原有 MySQL 交付 SQL 中 `013-dynamic-plugin-host-service-extension.sql` 的 `sys_kv_cache` 表结构或 `ENGINE=MEMORY` 引擎类型；SQLite 适配必须收敛在 `pkg/dialect` 执行期转译与 `kvcache incr` 方言中性 CAS 实现中
- [x] **FB-2**: 前端内部包 `unbuild --stub` 生成的 `dist` 存根不得包含构建机器本地绝对路径，源码分发后应能按当前仓库位置解析源码入口
- [x] **FB-3**: 角色管理新增和编辑抽屉中菜单权限选择器的已选统计缺少左侧间距，且修改权限复选框后关闭图标、取消按钮和遮罩关闭没有反馈且无法关闭
- [x] **FB-4**: `kvcache incr` 的数据库可重试写冲突判断不应在 `sqltable` 业务实现中硬编码 MySQL/SQLite 错误文案，应下沉到 `pkg/dialect` 的驱动错误分类能力中
- [x] **FB-5**: `host:state` 的 `state.get` 与 `state.delete` 不应绕过已生成的 `dao.SysPluginState` 直接使用 `g.DB().Model` 访问 `sys_plugin_state`
- [x] **FB-6**: 检查插件泛型资源与 `plugin-demo-source` 固定表实现的数据库访问方式；动态 manifest 表访问可保留动态模型，固定源码插件表访问应复用插件自己的 DAO 和生成列名
- [x] **FB-7**: MySQL 方言数据库初始化不得在实现逻辑中硬编码 `linapro` 库名，应完全使用 `database.default.link` 中配置的数据库名
- [x] **FB-8**: MySQL 方言数据库初始化不得自行解析 GoFrame 数据库链接，应复用 `gdb` 解析后的 `ConfigNode` 获取目标库名并构造 server-level 连接配置
- [x] **FB-9**: `pkg/dialect` 公共包不得直接暴露 `MySQLDialect` / `SQLiteDialect` 具体实现，MySQL 与 SQLite 方言实现应拆分到 `pkg/dialect/internal/mysql` 与 `pkg/dialect/internal/sqlite` 中，通过公共 `Dialect` 接口、工厂函数和必要公共门面能力对外提供稳定边界
- [x] **FB-10**: SQLite DDL 转译器中的 SQL 关键字与标识符匹配应避免大小写敏感缺口，尤其是表级主键列名与列定义大小写不一致时仍应正确识别并转译
- [x] **FB-11**: 降低主 CI 中 SQLite 验证耗时，CI 仅运行 SQLite 后端 smoke，完整 SQLite E2E 保留为手动验证入口
- [x] **FB-12**: SQLite 后端 smoke 脚本仅服务测试与 CI 验证，应放在 `hack/tests/scripts/` 测试脚本目录下，而不是通用 `hack/scripts/` 目录
- [x] **FB-13**: 英文环境下插件管理列表中未安装动态插件的名称和描述仍显示中文，应按当前语言展示插件清单本地化文案
- [x] **FB-14**: 角色管理新增和编辑抽屉中菜单权限已选统计与左侧模式选择缺少稳定间距，且勾选权限后关闭图标、取消按钮和遮罩关闭应能正常触发未保存确认并关闭
- [x] **FB-15**: 管理工作台 i18n 首次启动语言应自动识别浏览器语言，中文浏览器默认使用中文，其他浏览器默认使用英文，且不得覆盖用户已保存的语言偏好
- [x] **FB-16**: 角色管理新增和编辑抽屉的菜单权限树应保持目录包含菜单、菜单包含按钮的展示结构；动态新增权限应归入系统菜单之后的动态权限分组，而不是显示在最前面的根级位置
- [x] **FB-17**: 移除仅包装 Git clone 的 LinaPro 安装脚本，同时删除相关安装器文档、测试入口与 OpenSpec 能力描述
- [x] **FB-18**: `make build` 与 `make image` 需要支持跨平台构建；`make image platforms=linux/amd64,linux/arm64 registry=<registry> push=1` 必须构建并推送多架构 `Docker` 镜像
- [x] **FB-19**: 普通用户仅授权组织管理和内容管理目录权限后，访问对应插件页面不应因为页面复用字典选项接口缺少 `system:dict:query` 而报无权限
- [x] **FB-20**: `hack/config.yaml` 的 `build.os` / `build.arch` / `build.platform` 应收敛为 `build.platforms` 数组；数组项支持 `<goos>/<goarch>` 与 `auto`，命令行覆盖使用 `platforms=linux/amd64,linux/arm64`
- [x] **FB-21**: 新增 GitHub Actions nightly build workflow，每天凌晨自动构建并发布 `linux/amd64` 与 `linux/arm64` 多架构 Docker 镜像到 `ghcr.io`
- [x] **FB-22**: `make build` 与 `make image` 支持 `config=<path>` 构建配置参数，用于指定默认读取 `hack/config.yaml` 的配置文件路径
- [x] **FB-23**: 新增 GitHub Actions 标签发布镜像 workflow，仅在新建标签时构建并发布对应标签与 `latest` 的 `linux/amd64`、`linux/arm64` 多架构 Docker 镜像
- [x] **FB-24**: 新增 GitHub Actions nightly test workflow，每天 `Asia/Shanghai 00:00` 自动运行完整 Go/前端单元测试与完整 Playwright E2E 套件，持续验证项目健康度
- [x] **FB-25**: 移除 Nightly Test Go 单测目录硬编码，改为从 `go.work` 自动发现 workspace 模块并执行全量 `go test ./...`
- [x] **FB-26**: 主 CI 的 SQLite smoke 与 SQLite E2E 支持代码不应继续断言已删除的 “Switch database.default.link back to a MySQL link” 启动日志
- [x] **FB-27**: 定位并修复主 CI 中 `Go unit tests / Go unit tests` job 失败
- [x] **FB-28**: SQLite 运行时启用 `monitor-server` 插件后首次采集服务监控数据不应因隐式 `Save()` 缺少冲突列而报错
- [x] **FB-29**: SQLite 模式下执行日志管理列表接口不应因 `EXISTS ((SELECT ...))` 子查询语法报 `Database Operation Error`
- [x] **FB-30**: Go 单元测试入口与 GitHub Actions 后端单测流程应统一使用 `go test -race` 检测潜在竞态条件
  - 2026-05-08: 已新增 `make test-go` 本地入口，按 `go.work` 自动发现所有 Go workspace 模块并执行 `go test -race -v ./...`；GitHub Actions 后端单测 reusable workflow 同步改为 `go test -race -v ./...`，并将超时时间提高到 120 分钟以覆盖 race detector 开销。验证通过：`make test-go`。
  - 2026-05-08: lina-review 结论：本次仅调整 Go 单测执行入口、CI 命令、测试治理 allowlist 和 OpenSpec 任务记录；不涉及 API、数据库、前端运行时文案、i18n 资源、数据权限或缓存一致性变更。
