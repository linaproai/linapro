## Why

系统缺少通知公告管理功能，管理员无法向用户发布通知和公告信息。同时用户没有消息中心来接收和查看这些通知公告，缺少信息触达渠道。本次迭代新增通知公告管理模块和用户消息中心，并引入 Tiptap 富文本编辑器作为通用组件支撑公告内容编辑。

## What Changes

- 新增**通知公告管理**功能：支持通知/公告的创建、编辑、删除、发布，管理员可通过富文本编辑器编写公告内容
- 新增**用户消息中心**功能：页面右上角铃铛图标展示未读消息数，点击展开消息面板查看消息列表，支持标记已读、删除单条、清空全部消息，点击消息可跳转通知/公告详情页
- 新增 **Tiptap 富文本编辑器**通用组件：支持基础排版、图片（暂用 Base64 内联，预留 OSS 扩展点）、链接等功能
- 新增字典数据：`sys_notice_type`（通知类型）、`sys_notice_status`（公告状态）
- 消息分发采用 **Fan-out on Write** 策略：发布通知时为每个活跃用户创建消息记录
- 前端采用 **60 秒轮询**获取未读消息数量，预留 SSE 实时推送扩展点

## Capabilities

### New Capabilities

- `notice-management`: 通知公告管理模块，包含通知/公告的 CRUD、发布功能及管理列表页面
- `user-message`: 用户消息中心，包含右上角铃铛通知、消息面板、已读/删除/清空操作、通知详情页
- `tiptap-editor`: Tiptap 富文本编辑器通用组件，支持排版、图片上传、链接等功能

### Modified Capabilities

- `base-layout`: 页面顶部导航栏增加消息通知铃铛图标组件

## Impact

- **数据库**：新增 `sys_notice`、`sys_user_message` 两张表，新增字典种子数据
- **后端**：新增 notice、user-message 两个模块的 API/Controller/Service/DAO
- **前端**：新增通知公告管理页面、通知详情页、消息面板组件、Tiptap 编辑器组件
- **路由/菜单**：系统管理下新增"通知公告"菜单项
- **依赖**：前端新增 `@tiptap/vue-3`、`@tiptap/starter-kit` 等 Tiptap 相关依赖包
- **权限**：新增 `system:notice:list`、`system:notice:add`、`system:notice:edit`、`system:notice:remove` 权限码
