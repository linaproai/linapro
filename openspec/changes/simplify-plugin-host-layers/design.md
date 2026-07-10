## Context

当前插件宿主大致分层为：

- `pkg/plugin/capability`：源码/动态共享领域契约
- `pkg/plugin/pluginbridge`：动态 guest SDK、protocol、hostservice catalog
- `internal/service/plugin`：catalog / lifecycle / runtime / upgrade / wasm / store 等宿主编排

动态 host service 调用链较长，且存在：

1. catalog 元数据与 `internal/hostservice` 手写常量双维护
2. dedicated binary codec 与 JSON envelope 并存
3. `HostServiceCapabilityJSON*` 历史别名
4. 根 facade 同时持有 `lifecycle.Service` 与 `upgrade.Service`，升级与安装/启停拆成平行编排入口

这些不是产品能力错误，而是实现层样板与职责边界可以收敛。

## Goals / Non-Goals

**Goals:**

- 新增 core-owned host service 方法默认且强制走 JSON envelope
- catalog 成为 service/method 常量权威源，并通过生成消除双维护
- WASM JSON 分发样板可复用
- 删除仅宿主内部使用的 historical JSON 别名
- upgrade 编排归 lifecycle 门面拥有，根包只依赖 lifecycle 暴露生命周期与升级 API
- 保持 lifecycle → runtime 单向依赖，不引入额外窄接口层（复用 `runtime.Service`）

**Non-Goals:**

- 不迁移既有 dedicated codec 方法到 JSON（避免协议破坏与大爆炸回归）
- 不生成完整 typed guest client / 业务 handler 实现
- 不合并源码插件与动态插件运行时模型
- 不收缩业务 `*cap` 方法面（留待后续使用面审计）
- 不把 `runtime.Service` 拆成多个公开窄接口（架构规则优先复用既有 Service）

## Decisions

### D1：dedicated codec 冻结为方法级名单

- catalog 测试改为方法级 allowlist（`service.method`），而不是服务级白名单。
- 名单内存量方法可继续使用 dedicated codec。
- 名单外任何 `PayloadKindDedicated` 视为失败。
- 新增方法必须 `PayloadKindJSON` 或 `PayloadKindNone`。

### D2：wire 常量单一来源（无 go generate）

- host service/method wire 常量只维护在 `protocol/hostservices/wire_constants.go`。
- catalog 的 `Service` / method wire 必须引用这些常量，禁止重复字符串字面量。
- 不使用 `go generate`。
- catalog 治理测试校验常量与 catalog 一致。
- `protocol` 继续 re-export 既有公开名称；`internal/hostservice` 通过 `hostservices.` 引用常量。

### D3：JSON helper 统一命名

- 生产路径统一使用 `Marshal/UnmarshalHostServiceJSONRequest/Response`。
- 删除 `HostServiceCapabilityJSON*` 类型别名与包装函数。
- WASM `decodeCapabilityJSONRequest` / `capabilityJSONResponse` 继续作为 dispatcher 本地 helper，内部改调 JSON API。

### D4：upgrade 归 lifecycle 拥有

- `upgrade` 包保留实现文件与内部模型，但只由 lifecycle 构造与持有。
- lifecycle 公开升级相关方法，或通过组合把 upgrade 能力挂到 lifecycle.Service。
- 根 `plugin.New` 不再直接 `upgrade.New`，也不再持有 `upgradeSvc` 字段。
- 根类型别名可继续从 lifecycle（或 lifecycle 再导出的 upgrade 类型）获得，避免 API 层大范围改名。

### D5：runtime 保持执行副作用角色

- lifecycle 继续依赖 `runtime.Service` 完成动态 reconcile、artifact 校验、动态 lifecycle callback 与动态卸载副作用。
- 不为 lifecycle 单独定义公开窄接口，避免与架构规则“优先复用已有 Service”冲突。
- 管理入口统一为 lifecycle；runtime 的 `Uninstall*` 仅作为 lifecycle/reconciler 调用的实现细节。

## Risks / Trade-offs

| 风险 | 缓解 |
| --- | --- |
| 生成常量与手写 re-export 不同步 | generate + 测试比对 catalog 与常量值 |
| 删除 CapabilityJSON 别名漏改调用点 | 全仓检索 + 编译/单测 |
| lifecycle 构造参数变多 | 保持显式 DI；upgrade 依赖在 lifecycle.New 内装配 |
| 误伤 dedicated 协议兼容 | 只冻结，不改存量编解码行为 |
| 升级路径行为回归 | 复用现有 upgrade 单测，迁移后仍在 lifecycle/upgrade 路径运行 |

## Migration Plan

1. 落地 catalog 冻结测试与文档约定。
2. 生成常量并切换 re-export，验证既有测试。
3. 删除 CapabilityJSON 别名并更新调用点。
4. 把 upgrade 装配下沉到 lifecycle，调整根 facade。
5. 跑相关包测试与 `openspec validate --strict`。

## Open Questions

- 无。dedicated 存量迁移与 `*cap` 方法面收缩不在本次范围，后续单独提案。
