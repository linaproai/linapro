# 登录日志

## Purpose

定义 `linapro-monitor-loginlog` 源码插件承担的登录日志自动记录、查询、清理和导出行为，确保系统能够跟踪、审计和分析认证成功与失败。
## Requirements
### Requirement:登录日志自动记录
系统 SHALL 在登录成功、登录失败和成功退出等认证生命周期节点自动发出统一登录事件。当 `linapro-monitor-loginlog` 已安装并启用时，插件订阅事件并将日志持久化到 `plugin_linapro_monitor_loginlog` 表；当插件不可用时，宿主认证链路仍正常执行。

#### Scenario:登录日志插件已启用
- **当** 用户成功登录、登录失败或成功退出，且 `linapro-monitor-loginlog` 已安装并启用时
- **则** 宿主发出统一登录事件
- **且** `linapro-monitor-loginlog` 订阅事件后写入对应的登录日志记录

#### Scenario:登录日志插件缺失或禁用
- **当** 用户成功登录、登录失败或成功退出，但 `linapro-monitor-loginlog` 未安装、未启用或初始化失败时
- **则** 宿主仍正常返回认证结果
- **且** 宿主不因缺少特定登录日志实现而返回错误

### Requirement:登录日志记录内容
系统 SHALL 记录以下登录信息：用户名（user_name）、登录状态（status）、登录 IP（ip）、浏览器（browser）、操作系统（os）、登录消息（msg）、登录时间（login_time）。

#### Scenario:解析浏览器和操作系统
- **当** 记录日志时
- **则** 从 HTTP 请求头 `User-Agent` 字段解析浏览器名称和操作系统名称

### Requirement:登录日志列表查询
系统 SHALL 提供登录日志分页查询接口 `GET /api/v1/loginlog`，支持按用户名、IP、状态和时间范围筛选。

#### Scenario:分页查询登录日志
- **当** 管理员请求 `GET /api/v1/loginlog?pageNum=1&pageSize=10` 时
- **则** 返回按登录时间倒序排列的登录日志分页列表

#### Scenario:按条件筛选
- **当** 管理员请求带筛选条件的查询（如 `userName=admin&status=0&beginTime=2026-01-01&endTime=2026-03-15`）时
- **则** 返回满足所有条件的登录日志记录

### Requirement:登录日志详情查看
系统 SHALL 提供登录日志详情接口 `GET /api/v1/loginlog/{id}`，返回完整的登录日志信息。

#### Scenario:查看详情
- **当** 管理员请求 `GET /api/v1/loginlog/1` 时
- **则** 返回该登录日志的所有字段信息

### Requirement:按时间范围清理登录日志
系统 SHALL 提供登录日志清理接口 `DELETE /api/v1/loginlog/clean`，支持按时间范围硬删除日志记录。

#### Scenario:按时间范围清理日志
- **当** 管理员请求 `DELETE /api/v1/loginlog/clean?beginTime=2026-01-01&endTime=2026-01-31` 时
- **则** 硬删除该时间范围内的所有登录日志记录，返回删除记录数

#### Scenario:清理所有登录日志
- **当** 管理员请求 `DELETE /api/v1/loginlog/clean`（不带时间参数）时
- **则** 硬删除所有登录日志记录

### Requirement:登录日志批量删除
系统 SHALL 保留登录日志按 ID 列表删除接口 `DELETE /api/v1/loginlog/{ids}`，用于受控 API 调用或后续专门入口；登录日志管理页面 SHALL 不再提供表格勾选批量删除交互。

#### Scenario:按 ID 批量删除
- **当** 管理员请求 `DELETE /api/v1/loginlog/1,2,3` 时
- **则** 硬删除指定 ID 的登录日志记录，返回删除记录数

#### Scenario:前端不显示勾选批量删除
- **当** 管理员访问登录日志页面时
- **则** 表格不显示行复选框或表头全选框
- **且** 工具栏删除按钮不依赖选中行数量启用或禁用
- **且** 页面不会通过选中 ID 列表执行删除

### Requirement:登录日志导出
系统 SHALL 提供登录日志导出接口 `GET /api/v1/loginlog/export`，按当前筛选条件导出为 xlsx 格式。

#### Scenario:按筛选条件导出
- **当** 管理员请求 `GET /api/v1/loginlog/export?userName=admin&status=0` 时
- **则** 返回包含所有符合条件登录日志的 xlsx 文件

### Requirement:登录日志前端页面
系统 SHALL 通过 `linapro-monitor-loginlog` 源码插件在前端系统监控菜单下提供登录日志管理页面。

#### Scenario:登录日志列表页
- **当** 管理员访问登录日志页面时
- **则** 显示登录日志表格，包括筛选区域（用户名、IP 地址、状态、时间范围）、表格列（用户名、IP 地址、浏览器、操作系统、登录结果、提示消息、登录时间）、工具栏（清空、导出、删除按钮）、行操作（查看详情按钮）
- **且** 表格不显示多选框
- **且** 删除按钮始终作为范围删除入口展示，不因未勾选记录而置灰

#### Scenario:查看详情弹窗
- **当** 管理员点击某条日志的"详情"按钮时
- **则** 打开 Modal 显示该登录日志的完整信息

#### Scenario:清空操作
- **当** 管理员点击"清空"按钮并确认时
- **则** 执行不带时间范围参数的清理请求，清空当前可见权限范围内的所有登录日志

#### Scenario:范围删除操作
- **当** 管理员点击"删除"按钮时
- **则** 弹出日期范围选择对话框
- **且** 日期选择组件与上方提示区域保持清晰间隔
- **当** 管理员选择开始日期和结束日期并确认时
- **则** 前端请求 `DELETE /api/v1/loginlog/clean?beginTime=<开始日期>&endTime=<结束日期>`
- **且** 后端硬删除该时间范围内的登录日志记录并返回删除记录数

#### Scenario:删除所有登录日志
- **当** 管理员点击"删除"按钮并在对话框中选择删除所有日志时
- **则** 日期范围不再作为必填条件
- **当** 管理员确认删除时
- **则** 前端请求 `DELETE /api/v1/loginlog/clean`
- **且** 后端硬删除当前权限范围内的所有登录日志记录并返回删除记录数

#### Scenario:范围为空时阻止删除
- **当** 管理员打开删除对话框但未选择完整日期范围并确认时
- **则** 前端阻止提交并提示需要选择日志范围
- **且** 不发起清理请求

#### Scenario:导出操作
- **当** 管理员点击"导出"按钮时
- **则** 按当前筛选条件将登录日志导出为 xlsx 文件并下载

### Requirement:登录日志管理接口由源码插件交付

系统 SHALL 将登录日志查询、详情、导出、清理和页面能力作为 `linapro-monitor-loginlog` 源码插件交付。

#### Scenario:插件启用时暴露管理入口
- **当** `linapro-monitor-loginlog` 已安装并启用时
- **则** 宿主暴露登录日志查询、详情、导出、清理接口和前端页面
- **且** 插件菜单挂载到宿主 `系统监控` 目录，顶层 `parent_key` 为 `monitor`

#### Scenario:插件缺失时隐藏管理入口
- **当** `linapro-monitor-loginlog` 未安装或未启用时
- **则** 宿主不显示登录日志菜单和页面入口
- **且** 登录和退出流程继续正常运行

