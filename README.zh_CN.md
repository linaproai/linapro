# Lina

Lina 是一个 AI驱动的全栈开发框架。它结合了基于 GoFrame 的核心宿主服务、基于 Vue 3 + Vben 5 的默认管理工作台、插件扩展能力，以及基于 OpenSpec 的 AI 协作交付流程。

## 核心特性

- `apps/lina-core`：核心宿主服务，提供可复用的模块 API、共享平台能力、治理服务和插件运行时支持。
- `apps/lina-vben`：默认管理工作台，也是项目的参考前端应用。
- `apps/lina-plugins`：源码插件与动态插件样例，以及插件开发参考。
- `openspec/`：结构化交付所需的提案、设计、规范和任务文档。
- `hack/tests`：覆盖用户可见行为的 Playwright E2E 测试。

## 仓库结构

```text
apps/
  lina-core/      核心宿主服务（GoFrame）
  lina-vben/      默认管理工作台（Vue 3 + Vben 5）
  lina-plugins/   插件样例与插件开发参考
hack/
  tests/          Playwright E2E 测试集
openspec/
  changes/        活跃与已归档的 OpenSpec 变更
  specs/          当前生效的能力基线规范
```

## 快速开始

### 环境要求

- Go
- Node.js
- pnpm
- MySQL

### 常用命令

```bash
make dev          # 启动前后端
make stop         # 停止本地服务
make status       # 查看本地服务状态
make init         # 执行宿主 SQL 和 Seed 数据
make mock         # 加载演示 / Mock 数据
make test         # 运行 Playwright 测试
```

后端开发：

```bash
cd apps/lina-core
go run main.go
make build
make ctrl
make dao
```

前端开发：

```bash
cd apps/lina-vben
pnpm install
pnpm -F @lina/web-antd dev
pnpm run build
```

## 默认账号

- 用户名：`admin`
- 密码：`admin123`

## 交付流程

Lina 使用 OpenSpec 作为结构化交付主线。

1. 先探索需求与方案空间。
2. 在 `openspec/changes/` 下创建 OpenSpec 提案。
3. 按任务清单渐进式实现。
4. 在改代码的同时同步更新测试与文档。
5. 验证、审查并在验收后归档。

## 文档规范

- 所有目录级主说明文档统一使用英文 `README.md`。
- 任何存在 `README.md` 的目录，都必须同步提供中文镜像 `README.zh_CN.md`。
- 两份文档必须保持相同结构和相同技术事实。

## 入口索引

- `CLAUDE.md`：仓库级工程规则与流程说明。
- `apps/lina-plugins/README.md`：插件系统总览。
- `openspec/specs/`：当前生效的能力基线规范。
