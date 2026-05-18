# 为`LinaPro`做贡献

[English](CONTRIBUTING.md) | 简体中文

感谢你关注并参与`LinaPro`。本文档说明项目背景、开发环境、研发流程和贡献要求。

- [项目概述](#项目概述)
- [默认账号](#默认账号)
- [目录结构](#目录结构)
- [常用命令](#常用命令)
- [官方插件工作区](#官方插件工作区)
- [快速开始](#快速开始)
- [开发流程](#开发流程)
- [变更开发](#变更开发)
- [提交规范](#提交规范)
- [Pull Request 流程](#pull-request-流程)
- [发布标签](#发布标签)
- [代码规范](#代码规范)
- [测试](#测试)
- [i18n 指南](#i18n-指南)
- [社区](#社区)

# 项目概述

`LinaPro`是一个`面向可持续交付的AI原生全栈框架`，把`AI`作为核心生产力：`AI`主导分析、设计与实现，团队把握方向与关键决策。

`LinaPro`将后端服务、前端工作台、可插拔插件体系与规范驱动的`AI`研发工作流融为一体，构成一套完整的全栈交付框架。框架内置开箱即用的默认管理工作台，覆盖了绝大多数业务系统所需的权限管理、系统配置、任务调度等基础能力，让新项目无需从零搭建即可直接投入业务开发。

- **前端**：`Vben5 + Vue 3 + Ant Design Vue + TypeScript`
- **后端**：`GoFrame + PostgreSQL + JWT + 源码/WASM插件运行时`
- **研发流程**：`SDD + OpenSpec`

# 默认账号

- 用户名称：`admin`
- 登录密码：`admin123`

# 目录结构

```text
apps/                -> MonoRepo项目目录
  lina-core/         -> 全栈开发框架的核心宿主服务（GoFrame）
    api/             -> 请求/响应DTO（g.Meta路由定义）
    internal/        -> 后端核心代码实现
      cmd/           -> 服务启动与路由注册
      controller/    -> HTTP控制器（gf gen ctrl自动生成骨架）
      dao/           -> 数据访问层（gf gen dao自动生成）
      model/         -> 数据模型
        do/          -> 数据操作对象（自动生成）
        entity/      -> 数据库实体（自动生成）
      service/       -> 业务逻辑层
    manifest/        -> 交付清单
      config/        -> 后端配置文件
      sql/           -> DDL + Seed DML（版本SQL文件）
        mock-data/   -> Mock演示/测试数据（不随生产部署）
  lina-vben/         -> 默认管理工作台（Vben5前端pnpm monorepo）
    apps/web-antd/   -> 默认管理工作台应用（Ant Design Vue）
    packages/        -> 共享库（@core、effects、stores、utils等）
  lina-plugins/      -> 官方源码插件仓库submodule挂载入口
    <plugin-id>/     -> 源码插件目录（统一结构）
      backend/       -> 插件后端入口与实现
        api/         -> 插件API DTO与路由接口定义
        internal/    -> 插件后端内部实现
          controller/ -> HTTP控制器
          service/    -> 业务逻辑层
          dao/        -> 数据访问层（gf gen dao自动生成，按需生成）
          model/      -> 数据模型（按需生成）
            do/       -> 数据操作对象（自动生成）
            entity/   -> 数据库实体（自动生成）
        hack/        -> 插件codegen与开发配置，如hack/config.yaml
        plugin.go    -> 插件后端注册入口
      frontend/      -> 插件前端页面与资源
      manifest/      -> 插件安装、卸载与交付资源
      hack/          -> 插件级研发与测试资源
        tests/
          e2e/       -> 插件自有TC测试用例
          pages/     -> 插件自有E2E页面对象
          support/   -> 插件自有E2E helper
      plugin.yaml    -> 插件清单
      plugin_embed.go -> 插件嵌入资源入口
hack/                -> 项目脚本及测试用例文件
  tests/             -> 宿主与共享E2E测试（Playwright）
    e2e/             -> 宿主TC测试用例；源码插件自有E2E放在插件目录
    fixtures/        -> 测试fixtures（auth、config）
    pages/           -> 宿主/共享页面对象模型
openspec/            -> OpenSpec相关文档
  changes/           -> OpenSpec变更记录
```

# 常用命令

## 开发环境

```bash
make dev                         # 启动前后端（前端: 5666，后端: 8080）
make dev plugins=0               # 强制宿主模式启动（不加载官方源码插件）
make stop                        # 停止所有服务
make status                      # 查看服务状态
make dev.setup                  # 安装前端依赖及Playwright浏览器（仅首次）
make test                        # 运行完整E2E测试
make init                        # 初始化数据库（DDL + Seed数据）
make mock                        # 加载Mock演示数据（需先执行init）
make image tag=v0.6.0            # 构建生产Docker镜像
make release.tag.check tag=v0.6.0 # 校验发布标签必须等于metadata.yaml中的framework.version
make up                          # 默认用claude生成commit message并推送
make up tool=codex               # 使用codex生成commit message并推送
make up t=codex                  # tool的短别名
make up tool=codex               # codex默认模型为gpt-5.1-codex-mini
make up tool=codex model=gpt-5.2 # 指定AI工具和模型（兼容m=...）
```

当`apps/lina-plugins`存在插件清单时，`make image`会自动启用插件完整模式。可追加`registry=ghcr.io/linaproai push=1`构建并推送镜像。

## 后端

```bash
cd apps/lina-core
go run main.go          # 运行后端
make build              # 构建
make dao                # 修改SQL后生成DAO/DO/Entity
make ctrl               # 修改API定义后生成控制器骨架
```

## 前端

```bash
cd apps/lina-vben
pnpm install                   # 安装依赖
pnpm -F @lina/web-antd dev     # 开发模式
pnpm run build                 # 构建
```

## E2E 测试

首次运行前需安装前端依赖和 Playwright 浏览器：

```bash
make dev.setup
```

```bash
cd hack/tests
pnpm test              # 运行全部测试
pnpm test:headed       # 带浏览器界面运行
pnpm test:ui           # 交互式测试界面
pnpm test:debug        # 调试测试
pnpm report            # 查看HTML报告
```

测试文件命名规范为`TC{NNNN}*.ts`，例如`TC0001-login.ts`。宿主测试放在`hack/tests/e2e/`对应模块目录下；源码插件自有测试放在`apps/lina-plugins/<plugin-id>/hack/tests/e2e/`，插件专属页面对象和helper放在同插件的`hack/tests/pages/`、`hack/tests/support/`。

# 官方插件工作区

- 官方源码插件仓库独立维护在`https://github.com/linaproai/official-plugins.git`。
- 主仓库通过`apps/lina-plugins` `submodule`挂载官方插件仓库。
- `host-only`命令可在未初始化`submodule`时运行。
- `make dev`、`make build`、`make image`和`make image.build`在未显式传入`plugins`时会自动检测`apps/lina-plugins/*/plugin.yaml`；存在插件清单则启用插件完整模式，否则为宿主模式。需要强制宿主模式时传入`plugins=0`。
- 插件完整模式会基于宿主专用的根目录`go.work`自动生成或刷新已忽略的`temp/go.work.plugins`，并通过`GOWORK`解析源码插件`Go`模块；根目录`go.work`始终保持`host-only`。
- 插件专属测试和插件`E2E`需要先执行`git submodule update --init --recursive`。
- 本地`submodule`管理`remote`使用`git@github.com:linaproai/official-plugins.git`。

# 快速开始

## 环境要求

| 工具 | 版本 | 用途 |
|------|------|------|
| `Go` | `1.25.0+` | 后端开发 |
| `Node.js` | `22+` | 前端开发 |
| `pnpm` | `10+` | 前端包管理器 |
| `PostgreSQL` | `14+` | 默认数据库 |
| `make` | 任意 | 兼容任务入口 |

## 快速配置

```bash
# 1. Fork仓库并克隆你的fork
git clone https://github.com/<your-username>/linapro.git
cd linapro

# 2. 需要插件完整模式时初始化官方插件工作区
git submodule update --init --recursive

# 3. 初始化数据库（DDL + Seed数据）
make init

# 4. 启动全栈服务（前端: 5666，后端: 8080）
make dev
```

# 开发流程

`LinaPro`使用`OpenSpec`规范驱动开发流程。所有非平凡贡献都应遵循以下五阶段闭环：

```text
探索 -> 提案 -> 实现 -> 审查 -> 归档
```

| 阶段 | 命令 | 产物 |
|------|------|------|
| **探索** | `/opsx:explore` | 对问题与方案空间形成共识 |
| **提案** | `/opsx:propose <feature-name>` | 生成包含`proposal.md`、`design.md`、`specs/`和`tasks.md`的`openspec/changes/<feature>/` |
| **实现** | `/opsx:apply` | 与`tasks.md`对齐的代码、测试和文档 |
| **审查** | `/lina-review` | 代码质量与规范合规性自动审查 |
| **归档** | `/opsx:archive` | 变更移动到`openspec/changes/archive/` |

简单且独立的缺陷修复可以跳过探索和提案，直接进入实现，但所有变更在合并前都必须通过审查。

# 变更开发

## 后端（`lina-core`）

后端使用`GoFrame`。自动生成文件禁止手动修改。

```bash
cd apps/lina-core

# 修改api/{resource}/v1/*.go后
make ctrl

# 新增或修改manifest/sql/*.sql后
make init
make dao
```

关键规则：

- 返回给调用方的业务错误必须使用`pkg/bizerr`封装。
- 所有日志统一使用`pkg/logger`，禁止直接调用`g.Log()`。
- 始终沿完整调用链传递`ctx context.Context`。
- 所有`error`返回值都必须显式处理。
- 枚举必须定义为具名`Go`类型与常量，业务逻辑中不要使用裸字符串字面量。

## 前端（`lina-vben`）

```bash
cd apps/lina-vben
pnpm install
pnpm -F @lina/web-antd dev
```

关键规则：

- 表单使用`useVbenForm`，弹窗使用`useVbenModal`，抽屉使用`useVbenDrawer`。
- 表格页面使用`useVbenVxeGrid`与`Page`，操作按钮使用`ghost-button`和`Popconfirm`。
- 图标使用来自`@vben/icons`的`IconifyIcon`，图标名使用`Iconify`格式，例如`ant-design:inbox-outlined`。
- 路径别名`#/*`指向`./src/*`。
- 开发新页面前，先查看参考项目中的`UI`和交互模式。

## 插件开发

插件位于`apps/lina-plugins/<plugin-id>/`，并且必须包含：

```text
<plugin-id>/
  plugin.yaml         -> 插件清单
  plugin_embed.go     -> 嵌入资源入口
  backend/
    api/              -> 插件API DTO与路由定义
    plugin.go         -> 插件注册入口
    internal/
      controller/     -> HTTP控制器
      service/        -> 业务逻辑
      dao/            -> 需要数据库访问时生成的DAO
      model/          -> 生成模型
    hack/config.yaml  -> codegen配置
  frontend/pages/     -> 插件前端页面
  manifest/
    sql/              -> 安装SQL
    sql/uninstall/    -> 卸载SQL
    i18n/             -> 插件i18n资源
```

插件业务逻辑放在`backend/internal/service/`。只有实现宿主稳定扩展接缝的`provider`或`adapter`才放在`backend/provider/`。

# 提交规范

使用`Conventional Commits`：

```text
<type>(<scope>): <short summary>
```

| 类型 | 使用场景 |
|------|----------|
| `feat` | 新功能或新能力 |
| `fix` | 缺陷修复 |
| `docs` | 仅文档变更 |
| `refactor` | 不改变行为的代码重构 |
| `test` | 新增或更新测试 |
| `chore` | 构建脚本、CI、依赖等 |
| `perf` | 性能优化 |

示例：

```text
feat(plugin): add WASM sandbox memory limit configuration
fix(auth): prevent session token from persisting after force-logout
docs(contributing): add plugin workspace commands
```

摘要行保持在72个字符以内。扩展说明前保留一个空行。

# Pull Request 流程

1. `Fork`仓库，并从`main`创建分支：

   ```bash
   git checkout -b feat/my-feature
   ```

2. 非平凡变更遵循`OpenSpec`流程。在`Pull Request`描述中附上或引用相关`openspec/changes/<feature>/`产物。

3. 请求审查前确认所有检查通过：

   ```bash
   make test
   ```

4. 向`main`发起`Pull Request`，包含清晰标题、变更内容与原因、相关`Issue`或`OpenSpec`变更文档链接，以及`UI`变更的截图或录屏。

5. 合并前至少需要一名维护者批准。

# 发布标签

发布标签名称必须与`apps/lina-core/manifest/config/metadata.yaml`中的`framework.version`完全一致。维护者发布前应在本地检查目标版本：

```bash
make release.tag.check tag=v0.2.0
```

常规发布使用`Create Release Tag` `GitHub Actions`工作流。该工作流读取`framework.version`，通过`linactl release.tag.check`校验版本，拒绝移动已有标签，并创建匹配的`Git`标签。

仓库维护者应为`v*`等发布标签配置`GitHub`标签规则集，阻止普通用户直接创建、更新和删除标签，并仅允许受控发布主体绕过规则。

受控工作流会基于仓库变量`RELEASE_APP_CLIENT_ID`和仓库密钥`RELEASE_APP_PRIVATE_KEY`创建短期`GitHub App`安装令牌。该`GitHub App`必须安装到本仓库，并拥有`Contents`读写权限。

规则集绕过授权给主体，而不是授权给令牌字符串。请将受控工作流使用的`GitHub App`加入发布标签规则集的绕过列表。使用默认`GITHUB_TOKEN`推送标签时，通常无法可靠触发另一个工作流，因此受控工作流使用`GitHub App`令牌，让标签推送可以触发下游发布工作流。

# 代码规范

## Go

- 所有源码文件都必须在主文件中提供包级注释，或在其他文件中提供文件级注释。
- 代码中的时间长度使用`time.Duration`，配置中的时间长度使用带单位字符串，例如`"10s"`和`"5m"`。
- 多步数据库操作使用`dao.Xxx.Transaction()`闭包。
- 使用数据库无关的`ORM`构造。不要使用`FIND_IN_SET`、`GROUP_CONCAT`或`ANY(ARRAY[...])`等数据库特定函数。
- `DAO`、`DO`和`Entity`文件由`gf gen dao`生成，禁止手动修改。
- 控制器文件由`gf gen ctrl`生成骨架，只填写业务逻辑，不要修改生成骨架。

## TypeScript / Vue 3

- 遵循项目的`ESLint`和`Prettier`配置。
- 保持组件文件聚焦：一个文件一个组件。
- 使用`TypeScript`严格模式，避免使用`any`。
- `API`调用放在`src/api/`中，并使用项目的`requestClient`。
- `i18n`键遵循`module.subkey`约定；禁止硬编码用户可见文本。

# 测试

## E2E 测试（`Playwright`）

所有用户可观察行为变化都必须由`E2E`测试覆盖。

首次运行前安装前端依赖和 Playwright 浏览器：

```bash
make dev.setup
```

```bash
cd hack/tests
pnpm test              # 无头模式运行完整测试套件
pnpm test:headed       # 带可见浏览器运行
pnpm test:ui           # 打开交互式Playwright UI
pnpm report            # 查看HTML报告
```

测试文件命名为`TC{NNNN}-<description>.ts`，例如`TC0042-user-batch-export.ts`，并按相关模块放在`hack/tests/e2e/`下。

每个测试文件应覆盖其功能的完整操作集合，例如列表、新增、编辑、删除和特殊动作。测试过程中产生的截图、下载等文件放在项目根目录的`temp/`目录下。

# i18n 指南

`LinaPro`支持多语言。每个贡献都必须评估`i18n`影响：

- 新增用户可见文本必须在所有支持语言文件中提供对应条目。
- 前端运行时翻译位于`apps/lina-vben/`语言包。
- 宿主`API`文档翻译位于`apps/lina-core/manifest/i18n/<locale>/apidoc/`。
- 插件翻译自包含于`apps/lina-plugins/<plugin-id>/manifest/i18n/`。
- 禁止在`Go`或`TypeScript`源码中硬编码用户可见文本，必须使用`i18n`键。
- `en-US` `apidoc`文件只保留空对象占位；英文源文本来自`API` DTO本身。

如果某个变更不影响`i18n`资源，请在`Pull Request`描述中明确说明。

# 社区

| 渠道 | 链接 |
|------|------|
| `GitHub Issues` | https://github.com/linaproai/linapro/issues |
| 官方网站 | https://linapro.ai/ |
| 在线演示 | http://demo.linapro.ai/ |

提交新`Issue`前请先搜索已有问题。安全漏洞请不要公开提交`Issue`，请直接联系维护者。
