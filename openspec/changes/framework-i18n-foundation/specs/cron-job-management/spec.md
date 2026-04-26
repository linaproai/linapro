## ADDED Requirements

### Requirement: 内置定时任务展示元数据必须由后端本地化
系统 SHALL 在定时任务管理、任务分组管理和执行日志接口中，按当前请求语言返回内置调度数据的展示名称、描述和备注。宿主、源码插件和动态插件注册的内置定时任务 MUST 使用稳定 `handlerRef`、插件 ID、任务名或分组编码作为后端翻译锚点；前端不得通过中文原文、`handlerRef` 或分组编码维护任务名称与分组名称翻译映射。

#### Scenario: 代码注册内置任务使用英文源文案
- **WHEN** 宿主或插件注册内置定时任务、任务处理器或动态插件 cron contract
- **THEN** 注册链路中的 `Name`、`DisplayName` 和 `Description` 源文案使用可读英文
- **AND** 中文等非英文展示通过后端运行时 i18n 资源或插件 i18n 资源返回
- **AND** `handlerRef`、插件 ID、任务名和分组编码保持稳定，不作为用户可见翻译结果直接展示

#### Scenario: 查询任务列表时返回本地化名称
- **WHEN** 管理员以 `en-US` 请求 `GET /job`
- **THEN** 返回的内置任务 `name`、`description` 和 `groupName` 已经按英文投影
- **AND** 用户自建任务的 `name`、`description` 和所属自建分组名称保持数据库原值
- **AND** 前端任务列表直接渲染接口返回值，不再调用前端种子任务映射函数

#### Scenario: 查询任务分组时返回本地化默认分组
- **WHEN** 管理员以 `en-US` 请求 `GET /job-group`
- **THEN** 默认分组的 `name` 和 `remark` 已经按英文投影
- **AND** 用户自建分组继续返回数据库原值
- **AND** 前端分组列表和任务列表中的所属分组展示保持一致

#### Scenario: 查询执行日志时返回本地化任务名
- **WHEN** 管理员以 `en-US` 请求 `GET /job/log` 或 `GET /job/log/{id}`
- **THEN** 返回的内置任务日志 `jobName` 已经按英文投影
- **AND** 后端可使用日志快照中的稳定 `handlerRef` 或任务锚点解析当前语言展示值
- **AND** 前端日志列表和详情不再解析快照后执行本地翻译
