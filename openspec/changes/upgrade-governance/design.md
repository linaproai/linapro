## Context

当前 `upgrade-governance` 已经完成了一套可执行的**框架源码升级**方案：开发者在本地显式执行 `make upgrade`，升级工具比较当前与目标框架版本，覆盖代码后从第一条宿主 SQL 开始全量重放。

但插件升级治理仍然不完整，尤其是源码插件：

- **动态插件**已经具备 staged artifact / active release / upgrade migration 的运行时机制；
- **源码插件**仍然只有“发现、安装、卸载”语义，没有正式的 upgrade 入口；
- 当前源码扫描会把注册表版本直接写成源码里的新版本，导致“当前生效版本”和“源码树发现版本”混在一起；
- 如果开发者在源码插件版本提升后直接启动宿主，系统会进入“编译产物已经升级，但治理状态和数据库尚未完成升级”的不一致状态。

由于源码插件与宿主一起编译交付，它的升级语义更接近框架升级，而不是动态插件的运行时热升级。因此本轮设计决定：**源码插件升级必须在开发阶段显式完成，宿主运行时只负责校验并阻断启动，不负责自动升级。**

同时，团队已经明确本轮**不考虑回滚机制**。因此源码插件与动态插件的升级路径都只要求：升级计划清晰、执行过程显式、失败状态可见；自动回退与通用 rollback 作为后续迭代再处理。

## Goals / Non-Goals

**Goals**
- 扩展 `make upgrade`，让开发态升级入口同时支持框架升级和源码插件升级。
- 为源码插件建立清晰的“当前生效版本 / 源码发现版本”分离模型。
- 为源码插件提供显式升级流程，复用现有 release / migration / resource ref 治理表。
- 在宿主启动阶段阻断未完成源码插件升级的场景，确保“启动前升级完成”。
- 明确动态插件升级继续走运行时 upload + install/reconcile，不纳入 `make upgrade`。
- 补齐当前变更中的 OpenSpec 文档、任务清单与后续文档更新要求。

**Non-Goals**
- 本轮不实现插件回滚、自动 rollback SQL 或 release 自动回退。
- 本轮不实现业务系统运行时升级平台。
- 本轮不为插件新增独立的 `manifest/sql/upgrade/` 目录；源码插件升级继续复用既有安装 SQL 资产并以 `phase=upgrade` 记账。
- 本轮不把动态插件升级改造成开发态命令，也不替换其现有 runtime reconciler 模型。

## Decisions

### 1. `make upgrade` 扩展为统一的开发态升级入口

保留 `make upgrade` 作为唯一的开发阶段升级入口，但增加明确的范围参数，不再默认只表示框架升级。

建议命令形态：

```bash
make upgrade confirm=upgrade scope=framework target=v0.5.0
make upgrade confirm=upgrade scope=source-plugin plugin=plugin-demo
make upgrade confirm=upgrade scope=source-plugin plugin=all
make upgrade confirm=upgrade scope=source-plugin plugin=plugin-demo dry_run=1
```

规则：
- `scope=framework`：沿用当前框架升级语义；
- `scope=source-plugin`：执行源码插件升级；
- `plugin=` 只在 `scope=source-plugin` 下生效，支持单个插件 ID 或 `all`；
- `dry_run` 在两种范围下都必须支持。

本轮沿用现有仓库根目录开发态升级工具实现，不额外引入第二套命令入口，避免框架升级与源码插件升级分别维护不同的安全校验和计划输出逻辑。

同时，开发态升级工具的目录命名统一为 `hack/upgrade-source`，并约束根目录只保留 `main.go` 用于进程启动；框架升级与源码插件升级的具体逻辑分别收敛到 `internal/frameworkupgrade` 与 `internal/sourceupgrade` 组件中，避免仓库根目录升级工具继续堆积实现细节。

宿主侧的源码插件升级治理也必须保持同样的分层：真实的升级发现、升级执行与启动前 fail-fast 校验逻辑收敛到独立宿主组件中，由插件主服务仅作为组合门面暴露给宿主启动流程；对 `hack/upgrade-source` 开发态工具开放的 `pkg` 层只保留稳定 contract 与必要 facade，避免把真实治理逻辑散落在命令组件、`pkg` 包装层和 `plugin` 主包实现文件中。

### 2. 源码插件升级必须在开发阶段显式完成

源码插件与宿主一起编译交付，因此它不是运行中可热替换的独立产物。若源码插件版本从 `v0.1.0` 提升到 `v0.5.0`，最合理的升级时机不是宿主启动后自动补救，而是**开发者在启动前显式执行升级命令**。

因此本轮约束为：
- 源码插件升级只能通过开发态 `make upgrade scope=source-plugin ...` 触发；
- 宿主启动阶段不得自动执行源码插件升级 SQL；
- 若检测到存在待升级源码插件，宿主启动必须失败并提示正确命令。

这样可以避免宿主在半升级状态下启动，也能保持开发阶段操作边界与框架升级一致。

### 3. 源码插件必须区分“当前生效版本”与“源码发现版本”

当前源码扫描会直接把 `sys_plugin.version` 更新成 `plugin.yaml` 里的版本，这会破坏升级治理语义。为了解决这个问题，本轮定义：

- `sys_plugin.version` 与 `sys_plugin.release_id` 只表示**当前生效版本**；
- `sys_plugin_release` 中可以存在更高版本的源码插件 release，但在显式执行升级前，该 release 只能处于 `prepared` 或等价的“已发现未应用”状态；
- 源码扫描发现更高版本时，只 upsert 对应 `sys_plugin_release` 的快照和校验信息，不得直接覆盖 `sys_plugin.version`。

这样，数据库中就能同时表达：
- 当前正在使用哪个源码插件版本；
- 源码树里已经发现了哪个更高版本；
- 该更高版本是否已经正式升级生效。

### 4. 源码插件升级复用现有 release / migration / resource ref 治理表

本轮不新增插件升级专用元数据表，而是复用已有治理模型：

- `sys_plugin_release`：记录源码插件 release 快照；
- `sys_plugin_migration`：记录 upgrade phase 的 SQL 执行结果；
- `sys_plugin_resource_ref`：记录升级后的菜单、权限、前端页面和其他治理资源引用。

源码插件升级成功后：
- 当前 active release 从旧版本切换到新版本；
- `sys_plugin.version` 与 `release_id` 更新为新 release；
- 旧 release 保留历史快照，但本轮不实现 rollback。

### 5. 源码插件升级继续复用既有 `manifest/sql/` 资产

本轮不新增 `manifest/sql/upgrade/`。源码插件在执行升级时，仍然解析插件现有的安装 SQL 资产，但迁移账本中的 `phase` 必须记录为 `upgrade`。

这样做的原因是：
- 当前插件迁移执行器已经支持 `MigrationDirectionUpgrade`；
- 现有插件 SQL 规范已经要求幂等；
- 本轮先解决“有没有正式升级路径”的问题，不再同时扩充新的升级 SQL 目录约定。

### 6. 宿主启动前必须校验源码插件是否存在待升级版本

宿主启动阶段仍然需要扫描源码插件，以便同步最新的源码发现结果和 release 快照。但扫描之后，必须立即执行**待升级校验**：

- 若某源码插件已安装，且数据库当前生效版本低于源码树发现版本；
- 则宿主启动必须失败；
- 错误信息必须明确输出插件 ID、当前版本、发现版本，以及建议执行的 `make upgrade` 命令。

该校验必须发生在源码插件 HTTP 路由注册、插件 cron 注册和宿主正式对外提供服务之前。

### 7. 动态插件升级继续保留运行时模型，但需明确边界

动态插件当前已经具备：
- staged artifact 与 active release 分离；
- 上传高版本后作为待切换 release 暂存；
- install/reconcile 时检测版本漂移并执行 upgrade phase；
- 失败时保留失败 release 状态。

因此本轮不改变动态插件升级核心机制，只补充边界约束：
- `make upgrade` 不处理动态插件；
- 动态插件升级仍然由 upload + install/reconcile 驱动；
- 相关 OpenSpec 文档、命令帮助和接口说明需要明确这一点，避免将“插件升级”误解为统一使用开发态命令。

### 8. 本轮不实现回滚能力

团队已明确本轮不考虑回滚，因此本轮不做：
- source plugin rollback 命令；
- dynamic plugin 自动回退编排增强；
- rollback SQL 目录；
- 框架/插件升级失败后的自动恢复。

本轮失败处理策略统一为：
- 停止后续步骤；
- 保留失败状态、错误日志和 migration 记录；
- 由开发者人工修复后重新执行。

## Risks / Trade-offs

- **源码插件 release 快照仍引用同一源码目录**：由于源码插件与宿主一起演进，历史版本源码不会像动态插件 release 那样保留独立产物路径；本轮接受这一限制，以保证开发态升级流程先落地。
- **启动阻断会提高本地开发摩擦**：但这比带着待升级源码插件强行启动更安全，也更符合“启动前升级完成”的要求。
- **不做 rollback 会让失败恢复更依赖人工处理**：这是本轮有意收敛的边界，后续再补充回滚设计。
- **动态插件与源码插件仍采用不同升级触发方式**：这不是缺陷，而是两类插件交付形态不同导致的合理差异；本轮目标是把边界讲清而不是强行统一成同一流程。

## Migration Plan

1. 扩展现有 `make upgrade` 与开发态升级工具参数模型，支持 `scope=framework|source-plugin` 与 `plugin=<id|all>`。
2. 调整源码插件扫描与 release 同步逻辑，拆分当前生效版本与源码发现版本。
3. 实现源码插件升级计划与执行流程，复用 `phase=upgrade` 迁移记账。
4. 在宿主启动流程加入源码插件待升级校验，并在发现漂移时 fail fast。
5. 补充单元测试/集成测试，覆盖版本漂移、命令 dry-run、单插件升级、批量升级和启动阻断场景。
6. 更新相关文档，包括 OpenSpec 规格、任务清单、命令帮助，以及后续实现阶段所需的 `README.md` / `README.zh_CN.md` 与接口说明。

## Open Questions

- `scope=source-plugin plugin=all` 在遇到某个插件升级失败时，是否立即停止后续插件升级；本轮倾向于立即停止，保持与框架 SQL 执行“遇错即停”一致。
- 后续是否需要把动态插件当前已有的 upgrade 机制单独整理成独立 capability，并补齐与源码插件对照的治理文档；本轮先在当前变更中明确边界，不再拆新变更。
