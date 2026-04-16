## Context

当前 Lina 系统已完成用户管理、部门管理、岗位管理、字典管理、通知公告、操作日志等核心业务模块（v0.1.0 ~ v0.4.0）。系统缺乏自身的元信息展示能力——管理员无法查看 API 文档、系统运行状态、技术栈信息。

后端 GoFrame 框架已自动在 `/api.json` 暴露 OpenAPI v3 规范文件，前端基于 Vben5 + Ant Design Vue。Vben5 框架自带 `<About />` 组件可作为系统信息页面的样式参考。

## Goals / Non-Goals

**Goals:**
- 新增"系统信息"顶级菜单，包含系统接口、系统信息、组件演示三个子页面
- 集成 Scalar 作为 OpenAPI 文档 UI，支持在线测试接口
- 提供系统运行时信息（Go 版本、OS、数据库、启动时间等）的后端 API 和前端展示
- 通过 iframe 嵌入 vben5 官网演示页面，提供组件演示能力
- 外链第三方地址做成配置对象，方便后续修改

**Non-Goals:**
- 不自行托管 vben5 组件演示的静态文件
- 不实现自定义的 OpenAPI 文档渲染器
- 不展示系统性能监控指标（CPU/内存/磁盘等）

## Decisions

### 1. OpenAPI 文档 UI 选型：Scalar

**选择**: 使用 `@scalar/api-reference` npm 包

**替代方案**:
- Swagger UI：经典方案，但 UI 较为陈旧
- Rapidoc：Web Component 方案，维护活跃度一般
- Stoplight Elements：功能强大但体积较大

**理由**: Scalar UI 现代美观，支持在线接口测试（Try it），提供 Vue 组件可直接集成，活跃维护，体积合理。前端直接引入 Vue 组件，指向后端的 `/api.json` 端点即可。

### 2. 系统信息页面架构

**选择**: 后端提供 `GET /api/v1/system/info` 接口 + 前端配置对象

- **后端 API 返回**: Go 版本、GoFrame 版本、操作系统、数据库版本、系统启动时间、运行时长等运行时信息
- **前端配置对象**: 项目名称、版本、描述、许可证、主页链接、后端组件列表（名称+版本+链接）、前端组件列表（名称+版本+链接）
- 外链地址集中在前端配置文件中定义，修改时无需改动组件代码

**理由**: 运行时信息只有后端知道，必须通过 API 获取；而项目介绍和组件列表在编译时即确定，放在前端配置中更灵活。

### 3. 组件演示方案：iframe 嵌入外部网站

**选择**: iframe 嵌入 `https://www.vben.pro/`

**替代方案**:
- 嵌入 playground 构建产物：6.2MB 原始文件，增加后端体积
- Lina 菜单 + iframe 子页面：需 fork 修改 playground 源码，维护成本高

**理由**: vben.pro 未设置 X-Frame-Options 限制，可正常嵌入。零体积增加，零维护成本。加载失败时展示友好错误页面，用户体验可接受。

### 4. 菜单与路由结构

新增顶级菜单"系统信息"（路径 `/about`），与现有"系统管理"（`/system`）区分开：

```
系统信息 (/about)
├── 系统接口 (/about/api-docs)
├── 系统信息 (/about/system-info)
└── 组件演示 (/about/component-demo)
```

### 5. 后端模块组织

新增 `system` API 模块（注意：与已有的 `system` 菜单下的用户管理等模块不同，这里是系统自身信息）：

```
api/system/v1/info.go          → 系统信息请求/响应 DTO
internal/controller/system/    → 系统信息控制器
internal/service/system/       → 系统信息服务层
```

## Risks / Trade-offs

- **[风险] vben.pro 网站不可用** → 组件演示页面展示加载失败提示，不影响其他功能。提供配置项允许修改嵌入地址。
- **[风险] Scalar 包体积** → Scalar Vue 组件约 200-300KB gzip，作为独立路由按需加载，不影响首屏性能。
- **[权衡] iframe 嵌入 vs 本地静态文件** → 选择 iframe 牺牲了离线可用性，换取零维护和零体积增加。
- **[权衡] 系统信息 API 无鉴权 vs 鉴权** → 系统信息接口应放在鉴权路由组内，仅登录用户可查看，避免泄露系统版本信息。
