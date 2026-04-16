# Notice Management

## Purpose

定义通知公告模块的数据结构、列表查询、详情维护和消息派发行为，确保公告内容能够被宿主统一管理并传递给目标用户。

## Requirements

### Requirement: 通知公告数据库表设计
系统 SHALL 提供 `sys_notice` 表存储通知公告数据。

#### Scenario: sys_notice 表结构
- **WHEN** 查看 `sys_notice` 表结构
- **THEN** 表包含：`id`（BIGINT PK AUTO_INCREMENT）、`title`（VARCHAR(255) 标题）、`type`（TINYINT 类型：1=通知 2=公告）、`content`（LONGTEXT 富文本内容）、`status`（TINYINT 状态：0=草稿 1=已发布）、`remark`（VARCHAR(500) 备注）、`created_by`（BIGINT 创建人ID）、`updated_by`（BIGINT 更新人ID）、`created_at`（DATETIME）、`updated_at`（DATETIME）、`deleted_at`（DATETIME 软删除）

### Requirement: 通知公告列表查询
系统 SHALL 提供通知公告的分页列表查询接口。

#### Scenario: 查询通知公告列表
- **WHEN** 调用 `GET /api/v1/notice` 并传入分页参数 `pageNum` 和 `pageSize`
- **THEN** 返回通知公告列表和总数，格式为 `{list: [...], total: number}`
- **THEN** 列表按创建时间倒序排列

#### Scenario: 通知公告列表支持条件筛选
- **WHEN** 查询时传入筛选参数 `title`（标题）、`type`（类型）或 `createdBy`（创建人）
- **THEN** `title` 使用模糊匹配（LIKE），`type` 精确匹配
- **THEN** `createdBy` 通过关联用户表匹配创建人用户名（模糊匹配）
- **THEN** 返回符合条件的通知公告列表

#### Scenario: 通知公告列表排除已删除记录
- **WHEN** 查询通知公告列表
- **THEN** 结果中不包含已软删除的记录

#### Scenario: 列表返回创建人名称
- **WHEN** 查询通知公告列表
- **THEN** 每条记录包含 `createdByName` 字段，为创建人的用户昵称

### Requirement: 获取通知公告详情
系统 SHALL 提供通知公告详情查询接口。

#### Scenario: 查询通知公告详情
- **WHEN** 调用 `GET /api/v1/notice/{id}`
- **THEN** 返回该通知公告的完整信息，包含富文本内容

#### Scenario: 查询不存在的通知公告
- **WHEN** 调用 `GET /api/v1/notice/{id}` 且该 ID 不存在
- **THEN** 系统返回错误信息

### Requirement: 创建通知公告
系统 SHALL 提供创建通知公告的接口。

#### Scenario: 创建通知公告成功
- **WHEN** 调用 `POST /api/v1/notice` 并提交 `title`、`type`、`content`、`status` 字段
- **THEN** 系统创建通知公告并自动记录 `created_by` 为当前登录用户ID
- **THEN** 返回成功

#### Scenario: 创建并直接发布通知
- **WHEN** 创建通知公告时 `status` 为 1（已发布）
- **THEN** 系统创建通知公告后，自动为所有活跃用户（status=1 且排除当前用户）创建 `sys_user_message` 消息记录

#### Scenario: 创建草稿通知
- **WHEN** 创建通知公告时 `status` 为 0（草稿）
- **THEN** 仅创建通知公告记录，不分发用户消息

#### Scenario: 必填字段校验
- **WHEN** 创建通知公告时缺少 `title`、`type` 或 `content`
- **THEN** 系统返回参数校验错误

### Requirement: 更新通知公告
系统 SHALL 提供更新通知公告的接口。

#### Scenario: 更新通知公告成功
- **WHEN** 调用 `PUT /api/v1/notice/{id}` 并提交要更新的字段
- **THEN** 系统更新对应通知公告信息，自动记录 `updated_by` 为当前登录用户ID

#### Scenario: 草稿更新为已发布
- **WHEN** 更新通知公告时将 `status` 从 0 改为 1
- **THEN** 系统更新通知公告状态后，自动为所有活跃用户创建 `sys_user_message` 消息记录

#### Scenario: 已发布通知再次编辑
- **WHEN** 更新一条已发布的通知公告内容（不改变 status）
- **THEN** 仅更新通知公告记录，不重复分发用户消息

#### Scenario: 更新不存在的通知公告
- **WHEN** 更新一个不存在的通知公告 ID
- **THEN** 系统返回错误信息

### Requirement: 删除通知公告
系统 SHALL 提供删除通知公告的接口，支持批量删除。

#### Scenario: 删除通知公告成功
- **WHEN** 调用 `DELETE /api/v1/notice` 并传入 `ids` 参数（逗号分隔的ID列表）
- **THEN** 对应通知公告被软删除（设置 `deleted_at`）

### Requirement: 通知公告字典数据
系统 SHALL 提供通知公告相关的字典数据。

#### Scenario: 初始化通知类型字典
- **WHEN** 执行 v0.4.0 数据库迁移脚本
- **THEN** 创建字典类型 `sys_notice_type`（通知类型），包含字典数据：通知(1)、公告(2)

#### Scenario: 初始化公告状态字典
- **WHEN** 执行 v0.4.0 数据库迁移脚本
- **THEN** 创建字典类型 `sys_notice_status`（公告状态），包含字典数据：草稿(0)、已发布(1)

### Requirement: 通知公告管理前端列表页
系统 SHALL 提供通知公告管理列表页面。

#### Scenario: 列表页展示
- **WHEN** 用户进入通知公告管理页面
- **THEN** 以 VXE-Grid 表格展示通知公告列表，支持分页
- **THEN** 显示列：公告标题、公告类型（字典渲染）、状态（字典渲染）、创建人、创建时间
- **THEN** 支持复选框多选

#### Scenario: 搜索筛选
- **WHEN** 用户在搜索栏输入标题、选择类型或输入创建人并点击搜索
- **THEN** 表格刷新显示符合条件的通知公告

#### Scenario: 新增通知公告
- **WHEN** 用户点击"新增"按钮
- **THEN** 弹出弹窗（800px 宽），包含标题、状态（RadioButton）、类型（RadioButton）、内容（Tiptap 编辑器）字段

#### Scenario: 编辑通知公告
- **WHEN** 用户点击某条记录的"编辑"按钮
- **THEN** 弹出弹窗并回显通知公告信息，修改后提交更新

#### Scenario: 删除通知公告
- **WHEN** 用户点击某条记录的"删除"按钮
- **THEN** 弹出确认对话框，确认后删除该通知公告，列表自动刷新

#### Scenario: 批量删除
- **WHEN** 用户勾选多条记录后点击工具栏"删除"按钮
- **THEN** 弹出确认对话框，确认后批量删除选中的通知公告

### Requirement: 通知公告菜单与权限
系统 SHALL 在系统管理菜单下新增通知公告菜单项。

#### Scenario: 菜单显示
- **WHEN** 用户登录后查看侧边栏
- **THEN** 系统管理分组下显示"通知公告"菜单项

#### Scenario: 权限控制
- **WHEN** 通知公告页面渲染操作按钮
- **THEN** 新增按钮受 `system:notice:add` 权限控制
- **THEN** 编辑按钮受 `system:notice:edit` 权限控制
- **THEN** 删除按钮受 `system:notice:remove` 权限控制
