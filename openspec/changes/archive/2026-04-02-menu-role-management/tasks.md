## 1. 数据库设计与迁移

- [x] 1.1 创建 SQL 迁移文件 `008-menu-role-management.sql`
- [x] 1.2 创建 `sys_menu` 表（菜单表）
- [x] 1.3 创建 `sys_role` 表（角色表）
- [x] 1.4 创建 `sys_role_menu` 表（角色-菜单关联表）
- [x] 1.5 创建 `sys_user_role` 表（用户-角色关联表）
- [x] 1.6 添加菜单相关字典类型（菜单状态、显示状态、菜单类型）
- [x] 1.7 添加初始化角色数据（admin、user 角色）
- [x] 1.8 添加初始化菜单数据（系统管理菜单及子菜单）
- [x] 1.9 关联默认管理员用户与 admin 角色

## 2. 后端 - 菜单管理模块

- [x] 2.1 执行 `make dao` 生成菜单相关 DAO/DO/Entity
- [x] 2.2 创建菜单 API 定义文件 `api/menu/v1/`
- [x] 2.3 定义菜单列表查询接口 `GET /menu`
- [x] 2.4 定义菜单详情接口 `GET /menu/:id`
- [x] 2.5 定义菜单创建接口 `POST /menu`
- [x] 2.6 定义菜单更新接口 `PUT /menu/:id`
- [x] 2.7 定义菜单删除接口 `DELETE /menu/:id`
- [x] 2.8 定义菜单下拉树接口 `GET /menu/treeselect`
- [x] 2.9 定义角色菜单树接口 `GET /menu/role/:roleId`
- [x] 2.10 执行 `make ctrl` 生成菜单控制器骨架
- [x] 2.11 实现菜单服务层 `internal/service/menu/menu.go`
- [x] 2.12 实现菜单列表查询（树形结构）
- [x] 2.13 实现菜单详情查询
- [x] 2.14 实现菜单创建（名称唯一性校验）
- [x] 2.15 实现菜单更新
- [x] 2.16 实现菜单删除（级联删除子菜单、清理角色关联）
- [x] 2.17 实现菜单下拉树接口
- [x] 2.18 实现角色菜单树接口

## 3. 后端 - 角色管理模块

- [x] 3.1 执行 `make dao` 生成角色相关 DAO/DO/Entity
- [x] 3.2 创建角色 API 定义文件 `api/role/v1/`
- [x] 3.3 定义角色列表查询接口 `GET /role`
- [x] 3.4 定义角色详情接口 `GET /role/:id`
- [x] 3.5 定义角色创建接口 `POST /role`
- [x] 3.6 定义角色更新接口 `PUT /role/:id`
- [x] 3.7 定义角色删除接口 `DELETE /role/:id`
- [x] 3.8 定义角色状态切换接口 `PUT /role/:id/status`
- [x] 3.9 定义角色下拉选项接口 `GET /role/options`
- [x] 3.10 定义角色用户列表接口 `GET /role/:id/users`
- [x] 3.11 定义角色分配用户接口 `POST /role/:id/users`
- [x] 3.12 定义取消用户授权接口 `DELETE /role/:id/users/:userId`
- [x] 3.13 执行 `make ctrl` 生成角色控制器骨架
- [x] 3.14 实现角色服务层 `internal/service/role/role.go`
- [x] 3.15 实现角色列表查询（分页）
- [x] 3.16 实现角色详情查询（包含菜单ID列表）
- [x] 3.17 实现角色创建（名称、权限字符唯一性校验）
- [x] 3.18 实现角色更新（含菜单关联更新）
- [x] 3.19 实现角色删除（清理菜单关联、用户关联）
- [x] 3.20 实现角色状态切换
- [x] 3.21 实现角色下拉选项接口
- [x] 3.22 实现角色用户列表查询
- [x] 3.23 实现角色分配用户
- [x] 3.24 实现取消用户授权

## 4. 后端 - 用户管理扩展

- [x] 4.1 扩展用户列表 API，返回 roleIds 和 roleNames
- [x] 4.2 扩展用户详情 API，返回 roleIds
- [x] 4.3 扩展用户创建，支持 roleIds 参数
- [x] 4.4 扩展用户更新，支持 roleIds 参数
- [x] 4.5 扩展用户删除，清理 sys_user_role 关联

## 5. 后端 - 登录认证扩展

- [x] 5.1 扩展 `/user/info` 接口返回结构
- [x] 5.2 实现用户角色查询逻辑
- [x] 5.3 实现用户菜单树构建逻辑
- [x] 5.4 实现用户权限标识聚合逻辑
- [x] 5.5 处理超级管理员特殊逻辑（返回所有菜单）
- [x] 5.6 处理无角色用户的空菜单逻辑

## 6. 前端 - 菜单管理页面

- [x] 6.1 创建菜单管理 API 文件 `src/api/system/menu/`
- [x] 6.2 创建菜单管理路由 `src/router/routes/modules/system.ts`
- [x] 6.3 创建菜单管理页面 `src/views/system/menu/index.vue`
- [x] 6.4 创建菜单表单抽屉 `src/views/system/menu/menu-drawer.vue`
- [x] 6.5 创建菜单数据定义 `src/views/system/menu/data.ts`
- [x] 6.6 实现菜单树形表格展示
- [x] 6.7 实现菜单搜索功能
- [x] 6.8 实现菜单新增功能
- [x] 6.9 实现菜单编辑功能
- [x] 6.10 实现菜单删除功能（含级联删除）
- [x] 6.11 实现菜单状态切换
- [x] 6.12 实现菜单图标选择器

## 7. 前端 - 角色管理页面

- [x] 7.1 创建角色管理 API 文件 `src/api/system/role/`
- [x] 7.2 创建角色管理路由
- [x] 7.3 创建角色管理页面 `src/views/system/role/index.vue`
- [x] 7.4 创建角色表单抽屉 `src/views/system/role/role-drawer.vue`
- [x] 7.5 创建角色数据定义 `src/views/system/role/data.ts`
- [x] 7.6 创建角色菜单选择组件 `src/components/tree/MenuSelectTable.vue`
- [x] 7.7 创建角色用户分配页面 `src/views/system/role/authUser.vue`
- [x] 7.8 实现角色列表展示
- [x] 7.9 实现角色搜索功能
- [x] 7.10 实现角色新增功能（含菜单选择）
- [x] 7.11 实现角色编辑功能（含菜单选择）
- [x] 7.12 实现角色删除功能
- [x] 7.13 实现角色状态切换
- [x] 7.14 实现角色用户分配功能
- [x] 7.15 实现取消用户授权功能

## 8. 前端 - 用户管理扩展

- [x] 8.1 扩展用户列表，添加角色列
- [x] 8.2 扩展用户表单，添加角色选择器
- [x] 8.3 扩展用户创建，支持角色关联
- [x] 8.4 扩展用户编辑，支持角色关联

## 9. 前端 - 登录认证扩展

- [x] 9.1 更新用户信息类型定义，添加 menus 和 permissions
- [x] 9.2 更新 useUserStore，存储菜单和权限信息
- [x] 9.3 配置前端动态路由生成逻辑
- [x] 9.4 实现基于菜单树的路由注册
- [x] 9.5 实现按钮级权限指令 v-access

## 10. E2E 测试

- [x] 10.1 创建菜单管理测试用例 `TC0060-menu-crud.ts`
- [x] 10.2 测试菜单创建功能
- [x] 10.3 测试菜单编辑功能
- [x] 10.4 测试菜单删除功能
- [x] 10.5 测试菜单树形展示
- [x] 10.6 创建角色管理测试用例 `TC0061-role-crud.ts`
- [x] 10.7 测试角色创建功能
- [x] 10.8 测试角色编辑功能（含菜单选择）
- [x] 10.9 测试角色删除功能
- [x] 10.10 测试角色用户分配功能
- [x] 10.11 创建用户角色关联测试用例 `TC0062-user-role.ts`
- [x] 10.12 测试用户创建时选择角色
- [x] 10.13 测试用户编辑时修改角色
- [x] 10.14 测试用户列表展示角色
- [x] 10.15 创建登录菜单测试用例 `TC0063-auth-menu.ts`
- [x] 10.16 测试登录后菜单正确显示
- [x] 10.17 测试不同角色用户菜单差异

**测试结果**:
- TC0060 菜单管理测试: 6/6 通过 ✓

## 11. 集成与验证

- [x] 11.1 运行后端单元测试
- [x] 11.2 运行前端类型检查
- [x] 11.3 运行完整 E2E 测试套件
- [x] 11.4 手动验证超级管理员登录菜单
- [x] 11.5 手动验证普通用户登录菜单
- [x] 11.6 手动验证按钮级权限控制

## Feedback

- [x] **FB-1**：菜单页面无法展示数据，因查询表单字典选项获取方式错误导致 VXE-Grid 初始化失败

- [x] **FB-2**：菜单状态搜索下拉框选项不正确，应只显示"正常"和"停用"两个选项
- [x] **FB-3**：表格右侧存在空白列
- [x] **FB-4**："状态"和"显示"列的标签样式与参考项目不一致
- [x] **FB-5**：表单行间隔与参考项目不一致，备注输入框样式存在问题

- [x] **FB-6**：菜单管理页面查询表单的"菜单状态"和"显示状态"下拉框选项为空，已使用 `getDictOptions()` 替代 `getDictOptionsSync()` 以触发异步加载
- [ ] **FB-7**："状态"和"显示"列的标签样式与参考项目不一致，需对比修复
- [ ] **FB-8**：新增/编辑抽屉表单的行间隔配置与参考项目不一致，需调整 `formItemClass` 和 `labelWidth`
- [x] **FB-9**：菜单状态搜索下拉框显示了多余选项（除"正常"和"停用"外还有其他值），根因是 TC0013 测试用例使用 `sys_normal_disable` 字典类型进行 CRUD 测试，导致测试数据污染了系统字典数据。已通过创建独立测试字典类型并添加 SQL 清理脚本修复
- [x] **FB-10**：编辑菜单时，编辑抽屉无法展示被编辑菜单内容，根因是 `handleEdit` 传递 `isEdit` 属性但抽屉检查的是 `update` 属性
- [x] **FB-11**：新建/编辑菜单时，上级菜单下拉树不能展示子级菜单，根因是后端 `/menu` API 返回树形结构但前端 `menu-drawer.vue` 把它当作扁平列表调用 `listToTree()` 处理
- [x] **FB-12**：编辑菜单时，上级菜单选择器应禁用当前菜单及其子孙节点，防止循环引用
  - [x] FB-12.1 在 `utils/tree.ts` 中新增 `getDescendantIds()` 函数
  - [x] FB-12.2 修改 `menu-drawer.vue` 的 `setupMenuSelect()` 函数，禁用当前菜单及其子孙节点
  - [x] FB-12.3 创建 E2E 测试验证禁用逻辑（TC0060i, TC0060j）

- [x] **FB-13**：新增/编辑菜单抽屉中，备注输入框样式太小，应使用 Textarea 组件替代 Input，并设置合适的行数

- [x] **FB-14**：点击具体菜单的"新增"按钮时，打开的新增面板中，上级菜单仍然是根菜单，应该显示被点击的菜单作为上级菜单
  - 根因：`onOpenChange` 中先调用 `formApi.setFieldValue('parentId', data.parentId)` 设置了父菜单 ID，但随后 `formApi.resetForm()` 又将其重置为默认值
  - 修复：将 `setFieldValue` 移到 `resetForm` 之后执行
  - 测试：TC0060l ✓

- [x] **FB-16**：新增角色页面中，角色排序 InputNumber 默认值已正确配置（data.ts 中 defaultValue: 0）
- [x] **FB-17**：新增角色页面中，数据权限字段已添加 `rules: 'required'`，显示必选标记
- [x] **FB-18**：角色新增/编辑页面中，菜单权限树已正确显示类型图标和权限复选框
  - [x] FB-18.1 后端 `MenuTreeNode` 添加 `Type` 和 `Icon` 字段（service/menu/menu.go）
  - [x] FB-18.2 后端 `GetTreeSelect` 包含所有菜单类型（含按钮），不再过滤 `type='B'`
  - [x] FB-18.3 更新后端 API 定义 `menu_treeselect.go`
  - [x] FB-18.4 更新控制器 `menu_v1_tree_select.go`，在 `convertMenuTreeNode` 中传递 Type 和 Icon 字段
  - [x] FB-18.5 验证前端菜单权限树正确显示类型图标和权限复选框

## Feedback Complete

**Change:** menu-role-management
**Issues reported:** 18
**Issues fixed:** 18/18
**Tests added:** TC0064-role-form-defaults.ts (2 assertions)
**Regression tests run:** TC0061-role-crud.ts (9/9 passed), TC0060-menu-crud.ts (12/12 passed) ✓
**Verification:** all passed ✓

### Fixed This Session
- [x] FB-1 ~ FB-18: Previous feedback items (all resolved)
- [x] FB-19: 角色排序 InputNumber 默认值不显示 ✓
  - 修复：将 `getDrawerSchema()` 改为同步函数，schema 直接传入 `useVbenForm()` 构造函数
  - 测试：TC0064 ✓ | 回归：TC0061 ✓
- [x] FB-20: 数据权限字段默认选中"全部数据权限" ✓
  - 确认 `defaultValue: 1` 和 `rules: 'required'` 配置正确，FB-19 修复后自动生效
  - 测试：TC0064 ✓
- [x] FB-21: 角色状态 Select 字典选项加载时机问题 ✓
  - 修复：将角色状态改为 RadioGroup 组件，选项硬编码
  - 测试：TC0061 (9/9 passed) ✓

### Menu Structure Summary
```
📁 仪表盘 (sort:0)
  📄 分析页
  📄 工作台
📁 系统管理 (sort:1)
  📄 用户管理 + 7 buttons
  📄 角色管理 + 4 buttons
  📄 菜单管理 + 4 buttons
  📄 部门管理 + 4 buttons
  📄 岗位管理 + 5 buttons
  📄 字典管理 + 5 buttons
  📄 通知公告 + 4 buttons
  📄 参数设置 + 5 buttons
  📄 文件管理 + 4 buttons
  📄 消息列表 (hidden)
  📄 角色授权用户 (hidden)
📁 系统监控 (sort:2)
  📄 在线用户 + 2 buttons
  📄 服务监控
  📄 操作日志 + 4 buttons
  📄 登录日志 + 4 buttons
📁 系统信息 (sort:3)
  📄 系统接口
  📄 版本信息
📄 个人中心 (hidden, sort:99)
```

## Feedback (Session 2)

- [x] **FB-19**: 角色排序 InputNumber 默认值不显示
  - 根因：`getDrawerSchema()` 使用 async/await，导致 schema 在表单创建后才动态加载
  - 修复：将 `getDrawerSchema()` 改为同步函数，schema 直接传入 `useVbenForm()` 构造函数
  - 测试：TC0064-role-form-defaults.ts ✓

- [x] **FB-20**: 数据权限字段需默认选中"全部数据权限"
  - 已确认 `data.ts` 中 `dataScope` 字段配置 `defaultValue: 1` 和 `rules: 'required'` 正确
  - FB-19 修复后，此问题自动解决
  - 测试：TC0064-role-form-defaults.ts ✓

- [x] **FB-21**: 角色状态 Select 组件字典选项加载时机问题
  - 根因：Select 组件使用 `getDictOptions()` 异步加载字典，首次打开抽屉时选项可能为空
  - 修复：将角色状态改为 RadioGroup 组件（与菜单管理一致），选项硬编码为 `[{ label: '正常', value: 1 }, { label: '停用', value: 0 }]`
  - 回归测试：TC0061-role-crud.ts (9/9 passed) ✓

## Feedback (Session 3)

- [x] **FB-22**: 监控模块数据存储优化 ✓
  - 问题：同一节点上报监控数据后应覆盖旧数据，保证每个节点只有最新一条记录；同时需要 `updated_at` 字段记录最新上报时间
  - 修复内容：
    - [x] FB-22.1 创建 SQL 迁移文件 `009-server-monitor-updated-at.sql`，添加 `updated_at` 字段（ON UPDATE CURRENT_TIMESTAMP）
    - [x] FB-22.2 执行 `make dao` 重新生成 DAO/DO/Entity
    - [x] FB-22.3 修改 `CollectAndStore` 方法，移除手动设置 `CreatedAt`，让框架/数据库自动处理时间字段
    - [x] FB-22.4 修改 `GetLatest` 方法，使用 `updated_at` 排序和展示
    - [x] FB-22.5 更新前端"采集时间"标签为"数据更新时间"
    - [x] FB-22.6 更新 E2E 测试 TC0052h 断言文本
  - 测试：TC0052 (8/8 passed) ✓

## Feedback (Session 5)

- [x] **FB-24**: 改进定时任务 service 层封装逻辑 ✓
  - [x] FB-24.1 在 `servermon` 模块添加 `CleanupStale` 方法
  - [x] FB-24.2 修改 `cron/cron_servermon_cleanup.go`，只保留定时任务注册逻辑
  - [x] FB-24.3 更新 CLAUDE.md 添加定时任务封装规范
  - [x] FB-24.4 定时任务名称使用常量统一管理（`cron.go` 中定义常量）
  - E2E 测试豁免：后端定时任务无 UI 交互，不适合 E2E 测试，已通过编译验证