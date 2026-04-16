## Context

Lina 是一个全新的后台管理系统，当前项目仓库为空（仅有 git 初始化和配置文件）。需要从零搭建前后端项目基础架构，并实现第一个可用功能：用户认证和用户管理。

参考项目：
- 前端：ruoyi-plus-vben5（Vue 3 + Vben5 + Ant Design Vue + VXE-Grid）
- 后端：openspec-demo（GoFrame v2 标准分层架构）

技术约束：
- 数据库使用 SQLite（GoFrame 内置 driver），SQL 语法需兼容 MySQL
- 前端使用 Vben5 最新版官方模板初始化
- 后端严格遵循 GoFrame v2 规范（goframe-v2 技能）
- 每个迭代需要 E2E 测试覆盖

## Goals / Non-Goals

**Goals:**
- 搭建可运行的前后端项目基础架构
- 实现 JWT 用户名密码认证（登录/登出）
- 实现用户管理 CRUD（轻量用户表）
- 搭建管理后台基础布局（侧边栏、顶部导航）
- E2E 测试覆盖登录和用户管理功能

**Non-Goals:**
- 不实现角色/部门/岗位关联（后续迭代）
- 不实现前端动态权限路由（本迭代菜单静态配置）
- 不实现验证码、OAuth、记住密码等高级认证功能
- 不实现用户头像上传
- 不支持 MySQL（本迭代仅 SQLite，但 SQL 兼容）

## Decisions

### D1: 项目目录结构

采用 monorepo 结构，与 openspec-demo 保持一致：

```
apps/
  backend/           → GoFrame v2 后端
    api/             → 请求/响应 DTO（g.Meta 路由定义）
    internal/
      cmd/           → 服务启动 & 路由注册
      consts/        → 常量定义
      controller/    → HTTP 控制器（make ctrl 自动生成骨架）
      dao/           → 数据访问层（make dao 自动生成）
      model/
        do/          → 数据操作对象（自动生成）
        entity/      → 数据库实体（自动生成）
        context.go   → 上下文模型
      service/       → 业务逻辑层
        auth/        → JWT 认证服务
        user/        → 用户管理服务
        middleware/  → HTTP 中间件
        bizctx/      → 业务上下文提取
    manifest/
      config/        → 配置文件（config.yaml）
      sql/           → 数据库 schema
    hack/            → 开发工具
    resource/        → 静态资源
  frontend/          → Vben5 前端（pnpm monorepo）
    apps/web-antd/   → 主应用
    packages/        → 共享库
hack/
  tests/             → E2E 测试（Playwright）
```

**理由**：与 demo 项目结构一致，降低认知负担，便于参考。

### D2: 数据库设计 — 轻量用户表

```sql
CREATE TABLE sys_user (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    username    VARCHAR(64)  NOT NULL,
    password    VARCHAR(256) NOT NULL,
    nickname    VARCHAR(64)  NOT NULL DEFAULT '',
    email       VARCHAR(128) NOT NULL DEFAULT '',
    phone       VARCHAR(20)  NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 1,
    remark      VARCHAR(512) NOT NULL DEFAULT '',
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME,
    UNIQUE(username)
);
```

关键设计：
- `status`: 1=正常, 0=停用
- 软删除：`deleted_at` 非空表示已删除
- 密码使用 bcrypt 哈希存储
- 不包含角色/部门/岗位字段，后续通过关联表扩展
- 字段命名使用 snake_case，兼容 MySQL

**理由**：用户表保持轻量，关联信息通过中间表维护，方便后续迭代扩展。

### D3: 认证方案 — JWT + bcrypt

- 登录：验证用户名密码 → 签发 JWT（HS256）
- Token 载荷：`{userId, username, status}`
- Token 有效期：24 小时（可配置）
- 密码：bcrypt 哈希，默认 cost
- 中间件：从 `Authorization: Bearer <token>` 提取并验证
- 登出：前端清除 token（无服务端 session）

**替代方案**：Session-based auth — 排除，因为 JWT 无状态更简单，且 demo 项目已验证此方案。

### D4: 前端初始化 — Vben5 官方模板

使用 Vben5 最新版官方模板（通过 create-vben 脚手架），选择 Ant Design Vue 变体。然后：
- 清理模板中的示例页面
- 配置 API 代理（`/api` → backend:8080）
- 配置静态菜单（用户管理）
- 参考 ruoyi-plus-vben5 的页面模式实现用户管理

**理由**：从官方模板起步保证与最新版本兼容，比 fork ruoyi 再删减更干净。

### D5: API 设计

所有 API 统一在 `/api` 前缀下：

| 端点 | 方法 | 说明 | 认证 |
|------|------|------|------|
| `/api/v1/auth/login` | POST | 登录 | 否 |
| `/api/v1/auth/logout` | POST | 登出 | 是 |
| `/api/v1/user` | GET | 用户列表（分页） | 是 |
| `/api/v1/user` | POST | 创建用户 | 是 |
| `/api/v1/user/{id}` | GET | 用户详情 | 是 |
| `/api/v1/user/{id}` | PUT | 更新用户 | 是 |
| `/api/v1/user/{id}` | DELETE | 删除用户 | 是 |
| `/api/v1/user/{id}/status` | PUT | 修改用户状态 | 是 |
| `/api/v1/user/profile` | GET | 当前用户信息 | 是 |
| `/api/v1/user/profile` | PUT | 更新当前用户信息 | 是 |

统一响应格式：`{code: 0, message: "ok", data: {...}}`

### D6: E2E 测试方案

使用 Playwright，测试文件放在 `hack/tests/`：
- 测试文件命名：`TC{NNNN}_description.spec.ts`
- 需要前后端同时运行
- 测试覆盖：登录/登出流程、用户 CRUD 完整操作

### D7: 初始数据

系统初始化时需要一个默认管理员账号：
- 用户名：`admin`
- 密码：`admin123`（bcrypt 哈希）
- 状态：正常

通过 SQL 初始化脚本插入。

## Risks / Trade-offs

- **[SQLite 并发限制]** → SQLite 单写锁，但管理后台并发量低，完全够用。后续切换 MySQL 时只需改配置和 driver。
- **[JWT 无法服务端撤销]** → 本迭代不做 Token 黑名单，登出仅清除前端 token。如需强制下线功能，后续迭代可加 Redis Token 黑名单。
- **[静态菜单]** → 本迭代菜单硬编码在前端，不支持动态权限菜单。Change 2（菜单管理）完成后再切换为动态加载。
- **[Vben5 版本兼容]** → Vben5 活跃开发中，API 可能变化。使用最新稳定版并锁定版本。
