## ADDED Requirements

### Requirement: i18n.check 必须校验 sys_config 展示元数据翻译覆盖

`linactl i18n.check`（及 `make i18n.check`）SHALL 在既有硬编码扫描、locale 对等、`bizerr` messageKey、插件管理展示键与前端静态 `$t` 覆盖之外，校验参数设置投影所需的 `config.<config_key>.name` 与 `config.<config_key>.remark` 覆盖。

#### Scenario: 宿主 SQL seed 与受保护常量键缺失翻译时失败
- **WHEN** 宿主 `manifest/sql` 的 `sys_config` seed 或宿主声明的 `sys.*`/`demo.*` 配置常量缺少对应 `config.<key>.name` 或 `.remark`
- **THEN** `linactl i18n.check` 以非零退出码失败
- **AND** 输出指出 scope、locale 与缺失键

#### Scenario: 启用 i18n 的插件 SysConfigKey 缺失翻译时失败
- **WHEN** `i18n.enabled: true` 的插件声明 `SysConfigKey` 常量但插件运行时语言包缺少 `config.<key>.name` 或 `.remark`
- **THEN** `linactl i18n.check` 失败
- **AND** 未启用 i18n 的插件被跳过该项检查
