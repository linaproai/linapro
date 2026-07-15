## Context

插件硬依赖目前由统一的 `dependency.Resolver` 评估，生命周期路径中：

- 安装/启用共用 `CheckInstall`（只认 `Installed`）
- 卸载/禁用共用 `CheckReverse`（只认已安装下游）

结果是「下游已全部禁用仍不能禁用 core」这类反直觉行为。规范在部分场景已写「已启用 consumer 阻断禁用 provider」，但实现未区分两轴。

## Goals / Non-Goals

**Goals:**

- 明确两轴阻断矩阵并落到 resolver + lifecycle
- 启用：依赖必须已安装、已启用、版本满足
- 禁用：仅已启用下游硬依赖阻断
- 安装：依赖必须已安装、版本满足（不要求启用）
- 卸载：已安装下游硬依赖阻断（不论启用态）
- 升级：继续按已安装下游版本契约反向保护
- 错误文案按轴分流（已安装 vs 已启用）
- 单元测试覆盖正向/反向矩阵关键场景

**Non-Goals:**

- 不引入自动级联启用/禁用/安装/卸载
- 不改变 soft dependency / auto-install 策略（仍不支持）
- 不重做租户级 per-tenant 启用图；依赖检查继续以插件 registry 全局 `Installed`/`Status` 为准（与 `UpdateStatus` 一致）
- 不在首屏列表展开完整依赖检查

## Decisions

### D1：两轴模型，而不是一套状态机

| 操作 | 正向 | 反向 |
|------|------|------|
| 安装 | 依赖 installed + version | — |
| 卸载 | — | 下游 installed |
| 启用 | 依赖 installed + **enabled** + version | — |
| 禁用 | — | 下游 **enabled** |
| 升级 | 同安装轴候选版本 | 下游 installed 版本契约 |

**理由：** 安装/卸载是结构契约，启用/禁用是运行时激活；混用导致运维路径被过度约束。

**备选：** 禁用也要求先卸载下游 — 否决，破坏可逆运维。

### D2：在 `PluginSnapshot` 增加 `Enabled`，不引入第二套 dependency 声明

- `ApplyRegistrySnapshot` 从 registry `Status == enabled` 写入 `Enabled`
- 未安装插件 `Enabled=false`
- 租户作用域：与现网 `UpdateStatus` 一致，只认全局 registry 启用态；不按 tenant 切片做反向禁用判断

**备选：** 按当前请求 tenant 聚合启用态 — 复杂度高，且禁用 API 当前是全局治理动作；本期不做。

### D3：Resolver 通过 input 模式区分，而不是复制两套图遍历

- `InstallCheckInput.RequireEnabled`：启用检查为 true，安装检查为 false
- `ReverseCheckInput.OnlyEnabledDependents`：禁用为 true，卸载/升级为 false
- 新增依赖边状态 `not_enabled` 与 blocker `dependency_not_enabled`（启用轴）
- 禁用反向可复用 `reverse_dependency` blocker code，但业务错误码/文案与卸载分流

### D4：错误码与文案分流

- 卸载反向：保留 `PLUGIN_REVERSE_DEPENDENCY_BLOCKED`（已安装插件依赖…）
- 禁用反向：新增 `PLUGIN_REVERSE_ENABLED_DEPENDENCY_BLOCKED`（已启用插件依赖…）
- 启用正向：在现有 `PLUGIN_DEPENDENCY_BLOCKED` 的 blockers 摘要中体现 `dependency_not_enabled`

### D5：只阻断，不级联

任何轴都不自动启停/装卸依赖链；调用方必须按正确顺序手动操作。

## Risks / Trade-offs

- **[行为 BREAKING]** 下游仅禁用后允许禁用 core → 文档与 UI 文案同步；卸载仍保护结构
- **[半启用残留]** core 与 consumer 可同时 installed+disabled → 靠启用正向检查防止错误启用顺序
- **[租户粒度]** 全局 Status 可能与 tenant 运行时投影不完全一致 → 本期与 UpdateStatus 对齐，后续若引入租户级启停 API 再扩展
- **[API 枚举扩展]** `DependencyStatus` / `BlockerCode` 增加值 → 前端需容忍未知码并以服务端文案展示

## Migration Plan

1. 先落 resolver + 单测，再改 lifecycle 启停接线
2. 同步 i18n 错误文案
3. 更新相关集成测试
4. 无需数据迁移；无需配置开关（产品语义直接切换）

回滚：恢复启停路径共用 install 轴检查即可。

## Open Questions

- 无（租户粒度已按 D2 默认拍板；若后续要 per-tenant 反向检查，另开变更）
