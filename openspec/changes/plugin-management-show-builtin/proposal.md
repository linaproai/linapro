## Why

`distribution=builtin` 插件当前被普通插件管理列表默认隐藏，管理员无法在统一入口看到内建能力、查看详情或进入插件管理页。内建插件仍是项目交付面的一部分，应可见、可识别、只读治理，而不是从管理台完全消失。

## What Changes

- **BREAKING（列表语义）**：普通插件管理列表默认包含 `distribution=builtin` 插件，不再依赖诊断参数 `includeBuiltin` 才能看见。
- 列表与详情 UI 为 builtin 插件增加「内置插件」标识；若同时命中宿主 `plugin.autoEnable`，继续展示既有「自动启用」类标识。
- builtin 插件在页面上禁止安装、启用/禁用、卸载、手动升级与租户供应策略变更（隐藏写操作入口，非仅置灰）。
- 允许查看详情；已安装且存在管理页时，允许点击「管理」进入插件管理界面。
- 服务端写操作拒绝边界保持不变；`includeBuiltin` 查询参数调整为兼容字段（可忽略或始终等价于包含）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `plugin-manifest-lifecycle`：普通插件列表投影默认包含 builtin，并继续暴露 `distribution`。
- `plugin-ui-integration`：插件管理 UI 展示 builtin 及标识，隐藏写操作，保留详情与管理入口。

## Impact

- **后端**：`plugin` 列表服务过滤逻辑、`ListReq.includeBuiltin` 文档语义、相关单元测试。
- **前端**：插件管理列表/详情标识与操作列、i18n（中/英）。
- **测试**：`TC016` 及依赖「默认隐藏 builtin」的断言需改为「默认展示 + 只读治理」。
- **API 契约**：列表默认返回集合变大；调用方若曾假设默认无 builtin，需按 `distribution` 区分。
- **数据权限 / 缓存 / 开发工具**：无新增边界；列表仍为只读投影。
