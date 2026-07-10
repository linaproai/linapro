## Why

`localdocs/plugin-owned-domain-capability-design.md`审查未发现阻断性缺口，但现有规范仍默认由`lina-core/pkg/plugin/capability`拥有全部插件可消费领域契约。随着`AI`等非核心领域能力继续扩展，这会让`lina-core/pkg/plugin`持续膨胀，并把非核心领域 DTO、provider SPI、动态协议和版本兼容压力集中到核心宿主。

## What Changes

- **BREAKING**：将非核心领域能力改为由领域 owner 插件维护公开契约、源码插件 SDK、动态 guest SDK、provider SPI 和版本策略；`lina-core`只保留插件内核、依赖治理、通用能力注册、动态调用路由、授权、审计、缓存一致性和生命周期治理。
- **BREAKING**：以`linapro-ai-core`为第一批试点，将`AI`能力契约 owner 从`apps/lina-core/pkg/plugin/capability/aicap`迁移到`apps/lina-plugins/linapro-ai-core/backend/cap/aicap`，并迁移 provider SPI、动态 guest SDK、DTO、错误码和能力方法状态契约。
- **BREAKING**：移除`pluginhost`中`ProvideAIText`等非核心领域硬编码 provider 声明，改为通用 capability descriptor 注册，并由 owner 插件提供类型安全 helper。
- **BREAKING**：移除`pluginbridge`/WASM 中按`AI`方法硬编码的 dynamic catalog、codec 和 dispatcher 分支，改为 owner-aware dynamic host service descriptor、授权快照和通用转发。
- 扩展`plugin.yaml hostServices`结构，使动态插件申请 owner 插件能力时显式声明`owner`、`version`、`service`、`methods`和资源范围；申请 owner 能力时必须在`dependencies.plugins`中硬依赖对应 owner 插件。
- 增加跨插件 Go import 边界治理：生产代码只能依赖 owner 插件的`backend/cap/...`公开契约，禁止依赖`backend/internal`、`dao`、`do`、`entity`或`backend/pkg`作为领域能力入口。
- 同步更新`apps/lina-core/pkg/plugin`和`linapro-ai-core`文档、项目规则、测试、治理扫描和动态 demo。

## Capabilities

### New Capabilities

- `plugin-owned-domain-capabilities`：定义插件拥有非核心领域能力的 owner 模型、`backend/cap`公开契约目录、能力描述符、源码插件和动态插件消费边界。

### Modified Capabilities

- `plugin-host-domain-capabilities`：将固定“core capability 拥有领域契约”的模型扩展为 core-owned 与 plugin-owned 两类领域能力，并约束跨插件调用、数据权限、性能和降级语义。
- `framework-capability-registry`：将非核心领域 provider 注册从硬编码强类型 facade 调整为通用 capability descriptor + owner helper。
- `plugin-host-service-extension`：扩展动态`hostServices`协议，支持 owner-aware service 声明、catalog 合并、授权快照和通用 dispatcher。
- `plugin-dependency-management`：要求动态插件申请 owner 能力时同步声明 owner 插件硬依赖，并在安装、启用、升级、卸载路径阻断依赖缺失、版本不满足和反向依赖破坏。
- `plugin-package-boundary-governance`：允许 owner 插件在`backend/cap`发布领域契约，同时保持`lina-core/pkg/plugin`只承载宿主内核和通用插件治理契约。
- `framework-ai-capability-namespace`：保留`AI`命名空间和类型化子能力语义，但将非核心`AI`契约 owner 迁移到`linapro-ai-core`。
- `framework-ai-text-capability`：保留`plugin.linapro-ai-core.ai.text.v1`文本生成、状态、错误和安全语义，但迁移契约包、provider SPI 和动态 SDK owner。
- `framework-ai-multimodal-capabilities`：保留多模态`AI`方法语义和大对象/operation/ref 安全边界，但迁移契约包和动态方法发布 owner。
- `linapro-ai-core-plugin`：扩展`linapro-ai-core`职责，使其除管理面和 provider 实现外，还拥有`AI`领域公开契约、SDK、SPI、能力描述符和版本策略。

## Impact

- 后端 Go：`apps/lina-core/pkg/plugin/capability/aicap`、`pluginhost`、`pluginbridge`、`internal/service/plugin/internal/wasm`、`internal/service/plugin/internal/catalog`、`internal/service/plugin/internal/store`、`internal/service/plugin/internal/dependency`、`apps/lina-plugins/linapro-ai-core/backend`以及所有 AI 消费方 import。
- 动态插件协议：`plugin.yaml hostServices`结构、WASM artifact host service sections、dynamic catalog、授权确认快照、upgrade diff、runtime dispatcher、guest SDK 和 demo 插件。
- 插件目录与依赖：`linapro-ai-core/backend/cap/aicap`新增公开契约；源码消费插件需要`go.mod require`和`plugin.yaml dependencies.plugins`一致。
- 数据权限和安全：owner 能力调用必须携带调用插件、actor、tenant、资源声明和审计 envelope；涉及宿主数据时必须在 owner 查询阶段注入数据权限。
- 缓存一致性：能力注册表、授权快照、owner 可用性和 SDK catalog 属于关键运行时数据，必须定义权威源、事务后失效、集群同步、最大陈旧窗口和故障重建。
- 前端与 UI：插件管理页需要展示 owner 能力授权来源、缺失依赖、版本不满足和反向阻断；涉及用户可观察行为需补 E2E。
- 文档与治理：更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`、`linapro-ai-core`双语 README、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`和相关 OpenSpec specs。
