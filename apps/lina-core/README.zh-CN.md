# LinaPro Core Host

`apps/lina-core` 是 `LinaPro` 的 `GoFrame` 宿主服务。它负责提供可复用的后端模块、共享治理能力、插件运行时支持，以及默认管理工作台所消费的公开 / 受保护 `API`。

## 职责边界

- 提供系统管理、插件治理和平台公共能力对应的宿主 `RESTful API`。
- 负责数据库迁移、`Seed` 数据、运行时参数，以及生成的 `DAO` / `DO` / `Entity` 工件。
- 加载源码插件与动态 `Wasm` 插件，并将插件生命周期与宿主治理数据联动。
- 运行宿主级定时任务，以及面向管理工作台的持久化任务调度子系统。

## 定时任务管理

定时任务子系统属于宿主核心治理能力，因此落在 `lina-core` 中统一实现。

- `internal/service/jobhandler`：处理器注册表、宿主内置处理器装配、插件生命周期联动注册，以及参数 `Schema` 校验。
- `internal/service/jobmgmt`：任务分组、持久化任务、执行日志、`Cron` 预览、日志保留清理，以及内置任务约束规则。
- `internal/service/jobmgmt/scheduler`：持久化任务装载、`gcron` 注册、并发守卫、执行分发与取消控制。
- `internal/service/jobmgmt/shellexec`：受控 `Shell` 执行、超时处理、输出截断与手动终止支持。
- `internal/service/cron`：宿主启动后的统一入口，负责注册宿主定时任务并装载持久化任务。

## 关键目录

```text
api/                API DTO 与路由契约
internal/cmd/       服务启动与路由装配
internal/controller/ HTTP 控制器
internal/service/   业务服务与定时调度编排
manifest/config/    宿主运行配置
manifest/sql/       宿主 SQL 迁移与 Seed 数据
manifest/sql/mock-data/ 宿主可选`mock`/演示 SQL 资源
pkg/                宿主与插件共享的稳定公共包
```

## 常用命令

```bash
go run main.go
make build
make dao
make ctrl
make init confirm=init
```

发布构建可使用`make build platforms=linux/arm64`交叉编译单一目标，或在仓库根目录发布多平台镜像：

```bash
make image platforms=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=v0.6.0 push=1
```

## 数据库配置

宿主运行时数据库方言只从配置文件中的`database.default.link`读取。`PostgreSQL 14+`是默认生产数据库。运行`make init`或`make dev`之前，请先准备`PostgreSQL`；这些命令不会启动或管理数据库。

本地开发可使用以下容器：

```bash
docker run --name linapro-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=linapro \
  -p 5432:5432 \
  --health-cmd pg_isready \
  --health-interval 10s \
  --health-timeout 5s \
  --health-retries 5 \
  -d postgres:14-alpine
```

默认运行时连接为：

```yaml
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
```

`make init`是运维初始化命令，会使用配置中的数据库账号。该账号必须具备连接系统库、创建和删除目标数据库、终止目标库连接、建表、建索引、写入注释和写入`Seed`数据的权限。权限不足会直接失败，运行时不会提供低权限初始化兜底。

使用外部托管`PostgreSQL`，例如`RDS`或阿里云`PolarDB`时，请将`database.default.link`指向供应商端点，并使用具备上述权限的账号执行初始化。

如需单节点开发演示，可将链接改为`SQLite`：

```yaml
database:
  default:
    link: "sqlite::@file(./temp/sqlite/linapro.db)"
```

`SQLite`模式仅支持单节点，会自动强制`cluster.enabled=false`，不支持生产部署。

## 集群协调

单机部署保持轻量本地模式：不需要`Redis`，不会连接`Redis`，并在适用场景继续使用`PostgreSQL`和进程内缓存协调。

集群部署必须显式配置`Redis`协调后端：

```yaml
cluster:
  enabled: true
  coordination: redis
  redis:
    address: "127.0.0.1:6379"
    db: 0
    password: ""
    connectTimeout: 3s
    readTimeout: 2s
    writeTimeout: 2s
```

当前后端仅支持`redis`。`cluster.coordination`使用稳定的标量配置，便于后续扩展其他协调后端，而不暴露按存储类型拆分的协调开关。`Redis`集成测试默认不启用；运行依赖真实`Redis`的`Go`测试前，请设置`LINA_TEST_REDIS_ADDR`，例如`LINA_TEST_REDIS_ADDR=127.0.0.1:6379`。

## 源码插件升级

源码插件现在采用显式的开发阶段升级流程，不再允许在宿主启动期间静默切换版本。

- 宿主启动前会先扫描源码插件；如果某个已安装源码插件的 `plugin.yaml` 发现版本高于当前生效的 `sys_plugin.version`，启动会直接失败，直到版本差异处理完成。
- 通过受支持的插件工作区更新流程刷新源码插件代码，再启动宿主。
- 宿主启动前需要处理所有已安装且发现了更高版本的源码插件。
- 动态插件继续使用现有的运行时 `upload + install/reconcile` 升级模型。

## 相关入口

- `manifest/sql/011-scheduled-job-management.sql`：定时任务的表结构、种子数据、菜单权限与字典定义。
- `internal/cmd/cmd_http.go`：任务、分组、日志、处理器接口的宿主装配入口。
- `internal/service/cron/cron.go`：宿主定时任务启动入口。
