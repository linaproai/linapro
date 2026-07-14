## ADDED Requirements

### Requirement: 参数项必须声明封闭的值类型集合

系统 SHALL 为每条 `sys_config` 记录持久化 `value_type` 字段，取值必须属于封闭集合：`text`、`textarea`、`number`、`boolean`、`select`、`radio`、`multi_select`、`richtext`。缺失、空串或未知类型在管理面投影时 MUST 归一为 `text`。业务权威值仍存储于字符串字段 `value`，运行时读取路径 MUST 继续只依赖该字符串值。

#### Scenario: 新建参数默认类型为 text
- **WHEN** 管理员创建参数且未指定 `valueType`
- **THEN** 系统将 `value_type` 存为 `text`
- **AND** 后续详情返回 `valueType` 为 `text`

#### Scenario: 非法 valueType 被拒绝
- **WHEN** 调用方创建或更新参数且 `valueType` 不在封闭集合中
- **THEN** 系统拒绝写入并返回结构化校验错误
- **AND** 数据库记录不被修改

#### Scenario: 运行时读取不依赖 valueType
- **WHEN** 宿主通过 `GetRaw` 或 runtime snapshot 读取某配置键
- **THEN** 返回值仅基于有效字符串 `value`
- **AND** 不因 `value_type` 缺失或变更而改变读取优先级

### Requirement: 枚举类类型必须持久化选项列表

系统 SHALL 在 `sys_config.options` 中以 JSON 数组持久化选项，元素包含 `label` 与 `value` 字符串字段。当 `value_type` 为 `select`、`radio` 或 `multi_select` 时，options MUST 为至少包含一个元素的合法 JSON 数组；其他类型可将 options 存为空串或空数组。管理面与导入解析 MAY 接受简单行格式（`标签=值`、`标签|值` 或仅值），写入前 MUST 统一编码为 JSON 数组。

#### Scenario: 下拉参数保存选项列表
- **WHEN** 管理员创建 `valueType=select` 的参数，并提供选项 `[{label:"左侧",value:"panel-left"},{label:"右侧",value:"panel-right"}]`
- **THEN** 系统持久化该 options JSON
- **AND** 详情接口返回可解析的 options 数组

#### Scenario: 枚举类型缺少选项被拒绝
- **WHEN** 调用方创建或更新 `valueType` 为 `select`/`radio`/`multi_select` 且 options 为空或非法 JSON
- **THEN** 系统拒绝写入并返回结构化校验错误

### Requirement: 写路径必须按类型校验参数值

系统在创建、更新与导入参数时 SHALL 按 `valueType` 与 options 校验 `value`：

- `boolean`：仅允许 `true` 或 `false`
- `number`：必须可解析为十进制数字
- `select` / `radio`：必须等于某一 option 的 `value`
- `multi_select`：按英文分号 `;` 拆分后每一段都必须属于 options 的 `value` 集合（允许空串表示未选）
- `text` / `textarea` / `richtext`：不额外限制字符形态（仍受托管键既有校验约束）

托管运行时参数的既有格式校验（如 duration、上传上限）MUST 与类型校验叠加生效。

#### Scenario: 布尔参数拒绝非 true/false
- **WHEN** 管理员将 `valueType=boolean` 的参数值更新为 `yes`
- **THEN** 系统拒绝更新并返回校验错误
- **AND** 原值保持不变

#### Scenario: 下拉参数仅允许选项内的值
- **WHEN** 参数 `valueType=select` 且 options 仅包含 `panel-left` 与 `panel-right`
- **AND** 管理员提交 value=`panel-center`
- **THEN** 系统拒绝更新

#### Scenario: 多选参数以分号序列化
- **WHEN** 管理员保存 `valueType=multi_select` 且选中 option 值 `a` 与 `b`
- **THEN** 持久化的 `value` 为 `a;b`（顺序以实现定义为准，校验时顺序不敏感）
- **AND** 详情回显可供前端还原为多选数组

### Requirement: 内置参数的类型与选项不可通过管理面修改

对 `is_builtin=1` 的参数记录，系统 MUST 允许在权限范围内修改 `value`（及既有可编辑展示字段策略），MUST 拒绝通过管理面创建/更新/导入修改其 `value_type` 与 `options`。非内置参数允许完整编辑类型与选项。

#### Scenario: 内置参数禁止改类型
- **WHEN** 管理员尝试将内置参数的 `valueType` 从 `boolean` 改为 `text`
- **THEN** 系统拒绝该字段变更
- **AND** 若仅提交 value 变更则仍可成功更新 value

#### Scenario: 自定义参数可改类型与选项
- **WHEN** 管理员更新非内置参数的 `valueType` 与 options
- **THEN** 系统在通过校验后持久化新类型与选项

### Requirement: 管理面必须按类型渲染编辑组件

参数设置新增与编辑界面 SHALL 根据 `valueType` 渲染对应输入组件，而不是对所有参数统一使用自由文本域：

| valueType | 组件行为 |
|-----------|----------|
| text | 单行输入 |
| textarea | 多行文本 |
| number | 数字输入 |
| boolean | 开关或等价二值选择 |
| select | 下拉，选项来自 options |
| radio | 单选组，选项来自 options |
| multi_select | 多选，选项来自 options |
| richtext | 宿主富文本编辑器（Tiptap 或等价） |

新增流程 MUST 允许选择 `valueType`；当类型为 `select`/`radio`/`multi_select` 时 MUST 提供 options 编辑能力。内置参数编辑时 MUST 锁定类型与 options 为只读。

管理面 MUST 按 `valueType` 应用可扩展的弹窗密度策略（而非写死单次高度常量）：

- **compact**（默认，短字段类型）：保持默认中等弹窗宽度，可不提供全屏入口。
- **spacious**（`richtext`，以及长文 `textarea`）：弹窗加宽至约 720–960px 量级且不超过视口、提供全屏入口；`richtext` 编辑区高度 MUST 使用视口相对尺寸（如 `clamp`/`vh`），编辑区内滚动，避免固定过矮的像素高度。

类型切换（新增时选择类型）与编辑加载 MUST 同步刷新弹窗 chrome 与值组件布局。

#### Scenario: 编辑下拉型内置参数通过选择赋值
- **WHEN** 管理员编辑 `valueType=select` 且已配置 options 的参数
- **THEN** 值输入区域展示下拉组件且选项来自 options
- **AND** 管理员可通过选择完成赋值而无需从备注复制字符串

#### Scenario: 编辑布尔参数使用开关组件
- **WHEN** 管理员编辑 `valueType=boolean` 的参数
- **THEN** 值输入区域展示开关或等价二值组件
- **AND** 提交值为 `true` 或 `false` 字符串

#### Scenario: 富文本参数使用宽弹窗与可全屏编辑区
- **WHEN** 管理员新增或编辑 `valueType=richtext` 的参数
- **THEN** 编辑弹窗宽度明显大于默认短表单弹窗（约 520px）
- **AND** 弹窗提供全屏入口
- **AND** 富文本编辑区最小高度不低于 360px 量级（或等价视口相对高度）
- **AND** 编辑区在内容变长时内部滚动而不是把弹窗压成不可用的窄条

#### Scenario: 从富文本切回短类型恢复紧凑弹窗
- **WHEN** 管理员在新增流程中将 `valueType` 从 `richtext` 改为 `text`（或其它 compact 类型）
- **THEN** 弹窗恢复默认紧凑宽度策略
- **AND** 全屏入口按 compact 策略隐藏或关闭
