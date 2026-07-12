## Why

菜单管理列表中「状态」「是否显示」仅以字典标签展示，管理员需要进入编辑抽屉才能切换，效率低且与插件管理等列表的开关交互不一致。同时，父级目录停用或隐藏后子级菜单仍保持独立状态，容易造成导航与权限语义混乱。需要补齐父级级联收敛与列表快捷开关能力。

## What Changes

- 当菜单状态或显示字段被写入时，系统自动将其目标值同步到所有后代菜单
- 停用/隐藏与启用/显示均级联，保证父子子树状态一致
- 菜单管理列表的「状态」「是否显示」列改为开关组件，支持行内快速切换
- 切换成功后刷新列表树与当前用户可访问菜单/路由，使左侧导航即时收敛

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `menu-management`：补充状态/显隐级联停用与隐藏语义；补充列表页状态与显示开关交互要求

## Impact

- 后端：`apps/lina-core/internal/service/menu` 的 `Update` 级联写入后代 `status`/`visible`
- 前端：`apps/lina-vben/apps/web-antd/src/views/system/menu` 列表列与开关处理
- 前端 API：复用现有 `PUT /menu/{id}`，行内开关提交最小必要字段
- 测试：菜单服务单元测试 + 菜单管理 E2E
- 缓存/权限拓扑：沿用现有 `NotifyAccessTopologyChanged` 失效路径
- i18n：复用已有启用/停用、显示/隐藏文案，无新增键
