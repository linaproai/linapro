# 实施任务清单

## 1. Database Engine Migration: Spike 验证

- [x] 1.1 写最小 main.go 验证 GoFrame `gogf/gf/contrib/drivers/pgsql/v2` 驱动可连通
- [x] 1.2 在 PG 上建测试表验证 GoFrame ORM 对保留字列名（`key`/`value`）的自动加引号行为
- [x] 1.3 在 PG 上验证 `GENERATED ALWAYS AS IDENTITY` 的 `InsertAndGetId` 返回值正确
- [x] 1.4 验证 GoFrame PG 驱动对时间相关参数的需求，确认 link 字符串简化
- [x] 1.5 检查 `prepare-packed-assets.sh` 是否对 SQL 文件做 SQLite 语法预检查，如有则调整

## 2. Database Engine Migration: 方言层骨架改造

- [x] 2.1 删除 `apps/lina-core/pkg/dialect/internal/mysql/` 整个目录及其测试文件
- [x] 2.2 删除 `apps/lina-core/pkg/dialect/dialect_mysql_test.go`
- [x] 2.3 新建 `apps/lina-core/pkg/dialect/internal/postgres/` 目录
- [x] 2.4 新建 `postgres/dialect.go`：实现 `Name()`、`SupportsCluster()`、`OnStartup()`、`DatabaseVersion()`、`TranslateDDL`（no-op）
- [x] 2.5 新建 `postgres/error.go`：定义 PG 错误码常量和 `IsRetryableWriteConflict`
- [x] 2.6 新建 `postgres/prepare.go`：实现 `PrepareDatabase`（连系统库 + 终止连接 + DROP/CREATE）
- [x] 2.7 新建 `postgres/metadata.go`：实现 `QueryTableMetadata`（information_schema + pg_class）
- [x] 2.8 新建对应测试文件覆盖 Name、错误码、准备数据库、元数据查询
- [x] 2.9 修改 `dialect.go`：删除 `mysqlPrefix`，新增 `pgsqlPrefix`，`mysql:` 前缀返回"不再支持"错误
- [x] 2.10 在 `Dialect` 接口中新增 `QueryTableMetadata` 方法签名和 `TableMeta` 类型
- [x] 2.11 在 `pkg/dialect/internal/sqlite/dialect.go` 中实现 `QueryTableMetadata`
- [x] 2.12 新建 `dialect_postgres_test.go`：测试前缀分发、MySQL 前缀拒绝、`SupportsCluster()=true`

## 3. Database Engine Migration: SQLite 翻译器重写

- [x] 3.1 重写 `sqlite/translate.go`：用 PG 14+ 语法作为输入基准
- [x] 3.2 实现 `GENERATED ALWAYS AS IDENTITY` → `INTEGER PRIMARY KEY AUTOINCREMENT`
- [x] 3.3 实现整数类型映射：`INT/BIGINT/SMALLINT` → `INTEGER`
- [x] 3.4 实现字符串类型映射：`VARCHAR(n)/CHAR(n)/TEXT` → `TEXT`
- [x] 3.5 实现 `BYTEA` → `BLOB` 映射
- [x] 3.6 实现 `TIMESTAMP` → `DATETIME` 映射
- [x] 3.7 实现 `DECIMAL(m,n)` → `NUMERIC` 映射
- [x] 3.8 实现 `COMMENT ON TABLE/COLUMN` 整句丢弃规则
- [x] 3.9 确认 SQLite 翻译器不支持 `TRUNCATE`，本次 SQL 源不依赖该语法
- [x] 3.10 保留 `INSERT ... ON CONFLICT DO NOTHING`、`CREATE INDEX`、双引号标识符
- [x] 3.11 实现快速失败逻辑：识别并拒绝 PG 高级特性
- [x] 3.12 重写翻译器测试：覆盖每条翻译规则 + 项目实际 SQL 文件全文翻译
- [x] 3.13 删除原"MySQL → SQLite"翻译相关的过时正则与测试用例

## 4. Database Engine Migration: 宿主 SQL 源改写

- [x] 4.1 改写 `001-project-init.sql`
- [x] 4.2 改写 `002-dict-type-data.sql`
- [x] 4.3 改写 `005-file-storage.sql`
- [x] 4.4 改写 `006-online-session.sql`
- [x] 4.5 改写 `007-config-management.sql`
- [x] 4.6 改写 `008-menu-role-management.sql`
- [x] 4.7 改写 `010-distributed-locker.sql`
- [x] 4.8 改写 `011-plugin-framework.sql`
- [x] 4.9 改写 `012-plugin-host-call.sql`
- [x] 4.10 改写 `013-dynamic-plugin-host-service-extension.sql`
- [x] 4.11 改写 `014-scheduled-job-management.sql`
- [x] 4.12 改写 `015-distributed-cache-consistency.sql`
- [x] 4.13 改写所有 mock-data SQL
- [x] 4.13.1 盘点每个 INSERT IGNORE 目标表的幂等依据
- [x] 4.14 在 PG 数据库上逐个执行改写后的 SQL 文件验证
- [x] 4.15 用 SQLite 翻译器翻译 PG 源 SQL 并在内存 SQLite 上执行验证

## 5. Database Engine Migration: MEMORY 表改造

- [x] 5.1 确认 DDL 不含 ENGINE=MEMORY 与 UNLOGGED/TEMPORARY 关键字
- [x] 5.2 检查启动路径不对易失性表执行 TRUNCATE 或无条件全表 DELETE
- [x] 5.3 确认 sys_online_session 基于 last_active_time 处理过期会话
- [x] 5.4 确认 sys_locker 基于 expire_time 判断过期并允许抢占
- [x] 5.5 确认 sys_kv_cache 基于 expire_at 判定过期记录
- [x] 5.6 写集成测试：重启后未过期记录仍可用，过期记录按规则失效
- [x] 5.7 写多进程集群模拟测试

## 6. Database Engine Migration: plugin_data_table_comment 重构

- [x] 6.1 修改 `plugin_data_table_comment.go`：删除硬编码 information_schema 查询
- [x] 6.2 引入 `dialect.FromDatabase(g.DB()).QueryTableMetadata(...)`
- [x] 6.3 映射查询结果为现有调用方期望的数据结构
- [x] 6.4 写单元测试覆盖 PG 与 SQLite 两个分支

## 7. Database Engine Migration: 配置/工具链/驱动切换

- [x] 7.1 修改 config.yaml：默认 link 改为 PG link
- [x] 7.2 修改 config.template.yaml
- [x] 7.3 修改 hack/config.yaml
- [x] 7.4 修改 go.mod：移除 MySQL 驱动，新增 PG 驱动
- [x] 7.5 修改 main.go：驱动 import 切换
- [x] 7.6 运行 go mod tidy
- [x] 7.7 搜索并替换所有静态导入 mysql 驱动的源文件

## 8. Database Engine Migration: uint64 → int64

- [x] 8.1 全仓 grep uint64，按来源分组
- [x] 8.2 识别 service 层中由 MySQL UNSIGNED 派生的 uint64 字段
- [x] 8.3 识别 api DTO 中由 MySQL UNSIGNED 派生的 uint64 字段
- [x] 8.4 仅替换 MySQL UNSIGNED 派生类型为 int64
- [x] 8.5 编译验证
- [x] 8.6 单元测试覆盖

## 9. Database Engine Migration: DAO 重新生成

- [x] 9.1 启动本地 PG 容器
- [x] 9.2 运行 make init confirm=init
- [x] 9.3 运行 make dao 重新生成
- [x] 9.4 编译验证
- [x] 9.5 检查生成的 entity 类型
- [x] 9.6 对 8 个插件分别运行 make dao
- [x] 9.7 在插件业务代码中扫描 uint64 并替换

## 10. Database Engine Migration: 插件 SQL 改写

- [x] 10.1 改写 content-notice 插件 SQL
- [x] 10.2 改写 monitor-loginlog 插件 SQL
- [x] 10.3 改写 monitor-operlog 插件 SQL
- [x] 10.4 改写 monitor-server 插件 SQL
- [x] 10.5 改写 org-center 插件 SQL
- [x] 10.6 改写 monitor-online 插件 SQL
- [x] 10.7 改写 plugin-demo-dynamic 插件 SQL
- [x] 10.8 改写 plugin-demo-source 插件 SQL
- [x] 10.9 确认 demo-control 只有 .gitkeep 占位
- [x] 10.9.1 盘点插件 SQL 中每个 INSERT IGNORE 目标表的幂等依据
- [x] 10.10 安装/启用每个插件验证 install SQL
- [x] 10.11 卸载每个插件验证 uninstall SQL
- [x] 10.12 通过 SQLite 翻译器验证插件 SQL

## 11. Database Engine Migration: CI/镜像/文档

- [x] 11.1 更新本地 PostgreSQL 启动说明
- [x] 11.2 修改 GitHub Actions 使用 services.postgres
- [x] 11.3 确认应用不隐式管理数据库
- [x] 11.4 检查 Dockerfile 移除 mysql 客户端依赖
- [x] 11.5 确认镜像构建成功
- [x] 11.6 确认 make dev 不自动启动数据库
- [x] 11.7 调整 make init 连接错误提示
- [x] 12.1 修改项目根 README.md
- [x] 12.2 修改项目根 README.zh-CN.md
- [x] 12.3 修改 apps/lina-core/README.md
- [x] 12.4 修改 apps/lina-core/README.zh-CN.md
- [x] 12.5 修改 CLAUDE.md
- [x] 12.6 增加"切换到 SQLite"指南
- [x] 12.7 增加"PG 14+ 最低版本"说明
- [x] 12.8 增加"切换到外部托管 PG"指南
- [x] 12.8.1 说明 make init 权限要求
- [x] 12.9 检查归档文档中 MySQL 引用
- [x] 12.10 验证双语 README 内容一致

## 13. Database Engine Migration: 测试验证

- [x] 13.1 单元测试覆盖
- [x] 13.2 集成测试 1：PG 上 make init + mock + 启动
- [x] 13.2.1 重复执行验证幂等性
- [x] 13.3 集成测试 2：PG 上 rebuild=true
- [x] 13.4 集成测试 3：SQLite 上 make init + 启动
- [x] 13.4.1 SQLite 上重复执行验证幂等性
- [x] 13.5 E2E 用例：完整业务流
- [x] 13.6 E2E/集成用例：重启后易失性表不清空
- [x] 13.7 多进程集群模拟测试
- [x] 13.8 SQL 幂等性检查测试
- [x] 13.8.1 默认 collation 验证
- [x] 13.9 全部测试通过

## 14. Database Engine Migration: 手动验证

- [x] 14.1 全新 clone 按 README 步骤启动验证
- [x] 14.2 admin/admin123 登录成功
- [x] 14.3 用户管理 CRUD
- [x] 14.4 角色管理
- [x] 14.5 字典管理（含保留字列读写）
- [x] 14.6 配置管理（含保留字列读写）
- [x] 14.7 部门/岗位 CRUD
- [x] 14.8 文件管理
- [x] 14.9 定时任务
- [x] 14.10 监控页面显示 PostgreSQL 版本
- [x] 14.11 插件管理安装/启用/禁用/卸载
- [x] 14.12 重启后易失性表验证
- [x] 14.13 切换到 SQLite 验证
- [x] 14.14 切换到 mysql: 验证启动失败

## 15. Database Engine Migration: 文档自检

- [x] 15.1 声明有限 i18n 影响
- [x] 15.2 声明不引入新缓存策略
- [x] 15.3 检查双语 README 一致
- [x] 15.4 openspec validate 通过
- [x] 15.5 lina-review 审查通过

## 16. Startup SQL Efficiency: 基线与配置

- [x] 16.1 记录当前启动基线
- [x] 16.2 调整 config.yaml 的 database.default.debug 默认值为 false
- [x] 16.3 增加默认启动日志 smoke 断言
- [x] 16.4 增加显式 debug 配置测试

## 17. Startup SQL Efficiency: 启动共享上下文

- [x] 17.1 设计并实现 StartupContext
- [x] 17.2 调整启动编排函数签名
- [x] 17.3 将 catalog.WithStartupDataSnapshot 接入共享上下文
- [x] 17.4 将 integration.WithStartupDataSnapshot 接入共享上下文
- [x] 17.5 将 jobmgmt.withStartupDataSnapshot 接入共享上下文
- [x] 17.6 增加启动快照复用测试

## 18. Startup SQL Efficiency: 插件同步 no-op fast path

- [x] 18.1 为 registry 和 release metadata 同步补齐写后 snapshot 更新能力
- [x] 18.2 为 manifest menu 和 dynamic route permission menu 增加差异比较函数
- [x] 18.3 为 plugin resource ref 同步前置差异判断
- [x] 18.4 调整 SyncManifest/syncMetadata 编排确保无差异时不写库
- [x] 18.5 增加源码插件 no-op 同步测试
- [x] 18.6 增加差异同步测试

## 19. Startup SQL Efficiency: Cron 启动注册去重

- [x] 19.1 审查 cron.Start、syncBuiltinScheduledJobs、persistentScheduler.LoadAndRegister 查询边界
- [x] 19.2 调整调度器启动扫描条件排除 is_builtin=1
- [x] 19.3 增加或更新调度器测试
- [x] 19.4 评估 monitor-server 首轮采集任务延迟

## 20. Startup SQL Efficiency: 启动摘要与可观测性

- [x] 20.1 增加启动统计采集器
- [x] 20.2 启动完成后输出结构化摘要日志
- [x] 20.3 区分宿主启动编排 SQL 和启动后 SQL
- [x] 20.4 增加启动摘要日志测试

## 21. Startup SQL Efficiency: 回归验证

- [x] 21.1 运行 gofmt 和相关后端单元测试
- [x] 21.2 MySQL 配置下执行 make init/mock/启动 smoke
- [x] 21.3 SQLite 配置下执行启动 smoke
- [x] 21.4 对比优化前后启动 SQL 基线
- [x] 21.5 明确记录 i18n 影响判断
- [x] 21.6 明确记录缓存一致性判断
- [x] 21.7 lina-review 审查通过

## 22. Cross-Platform Dev Commands: 工具骨架与入口

- [x] 22.1 新增 hack/tools/linactl Go CLI 工具目录
- [x] 22.2 将 linactl 加入 go.work，提供双语 README
- [x] 22.3 实现 make 风格 key=value 参数解析
- [x] 22.4 新增根目录 make.cmd

## 23. Cross-Platform Dev Commands: 低风险目标迁移

- [x] 23.1 实现 help 命令
- [x] 23.2 实现 prepare-packed-assets 命令
- [x] 23.3 实现 wasm 命令的插件扫描和构建参数
- [x] 23.4 更新 Makefile 将已迁移目标改为调用 linactl

## 24. Cross-Platform Dev Commands: 开发服务命令迁移

- [x] 24.1 实现 status 命令
- [x] 24.2 实现 dev 命令
- [x] 24.3 实现 stop 命令
- [x] 24.4 更新 dev.mk

## 25. Cross-Platform Dev Commands: 构建/数据库/测试命令迁移

- [x] 25.1 实现 build 命令
- [x] 25.2 实现 image-build/image 包装
- [x] 25.3 实现 init 和 mock 命令
- [x] 25.4 实现 test/test-go/check-runtime-i18n 包装
- [x] 25.5 实现 cli.install/ctrl/dao/enums/service/pb/pbentity 包装

## 26. Cross-Platform Dev Commands: 脚本治理与兼容层收敛

- [x] 26.1 评估并删除或降级不再使用的 .sh 脚本
- [x] 26.2 将根目录 Makefile 已迁移目标收敛为薄包装
- [x] 26.3 将子模块 Makefile 已迁移目标收敛为薄包装
- [x] 26.4 明确 make.cmd、Makefile 与 linactl 的入口优先级

## 27. Cross-Platform Dev Commands: 文档与验证

- [x] 27.1 更新根目录 README/README.zh-CN
- [x] 27.2 更新 hack/tools/README 双语说明
- [x] 27.3 新增 Go 单元测试
- [x] 27.4 新增命令级 smoke 验证
- [x] 27.5 更新 GitHub Actions 增加 Windows runner 验证
- [x] 27.6 Windows CI 覆盖 linactl help/status 和轻量命令
- [x] 27.7 Windows CI 覆盖 cmd.exe 和 PowerShell 的 make 用法
- [x] 27.8 运行 go test 覆盖新增工具
- [x] 27.9 运行 openspec validate
- [x] 27.10 执行 i18n 影响确认
- [x] 27.11 执行缓存一致性影响确认
