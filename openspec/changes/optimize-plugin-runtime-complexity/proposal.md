## Why

当前 `apps/lina-core/internal/service/plugin` 与 `apps/lina-core/pkg/plugin` 已经承担插件宿主、动态插件、WASM bridge、host service、生命周期和缓存治理等核心职责，但审查发现部分行数增长来自可压缩的手写协议样板、隐式默认依赖和重复运行时扫描。继续放任这些问题会增加后续宿主能力扩展成本，并在插件数量增长或集群部署下放大性能与一致性风险。

本次变更跟进插件运行时复杂度治理，目标是在不削弱 `lina-core` 框架宿主通用性的前提下，消除实际风险、收敛维护样板，并建立后续扩展的可验证约束。

## What Changes

- 修复插件 runtime state 列表路径中的逐插件 manifest 重复扫描问题，避免 `ListRuntimeStates` 随插件数触发重复 `ScanManifests` 和动态 `.wasm` artifact 解析。
- 移除 WASM host service 与插件 runtime 中缓存、配置、session 等敏感依赖的隐式默认服务图，要求启动期或测试 fixture 显式注入共享实例或共享后端。
- 将 `pluginbridge` host service 协议、payload codec、public protocol alias、guest WASI client、非 WASI stub 和 host dispatcher 覆盖关系纳入单一元数据或生成式治理，减少手写线性样板。
- 收敛插件 runtime 子组件 wiring，区分必需依赖和可选能力；必需依赖缺失时在构造或启动校验阶段显式失败，不再通过 nil-safe no-op 静默跳过。
- 清理重复注释和可合并的小文件，但不以牺牲职责边界为目标做大规模目录扁平化。
- 不改变动态插件公开 HTTP 路由、host service service/method 字符串、payload wire 字段编号、插件生命周期语义或前端可见交互。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-runtime-loading`：补充插件运行时列表、manifest discovery 和 active release 资源读取的批量/快照性能要求，并约束运行时 wiring 必须显式校验。
- `plugin-host-service-extension`：补充 WASM host service 依赖必须由宿主启动期统一注入，禁止生产路径使用包级默认服务实例。
- `pluginbridge-subcomponent-architecture`：补充 host service 协议样板必须由单一描述源或生成流程覆盖，并要求自动化验证所有公开 alias、guest client、stub 和 dispatcher 同步。
- `distributed-cache-coordination`：补充缓存敏感服务实例来源审查要求，明确插件 WASM host service 和 runtime session/cache/config 依赖不得创建孤立默认实例。

## Impact

- 影响 `apps/lina-core/internal/service/plugin/internal/runtime` 的 runtime state 查询、manifest 读取、依赖 wiring 和测试。
- 影响 `apps/lina-core/internal/service/plugin/internal/catalog` 的 manifest 批量读取或快照能力。
- 影响 `apps/lina-core/internal/service/plugin/internal/wasm` 的 cache、storage、config、session 相关 host service 配置入口和测试 fixture。
- 影响 `apps/lina-core/pkg/plugin/pluginbridge/internal/hostservice`、`pkg/plugin/pluginbridge/protocol`、`pkg/plugin/capability/guest` 的 host service 协议维护方式、生成或覆盖校验。
- 影响 `apps/lina-core/pkg/plugin/pluginhost` 的生命周期 handler 内部存储方式时，只允许在保持公开契约清晰的前提下进行小范围压缩。
- 不新增 HTTP API、数据库表、DAO、前端页面、插件清单字段或用户可见文案。
- `i18n` 影响：无运行时文案或翻译资源变更；若实现阶段修改用户可见错误 fallback 或 API 文档源文本，必须同步按 `i18n` 规则处理。
- 缓存一致性影响：涉及插件运行时、WASM host service、cache/config/session 依赖来源，需要验证共享实例和集群一致性边界。
- 数据权限影响：本变更不新增业务数据读写或租户/组织数据可见性路径；host service 行为保持既有授权语义。
- 开发工具跨平台影响：若新增协议生成工具或治理脚本，必须按开发工具规则补充跨平台入口和验证。
