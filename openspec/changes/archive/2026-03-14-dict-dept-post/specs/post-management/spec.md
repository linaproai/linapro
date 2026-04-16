## ADDED Requirements

### Requirement: 岗位列表查询
系统 SHALL 提供岗位的分页列表查询接口，支持按部门过滤。

#### Scenario: 查询岗位列表
- **WHEN** 调用 `GET /api/v1/post` 并传入分页参数 `pageNum` 和 `pageSize`
- **THEN** 返回岗位列表和总数，格式为 `{list: [...], total: number}`

#### Scenario: 按部门过滤岗位
- **WHEN** 查询时传入 `deptId` 参数
- **THEN** 仅返回属于该部门的岗位

#### Scenario: 岗位列表支持条件筛选
- **WHEN** 查询时传入筛选参数 `code`（岗位编码）、`name`（岗位名称）或 `status`（状态）
- **THEN** `code` 和 `name` 使用模糊匹配（LIKE）
- **THEN** `status` 使用精确匹配

#### Scenario: 岗位列表排除已删除记录
- **WHEN** 查询岗位列表
- **THEN** 结果中不包含已软删除的记录

### Requirement: 创建岗位
系统 SHALL 提供创建岗位的接口。

#### Scenario: 创建岗位成功
- **WHEN** 调用 `POST /api/v1/post` 并提交 deptId、code、name、sort 等字段
- **THEN** 系统创建岗位并返回成功

#### Scenario: 岗位编码重复
- **WHEN** 创建岗位时提交已存在的 code 值
- **THEN** 系统返回错误信息，提示岗位编码已存在

#### Scenario: 必填字段校验
- **WHEN** 创建岗位时缺少 deptId、code 或 name
- **THEN** 系统返回参数校验错误

### Requirement: 更新岗位
系统 SHALL 提供更新岗位信息的接口。

#### Scenario: 更新岗位成功
- **WHEN** 调用 `PUT /api/v1/post/{id}` 并提交要更新的字段
- **THEN** 系统更新对应岗位信息并返回成功

#### Scenario: 更新不存在的岗位
- **WHEN** 更新一个不存在的岗位 ID
- **THEN** 系统返回错误信息

### Requirement: 删除岗位
系统 SHALL 提供删除岗位的接口，支持批量删除。

#### Scenario: 删除单个岗位
- **WHEN** 调用 `DELETE /api/v1/post/{id}`
- **THEN** 岗位被软删除

#### Scenario: 批量删除岗位
- **WHEN** 调用 `DELETE /api/v1/post/{ids}`，ids 为逗号分隔的多个 ID
- **THEN** 所有指定岗位被软删除

#### Scenario: 不能删除有关联用户的岗位
- **WHEN** 删除一个在 sys_user_post 中有关联用户的岗位
- **THEN** 系统返回错误信息，提示该岗位下存在用户，须先移除用户

### Requirement: 查看岗位详情
系统 SHALL 提供岗位详情查询接口。

#### Scenario: 查询岗位详情
- **WHEN** 调用 `GET /api/v1/post/{id}`
- **THEN** 返回该岗位的完整信息

### Requirement: 导出岗位
系统 SHALL 提供将岗位列表导出为 Excel 文件的功能。

#### Scenario: 导出岗位
- **WHEN** 调用 `GET /api/v1/post/export` 并传入筛选参数
- **THEN** 返回 Excel 文件流
- **THEN** 导出字段包括：岗位编码、岗位名称、排序、状态、备注、创建时间

### Requirement: 岗位部门树接口
系统 SHALL 提供用于岗位管理左侧筛选的部门树接口，包含"未分配部门"虚拟节点。

#### Scenario: 获取岗位部门树
- **WHEN** 调用 `GET /api/v1/post/dept-tree`
- **THEN** 返回部门树形结构数据

#### Scenario: 未分配部门虚拟节点
- **WHEN** 部门树返回数据
- **THEN** 包含一个 id 为 -1 的"未分配部门"虚拟节点

#### Scenario: 按未分配部门过滤岗位
- **WHEN** 查询岗位列表时传入 `deptId=-1`
- **THEN** 返回所有 dept_id 为 0 的岗位（未分配部门的岗位）

### Requirement: 按部门获取岗位选项
系统 SHALL 提供按部门获取岗位选项的接口，供用户编辑表单使用。

#### Scenario: 获取部门下的岗位选项
- **WHEN** 调用 `GET /api/v1/post/option-select` 并传入 `deptId` 参数
- **THEN** 返回该部门下所有正常状态的岗位列表，包含 id 和 name

#### Scenario: 部门下无岗位
- **WHEN** 查询的部门下没有岗位
- **THEN** 返回空列表

### Requirement: 岗位数据表设计
系统 SHALL 提供 sys_post 表和 sys_user_post 关联表。

#### Scenario: sys_post 表结构
- **WHEN** 查看 sys_post 表结构
- **THEN** 表包含：id、dept_id（INTEGER，引用 sys_dept.id）、code（VARCHAR，UNIQUE）、name、sort、status、remark、created_at、updated_at、deleted_at

#### Scenario: sys_user_post 关联表结构
- **WHEN** 查看 sys_user_post 表结构
- **THEN** 表包含：user_id（INTEGER）、post_id（INTEGER），联合主键
- **THEN** user_id 引用 sys_user.id，post_id 引用 sys_post.id

### Requirement: 岗位管理前端左树右表布局
系统 SHALL 在岗位管理页面采用左侧部门树 + 右侧岗位列表的布局。

#### Scenario: 布局结构
- **WHEN** 打开岗位管理页面
- **THEN** 左侧显示 DeptTree 组件（260px 宽度），右侧显示岗位列表（flex-1）

#### Scenario: 部门筛选联动
- **WHEN** 在左侧选择某个部门
- **THEN** 右侧岗位列表自动按该部门过滤
- **WHEN** 取消部门选择
- **THEN** 右侧显示全部岗位

#### Scenario: 表格列定义
- **WHEN** 查看岗位列表表格
- **THEN** 显示以下列：勾选框、岗位编码、岗位名称、排序、状态（DictTag 渲染）、创建时间、操作

#### Scenario: 工具栏操作
- **WHEN** 查看工具栏
- **THEN** 显示：新增按钮（primary）、批量删除按钮（danger，勾选后启用）、导出按钮

#### Scenario: 行操作按钮
- **WHEN** 查看每行的操作列
- **THEN** 显示两个按钮：编辑（ghost）、删除（ghost，红色，Popconfirm 确认）

### Requirement: 岗位编辑抽屉
系统 SHALL 提供 600px 宽度的 Drawer 用于新增和编辑岗位。

#### Scenario: 岗位表单字段
- **WHEN** 打开岗位编辑 Drawer
- **THEN** 表单字段包括：所属部门（TreeSelect，必填，显示完整路径）、岗位名称（必填）、岗位编码（必填）、排序（必填，默认 0）、状态（RadioGroup 按钮样式，默认正常）、备注（Textarea，全宽）
- **THEN** 表单采用 2 列网格布局

### Requirement: 岗位初始化数据
系统 SHALL 提供基础的岗位初始化数据。

#### Scenario: 初始化岗位数据
- **WHEN** 执行 v0.2.0 数据库迁移脚本
- **THEN** 创建以下岗位数据：
  - 总经理（code: CEO，dept: Lina科技，sort: 1）
  - 技术总监（code: CTO，dept: 研发部门，sort: 2）
  - 项目经理（code: PM，dept: 研发部门，sort: 3）
  - 开发工程师（code: DEV，dept: 研发部门，sort: 4）
  - 测试工程师（code: QA，dept: 测试部门，sort: 5）
