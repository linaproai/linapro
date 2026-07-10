## Context

专题设计见 `localdocs/core-plugin-id-decoupling-design.md`。当前仓库中：

1. `apps/lina-core/internal/cmd/cmd_test.go` 的 `productionPanicPolicy.Allowances` 为约 10 个官方源码插件路径维护 Path + Function + Count 白名单；`ScanRoots` 已覆盖 `apps/lina-plugins`，但放行仍依赖插件 ID 枚举。
2. `service-dependency-injection-governance` 与 `backend-conformance` 要求插件 `init` 注册 fail-fast panic 被 allowlist 记录；实现上被解读为“主框架测试维护完整插件清单”。
3. `CodePluginTenantProvisioningPolicyInvalid` 默认文案与 `manifest/i18n/en-US/error.json` 含 `linapro-tenant-core` 品牌名；错误码本身稳定。

本变更只落地 P0-A（panic 治理）与 P0-C（文案能力化）。P0-B 启动一致性语义、路径收口、测试夹具虚构化不在范围。

## Goals / Non-Goals

**Goals:**

- 官方源码插件按既有模式在 `backend/plugin.go` 的 `init` 中 `panic(err)` 注册失败时，**无需**修改 `lina-core` 测试白名单。
- 非常规插件 panic 仍被治理测试拒绝。
- `lina-core` 自身 panic 继续精确白名单与计数校验。
- 用户可见租户配置策略错误文案改为能力语义，不再出现官方插件品牌 ID。
- 变更可独立验证、回滚，不影响启动与插件生命周期行为。

**Non-Goals:**

- 不修改 `validateProviderStartupConsistency` 或删除 `tenantcap.ProviderPluginID` / `orgcap.ProviderPluginID`。
- 不引入 `pluginworkspace` 路径收口。
- 不批量替换测试 fixture 中的 `linapro-*` 插件 ID。
- 不取消 monorepo `apps/lina-plugins` 工作区与 `official_plugins` 构建装配。
- 不重做 tenant/org 能力归属。

## Decisions

### D1：插件 panic 采用 AST 模式自动放行（方案 A）

| 选项 | 说明 | 结论 |
| --- | --- | --- |
| A. AST 模式自动 allow | 扫描 `apps/lina-plugins/*/backend/plugin.go` 的 `init`，识别注册 fail-fast `panic(err)` | **采用**：主框架零清单 |
| B. 插件侧自声明 yaml/注释 | 每插件维护 allow 元数据，宿主汇总 | 多约定，暂不引入 |
| C. 审计迁出 lina-core | 放到 linactl 或插件仓 | 边界清晰但拆分测试入口成本高 |

**规则：**

1. `apps/lina-core/**`：继续精确 `Path + Function + Count + Category + Reason`。
2. `apps/lina-plugins/**`：
   - 位于 `backend/plugin.go` 的 `init`；
   - 每个 panic 均为“对 error-returning 注册 API 的错误结果 fail-fast”（`panic(err)` 或等价 `panic(ident)`，且 `ident` 来自该错误路径）；
   - 自动归入 `plugin-registration` 类别并通过。
3. 字面量 `panic("...")`、`panic(fmt.Sprintf(...))`、非 `init`、非 `backend/plugin.go` 的生产 panic：失败；若确有需要，再进入显式 allowlist（本变更默认不为插件新增显式条目）。
4. 官方工作区未初始化时继续 `Skip`，复用 `OfficialPluginsWorkspaceReady`。

**实现落点：** 优先在 `cmd` 包测试内抽取 helper（如 `isPluginRegistrationFailFast`），避免新增生产依赖；若 helper 过长可放同包 `_test.go` 或 `cmd` 测试辅助文件。

### D2：宿主精确白名单与插件模式路径分离

扫描仍可共用 `scanProductionPanicCalls`，但匹配逻辑改为：

```text
found panic
  → 若在 lina-core 精确 allowlist：按 Count 校验
  → 否则若在 lina-plugins 且模式自动放行：通过
  → 否则失败
```

过期白名单检测：仅对 `lina-core` 精确条目做 stale 检查；插件条目删除后不再存在“stale plugin allowance”。

### D3：错误文案能力化，错误码不变

- 源文案（示意）：
  - 旧：`must support linapro-tenant-core governance ...`
  - 新：`must support multi-tenant governance ...`  
    或 `must support framework tenant governance ...`
- 同步 `apps/lina-core/manifest/i18n/<locale>/error.json` 对应键。
- 测试断言改为能力语义子串；错误码 `PLUGIN_TENANT_PROVISIONING_POLICY_INVALID` 不变。
- 不修改启动一致性错误串中的 `linapro-tenant-core`（属 P0-B）。

### D4：范围冻结

任何“既然在改，顺便把 ProviderPluginID 去掉”的冲动均拒绝；防止启动检查变松。P0-B 另开变更。

## Risks / Trade-offs

| 风险 | 缓解 |
| --- | --- |
| AST 过宽误放行危险 panic | 仅限 `backend/plugin.go` + `init` + error fail-fast 模式；字面量/非常规失败；补 AST 单测或反向 fixture 说明 |
| 某插件使用非标准 fail-fast 写法导致测试失败 | 先判断是否应改写为标准 `panic(err)`；例外才加显式 allow（默认避免） |
| 文案变更导致依赖英文字符串的外部断言失败 | 仓库内测试同步；对外依赖错误码而非自然语言 |
| 与 service-dependency-injection 规范“必须记录 allowlist”字面冲突 | 本变更同步修订规范：允许模式归类等价于记录 |

## Migration Plan

1. 落地 panic 扫描逻辑与删除插件枚举条目。
2. 跑 `cmd` 包相关测试，确认官方工作区就绪时全绿。
3. 改错误源文案与 i18n，同步断言。
4. 静态检索：
   - `rg -n 'apps/lina-plugins/linapro-' apps/lina-core/internal/cmd/cmd_test.go` 应为 0（allowlist 部分）。
   - 生产错误定义中 `PLUGIN_TENANT_PROVISIONING` 相关文案不再含 `linapro-tenant-core`。
5. 回滚：还原 `cmd_test.go` 与错误/i18n 文件即可，无数据迁移。

## Open Questions

- 无阻塞项。推荐英文源文案采用 `multi-tenant governance`（简短清晰）；若审查偏好更精确的 `framework.tenant.v1` 表述，实现时可二选一并在 tasks 中固定一种。
