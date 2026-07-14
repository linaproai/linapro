## MODIFIED Requirements

### Requirement:内置系统参数名称和默认文案必须以英文本地化

配置管理页面 SHALL 按当前语言本地化内置系统参数名称、描述和默认显示值，使英文环境不显示默认中文系统文案。投影键 MUST 使用 `config.<config_key>.name` 与 `config.<config_key>.remark`，其中 `<config_key>` 为 `sys_config.key` 原值。

列表、按 key 查询与**按 ID 的编辑详情** MUST 对 `name`/`remark` 返回当前请求语言投影。编辑详情的 `value` MUST 保持库内实际存储值，不得用公共前端默认文案投影覆盖可编辑值。

#### Scenario:登录和 IP 黑名单参数显示英文元数据
- **当** 管理员以 `en-US` 打开系统配置时
- **则** 内置登录、页面标题、页面描述、副标题和 IP 黑名单参数元数据以英文显示
- **且** 页面不显示这些参数的中文内置标签

#### Scenario:内置公共前端文案可投射英文显示内容
- **当** 配置列表以 `en-US` 显示默认登录页标题、描述或副标题时
- **则** 可见显示内容使用英文投射或英文默认值
- **且** 编辑详情仍保留稳定的 `configKey` 和实际存储的 `value`

#### Scenario:英文环境编辑内置参数时元数据为英文
- **当** 管理员以 `en-US` 打开某内置参数的编辑详情
- **则** 详情中的 `name` 与 `remark` 为英文投影
- **且** `value` 等于库内存储值（含管理员自定义后的原文）
- **且** 编辑表单不展示该参数的中文 seed 名称或描述

#### Scenario:配置本地化资源保持完整
- **当** 内置配置翻译键被添加或更改时
- **则** 宿主全部运行时 locale 的 `config.<config_key>.name` 与 `config.<config_key>.remark` 保持覆盖
- **且** `make i18n.check` 对缺失的内置配置展示键报告失败

## ADDED Requirements

### Requirement:内置系统参数更新不得写回本地化名称与描述

系统 SHALL 在更新内置（`isBuiltin` 或受管系统键）配置记录时忽略请求中的 `name` 与 `remark`，仅允许在既有规则下更新可编辑字段（至少包含 `value`）。非内置自定义参数仍可更新 `name` 与 `remark`。

#### Scenario:内置参数保存不污染 name/remark
- **当** 调用方以 `en-US` 获取内置参数详情并将投影后的英文 `name`/`remark` 连同新 `value` 提交更新
- **则** 系统更新 `value`
- **且** 库内 `name` 与 `remark` 仍为更新前的存储原文
- **且** 后续中文环境下列表/详情仍可从 i18n 或库内 fallback 得到正确展示

#### Scenario:自定义参数仍可修改名称与备注
- **当** 管理员更新非内置参数的 `name` 或 `remark`
- **则** 系统按请求写入对应字段
