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

发布构建可使用`make build os=linux arch=arm64`交叉编译单一目标，或在仓库根目录发布多平台镜像：

```bash
make image platform=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=v0.6.0 push=1
```

## 数据库配置

宿主运行时数据库方言只从配置文件中的 `database.default.link` 读取。MySQL 仍是默认生产数据库。如需无需 MySQL 的演示或本地测试模式，可将链接改为 SQLite，例如：

```yaml
database:
  default:
    link: "sqlite::@file(./temp/sqlite/linapro.db)"
```

SQLite 模式仅支持单节点，会自动强制 `cluster.enabled=false`，不支持生产部署。

## 源码插件升级

源码插件现在采用显式的开发阶段升级流程，不再允许在宿主启动期间静默切换版本。

- 宿主启动前会先扫描源码插件；如果某个已安装源码插件的 `plugin.yaml` 发现版本高于当前生效的 `sys_plugin.version`，启动会直接失败，直到显式升级命令执行完成。
- 通过 AI 工具调用 `lina-upgrade` 技能升级单个插件，例如：`upgrade source plugin plugin-demo-source`。
- 通过同一技能执行 `upgrade all source plugins`，可批量处理全部已安装且发现了更高版本的源码插件。
- 动态插件继续使用现有的运行时 `upload + install/reconcile` 升级模型，不纳入开发态升级技能。

## 相关入口

- `manifest/sql/014-scheduled-job-management.sql`：定时任务的表结构、种子数据、菜单权限与字典定义。
- `internal/cmd/cmd_http.go`：任务、分组、日志、处理器接口的宿主装配入口。
- `internal/service/cron/cron.go`：宿主定时任务启动入口。
