## ADDED Requirements

### Requirement: 参数设置列表查询
系统 SHALL 提供系统参数的分页列表查询接口 `GET /api/v1/config`，支持按参数名称（模糊匹配）、参数键名（模糊匹配）和创建时间范围筛选。列表默认按 ID 倒序排列。

#### Scenario: 无筛选条件查询
- **WHEN** 用户请求 `GET /api/v1/config?pageNum=1&pageSize=10`
- **THEN** 系统返回前 10 条参数记录，按 ID 倒序排列，附带总数

#### Scenario: 按参数名称筛选
- **WHEN** 用户请求 `GET /api/v1/config?name=邮件`
- **THEN** 系统仅返回名称包含"邮件"的参数记录

#### Scenario: 按参数键名筛选
- **WHEN** 用户请求 `GET /api/v1/config?key=smtp`
- **THEN** 系统仅返回键名包含"smtp"的参数记录

#### Scenario: 按时间范围筛选
- **WHEN** 用户请求 `GET /api/v1/config?beginTime=2026-01-01&endTime=2026-12-31`
- **THEN** 系统仅返回在指定时间范围内创建的参数记录

### Requirement: 获取参数详情
系统 SHALL 提供按 ID 获取单条参数记录的接口 `GET /api/v1/config/{id}`。

#### Scenario: 获取已存在的参数详情
- **WHEN** 用户请求 `GET /api/v1/config/1`（ID 存在）
- **THEN** 系统返回完整的参数记录，包含 id、name、key、value、remark、created_at、updated_at

#### Scenario: 获取不存在的参数详情
- **WHEN** 用户请求 `GET /api/v1/config/99999`（ID 不存在）
- **THEN** 系统返回错误信息，提示参数不存在

### Requirement: 新增参数
系统 SHALL 支持通过 `POST /api/v1/config` 创建新的系统参数。参数键名 MUST 在所有未删除的记录中唯一。

#### Scenario: 创建参数成功
- **WHEN** 用户请求 `POST /api/v1/config`，body 为 `{name: "邮件服务地址", key: "sys.mail.host", value: "smtp.example.com"}`
- **THEN** 系统创建参数记录并返回新记录 ID

#### Scenario: 键名重复
- **WHEN** 用户请求 `POST /api/v1/config`，但 key 已被其他记录使用
- **THEN** 系统返回错误信息，提示键名已存在

### Requirement: 修改参数
系统 SHALL 支持通过 `PUT /api/v1/config/{id}` 修改已有参数记录的 name、key、value、remark 字段。修改后的键名 MUST 保持唯一。

#### Scenario: 修改参数成功
- **WHEN** 用户请求 `PUT /api/v1/config/1`，body 为 `{value: "new-value"}`
- **THEN** 系统更新记录并设置 updated_at 为当前时间

#### Scenario: 修改时键名与其他记录冲突
- **WHEN** 用户请求 `PUT /api/v1/config/1`，但 key 与另一条记录的 key 相同
- **THEN** 系统返回错误信息，提示键名已存在

### Requirement: 删除参数
系统 SHALL 支持通过 `DELETE /api/v1/config/{id}` 软删除单条参数记录，同时支持批量删除。

#### Scenario: 删除单条参数
- **WHEN** 用户请求 `DELETE /api/v1/config/1`（ID 存在）
- **THEN** 系统软删除该记录（设置 deleted_at 时间戳）

#### Scenario: 删除不存在的参数
- **WHEN** 用户请求 `DELETE /api/v1/config/99999`（ID 不存在）
- **THEN** 系统返回错误信息，提示参数不存在

### Requirement: 按键名查询参数
系统 SHALL 提供按键名查询接口 `GET /api/v1/config/key/{key}`，供其他模块内部调用获取参数值。

#### Scenario: 查询已存在的键名
- **WHEN** 用户请求 `GET /api/v1/config/key/sys.mail.host`
- **THEN** 系统返回匹配该键名的参数记录

#### Scenario: 查询不存在的键名
- **WHEN** 用户请求 `GET /api/v1/config/key/not.exist.key`
- **THEN** 系统返回错误信息，提示参数键名不存在

### Requirement: 参数导出
系统 SHALL 支持通过 `GET /api/v1/config/export` 将参数记录导出为 Excel 文件。导出应用与列表查询相同的筛选条件。

#### Scenario: 导出全部参数
- **WHEN** 用户请求 `GET /api/v1/export`（不带筛选条件）
- **THEN** 系统生成并返回包含所有未删除参数记录的 Excel 文件

#### Scenario: 按筛选条件导出
- **WHEN** 用户请求 `GET /api/v1/config/export?name=邮件`
- **THEN** 系统生成仅包含名称匹配的参数记录的 Excel 文件

### Requirement: 参数导入
系统 SHALL 支持通过 `POST /api/v1/config/import` 从 Excel 文件导入参数记录。系统 SHALL 提供模板下载端点，并在持久化前校验导入数据。

#### Scenario: 下载导入模板
- **WHEN** 用户请求 `GET /api/v1/config/import-template`
- **THEN** 系统返回包含示例数据的 Excel 模板，列包含：参数名称、参数键名、参数键值、备注

#### Scenario: 导入有效数据
- **WHEN** 用户上传有效的 Excel 文件到 `POST /api/v1/config/import`
- **THEN** 系统校验所有行，创建记录，返回成功数量

#### Scenario: 导入数据校验失败
- **WHEN** 用户上传包含无效数据的 Excel 文件（缺少必填字段、键名重复）
- **THEN** 系统拒绝整个导入，返回包含行号和原因的错误详情

#### Scenario: 覆盖模式导入
- **WHEN** 用户上传 Excel 文件时设置 `updateSupport=true`，文件中包含已存在的键名
- **THEN** 系统使用导入的值更新已有记录

#### Scenario: 忽略模式导入
- **WHEN** 用户上传 Excel 文件时设置 `updateSupport=false`（默认），文件中包含已存在的键名
- **THEN** 系统跳过已存在的记录，仅创建新记录

#### Scenario: 导入弹窗 UI
- **WHEN** 用户点击参数设置页面的"导入"按钮
- **THEN** 系统展示弹窗，包含模板下载链接、拖拽上传区域、文件类型提示（xlsx/xls）、覆盖/忽略模式开关

### Requirement: 参数设置前端页面
前端 SHALL 在系统管理菜单下提供参数设置管理页面，包含搜索栏、工具栏、数据表格和新增/编辑弹窗。

#### Scenario: 展示参数列表页面
- **WHEN** 用户导航到参数设置页面
- **THEN** 页面展示搜索栏（参数名称、参数键名、创建时间范围）、工具栏（导出、批量删除、新增）、VXE-Grid 表格，列包含：复选框、参数名称、参数键名、参数键值、备注、修改时间、操作（编辑/删除）

#### Scenario: 通过弹窗新增参数
- **WHEN** 用户点击"新增"按钮并填写表单（参数名称、参数键名、参数键值、备注）
- **THEN** 系统创建参数并刷新表格

#### Scenario: 通过弹窗编辑参数
- **WHEN** 用户点击某行的"编辑"按钮
- **THEN** 系统打开预填充的弹窗，用户编辑后保存，表格刷新

#### Scenario: 通过确认框删除参数
- **WHEN** 用户点击某行的"删除"按钮并确认
- **THEN** 系统删除参数并刷新表格

#### Scenario: 批量删除参数
- **WHEN** 用户勾选多行后点击"批量删除"并确认
- **THEN** 系统删除所有选中的参数并刷新表格

### Requirement: 参数设置菜单和权限
系统 SHALL 在系统管理菜单下包含"参数设置"菜单项。参数操作 SHALL 受权限控制。

#### Scenario: 菜单可见性
- **WHEN** 用户拥有 `system:config:list` 权限
- **THEN** 系统管理菜单中可见"参数设置"菜单项

#### Scenario: 权限控制操作
- **WHEN** 用户缺少 `system:config:add` 权限
- **THEN** 参数设置页面的"新增"按钮隐藏
