## 1. 数据库与基础设施

- [x] 1.1 创建 `manifest/sql/v0.3.0.sql`，包含 `sys_oper_log` 和 `sys_login_log` 建表语句，以及 `sys_oper_type` 字典类型和字典数据的 Seed DML
- [x] 1.2 执行 `make dao` 生成 DAO/DO/Entity 代码
- [x] 1.3 添加 User-Agent 解析依赖（如 `mssola/useragent`）

## 2. 后端 - 登录日志模块

- [x] 2.1 创建 `api/loginlog/v1/` 接口定义：List、Get、Clean、Export
- [x] 2.2 执行 `make ctrl` 生成控制器骨架
- [x] 2.3 实现 `internal/service/loginlog/` 服务层：Create、List、Get、Clean、Export
- [x] 2.4 填写控制器方法实现
- [x] 2.5 在 `cmd_http.go` 中注册登录日志路由

## 3. 后端 - 登录日志集成到认证模块

- [x] 3.1 修改 `internal/service/auth/` 的登录方法，在登录成功/失败时调用 LoginLog Service 写入日志（包括解析 User-Agent 获取浏览器和操作系统信息）
- [x] 3.2 修改登出方法，在登出时写入登录日志

## 4. 后端 - 操作日志模块

- [x] 4.1 创建 `api/operlog/v1/` 接口定义：List、Get、Clean、Export
- [x] 4.2 执行 `make ctrl` 生成控制器骨架
- [x] 4.3 实现 `internal/service/operlog/` 服务层：Create、List、Get、Clean、Export
- [x] 4.4 填写控制器方法实现
- [x] 4.5 在 `cmd_http.go` 中注册操作日志路由

## 5. 后端 - 操作日志中间件

- [x] 5.1 实现操作日志中间件 `internal/service/middleware/operlog.go`：拦截写操作、解析 `g.Meta` 标签获取模块名和操作类型、记录请求参数（截断+脱敏）和响应结果（截断）、计算耗时、异步写入数据库
- [x] 5.2 在 `cmd_http.go` 中将操作日志中间件注册到 Auth 中间件之后
- [x] 5.3 为现有导出接口（用户导出、字典导出、岗位导出）的 `g.Meta` 添加 `operLog:"4"` 标签

## 6. 前端 - 操作日志页面

- [x] 6.1 创建操作日志 API 层：`src/api/monitor/operlog/index.ts` 和 `model.d.ts`
- [x] 6.2 创建操作日志列表页：`src/views/monitor/operlog/index.vue` 和 `data.ts`（表格+筛选）
- [x] 6.3 创建操作日志详情抽屉组件：`src/views/monitor/operlog/operlog-detail-drawer.vue`
- [x] 6.4 实现清理功能：弹窗选择时间范围后硬删除

## 7. 前端 - 登录日志页面

- [x] 7.1 创建登录日志 API 层：`src/api/monitor/loginlog/index.ts` 和 `model.d.ts`
- [x] 7.2 创建登录日志列表页：`src/views/monitor/loginlog/index.vue` 和 `data.ts`（表格+筛选）
- [x] 7.3 创建登录日志详情弹窗组件：`src/views/monitor/loginlog/loginlog-detail-modal.vue`
- [x] 7.4 实现清理功能：弹窗选择时间范围后硬删除

## 8. 前端 - 路由与菜单

- [x] 8.1 创建系统监控路由模块：`src/router/routes/modules/monitor.ts`，包含操作日志和登录日志两个子路由
- [x] 8.2 配置菜单图标和排序

## 9. E2E 测试

- [x] 9.1 创建 `hack/tests/e2e/monitor/` 目录
- [x] 9.2 TC0026：操作日志列表查询与筛选测试（验证增删改操作自动记录日志、筛选功能）
- [x] 9.3 TC0027：操作日志详情查看测试（验证详情抽屉展示完整信息）
- [x] 9.4 TC0028：操作日志清理测试（验证按时间范围清理功能）
- [x] 9.5 TC0029：操作日志导出测试（验证按筛选条件导出 xlsx）
- [x] 9.6 TC0030：登录日志自动记录测试（验证登录成功/失败/登出均记录日志）
- [x] 9.7 TC0031：登录日志列表查询与筛选测试
- [x] 9.8 TC0032：登录日志详情查看测试（验证详情弹窗展示完整信息）
- [x] 9.9 TC0033：登录日志清理测试
- [x] 9.10 TC0034：登录日志导出测试

## Feedback

- [x] **FB-1**：操作日志"系统模块"列标题改为"模块名称"，"操作类型"列改为"操作名称"并显示API的summary标签文本（需新增DB字段、中间件采集summary、前端列定义调整）
- [x] **FB-2**：操作日志和登录日志增加批量删除功能（新增DELETE API、前端增加删除按钮，仅在勾选记录后可点击，参考ruoyi-plus-vben5实现）
- [x] **FB-3**：检查并修复API输入输出为二进制文件时操作日志的记录逻辑，确保存储提示信息而非实际二进制数据
- [x] **FB-4**：修复操作日志和登录日志的导出功能报错，补充缺失的E2E测试用例
- [x] **FB-5**：操作日志和登录日志的清空功能（Clean）在未指定时间范围时报错 "there should be WHERE condition statement for DELETE operation"，需在无筛选条件时添加 Where(1) 以通过 GoFrame 安全检查
- [x] **FB-6**：TC0028/TC0033 清空测试用例仅测试了弹窗确认，未测试实际确认清空操作是否成功，需补充确认清空后验证 API 返回成功的测试
- [x] **FB-7**：操作日志列表增加日志编号列，在操作名称列右侧增加操作类型列
- [x] **FB-8**：操作日志详情抽屉调整：去掉"方法"字段、在"操作名称"下新增"操作类型"字段、"模块名称"去掉操作类型DictTag、"操作状态"改为"操作结果"、请求参数和响应结果使用vue-json-pretty实现JSON代码高亮
- [x] **FB-9**：操作日志页面的查询表单和表格列中"操作状态"字段名称修改为"操作结果"
- [x] **FB-10**：操作日志页面的"操作类型"和"操作结果"搜索字段的 Select 下拉选项未从字典数据中填充，无法选择枚举值进行检索
- [x] **FB-11**：登录日志页面的"登录状态"搜索字段的 Select 下拉选项未从字典数据中填充，无法选择枚举值进行检索
- [x] **FB-12**：审查并完善操作日志和登录日志的 E2E 测试用例，覆盖各搜索条件（模块名称、操作人员、操作类型、操作结果、操作时间、登录状态等）的检索功能验证
- [x] **FB-13**：操作日志和登录日志页面的字典 Select 下拉选项出现重复，原因是 sys_dict_data 表缺少 UNIQUE 约束导致 INSERT OR IGNORE 无效，需添加唯一索引并清理已有重复数据
- [x] **FB-14**：操作日志和登录日志的时间范围筛选不生效，原因是 RangePicker 未设置 valueFormat 导致传给后端的是 dayjs 对象而非日期字符串，需添加 valueFormat 配置
