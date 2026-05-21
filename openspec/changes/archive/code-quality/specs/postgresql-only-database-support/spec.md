## ADDED Requirements

### Requirement: 运行时数据库支持必须收敛到 PostgreSQL

系统 SHALL 仅支持 PostgreSQL 14+ 作为运行时数据库。`database.default.link` MUST 使用 `pgsql:` 前缀。`sqlite:`、`mysql:` 和其他未知前缀 MUST 在方言解析、启动、初始化和 mock 加载前返回明确的不支持错误。

### Requirement: 交付和验证链路不得包含 SQLite 通道

系统 SHALL 从默认开发、CI、release、nightly、E2E 和测试脚本入口中移除 SQLite 专属验证通道。

### Requirement: 配置和文档必须表达 PostgreSQL-only 支持矩阵

系统 SHALL 在配置模板、镜像运行配置、README 和测试文档中将数据库支持矩阵表达为 PostgreSQL-only。
