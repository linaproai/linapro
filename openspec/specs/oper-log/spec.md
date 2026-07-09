# 操作日志

## Purpose

定义 `linapro-monitor-operlog` 源码插件承担的操作日志自动记录、查询、删除和导出行为，确保系统的关键写操作和指定读操作具有可追溯和可审计的操作痕迹。
## Requirements
### Requirement:操作日志自动记录
系统 SHALL 通过 `linapro-monitor-operlog` 源码插件使用宿主统一 HTTP 注册入口声明的全局审计中间件，为所有写操作（POST/PUT/DELETE）和标记了 `operLog` 标签的查询操作自动发出统一审计事件。宿主仅提供托管的全局中间件注册接缝和统一事件分发，不保留固定的操作日志业务中间件。当 `linapro-monitor-operlog` 已安装并启用时，插件中间件参与请求链路并将日志持久化到 `plugin_linapro_monitor_operlog` 表；当插件不可用时，宿主核心请求链路必须绕过采集逻辑并继续正常执行。

#### Scenario:操作日志插件已启用
- **当** 用户发起已审计的请求且 `linapro-monitor-operlog` 已安装并启用时
- **则** `linapro-monitor-operlog` 通过宿主的全局 HTTP 中间件注册器包装匹配的请求
- **且** 宿主发出统一审计事件
- **且** `linapro-monitor-operlog` 写入对应的操作日志记录

#### Scenario:操作日志插件缺失或禁用
- **当** 用户发起已审计的请求但 `linapro-monitor-operlog` 未安装、未启用或初始化失败时
- **则** 宿主绕过插件自注册的审计中间件逻辑
- **且** 宿主仍正常完成原始业务请求
- **且** 宿主不因缺少特定操作日志实现而返回错误

#### Scenario:下游中间件提前结束请求
- **当** `linapro-monitor-operlog` 的全局审计中间件已包装请求，而后续中间件或处理器在写入响应后提前结束当前请求时
- **则** 审计中间件仍可在 `Next` 返回后读取当前响应快照并发出匹配的审计事件
- **且** 提前结束请求不会导致操作日志被遗漏

### Requirement:操作日志记录内容
系统 SHALL 记录以下操作信息：模块名称（title、`g.Meta` 的 `tags` 标签）、操作名称（oper_summary、`g.Meta` 的 `summary` 标签）、操作类型（oper_type）、请求方法（request_method）、请求 URL（oper_url）、操作人员用户名（oper_name）、操作 IP（oper_ip）、请求参数（oper_param）、响应结果（json_result）、操作状态（status）、错误信息（error_msg）、耗时（cost_time）、操作时间（oper_time）。

#### Scenario:记录完整操作信息
- **当** 一次写操作完成（成功或不成功）时
- **则** 日志记录包含上述所有字段，`status` 为 0（成功）或 1（失败），`cost_time` 为请求处理时间（毫秒）

#### Scenario:请求参数长度截断
- **当** 请求参数 JSON 长度超过 2000 字符时
- **则** 截断到 2000 字符并追加 `... (truncated)`

#### Scenario:响应结果长度截断
- **当** 响应结果 JSON 长度超过 2000 字符时
- **则** 截断到 2000 字符并追加 `... (truncated)`

#### Scenario:密码字段脱敏
- **当** 请求参数包含 `password` 或 `Password` 字段时
- **则** 将字段值替换为 `* * *`

#### Scenario:操作类型使用语义化字符串常量
- **当** 系统记录、查询或导出操作日志时
- **则** `oper_type` 使用 `create`、`update`、`delete`、`export`、`import`、`other` 等语义化字符串
- **且** 宿主和插件代码通过强类型常量复用这些值，而非散落的硬编码或 `1~6` 整数

### Requirement:操作日志列表查询
系统 SHALL 提供操作日志分页查询接口 `GET /api/v1/operlog`，支持按操作模块、操作人员、操作类型、状态和时间范围筛选。

#### Scenario:分页查询操作日志
- **当** 管理员请求 `GET /api/v1/operlog?pageNum=1&pageSize=10` 时
- **则** 返回按操作时间倒序排列的操作日志分页列表

#### Scenario:按条件筛选
- **当** 管理员请求带筛选条件的查询（如 `title=用户&operName=admin&operType=create&status=0&beginTime=2026-01-01&endTime=2026-03-15`）时
- **则** 返回满足所有条件的日志记录

### Requirement:查看操作日志详情
系统 SHALL 提供操作日志详情接口 `GET /api/v1/operlog/{id}`，返回完整的日志信息。

#### Scenario:查看详情
- **当** 管理员请求 `GET /api/v1/operlog/1` 时
- **则** 返回该日志的所有字段信息，包括完整的请求参数和响应结果

### Requirement:按时间范围清理操作日志
系统 SHALL 提供操作日志清理接口 `DELETE /api/v1/operlog/clean`，支持按时间范围硬删除日志记录。

#### Scenario:按时间范围清理日志
- **当** 管理员请求 `DELETE /api/v1/operlog/clean?beginTime=2026-01-01&endTime=2026-01-31` 时
- **则** 硬删除时间范围内的所有操作日志记录，返回删除记录数

#### Scenario:清理所有日志
- **当** 管理员请求 `DELETE /api/v1/operlog/clean`（不带时间参数）时
- **则** 硬删除所有操作日志记录

### Requirement:批量删除操作日志
系统 SHALL 保留操作日志按 ID 列表删除接口 `DELETE /api/v1/operlog/{ids}`，用于受控 API 调用或后续专门入口；操作日志管理页面 SHALL 不再提供表格勾选批量删除交互。

#### Scenario:按 ID 批量删除
- **当** 管理员请求 `DELETE /api/v1/operlog/1,2,3` 时
- **则** 硬删除指定 ID 的操作日志记录，返回删除记录数

#### Scenario:前端不显示勾选批量删除
- **当** 管理员访问操作日志页面时
- **则** 表格不显示行复选框或表头全选框
- **且** 工具栏删除按钮不依赖选中行数量启用或禁用
- **且** 页面不会通过选中 ID 列表执行删除

### Requirement:操作日志导出
系统 SHALL 提供操作日志导出接口 `GET /api/v1/operlog/export`，按当前筛选条件导出为 xlsx 格式。

#### Scenario:按筛选条件导出
- **当** 管理员请求 `GET /api/v1/operlog/export?title=用户&status=0` 时
- **则** 返回包含所有符合条件日志记录的 xlsx 文件，包含所有字段

#### Scenario:导出全部
- **当** 管理员请求 `GET /api/v1/operlog/export`（不带筛选条件）时
- **则** 返回包含所有操作日志的 xlsx 文件

### Requirement:操作日志前端页面
系统 SHALL 通过 `linapro-monitor-operlog` 源码插件在前端系统监控菜单下提供操作日志管理页面。

#### Scenario:操作日志列表页
- **当** 管理员访问操作日志页面时
- **则** 显示操作日志表格，包括筛选区域（模块名称、操作人员、操作类型、状态、时间范围）、表格列（模块名称、操作名称、操作人员、操作 IP、操作状态、操作日期、操作时间）、工具栏（清空、导出、删除按钮）、行操作（查看详情按钮）
- **且** 表格不显示多选框
- **且** 删除按钮始终作为范围删除入口展示，不因未勾选记录而置灰

#### Scenario:查看详情抽屉
- **当** 管理员点击某条日志的"详情"按钮时
- **则** 打开右侧抽屉显示该日志的完整信息，包括格式化的请求参数和响应结果 JSON

#### Scenario:清空操作
- **当** 管理员点击"清空"按钮并确认时
- **则** 执行不带时间范围参数的清理请求，清空当前可见权限范围内的所有操作日志

#### Scenario:范围删除操作
- **当** 管理员点击"删除"按钮时
- **则** 弹出日期范围选择对话框
- **且** 日期选择组件与上方提示区域保持清晰间隔
- **当** 管理员选择开始日期和结束日期并确认时
- **则** 前端请求 `DELETE /api/v1/operlog/clean?beginTime=<开始日期>&endTime=<结束日期>`
- **且** 后端硬删除该时间范围内的操作日志记录并返回删除记录数

#### Scenario:删除所有操作日志
- **当** 管理员点击"删除"按钮并在对话框中选择删除所有日志时
- **则** 日期范围不再作为必填条件
- **当** 管理员确认删除时
- **则** 前端请求 `DELETE /api/v1/operlog/clean`
- **且** 后端硬删除当前权限范围内的所有操作日志记录并返回删除记录数

#### Scenario:范围为空时阻止删除
- **当** 管理员打开删除对话框但未选择完整日期范围并确认时
- **则** 前端阻止提交并提示需要选择日志范围
- **且** 不发起清理请求

#### Scenario:导出操作
- **当** 管理员点击"导出"按钮时
- **则** 按当前筛选条件将操作日志导出为 xlsx 文件并下载

### Requirement:操作日志类型使用语义化字符串常量

系统 SHALL 使用具有业务语义的字符串常量表达操作日志类型，而非在宿主、插件、接口和存储层传播位置敏感的 `1~6` 整数编码。

#### Scenario:审计事件入库时写入语义化类型
- **当** 宿主发出操作日志审计事件时
- **则** `linapro-monitor-operlog` 使用强类型常量写入 `oper_type`
- **且** `oper_type` 的持久化值为 `create`、`update`、`delete`、`export`、`import`、`other` 之一

#### Scenario:操作日志接口返回语义化类型
- **当** 管理员查询或导出操作日志时
- **则** 接口中的 `operType` 字段返回语义化字符串值
- **且** 前端继续通过 `sys_oper_type` 字典渲染对应的本地化标签

### Requirement:操作日志管理接口由源码插件交付

系统 SHALL 将操作日志查询、详情、导出、清理和页面能力作为 `linapro-monitor-operlog` 源码插件交付。

#### Scenario:插件启用时暴露管理入口
- **当** `linapro-monitor-operlog` 已安装并启用时
- **则** 宿主暴露操作日志查询、详情、导出、清理接口和前端页面
- **且** 插件菜单挂载到宿主 `系统监控` 目录，顶层 `parent_key` 为 `monitor`

#### Scenario:插件缺失时隐藏管理入口
- **当** `linapro-monitor-operlog` 未安装或未启用时
- **则** 宿主不显示操作日志菜单和页面入口
- **且** 普通业务请求链路继续正常运行

