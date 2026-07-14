## Context

`sys_config` 同时承载宿主内置参数与插件业务参数。管理面曾全量展示，导致双轨维护。

## Goals / Non-Goals

**Goals:**

- 数据驱动的 `system_manageable` 管理面分流。
- 能力 API 以四参数 `SetValue(ctx, key, value, options)` 控制管理面标记。
- 插件 settings 保存路径显式传 `SystemManageable: false`。

**Non-Goals:**

- 参数 category 分组。
- 按 `plugin.*` 命名空间硬过滤。

## Decisions

### 1. 字段

`system_manageable SMALLINT NOT NULL DEFAULT 1`。

### 2. 管理面边界

List/Export 仅 `=1`；Get/Update/Delete/Import 对 `=0` 拒绝或 not-found；Create 固定 `1`。

### 3. SetValue / BatchSetValue 能力扩展

```go
type SetSysConfigValueOptions struct {
    SystemManageable *bool // nil insert→0; nil update→keep; non-nil→write
}
type SetSysConfigValueItem struct { Key SysConfigKey; Value string }

SetValue(ctx, key, value, options *SetSysConfigValueOptions) error
BatchSetValue(ctx, items []SetSysConfigValueItem, options *SetSysConfigValueOptions) error
```

- `SetValue` 委托 `BatchSetValue`（单键批）。
- `BatchSetValue`：一事务写全部 items，成功后仅一次 revision；空 items 成功无副作用。
- 多字段插件 settings MUST 用 `BatchSetValue`，不得循环 `SetValue`。
- `options == nil` 等价于未指定 `SystemManageable`。
- 插件入口闭环 MUST 传 `SystemManageable: gconv.PtrBool(false)`。

### 4. SQL

仅改 `005` 建表；无独立增量；不考虑兼容迁移。

## Risks / Trade-offs

- 旧插件未传 options 时首次插入仍为 `0`（安全默认）。
- 开发库需 `db.init` 对齐表结构。

## Migration Plan

`make db.init` 重建；`make dao` 已刷新。

## Open Questions

无。
