## 1. 数据库与基础设施

- [x] 1.1 创建 v0.4.0 SQL 文件：`sys_notice` 表 DDL、`sys_user_message` 表 DDL、字典种子数据（`sys_notice_type`、`sys_notice_status`）、菜单权限数据
- [x] 1.2 执行 `make init` 更新数据库，执行 `make dao` 生成 DAO/DO/Entity 代码

## 2. 后端 — 通知公告管理

- [x] 2.1 创建通知公告 API 定义（`api/notice/v1/`）：列表查询、详情、创建、更新、删除接口的 Request/Response 结构体
- [x] 2.2 执行 `make ctrl` 生成 Controller 骨架，注册路由到 `cmd_http.go`
- [x] 2.3 实现通知公告 Service 层：列表查询（支持标题模糊搜索、类型筛选、创建人关联查询）、详情查询
- [x] 2.4 实现通知公告 Service 层：创建、更新（含发布时 fan-out 消息分发逻辑）、删除
- [x] 2.5 实现通知公告 Controller 层：填充各接口的业务逻辑调用

## 3. 后端 — 用户消息

- [x] 3.1 创建用户消息 API 定义（`api/usermsg/v1/`）：未读数量、消息列表、标记已读、标记全部已读、删除、清空接口
- [x] 3.2 执行 `make ctrl` 生成 Controller 骨架，注册路由
- [x] 3.3 实现用户消息 Service 层：未读数量查询、消息列表分页查询
- [x] 3.4 实现用户消息 Service 层：标记已读、标记全部已读、删除单条、清空全部（物理删除）
- [x] 3.5 实现用户消息 Controller 层：填充各接口的业务逻辑调用

## 4. 前端 — Tiptap 富文本编辑器组件

- [x] 4.1 安装 Tiptap 依赖包（`@tiptap/vue-3`、`@tiptap/starter-kit`、`@tiptap/extension-image`、`@tiptap/extension-link`、`@tiptap/extension-placeholder`、`@tiptap/extension-underline`）
- [x] 4.2 实现 Tiptap 编辑器组件（`src/components/tiptap/`）：主组件 `editor.vue`、工具栏 `toolbar.vue`、扩展配置 `extensions.ts`，支持 v-model、disabled、height props，图片以 Base64 内联并预留 uploadHandler 扩展点

## 5. 前端 — 通知公告管理页面

- [x] 5.1 创建通知公告 API 文件（`src/api/system/notice/`）：列表查询、详情、创建、更新、删除接口
- [x] 5.2 创建通知公告管理路由配置，添加至系统管理菜单下
- [x] 5.3 实现通知公告列表页（`src/views/system/notice/index.vue`）：VXE-Grid 表格、搜索栏（标题/类型/创建人）、工具栏按钮（新增/批量删除）、行操作（编辑/删除）
- [x] 5.4 实现通知公告新增/编辑弹窗（`notice-modal.vue`）：标题、状态 RadioButton、类型 RadioButton、Tiptap 编辑器内容字段
- [x] 5.5 实现通知公告详情页（`src/views/system/notice/detail.vue`）：展示标题、类型、创建人、创建时间和富文本内容

## 6. 前端 — 用户消息中心

- [x] 6.1 创建用户消息 API 文件（`src/api/system/message/`）：未读数量、消息列表、标记已读、标记全部已读、删除、清空接口
- [x] 6.2 实现消息 Pinia Store（`src/store/message.ts`）：未读数量状态、60 秒轮询逻辑、启动/停止方法，预留 SSE 扩展点
- [x] 6.3 实现消息通知铃铛组件（`src/layouts/` 或 `src/components/`）：铃铛图标 + 未读数量徽标 + Popover 消息面板
- [x] 6.4 实现消息面板功能：消息列表展示、点击跳转详情、全部已读、清空、删除单条
- [x] 6.5 集成铃铛组件到顶部导航栏 layout，登录后启动轮询，退出时停止轮询

## 7. E2E 测试

- [x] 7.1 创建通知公告 Mock 数据 SQL 文件（`manifest/sql/mock-data/`）
- [x] 7.2 编写 E2E 测试：TC0037 通知公告 CRUD（新增、编辑、删除通知公告）
- [x] 7.3 编写 E2E 测试：TC0038 通知公告搜索筛选（按标题、类型、创建人搜索）
- [x] 7.4 编写 E2E 测试：TC0039 通知公告发布与消息分发（发布后铃铛显示未读数量）
- [x] 7.5 编写 E2E 测试：TC0040 消息面板操作（查看消息、标记已读、删除、清空）
- [x] 7.6 编写 E2E 测试：TC0041 通知公告详情页（从消息面板跳转查看详情）
- [x] 7.7 运行全部 E2E 测试套件，确认无回归

## Feedback

- [x] **FB-1**：通知公告列表创建人列显示用户数字ID，应关联sys_user表返回用户名（如admin）
- [x] **FB-2**：通知公告新增/编辑弹窗中删除无必要的"备注"字段
- [x] **FB-3**：消息通知轮询时间间隔硬编码为60秒，应支持可配置
- [x] **FB-4**：消息通知面板"查看所有消息"按钮点击无响应，viewAll事件未绑定处理函数
- [x] **FB-5**：发布通知/公告后其他用户右上角消息铃铛无法收到未读提示
- [x] **FB-6**：草稿发布后其他用户收不到消息通知，排查 fan-out 逻辑和前端轮询是否正确触发
- [x] **FB-7**：消息通知面板悬浮预览内容为空，应显示通知内容摘要而非类型标签
- [x] **FB-8**：点击消息列表项无法跳转到通知详情页，需在 NotificationItem 映射中添加 link 属性
- [x] **FB-9**：创建独立的用户消息列表页面（/system/message），支持分页、删除、清空，替换当前"查看所有"跳转目标
- [x] **FB-10**：通知详情页使用 Card 组件包裹内容区域，提升视觉层次感
- [x] **FB-11**：通知管理页面操作列增加"预览"按钮，点击弹窗展示通知公告内容
- [x] **FB-12**：删除通知公告时应级联删除 sys_user_message 中对应的消息记录
- [x] **FB-13**：通知预览弹窗主体内容与顶部信息表格间距过小，需增加空白
- [x] **FB-14**：消息列表页"删除"按钮缺少边框，应与其他页面删除按钮样式保持一致
- [x] **FB-15**：消息列表页点击消息应弹出预览窗口查看详情，而非跳转到新页面
- [x] **FB-16**：右上角消息通知列表点击消息应弹出预览窗口查看详情，同时删除消息详情页及其路由
