## Context

Lina 项目是一个基于 GoFrame + Vben5 的前后端分离管理后台系统。当前已实现用户管理、部门管理、岗位管理、字典管理等基础模块，但缺少菜单管理和角色管理功能，无法实现精细化的权限控制。

### 现有架构

- **后端**: GoFrame v2 框架，采用 Controller → Service → DAO 三层架构
- **前端**: Vben5 + Vue 3 + Ant Design Vue，采用 pnpm monorepo
- **认证**: JWT Token 认证，Session 存储在数据库中
- **权限**: 当前为硬编码的 admin 角色，无动态权限控制

### 参考项目

ruoyi-plus-vben5 项目提供了完整的菜单管理和角色管理实现，本设计将参考其数据模型和交互设计，同时遵循 Lina 项目的架构规范。

## Goals / Non-Goals

**Goals:**

- 实现菜单管理功能，支持树形层级结构、三种菜单类型（目录/菜单/按钮）
- 实现角色管理功能，支持角色与菜单的关联、角色与用户的关联
- 扩展用户管理，支持为用户分配角色
- 扩展登录认证，返回用户菜单树用于前端动态路由生成
- 实现简化的数据权限（全部/本部门/仅本人）

**Non-Goals:**

- 不实现完整的六种数据权限范围（全部/自定义/本部门/本部门及以下/仅本人/部门及以下或本人）
- 不实现数据权限的实际过滤逻辑（后续迭代）
- 不实现角色的"权限"操作功能（数据权限编辑页面）
- 不实现菜单权限的父子不联动选项

## Decisions

### 1. 数据库设计

**决策**: 新增四张表实现菜单-角色-用户关联

```
sys_menu (菜单表)
├── id, parent_id, name, path, component, perms, icon
├── type (D=目录, M=菜单, B=按钮)
├── sort, visible, status, is_frame, is_cache
├── query_param, remark
└── created_at, updated_at, deleted_at

sys_role (角色表)
├── id, name, key, sort
├── data_scope (1=全部, 2=本部门, 3=仅本人)
├── status, remark
└── created_at, updated_at, deleted_at

sys_role_menu (角色-菜单关联表)
├── role_id, menu_id
└── PRIMARY KEY (role_id, menu_id)

sys_user_role (用户-角色关联表)
├── user_id, role_id
└── PRIMARY KEY (user_id, role_id)
```

**理由**:
- 参考 ruoyi-plus-vben5 的成熟设计
- 菜单支持软删除，角色支持软删除
- 多对多关联表设计灵活，支持一个用户多个角色、一个角色多个菜单

### 2. 菜单类型设计

**决策**: 采用三种菜单类型（D/M/B）

| 类型 | 标识 | 说明 | 特性 |
|------|------|------|------|
| 目录 | D (Directory) | 包含子菜单的容器 | 有图标、路由地址、显示排序 |
| 菜单 | M (Menu) | 实际页面 | 有组件路径、权限标识、是否缓存 |
| 按钮 | B (Button) | 页面内按钮权限 | 仅权限标识，无路由 |

**理由**:
- 目录类型用于组织菜单层级结构
- 菜单类型对应实际可访问的页面
- 按钮类型用于控制页面内操作权限

### 3. 登录时菜单获取

**决策**: 登录后通过 `/auth/info` 接口返回用户菜单树

```json
{
  "userId": 1,
  "username": "admin",
  "realName": "管理员",
  "avatar": "/avatar.png",
  "roles": ["admin"],
  "menus": [
    {
      "id": 1,
      "parentId": 0,
      "name": "系统管理",
      "path": "system",
      "icon": "ant-design:setting-outlined",
      "type": "D",
      "children": [...]
    }
  ],
  "permissions": ["system:user:list", "system:user:add"],
  "homePath": "/dashboard"
}
```

**理由**:
- 前端 Vben5 框架支持 backend 模式的动态路由
- 一次性返回菜单树减少请求次数
- 同时返回权限标识列表用于按钮级权限控制

### 4. 菜单父子联动

**决策**: 角色分配菜单时默认父子联动

- 勾选父菜单时自动勾选所有子菜单
- 取消勾选子菜单时自动取消父菜单勾选（直到有其他子菜单被勾选）

**理由**: 简化用户操作，符合直觉，减少配置错误

### 5. 数据权限范围

**决策**: 简化为三种数据权限范围

| 值 | 含义 | 说明 |
|----|------|------|
| 1 | 全部数据 | 可查看所有数据 |
| 2 | 本部门数据 | 仅可查看本部门数据 |
| 3 | 仅本人数据 | 仅可查看自己创建的数据 |

**理由**: 覆盖常见场景，降低实现复杂度，后续可扩展

### 6. API 设计

**菜单管理 API**:
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /menu | 获取菜单列表（树形） |
| GET | /menu/:id | 获取菜单详情 |
| POST | /menu | 创建菜单 |
| PUT | /menu/:id | 更新菜单 |
| DELETE | /menu/:id | 删除菜单 |
| GET | /menu/treeselect | 获取菜单下拉树 |
| GET | /menu/role/:roleId | 获取角色的菜单树 |

**角色管理 API**:
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /role | 获取角色列表（分页） |
| GET | /role/:id | 获取角色详情 |
| POST | /role | 创建角色 |
| PUT | /role/:id | 更新角色 |
| DELETE | /role/:id | 删除角色 |
| PUT | /role/:id/status | 更新角色状态 |
| GET | /role/:id/users | 获取角色的用户列表 |
| POST | /role/:id/users | 为角色分配用户 |
| DELETE | /role/:id/users/:userId | 取消用户角色授权 |

**用户管理 API 扩展**:
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /role/options | 获取角色下拉选项 |

## Risks / Trade-offs

### 风险1: 菜单删除后角色权限断裂

**风险**: 删除菜单后，角色-菜单关联表中存在无效数据

**缓解措施**:
- 删除菜单时同步删除 `sys_role_menu` 中的关联记录
- 使用数据库外键约束（ON DELETE CASCADE）或在代码层面处理

### 风险2: 角色删除后用户权限丢失

**风险**: 删除角色后，用户-角色关联表中存在无效数据

**缓解措施**:
- 删除角色时同步删除 `sys_user_role` 和 `sys_role_menu` 中的关联记录
- 提示用户该角色已分配给多少用户

### 风险3: 超级管理员权限处理

**风险**: 超级管理员角色的菜单权限如何处理

**缓解措施**:
- 预置 `admin` 角色，该角色拥有所有菜单权限（代码层面判断，不依赖关联表）
- 非超级管理员角色通过 `sys_role_menu` 关联表获取权限

### 风险4: 前端动态路由刷新

**风险**: 页面刷新后动态路由丢失

**缓解措施**:
- 前端在路由守卫中检测刷新，重新请求菜单接口
- 或将菜单树存储在 localStorage 中（需处理过期问题）

## Migration Plan

### 部署步骤

1. 执行 SQL 迁移脚本，创建新表和初始化数据
2. 部署后端新版本
3. 部署前端新版本
4. 验证登录、菜单、角色功能

### 初始化数据

- 预置 `admin` 角色（超级管理员）
- 预置 `user` 角色（普通用户）
- 预置系统菜单（系统管理、用户管理等）
- 将 `admin` 角色分配给默认管理员用户

### 回滚策略

- SQL 回滚脚本删除新增表
- 后端回滚到上一版本
- 前端回滚到上一版本

## Open Questions

1. **菜单排序字段**: 使用 `sort` 还是 `order_num`？
   - 决定: 使用 `sort`，与岗位表命名风格一致

2. **角色标识字段**: 使用 `key` 还是 `role_key`？
   - 决定: 使用 `key`，简洁明了

3. **菜单名称是否支持 i18n**: 如何处理？
   - 决定: `name` 字段存储 i18n key（如 `menu.system.user`），前端根据语言包翻译