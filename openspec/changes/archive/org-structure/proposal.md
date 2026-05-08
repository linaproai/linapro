## Why

v0.1.0 完成了用户管理的基础 MVP。作为后台管理系统，字典管理、部门管理和岗位管理是核心基础模块，为后续的角色权限、菜单管理等功能提供数据支撑。字典管理提供全局数据枚举能力（如状态、性别等），部门管理建立组织层级结构，岗位管理定义职位体系。这三个模块的完整实现是系统走向生产可用的关键一步。

## What Changes

- 新增字典管理模块：字典类型和字典数据的完整 CRUD，支持 Tag 样式配置（预设色 + 自定义色 + CSS 类）、导出功能，前端双面板布局
- 新增全局 DictTag 组件和 Pinia 字典缓存 Store，为所有模块提供字典渲染能力
- 新增部门管理模块：树形结构的完整 CRUD，支持层级关系管理、展开/折叠、负责人选择
- 新增岗位管理模块：完整 CRUD，关联部门，左侧部门树筛选，支持导出
- 新增 DeptTree 可复用组件，供用户管理和岗位管理共同使用
- 新增 sys_user_dept 和 sys_user_post 关联表，使用独立表实现用户与部门/岗位的关联（不修改 sys_user 表结构）
- 扩展用户管理：列表页增加部门树侧边栏筛选，编辑表单增加部门选择（TreeSelect）和岗位多选，列表增加部门名称列

## Capabilities

### New Capabilities

- `dict-management`: 字典类型和字典数据的增删改查，Tag 样式系统（预设色/自定义色/CSS 类），双面板 UI，导出功能，全局 DictTag 组件和 Pinia 缓存 Store
- `dept-management`: 部门的增删改查，树形层级结构管理，展开/折叠，负责人从部门用户中选择，DeptTree 可复用组件
- `post-management`: 岗位的增删改查，关联部门，左侧部门树筛选，批量删除，导出功能

### Modified Capabilities

- `user-management`: 用户列表增加部门树侧边栏筛选，用户编辑表单增加部门选择和岗位多选字段，用户与部门/岗位通过独立关联表（sys_user_dept / sys_user_post）建立关联

## Impact

- **数据库**: 新增 6 张表（sys_dict_type, sys_dict_data, sys_dept, sys_post, sys_user_dept, sys_user_post），均在 `manifest/sql/v0.2.0.sql` 中定义
- **后端 API**: 新增字典类型、字典数据、部门、岗位四组 RESTful API；扩展用户 API 支持 deptId/postIds 参数和部门树查询
- **后端 Service**: 新增 dict、dept、post 三个 service 包；扩展 user service 处理关联表
- **前端页面**: 新增字典管理、部门管理、岗位管理三个页面模块；修改用户管理页面
- **前端组件**: 新增 DictTag 全局组件、TagStylePicker 组件、DeptTree 可复用组件
- **前端状态**: 新增 Pinia dict store 用于字典数据缓存和去重请求
- **路由**: system.ts 新增字典管理、部门管理、岗位管理路由
