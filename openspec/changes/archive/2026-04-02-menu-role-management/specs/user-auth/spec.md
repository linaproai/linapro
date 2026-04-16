## MODIFIED Requirements

### Requirement: 登录后返回用户信息

系统必须在用户登录成功后返回用户信息，包括角色和菜单树。

#### Scenario: 登录成功返回用户信息
- **WHEN** 用户使用正确的用户名和密码登录
- **THEN** 系统返回 userId、username、realName、email、avatar
- **AND** 系统返回 roles 字段，包含用户所有角色的 key 列表
- **AND** 系统返回 menus 字段，包含用户可访问的菜单树
- **AND** 系统返回 permissions 字段，包含用户所有的权限标识列表
- **AND** 系统返回 homePath 字段，指定用户的首页路径

#### Scenario: 超级管理员登录
- **WHEN** admin 角色的用户登录
- **THEN** 系统返回所有菜单（不检查 sys_role_menu 关联）
- **AND** roles 包含 "admin"
- **AND** permissions 包含 "*:*:*" 表示所有权限

#### Scenario: 普通用户登录
- **WHEN** 非超级管理员用户登录
- **THEN** 系统根据用户的角色查询 sys_role_menu 获取菜单ID列表
- **AND** 系统根据菜单ID列表构建菜单树
- **AND** 系统过滤掉停用状态（status=0）的菜单
- **AND** 系统过滤掉隐藏状态（visible=0）的菜单

#### Scenario: 用户无角色登录
- **WHEN** 没有分配任何角色的用户登录
- **THEN** 系统返回空的菜单树
- **AND** roles 为空数组
- **AND** permissions 为空数组

#### Scenario: 用户角色全部停用
- **WHEN** 用户的所有角色都被停用
- **THEN** 系统返回空的菜单树
- **AND** roles 为空数组
- **AND** permissions 为空数组

## ADDED Requirements

### Requirement: 菜单树结构

系统返回的菜单树必须符合前端路由生成要求。

#### Scenario: 菜单树包含必要字段
- **WHEN** 系统返回菜单树
- **THEN** 每个菜单节点包含 id、parentId、name、path、component、icon、type、sort、visible、status 字段
- **AND** 目录类型（type="D"）的菜单包含 children 子节点
- **AND** 菜单类型（type="M"）的菜单为叶子节点
- **AND** 按钮类型（type="B"）不在菜单树中返回

#### Scenario: 菜单树按排序字段排序
- **WHEN** 系统返回菜单树
- **THEN** 同级菜单按 sort 字段升序排列

### Requirement: 权限标识列表

系统必须返回用户的所有权限标识。

#### Scenario: 权限标识聚合
- **WHEN** 用户有多个角色
- **THEN** 系统聚合所有角色的权限标识（去重）
- **AND** 权限标识来自菜单表中 type="M" 或 type="B" 的 perms 字段

#### Scenario: 超级管理员权限
- **WHEN** 用户是超级管理员（有 admin 角色）
- **THEN** permissions 返回 ["*:*:*"]
- **AND** 前端判断此权限标识为拥有所有权限