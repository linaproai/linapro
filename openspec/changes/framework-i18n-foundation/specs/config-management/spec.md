## ADDED Requirements

### Requirement: 配置元数据必须支持按当前语言返回本地化名称与备注
The system SHALL return localized config metadata for config list, detail, import/export templates, and protected-setting projections. Config metadata localization MUST use stable config keys as translation anchors and MUST NOT change the actual config key or stored config value.

#### Scenario: 查询配置列表时返回英文元数据
- **WHEN** 管理员以 `en-US` 查询配置列表或配置详情
- **THEN** 返回结果中的配置名称和备注使用英文本地化值
- **AND** `configKey` 与 `configValue` 仍保持原始治理语义

#### Scenario: 配置元数据翻译缺失时回退默认语言
- **WHEN** 某个配置项在当前语言下缺少名称或备注翻译
- **THEN** 系统回退到默认语言元数据或基线名称
- **AND** 配置读写能力不受影响

### Requirement: 公共前端配置文案必须支持国际化投影
The system SHALL let the public frontend config endpoint return localized brand and authentication copy according to the current request language, while keeping non-textual fields such as layout and theme mode stable.

#### Scenario: 登录页公共配置返回英文文案
- **WHEN** 浏览器以 `en-US` 请求公共前端配置接口
- **THEN** 返回的应用名称、登录页标题、登录页说明和登录副标题均为英文本地化结果
- **AND** `panelLayout`、`themeMode`、`layout` 等非文案字段保持原值

#### Scenario: 工作台刷新后显示最新本地化品牌文案
- **WHEN** 管理员更新了某语言下的公共前端文案并刷新登录页或工作台
- **THEN** 刷新后的页面显示新的本地化品牌名称和登录展示文案
- **AND** 不需要修改页面组件代码
