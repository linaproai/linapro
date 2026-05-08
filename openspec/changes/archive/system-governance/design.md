## Context

LinaPro 管理后台当前已具备用户管理、部门管理、岗位管理、字典管理、通知公告、文件管理等基础业务模块。系统使用纯无状态 JWT 认证，后端不跟踪用户会话。缺少操作审计、系统监控、API 文档展示和运行时参数配置等系统治理能力。

技术栈：GoFrame v2 + MySQL + JWT（HS256），前端 Vben5 + Vue 3 + Ant Design Vue。已有完整的中间件链（CORS → 响应包装 → 上下文注入 → JWT 认证），以及基于 `g.Meta` 的路由定义机制。现有导出功能使用 `excelize` 库生成 xlsx 格式文件。

## Goals / Non-Goals

**Goals:**
- 通过中间件自动记录所有写操作到操作日志表，在登录/登出流程中记录登录日志
- 实现基于 MySQL MEMORY 引擎的会话跟踪机制，支持在线用户列表查询和强制下线
- 实现基于 gopsutil 的服务器指标定时采集，写入 MySQL，支持多节点分布式部署
- 集成 Stoplight Elements 作为 OpenAPI 文档 UI，提供系统运行时信息展示
- 提供系统参数的完整 CRUD 管理，支持 Excel 导出导入
- 优化字典管理导出导入功能，支持合并导出导入

**Non-Goals:**
- 不实现 IP 地理位置解析
- 不实现日志的软删除（日志清理为硬删除）
- 不实现实时推送（WebSocket）刷新监控数据
- 不实现历史指标趋势图表，仅展示最新一次采集的快照数据
- 不实现报警/告警功能
- 不实现参数缓存机制
- 不自行托管 vben5 组件演示的静态文件

## Decisions

### 一、审计日志体系

#### 1. 操作日志记录方式 — 中间件自动拦截 + API 标签混合方案

**选择**：在中间件层自动拦截写操作，并通过 `g.Meta` 标签标记需要特殊记录的查询操作。

**备选方案**：
- A. 纯中间件方案：全局拦截所有写操作，无需修改业务代码。但无法精确区分操作类型和模块名。
- B. 纯服务层埋点：在每个 Service 方法中手动调用日志记录。精确但侵入性强，容易遗漏。
- C. 本方案（A+B 混合）：中间件自动处理 POST/PUT/DELETE，通过 `g.Meta` 的 `tags` 字段获取模块名、`summary` 获取操作描述。对于需要记录的 GET 操作（如导出），在 `g.Meta` 中添加 `operLog` 标签。

**理由**：GoFrame 的 `g.Meta` 已包含 `tags`（模块名）和 `summary`（操作描述），中间件可通过 `r.GetServeHandler()` 获取这些元信息，无需额外编码即可获得操作语义。

#### 2. 操作类型（oper_type）推断规则

根据 HTTP 方法和 `g.Meta` 中的 `operLog` 标签自动推断：

| HTTP 方法 | 默认 oper_type | 说明 |
|-----------|---------------|------|
| POST | 1（新增） | 如果路径含 `import` 则为 5（导入） |
| PUT | 2（修改） | — |
| DELETE | 3（删除） | — |
| GET + `operLog:"4"` | 4（导出） | 仅标记了 operLog 的 GET 请求 |
| 其他 | 6（其他） | 自定义标签值 |

#### 3. 请求/响应参数记录策略

**截断长度**：请求参数和响应结果各截断至 **2000 字符**。超出部分直接截断并追加 `...(truncated)`。

**脱敏规则**：对请求参数 JSON 中的 `password`、`Password` 字段值替换为 `***`。

**理由**：2000 字符足以记录绝大多数业务操作的参数，同时避免批量导入等场景下日志表膨胀。仅对密码字段脱敏，其他字段保留原值以便审计。

#### 4. 日志不使用软删除

操作日志和登录日志表**不设 `deleted_at` 字段**，清理操作直接执行 `DELETE FROM` 硬删除。

**理由**：日志数据本身是审计记录，对日志进行软删除会导致日志表无限增长且无法真正释放存储空间。清理功能的目的就是释放空间，硬删除更符合预期。

#### 5. 中间件注册位置

操作日志中间件注册在 Auth 中间件**之后**，确保已获取当前用户信息（用户名）。

```
CORS → ResponseHandler → Ctx → Auth → OperLog → Controller
```

#### 6. 登录日志记录位置

直接在 Auth Service 的登录/登出方法中调用 LoginLog Service 写入记录，而非通过中间件。

**理由**：登录接口在 Auth 中间件之前（公开接口），中间件无法拦截。且登录日志需要记录登录结果（成功/失败），只有在业务逻辑中才能准确获取。

#### 7. 异步写入

操作日志中间件在请求处理完成后，通过 **goroutine 异步写入**数据库，避免影响接口响应时间。

### 二、系统可观测性

#### 8. 会话存储使用 MySQL MEMORY 引擎

**选择**: MySQL MEMORY 引擎表 `sys_online_session`

**备选方案**:
- A) gcache 内存缓存：性能最优，但进程重启丢失，且不支持跨实例共享
- B) Redis：支持分布式，但引入额外组件依赖
- C) MySQL MEMORY 引擎：通过 MySQL 存储在内存中，性能接近内存缓存，无需额外依赖

**理由**: MEMORY 引擎降低组件依赖复杂度，利用已有 MySQL 基础设施。MySQL 重启后数据丢失可以接受（用户重新登录即可）。通过定义 `SessionStore` 抽象接口，未来可无缝切换到 gcache + Redis 方案。

#### 9. 服务监控采用定时采集 + 数据库存储

**选择**: gopsutil 定时采集 → 写入 `sys_server_monitor` 表 → API 从数据库读取

**备选方案**:
- A) 实时查询（API 请求时现场采集）：简单但不支持多节点，且响应慢
- B) 定时采集写数据库：支持多节点、无状态部署，历史数据可查

**理由**: 定时采集使 Lina 服务保持无状态。每个节点独立采集自身指标并写入数据库，新增节点只需部署服务即可自动上报。

**采集参数**:
- 采集频率：默认 30 秒
- 数据保留：每个节点只保留最新一条记录（UPSERT 策略）
- 节点标识：hostname + 本机 IP 自动获取

#### 10. 会话活跃时间跟踪与自动清理

`sys_online_session` 表新增 `last_active_time` 字段，登录时初始化，每次请求通过 UPDATE 操作更新并根据受影响行数判断会话是否存在。定时任务清理超时会话，超时阈值和清理频率通过配置文件调整。

### 三、系统自我描述

#### 11. OpenAPI 文档 UI 选型：Stoplight Elements（iframe 方案）

**选择**: 使用静态 HTML 文件 + iframe 嵌入 Stoplight Elements

**演进路径**: Scalar → Stoplight Elements（Web Component）→ Stoplight Elements（iframe）

**理由**: Scalar 的 API Client 弹窗被遮挡，Stoplight Elements 通过 Web Component 集成后 CSS 污染全局样式，最终采用 iframe 嵌入实现样式完全隔离。文档 HTML 改为静态文件方式提供，移除后端 API 路由，减少系统复杂度。

#### 12. 系统信息页面架构

**选择**: 后端提供 `GET /api/v1/system/info` 接口 + 前端配置对象

- **后端 API 返回**: Go 版本、GoFrame 版本、操作系统、数据库版本、系统启动时间、运行时长等运行时信息
- **前端配置对象**: 项目名称、版本、描述、许可证、主页链接、后端组件列表、前端组件列表
- 外链地址集中在前端配置文件中定义，修改时无需改动组件代码

#### 13. 组件演示方案：iframe 嵌入外部网站

**选择**: iframe 嵌入 `https://www.vben.pro/`

**理由**: vben.pro 未设置 X-Frame-Options 限制，可正常嵌入。零体积增加，零维护成本。加载失败时展示友好错误页面。

### 四、运行时配置管理

#### 14. 参数设置数据表设计

表名 `sys_config`，字段保持简洁：id、name、key（UNIQUE）、value、remark、created_at、updated_at、deleted_at。

**决策**：`key` 和 `value` 虽为 MySQL 保留字，但在 GoFrame 的 ORM 中会自动用反引号包裹，不影响使用。

#### 15. 参数设置 API 设计

遵循 RESTful 规范：GET `/config`（列表）、GET `/config/{id}`（详情）、POST `/config`（新增）、PUT `/config/{id}`（修改）、DELETE `/config/{id}`（删除）、GET `/config/key/{key}`（按键名查询）、GET `/config/export`（导出）、POST `/config/import`（导入）。

#### 16. 字典合并导出导入

新增 `GET /dict/export` 合并导出接口，同时导出字典类型和字典数据到双 Sheet Excel 文件。新增 `POST /dict/import` 合并导入接口，支持同时导入字典类型和字典数据。前端字典类型面板使用合并接口，字典数据面板移除独立的导出导入按钮。

### 五、通用设计决策

#### 17. 浏览器和操作系统解析

通过解析 HTTP 请求头 `User-Agent` 字段获取浏览器和操作系统信息。使用 `mssola/useragent` 库。

#### 18. 前端菜单与路由结构

```
系统监控 (/monitor)
├── 操作日志 (/monitor/operlog)
├── 登录日志 (/monitor/loginlog)
├── 在线用户 (/monitor/online)
└── 服务监控 (/monitor/server)

系统信息 (/about)
├── 系统接口 (/about/api-docs)
├── 版本信息 (/about/system-info)
└── 组件演示 (/about/component-demo)

系统管理 (/system)
└── 参数设置 (/system/config)
```

## Risks / Trade-offs

- **[日志丢失]** 异步写入可能在进程崩溃时丢失少量日志 → 对于后台管理系统可接受
- **[存储增长]** 操作日志记录请求/响应参数，长期运行会占用大量存储 → 提供按时间范围清理功能
- **[MEMORY 引擎限制]** MEMORY 引擎不支持 BLOB/TEXT 类型，所有字段需使用定长类型 → VARCHAR 即可满足需求
- **[MySQL 重启丢失会话]** MySQL 重启后所有在线会话丢失，用户需重新登录 → 对管理后台场景影响可控
- **[采集间隔与实时性]** 30 秒采集间隔意味着前端看到的数据最多有 30 秒延迟 → 对监控场景可接受
- **[iframe 嵌入局限]** 组件演示依赖外部网站可用性 → 展示加载失败提示，不影响其他功能
- **[保留字字段名]** `key`、`value` 为 MySQL 保留字 → GoFrame ORM 自动处理反引号包裹
- **[无缓存机制]** 每次按键名查询都查数据库 → 参数数据量小、查询频率低，当前阶段不需要缓存
