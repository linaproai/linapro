## Why

插件业务参数经 `HostConfig.SysConfig().SetValue` 落库到 `sys_config` 后，会与宿主内置参数一并出现在「参数设置」管理面，形成系统页与插件设置页的双轨维护。持久化可继续统一使用 `sys_config`，但管理入口必须按数据标记分流，且能力契约必须允许插件显式声明是否进入系统参数页。

## What Changes

- 为 `sys_config` 增加 `system_manageable` 字段（`1`=可在系统参数页维护，`0`=否）。
- 系统参数 List / Export / 详情 / 增删改 / 导入仅面向 `system_manageable = 1`。
- 扩展 `HostConfig.SysConfig().SetValue` 为四参数 `(ctx, key, value, options)`，其中 `options *SetSysConfigValueOptions` 可控制 `SystemManageable`。
- 首次插入且未指定时默认 `0`；插件 settings 保存路径显式传 `false`。
- 宿主 seed / 管理面新建默认 `1`。
- 运行时读路径不受影响。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `config-management`：管理面按 `system_manageable` 过滤与锁定；插件 `SetValue` 能力扩展管理面标记。

## Impact

- **数据库**：`005-config-management.sql` 建表含 `system_manageable`。
- **能力契约**：`SetValue(ctx, key, value, options *SetSysConfigValueOptions)`；动态 JSON 可选 `systemManageable`。
- **插件**：storage / oidc / ldap 等 settings 写入显式 `gconv.PtrBool(false)`。
- **i18n**：`SYSCONFIG_SYSTEM_MANAGE_DENIED` 中英错误文案。
- **缓存 / 数据权限 / DI**：无语义变更；无新增运行期依赖。
