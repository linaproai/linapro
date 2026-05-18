<div align=center>
<img src="https://linapro.ai/img/linapro-logo.png" width="300" alt="linapro logo"/>

[![LinaPro CI](https://github.com/linaproai/linapro/actions/workflows/main-ci.yml/badge.svg)](https://github.com/linaproai/linapro/actions/workflows/main-ci.yml)
[![LinaPro Stable Release](https://img.shields.io/github/v/release/linaproai/linapro?style=flat&v=2)](https://github.com/linaproai/linapro/releases)
[![LinaPro License](https://img.shields.io/github/license/linaproai/linapro.svg?style=flat&v=2)](https://github.com/linaproai/linapro)
[![LinaPro Is Production Ready](https://img.shields.io/badge/production-ready-blue.svg?style=flat)](https://github.com/linaproai/linapro)

[![Vue](https://img.shields.io/badge/Vue-3.x-42b883.svg)](https://vuejs.org/)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)](https://golang.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.x-3178C6.svg)](https://www.typescriptlang.org/)
[![Vite](https://img.shields.io/badge/Vite-6.x-646CFF.svg)](https://vitejs.dev/)
[![Element Plus](https://img.shields.io/badge/Element_Plus-2.x-409EFF.svg)](https://element-plus.org/)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-3.x-06B6D4.svg)](https://tailwindcss.com/)
[![GoFrame](https://img.shields.io/badge/GoFrame-v2-00ADD8.svg)](https://goframe.org/)
[![Art Design Pro](https://img.shields.io/badge/Art_Design_Pro-UI-FF6B6B.svg)](https://github.com/Daymychen/art-design-pro)

[English](README.md) | 简体中文

</div>

# 项目介绍

`LinaPro`是一款**面向可持续交付的`AI`原生全栈框架**，将规范驱动的`AI`研发工作流、全生命周期`AI`技能体系、完整插件运行时与前后端一体化全栈设计融为一体，并内置权限管理、系统配置、任务调度等企业级基础能力，为团队构建起一套完整的`AI`原生交付底座。

团队无需从零搭建基础设施，从第一天起就能以`AI`作为主力驱动业务开发和持续交付。


# 快速链接

| 资源 | 地址 |
|------|------|
| **开源仓库** | https://github.com/linaproai/linapro |
| **后台演示** | http://demo.linapro.ai/ <br/>账号：`admin` <br/>密码：`admin123`|
| **官方网站** | https://linapro.ai/ |

# 项目定位

`LinaPro`面向独立开发者、研发团队和企业，提供以下核心能力：

- **AI 原生研发工作流**：内置规范驱动的`AI`研发工作流，对可选但推荐的`OpenSpec`提供良好支持，让`AI`主导分析、设计与实现，每次变更均锚定在增量规范与强制`E2E`测试上，团队专注于方向决策
- **丰富的 AI 技能体系**：内置十余项覆盖研发全生命周期的专属`AI`技能，涵盖后端开发、前端设计、测试编写、代码审查、性能审计、版本升级等场景，让`AI`在每个具体工作环节都能做出符合框架约束的专业决策
- **快速业务开发**：开箱即用的管理工作台与丰富的内置模块，显著缩短项目从零到上线的时间
- **全栈一体化**：前后端统一设计，接口契约、权限模型与设计规范完全对齐，无需独立集成两套框架
- **完整 API 文档**：自动聚合宿主与所有插件接口，支持在线浏览与调试
- **插件生态**：双模式插件系统（源码插件 +`WASM`动态插件），任意能力均可通过插件扩展或替换；官方插件以`submodule`形式独立维护，按需引入，不增加主框架负担
- **多租户支持**：框架原生支持多租户能力，提供官方多租户管理插件，未启用时自动回退单租户模式，迁移零成本
- **企业级治理**：`JWT`认证配合声明式`RBAC`权限体系，内置操作日志、登录日志、会话管理等审计能力
- **原生分布式**：底层支持分布式锁、键值缓存、水平扩展，集群模式基于`Redis`协调器实现高可用，无需改造业务代码

# 技术架构

```mermaid
graph TB
    subgraph Workflow["AI 研发工作流  openspec/"]
        direction LR
        Explore["🔍 探索"] --> Propose["📋 提案"] --> Implement["⚙️ 实现"] --> Review["🔎 审查"] --> Archive["📦 归档"]
    end

    subgraph Frontend["默认管理工作台  lina-vben"]
        UI["Vue 3 + Vben5 + Ant Design"]
    end

    subgraph Host["核心宿主服务  lina-core"]
        direction TB
        API["API 接口层\n(g.Meta 路由定义 + DTO)"]
        Ctrl["控制器层\n(HTTP 请求处理)"]
        Svc["业务服务层\n(核心业务逻辑)"]
        Plugin["插件运行时\n(生命周期编排 · 沙箱隔离)"]
        Tenant["原生多租户\n(bizctx · tenant_id)"]
        Gov["治理服务\n(JWT · RBAC · 日志 · 会话)"]
        API --> Ctrl --> Svc
        Svc --> Plugin
        Svc --> Tenant
        Svc --> Gov
    end

    subgraph Plugins["插件系统  apps/lina-plugins"]
        direction LR
        Source["源码插件\n随宿主编译交付"]
        Dynamic["WASM 动态插件\n运行时热加载"]
    end

    DB[("数据存储\nPostgreSQL")]
    Redis[("集群协调\nRedis")]

    Workflow -.->|规范驱动| Frontend
    Workflow -.->|规范驱动| Host
    UI -->|HTTP| API
    Plugin -->|编译加载| Source
    Plugin -->|沙箱执行| Dynamic
    Svc --> DB
    Gov --> DB
    Svc -.->|cluster.enabled=true| Redis
```

# 工作台示例

<table>
  <tr>
    <td><img src="https://linapro.ai/img/preview/linapro-i18n.webp?v=v0.1.0" /></td>
    <td><img src="https://linapro.ai/img/preview/linapro-plugin.webp?v=v0.1.0" /></td>
    <td><img src="https://linapro.ai/img/preview/linapro-apidoc.webp?v=v0.1.0" /></td>
  </tr>
  <tr>
    <td><img src="https://linapro.ai/img/preview/linapro-menu.webp?v=v0.1.0" /></td>
    <td><img src="https://linapro.ai/img/preview/linapro-cron.webp?v=v0.1.0" /></td>
    <td><img src="https://linapro.ai/img/preview/linapro-monitor.webp?v=v0.1.0" /></td>
  </tr>
  <tr>
    <td><img src="https://linapro.ai/img/preview/linapro-sysconfig.webp?v=v0.1.0" /></td>
    <td><img src="https://linapro.ai/img/preview/linapro-user.webp?v=v0.1.0" /></td>
    <td><img src="https://linapro.ai/img/preview/linapro-multitenant-select.webp?v=v0.1.0" /></td>
  </tr>
</table>

# 主要技术栈

| 类别 | 技术 | 说明 |
|------|------|------|
| 后端语言 | `Go` | `v1.25.0` |
| 后端框架 | `GoFrame` | `v2.10.1`，提供路由、`ORM`、配置等全套能力 |
| 前端框架 | `Vue 3` | 基于`Vben 5`管理台模板 |
| 前端 UI | `Ant Design Vue` | 企业级 `UI` 组件库 |
| 构建工具 | `Vite` | 极速前端构建 |
| 数据库 | `PostgreSQL` | 默认数据存储 |
| 插件运行时 | `WebAssembly` | `tetratelabs/wazero`，支持`WASM`动态插件 |
