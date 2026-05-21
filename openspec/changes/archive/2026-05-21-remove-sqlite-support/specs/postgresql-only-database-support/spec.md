## ADDED Requirements

### Requirement: 运行时数据库支持必须收敛到 PostgreSQL

系统 SHALL 仅支持 PostgreSQL 14+ 作为运行时数据库。`database.default.link` MUST 使用 `pgsql:` 前缀。`sqlite:`、`mysql:` 和其他未知前缀 MUST 在方言解析、启动、初始化和 mock 加载前返回明确的不支持错误，不得静默回退到默认方言。

#### Scenario: PostgreSQL 链接正常解析

- **WHEN** `database.default.link` 以 `pgsql:` 开头
- **THEN** 系统解析为 PostgreSQL 方言
- **AND** 后端启动、`init` 和 `mock` 可继续进入 PostgreSQL 数据库连接流程

#### Scenario: SQLite 链接被拒绝

- **WHEN** `database.default.link` 以 `sqlite:` 开头
- **THEN** 系统返回明确错误说明 SQLite 不再支持
- **AND** 错误消息列出当前支持的前缀仅为 `pgsql:`
- **AND** 系统不得注册 SQLite 驱动、创建 SQLite 文件或执行 SQLite DDL 转译

### Requirement: 交付和验证链路不得包含 SQLite 通道

系统 SHALL 从默认开发、CI、release、nightly、E2E 和测试脚本入口中移除 SQLite 专属验证通道。共享测试套件 MUST 不再暴露 SQLite smoke 输入，workflow MUST 不再调用 SQLite smoke workflow，测试文档 MUST 不再指导运行 SQLite 专属测试。

#### Scenario: 共享 CI 不运行 SQLite smoke

- **WHEN** main、nightly 或 release workflow 调用共享测试套件
- **THEN** 调用参数中不存在 SQLite smoke 开关
- **AND** 共享测试套件不包含 SQLite smoke job
- **AND** 仓库不保留可复用 SQLite smoke workflow

#### Scenario: E2E 脚本不提供 SQLite 命令

- **WHEN** 开发者查看 `hack/tests/package.json`
- **THEN** 不存在 `test:sqlite` 或 `test:sqlite:e2e-smoke` 脚本
- **AND** `hack/tests/e2e` 不包含 SQLite 专属测试用例

### Requirement: 配置和文档必须表达 PostgreSQL-only 支持矩阵

系统 SHALL 在配置模板、镜像运行配置、README 和测试文档中将数据库支持矩阵表达为 PostgreSQL-only。文档不得继续推荐 SQLite 作为开发、演示或容器默认数据库。

#### Scenario: 配置模板只展示 PostgreSQL 链接

- **WHEN** 开发者查看 `manifest/config/config.template.yaml`
- **THEN** 数据库链接示例只展示 PostgreSQL 14+ `pgsql:` 链接
- **AND** 不包含 SQLite 文件链接示例

#### Scenario: 镜像运行配置使用 PostgreSQL

- **WHEN** 发布镜像读取默认 `.github/image/config.runtime.yaml`
- **THEN** `database.default.link` 使用 `pgsql:` 链接
- **AND** 配置注释不再声明 SQLite 会强制关闭集群模式
