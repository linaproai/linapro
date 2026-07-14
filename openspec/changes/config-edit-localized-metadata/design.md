## Context

`sys_config` 内置参数的展示名与备注由运行时键 `config.<key>.name` / `config.<key>.remark` 投影；库内 seed 多为中文。列表 `List` 与 `GetByKey` 已本地化，而 `GetById` 返回原文，供编辑回填，避免用户保存时把英文投影写回库。

前端编辑弹窗直接使用 `configInfo` 回填 `name`/`remark`，且内置项未锁定这两字段，造成英文环境下的可见中文。

## Goals / Non-Goals

**Goals:**

- 英文（及任意请求语言）下编辑内置参数时，名称与描述与列表一致为本地化文案。
- 保存内置参数不得把投影后的 name/remark 写回 `sys_config`。
- 参数值编辑仍使用库内真实 `value`，自定义文案不被默认公共前端投影覆盖。

**Non-Goals:**

- 不为 name/remark 引入独立持久化多语言列或拆分 display DTO。
- 不改变自定义（非内置）参数的 name/remark 可编辑语义。
- 不改变导出行数据使用库内原文的策略。

## Decisions

### D1：详情仅本地化 name/remark，不本地化 value

- **决策**：`GetById` 对实体调用与列表相同的 name/remark 投影；**不**对详情 value 做公共前端默认值投影。
- **原因**：value 是可编辑真源；列表上的默认值展示投影若用于详情回填，保存会把默认英文写回库。

### D2：内置参数 Update 忽略 name/remark

- **决策**：`isBuiltInConfigRecord` 为真时，`Update` 不写入 `Name`/`Remark` 字段（即使请求体携带）。
- **原因**：内置元数据真源为 i18n 资源；写回任一语言投影都会污染存储并破坏其他语言列表的 fallback。

### D3：前端内置 name/remark 只读

- **决策**：`buildModalSchema` 在 `isBuiltin` 时禁用 name、remark 输入。
- **原因**：与后端「不可写回」一致，避免误导用户以为可改显示名。

### D4：自定义参数行为不变

- 非内置：`GetById` 仍可对有翻译键的 key 做投影（`Translate` 无键时 fallback 原文）；Update 仍可写 name/remark。
- 无翻译资源的自定义参数详情与编辑与现状一致。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 曾依赖「详情 name 即库原文」的调用方 | 文档与规范明确详情为展示投影；管理面是主消费者 |
| 内置参数管理员无法改备注 | 按产品决策：内置备注走 i18n；若未来需要可再开独立需求 |
| Update 内部仍用 `GetById` 读现有行 | 需保证 mutation 路径使用**未投影**实体做比较与写回；见实现注意 |

### 实现注意：mutation 路径与投影解耦

`Update`/`Delete` 当前通过 `GetById` 加载现有行。若 `GetById` 就地修改 `Name`/`Remark` 为投影值，内建校验与日志不受影响（忽略写 name/remark），但 **`previousValue` 等 value 字段不受影响**。为安全起见：

- 方案 A（推荐）：抽取 `getByIdRaw` 供 mutation；`GetById` 对外在 raw 之上再 localize name/remark。
- 方案 B：`GetById` 本地化后，Update 仍忽略 name/remark 写回（value 未本地化则足够）。

采用 **方案 A**，避免未来在详情投影 value 时引入回归。

## Migration Plan

- 无数据迁移。
- 已存在的被英文化污染的内置 name/remark 行：不在本变更范围；可用 seed/修复脚本另议。部署后新保存不再写入这两字段。

## Open Questions

- 无（产品方案已在反馈中确认继续落地）。
