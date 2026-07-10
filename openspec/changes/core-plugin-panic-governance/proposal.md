## Why

`apps/lina-core` 的生产 panic 治理测试通过硬编码官方插件路径与 panic 次数维护 allowlist，新增、重命名或调整官方源码插件 `init` 注册 fail-fast 时必须反向修改宿主测试，形成宿主与插件集合的静态耦合。同时，用户可见的租户治理错误文案仍绑定官方插件品牌名 `linapro-tenant-core`，与“主框架只认能力契约”的方向不一致。本变更优先落地低风险、高维护收益的治理解耦，并顺手完成文案能力化；不重做 tenant/org 启动一致性判定语义。

## What Changes

- 重构 `TestProductionPanicsMatchAllowlist` 相关策略：`apps/lina-core/**` 继续使用精确 Path/Function/Count 白名单；`apps/lina-plugins/**` 中符合约定的插件 `init` 注册 fail-fast `panic(err)` 按 AST 模式自动放行，不再枚举官方插件 ID。
- 非常规插件 panic（字面量 panic、非 `init` 生产 panic、非注册 fail-fast 模式）仍必须失败，或要求显式 allowlist 条目。
- 将 `CodePluginTenantProvisioningPolicyInvalid` 默认英文源文案及对应 `manifest/i18n/**/error.json` 从“linapro-tenant-core governance”改为能力语义描述（multi-tenant / framework tenant governance），同步测试断言。
- **不在本变更范围内**：`validateProviderStartupConsistency` 的能力可用性重写、`ProviderPluginID` 删除、插件工作区路径收口、测试夹具全面虚构化、文档示例清洗（见 `localdocs/core-plugin-id-decoupling-design.md` 后续变更）。

## Capabilities

### New Capabilities

- （无）本变更不引入新的业务能力域。

### Modified Capabilities

- `backend-conformance`: 明确生产 panic 治理对官方插件工作区采用模式匹配自动放行，宿主测试不得维护完整官方插件 ID 清单。
- `service-dependency-injection-governance`: 调整“插件 `init` 注册 fail-fast panic 必须写入 allowlist”的表述，允许由治理扫描按类别/AST 模式自动归类，而非逐插件枚举。
- `runtime-message-i18n-governance`: 要求宿主运行时业务错误在描述能力约束时使用能力语义，不得将官方插件品牌 ID 写入用户可见错误文案。

## Impact

- **代码**：`apps/lina-core/internal/cmd/cmd_test.go` 及 panic 扫描 helper；`internal/service/plugin/plugin_code.go`；`manifest/i18n/**/error.json`；依赖该错误文案的测试。
- **API / DTO 结构**：无变更；错误码 `PLUGIN_TENANT_PROVISIONING_POLICY_INVALID` 保持稳定，仅默认文案与翻译变更。
- **数据库 / 前端 / 插件 ABI**：无影响。
- **启动行为**：无变更（不修改启动一致性判定）。
- **i18n**：有影响，需同步错误语言包。
- **缓存 / 数据权限 / 跨平台工具**：无影响。
- **设计输入**：`localdocs/core-plugin-id-decoupling-design.md` 中的 P0-A 与 P0-C。
