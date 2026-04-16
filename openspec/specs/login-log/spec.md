# Login Log

## Purpose

定义登录日志的自动记录、查询、清理与导出行为，确保系统能够对认证成功与失败事件进行追踪、审计和运维分析。

## Requirements

### Requirement: 登录日志自动记录
系统 SHALL 在用户登录（成功或失败）时自动记录登录日志到 `sys_login_log` 表。

#### Scenario: 登录成功记录
- **WHEN** 用户通过 `POST /api/v1/auth/login` 成功登录
- **THEN** 系统写入一条登录日志，`status` 为 0（成功），`msg` 为 "登录成功"，包含用户名、IP、浏览器、操作系统信息

#### Scenario: 登录失败记录
- **WHEN** 用户登录失败（密码错误、用户不存在、用户已停用等）
- **THEN** 系统写入一条登录日志，`status` 为 1（失败），`msg` 为具体失败原因（如"用户名或密码错误"、"用户已停用"）

### Requirement: 登录日志记录内容
系统 SHALL 记录以下登录信息：用户名（user_name）、登录状态（status）、登录IP（ip）、浏览器（browser）、操作系统（os）、登录消息（msg）、登录时间（login_time）。

#### Scenario: 解析浏览器和操作系统
- **WHEN** 记录登录日志
- **THEN** 从 HTTP 请求头 `User-Agent` 字段解析出浏览器名称和操作系统名称

### Requirement: 登录日志列表查询
系统 SHALL 提供登录日志分页查询接口 `GET /api/v1/loginlog`，支持按用户名、IP、状态、时间范围筛选。

#### Scenario: 分页查询登录日志
- **WHEN** 管理员请求 `GET /api/v1/loginlog?pageNum=1&pageSize=10`
- **THEN** 返回登录日志分页列表，按登录时间倒序排列

#### Scenario: 按条件筛选
- **WHEN** 管理员请求带筛选条件的查询（如 `userName=admin&status=0&beginTime=2026-01-01&endTime=2026-03-15`）
- **THEN** 返回符合所有条件的登录日志记录

### Requirement: 登录日志详情查看
系统 SHALL 提供登录日志详情接口 `GET /api/v1/loginlog/{id}`，返回完整的登录日志信息。

#### Scenario: 查看登录日志详情
- **WHEN** 管理员请求 `GET /api/v1/loginlog/1`
- **THEN** 返回该登录日志的所有字段信息

### Requirement: 登录日志按时间范围清理
系统 SHALL 提供登录日志清理接口 `DELETE /api/v1/loginlog/clean`，支持按时间范围硬删除日志记录。

#### Scenario: 按时间范围清理登录日志
- **WHEN** 管理员请求 `DELETE /api/v1/loginlog/clean?beginTime=2026-01-01&endTime=2026-01-31`
- **THEN** 硬删除该时间范围内的所有登录日志记录，返回删除的记录数

#### Scenario: 清理全部登录日志
- **WHEN** 管理员请求 `DELETE /api/v1/loginlog/clean`（不带时间参数）
- **THEN** 硬删除所有登录日志记录

### Requirement: 登录日志批量删除
系统 SHALL 提供登录日志批量删除接口 `DELETE /api/v1/loginlog/{ids}`，支持按 ID 列表硬删除日志记录。

#### Scenario: 按 ID 批量删除
- **WHEN** 管理员请求 `DELETE /api/v1/loginlog/1,2,3`
- **THEN** 硬删除指定 ID 的登录日志记录，返回删除的记录数

#### Scenario: 前端批量删除操作
- **WHEN** 管理员勾选一条或多条登录日志后点击"删除"按钮
- **THEN** 弹出确认对话框显示选中数量，确认后执行批量删除

### Requirement: 登录日志导出
系统 SHALL 提供登录日志导出接口 `GET /api/v1/loginlog/export`，按当前筛选条件导出为 xlsx 格式文件。

#### Scenario: 按筛选条件导出
- **WHEN** 管理员请求 `GET /api/v1/loginlog/export?userName=admin&status=0`
- **THEN** 返回包含符合条件的所有登录日志的 xlsx 文件

### Requirement: 登录日志前端页面
系统 SHALL 在前端系统监控菜单下提供登录日志管理页面。

#### Scenario: 登录日志列表页
- **WHEN** 管理员访问登录日志页面
- **THEN** 展示登录日志表格，包含筛选区域（用户名、IP地址、状态、时间范围），表格列（用户名、IP地址、浏览器、操作系统、登录结果、提示消息、登录时间），工具栏（清空、导出、删除按钮），行操作（查看详情按钮）。删除按钮在未勾选任何记录时置灰不可点击。

#### Scenario: 查看详情弹窗
- **WHEN** 管理员点击某条日志的"详情"按钮
- **THEN** 打开弹窗（Modal），展示登录日志的完整信息

#### Scenario: 清理操作
- **WHEN** 管理员点击"清理"按钮
- **THEN** 弹出确认对话框，包含时间范围选择器，确认后执行清理

#### Scenario: 导出操作
- **WHEN** 管理员点击"导出"按钮
- **THEN** 按当前筛选条件导出登录日志为 xlsx 文件并下载
