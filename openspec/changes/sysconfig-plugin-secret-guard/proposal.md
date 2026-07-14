# 系统参数：插件配置展示（已还原）

## 状态

**已撤销。** 曾要求在系统参数管理面隐藏/锁定 `plugin.*`，后确认该需求有误。

## 当前行为

- 系统参数 List/Export/**有什么展示什么**，不对 `plugin.*` 做过滤。
- 插件配置仍通过 `HostConfig.SysConfig().SetValue` 写入 `sys_config`。
- 管理面 Create/Update/Delete/Import/Get 对 `plugin.*` 不再做额外拒绝（与改动前一致；内置 `is_builtin`/托管键保护仍生效）。

## 非目标

- 不改变插件设置页与 `SetValue` 存储路径。
