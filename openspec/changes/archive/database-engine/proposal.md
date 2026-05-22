## Why

LinaPro在数据库引擎治理上先后经历了两次关键收敛：先将默认数据库与SQL源语法从`MySQL`切换到更接近标准SQL、也更符合企业部署偏好的`PostgreSQL`，再把一度保留用于开发演示的`SQLite`运行链路彻底移除。前一阶段解决了`MySQL`专属语法、`ENGINE=MEMORY`语义和初始化流程深度耦合的问题；后一阶段则解决了“运行时支持矩阵过宽、测试与文档长期维护两套路径、SQL源持续受非生产方言约束”的问题。

最终目标是让`LinaPro`围绕`PostgreSQL 14+`形成单一、清晰、可持续维护的数据库基线：运行时、初始化命令、插件SQL生命周期、缓存与集群协调、CI、镜像交付和双语文档全部以`PostgreSQL`为准，不再保留`MySQL`或`SQLite`作为受支持的运行时数据库。

## What Changes

### 数据库支持范围与配置口径收敛

- 运行时数据库支持统一为`PostgreSQL 14+`，`database.default.link`仅支持`pgsql:`前缀。
- `sqlite:`、`mysql:`和其他未知前缀在方言解析阶段即明确失败，不再进入启动、初始化、mock加载、cluster coordination或业务运行流程。
- 配置模板、镜像运行配置、README与测试文档统一改为`PostgreSQL-only`口径，不再推荐`SQLite`作为开发、演示或容器默认数据库。

### 保留方言抽象层，但只围绕PostgreSQL治理

- 保留`apps/lina-core/pkg/dialect/`作为公共稳定边界，继续统一承载数据库准备、SQL入口、数据库版本查询、表元数据查询、驱动错误分类和驱动/ORM只读SQL分类能力。
- 删除`SQLite`方言实现、DDL转译器、错误分类和启动降级钩子，只保留`PostgreSQL`具体实现以及明确的不支持错误。
- `init`通过连接系统库`postgres`执行`PrepareDatabase`完成建库、删库和重建；`mock`只连接已初始化数据库，不负责准备数据库。

### SQL源、易失性表与缓存语义统一

- 宿主与插件SQL资源统一使用受治理的`PostgreSQL 14+`语法子集编写，移除`MySQL`专属语法遗留，不再为了`SQLite`翻译能力限制运行时支持矩阵。
- `sys_online_session`、`sys_locker`、`sys_kv_cache`作为原易失性表，统一改为`PostgreSQL`普通持久表，依赖`last_active_time`、`expire_time`、`expire_at`和既有TTL清理自然收敛。
- 插件缓存服务继续使用`sqltable`后端与CAS递增语义，在单机默认路径下落到共享`PostgreSQL`表，在集群模式下依赖coordination KV backend；缓存始终视为有损缓存。

### 集群、协调、工具链与发布链路收敛

- `cluster.enabled`只在`PostgreSQL`路径下按配置生效；不支持的数据库方言必须在cluster初始化前失败，不再通过`SQLite`单机锁定或启动告警继续运行。
- 删除`SQLite` smoke、E2E通道、脚本入口、CI workflow输入与package scripts，主干、nightly、release和共享验证模板统一为`PostgreSQL-only`。
- Release镜像与测试门禁继续复用共享测试模板，但不再包含任何`SQLite`验证通道。

### 文档与README治理同步

- 根目录与`apps/lina-core/`的双语README同步更新数据库支持矩阵、初始化说明和外部数据库依赖口径。
- 目录级说明文档继续采用`README.md`与`README.zh-CN.md`双语镜像约定，避免数据库治理相关文档口径漂移。

## Capabilities

### New Capabilities

- `postgresql-only-database-support`：定义运行时、初始化、测试、CI与交付链路只支持`PostgreSQL 14+`的能力边界。
- `cluster-coordination-config`：定义非`PostgreSQL`数据库链接必须在coordination启动前失败的约束。
- `readme-localization-governance`：定义目录级README双语镜像命名与同步治理规则。

### Modified Capabilities

- `project-setup`：项目默认数据库从“`PostgreSQL + SQLite`开发演示模式”收敛为“`PostgreSQL-only`运行时”。
- `database-dialect-abstraction`：保留方言抽象边界，但删除`SQLite`实现与转译路径，只保留`PostgreSQL`实现、元数据查询和只读SQL分类能力。
- `database-bootstrap-commands`：初始化与mock命令仅围绕`PostgreSQL`准备与SQL执行工作流运行，不再支持`SQLite`准备或转译。
- `cluster-deployment-mode`：移除`SQLite`专属单机锁定与启动警告，集群模式只在`PostgreSQL`支持矩阵内生效。
- `sql-source-syntax`：继续以受治理的`PostgreSQL 14+`子集约束SQL源，但不再把`SQLite`运行时支持作为前提。
- `volatile-table-bootstrap`：易失性表自然过期规范收敛到`PostgreSQL`持久表路径。
- `plugin-cache-service`：插件缓存规范移除`SQLite`专属语义，仅保留`PostgreSQL`单机与集群路径。
- `release-image-build`：发布构建与共享测试模板移除`SQLite` smoke门禁。

## Impact

### Affected Code

- `apps/lina-core/pkg/dialect`、`apps/lina-core/pkg/dbdriver`、初始化命令、系统信息和插件数据治理相关代码。
- `apps/lina-core/manifest/config/`、镜像运行配置、`hack/tests`、CI workflow、`linactl`与`make`相关说明入口。
- 宿主与插件SQL治理基线，以及`sys_online_session`、`sys_locker`、`sys_kv_cache`相关自然过期路径。

### Affected Dependencies

- 移除`github.com/gogf/gf/contrib/drivers/sqlite/v2`及其SQLite驱动链路残留。
- 保持`PostgreSQL`驱动与相关测试依赖作为唯一数据库后端依赖。

### Affected Verification

- PostgreSQL编译、单元测试、工具链验证、OpenSpec校验和静态扫描成为默认门禁。
- 原`SQLite` smoke、专属E2E与package script入口被移除，不再作为交付验证路径。

### i18n / Cache Impact

- 本次不新增、修改或删除运行时语言包、插件`manifest/i18n`或apidoc i18n JSON，仅同步更新配置、文档和规范中的数据库支持描述。
- 本次不新增缓存策略，只删除`SQLite`特有的运行分支；单机仍使用SQL表缓存后端，集群仍依赖coordination能力保证一致性。
