# 后端拥有数据的前端翻译映射审查

## 审查范围

- 源码扫描：`apps/lina-vben/apps/web-antd/src` 中所有 `localizeSeed*`、`display-l10n`、按当前语言分支的硬编码映射。
- 使用点核对：确认哪些映射仍被页面实际引用，哪些只是历史遗留函数。
- 后端链路核对：对照角色、定时任务、任务分组与执行日志接口，确认接口是否已经按请求语言返回展示值。

## 结论

- 当前仍在页面实际生效的前端侧后端数据翻译集中在角色管理、定时任务管理、任务分组和执行日志。
- `display-l10n.ts` 保留了大量历史映射函数，覆盖部门、岗位、角色、配置、字典、通知、操作日志、登录日志、插件、菜单和调度数据。多数函数当前未被引用，但这些函数本身仍把后端业务数据的翻译知识固化在前端，后续容易被误用。
- 后端拥有的数据必须在后端根据请求语言完成本地化投影；前端只消费接口返回值并负责静态界面文案、组件布局和纯展示格式化。

## 实际在用的问题点

1. 角色管理列表
   - 文件：`apps/lina-vben/apps/web-antd/src/views/system/role/data.ts`
   - 问题：`admin` 角色名称通过 `localizeSeedRoleName(row.key, row.name)` 在前端投影为英文。
   - 期望：角色列表接口在当前语言下返回内置受保护角色的本地化展示值；可编辑角色仍返回数据库原值。

2. 任务分组列表
   - 文件：`apps/lina-vben/apps/web-antd/src/adapter/vxe-table.ts`
   - 文件：`apps/lina-vben/apps/web-antd/src/views/system/job-group/index.vue`
   - 问题：默认分组名称与备注通过 `localizeSeedJobGroupName`、`localizeSeedJobGroupRemark` 在前端翻译。
   - 期望：`GET /job-group` 按请求语言返回默认分组的本地化名称与备注；用户自建分组仍返回数据库原值。

3. 定时任务列表
   - 文件：`apps/lina-vben/apps/web-antd/src/adapter/vxe-table.ts`
   - 问题：任务名称和所属分组通过 `localizeSeedJobName`、`localizeSeedJobGroupName` 在前端翻译。
   - 期望：`GET /job` 按请求语言返回内置任务名称、描述和分组名称；用户自建任务名称仍返回数据库原值。

4. 执行日志列表与详情
   - 文件：`apps/lina-vben/apps/web-antd/src/adapter/vxe-table.ts`
   - 文件：`apps/lina-vben/apps/web-antd/src/views/system/job-log/detail.vue`
   - 问题：执行日志中的任务名通过解析 `jobSnapshot.handlerRef` 后在前端翻译。
   - 期望：`GET /job/log` 与 `GET /job/log/{id}` 根据当前语言返回本地化任务名；日志快照中的稳定 `handlerRef` 或翻译锚点用于后端解析，前端不再维护映射表。

## 历史遗留映射清理范围

`apps/lina-vben/apps/web-antd/src/utils/display-l10n.ts` 当前包含以下后端数据映射，应在实现阶段整体清理或拆除：

- 组织与权限：`localizeSeedDept*`、`localizeSeedPostName`、`localizeSeedRole*`
- 系统设置：`localizeSeedConfig*`、`localizeSeedDictType*`
- 内容与日志：`localizeSeedNotice*`、`localizeSeedOperLog*`、`localizeSeedLoginLogMessage`
- 插件与菜单：`localizeDynamicPluginSeedRecord*`、`localizeSeedPlugin*`、`localizeSeedMenuName`
- 调度中心：`localizeSeedJobGroup*`、`localizeSeedJob*`

这些映射不应迁移到新的前端工具函数中；需要保留展示逻辑时，应改为后端翻译字段、运行时消息键或业务内容多语言模型。

## 不属于本次问题的前端逻辑

- `permission-display.ts` 对权限字符串进行分段格式化，属于把稳定权限码转换为可读展示的 UI 格式化，不是按后端业务文本做中英文翻译。该类逻辑可以保留，但不得扩展为业务数据翻译映射。
- 英文环境下表格列宽、表单标签宽度、分页选择器宽度等布局适配逻辑可以保留，因为它们不改变后端数据语义。

## 后续修正要求

- 定时任务注册时，宿主和插件内置任务的源文案统一使用英文 `Name`、`Description`、`DisplayName`，非英文展示通过后端 i18n 资源投影。
- 对已经持久化在 `sys_job`、`sys_job_group`、`sys_job_log` 中的内置治理数据，后端接口必须基于稳定锚点返回本地化展示值。
- 前端移除 `display-l10n.ts` 对后端拥有数据的映射依赖，并同步更新 E2E：断言接口返回值已经本地化，而不是只断言表格 formatter 显示结果。
