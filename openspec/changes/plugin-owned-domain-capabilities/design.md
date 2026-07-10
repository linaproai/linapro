## Context

`localdocs/plugin-owned-domain-capability-design.md`审查结论是方案可落地，无阻断性缺口。代码库核对显示当前`AI`能力确实处于“实现归插件、契约归 core”的混合状态：

- `apps/lina-core/pkg/plugin/capability/aicap`拥有`AI`公开契约和多个子能力 DTO。
- `apps/lina-core/pkg/plugin/pluginhost`拥有`ProvideAIText`、`GetAITextProviderFactory`和`aiTextProvider`字段。
- `apps/lina-core/pkg/plugin/pluginbridge/protocol/hostservices/catalog.go`在 core 中硬编码`ai`动态方法。
- `apps/lina-core/internal/service/plugin/internal/wasm/wasm_host_service_ai.go`通过逐方法`switch`分发`AI`host service。
- `apps/lina-plugins/linapro-ai-core`已经拥有 provider、模型、档位、调用日志、管理 API、前端和`i18n`资源，但尚未拥有`backend/cap`公开契约。

这次变更是跨 core 插件内核、动态插件协议、源码插件 provider 注册、官方`AI`插件、文档和治理扫描的架构迁移，不进入兼容保留模式。项目无历史负担，允许一次性破坏性迁移，但必须在同一变更内修复所有调用方和测试。

## Goals / Non-Goals

**Goals:**

- 建立 plugin-owned 非核心领域能力模型，避免`lina-core/pkg/plugin`随业务领域增长持续膨胀。
- 将`AI`作为第一批试点，把契约、DTO、错误、SDK、SPI、provider helper 和版本策略迁移到`linapro-ai-core/backend/cap/aicap`。
- 用通用 capability descriptor 取代`ProvideAIText`、AI 专属动态 catalog、AI 专属 codec 和 WASM AI dispatcher。
- 扩展动态`hostServices`为 owner-aware 声明，确保 owner 插件依赖、方法授权、资源范围和运行时调用一致。
- 增加跨插件 import 边界扫描，只允许生产代码依赖 owner 插件`backend/cap/...`。
- 同步更新 OpenSpec specs、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`和双语 README。

**Non-Goals:**

- 不迁移宿主内核能力，例如`runtime`、`storage`、`cache`、`lock`、`manifest`、`plugins`、`bizctx`和`route`。
- 不把`Users`、`Auth`、`Files`、`Dict`等基础宿主通用业务能力迁出 core。
- 不引入`dependencies.capabilities`、软依赖或自动安装策略。
- 不允许插件间 import owner 插件`backend/internal`、DAO、DO、Entity 或 provider 实现。
- 不在本变更中评估并迁移`Org`、`Tenant`等其它领域，只保留后续评估入口。

## Decisions

### 1. 使用`backend/cap`作为 owner 插件公开契约目录

采用`apps/lina-plugins/<plugin-id>/backend/cap/<domain>cap`，而不是复用`backend/pkg`。`cap`语义更窄，治理扫描可以精准识别跨插件领域契约；`pkg`在 Go 项目中容易被普通 helper、client 和常量污染。

替代方案：继续由 core `pkg/plugin/capability/<domain>cap`拥有契约。该方案实现成本低，但会继续扩大 core 边界，与技术方案目标冲突。

### 2. owner-aware `hostServices`使用显式`owner`和`version`字段

采用：

```yaml
hostServices:
  - service: ai
    owner: linapro-ai-core
    version: v1
    methods:
      - text.generate
```

而不是`service: plugin:linapro-ai-core:ai:v1`。结构化字段更利于管理端展示、授权 diff、依赖交叉校验和 catalog 合并，也避免把 service 字符串变成二级编码协议。

### 3. core 保留通用 descriptor 和路由，不保留 AI 方法 owner

core 新增通用 capability descriptor、注册表、owner-aware catalog 合并、授权快照和 dispatcher。`AI`方法、DTO、codec、guest helper 和 provider SPI 迁移到`linapro-ai-core`。core 不再维护`AI`专属`switch`和专属 codec，只负责：

- 校验调用插件依赖 owner 插件。
- 校验 owner 插件安装、启用和版本。
- 校验 method、resource、payload 和授权快照。
- 转发到 owner handler。
- 记录审计 envelope 和错误映射。

### 4. `AI`迁移覆盖全量已发布契约，动态方法按可运行路径发布

不能只迁移`aitext`契约。当前 core catalog 已发布文本、图片、Embedding、音频、视觉、文档、安全、视频和 operation 方法的公开语义。实施时必须把这些方法的 DTO、方法常量和错误语义迁到`linapro-ai-core/backend/cap/aicap`，避免 core/owner 双契约。动态`ai.v1`descriptor 仅发布当前具备真实 invoker 路径的方法；尚未接线的多模态方法保留在 owner 契约中，待 provider SPI 落地后再进入授权 catalog。

### 5. owner 能力依赖复用`dependencies.plugins`

源码插件 import owner `backend/cap/...`时，必须在`plugin.yaml dependencies.plugins`声明 owner 插件依赖；动态插件声明`hostServices.owner`时也必须声明 owner 插件依赖。不新增 capability 依赖模型，避免依赖图、安装阻断和反向卸载保护出现第二套规则。

### 6. 能力注册表和授权快照按关键运行时缓存治理

能力注册表、授权快照、owner 可用性和 SDK catalog 的权威源分别是已发现插件、源码注册、动态 artifact 描述符、插件 manifest 和 owner descriptor。插件安装、启用、禁用、升级、卸载事务提交后才能发布失效。`cluster.enabled=true`时必须复用宿主运行时修订、共享修订号、事件广播或等价协调机制；不得只刷新本地内存。

## Migration Plan

1. 规范冻结：更新本变更 specs、项目规则和 README 范围，明确 core-owned/plugin-owned 分类、`backend/cap`目录、owner-aware host service schema、依赖治理和缓存一致性。
2. core 通用注册：实现 capability descriptor、注册表、owner-aware catalog 合并、host service schema 校验、授权快照、升级 diff 和通用 dispatcher。
3. AI owner 迁移：在`linapro-ai-core/backend/cap/aicap`建立契约、`spi`和`bridge`，迁移 core `aicap`、AI provider 注册、动态 SDK、codec 和 dispatcher。
4. 调用方迁移：更新`linapro-ai-core`自身、动态 demo、测试替身和所有 AI 消费方 import；删除或替换 core `aicap`生产入口，不保留长期 re-export。
5. 治理扫描：增加跨插件 import 边界扫描、owner dependency 与 host service 交叉校验、AI core 残留 import 检索。
6. 验证与审查：运行 OpenSpec strict、Go 包测试、启动绑定测试、动态 host call 测试、i18n 检查、治理扫描和`lina-review`。

Rollback 策略：本项目无历史负担，不做发布兼容回滚。若实施中发现迁移面超出单变更承载能力，应回退本变更未完成代码并拆分 OpenSpec，而不是保留 core/owner 双契约长期并存。

## Risks / Trade-offs

- [跨插件 Go module 依赖复杂] -> 通过插件 module path 映射、构建期本地`replace`和 import/dependency 扫描约束。
- [owner-aware host service schema 牵涉面广] -> 在同一变更内覆盖 manifest 解析、artifact section、授权快照、upgrade diff、runtime dispatcher 和 demo。
- [core 过度通用化] -> core descriptor 只表达 owner、version、method、风险、资源和 handler，不承载领域业务 DTO 或弱类型业务网关。
- [动态 SDK 淡化安全边界] -> SDK 只编码和声明 helper，运行时必须经过 host call 授权、依赖检查和审计。
- [缓存陈旧导致越权调用] -> owner 启停和授权变化事务提交后失效，集群模式接入共享修订或事件，缓存不可用时回源重建或拒绝。
- [文档和规则不同步] -> tasks 中强制更新 core/plugin README、AI 插件 README、`.agents/rules/architecture.md`和`.agents/rules/plugin.md`。

## Open Questions

- 无阻断性开放问题。实现时若发现某个已发布 AI 方法无法在同一变更中迁移，必须先更新本变更 specs 和 tasks，明确排除原因、剩余风险和后续变更，而不是默认保留双 owner。
