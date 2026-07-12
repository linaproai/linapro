## 1. 后端列表投影

- [x] 1.1 更新 `ListReq.includeBuiltin` 文档语义为兼容字段，并移除默认隐藏 builtin 的列表过滤
- [x] 1.2 更新 `plugin_list_test`：默认列表包含 builtin 且返回 `distribution`

## 2. 前端标识与只读操作

- [x] 2.1 插件列表名称列增加「内置插件」标识（可与自动启用标识并存）
- [x] 2.2 确认 builtin 隐藏安装/启停/卸载/升级/租户策略，保留详情与管理入口
- [x] 2.3 详情弹窗展示内置插件相关只读信息（如适用）
- [x] 2.4 补充中英文 i18n 文案

## 3. 测试与校验

- [x] 3.1 更新 `TC016`：默认展示 builtin、内置标识、写操作隐藏、详情可用
- [x] 3.2 运行相关单测与 E2E / 静态校验
- [x] 3.3 `openspec validate plugin-management-show-builtin --strict`
