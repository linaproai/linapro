## Context

探索阶段确认主框架里仍存在三类不必要复杂度：

- `RuntimeDelegate` 和若干内部 provider 为了解决插件根服务构造顺序而存在，但部分运行期方法在未绑定时返回 nil 或原始入参，导致启动接线缺陷可能被隐藏到请求期。
- `upgradeCachePublisher`、`upgradeCacheFreshener` 等小 adapter 只包装根 `serviceImpl` 的缓存方法，却在 service 为 nil 时静默成功。
- `httpstartup` 通过 `kvcache.SetDefaultProvider` 设置进程级默认 provider，再由 `kvcache.New()` 读取该默认值创建共享服务。该路径让生产拓扑选择隐含在全局可变状态里，不利于审查共享实例来源。

这些问题不要求大规模重构。`RuntimeDelegate` 仍然是当前插件根服务组合中有价值的启动接缝；本变更只收紧错误语义和后端选择来源，避免把局部循环拆解成更多抽象层。

## Goals / Non-Goals

**Goals：**

- 保留必要 delegate，但让绑定状态显式可检查。
- 对能够返回 `error` 的运行期 delegate 调用采用 fail-fast，不再把未绑定当作成功。
- 对插件内部 cache/upgrade 适配器的 nil service 路径返回明确错误或在构造期避免 nil。
- 让 HTTP 启动期直接按 `cluster.enabled` 选择 `kvcache` provider 并创建共享服务。
- 用单元测试和启动包编译门禁证明行为收紧不会破坏生产装配。

**Non-Goals：**

- 不重写插件服务内部拓扑，不移除所有 delegate。
- 不改变插件 host service 协议、`plugin.yaml`、HTTP API、数据库或前端。
- 不把 capability provider manager、租户/组织/AI 能力目录纳入本轮简化。
- 不修改 `apps/lina-plugins/<plugin-id>/` 下任何业务插件资源。

## Decisions

### 1. Delegate 保留为组合接缝，但绑定状态必须可诊断

`RuntimeDelegate` 的职责限定为启动期循环依赖接缝。它应提供 `Bound()` 或等价可测试状态，并在 `BindService(nil)` 等无效绑定时保持未绑定且可被测试识别。生产启动路径必须在完成插件根服务构造后显式绑定，并处理绑定失败或通过测试证明绑定完成。

能够返回 `error` 的运行期回调，例如登录成功/失败、登出、cache notifier、dependency validator、upgrade cache publisher/freshener，在未绑定时返回内部诊断错误。只读降级方法如果当前接口没有错误返回值，可以继续 fail-closed 或返回输入投影，但必须避免伪装写入、发布、刷新或校验已经成功。

### 2. 内部 adapter 不再把 nil service 当作成功

`upgradeCachePublisher` 和 `upgradeCacheFreshener` 这类 adapter 仍可保留在根包内，原因是它们承接内部子组件的窄接口，避免让子组件持有插件根门面。但 adapter 方法不得在 service 为 nil 时返回 nil。测试可以直接覆盖 nil adapter 的失败语义，生产路径则由 `plugin.New()` 传入根 `serviceImpl`。

### 3. `kvcache` 后端选择在 HTTP 启动期显式完成

`httpstartup` 根据当前 cluster/coordination 初始化结果创建 `kvCacheSvc`：

| 模式 | Provider | 创建位置 |
|------|----------|----------|
| `cluster.enabled=false` | `kvcache.NewSQLTableProvider()` | HTTP runtime 启动装配 |
| `cluster.enabled=true` | `kvcache.NewCoordinationKVProvider(coordinationSvc)` | HTTP runtime 启动装配 |

`kvcache.New(kvcache.WithProvider(provider))` 或等价显式入口作为生产路径。`SetDefaultProvider` 可以暂时保留给测试兼容或后续清理，但 HTTP 启动装配不得依赖它完成拓扑选择。

### 4. 缓存一致性七要素

| 要素 | 决策 |
|------|------|
| 权威数据源 | `kvcache` 仍是有损缓存，不作为权限、配置、插件稳定状态或 revision 的权威源 |
| 一致性模型 | 单机使用 SQL table 后端；集群使用 coordination KV 后端 |
| 失效触发点 | 缓存条目按既有 TTL、delete、expire、cleanup 或后端原生过期处理 |
| 跨实例同步 | 集群模式复用统一 coordination provider；不创建独立 Redis client |
| 最大可接受陈旧 | 由插件缓存 TTL 和调用方有损缓存语义决定，不承担关键状态 freshness |
| 故障降级 | 后端失败返回错误，不向插件或调用方伪装写入成功 |
| 恢复路径 | SQL table cleanup、coordination 后端原生 TTL、缓存未命中回源由调用方决定 |

## Impact Analysis

- `i18n`：无运行时用户可见文案、API 文档源文本、插件清单或语言包资源变更。
- 数据权限：不新增或修改列表、详情、导出、下载、聚合、写入或授权关系接口；无数据权限边界变化。
- API 契约：不新增或修改 HTTP 路由、DTO、`g.Meta`、权限标签或响应结构。
- 数据库：不新增或修改 SQL、DAO、DO、Entity、索引或软删除语义。
- 开发工具跨平台：不修改 Makefile、脚本、CI、`linactl` 或工具型 Go 代码。
- 测试策略：新增或更新 Go 单元测试覆盖未绑定 delegate、nil adapter 和显式 provider 选择；运行插件服务、`kvcache` 和 HTTP 启动包测试。
- `apps/lina-core/pkg/plugin` README：本变更只收紧宿主内部组合和共享缓存服务创建，不改变公开插件 SDK 或 host service 契约，预计无需同步 README。

## Risks / Trade-offs

- 某些测试可能依赖未绑定 delegate 的旧静默成功语义。处理方式是更新测试 fixture 明确绑定，或在确认为纯只读零值语义时保留 fail-closed。
- 直接删除 `SetDefaultProvider` 可能造成大范围测试 churn。本轮优先从生产启动路径移除依赖，后续如确认无测试价值再单独清理。
- 过度 fail-fast 可能影响启动前短暂调用路径。因此只对可返回 `error` 且代表写入、发布、刷新、校验成功的操作收紧，避免改变没有错误返回值的只读投影签名。
