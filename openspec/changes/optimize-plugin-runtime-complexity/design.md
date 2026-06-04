## Context

本次变更来自对 `apps/lina-core/internal/service/plugin` 和 `apps/lina-core/pkg/plugin` 的代码行数与复杂度审查。两个目录合计约 9.3 万行 Go，其中约三分之一是测试；生产代码高行数主要集中在动态插件运行时、WASM host service、`pluginbridge` 协议样板、源码插件生命周期 API 和跨子组件 wiring。

这些职责大多属于 `apps/lina-core` 作为全栈框架核心宿主服务的真实边界，不能简单按行数删除。但审查发现以下问题需要通过正式变更跟进：

- `ListRuntimeStates` 在插件 registry 循环内逐个读取 desired manifest，实际会重复扫描 source/dynamic manifests，并对动态插件重复解析 `.wasm` artifact。
- 部分 WASM host service 使用包级默认 `kvcache.New()` 或 `config.New()`，`runtime.New` 内部也存在默认 `session.NewDBStore()` 后再被 setter 覆盖的 fallback。
- `pluginbridge` host service 相关代码存在大量手写 alias、codec、guest client、非 WASI stub 和 dispatcher 同步点。
- 根插件服务和 runtime 服务使用宽 facade 与 setter wiring，部分必需依赖缺失时通过 nil-safe no-op 静默退化。

## Goals / Non-Goals

**Goals:**

- 消除 runtime state 列表路径中的 N+1 manifest 扫描和 artifact 解析风险。
- 确保 WASM host service、runtime session/cache/config 等缓存敏感依赖都来自启动期共享实例或测试 fixture 显式注入。
- 为 host service 协议样板建立单一描述源、生成流程或等价覆盖校验，降低新增方法时多处遗漏风险。
- 将必需 runtime wiring 从静默 nil-safe no-op 收敛为构造或启动阶段可验证的显式依赖。
- 只清理明确重复注释、小文件和样板，不破坏现有职责边界。

**Non-Goals:**

- 不重写插件系统，不删除源码插件或动态插件任一运行模式。
- 不改变动态插件 ABI、WASM custom section、host service service/method 字符串、payload wire 字段编号或 guest SDK 公开行为。
- 不新增 HTTP API、前端页面、数据库表、DAO、插件清单字段或用户可见文案。
- 不把所有子组件合并回单个大文件；目录拆分仍以职责清晰为优先。
- 不在本变更中实现完整协议 IDL 生态，允许先以 Go 描述表和静态覆盖测试作为过渡。

## Decisions

### 决策 1：先修批量读取路径，再讨论更大范围重构

`ListRuntimeStates` 应在一次调用中构建 manifest map，而不是在 registry 循环中调用单项 `GetDesiredManifest`。实现可选两条路径：

- 在 runtime state 查询内先调用一次 `ScanManifests()`，按 `manifest.ID` 建 map。
- 在 catalog 增加显式 manifest snapshot 或 batch reader，并让 `GetDesiredManifest` 可在快照上下文中复用已扫描结果。

优先选择第二种或等价可复用方案，因为插件启动、列表、runtime projection 都可能受益；但第一种可作为低风险第一步。无论采用哪种实现，测试必须覆盖扫描次数或 artifact parse 次数不会随 registry 行数线性增长。

### 决策 2：WASM host service 默认实例改为显式配置

WASM host service 的 `cache`、`storage/config`、`lock`、`notify`、`config`、`manifest`、`org`、`tenant`、`AI` 等 dispatcher 运行时依赖必须由 `ConfigureWasmHostServices` 或测试 fixture 显式配置。生产包级变量不得默认调用 `kvcache.New()`、`config.New()` 或其他关键服务 `New()` 形成孤立服务图。

缺失配置时，dispatcher 应返回结构化 internal error；启动路径应在 `ConfigureWasmHostServices` 阶段 fail fast。测试可通过 fixture 显式注入 fake/shared service，不依赖包默认状态。

### 决策 3：协议样板先建立单一描述源和覆盖校验

当前 `pluginbridge` 的 host service 方法需要同步维护：

- service/method/capability 元数据。
- payload DTO 和 codec。
- `protocol` public type/function alias。
- guest contract。
- `wasip1` client。
- `!wasip1` unsupported stub。
- host dispatcher 和授权校验。

最终目标是生成这些同步点。第一阶段可以先建立 Go 描述表或 YAML 描述文件作为单一事实来源，并新增静态测试，验证每个描述的方法都被 alias、guest client、stub 和 dispatcher 覆盖。第二阶段再把 codec、alias、client/stub 迁移为生成代码，生成文件必须有 `Code generated` 标记并纳入 `go test`。

### 决策 4：wiring 校验优先于大规模 facade 拆分

根插件 `Service` 和 runtime `Service` 的 facade 行数很高，但其中很多方法是宿主、控制器、cron、中间件、OpenAPI 和插件生命周期的真实入口。直接拆掉 facade 会放大调用方迁移范围。

本次优先做两件事：

- 将必需依赖从 setter 后置覆盖改为构造函数参数或私有 composition root 校验项。
- 保留真正可选能力的 nil-safe 降级，但在接口注释和测试中明确可选原因。

如果实现阶段发现某个调用方只需要小接口，可新增窄接口注入调用方，但不得为了减少行数制造新的转发层。

### 决策 5：小规模行数清理只处理确定收益项

可以清理重复注释、合并同一组件下少于 50 行且只是 alias/转发的小文件，但不得合并职责不同的 runtime、catalog、wasm、integration 子组件。源码插件生命周期公开 API 可保持 typed 方法以保护插件作者体验；内部存储可改为 hook key/map，前提是测试证明公开方法和回调执行语义不变。

## Risks / Trade-offs

- 批量 manifest snapshot 过期 → snapshot 只在单次读取或启动编排上下文内使用，不跨请求长期缓存；动态 active release 仍以既有 release/generation 事实源为准。
- 移除默认服务实例导致部分测试失败 → 测试 fixture 必须显式配置 host service 依赖，失败可以暴露此前被默认实例掩盖的 wiring 问题。
- 协议生成引入工具复杂度 → 第一阶段允许先用 Go 描述表和覆盖测试，不强制一次性完成所有代码生成。
- facade 拆分过度影响调用方 → 本变更优先校验依赖完整性和删除静默 no-op，不以拆分 facade 作为验收目标。
- 公开协议行为回归 → 所有 protocol、guest、wasm host service 测试必须证明 wire 字段、service/method 字符串和错误语义保持不变。

## Migration Plan

1. 新增或修改增量规格，明确性能、DI、协议样板和缓存敏感实例来源要求。
2. 修复 runtime state manifest 批量读取路径，并补充扫描次数或 artifact parse 次数验证。
3. 移除 WASM host service 和 runtime 内部关键依赖默认 `New()` fallback，更新启动配置和测试 fixture。
4. 建立 host service 单一描述源和覆盖校验；必要时再迁移部分样板为生成代码。
5. 对 runtime wiring 做构造或启动校验，保留有明确理由的可选依赖降级。
6. 清理重复注释和确定可合并的小型 alias/转发文件。
7. 运行相关 Go 包测试、OpenSpec 严格校验和静态检索，记录 `i18n`、缓存一致性、数据权限、开发工具跨平台和测试策略影响。

## Open Questions

- 协议样板第一阶段采用 Go 描述表还是 YAML/JSON 描述文件，需要实现前根据现有工具链和跨平台要求确认。
- 是否同步引入生成命令到 `linactl` 或 `make`，取决于第一阶段是否已经需要产出生成文件。
