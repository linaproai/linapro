## Context

当前插件升级逻辑已经具备统一入口的外部形态，但内部 owner 仍分散：

- `plugin`根门面持有 runtime upgrade preview/execute 编排和分布式锁。
- `internal/sourceupgrade`负责源码插件升级状态扫描、执行、失败记账和 release 提升。
- `internal/runtimeupgrade`只提供 dynamic runtime upgrade preview 的纯函数规划。
- `internal/store`负责 runtime upgrade 状态投影和失败诊断读取。
- `internal/runtime`负责 dynamic release 切换、reconciler upgrade 和失败后旧发布保留。

这种形态已经避开了构造 setter 和反向持有门面，但仍保留 source/dynamic 两套升级骨架。方案 D 的目标是在 C 阶段确立 lifecycle 和统一缓存发布入口后，把升级治理收敛到一个 owner，避免治理守卫、失败账本、缓存失效和 release 提升继续散落在多个包。

## Goals / Non-Goals

**Goals:**

- 新建`internal/upgrade`作为 source/dynamic 升级编排 owner。
- 删除`internal/sourceupgrade`和`internal/runtimeupgrade`两个平行包，根门面升级文件只保留平台治理守卫、锁入口协调和委托。
- 将 runtime upgrade preview、source upgrade status、execute 和失败诊断统一到同一服务契约与测试边界。
- 将`sys_plugin_migration`升级失败诊断收敛为一套读写约定，保持既有 phase、message key、fallback 和可诊断字段。
- 保持`publishPluginChange`作为升级成功和失败后派生缓存发布的唯一入口。
- 用静态边界测试阻断`sourceupgrade`、`runtimeupgrade`和根门面升级长流程回流。

**Non-Goals:**

- 不修改 HTTP API、DTO、OpenAPI 文案或前端页面。
- 不新增、删除或迁移数据库表和字段。
- 不改变插件 manifest wire、WASM guest 协议、host service wire 字符串或动态 artifact 格式。
- 不拆分`runtime/route.go`，不继续改造 WASM host service 分发层；这些仍属于后续可选 E 阶段。
- 不把插件升级变成自动级联升级；依赖插件仍必须由管理员显式处理。

## Decisions

### D1：以`internal/upgrade`承载统一升级编排

`upgrade.Service`提供 source upgrade status、runtime upgrade preview 和 execute 等能力。根门面保留公开`Service`契约、平台治理守卫、必要分布式锁入口和兼容类型别名，但不再持有 source/dynamic 分叉长流程。

替代方案是只把`sourceupgrade`改名为`upgrade`，再让 dynamic 继续留在根门面。该方案迁移成本低，但无法消除`runtimeupgrade`纯函数、根门面 execute 分叉和 source 再入公开方法，不能满足方案 D 的“升级逻辑单包闭环”验收。

### D2：统一 preview 和 execute 的共享骨架

升级服务内部使用同一套规划输入表达有效发布、目标发布、插件类型、依赖状态、host service diff、SQL 摘要和风险提示。source/dynamic 差异作为策略函数或窄执行器存在：

- source 策略负责源码发现版本、源码 release 提升、源码 lifecycle callback 和源码治理资源同步。
- dynamic 策略负责 dynamic authorization 持久化、runtime upgrade request、release artifact 和 rollback 语义。

共享骨架负责依赖校验、反向依赖保护、分布式锁内状态复查、失败记账、release 提升后缓存发布和结构化返回。

### D3：失败诊断只保留一套`sys_plugin_migration`约定

升级失败阶段统一使用`plugintypes.RuntimeUpgradeFailurePhase`和 migration ledger 中的 upgrade phase 映射。source/dynamic 不再分别维护两套失败码和阶段归一化逻辑。`store`可以继续作为底层治理投影 owner，但升级失败诊断构造逻辑归属`upgrade`或升级专属 store helper，不再由根门面、sourceupgrade 和 runtimeupgrade 分别拼装。

该决策不新增 SQL schema，也不改变现有`sys_plugin_migration`数据分类。

### D4：治理守卫只在门面入口执行一次

公开升级入口继续由根门面执行`ensurePlatformGovernance`。门面进入`upgrade.Service`后，内部 source/dynamic 策略不得再调用公开门面方法，也不得再次执行平台治理守卫。启动期或内部治理路径如需升级状态计算，只能调用无副作用查询或明确的内部方法，并在任务记录中说明租户上下文和治理边界。

### D5：缓存发布继续经`publishPluginChange`

升级服务不直接操作 frontend、WASM、i18n 或 management list cache。它通过构造函数接收 cache publisher 窄契约，成功或失败路径都发布包含`pluginID`、`pluginType`和 reason 的插件变化。该 publisher 仍由根门面绑定到`publishPluginChange`，继续复用`plugin-runtime`revision controller。

### D6：静态治理测试固化范围

扩展`plugin_boundary_test.go`：

- 阻断生产代码 import`internal/sourceupgrade`和`internal/runtimeupgrade`。
- 阻断新建`internal/sourceupgrade`、`internal/runtimeupgrade`目录。
- 阻断根门面升级文件直接调用`UpgradeSourcePlugin`形成再入。
- 要求`internal/upgrade`存在并由`plugin.New()`显式构造。
- 要求根门面升级文件不直接导入`runtimeupgrade`或`sourceupgrade`，不承载长流程。

## Risks / Trade-offs

- [Risk] 升级路径迁移同时触碰 source 和 dynamic 两条高风险流程。→ Mitigation：按 preview/status、source execute、dynamic execute、失败诊断、缓存发布五批迁移；每批运行对应窄测试和插件全包测试。
- [Risk] 消除 source 再入公开方法后可能改变平台治理守卫执行次数。→ Mitigation：新增或调整测试断言升级入口只执行一次治理守卫，内部 source 策略仍保留租户上下文和错误语义。
- [Risk] 失败诊断 owner 迁移可能改变 message key 或 fallback。→ Mitigation：迁移前盘点所有失败码、phase 和 message key；迁移后用现有测试和新增断言固定原值。
- [Risk] cache publisher 从根门面传入`upgrade`后可能遗漏失败路径发布。→ Mitigation：扩展静态边界测试和升级失败测试，覆盖 source callback、SQL、release switch、cache invalidation 等失败 phase。
- [Risk] `store`与`upgrade`之间的职责边界可能过细。→ Mitigation：`store`只保留治理数据读写和稳定投影，升级业务状态机、phase 归一化和失败语义归属`upgrade`。

## Migration Plan

1. 建立`internal/upgrade`主契约和只读 status/preview 能力，先包装现有 source/dynamic 逻辑并保持测试通过。
2. 迁移 source upgrade status 和 execute 编排，删除根门面公开方法再入路径。
3. 迁移 runtime upgrade preview/execute 编排和`runtimeupgrade`纯函数，保留 dynamic runtime 的底层 release/reconciler 能力。
4. 收敛失败诊断、release 提升和缓存发布，补齐单一账本约定记录。
5. 删除`internal/sourceupgrade`和`internal/runtimeupgrade`目录，扩展静态边界测试。
6. 运行插件全包、`i18n`、`cachecoord`、`internal/cmd`和 OpenSpec 严格校验。

回滚策略为`git revert`当前 OpenSpec 变更对应提交；本变更不涉及数据库 schema 或外部 wire 格式迁移。

## Open Questions

- 无阻塞问题。实现中如果发现`store`现有 runtime upgrade projection 与`upgrade`业务状态机边界不清，优先保持`store`为数据 owner，业务解释逻辑迁入`upgrade`。
