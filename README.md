<div align="center">
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

English | [简体中文](README.zh-CN.md)

</div>

# Overview

`LinaPro` is an **AI-native full-stack framework built for sustainable delivery**. It brings together a specification-driven AI development workflow, a comprehensive AI skill system spanning the entire development lifecycle, a complete plugin runtime, and an integrated full-stack design — with enterprise-grade capabilities like access control, system configuration, and job scheduling built right in.

Teams can skip the infrastructure-from-scratch phase and put AI to work driving real business development from day one.

# Quick Links

| Resource | URL |
|----------|-----|
| **Repository** | https://github.com/linaproai/linapro |
| **Live Demo** | http://demo.linapro.ai/ <br/>Username: `admin` <br/>Password: `admin123` |
| **Website** | https://linapro.ai/ |

# Core Capabilities

`LinaPro` is designed for individual developers, engineering teams, and enterprises. Here's what it brings to the table:

- **AI-native development workflow**: Ships with a specification-driven AI development workflow, with first-class support for the optional but recommended `OpenSpec` tool. AI leads analysis, design, and implementation while every change is anchored to incremental specs and mandatory E2E tests — freeing your team to focus on direction rather than execution details.
- **A rich AI skill ecosystem**: Over a dozen built-in AI skills cover the full development lifecycle — backend development, frontend design, test writing, code review, performance auditing, version upgrades, and more. AI makes framework-aware decisions in every context without needing to be re-briefed each session.
- **Fast business development**: A batteries-included management workspace and a rich set of built-in modules dramatically shorten the path from zero to production.
- **Integrated full-stack design**: Frontend and backend are designed as a unified whole — API contracts, permission models, and design conventions are fully aligned, with no manual cross-framework integration overhead.
- **Complete API documentation**: All host and plugin API endpoints are automatically aggregated and exposed as a single browsable, debuggable doc site.
- **Extensible plugin ecosystem**: A dual-mode plugin system — source plugins and `WASM` dynamic plugins — lets any capability be extended or replaced. Official plugins are maintained as a separate submodule and pulled in only when needed, keeping the core framework lean.
- **Multi-tenant support**: Native multi-tenant capability with an official multi-tenant management plugin. When the plugin is not enabled, the system automatically falls back to single-tenant mode with zero migration cost.
- **Enterprise-grade governance**: JWT authentication paired with a declarative RBAC permission system, plus built-in operation logs, login logs, and session management for comprehensive auditability.
- **Distribution-ready by design**: Built-in distributed locking, key-value caching, and horizontal scaling. Cluster mode is coordinated via Redis for high availability — no changes to business code required.

# Architecture

```mermaid
graph TB
    subgraph Workflow["AI Development Workflow  openspec/"]
        direction LR
        Explore["🔍 Explore"] --> Propose["📋 Propose"] --> Implement["⚙️ Implement"] --> Review["🔎 Review"] --> Archive["📦 Archive"]
    end

    subgraph Frontend["Management Workspace  lina-vben"]
        UI["Vue 3 + Vben5 + Ant Design"]
    end

    subgraph Host["Core Host Service  lina-core"]
        direction TB
        API["API Layer\n(g.Meta route definitions + DTO)"]
        Ctrl["Controller Layer\n(HTTP request handling)"]
        Svc["Service Layer\n(core business logic)"]
        Plugin["Plugin Runtime\n(lifecycle orchestration · sandbox isolation)"]
        Tenant["Native Multi-Tenant\n(bizctx · tenant_id)"]
        Gov["Governance\n(JWT · RBAC · Logs · Sessions)"]
        API --> Ctrl --> Svc
        Svc --> Plugin
        Svc --> Tenant
        Svc --> Gov
    end

    subgraph Plugins["Plugin System  apps/lina-plugins"]
        direction LR
        Source["Source Plugins\ncompiled with host"]
        Dynamic["WASM Dynamic Plugins\nhot-loaded at runtime"]
    end

    DB[("Data Store\nPostgreSQL")]
    Redis[("Cluster Coordination\nRedis")]

    Workflow -.->|spec-driven| Frontend
    Workflow -.->|spec-driven| Host
    UI -->|HTTP| API
    Plugin -->|compiled load| Source
    Plugin -->|sandbox execution| Dynamic
    Svc --> DB
    Gov --> DB
    Svc -.->|cluster.enabled=true| Redis
```

# Screenshots

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

# Tech Stack

| Category | Technology | Notes |
|----------|------------|-------|
| Backend Language | `Go` | `v1.25.0` |
| Backend Framework | `GoFrame` | `v2.10.1` — routing, ORM, configuration, and more |
| Frontend Framework | `Vue 3` | Built on the `Vben 5` admin template |
| Frontend UI | `Ant Design Vue` | Enterprise-grade UI component library |
| Build Tool | `Vite` | Lightning-fast frontend builds |
| Database | `PostgreSQL` | Default data store |
| Plugin Runtime | `WebAssembly` | `tetratelabs/wazero`, powering WASM dynamic plugins |
