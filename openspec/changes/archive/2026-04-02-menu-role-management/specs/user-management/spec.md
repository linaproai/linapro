## MODIFIED Requirements

### Requirement: 用户列表查询

系统必须支持用户列表的分页查询，并包含角色信息。

#### Scenario: 查询用户列表
- **WHEN** 用户访问用户管理页面
- **THEN** 系统返回用户分页列表
- **AND** 每个用户包含 id、username、nickname、email、phone、sex、avatar、status、deptId、deptName、roleIds、roleNames 等信息

#### Scenario: 用户列表包含角色信息
- **WHEN** 用户请求用户列表
- **THEN** 系统通过 sys_user_role 表查询用户的角色关联
- **AND** 系统通过 sys_role 表查询角色名称
- **AND** 返回的 roleIds 为角色ID数组
- **AND** 返回的 roleNames 为角色名称数组

### Requirement: 用户详情查询

系统必须返回用户详情及其关联的角色信息。

#### Scenario: 查询用户详情
- **WHEN** 用户请求用户详情
- **THEN** 系统返回用户基本信息
- **AND** 系统返回 deptId、deptName
- **AND** 系统返回 postIds（岗位ID数组）
- **AND** 系统返回 roleIds（角色ID数组）

### Requirement: 创建用户

系统必须支持创建用户时关联角色。

#### Scenario: 创建用户并关联角色
- **WHEN** 用户提交创建用户表单
- **THEN** 系统创建用户记录
- **AND** 如果传入 deptId，系统在 sys_user_dept 表插入关联记录
- **AND** 如果传入 postIds，系统在 sys_user_post 表插入关联记录
- **AND** 如果传入 roleIds，系统在 sys_user_role 表插入关联记录

### Requirement: 更新用户

系统必须支持更新用户时修改角色关联。

#### Scenario: 更新用户角色关联
- **WHEN** 用户提交更新用户表单
- **THEN** 系统更新用户基本信息
- **AND** 如果传入 deptId，系统更新 sys_user_dept 关联
- **AND** 如果传入 postIds，系统更新 sys_user_post 关联
- **AND** 如果传入 roleIds，系统更新 sys_user_role 关联

#### Scenario: 更新用户事务处理
- **WHEN** 更新用户时发生错误
- **THEN** 系统回滚所有操作（用户基本信息、部门关联、岗位关联、角色关联）

### Requirement: 删除用户

系统必须支持删除用户时清理所有关联数据。

#### Scenario: 删除用户清理关联数据
- **WHEN** 用户删除一个用户
- **THEN** 系统软删除用户记录
- **AND** 系统删除 sys_user_dept 中的关联记录
- **AND** 系统删除 sys_user_post 中的关联记录
- **AND** 系统删除 sys_user_role 中的关联记录