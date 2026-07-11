# 插件管理列表增加「管理」入口

## Why

插件管理列表当前只能查看治理信息并执行安装/启用/卸载等生命周期操作。运维人员要进入某个插件的业务管理页（如 LDAP 设置、登录日志、通知管理）时，必须先在左侧菜单中自行定位，路径长且容易遗漏。需要在列表操作列提供直达入口，并在插件无管理页时明确置灰，避免无效点击。

## What Changes

- 在插件管理列表每一行操作列中增加「管理」按钮。
- 点击后跳转到该插件的管理页面（宿主已装配的可导航插件页面）。
- 若该插件不存在管理页面，按钮置灰且不可点击。
- 补充中英文 UI 文案与 E2E 验收覆盖。

## Capabilities

### New Capabilities

- `plugin-management-manage-entry`：插件管理列表操作列的「管理」入口语义、启用/置灰规则与跳转行为。

### Modified Capabilities

- `plugin-ui-integration`：在既有插件管理列表操作语义上补充「管理」入口。

## Impact

- 前端：`apps/lina-vben/apps/web-antd/src/views/system/plugin/index.vue` 操作列与路由跳转。
- 前端：插件页面注册表用于判断是否存在管理页及解析跳转目标。
- i18n：宿主前端 `pages.system.plugin.actions.manage` 等文案。
- 测试：插件管理 E2E 与相关单元测试。
- 无后端 API / 数据模型 / 权限码变更。
