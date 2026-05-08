## 1. 数据库与基础设施

- [x] 1.1 创建操作日志和登录日志 SQL 文件，包含 `sys_oper_log` 和 `sys_login_log` 建表语句，以及 `sys_oper_type` 字典类型和字典数据的 Seed DML
- [x] 1.2 创建在线用户和服务监控 SQL 文件，新增 `sys_online_session`（MEMORY 引擎）和 `sys_server_monitor` 表的 DDL，新增系统监控菜单数据
- [x] 1.3 创建参数设置 SQL 文件，新增 `sys_config` 表 DDL 和菜单/按钮权限种子数据
- [x] 1.4 执行 `make init` 更新数据库，执行 `make dao` 生成 DAO/DO/Entity 文件
- [x] 1.5 添加 User-Agent 解析依赖（`mssola/useragent`）和系统指标采集依赖（`github.com/shirou/gopsutil/v4`）

## 2. 后端 - 审计日志模块

### 2.1 登录日志

- [x] 2.1.1 创建 `api/loginlog/v1/` 接口定义：List、Get、Clean、Export、Delete（批量删除）
- [x] 2.1.2 执行 `make ctrl` 生成控制器骨架
- [x] 2.1.3 实现 `internal/service/loginlog/` 服务层：Create、List、Get、Clean、Export、Delete
- [x] 2.1.4 填写控制器方法实现，在 `cmd_http.go` 中注册登录日志路由

### 2.2 操作日志

- [x] 2.2.1 创建 `api/operlog/v1/` 接口定义：List、Get、Clean、Export、Delete（批量删除）
- [x] 2.2.2 执行 `make ctrl` 生成控制器骨架
- [x] 2.2.3 实现 `internal/service/operlog/` 服务层：Create、List、Get、Clean、Export、Delete
- [x] 2.2.4 填写控制器方法实现，在 `cmd_http.go` 中注册操作日志路由

### 2.3 操作日志中间件

- [x] 2.3.1 实现操作日志中间件 `internal/service/middleware/operlog.go`：拦截写操作、解析 `g.Meta` 标签获取模块名和操作类型、记录请求参数（截断+脱敏）和响应结果（截断）、计算耗时、异步写入数据库
- [x] 2.3.2 在 `cmd_http.go` 中将操作日志中间件注册到 Auth 中间件之后
- [x] 2.3.3 为现有导出接口的 `g.Meta` 添加 `operLog:"4"` 标签

## 3. 后端 - 认证模块改造

- [x] 3.1 定义会话存储抽象接口 `SessionStore`，实现基于 MySQL 的 `DBSessionStore`（创建、查询、删除、列表过滤、TouchOrValidate、CleanupInactive）
- [x] 3.2 改造认证服务（`internal/service/auth/`）：登录成功后创建会话记录并写入登录日志，登出时删除会话记录并写入登录日志
- [x] 3.3 改造认证中间件（`internal/service/middleware/`）：在 JWT 校验后通过 TouchOrValidate 检查会话有效性，会话不存在时返回 401
- [x] 3.4 实现不活跃会话自动清理定时任务，超时阈值和清理频率通过配置文件调整

## 4. 后端 - 在线用户模块

- [x] 4.1 创建在线用户 API 定义（`api/monitor/v1/`）：`GET /monitor/online/list`、`DELETE /monitor/online/{tokenId}`
- [x] 4.2 实现在线用户 Controller 和 Service：列表查询、强制下线逻辑
- [x] 4.3 注册在线用户路由到 `cmd_http.go`

## 5. 后端 - 服务监控模块

- [x] 5.1 实现指标采集服务（CPU、内存、磁盘、网络、Go 运行时、服务器基本信息）
- [x] 5.2 实现定时采集任务：服务启动时立即采集一次，此后每 30 秒采集一次，采用 UPSERT 策略每个节点只保留最新一条记录
- [x] 5.3 创建服务监控 API 定义（`api/monitor/v1/`）：`GET /monitor/server`
- [x] 5.4 实现服务监控 Controller 和 Service：读取数据库中各节点最新监控数据
- [x] 5.5 注册服务监控路由和定时采集任务到 `cmd_http.go`

## 6. 后端 - 系统信息模块

- [x] 6.1 创建 API 定义 `api/sysinfo/v1/info.go`，定义 `GET /api/v1/system/info` 的请求/响应结构体
- [x] 6.2 实现 `internal/service/sysinfo/sysinfo.go` 系统信息服务层，获取运行时信息
- [x] 6.3 填写控制器方法实现，在 `cmd_http.go` 中注册系统信息路由（鉴权路由组内）

## 7. 后端 - 参数设置模块

- [x] 7.1 创建 API 定义 `api/config/v1/`：List、Get、Create、Update、Delete、ByKey、Export、Import、ImportTemplate（7 个文件）
- [x] 7.2 执行 `make ctrl` 生成控制器骨架
- [x] 7.3 实现 `internal/service/sysconfig/` 服务层：完整 CRUD、按键名查询、导出、导入（覆盖/忽略模式）
- [x] 7.4 填写控制器方法实现，在 `cmd_http.go` 中注册参数设置路由

## 8. 后端 - 字典导出导入优化

- [x] 8.1 新增字典合并导出接口（`GET /dict/export`），同时导出字典类型和字典数据到双 Sheet Excel 文件
- [x] 8.2 新增字典合并导入接口（`POST /dict/import`），支持同时导入字典类型和字典数据
- [x] 8.3 新增字典导入模板下载接口，返回包含两个 Sheet 的模板文件
- [x] 8.4 字典类型删除逻辑改为级联删除，删除字典类型时同时删除关联的字典数据

## 9. 后端 - 定时任务与配置重构

- [x] 9.1 将所有定时任务从 gtimer 迁移到 gcron 组件，使用 crontab 表达式
- [x] 9.2 将所有硬编码配置读取改为 struct 维护，按配置分组创建配置结构体
- [x] 9.3 将 `internal/config/` 迁移到 `internal/service/config/`，按模块拆分为独立 Go 文件
- [x] 9.4 将 cmd_http.go 中的定时任务逻辑提取到 service/cron 独立组件中封装

## 10. 前端 - 操作日志页面

- [x] 10.1 创建操作日志 API 层：`src/api/monitor/operlog/`
- [x] 10.2 创建操作日志列表页：`src/views/monitor/operlog/index.vue` 和 `data.ts`（表格+筛选）
- [x] 10.3 创建操作日志详情抽屉组件，请求参数和响应结果使用 vue-json-pretty 实现 JSON 代码高亮
- [x] 10.4 实现清理功能（弹窗选择时间范围后硬删除）和批量删除功能

## 11. 前端 - 登录日志页面

- [x] 11.1 创建登录日志 API 层：`src/api/monitor/loginlog/`
- [x] 11.2 创建登录日志列表页：`src/views/monitor/loginlog/index.vue` 和 `data.ts`
- [x] 11.3 创建登录日志详情弹窗组件
- [x] 11.4 实现清理功能和批量删除功能

## 12. 前端 - 在线用户页面

- [x] 12.1 创建前端 API 文件 `src/api/monitor/online/`
- [x] 12.2 创建在线用户页面 `src/views/monitor/online/index.vue` 和 `data.ts`：搜索表单、VXE-Grid 表格、工具栏在线人数统计、强制下线 Popconfirm
- [x] 12.3 新增系统监控路由模块 `src/router/routes/modules/monitor.ts`

## 13. 前端 - 服务监控页面

- [x] 13.1 创建前端 API 文件 `src/api/monitor/server/`
- [x] 13.2 创建服务监控页面 `src/views/monitor/server/index.vue` 及子组件：服务器信息卡片、CPU/内存圆形进度条、Go 运行时信息、磁盘使用表格、网络流量信息
- [x] 13.3 实现多节点切换逻辑和可折叠节点列表布局

## 14. 前端 - 系统信息页面

- [x] 14.1 创建路由模块 `src/router/routes/modules/about.ts`，定义"系统信息"顶级菜单及三个子路由
- [x] 14.2 创建前端配置文件 `src/views/about/config.ts`，定义可配置项
- [x] 14.3 实现系统接口页面：iframe 嵌入 Stoplight Elements 静态文档页面
- [x] 14.4 实现版本信息页面：关于项目、后端组件、前端组件三个区块
- [x] 14.5 实现组件演示页面：iframe 嵌入 vben5 官网演示，含加载失败处理
- [x] 14.6 创建 API 文件 `src/api/about/index.ts`，调用 `GET /api/v1/system/info`

## 15. 前端 - 参数设置页面

- [x] 15.1 创建前端 API 层 `src/api/system/config/`
- [x] 15.2 创建参数设置页面 `src/views/system/config/index.vue`、`config-modal.vue`、`data.ts`
- [x] 15.3 添加参数设置路由到系统路由模块

## 16. 前端 - 字典导出导入优化

- [x] 16.1 前端字典类型面板更新导出导入功能，使用合并接口
- [x] 16.2 前端字典数据面板移除导出和导入按钮
- [x] 16.3 抽象通用导出确认弹窗组件 `ExportConfirmModal`，复用至所有导出模块
- [x] 16.4 统一所有模块导出文件命名规范

## 17. 前端 - 通用改进

- [x] 17.1 用户管理页面状态字段改为从字典模块动态读取
- [x] 17.2 部门/岗位管理页面状态选项改为从字典模块动态读取
- [x] 17.3 字典管理页面修改字典数据后清除 dictStore 缓存
- [x] 17.4 移除头像下拉菜单中的多余菜单项，修正用户邮箱和昵称显示
- [x] 17.5 个人中心表单字段调整（昵称必填、非必填字段修正）
- [x] 17.6 全局分页选项增加 100 条/页

## 18. 接口文档完善

- [x] 18.1 完善 auth 模块接口文档：为 g.Meta 添加 dc 标签，为所有输入输出字段补充 dc 和 eg 标签
- [x] 18.2 完善 user 模块接口文档（13 个文件）
- [x] 18.3 完善 dept 模块接口文档（8 个文件）
- [x] 18.4 完善 post 模块接口文档（8 个文件）
- [x] 18.5 完善 dict 模块接口文档（14 个文件）
- [x] 18.6 完善 notice 模块接口文档（5 个文件）
- [x] 18.7 完善 loginlog 模块接口文档（5 个文件）
- [x] 18.8 完善 operlog 模块接口文档（5 个文件）
- [x] 18.9 完善 usermsg 模块接口文档（6 个文件）
- [x] 18.10 完善 sysinfo 模块接口文档（1 个文件）

## 19. E2E 测试

- [x] 19.1 TC0026-TC0034：操作日志和登录日志的列表查询、详情查看、清理、导出、自动记录测试
- [x] 19.2 TC0044-TC0045：系统接口页面和版本信息页面加载测试
- [x] 19.3 TC0049-TC0052：在线用户列表、搜索、强制下线、服务监控页面测试
- [x] 19.4 参数设置页面 CRUD、搜索、导出导入测试
- [x] 19.5 字典合并导出导入测试
- [x] 19.6 导出确认弹窗测试（所有导出模块）
- [x] 19.7 字典修改后全局生效测试
- [x] 19.8 运行全部 E2E 测试确认无回归

## 20. 代码质量与重构

- [x] 20.1 user/dept/file 服务的事务管理修复，Create/Update 方法使用事务确保数据一致性
- [x] 20.2 user 和 post 服务中重复的部门树遍历逻辑抽取到 dept 服务复用
- [x] 20.3 将部门层级查询中的 MySQL 特有 FIND_IN_SET 替换为基于 parent_id 迭代查询的跨数据库通用实现
- [x] 20.4 用户列表查询 N+1 问题修复，批量查询部门信息
- [x] 20.5 字典类型更新时校验 Type 字段唯一性
- [x] 20.6 日志导出方法添加条数限制防止内存溢出
- [x] 20.7 文件上传添加文件名清洗防止路径遍历攻击
- [x] 20.8 通知公告 NoticeModal formRules 改为 reactive 对象修复 Vue warn
- [x] 20.9 暗黑模式下面包屑链接颜色修正
- [x] 20.10 容器环境兼容优化：过滤虚拟文件系统挂载点，采集失败时优雅降级
