## Context

当前实现中：

- 后端 `List` 在 `IncludeBuiltin=false`（默认）时过滤掉 `distribution=builtin`。
- 前端列表不传 `includeBuiltin`，因此管理台看不到内建插件。
- 前端已有 `isBuiltinPlugin` 守卫，对诊断暴露的 builtin 行隐藏启用/安装/卸载/升级/租户策略开关。
- 详情与「管理」按钮本身不依赖 distribution，只需列表可见且已安装并存在管理页。

用户要求：builtin 出现在插件管理中，带「内置插件」标识（并保留自动启用类标识），不可卸载，可详情与管理。

## Goals / Non-Goals

**Goals:**

- 普通管理列表默认展示 builtin 与 managed。
- UI 用明确标识区分内建插件；写操作入口对 builtin 继续隐藏。
- 保留详情与管理入口可用性。
- 服务端生命周期写拒绝语义不变。

**Non-Goals:**

- 不开放 builtin 的安装、卸载、启用/禁用、运行时升级或租户供应策略写路径。
- 不改变 manifest `distribution` 校验与启动升级治理。
- 不引入独立「诊断列表」页面。

## Decisions

### D1：列表默认包含 builtin，去掉隐藏过滤

- **选择**：`List` 不再按 `IncludeBuiltin` 过滤；默认响应包含 builtin。
- **备选**：前端固定传 `includeBuiltin=true` —— 仍把「普通管理」与「诊断」语义耦合，易遗漏其它调用方。
- **兼容**：保留 `includeBuiltin` 查询字段，文档标注为兼容字段（忽略或始终视为包含），避免旧客户端因未知字段失败。

### D2：UI 标识分层

- **内置插件**：`distribution === 'builtin'` 时展示 badge（文案 i18n：中文「内置插件」/ 英文 “Built-in”）。
- **自动启用**：继续使用现有 `autoEnableManaged` → `autoEnableBadge`，与内置标识可并存。
- 写操作：继续 `v-if` 隐藏，不展示禁用态按钮。

### D3：管理按钮条件不变

- `canOpenPluginManagement` 仍为「已安装 + 存在管理页」。
- builtin 通常已安装且可有配置页，因此管理入口对用户可用。

### D4：测试策略

- 单元测试：默认列表包含 builtin 且带 `distribution`。
- E2E `TC016`：默认列表可见 builtin、有内置标识、无卸载等写操作、有详情与管理（管理在有管理页 mock 时；若 mock 无管理页则验证按钮按既有禁用语义，至少保证详情可用且无卸载）。

## Risks / Trade-offs

- **[Risk] 列表项增多，首屏更嘈杂** → Mitigation：靠 badge 区分；分页与筛选不变。
- **[Risk] 调用方曾假设默认无 builtin** → Mitigation：提案标记 BREAKING 列表语义；`distribution` 字段可过滤。
- **[Risk] 用户误以为可卸载内建插件** → Mitigation：隐藏卸载入口 + 服务端拒绝。

## Migration Plan

1. 更新增量规范与任务。
2. 改后端列表过滤与单测。
3. 改前端标识/i18n，复核写操作隐藏。
4. 更新 E2E TC016。
5. 本地 `openspec validate` + 相关测试。

回滚：恢复默认过滤与前端标识即可。

## Open Questions

无。
