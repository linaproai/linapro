## MODIFIED Requirements

### Requirement: 宿主必须通过统一的方言抽象层收敛数据库引擎差异

系统 SHALL 在 `apps/lina-core/pkg/dialect/` 提供公共稳定的 `Dialect` 接口与方言辅助能力作为数据库引擎差异的唯一收敛点。当前唯一支持的具体数据库方言 SHALL 为 PostgreSQL。所有数据库引擎相关的差异化行为必须通过该包暴露，业务模块不得在自身代码路径中出现数据库引擎判断。

#### Scenario: 业务模块不感知数据库引擎差异

- **当** 业务模块通过 DAO 层执行查询、写入、更新、删除操作时
- **则** 业务代码不包含针对数据库引擎的分支判断
- **且** 同一份业务代码只承诺在 PostgreSQL 支持矩阵下运行

#### Scenario: 方言根据数据库链接前缀自动分发

- **当** 配置文件 `database.default.link` 以 `pgsql:` 开头时
- **则** `dialect.From(link)` 返回 PostgreSQL 方言实例
- **且** `sqlite:`、`mysql:` 和其他未识别前缀 MUST 被识别为不支持的方言并返回明确错误

#### Scenario: PostgreSQL 方言 DDL 入口为无操作

- **当** PostgreSQL 方言实例的 `TranslateDDL` 被调用时
- **则** 返回值与输入字节级别完全一致
- **且** 不返回错误

#### Scenario: 方言必须暴露数据库准备入口

- **当** 需要执行数据库准备时
- **则** 调用方通过 `Dialect.PrepareDatabase` 完成 PostgreSQL 数据库准备工作
- **且** PostgreSQL 通过连接系统库 `postgres` 执行 `pg_terminate_backend` + `DROP DATABASE IF EXISTS` + `CREATE DATABASE`

#### Scenario: 驱动与 ORM 只读 SQL 分类由 dialect 公共包提供

- **当** 治理层需要允许驱动或 ORM 发出的表元数据读 SQL 时
- **则** 调用方通过 `Dialect.ClassifyReadSQL(sql)` 获取语义分类
- **且** 治理层不得直接硬编码 PostgreSQL 特有 SQL 片段
