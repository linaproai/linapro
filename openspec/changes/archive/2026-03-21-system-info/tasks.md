## 1. 后端：系统信息 API

- [x] 1.1 创建 SQL 文件 `manifest/sql/v0.5.0.sql`（本版本无 DDL 变更，仅作为版本占位）
- [x] 1.2 创建 API 定义 `api/sysinfo/v1/info.go`，定义 `GET /api/v1/system/info` 的请求/响应结构体（返回 Go 版本、GoFrame 版本、OS、数据库版本、启动时间、运行时长）
- [x] 1.3 执行 `make ctrl` 生成控制器骨架
- [x] 1.4 实现 `internal/service/sysinfo/sysinfo.go` 系统信息服务层，获取运行时信息
- [x] 1.5 在控制器中调用服务层并返回数据
- [x] 1.6 在 `internal/cmd/cmd_http.go` 中注册系统信息路由（鉴权路由组内）

## 2. 前端：路由与菜单结构

- [x] 2.1 创建路由模块 `src/router/routes/modules/about.ts`，定义"系统信息"顶级菜单及三个子路由（系统接口、系统信息、组件演示）
- [x] 2.2 创建视图目录 `src/views/about/`，包含 `api-docs/index.vue`、`system-info/index.vue`、`component-demo/index.vue` 三个页面组件骨架

## 3. 前端：系统接口页面（Scalar OpenAPI UI）

- [x] 3.1 安装 `@scalar/api-reference` npm 依赖
- [x] 3.2 实现 `src/views/about/api-docs/index.vue`，集成 Scalar Vue 组件，加载后端 `/api.json`
- [x] 3.3 创建前端配置文件 `src/views/about/config.ts`，定义 OpenAPI 规范地址、组件演示地址、项目信息、后端/前端组件列表等可配置项

## 4. 前端：系统信息页面

- [x] 4.1 创建 API 文件 `src/api/about/index.ts`，定义 `getSystemInfo` 方法调用 `GET /api/v1/system/info`
- [x] 4.2 实现 `src/views/about/system-info/index.vue`，包含四个 Card 区块：关于项目、基本信息、后端组件、前端组件
- [x] 4.3 关于项目区块：从配置对象读取项目名称、描述、版本、许可证、主页链接
- [x] 4.4 基本信息区块：调用后端 API 展示运行时数据（Go 版本、OS、数据库版本、启动时间、运行时长等）
- [x] 4.5 后端/前端组件区块：从配置对象读取组件列表，以网格布局展示名称、版本、可点击外链

## 5. ~~前端：组件演示页面~~（已取消）

- [x] ~~5.1 实现组件演示页面~~（已取消）
- [x] ~~5.2 实现 iframe 加载失败检测~~（已取消）

## 6. E2E 测试

- [x] 6.1 创建 `hack/tests/e2e/about/` 测试目录
- [x] 6.2 编写 `TC0044-api-docs-page.ts`：验证系统接口页面加载 Scalar UI 正常展示
- [x] 6.3 编写 `TC0045-system-info-page.ts`：验证系统信息页面四个区块正常展示，后端数据正确加载
- [x] ~~6.4 编写 `TC0046-component-demo-page.ts`~~（已取消）
- [x] 6.5 运行全部 E2E 测试确认无回归（114 passed，6 failed 均为已有问题，新增 3 个测试全部通过）

## Feedback

- [x] **FB-1**：~~系统接口页面 Scalar API Client 弹窗被遮挡~~ → 已改用 Stoplight Elements 替代 Scalar，通过 Web Component 方式集成，样式完全隔离无冲突
- [x] **FB-2**：系统接口页面顶部空白过多，需减少顶部间距
- [x] **FB-3**：系统接口页面左侧 Overview 菜单点击后右侧内容为空白
- [x] **FB-4**：系统接口页面左侧接口分类菜单应粗体展示
- [x] **FB-5**：系统接口页面 Stoplight CSS 污染全局页面样式（边框消失等），改用 iframe 嵌入实现样式隔离
- [x] **FB-6**：系统接口文档 HTML 页面应改为静态文件方式提供，移除后端 API 路由，减少系统复杂度
- [x] **FB-7**：将系统信息子菜单的"系统信息"标题修改为"版本信息"
- [x] **FB-8**：移除头像下拉菜单中的"文档"、"Github"、"问题 & 帮助"菜单项
- [x] **FB-9**：修正头像下拉菜单中的用户邮箱（后端增加 email 字段）和昵称显示（无昵称时展示用户名）
- [x] **FB-10**：个人中心上传头像后提示上传成功，但页面未展示新上传的头像
- [x] **FB-11**：个人中心基本设置页面中昵称、邮箱、手机号码、性别不应标记为必填项
- [x] **FB-12**：页面右上角头像未展示用户头像，用户未设置头像时应展示默认头像
- [x] **FB-13**：后端创建用户时，如果昵称为空则默认设置为用户名；后端 API 层和前端创建/编辑表单增加昵称必填校验
- [x] **FB-14**：个人中心基本设置页面昵称增加必填校验，后端 UpdateProfile API 增加昵称必填校验
- [x] **FB-15**：用户创建/修改页面选择部门后，岗位列表未正确展示子部门岗位（后端 getDeptAndDescendantIds 使用 || 拼接字符串在 MySQL 中无效，需改用 FIND_IN_SET 或 CONCAT）
- [x] **FB-16**：用户创建/修改页面"所属部门"字段名称改为"部门"，移至岗位字段上方，改为非必填项
- [x] **FB-17**：用户列表中用户不属于任何部门时，部门列应展示"未分配部门"而非空白
- [x] **FB-18**：将部门层级查询中的 MySQL 特有 FIND_IN_SET 替换为基于 parent_id 迭代查询的跨数据库通用实现（涉及 dept/post/user 三个服务共 3 处）
- [x] **FB-19**：通知公告 NoticeModal 中 formRules 使用普通对象导致 Vue warn "Invalid watch source"，需改为 reactive 对象
- [x] **FB-20**：系统接口页面嵌套的接口文档 iframe 中 html/body 背景透明，导致外层页面灰色背景透过 iframe 渗透到 Stoplight 侧边栏与内容区之间的分隔条区域
- [x] **FB-21**：系统接口文档左侧菜单中只有模块名称粗体展示，接口名称不应粗体展示
- [x] **FB-22**：去掉系统接口文档左下角的"powered by Stoplight"展示
- [x] **FB-23**：接口文档中接口地址背景块（如"GET /api/v1/notice"）长度应与当前板块宽度一致，GET 在左侧、接口地址在右侧
- [x] **FB-24**：系统接口文档左侧 SCHEMAS 区域做成可折叠的（默认折叠）
- [x] **FB-25**：用户管理页面状态字段改为从字典模块 `sys_normal_disable` 读取标签，包括查询表单、创建/编辑表单和表格 Switch
- [x] **FB-26**：字典管理页面修改/新增/删除字典数据后，未清除 dictStore 缓存，导致其他页面状态标签不同步
- [x] **FB-27**：部门管理页面查询表单和创建/编辑表单的状态选项改为从字典模块 sys_normal_disable 动态读取
- [x] **FB-28**：岗位管理页面查询表单、创建/编辑表单和表格 Switch 的状态选项改为从字典模块 sys_normal_disable 动态读取
- [x] **FB-29**：编写 E2E 测试用例验证字典修改后全局生效（修改字典标签 → 验证其他页面同步更新）
- [x] **FB-30**：系统接口文档左侧菜单顶部的"API文档"标题需要隐藏，仅保留 Overview
- [x] **FB-31**：系统接口文档中每个接口详情页展示的请求地址是前端地址，应改为后端接口地址（需在 OpenAPI spec 中配置 servers）
- [x] **FB-32**：OpenAPI 接口文档需要使用当前登录用户的 JWT Token 信息，方便在页面中直接请求测试接口
- [x] **FB-33**：后端 OpenAPI 文档信息（标题、描述、版本等）从硬编码改为通过配置文件配置
- [x] **FB-34**：去掉版本信息页面顶部的系统信息介绍板块
- [x] **FB-35**：重新布局"关于项目"板块：项目名称+项目介绍放第一行，其他信息放第二行；"项目描述"改为"项目介绍"
- [x] **FB-36**：动态获取 GoFrame 版本（当前硬编码为 "v2.10.0"，改用 gf.VERSION 运行时获取）
- [x] **FB-37**：将前后端组件展示信息移至后端配置文件配置，新增 API 返回组件信息，前端从 API 获取（GoFrame 描述改为"Go语言应用开发框架"，分析 go.mod 中关键第三方组件加入后端组件列表）
- [x] **FB-38**：去掉版本信息页面顶部的"版本信息"标题栏
- [x] **FB-39**：减少"关于项目"板块中项目名称与项目介绍之间的空白间距
- [x] **FB-40**：暗黑模式下面包屑菜单中链接颜色显示为主题蓝色（Ant Design Vue 全局 a 标签样式覆盖），应与参考项目一致显示为 muted-foreground 颜色
- [x] **FB-41**：完善 auth 模块接口文档：为 g.Meta 添加 dc 标签，为所有输入输出字段补充 dc 和 eg 标签（3个文件）
- [x] **FB-42**：完善 user 模块接口文档：为 g.Meta 添加 dc 标签，为所有输入输出字段补充 dc 和 eg 标签（13个文件）
- [x] **FB-43**：完善 dept 模块接口文档：为 g.Meta 添加 dc 标签，为所有输入输出字段补充 dc 和 eg 标签（8个文件）
- [x] **FB-44**：完善 post 模块接口文档：为 g.Meta 添加 dc 标签，为所有输入输出字段补充 dc 和 eg 标签（8个文件）
- [x] **FB-45**：完善 dict 模块接口文档：为 g.Meta 添加 dc 标签，为所有输入输出字段补充 dc 和 eg 标签（14个文件）
- [x] **FB-46**：完善 notice 模块接口文档：为 g.Meta 添加 dc 标签，为所有输入输出字段补充 dc 和 eg 标签（5个文件）
- [x] **FB-47**：完善 loginlog 模块接口文档：为 g.Meta 添加 dc 标签，为所有输入输出字段补充 dc 和 eg 标签（5个文件）
- [x] **FB-48**：完善 operlog 模块接口文档：为 g.Meta 添加 dc 标签，为所有输入输出字段补充 dc 和 eg 标签（5个文件）
- [x] **FB-49**：完善 usermsg 模块接口文档：为 g.Meta 添加 dc 标签，为所有输入输出字段补充 dc 和 eg 标签（6个文件）
- [x] **FB-50**：完善 sysinfo 模块接口文档：为 g.Meta 添加 dc 标签，为所有输入输出字段补充 dc 和 eg 标签（1个文件）
