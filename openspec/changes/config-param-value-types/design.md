## Context

参数设置（`sys_config` + 管理面 `system/config`）当前模型仅包含 `name` / `key` / `value` / `remark` 等字段，`value` 统一为 `TEXT` 字符串。前端新增/编辑统一使用 `Textarea`，枚举与开关类参数只能把合法取值写在 `remark` 中靠人工复制。

宿主运行时与插件仍以字符串 `value` 为业务真源（如 `true`/`false`、`panel-right`、`100`、Go duration 字符串）。类型化能力应服务**管理面编辑体验与写路径校验**，而不是改变运行时读取契约。

相关现状：

| 层次 | 现状 |
|------|------|
| 表结构 | `sys_config` 无类型/选项字段 |
| API | `ConfigItem` / Create / Update 无类型字段 |
| 前端 | `modalSchema` 固定 `Textarea` 编辑 `value` |
| 种子 | 布尔、布局、主题等合法值写在备注中 |
| 导入导出 | 列：`name`、`key`、`value`、`remark`、时间 |

## Goals / Non-Goals

**Goals:**

1. 为每个参数项声明稳定的 `valueType`，管理面按类型渲染输入组件。
2. 为需要枚举的类型持久化选项列表，编辑时下拉/单选/多选，而非拷贝备注。
3. 写路径（创建/更新/导入）按类型与选项校验 `value`。
4. 内置参数种子补齐类型与选项，开箱即用。
5. 历史无类型数据默认 `text`，运行时读路径与 revision 语义不变。

**Non-Goals:**

- 不将 `value` 拆成多类型存储列或 JSON 业务值。
- 不在本迭代引入“关联字典类型”作为选项来源（选项以内联 JSON 为主）。
- 不改造插件专属设置页、密钥掩码或 public frontend 白名单语义。
- 不新增独立“参数类型字典管理”后台模块。
- 不强制插件 `SetValue` 声明类型（缺省 `text` 即可）。

## Decisions

### D1：类型枚举固定为封闭集合

**决策**：`value_type` 使用封闭字符串枚举，后端与前端共用同一集合：

| valueType | 含义 | 管理面组件 | value 存储约定 |
|-----------|------|------------|----------------|
| `text` | 单行文本 | Input | 任意字符串（默认） |
| `textarea` | 多行文本 | Textarea | 任意字符串（可含换行） |
| `number` | 数字 | InputNumber | 十进制数字字符串，如 `100` |
| `boolean` | 布尔 | Switch 或 Radio | 仅 `true` / `false` |
| `select` | 下拉单选 | Select | 必须落在 options 的 `value` 集合 |
| `radio` | 单选组 | RadioGroup | 同 select |
| `multi_select` | 多选 | Select mode=multiple | 多个选项 value，使用英文分号 `;` 连接，顺序不敏感 |
| `richtext` | 富文本 | 富文本编辑器 | HTML/Markdown 字符串（实现选用项目已有富文本组件） |

空值、未知值或历史 NULL 在读投影时归一为 `text`。

**备选**：开放字符串自由扩展 — 拒绝，管理面组件映射无法保证；`duration` 独立类型 — 延后，现有 duration 参数可先用 `text` + 既有托管键校验。

### D2：选项持久化为 JSON，管理面使用简单行格式编辑

**决策**：`sys_config.options` 为 `TEXT`，**库内权威格式仍为 JSON 数组**：

```json
[
  {"label": "左侧", "value": "panel-left"},
  {"label": "居中", "value": "panel-center"},
  {"label": "右侧", "value": "panel-right"}
]
```

管理面与 Excel 导入/导出的**用户编辑格式**使用简单行文本（一行一项）：

```text
左侧=panel-left
居中=panel-center
panel-right
```

- 支持 `标签=值`、`标签|值`，或仅写 `值`（标签默认等于值）。
- 后端 `ParseOptions` 同时接受 JSON 数组与简单行格式；写入前统一 `EncodeOptions` 为 JSON。
- 仅 `select` / `radio` / `multi_select` 要求非空 options。
- `boolean` 固定合法值为 `true`/`false`，不依赖 options。
- 租户覆盖行：覆盖时默认继承平台行的 `value_type`/`options`（见 D5）。

**备选**：独立 options 表 — 过度设计；关联 `sys_dict` — 有复用价值但耦合字典治理，列为后续增强。

### D3：业务权威值仍是字符串 `value`

**决策**：所有类型最终序列化为现有 `value` 字段字符串。宿主 `GetRaw`、runtime snapshot、插件 `SetValue` 继续只关心字符串。类型元数据不进入 runtime snapshot 热路径（管理 CRUD 返回即可）。

### D4：校验分层

1. **通用类型校验**（所有参数）：`valueType` 合法；options 结构合法；`select`/`radio`/`multi_select` 的 value ⊆ options；`boolean` ∈ {true,false}；`number` 可解析为数字。
2. **托管键既有校验**（`validateManagedConfigValue`）：duration、上传上限等继续保留，与类型校验叠加。
3. **内置参数**：允许管理员改 `value`；是否允许改 `valueType`/`options`：
   - 内置（`is_builtin=1`）记录：**允许**改 value；**禁止**改 `valueType` 与 `options`（避免破坏宿主语义与种子约定）。
   - 非内置：可完整编辑类型与选项。

### D5：租户覆盖与类型继承

租户创建 fallback 覆盖时，从平台默认行复制 `value_type` 与 `options`，仅写入租户侧 `value`（及可编辑 name/remark 策略保持现状）。列表/详情投影继续返回有效行的类型元数据，供编辑组件使用。

### D6：前端动态表单

- 新增：表单含 `valueType` 选择器；切换类型时切换 value 输入组件，并在 `select`/`radio`/`multi_select` 时展示 options 编辑器（label/value 列表）。
- 编辑：加载详情后按 `valueType` 渲染；内置参数锁定类型与 options 字段为只读。
- `multi_select` 提交前将数组 join 为 `;`；回显时 split 过滤空串。
- 列表可不展示类型列（避免噪音），详情/表单必显；可选在表格增加类型列（低优先级，任务中列为可选）。

### D7：导入导出

在既有 Excel 列基础上增加：

- `valueType`（`config.field.valueType`）
- `options`（`config.field.options`，单元格存 JSON 字符串）

导入时缺省 `valueType` → `text`；缺省 options → 空；若类型需要 options 而单元格非法 JSON，则该行失败。

### D8：数据库迭代

不单独保留增量 SQL；在宿主既有 SQL 中直接落地类型元数据（无兼容迁移负担）：

1. `005-config-management.sql`：`sys_config` 建表即含 `value_type`（默认 `text`）与 `options`（默认空串）；内置宿主参数 Seed `INSERT` 直接写入匹配的 `value_type`/`options`。
2. `011-scheduled-job-management.sql`：定时任务相关内置参数 Seed 同步写入 `value_type`/`options`（如 `sys.cron.shell.enabled=boolean`、`sys.cron.log.retention=textarea`）。

种子只定义类型元数据与初始 `value`；`INSERT ... ON CONFLICT DO NOTHING` 保证重复执行不覆盖已有行。

### D9：API 字段命名

JSON 使用 camelCase：`valueType`、`options`（`options` 为对象数组；导入导出单元格为 JSON 文本）。

`options` 元素：

```ts
interface ConfigValueOption {
  label: string;
  value: string;
}
```

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 历史自定义参数无类型，体验与现在相同 | 默认 `text`，无破坏 |
| `multi_select` 与 IP 黑名单等“分号多值文本”语义混淆 | 黑名单继续用 `textarea`；`multi_select` 仅用于有限选项集合 |
| 富文本组件体积/依赖 | 优先复用项目已有富文本能力；若无则首版 `richtext` 降级为增强 Textarea，并在任务中注明 |
| 内置类型锁定后无法热修选项 | 通过 SQL 迭代与升级脚本修正种子；不开放 UI 改内置类型 |
| options label i18n 不完整 | 本迭代补齐主要内置枚举翻译键；未翻译时回退 label 原文 |
| 导入 options JSON 手误 | 行级失败原因本地化；模板示例行演示格式 |

## Migration Plan

1. 落地 DDL + 内置 key 类型 UPDATE。
2. `make dao` 刷新实体。
3. 后端 API/服务/校验/导入导出。
4. 前端模型与动态表单。
5. i18n（字段名、类型名、校验错误、内置 options 可选翻译）。
6. 单测 + E2E（布尔开关、下拉选择、非法枚举拒绝）。
7. `openspec validate` 与 `lina-review`。

回滚：DDL 列为可空默认兼容；若需回滚代码，旧前端忽略未知字段仍可编辑 value。

## Open Questions

1. **富文本**：若仓库尚无统一富文本编辑器，首版是否将 `richtext` 映射为增强多行文本？**建议**：首版映射为增强 Textarea，API 保留 `richtext` 枚举，后续再接编辑器。
2. **列表类型列**：是否展示？**建议**：本迭代不展示，降低表格噪音。
3. **options 的 i18n 键规范**：是否在本迭代强制 `config.<key>.option.<value>`？**建议**：强制为内置枚举补翻译；自定义参数允许纯文本 label。
