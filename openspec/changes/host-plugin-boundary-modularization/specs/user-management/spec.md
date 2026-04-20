## MODIFIED Requirements

### Requirement: 用户列表查询
系统 SHALL 提供用户列表分页查询接口，支持多字段排序、增强的条件筛选和角色信息聚合。`org-management` 已安装并启用时，系统额外支持按部门过滤并返回部门字段；插件缺失时，宿主忽略组织扩展过滤并保持用户列表主体功能可用。

#### Scenario: 组织插件可用时按部门过滤用户列表
- **WHEN** `org-management` 已安装并启用，且查询时传入 `deptId`
- **THEN** 系统通过组织插件提供的组织关系筛选属于该部门的用户
- **AND** 返回的用户数据中可包含 `deptId` 和 `deptName` 字段

#### Scenario: 组织插件缺失时查询用户列表
- **WHEN** `org-management` 未安装或未启用，且查询用户列表
- **THEN** 系统仍返回用户分页列表和角色信息
- **AND** 与部门相关的筛选条件和字段被安全忽略或省略

### Requirement: 创建用户
系统 SHALL 提供创建用户接口，始终支持角色关联；当 `org-management` 已安装并启用时，系统额外支持关联部门和岗位；插件缺失时，这些组织扩展字段不阻塞用户创建。

#### Scenario: 组织插件缺失时创建用户
- **WHEN** `org-management` 未安装或未启用，且管理员创建用户
- **THEN** 系统仍成功创建用户并处理角色关联
- **AND** 缺少部门和岗位信息不会导致创建失败

### Requirement: 更新用户信息
系统 SHALL 提供更新用户信息接口，始终支持角色关联；当 `org-management` 已安装并启用时，系统额外支持更新部门和岗位关联；插件缺失时，这些组织扩展字段不阻塞用户更新。

#### Scenario: 组织插件缺失时更新用户
- **WHEN** `org-management` 未安装或未启用，且管理员更新用户
- **THEN** 系统仍成功更新用户基础信息与角色关联
- **AND** 与部门、岗位相关的字段被安全忽略

### Requirement: 查看用户详情
系统 SHALL 提供用户详情查询接口。`org-management` 已安装并启用时返回关联的部门和岗位信息；插件缺失时仍返回用户基础信息与角色信息。

#### Scenario: 组织插件缺失时查询用户详情
- **WHEN** `org-management` 未安装或未启用，且调用 `GET /api/v1/user/{id}`
- **THEN** 系统返回该用户的完整基础信息（不含密码）与角色信息
- **AND** `deptId`、`deptName`、`postIds` 等组织扩展字段被省略、置零值或置空集合

### Requirement: 用户部门树接口
系统 SHALL 在 `org-management` 已安装并启用时提供用于用户管理左侧筛选的部门树接口；插件缺失时，宿主不再暴露该组织扩展接口。

#### Scenario: 组织插件可用时获取用户部门树
- **WHEN** `org-management` 已安装并启用，且调用 `GET /api/v1/user/dept-tree`
- **THEN** 系统返回部门树形结构数据，每个节点包含 id、label、children、userCount
- **AND** 树的第一层仍可包含 `未分配部门` 虚拟节点

#### Scenario: 组织插件缺失时用户部门树不可用
- **WHEN** `org-management` 未安装或未启用
- **THEN** 宿主不再暴露 `GET /api/v1/user/dept-tree` 作为默认用户管理依赖接口
- **AND** 用户管理主体流程不依赖该接口才能正常工作

### Requirement: 用户管理前端部门树筛选
系统 SHALL 仅在 `org-management` 已安装并启用时在用户管理页面左侧展示 `DeptTree` 筛选区；插件缺失时页面退化为全宽用户列表。

#### Scenario: 组织插件缺失时页面布局降级
- **WHEN** `org-management` 未安装或未启用，且管理员打开用户管理页面
- **THEN** 页面不显示 `DeptTree` 组件
- **AND** 用户列表区域以单栏全宽布局展示

### Requirement: 用户编辑表单增加部门和岗位字段
系统 SHALL 仅在 `org-management` 已安装并启用时在用户编辑表单中展示部门选择和岗位多选字段；插件缺失时这些字段被隐藏。

#### Scenario: 组织插件缺失时隐藏部门岗位字段
- **WHEN** `org-management` 未安装或未启用，且管理员打开用户编辑抽屉
- **THEN** 表单中不显示部门字段和岗位字段
- **AND** 用户仍可完成基础信息和角色信息的编辑

### Requirement: 用户列表增加部门名称列
系统 SHALL 仅在 `org-management` 已安装并启用时在用户列表表格中展示部门名称列；插件缺失时该列被隐藏。

#### Scenario: 组织插件缺失时隐藏部门列
- **WHEN** `org-management` 未安装或未启用，且管理员查看用户列表表格
- **THEN** 表格中不显示 `部门` 列
- **AND** 其余核心用户列继续正常展示
