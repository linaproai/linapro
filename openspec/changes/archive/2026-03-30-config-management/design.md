## Context

Lina 系统目前缺少运行时可配置的参数管理能力。本次变更新增参数设置模块（`sys_config`），实现键值对形式的系统参数 CRUD 管理。

现有项目已有成熟的模块实现模式（以 `dict` 模块为参考），前后端均有统一的分层架构：
- 后端：API 定义 → Controller → Service → DAO
- 前端：API 层 → VXE-Grid 表格页面 + Modal 弹窗

参数设置模块将完全复用这套模式，无需引入新的架构或依赖。

## Goals / Non-Goals

**Goals:**
- 提供系统参数的完整 CRUD 管理（列表查询、新增、编辑、删除、批量删除）
- 支持按键名查询参数值（供其他模块内部调用）
- 支持 Excel 导出
- 前端页面与参考项目（ruoyi-plus-vben5）保持交互一致

**Non-Goals:**
- 不实现参数缓存机制和缓存刷新功能
- 不区分「系统内置」与「用户自定义」参数
- 不实现参数值的类型校验（统一以字符串存储）
- 不实现参数变更的审计日志（复用现有操作日志即可）

## Decisions

### 1. 数据表设计

表名 `sys_config`，字段保持简洁：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT UNSIGNED PK | 主键 |
| name | VARCHAR(100) | 参数名称 |
| key | VARCHAR(100) UNIQUE | 参数键名（唯一） |
| value | VARCHAR(500) | 参数键值 |
| remark | VARCHAR(500) | 备注 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 修改时间 |
| deleted_at | DATETIME | 软删除时间 |

**决策**：`key` 和 `value` 虽为 MySQL 保留字，但在 GoFrame 的 ORM 中会自动用反引号包裹，不影响使用。保持简洁的字段命名，不加 `config_` 前缀。

### 2. API 路径设计

遵循 RESTful 规范，与现有模块风格一致：

| 操作 | 方法 | 路径 |
|------|------|------|
| 列表查询 | GET | `/config` |
| 获取详情 | GET | `/config/{id}` |
| 新增 | POST | `/config` |
| 修改 | PUT | `/config/{id}` |
| 删除 | DELETE | `/config/{id}` |
| 按键名查询 | GET | `/config/key/{key}` |
| 导出 | GET | `/config/export` |

### 3. 后端分层

完全复用 dict 模块的实现模式：
- `api/config/v1/` — 每个接口一个文件（`config_list.go`、`config_create.go` 等）
- `internal/controller/config/` — gf gen ctrl 自动生成骨架
- `internal/service/config/` — `config.go` 主文件，包含所有 CRUD 方法
- `internal/dao/` — gf gen dao 自动生成

### 4. 前端实现

复用 dict type 页面的交互模式：
- 搜索栏：参数名称（Input）、参数键名（Input）、创建时间（RangePicker）
- 表格列：☐、参数名称、参数键名、参数键值、备注、修改时间、操作
- 工具栏：导出、批量删除、新增
- 弹窗表单：参数名称、参数键名、参数键值（Textarea）、备注（Textarea）

### 5. 菜单和权限

在系统管理菜单下新增「参数设置」菜单项，权限标识使用 `system:config:*` 前缀：
- `system:config:list` — 查看
- `system:config:add` — 新增
- `system:config:edit` — 编辑
- `system:config:remove` — 删除
- `system:config:export` — 导出

## Risks / Trade-offs

- **[保留字字段名]** `key`、`value` 为 MySQL 保留字 → GoFrame ORM 自动处理反引号包裹，gf gen dao 生成的代码可正常工作，风险可控
- **[无缓存机制]** 每次按键名查询都查数据库 → 参数数据量小、查询频率低，当前阶段不需要缓存，后续按需添加
