## 1. 后端项目初始化

- [x] 1.1 创建 GoFrame v2 后端项目骨架（`apps/lina-core/`），包含标准目录结构：api、internal/cmd、internal/controller、internal/service、internal/dao、internal/model、manifest/config、manifest/sql
- [x] 1.2 配置 `go.mod`，引入 GoFrame v2、SQLite driver（`github.com/gogf/gf/contrib/drivers/sqlite/v2`）、JWT（`github.com/golang-jwt/jwt/v5`）、bcrypt（`golang.org/x/crypto`）
- [x] 1.3 编写 `manifest/config/config.yaml`，配置 server（端口 8080）、logger、database（SQLite）、jwt（secret、expireHour）
- [x] 1.4 编写 `manifest/sql/init.sql`，创建 `sys_user` 表并插入默认管理员账号（admin/admin123 bcrypt 哈希）
- [x] 1.5 编写 `internal/cmd/cmd.go`，实现服务启动、路由注册框架（公开路由组 + 认证路由组）
- [x] 1.6 编写 `Makefile`（`apps/lina-core/Makefile`），支持 build、ctrl、dao、service 命令
- [x] 1.7 验证后端项目可编译运行并监听 8080 端口

## 2. 后端认证模块

- [x] 2.1 编写认证 API 定义（`api/auth/v1/`）：LoginReq/LoginRes、LogoutReq/LogoutRes
- [x] 2.2 实现认证服务（`internal/service/auth/`）：Login（验证用户名密码、签发 JWT）、ParseToken（解析验证 JWT）、HashPassword/VerifyPassword（bcrypt）
- [x] 2.3 实现认证控制器（`internal/controller/auth/`）：登录和登出处理
- [x] 2.4 实现中间件服务（`internal/service/middleware/`）：CORS、HandlerResponse（统一响应格式）、Auth（JWT 认证）
- [x] 2.5 实现业务上下文服务（`internal/service/bizctx/`）：从请求上下文提取当前用户信息
- [x] 2.6 定义常量（`internal/consts/`）和上下文模型（`internal/model/context.go`）

## 3. 后端用户管理模块

- [x] 3.1 编写用户 API 定义（`api/user/v1/`）：ListReq/ListRes、CreateReq/CreateRes、UpdateReq/UpdateRes、DeleteReq/DeleteRes、GetReq/GetRes、UpdateStatusReq/UpdateStatusRes、GetProfileReq/GetProfileRes、UpdateProfileReq/UpdateProfileRes
- [x] 3.2 使用 `make dao` 生成 DAO/Entity/DO 代码（或手动编写 `internal/dao/`、`internal/model/entity/`、`internal/model/do/`）
- [x] 3.3 实现用户服务（`internal/service/user/`）：List（分页查询+条件筛选）、Create（用户名查重+bcrypt密码）、Update、Delete（软删除+防止删除管理员和自己）、GetById、UpdateStatus（防止停用自己）、GetProfile、UpdateProfile
- [x] 3.4 实现用户控制器（`internal/controller/user/`）：绑定各 API 处理方法
- [x] 3.5 在 `cmd.go` 中注册认证路由和用户管理路由

## 4. 前端项目初始化

- [x] 4.1 使用 Vben5 最新版官方脚手架初始化前端项目到 `apps/lina-vben/`，选择 Ant Design Vue 变体
- [x] 4.2 清理模板示例页面，配置项目名称为 Lina
- [x] 4.3 配置 API 代理：开发环境将 `/api` 前缀请求代理到 `http://localhost:8080`
- [x] 4.4 配置请求客户端（`src/api/request.ts`）：Bearer Token 认证头、统一错误处理、401 自动跳转登录

## 5. 前端认证模块

- [x] 5.1 实现登录 API（`src/api/core/auth.ts`）：loginApi、logoutApi
- [x] 5.2 适配 Vben5 的认证流程：对接 login、logout、getUserInfo 接口
- [x] 5.3 配置路由守卫：未登录跳转登录页、401 响应处理

## 6. 前端布局与菜单

- [x] 6.1 配置管理后台基础布局：侧边栏 + 顶部导航栏 + 内容区域
- [x] 6.2 配置静态菜单：系统管理分组下设"用户管理"菜单项
- [x] 6.3 配置顶部导航栏：显示当前用户信息、退出登录按钮

## 7. 前端用户管理页面

- [x] 7.1 实现用户管理 API（`src/api/system/user/`）：userList、userAdd、userUpdate、userDelete、userInfo、userStatusChange、getProfile、updateProfile
- [x] 7.2 实现用户管理列表页（`src/views/system/user/index.vue`）：VXE-Grid 表格 + 搜索栏（用户名、状态）+ 分页 + 操作按钮（新增/编辑/删除/状态切换）
- [x] 7.3 实现用户管理数据定义（`src/views/system/user/data.tsx`）：表格列定义、查询表单 schema、抽屉表单 schema
- [x] 7.4 实现用户新增/编辑抽屉（`src/views/system/user/user-drawer.vue`）：表单校验、创建/更新逻辑
- [x] 7.5 配置用户管理路由

## 8. 开发环境配置

- [x] 8.1 编写根目录 `Makefile`：dev（启动前后端）、stop（停止服务）、status（查看状态）
- [x] 8.2 编写 `CLAUDE.md`：项目概述、常用命令、架构说明、开发规范
- [x] 8.3 更新 `.gitignore`：补充 backend/lina-vben 相关忽略规则

## 9. E2E 测试

- [x] 9.1 初始化 Playwright 测试项目（`hack/tests/`），配置 playwright.config.ts
- [x] 9.2 编写 TC0001：登录成功流程（输入正确用户名密码 → 跳转到管理后台）
- [x] 9.3 编写 TC0002：登录失败流程（错误密码 → 显示错误提示）
- [x] 9.4 编写 TC0003：登出流程（点击退出 → 跳转到登录页）
- [x] 9.5 编写 TC0004：用户管理 CRUD 完整流程（新增用户 → 列表查询 → 编辑用户 → 状态切换 → 删除用户）
- [x] 9.6 编写 TC0005：未登录访问受保护页面（直接访问管理后台 URL → 重定向到登录页）
- [x] 9.7 所有 E2E 测试通过

## Feedback

- [x] **FB-1**：后端用户列表接口支持字段排序（增加 orderBy/orderDirection 参数，支持 id、username、nickname、phone、email、status、created_at 字段排序）
- [x] **FB-2**：前端用户管理列表表头支持点击排序（VXE-Grid 列头排序联动后端接口）
- [x] **FB-3**：后端用户列表接口增强搜索（支持 nickname 模糊搜索、beginTime/endTime 创建时间范围筛选）
- [x] **FB-4**：前端搜索表单联动后端（确保 nickname、createdAt 时间范围等筛选参数正确传递到后端）
- [x] **FB-5**：后端实现用户导出接口（GET /api/v1/user/export，返回 Excel 文件）
- [x] **FB-6**：前端实现用户导出功能（点击导出按钮下载 Excel 文件）
- [x] **FB-7**：后端实现用户导入接口（POST /api/v1/user/import，解析 Excel 并批量创建用户）和导入模板下载接口（GET /api/v1/user/import-template）
- [x] **FB-8**：前端实现用户导入功能（上传 Excel 文件、显示导入结果）
- [x] **FB-9**：创建 100 条测试用户数据（通过 SQL 初始化脚本，覆盖各种状态和字段）
- [x] **FB-10**：编写排序功能 E2E 测试用例
- [x] **FB-11**：编写搜索功能 E2E 测试用例
- [x] **FB-12**：编写导出功能 E2E 测试用例
- [x] **FB-13**：编写导入功能 E2E 测试用例
- [x] **FB-14**：用户列表可排序列头鼠标指针应显示为手型（pointer），当前仅排序小图标有手型
- [x] **FB-15**：用户列表导出功能执行报错，无法正常导出 Excel 文件
- [x] **FB-16**：用户导入弹出框 UI 需改为 UploadDragger 拖拽上传样式，增加文件类型提示、下载模板链接（带 Excel 图标）、"是否更新/覆盖已存在的用户数据" Switch 开关，参考 ruoyi-plus-vben5 项目
- [x] **FB-17**：用户编辑表单中 RadioGroup 单选项样式改为 button 样式（optionType: 'button', buttonStyle: 'solid'），与参考项目保持一致
- [x] **FB-18**：实现用户列表页面的重置密码功能（后端新增重置密码 API + 前端弹出框显示用户信息及新密码输入）
- [x] **FB-19**：将 UI 参考规范写入 CLAUDE.md，要求所有 UI 设计和实现参考 ruoyi-plus-vben5 项目保持一致性
- [x] **FB-20**：工具栏的导出和删除按钮仅在勾选了用户行时显示，且导出仅导出选中的用户而非全部用户
- [x] **FB-21**：前端用户管理列表中，当前登录用户的行禁止操作（禁用编辑、删除、状态切换、重置密码等按钮，禁止勾选复选框）
- [x] **FB-22**：数据库 sys_user 表增加 avatar 字段，重新生成 DAO/Entity/DO
- [x] **FB-23**：后端新增头像上传接口（POST /user/profile/avatar），支持 multipart/form-data 文件上传，保存文件到本地并返回 URL
- [x] **FB-24**：后端配置静态文件服务，使上传的头像文件可通过 URL 访问
- [x] **FB-25**：后端 GetInfo 接口返回真实的 avatar 字段值（当前硬编码为空字符串）
- [x] **FB-26**：前端注册个人中心路由（/profile），使顶部用户下拉菜单中的"个人中心"链接可用
- [x] **FB-27**：前端创建 CropperAvatar 头像裁剪上传组件（参考 ruoyi-plus-vben5 项目的 cropper 组件）
- [x] **FB-28**：前端个人中心页面对接真实后端 API，基本设置页展示并支持修改个人信息和头像
- [x] **FB-29**：前端用户管理列表页增加头像列，使用 Avatar 组件展示用户头像
- [x] **FB-30**：个人中心页面右侧样式调整，参考 ruoyi-plus-vben5 项目布局和样式保持 UI 一致（增加 overflow-hidden、表单响应式宽度等）
- [x] **FB-31**：个人中心页面去掉右侧菜单中的"安全设置"Tab
- [x] **FB-32**：个人中心页面左侧面板增加"上次登录"信息（数据库增加 login_date 字段、后端登录时记录时间、前端展示）
- [x] **FB-33**：个人中心右侧面板样式修复：基本设置和修改密码表单的提交按钮应使用 useVbenForm 内置的 submitButtonOptions 渲染（而非手动添加外部按钮），与参考项目保持一致
- [x] **FB-34**：将个人中心左侧面板的"手机号码"展示描述改为"手机"
- [x] **FB-35**：个人中心基本设置表单字段增加必填校验（昵称 required、邮箱 email 格式校验、手机 正则校验），参考 ruoyi-plus-vben5 项目
- [x] **FB-36**：数据库 sys_user 表增加 sex 字段（TINYINT DEFAULT 0），重新生成 DAO/Entity/DO，后端 API 定义（CreateReq/UpdateReq/UpdateProfileReq/ListReq）增加 Sex 字段，后端 Service 层支持 sex 字段的读写和筛选
- [x] **FB-37**：前端个人中心基本设置表单增加性别字段（RadioGroup，button 样式，选项：未知/男/女，必填），前端用户管理抽屉表单增加性别字段
- [x] **FB-38**：前端用户管理表格增加性别列，后端 Excel 导出/导入增加性别字段
- [x] **FB-39**：将数据库 SQL 初始化从后端启动流程中剥离，改为独立的 `make init` 命令（后端增加 init 子命令，根 Makefile 增加 init target）

