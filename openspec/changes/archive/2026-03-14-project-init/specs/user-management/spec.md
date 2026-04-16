## ADDED Requirements

### Requirement: 用户列表查询
系统 SHALL 提供用户列表分页查询接口。

#### Scenario: 查询用户列表
- **WHEN** 调用 `GET /api/v1/user` 并传入分页参数 `pageNum` 和 `pageSize`
- **THEN** 返回用户列表和总数，格式为 `{list: [...], total: number}`

#### Scenario: 用户列表支持条件筛选
- **WHEN** 查询时传入筛选参数（如 `username`、`status`、`phone`）
- **THEN** 返回符合条件的用户列表

#### Scenario: 用户列表排除已删除用户
- **WHEN** 查询用户列表
- **THEN** 结果中不包含已软删除的用户

#### Scenario: 用户列表不返回密码
- **WHEN** 查询用户列表或用户详情
- **THEN** 返回数据中不包含密码字段

### Requirement: 创建用户
系统 SHALL 提供创建用户接口。

#### Scenario: 创建用户成功
- **WHEN** 调用 `POST /api/v1/user` 并提交用户名、密码、昵称等信息
- **THEN** 系统创建用户并返回用户 ID

#### Scenario: 用户名重复
- **WHEN** 创建用户时提交已存在的用户名
- **THEN** 系统返回错误信息，提示用户名已存在

#### Scenario: 必填字段校验
- **WHEN** 创建用户时缺少用户名或密码
- **THEN** 系统返回参数校验错误

### Requirement: 更新用户信息
系统 SHALL 提供更新用户信息接口。

#### Scenario: 更新用户成功
- **WHEN** 调用 `PUT /api/v1/user/{id}` 并提交要更新的字段
- **THEN** 系统更新对应用户信息并返回成功

#### Scenario: 更新不存在的用户
- **WHEN** 更新一个不存在的用户 ID
- **THEN** 系统返回错误信息，提示用户不存在

### Requirement: 删除用户
系统 SHALL 提供软删除用户接口。

#### Scenario: 删除用户成功
- **WHEN** 调用 `DELETE /api/v1/user/{id}`
- **THEN** 用户被软删除（设置 deleted_at），不做物理删除

#### Scenario: 不能删除自己
- **WHEN** 当前登录用户尝试删除自己
- **THEN** 系统返回错误信息，提示不能删除当前登录用户

#### Scenario: 不能删除默认管理员
- **WHEN** 尝试删除 ID 为 1 的默认管理员
- **THEN** 系统返回错误信息，提示不能删除默认管理员

### Requirement: 修改用户状态
系统 SHALL 提供独立的用户状态修改接口。

#### Scenario: 启用用户
- **WHEN** 调用 `PUT /api/v1/user/{id}/status` 并设置 status 为 1
- **THEN** 用户状态变为正常

#### Scenario: 停用用户
- **WHEN** 调用 `PUT /api/v1/user/{id}/status` 并设置 status 为 0
- **THEN** 用户状态变为停用

#### Scenario: 不能停用自己
- **WHEN** 当前登录用户尝试停用自己
- **THEN** 系统返回错误信息

### Requirement: 查看用户详情
系统 SHALL 提供用户详情查询接口。

#### Scenario: 查询用户详情
- **WHEN** 调用 `GET /api/v1/user/{id}`
- **THEN** 返回该用户的完整信息（不含密码）

### Requirement: 当前用户个人信息
系统 SHALL 提供当前登录用户查看和修改自身信息的接口。

#### Scenario: 查看个人信息
- **WHEN** 调用 `GET /api/v1/user/profile`
- **THEN** 返回当前登录用户的个人信息

#### Scenario: 修改个人信息
- **WHEN** 调用 `PUT /api/v1/user/profile` 并提交要修改的字段（昵称、邮箱、手机）
- **THEN** 系统更新当前用户信息并返回成功

#### Scenario: 修改个人密码
- **WHEN** 调用 `PUT /api/v1/user/profile` 并提交新密码
- **THEN** 系统使用 bcrypt 哈希新密码并更新

### Requirement: 用户数据表设计
用户表（sys_user）SHALL 为轻量设计，仅包含基础用户信息。

#### Scenario: 用户表字段
- **WHEN** 查看 sys_user 表结构
- **THEN** 表包含：id、username、password、nickname、email、phone、status、remark、created_at、updated_at、deleted_at

#### Scenario: 用户名唯一约束
- **WHEN** 尝试插入重复的用户名
- **THEN** 数据库拒绝插入并返回唯一约束错误

## MODIFIED Requirements

### Requirement: 用户列表查询
系统 SHALL 提供用户列表分页查询接口，支持多字段排序和增强的条件筛选。

#### Scenario: 用户列表支持字段排序
- **WHEN** 调用 `GET /api/v1/user` 并传入排序参数 `orderBy`（字段名）和 `orderDirection`（`asc` 或 `desc`）
- **THEN** 返回按指定字段和方向排序的用户列表
- **THEN** 支持排序的字段包括：`id`、`username`、`nickname`、`phone`、`email`、`status`、`created_at`

#### Scenario: 默认排序
- **WHEN** 调用 `GET /api/v1/user` 未传入排序参数
- **THEN** 默认按 `id` 降序排列

#### Scenario: 用户列表支持增强条件筛选
- **WHEN** 查询时传入筛选参数（`username`、`nickname`、`status`、`phone`、`beginTime`、`endTime`）
- **THEN** `username` 和 `nickname` 使用模糊匹配（LIKE）
- **THEN** `phone` 使用模糊匹配（LIKE）
- **THEN** `status` 使用精确匹配
- **THEN** `beginTime` 和 `endTime` 筛选 `created_at` 在该时间范围内的用户

## ADDED Requirements

### Requirement: 用户列表导出
系统 SHALL 提供将用户列表导出为 Excel 文件的功能。

#### Scenario: 导出全部用户
- **WHEN** 在用户管理页面点击"导出"按钮
- **THEN** 系统根据当前搜索条件导出符合条件的用户数据为 `.xlsx` 文件
- **THEN** 导出字段包括：用户名、昵称、手机号、邮箱、状态、备注、创建时间
- **THEN** 不导出密码字段

#### Scenario: 导出 API
- **WHEN** 调用 `GET /api/v1/user/export` 并传入与列表查询相同的筛选参数
- **THEN** 返回 Excel 文件流（Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet）

### Requirement: 用户列表导入
系统 SHALL 提供从 Excel 文件批量导入用户的功能。

#### Scenario: 导入用户
- **WHEN** 在用户管理页面点击"导入"按钮并上传 `.xlsx` 文件
- **THEN** 系统解析文件中的用户数据并批量创建用户
- **THEN** 导入字段包括：用户名（必填）、密码（必填）、昵称、手机号、邮箱、状态、备注
- **THEN** 用户名重复的记录跳过并记录到失败列表中

#### Scenario: 导入结果反馈
- **WHEN** 导入完成
- **THEN** 返回导入结果，包括成功条数和失败条数及失败原因

#### Scenario: 导入 API
- **WHEN** 调用 `POST /api/v1/user/import` 并上传 Excel 文件
- **THEN** 解析并批量创建用户，返回导入结果 `{success: number, fail: number, failList: [{row, reason}]}`

#### Scenario: 下载导入模板
- **WHEN** 调用 `GET /api/v1/user/import-template`
- **THEN** 返回标准导入模板 Excel 文件，包含表头和示例数据

### Requirement: 测试数据
系统 SHALL 提供足够的测试用户数据以验证分页、排序和筛选功能。

#### Scenario: 初始化测试数据
- **WHEN** 运行数据初始化脚本
- **THEN** 系统创建 100 条测试用户数据
- **THEN** 测试数据覆盖各种状态（正常/停用）、不同的用户名、昵称、手机号、邮箱
