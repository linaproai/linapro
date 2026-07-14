# config-management Specification

## Purpose

记录本分组对配置管理展示本地化与`i18n.check`交叉约束的可验收场景；完整配置管理行为以主规范`openspec/specs/config-management/spec.md`为准。

## Requirements

### Requirement: 内置系统参数名称和默认文案必须以英文本地化

配置管理页面 SHALL 按当前语言本地化内置系统参数名称、描述和默认显示值，使英文环境不显示默认中文系统文案。投影键 MUST 使用`config.<config_key>.name`与`config.<config_key>.remark`，其中`<config_key>`为`sys_config.key`原值。

#### Scenario: 登录和 IP 黑名单参数显示英文元数据

- **当** 管理员以`en-US`打开系统配置时
- **则** 内置登录、页面标题、页面描述、副标题和 IP 黑名单参数元数据以英文显示
- **且** 页面不显示这些参数的中文内置标签

#### Scenario: 内置公共前端文案可投射英文显示内容

- **当** 配置列表以`en-US`显示默认登录页标题、描述或副标题时
- **则** 可见显示内容使用英文投射或英文默认值
- **且** 编辑详情仍保留稳定的`configKey`和实际存储值

#### Scenario: 配置本地化资源保持完整

- **当** 内置配置翻译键被添加或更改时
- **则** 宿主全部运行时 locale 的`config.<config_key>.name`与`config.<config_key>.remark`保持覆盖
- **且** `make i18n.check`对缺失的内置配置展示键报告失败

### Requirement: 启用 i18n 的插件 sys_config 展示键必须可本地化

对`plugin.yaml`中`i18n.enabled: true`且以`SysConfigKey`常量声明的插件配置键，系统 SHALL 在插件`manifest/i18n/<locale>/`中提供`config.<config_key>.name`与`config.<config_key>.remark`，使参数设置页列表投影不为技术 key 裸展示。

#### Scenario: 插件 SysConfigKey 缺译被门禁阻断

- **当** 启用 i18n 的插件新增`hostconfigcap.SysConfigKey = "plugin.<id>...."`常量且未补齐对应`config.<key>.name/remark`
- **则** `make i18n.check`失败并指出缺失键与 locale
- **且** 未启用 i18n 的插件不要求上述插件侧资源
