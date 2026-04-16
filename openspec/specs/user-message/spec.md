# User Message

## Purpose

定义用户消息的数据结构、未读统计、列表查询、已读处理与清空行为，确保系统能够稳定管理用户侧的站内消息体验。

## Requirements

### Requirement: 用户消息数据库表设计
系统 SHALL 提供 `sys_user_message` 表存储用户消息数据。

#### Scenario: sys_user_message 表结构
- **WHEN** 查看 `sys_user_message` 表结构
- **THEN** 表包含：`id`（BIGINT PK AUTO_INCREMENT）、`user_id`（BIGINT 接收用户ID）、`title`（VARCHAR(255) 消息标题）、`type`（TINYINT 消息类型：1=通知 2=公告）、`source_type`（VARCHAR(50) 来源类型）、`source_id`（BIGINT 来源ID）、`is_read`（TINYINT 是否已读：0=未读 1=已读）、`read_at`（DATETIME 阅读时间）、`created_at`（DATETIME 创建时间）
- **THEN** 在 `user_id` + `is_read` 上建立联合索引

### Requirement: 查询未读消息数量
系统 SHALL 提供查询当前用户未读消息数量的接口。

#### Scenario: 获取未读消息数
- **WHEN** 调用 `GET /api/v1/user/message/count`
- **THEN** 返回当前登录用户的未读消息总数 `{count: number}`

#### Scenario: 无未读消息
- **WHEN** 当前用户没有未读消息
- **THEN** 返回 `{count: 0}`

### Requirement: 查询消息列表
系统 SHALL 提供查询当前用户消息列表的接口。

#### Scenario: 获取消息列表
- **WHEN** 调用 `GET /api/v1/user/message` 并传入分页参数
- **THEN** 返回当前登录用户的消息列表，按创建时间倒序排列
- **THEN** 每条消息包含：`id`、`title`、`type`、`sourceType`、`sourceId`、`isRead`、`readAt`、`createdAt`

### Requirement: 标记消息已读
系统 SHALL 提供标记消息为已读的接口。

#### Scenario: 标记单条消息已读
- **WHEN** 调用 `PUT /api/v1/user/message/{id}/read`
- **THEN** 将该消息的 `is_read` 设为 1，`read_at` 设为当前时间
- **THEN** 仅允许操作当前登录用户自己的消息

#### Scenario: 标记所有消息已读
- **WHEN** 调用 `PUT /api/v1/user/message/read-all`
- **THEN** 将当前登录用户的所有未读消息标记为已读

### Requirement: 删除消息
系统 SHALL 提供删除消息的接口。

#### Scenario: 删除单条消息
- **WHEN** 调用 `DELETE /api/v1/user/message/{id}`
- **THEN** 物理删除该消息记录
- **THEN** 仅允许删除当前登录用户自己的消息

#### Scenario: 清空所有消息
- **WHEN** 调用 `DELETE /api/v1/user/message/clear`
- **THEN** 物理删除当前登录用户的所有消息记录

### Requirement: 消息通知铃铛组件
系统 SHALL 在页面顶部导航栏提供消息通知铃铛组件。

#### Scenario: 铃铛图标展示
- **WHEN** 用户登录后查看顶部导航栏
- **THEN** 显示铃铛图标，当有未读消息时在图标右上角显示未读数量徽标

#### Scenario: 无未读消息
- **WHEN** 用户没有未读消息
- **THEN** 铃铛图标不显示数量徽标

#### Scenario: 定时轮询
- **WHEN** 用户登录后
- **THEN** 前端每 60 秒自动调用未读消息数量接口更新徽标

### Requirement: 消息面板
系统 SHALL 提供消息面板，用户点击铃铛图标后展开。

#### Scenario: 打开消息面板
- **WHEN** 用户点击铃铛图标
- **THEN** 展开 Popover 消息面板，显示消息列表
- **THEN** 每条消息显示标题、消息类型、时间
- **THEN** 未读消息有视觉标识（如加粗或圆点标记）

#### Scenario: 点击消息跳转详情
- **WHEN** 用户点击消息面板中的某条消息
- **THEN** 自动标记该消息为已读
- **THEN** 跳转至通知公告详情页 `/system/notice/detail/{sourceId}`

#### Scenario: 全部已读
- **WHEN** 用户点击消息面板中的"全部已读"按钮
- **THEN** 调用标记所有消息已读接口，更新面板中所有消息状态

#### Scenario: 清空消息
- **WHEN** 用户点击消息面板中的"清空"按钮
- **THEN** 弹出确认对话框，确认后调用清空接口，消息面板清空

#### Scenario: 删除单条消息
- **WHEN** 用户在消息面板中点击某条消息的删除按钮
- **THEN** 调用删除单条消息接口，该消息从面板中移除

### Requirement: 通知公告详情页
系统 SHALL 提供通知公告详情页面。

#### Scenario: 详情页展示
- **WHEN** 用户访问 `/system/notice/detail/{id}`
- **THEN** 页面显示通知公告的标题、类型（字典渲染）、创建人、创建时间
- **THEN** 以富文本方式展示公告内容

#### Scenario: 从消息面板跳转
- **WHEN** 用户从消息面板点击消息跳转至详情页
- **THEN** 该消息自动标记为已读

### Requirement: 消息 Store（Pinia）
系统 SHALL 提供 Pinia Store 管理用户消息状态。

#### Scenario: 初始化轮询
- **WHEN** 用户登录后进入宿主工作区
- **THEN** 消息 Store 启动 60 秒间隔轮询，获取未读消息数量

#### Scenario: 停止轮询
- **WHEN** 用户退出登录
- **THEN** 消息 Store 停止轮询

#### Scenario: 未读数量响应式更新
- **WHEN** 轮询获取到新的未读数量
- **THEN** Store 中的 `unreadCount` 响应式更新，铃铛徽标同步变化

#### Scenario: 预留 SSE 扩展点
- **WHEN** 查看消息 Store 的轮询实现
- **THEN** 轮询逻辑封装在独立方法中，未来可替换为 SSE 监听而无需修改 Store 外部接口
