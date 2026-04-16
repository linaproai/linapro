## Context

Lina 是基于 GoFrame + Vue 3 (Vben5) 的后台管理系统。当前已具备用户管理、部门管理、岗位管理、字典管理等基础模块。本次 v0.4.0 迭代新增通知公告管理和用户消息中心功能，同时引入 Tiptap 富文本编辑器作为通用组件。

现有技术栈：后端 GoFrame v2 + MySQL，前端 Vue 3 + Ant Design Vue + VXE-Grid + Pinia。

## Goals / Non-Goals

**Goals:**
- 实现通知公告的完整 CRUD 管理功能（草稿/发布状态）
- 实现用户消息中心，支持右上角铃铛未读提示、消息面板、消息详情查看
- 引入 Tiptap 富文本编辑器通用组件，支持基础排版和图片插入
- 消息分发采用 Fan-out on Write，发布通知时为所有活跃用户创建消息记录
- 前端采用 60 秒轮询获取未读消息数量

**Non-Goals:**
- 不实现 SSE/WebSocket 实时推送（预留扩展点）
- 不集成 OSS 文件存储（图片暂用 Base64 内联，预留接口扩展）
- 不实现通知的定向发送（本期发给所有用户，不做角色/部门筛选）
- 不实现消息通知偏好设置页面

## Decisions

### 决策 1：消息分发策略 — Fan-out on Write

**选择**: 发布通知时为每个活跃用户创建 `sys_user_message` 记录。

**备选方案**: Fan-out on Read（仅存通知，用 `sys_notice_read` 记录已读状态，查询时动态计算未读）。

**理由**: 管理后台用户量通常在千级以内，写入压力可控。Fan-out on Write 使查询简单（直接查 `sys_user_message` 表），支持用户独立删除消息，且未读计数查询高效（单表 `COUNT`）。

### 决策 2：实时性方案 — HTTP 轮询先行

**选择**: 前端每 60 秒轮询 `GET /api/v1/user/message/count` 获取未读数量，用户打开面板时拉取消息列表。

**备选方案**: SSE (Server-Sent Events) 实时推送。

**理由**: 通知公告是低频操作（每天 1-2 条），60 秒延迟完全可接受。轮询方案的后端是标准 REST API，与现有架构一致，无需额外的连接管理。未来升级到 SSE 只需增加推送通道，不影响数据模型。

**扩展点设计**: 前端消息 Store 的 `startPolling()` 方法封装为可替换接口，未来切换到 SSE 时仅需替换此方法实现。

### 决策 3：富文本编辑器 — Tiptap

**选择**: 使用 Tiptap 编辑器（`@tiptap/vue-3` + `@tiptap/starter-kit`）。

**备选方案**: Tinymce（参考项目使用）、WangEditor。

**理由**: Tiptap 基于 ProseMirror，体积小、扩展性强、Vue 3 集成好。Tinymce 需要额外加载 400KB+ 的库文件且需要 license，WangEditor 社区维护不够活跃。Tiptap 的模块化扩展机制便于后续按需添加功能。

**图片处理**: 暂时使用 Base64 内联存储。通过抽象 `uploadHandler` 回调函数，后续接入 OSS 时仅需替换此函数实现。

### 决策 4：消息删除策略 — 物理删除

**选择**: 用户删除消息时物理删除 `sys_user_message` 记录。

**备选方案**: 软删除（增加 `deleted_at` 字段）。

**理由**: 用户消息是个人行为数据，无需回收站或审计追溯。物理删除减少数据膨胀，清空全部消息时 `DELETE` 语句简洁高效。

### 决策 5：通知详情页路由设计

**选择**: 新增 `/system/notice/detail/:id` 路由作为通知详情页，用户从消息面板点击消息后跳转至此页面查看完整内容。

**理由**: 通知内容为富文本 HTML，在消息面板 Popover 中展示空间有限且体验不佳。独立详情页可完整展示通知内容，且 URL 可分享。

## Risks / Trade-offs

- **[用户量膨胀]** Fan-out on Write 在用户量大时写入放大明显 → 管理后台场景用户量可控（<10000），暂不构成问题。若未来用户量增长，可切换至 Fan-out on Read 或引入消息队列异步处理
- **[轮询性能]** 60 秒轮询产生定期请求 → `GET /api/v1/user/message/count` 仅返回一个数字，查询走索引（`user_id + is_read`），性能开销极低
- **[Base64 图片]** 富文本中图片以 Base64 存储会增大数据库记录体积 → 本期可接受，后续接入 OSS 时通过 `uploadHandler` 替换即可
- **[Tiptap 依赖]** 引入新的前端依赖包 → Tiptap 核心包体积小（~50KB gzipped），且采用模块化按需引入
