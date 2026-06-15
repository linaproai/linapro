## Context

`localdocs/plugin-scale-performance-optimization-plan.md`确认当前插件运行时的骨架设计基本正确：动态路由前缀匹配为`O(1)`，host call 注册表为只读 map，store 启动快照已把部分`N+1`收敛为批量读取，reconciler 也已按修订号跳过无变化扫描。剩余高价值瓶颈集中在读模型缓存覆盖不完整和失效粒度过粗：

- 动态路由请求在进入 guest 前会重复解析同一动态插件 artifact 和 release manifest。
- 插件详情、依赖检查、OpenAPI 投影和 hook 分发会反复全量`ScanManifests`。
- 单插件启停或升级会触发全部`WASM`编译缓存失效，且当前编译过程持有全局缓存写锁。
- 依赖检查在列表投影中存在按插件遍历全部插件快照的超线性成本。

本变更只处理清单读模型、`WASM`缓存和治理读投影性能，不处理动态路由认证、session 续期、guest 并发护栏和 datahost 微优化。

## Goals / Non-Goals

**Goals:**

- 让动态插件请求热路径在缓存命中时不执行完整 artifact 解析。
- 让单插件治理变更只失效受影响插件或 artifact 的派生缓存。
- 让插件详情、依赖检查、OpenAPI 投影和 hook 分发复用有界清单快照。
- 保持`plugin-runtime`域缓存一致性语义，集群模式不得退化为仅本节点可见的本地缓存。
- 通过单元测试、静态检索或计数桩验证解析次数、扫描次数和编译次数随插件数增长保持有界。

**Non-Goals:**

- 不拆分 manifest 与 artifact 存储。
- 不改变`plugin.yaml`、host service 协议或动态插件交付形态。
- 不新增外部基础设施依赖。
- 不承诺启用`wazero.CompilationCache`；该项只允许在基准证明收益后作为后续增强。
- 不处理权限快照、session 有效性校验、guest 并发护栏和 datahost 表契约缓存。

## Decisions

### 1. 在`catalog`和`store`边界内实现清单读模型缓存

新增缓存作为插件清单事实源的内部实现细节，不对外暴露缓存结构。调用方继续通过`catalog`读取清单，通过`store`读取 release 快照。这样不会把缓存状态泄漏给插件根门面、runtime 或管理投影。

缓存条目按权威数据源拆分：

| 条目 | 键 | 一致性 |
|------|----|--------|
| 源码插件 manifest | `pluginID` | 编译期嵌入，进程生命周期内不可变 |
| 动态 desired manifest | artifact 路径加文件`size`和`mtime` | 文件`stat`守卫加显式按插件失效 |
| release manifest 快照 | `pluginID`、`releaseID`、`checksum` | release 行写后不改，卸载或 LRU 淘汰 |

替代方案是让每个调用点自行加缓存。该方案会复制失效逻辑，并增加多个缓存事实源；因此拒绝。

### 2. 请求级复用 manifest，但不让 manifest 缓存承担执行授权

`matchDynamicRoute`定位候选插件时可以把 manifest 快照挂入请求上下文，`prepareDynamicRouteRuntime`复用该快照，避免同一请求二次解析。进入 guest 前仍必须确认`plugin-runtime` freshness、enabled snapshot 和 active release 状态；缓存只能降低解析成本，不能替代执行安全检查。

### 3. `WASM`编译缓存按插件和 artifact 精细失效

已知`pluginID`或 active release artifact 的治理变更必须调用按路径失效，不能走全量失效。集群 peer 只收到修订号时，通过清单缓存和注册表投影做差异对账，失效不再活跃或文件状态变化的 artifact。

编译缓存 map 的锁只保护条目读写；实际`os.ReadFile`和`CompileModule`通过 per-artifact single-flight 在锁外执行。这样同 artifact 并发只编译一次，不同 artifact 互不阻塞。

### 4. 治理读投影共享快照和索引

管理列表、详情、依赖检查、OpenAPI 投影和 hook 分发应复用同一批 manifest、runtime 和 store 投影。反向依赖索引在快照构建时生成，复杂度从按插件遍历全部插件收敛为`O(V+E)`。

OpenAPI 文档缓存键必须包含`plugin-runtime`修订号、locale 和运行时翻译包版本。该缓存属于展示/运维投影，不作为运行时授权依据。

### 5. 缓存一致性七要素

| 要素 | 决策 |
|------|------|
| 权威数据源 | 源码插件嵌入资源、动态插件 artifact 文件、`sys_plugin_release`发布快照 |
| 一致性模型 | 复用`plugin-runtime`修订号，动态文件补充`stat`守卫 |
| 失效触发点 | 治理写成功后`publishPluginChange`按插件失效并发布修订号；peer 观察修订号后差异对账 |
| 跨实例同步 | `cluster.enabled=false`使用本地 revision；`cluster.enabled=true`使用 Redis revision/event |
| 最大可接受陈旧 | 展示类投影可按`plugin-runtime`域预算短暂陈旧；动态执行路径不可在 freshness 不可确认时继续放行 |
| 故障降级 | 缓存未命中回源解析；freshness 不可确认时按 conservative-hide 或结构化错误处理 |
| 恢复路径 | 修订号请求路径检查、事件监听、低频 safety sweep 和进程重启重建 |

## Risks / Trade-offs

- 共享 manifest 被调用方原地修改 → 实施时静态检索消费点，必要时对高风险路径保留显式拷贝，并增加防御性测试。
- 文件`mtime`精度导致同大小同时间替换未被发现 → 治理写路径同时执行显式按插件失效，`stat`守卫只作为稳态读优化和兜底。
- 差异对账遗漏回滚或路径复用边界 → 对账以 active release 集合、checksum/generation 和文件状态共同判断，旧 lease 允许在途请求自然结束。
- OpenAPI 投影缓存返回旧语言内容 → 缓存键包含 locale 和运行时翻译包版本，并在 i18n bundle 变化后失效。
- 新增缓存增加内存占用 → release manifest 缓存使用 LRU 上限，动态 manifest 缓存避免长期保留不需要的前端大字节资产。

## Migration Plan

该变更不涉及数据库迁移和外部依赖。实施顺序为：先建立清单缓存和测试计数桩，再迁移动态路由与单插件读取调用点，然后收敛`WASM`编译缓存失效和锁粒度，最后迁移依赖检查与 OpenAPI 投影缓存。每一步都保持旧回源路径可用，出现问题时可回退到直接解析或全量失效的保守行为。

## Open Questions

- OpenAPI 文档投影缓存是否纳入本批实现，还是在核心热路径完成后作为可选任务后移。当前任务将其列为低优先级子任务，可在实施时根据风险裁剪。
- 是否需要把 30 插件性能基准脚本固化为`hack/`工具。如果固化，需要额外读取并执行`dev-tooling.md`规则；当前默认仅用单元测试计数桩和可选本地基准记录作为验证。
