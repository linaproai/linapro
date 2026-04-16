## MODIFIED Requirements

### Requirement: 用户列表查询
系统 SHALL 提供用户列表分页查询接口，支持多字段排序、增强的条件筛选和按部门过滤。

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

#### Scenario: 按部门过滤用户列表
- **WHEN** 查询时传入 `deptId` 参数
- **THEN** 通过 sys_user_dept 关联表筛选属于该部门的用户
- **THEN** 返回的用户数据中包含 deptId 和 deptName 字段

#### Scenario: 用户列表返回部门名称
- **WHEN** 查询用户列表
- **THEN** 每条用户数据中包含 deptName 字段（通过 LEFT JOIN sys_user_dept 和 sys_dept 获取）

### Requirement: 创建用户
系统 SHALL 提供创建用户接口，支持关联部门和岗位。

#### Scenario: 创建用户成功
- **WHEN** 调用 `POST /api/v1/user` 并提交用户名、密码、昵称等信息
- **THEN** 系统创建用户并返回用户 ID

#### Scenario: 创建用户关联部门
- **WHEN** 创建用户时提交 deptId 参数
- **THEN** 系统在 sys_user_dept 表中创建用户与部门的关联记录

#### Scenario: 创建用户关联岗位
- **WHEN** 创建用户时提交 postIds 参数（数组）
- **THEN** 系统在 sys_user_post 表中创建用户与各岗位的关联记录

#### Scenario: 用户名重复
- **WHEN** 创建用户时提交已存在的用户名
- **THEN** 系统返回错误信息，提示用户名已存在

#### Scenario: 必填字段校验
- **WHEN** 创建用户时缺少用户名或密码
- **THEN** 系统返回参数校验错误

### Requirement: 更新用户信息
系统 SHALL 提供更新用户信息接口，支持更新部门和岗位关联。

#### Scenario: 更新用户成功
- **WHEN** 调用 `PUT /api/v1/user/{id}` 并提交要更新的字段
- **THEN** 系统更新对应用户信息并返回成功

#### Scenario: 更新用户部门关联
- **WHEN** 更新用户时提交 deptId 参数
- **THEN** 系统更新 sys_user_dept 表中的关联记录（先删后插）

#### Scenario: 更新用户岗位关联
- **WHEN** 更新用户时提交 postIds 参数（数组）
- **THEN** 系统更新 sys_user_post 表中的关联记录（先删后插）

#### Scenario: 更新不存在的用户
- **WHEN** 更新一个不存在的用户 ID
- **THEN** 系统返回错误信息，提示用户不存在

### Requirement: 查看用户详情
系统 SHALL 提供用户详情查询接口，返回关联的部门和岗位信息。

#### Scenario: 查询用户详情
- **WHEN** 调用 `GET /api/v1/user/{id}`
- **THEN** 返回该用户的完整信息（不含密码）
- **THEN** 包含 deptId（关联部门 ID）、deptName（部门名称）
- **THEN** 包含 postIds（关联岗位 ID 数组）

### Requirement: 用户部门树接口
系统 SHALL 提供用于用户管理左侧筛选的部门树接口，包含"未分配部门"虚拟节点和各节点用户数量。

#### Scenario: 获取用户部门树
- **WHEN** 调用 `GET /api/v1/user/dept-tree`
- **THEN** 返回部门树形结构数据，每个节点包含 id、label、children、userCount
- **THEN** 每个部门节点的 label 格式为"部门名(N)"，N 为该部门关联的用户数量
- **THEN** 树的第一层（与根节点同级）包含一个"未分配部门"虚拟节点（id 为 -1）

#### Scenario: 未分配部门虚拟节点
- **WHEN** 部门树返回数据
- **THEN** 包含一个 id 为 -1 的"未分配部门"虚拟节点
- **THEN** 该节点的 userCount 为未关联任何部门的用户总数

#### Scenario: 按未分配部门过滤用户
- **WHEN** 查询用户列表时传入 `deptId=-1`
- **THEN** 返回所有未在 sys_user_dept 表中有关联记录的用户

## ADDED Requirements

### Requirement: 用户管理前端部门树筛选
系统 SHALL 在用户管理页面左侧增加 DeptTree 组件用于按部门筛选用户。

#### Scenario: 左树右表布局
- **WHEN** 打开用户管理页面
- **THEN** 左侧显示 DeptTree 组件，右侧显示用户列表
- **THEN** 布局与岗位管理页面一致

#### Scenario: 部门筛选联动
- **WHEN** 在左侧选择某个部门
- **THEN** 右侧用户列表自动按该部门过滤（传入 deptId 参数）
- **WHEN** 取消部门选择
- **THEN** 右侧显示全部用户

### Requirement: 用户编辑表单增加部门和岗位字段
系统 SHALL 在用户编辑表单中增加部门选择和岗位多选字段。

#### Scenario: 部门选择字段
- **WHEN** 打开用户编辑抽屉
- **THEN** 表单中包含部门字段（TreeSelect 组件）
- **THEN** TreeSelect 显示完整部门路径（如 "Lina科技 / 研发部门"）
- **THEN** 支持搜索、展开全部节点

#### Scenario: 岗位联动选择
- **WHEN** 用户选择了部门
- **THEN** 自动加载该部门下的岗位选项到岗位多选字段
- **THEN** 清空之前选择的岗位
- **WHEN** 部门下无岗位
- **THEN** 岗位字段 placeholder 显示"该部门下暂无岗位"

#### Scenario: 编辑时回显
- **WHEN** 编辑已有用户
- **THEN** 自动回显用户关联的部门和岗位
- **THEN** 岗位选项列表为该用户所属部门下的岗位

### Requirement: 用户列表增加部门名称列
系统 SHALL 在用户列表表格中增加部门名称列。

#### Scenario: 显示部门名称
- **WHEN** 查看用户列表表格
- **THEN** 表格中包含"部门"列，显示用户所属部门名称
- **THEN** 未关联部门的用户该列为空
