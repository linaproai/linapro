## ADDED Requirements

### Requirement: 部门列表查询
系统 SHALL 提供部门的树形列表查询接口（不分页）。

#### Scenario: 查询部门列表
- **WHEN** 调用 `GET /api/v1/dept`
- **THEN** 返回全部部门数据的平铺列表，前端通过 parentId 构建树形结构
- **THEN** 按 order_num 升序排列

#### Scenario: 部门列表支持条件筛选
- **WHEN** 查询时传入筛选参数 `name`（部门名称）或 `status`（状态）
- **THEN** `name` 使用模糊匹配（LIKE）
- **THEN** `status` 使用精确匹配

#### Scenario: 部门列表排除已删除记录
- **WHEN** 查询部门列表
- **THEN** 结果中不包含已软删除的记录

### Requirement: 创建部门
系统 SHALL 提供创建部门的接口。

#### Scenario: 创建部门成功
- **WHEN** 调用 `POST /api/v1/dept` 并提交 parentId、name、orderNum 等字段
- **THEN** 系统创建部门，自动计算 ancestors 字段（如 "0,1,2"），并返回成功

#### Scenario: 创建根部门
- **WHEN** 创建部门时 parentId 为 0
- **THEN** 该部门为根部门，ancestors 为 "0"

#### Scenario: 必填字段校验
- **WHEN** 创建部门时缺少 name
- **THEN** 系统返回参数校验错误

### Requirement: 更新部门
系统 SHALL 提供更新部门信息的接口。

#### Scenario: 更新部门成功
- **WHEN** 调用 `PUT /api/v1/dept/{id}` 并提交要更新的字段
- **THEN** 系统更新对应部门信息并返回成功

#### Scenario: 不能将部门设为自身的子部门
- **WHEN** 更新部门时将 parentId 设为自身 ID 或自身的子部门 ID
- **THEN** 系统返回错误信息，提示上级部门不能是自身或其子部门

#### Scenario: 更新部门时同步更新子部门 ancestors
- **WHEN** 部门的 parentId 发生变更
- **THEN** 系统自动更新该部门及所有子部门的 ancestors 字段

### Requirement: 删除部门
系统 SHALL 提供删除部门的接口。

#### Scenario: 删除部门成功
- **WHEN** 调用 `DELETE /api/v1/dept/{id}`
- **THEN** 部门被软删除

#### Scenario: 不能删除有子部门的部门
- **WHEN** 删除一个有子部门的部门
- **THEN** 系统返回错误信息，提示该部门下存在子部门，须先删除子部门

#### Scenario: 不能删除有关联用户的部门
- **WHEN** 删除一个在 sys_user_dept 中有关联用户的部门
- **THEN** 系统返回错误信息，提示该部门下存在用户，须先移除用户

### Requirement: 查看部门详情
系统 SHALL 提供部门详情查询接口。

#### Scenario: 查询部门详情
- **WHEN** 调用 `GET /api/v1/dept/{id}`
- **THEN** 返回该部门的完整信息

### Requirement: 部门树形结构接口
系统 SHALL 提供用于 TreeSelect 组件的部门树接口。

#### Scenario: 获取完整部门树
- **WHEN** 调用 `GET /api/v1/dept/tree`
- **THEN** 返回树形结构数据，每个节点包含 id、label（部门名称）、children

#### Scenario: 获取排除指定节点的部门树
- **WHEN** 调用 `GET /api/v1/dept/exclude/{id}`
- **THEN** 返回排除该节点及其所有子节点的部门列表
- **THEN** 用于编辑部门时选择上级部门（避免循环引用）

### Requirement: 部门数据表设计
系统 SHALL 提供 sys_dept 表和 sys_user_dept 关联表。

#### Scenario: sys_dept 表结构
- **WHEN** 查看 sys_dept 表结构
- **THEN** 表包含：id、parent_id、ancestors、name、order_num、leader（INTEGER，引用 sys_user.id）、phone、email、status、remark、created_at、updated_at、deleted_at

#### Scenario: sys_user_dept 关联表结构
- **WHEN** 查看 sys_user_dept 表结构
- **THEN** 表包含：user_id（INTEGER）、dept_id（INTEGER），联合主键
- **THEN** user_id 引用 sys_user.id，dept_id 引用 sys_dept.id

### Requirement: 部门管理前端树形表格
系统 SHALL 在部门管理页面使用 VXE-Grid 的树形模式展示部门层级。

#### Scenario: 树形展示
- **WHEN** 打开部门管理页面
- **THEN** 使用 VXE-Grid treeConfig（parentField: 'parentId', rowField: 'id', transform: true）渲染树形表格
- **THEN** 默认展开所有节点

#### Scenario: 展开/折叠操作
- **WHEN** 点击工具栏"展开全部"按钮
- **THEN** 展开所有树节点
- **WHEN** 点击工具栏"折叠全部"按钮
- **THEN** 折叠所有树节点
- **WHEN** 双击某一行
- **THEN** 切换该节点的展开/折叠状态

#### Scenario: 表格列定义
- **WHEN** 查看部门列表表格
- **THEN** 显示以下列：部门名称（树节点）、排序、状态（DictTag 渲染）、创建时间、操作

#### Scenario: 行操作按钮
- **WHEN** 查看每行的操作列
- **THEN** 显示三个按钮：编辑（ghost）、新增子部门（ghost，绿色）、删除（ghost，红色，Popconfirm 确认）

### Requirement: 部门编辑抽屉
系统 SHALL 提供 600px 宽度的 Drawer 用于新增和编辑部门。

#### Scenario: 新增部门表单
- **WHEN** 点击"新增根部门"或"新增子部门"按钮
- **THEN** 打开 Drawer，表单字段包括：上级部门（TreeSelect，显示完整路径）、部门名称（必填）、排序（必填，默认 0）、负责人（Select，disabled）、联系电话（正则校验）、邮箱（邮箱校验）、状态（RadioGroup 按钮样式）
- **THEN** 新增子部门时，上级部门自动填入当前部门

#### Scenario: 编辑部门表单
- **WHEN** 点击编辑按钮
- **THEN** 打开 Drawer，加载现有数据
- **THEN** 负责人字段变为可用（Select），选项列表为该部门下的用户（通过 sys_user_dept 查询）
- **THEN** 上级部门 TreeSelect 排除自身及子部门节点

### Requirement: DeptTree 可复用组件
系统 SHALL 提供可复用的 DeptTree 组件，供用户管理和岗位管理使用。

#### Scenario: DeptTree 组件功能
- **WHEN** 使用 DeptTree 组件
- **THEN** 显示部门树形结构，支持搜索、刷新、单选
- **THEN** 通过 v-model:selectDeptId 绑定选中的部门 ID
- **THEN** 默认展开所有节点

#### Scenario: DeptTree 搜索
- **WHEN** 在搜索框输入关键字
- **THEN** 过滤显示匹配的部门节点

#### Scenario: DeptTree 刷新
- **WHEN** 点击刷新按钮
- **THEN** 重新加载部门树数据并触发 reload 事件

### Requirement: 部门初始化数据
系统 SHALL 提供基础的部门初始化数据。

#### Scenario: 初始化部门结构
- **WHEN** 执行 v0.2.0 数据库迁移脚本
- **THEN** 创建以下部门结构：
  - Lina科技（根部门，id=1）
    - 研发部门（id=2）
    - 市场部门（id=3）
    - 测试部门（id=4）
    - 财务部门（id=5）
    - 运维部门（id=6）
