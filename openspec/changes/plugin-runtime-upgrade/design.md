## Context

当前插件交付链路分为两类：源码插件通过 `apps/lina-plugins/<plugin-id>` 源码目录参与宿主编译，动态插件通过运行时 artifact 上传或文件投放参与加载。文件更新本身只能代表“发现了新版本”，不能代表数据库中的有效版本、菜单权限、i18n、apidoc、路由、任务、缓存和插件自有数据已经完成升级。

旧的源码插件升级治理把待升级视为启动失败，并把升级描述为开发期升级操作。这与实际能力不一致：当前没有可交付的开发期升级指令；更重要的是，运行时状态和业务数据不能在开发阶段离线完成升级。因此本设计把文件覆盖和运行时升级拆开：开发阶段只覆盖文件，宿主启动只标记状态，运行阶段由管理员在插件管理页显式触发升级。

源码插件存在一个特殊约束：文件覆盖并重新编译/启动后，进程内只有目标版本代码，宿主无法再执行旧版本插件代码。升级回调必须由目标版本插件实现，并通过升级前 manifest 快照理解旧版本状态。动态插件若保留旧 artifact，可以选择读取旧 release 快照，但统一升级编排仍以目标版本升级逻辑为准。

## Goals / Non-Goals

**Goals:**

- 将开发阶段离线文件覆盖与运行时状态/数据升级明确分离。
- 宿主启动时发现插件版本漂移后不阻断启动，而是把插件标记为 `pending_upgrade` 或 `abnormal` 等运行时状态。
- 在插件管理页提供升级入口、升级内容预览、确认执行、失败诊断和异常修复提示。
- 提供统一运行时升级 API，覆盖源码插件和动态插件的有效版本切换、升级 SQL、治理资源同步、缓存失效和集群通知。
- 扩展源码插件生命周期接口，提供可选升级回调，参数包含升级前和目标 manifest 快照。
- 删除旧 lifecycle guard 与 `Can*` 接口，统一到 `Before*`/`After*` 生命周期回调模型。
- 明确分布式部署下升级只能由一个协调者执行，其他节点通过共享状态、修订号和事件广播收敛。

**Non-Goals:**

- 不实现自动升级；宿主启动不得自动执行插件升级 SQL、插件回调或有效版本切换。
- 不支持离线开发工具修改运行时数据库、插件状态或插件自有数据。
- 不支持无人工确认的自动降级；文件版本低于数据库有效版本时先进入异常状态。
- 不把框架自身版本升级、宿主 SQL 重放和业务系统整体升级纳入本插件运行时升级范围。
- 不要求所有插件必须实现自定义升级回调；未实现时宿主仍执行标准升级流程。

## Decisions

### 决策 1：运行时状态独立于安装和启用状态

插件状态需要拆成三类：

- `installed`：是否已安装。
- `enabled`：是否启用。
- `runtimeState`：插件文件与数据库有效状态是否一致。

建议的运行时状态：

```text
normal            数据库有效版本与发现版本一致
pending_upgrade   数据库有效版本低于发现版本
abnormal          数据库有效版本高于发现版本，或 manifest/release 不可安全匹配
upgrade_running   当前节点或集群中已有升级任务执行中
upgrade_failed    最近一次升级失败，需要查看错误并重试或人工修复
```

替代方案是复用 `sys_plugin.status`。不采用该方案，因为 `status` 当前表达启用/禁用，混入升级状态会破坏菜单、路由、权限和租户启用语义。

### 决策 2：启动扫描只标记，不升级也不失败

启动阶段需要扫描源码插件目录和动态插件 artifact/release 元数据，比较数据库有效版本与文件发现版本：

```text
effective < discovered  => pending_upgrade
effective = discovered  => normal
effective > discovered  => abnormal
manifest invalid        => abnormal 或跳过未安装插件，视是否已有有效记录而定
```

宿主不得因为 `pending_upgrade` 拒绝启动。对于 `pending_upgrade` 插件，插件管理页和宿主基础治理接口必须可用；插件业务入口应进入受控状态，避免目标版本代码直接访问旧数据库结构。受控策略可以是暂停插件路由、隐藏/禁用插件菜单、禁用插件 cron，或让运行时路由返回稳定的 `plugin_upgrade_required` 错误。

替代方案是继续 fail-fast。该方案能避免不兼容代码运行，但会导致用户无法打开插件管理页执行升级，无法处理运行时数据升级，因此不采用。

### 决策 3：运行时升级通过管理 API 显式触发

新增升级预览和升级执行能力：

- `GET /plugins/{id}/upgrade/preview`：返回升级前后版本、manifest 差异、SQL 数量、依赖检查、hostServices 变化、菜单/权限/i18n/apidoc 变化摘要和风险提示。
- `POST /plugins/{id}/upgrade`：执行显式升级动作。

`POST` 用于升级执行，因为升级有副作用；预览使用 `GET`，因为只读。执行请求必须至少携带确认标记，可携带 hostServices 授权确认、是否包含 mock 数据等明确选项，但 mock 数据默认不得在升级中自动加载。

### 决策 4：升级编排使用固定顺序和可恢复记录

升级流程建议固定为：

```text
1. 加锁和状态切换为 upgrade_running
2. 重新读取有效 manifest 快照和目标 manifest 快照
3. 校验依赖、反向依赖、框架版本和 hostServices 授权变化
4. 执行 BeforeUpgrade 前置回调，允许插件阻断
5. 执行插件自定义 Upgrade 回调
6. 执行 manifest/sql upgrade SQL 并记录 phase=upgrade
7. 同步菜单、权限、资源引用、i18n、apidoc、路由、cron 等治理资源
8. 切换 sys_plugin.version、release_id 和 release 状态
9. 精确失效插件相关缓存并广播集群事件
10. 执行 AfterUpgrade 事件回调
11. 状态切换为 normal
```

如果任一步骤失败，状态切换为 `upgrade_failed`，保留失败阶段、错误码、错误详情、from/to manifest 快照和可重试信息。第一阶段不做自动回滚，因为插件自定义迁移和 SQL DDL/DML 未必可逆；必须通过重试或人工修复恢复。

### 决策 5：升级回调由目标版本插件实现

源码插件文件覆盖后，宿主只能调用目标版本代码。因此升级接口应由目标版本插件实现：

```go
type SourcePluginUpgrader interface {
    Upgrade(ctx context.Context, req SourcePluginUpgradeRequest) error
}

type SourcePluginUpgradeRequest struct {
    PluginID string
    FromManifest ManifestSnapshot
    ToManifest ManifestSnapshot
    FromVersion string
    ToVersion string
}
```

宿主传入升级前 manifest 快照和目标 manifest 快照。插件如需处理旧字段、旧状态或旧数据，必须在目标版本回调中基于 `FromManifest` 和数据库实际状态完成兼容逻辑。

### 决策 6：统一使用生命周期前置回调

项目为全新项目，不需要保留历史 lifecycle guard 兼容层。旧 `RegisterLifecycleGuard`、`CanUninstall`、`CanDisable`、`CanTenantDelete` 等接口和方法应删除，新模型统一为生命周期回调：

```text
BeforeInstall     可阻断
AfterInstall      事件通知
BeforeUpgrade     可阻断
Upgrade           自定义升级执行
AfterUpgrade      事件通知
BeforeDisable     可阻断
BeforeUninstall   可阻断
BeforeTenantDisable / BeforeTenantDelete / BeforeInstallModeChange 可阻断
```

`Before*` 返回允许/拒绝、稳定 reason key 和错误；`After*` 用于不可阻断的事件通知。插件作者只能通过新生命周期 facade 注册回调，不再暴露旧 Guard/`Can*` 入口。

### 决策 7：集群环境使用主节点/锁协调和作用域失效

`cluster.enabled=false` 时，升级可以使用本地锁和本地缓存失效。`cluster.enabled=true` 时，升级执行必须依赖宿主统一的 `cluster.Service`、分布式锁、共享修订号或等价协调机制，确保同一插件同一时间只有一个升级执行者。

升级成功后，必须按插件 ID、语言、菜单/权限/路由/cron/i18n/apidoc 等作用域精确失效缓存，并广播运行时状态变化。禁止只清理当前节点内存缓存。

## Risks / Trade-offs

- [风险] 源码插件待升级时目标版本代码可能访问旧结构。缓解：待升级状态下暂停或受控暴露插件业务入口，仅允许插件管理和升级 API 工作。
- [风险] 升级失败后数据库可能处于部分迁移状态。缓解：记录失败阶段和迁移账本，提供重试入口，要求 SQL 幂等，第一阶段不承诺自动回滚。
- [风险] 动态插件存在旧 artifact 和新 artifact 双版本运行差异。缓解：有效 release 仍指向旧版本，发现 release 只作为目标版本；升级前不切换有效 release。
- [风险] 生命周期回调模型过大。缓解：接口保持可选和小粒度，先实现安装、升级、禁用、卸载等核心点，并删除旧 Guard/`Can*` 入口以避免双模型并存。
- [风险] 分布式升级竞态导致多节点重复执行。缓解：升级 API 在 cluster 模式下必须获取分布式锁并基于共享状态做幂等检查。
- [风险] 前端状态与后端状态短暂不一致。缓解：列表 API 返回权威 `runtimeState`、`effectiveVersion`、`discoveredVersion` 和 `upgradeAvailable`，操作后刷新列表并监听状态键变化。

## Migration Plan

1. 先删除旧开发期升级入口规范引用，并把现有启动 fail-fast 检查调整为运行时状态标记设计。
2. 增加插件运行时状态模型和 API DTO，扩展插件列表投影。
3. 改造启动扫描流程：同步发现 release，比较有效版本和发现版本，写入运行时状态，不阻断启动。
4. 实现升级预览、升级执行 API 和服务编排。
5. 扩展源码插件生命周期接口，并删除旧 lifecycle guard 与 `Can*` 兼容层。
6. 改造插件管理页动作、状态标签、升级弹窗和异常诊断展示。
7. 补齐缓存失效、集群协调、单元测试、启动包测试和插件管理 E2E。

Rollback 策略：若运行时升级流程在实现阶段风险过高，可先交付状态标记和管理页只读诊断，暂不开放 `POST /plugins/{id}/upgrade`，保持插件业务入口在 `pending_upgrade` 下受控暂停，避免启动失败。

## Cross-Cutting Assessments

- i18n：新增插件状态、升级按钮、升级弹窗、错误码、接口文档和插件 manifest 差异提示，必须同步维护前端运行时语言包、后端 bizerr messageKey、apidoc i18n 和相关插件 manifest/i18n 资源。
- 缓存一致性：升级涉及插件状态、菜单、权限、路由、cron、i18n、apidoc 和 hostServices 授权缓存，必须按作用域精确失效；集群模式下必须广播或共享修订号，禁止仅本地失效。
- 数据权限：升级 API 属于插件治理操作，不读取业务范围数据；自定义插件升级回调如访问宿主数据服务，必须遵循对应 host service 的权限和租户边界，不得绕过数据权限。
- RESTful API：升级预览使用 `GET`，升级执行使用 `POST /plugins/{id}/upgrade`；不得用 `GET` 执行有副作用的升级。
- 开发工具脚本：本变更不新增开发期升级工具；若清理旧升级入口残留，只能删除或修正文档/包装入口，不得新增平台专属脚本。
- E2E：插件管理页新增用户可观察状态和升级操作，必须在宿主级插件框架 E2E 中新增或更新测试，TC ID 从 `TC0236` 起预留。
