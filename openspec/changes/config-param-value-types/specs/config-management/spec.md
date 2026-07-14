## ADDED Requirements

### Requirement: 参数管理 API 必须暴露 valueType 与 options

参数设置的列表、详情、按键查询、创建与更新契约 SHALL 包含 `valueType` 与 `options` 字段。创建与更新请求可提交上述字段；响应 MUST 返回持久化后的有效值。`options` 在 JSON 中为对象数组；空选项以空数组表示。

#### Scenario: 创建带类型的参数

- **WHEN** 管理员调用创建接口提交 `name`、`key`、`value`、`valueType=select` 与非空 `options`
- **THEN** 系统创建成功
- **AND** 详情响应包含相同的 `valueType` 与 `options`

#### Scenario: 列表项携带类型元数据

- **WHEN** 管理员查询参数列表
- **THEN** 每条 `ConfigItem` 包含 `valueType`
- **AND** 对枚举类型包含可用于编辑的 `options`

### Requirement: 配置导入导出必须包含类型与选项列

配置 Excel 导出、导入与导入模板 SHALL 在既有列基础上包含 `valueType` 与 `options` 列，列头通过 `config.field.valueType` 与 `config.field.options` 翻译键按当前语言解析。导入时缺失 `valueType` 视为 `text`；`options` 单元格存 JSON 文本。

#### Scenario: 导出包含类型列

- **WHEN** 管理员导出参数设置
- **THEN** Excel 包含本地化的 valueType 与 options 列头
- **AND** 行数据写出对应字段

#### Scenario: 导入非法 options JSON 行失败

- **WHEN** 管理员导入一行 `valueType=select` 且 options 单元格不是合法 JSON 数组
- **THEN** 该行计入失败列表并给出可理解原因
- **AND** 其他合法行仍可成功导入

### Requirement: 内置参数种子必须携带类型与选项元数据

宿主初始化 SQL SHALL 为已有内置运行时与公共前端参数写入匹配语义的 `value_type` 与（如需要）`options`，使管理面开箱按正确组件编辑。至少覆盖：

- 布尔开关类（如忘记密码入口、注册入口、水印开关）→ `boolean`
- 布局/主题等有限枚举 → `select` 或 `radio`，并给出完整 options
- 上传上限、日志保留天数等 → `number`
- 隐私政策/服务条款等长文 → `textarea` 或 `richtext`
- 其余短文案/路径/duration → `text` 或 `textarea`

#### Scenario: 登录框位置参数为下拉类型

- **WHEN** 管理员完成宿主数据库初始化后打开参数设置并编辑 `sys.auth.loginPanelLayout`
- **THEN** 该参数 `valueType` 为 `select`（或 `radio`）
- **AND** options 包含 `panel-left`、`panel-center`、`panel-right`
- **AND** 编辑界面以下拉或单选方式选择，而非自由文本猜测

#### Scenario: 忘记密码入口参数为布尔类型

- **WHEN** 管理员编辑 `sys.auth.forgetPasswordEnabled`
- **THEN** 该参数 `valueType` 为 `boolean`
- **AND** 编辑界面提供开关或二值选择组件
