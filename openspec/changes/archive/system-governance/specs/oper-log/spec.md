## ADDED Requirements

### Requirement: 操作日志自动记录
系统 SHALL 通过中间件自动记录所有写操作（POST/PUT/DELETE）以及标记了 `operLog` 标签的查询操作到 `sys_oper_log` 表。普通 GET 查询请求不记录。

#### Scenario: POST 请求自动记录为新增操作
- **WHEN** 用户发起 `POST` 请求（如创建用户 `POST /api/v1/user`）
- **THEN** 系统自动写入一条操作日志，`oper_type` 为 1（新增），`title` 从 `g.Meta` 的 `tags` 字段获取，`oper_name` 为当前登录用户名

#### Scenario: PUT 请求自动记录为修改操作
- **WHEN** 用户发起 `PUT` 请求（如修改用户 `PUT /api/v1/user/1`）
- **THEN** 系统自动写入一条操作日志，`oper_type` 为 2（修改）

#### Scenario: DELETE 请求自动记录为删除操作
- **WHEN** 用户发起 `DELETE` 请求（如删除用户 `DELETE /api/v1/user/1`）
- **THEN** 系统自动写入一条操作日志，`oper_type` 为 3（删除）

#### Scenario: 标记的 GET 请求记录为导出操作
- **WHEN** 用户发起带 `operLog` 标签的 GET 请求（如导出用户 `GET /api/v1/user/export`）
- **THEN** 系统自动写入一条操作日志，`oper_type` 从 `operLog` 标签值获取（如 4=导出）

#### Scenario: 普通 GET 请求不记录
- **WHEN** 用户发起未标记 `operLog` 的 GET 请求（如查询用户列表 `GET /api/v1/user`）
- **THEN** 系统不写入操作日志

#### Scenario: 导入请求记录为导入操作
- **WHEN** 用户发起 POST 请求且路径包含 `import`（如 `POST /api/v1/user/import`）
- **THEN** 系统自动写入一条操作日志，`oper_type` 为 5（导入）

### Requirement: 操作日志记录内容
系统 SHALL 记录以下操作信息：模块名称（title，来自 `g.Meta` 的 `tags` 标签）、操作名称（oper_summary，来自 `g.Meta` 的 `summary` 标签）、操作类型（oper_type）、请求方法（request_method）、请求路由（method）、请求URL（oper_url）、操作人用户名（oper_name）、操作IP（oper_ip）、请求参数（oper_param）、响应结果（json_result）、操作状态（status）、错误信息（error_msg）、耗时（cost_time）、操作时间（oper_time）。

#### Scenario: 记录完整操作信息
- **WHEN** 一次写操作完成（无论成功或失败）
- **THEN** 日志记录包含上述所有字段，`status` 为 0（成功）或 1（失败），`cost_time` 为请求处理耗时（毫秒）

#### Scenario: 请求参数长度截断
- **WHEN** 请求参数 JSON 长度超过 2000 字符
- **THEN** 截断至 2000 字符并追加 `...(truncated)`

#### Scenario: 响应结果长度截断
- **WHEN** 响应结果 JSON 长度超过 2000 字符
- **THEN** 截断至 2000 字符并追加 `...(truncated)`

#### Scenario: 密码字段脱敏
- **WHEN** 请求参数中包含 `password` 或 `Password` 字段
- **THEN** 该字段值替换为 `***`

### Requirement: 操作日志列表查询
系统 SHALL 提供操作日志分页查询接口 `GET /api/v1/operlog`，支持按操作模块、操作人、操作类型、状态、时间范围筛选。

#### Scenario: 分页查询操作日志
- **WHEN** 管理员请求 `GET /api/v1/operlog?pageNum=1&pageSize=10`
- **THEN** 返回操作日志分页列表，按操作时间倒序排列

#### Scenario: 按条件筛选
- **WHEN** 管理员请求带筛选条件的查询（如 `title=User&operName=admin&operType=1&status=0&beginTime=2026-01-01&endTime=2026-03-15`）
- **THEN** 返回符合所有条件的日志记录

### Requirement: 操作日志详情查看
系统 SHALL 提供操作日志详情接口 `GET /api/v1/operlog/{id}`，返回完整的日志信息。

#### Scenario: 查看日志详情
- **WHEN** 管理员请求 `GET /api/v1/operlog/1`
- **THEN** 返回该日志的所有字段信息，包括完整的请求参数和响应结果

### Requirement: 操作日志按时间范围清理
系统 SHALL 提供操作日志清理接口 `DELETE /api/v1/operlog/clean`，支持按时间范围硬删除日志记录。

#### Scenario: 按时间范围清理日志
- **WHEN** 管理员请求 `DELETE /api/v1/operlog/clean?beginTime=2026-01-01&endTime=2026-01-31`
- **THEN** 硬删除该时间范围内的所有操作日志记录，返回删除的记录数

#### Scenario: 清理全部日志
- **WHEN** 管理员请求 `DELETE /api/v1/operlog/clean`（不带时间参数）
- **THEN** 硬删除所有操作日志记录

### Requirement: 操作日志批量删除
系统 SHALL 提供操作日志批量删除接口 `DELETE /api/v1/operlog/{ids}`，支持按 ID 列表硬删除日志记录。

#### Scenario: 按 ID 批量删除
- **WHEN** 管理员请求 `DELETE /api/v1/operlog/1,2,3`
- **THEN** 硬删除指定 ID 的操作日志记录，返回删除的记录数

#### Scenario: 前端批量删除操作
- **WHEN** 管理员勾选一条或多条操作日志后点击"删除"按钮
- **THEN** 弹出确认对话框显示选中数量，确认后执行批量删除

### Requirement: 操作日志导出
系统 SHALL 提供操作日志导出接口 `GET /api/v1/operlog/export`，按当前筛选条件导出为 xlsx 格式文件。

#### Scenario: 按筛选条件导出
- **WHEN** 管理员请求 `GET /api/v1/operlog/export?title=User&status=0`
- **THEN** 返回包含符合条件的所有日志记录的 xlsx 文件，包含所有字段

#### Scenario: 导出全部
- **WHEN** 管理员请求 `GET /api/v1/operlog/export`（不带筛选条件）
- **THEN** 返回包含所有操作日志的 xlsx 文件

### Requirement: 操作日志前端页面
系统 SHALL 在前端系统监控菜单下提供操作日志管理页面。

#### Scenario: 操作日志列表页
- **WHEN** 管理员访问操作日志页面
- **THEN** 展示操作日志表格，包含筛选区域（模块名称、操作人员、操作类型、状态、时间范围），表格列（模块名称、操作名称、操作人员、操作IP、操作状态、操作日期、操作耗时），工具栏（清空、导出、删除按钮），行操作（查看详情按钮）。删除按钮在未勾选任何记录时置灰不可点击。

#### Scenario: 查看详情抽屉
- **WHEN** 管理员点击某条日志的"详情"按钮
- **THEN** 打开右侧抽屉（Drawer），展示日志的完整信息，包括格式化的请求参数和响应结果 JSON

#### Scenario: 清理操作
- **WHEN** 管理员点击"清理"按钮
- **THEN** 弹出确认对话框，包含时间范围选择器，确认后执行清理

#### Scenario: 导出操作
- **WHEN** 管理员点击"导出"按钮
- **THEN** 按当前筛选条件导出操作日志为 xlsx 文件并下载
