# Design

## Go Static Lint Gate

仓库通过根目录`.golangci.yml`、`.golangci-lint-version`和`.staticcheck-version`固定静态检查规则与工具版本。`linactl lint.go`是唯一跨平台扫描入口：省略`plugins`时按官方插件工作区自动探测；`plugins=0`扫描宿主与`linactl`；`plugins=1`生成临时完整`go.work`覆盖官方插件模块。根`Makefile`与`make.cmd`只做薄转发。

死代码门禁不启用独立`unused` linter，而由固定版本`staticcheck U1000`承担；含`wasip1`/`!wasip1`构建约束的包按宿主目标与`GOOS=wasip1 GOARCH=wasm` guest 目标归并结果，避免 guest 专属符号被宿主目标误判。`lint.go`与`env.setup`在扫描或初始化前按锁定版本自动检测/安装`golangci-lint`与`staticcheck`，安装使用`GOWORK=off`并剥离可能污染的构建变量。`CI`复用`make lint.go`/`linactl lint.go`路径，主验证与发布验证均阻断宿主与插件完整模式失败，不使用`only-new-issues`作为长期豁免。

动态插件 builder 配置与静态检查治理同期收敛：`wasm.hooks`、`wasm.resources`与`wasm.lifecycle.timeouts`统一放在插件根`hack/config.yaml`，时长必须使用带单位字符串；构建工具与宿主本地目录加载均不得再扫描`backend/*/*.yaml`。

公开参数只接受标准契约：`plugins`显式值仅标准布尔，省略时按工作区探测；参数 key 仅 kebab-case；布尔仅`true/false/1/0`；`wasm`单插件仅`dir=`，拒绝`p`/`plugin-dir`；发布标签仅显式`tag=`；镜像 registry 仅配置或`registry=`，不读`LINAPRO_IMAGE_REGISTRY`；构建变量展开仅`TARGET_DIR`/`BUILD_DIR`/`REPO_ROOT`，定向构建缺`hack/config.yaml`时拒绝且不回退`package.json`。

## Command Entry And Tool Consolidation

`linactl`是仓库默认跨平台开发命令承载者。所有常用命令使用 Go 标准库处理路径、进程、HTTP readiness、端口、文件复制、PID、日志和参数解析。`Makefile`继续服务 Linux/macOS 与既有 CI，Windows`make.cmd`服务`cmd.exe`和 PowerShell，但两者只转发到同一套`linactl`实现。命令保留 make-style`key=value`参数，避免迁移破坏既有开发习惯。

工具实现从独立 Go 模块收敛到`hack/tools/linactl/internal/`。镜像构建、动态插件 Wasm 打包、运行时 i18n 扫描、GoFrame controller/DAO 生成分别由内部子组件承载；公开命令仍保持`make image`、`make image.build`、`make wasm`、`make i18n.check`、`make ctrl`、`make dao`等稳定入口。GoFrame CLI 通过仓库锁定的 module 版本嵌入到隐藏隔离入口执行，避免开发者本地`gf`版本或`latest`二进制漂移。

`env.check`只做轻量工具级健康检查，不启动服务、不修改依赖；`env.setup`承接原`dev.setup`语义，用于前端依赖和 Playwright Chromium headless shell。`linactl dev`只等待自己管理的旧进程端口释放，仍拒绝杀死未知外部端口占用。

## GoFrame Code Generation With Target Directory Support

`linactl ctrl`和`linactl dao`只接受`dir=`目标参数，删除旧`p=`、`plugin=`和`target=`参数。不传目标时默认宿主`apps/lina-core`。

**插件代码生成配置迁移决策**：将官方插件的 GoFrame 代码生成配置从`backend/hack/config.yaml`迁移到插件根`hack/config.yaml`。`linactl`内部使用`Target{WorkDir, ConfigDir}`表示代码生成目标，插件后端目标为`WorkDir=apps/lina-plugins/<plugin-id>/backend`、`ConfigDir=apps/lina-plugins/<plugin-id>/hack`。标准插件识别只作用于`dir=<plugin>/backend`：当目标目录名为`backend`、父目录存在`plugin.yaml`、父目录位于`apps/lina-plugins/`下且`plugin.yaml`中的`id`与目录名一致时，目标被识别为标准插件后端并读取插件根`hack/`。

**插件自定义构建指令迁移决策**：`linactl build`改为读取插件根`hack/config.yaml`中的`build.commands`，支持`$(PLUGIN_ROOT)`和`$(REPO_ROOT)`变量展开。删除`apps/lina-plugins`根`go.mod`、`go.sum`和`lina-plugins.go`，改为由`linactl`在插件完整构建时自动生成源码插件聚合模块和临时`go.work`。

公开命令负责解析目标目录，隐藏`__goframe`子命令仍只接受`gen ctrl`或`gen dao`，并在父进程指定的工作目录中运行内嵌 GoFrame CLI。`dao`会提前校验目标`hack/config.yaml`，`ctrl`只要求目标后端目录存在，避免把 DAO 配置前置条件错误套到只生成 controller 的插件上。`plugins.check`扫描插件根`hack/config.yaml`，并阻断旧路径`backend/hack/config.yaml`。

根`Makefile`通过`hack/makefiles/database.mk`暴露统一`ctrl`和`dao`目标。`apps/lina-core/Makefile`删除旧`hack/hack-cli.mk`、`hack/hack.mk`依赖和旧`gf`安装/镜像重复入口，只保留宿主相关薄转发。

新增根目录共享片段`hack/makefiles/plugin.codegen.mk`，统一维护插件`ctrl`和`dao`目标。所有官方插件根目录`Makefile`改为计算自身`PLUGIN_ROOT`和`REPO_ROOT`后 include 共享片段，通过`$(PLUGIN_ROOT)/backend`推导目标后端目录，不再硬编码具体插件 ID 或`apps/lina-plugins/<plugin-id>/backend`。

## GoFrame Shutdown Config Reuse

**决策**：移除 LinaPro 自定义`shutdown.timeout`配置入口和`config.Service.GetShutdown`读取契约，改用 GoFrame Server 的`server.gracefulShutdownTimeout`作为停机超时唯一来源。宿主运行时资源清理继续在 GoFrame`Server.Run()`返回后执行，清理 deadline 从`Server.GetGracefulShutdownTimeout()`派生。

**关键实现**：
- 删除`config.Service.GetShutdown`接口及其实现
- `httpstartup`模块从 GoFrame Server 实例读取停机超时，不再依赖自定义配置键
- 配置样例将停机超时迁移到`server.gracefulShutdownTimeout`，移除顶层`shutdown`配置段
- 不新增 HTTP API、前端页面或数据库变更

## Agent Resource Bridges

Agent 资源桥接从单一`skills`命令扩展为`agents.<resource>.<action>`命令树，管理`skills`、`prompts`和`md`三类资源。实现分为`internal/agents/common`公共状态机与各资源专属注册表；所有 link/unlink 只操作仓库根内路径，不接受外部 root 参数，不修改`HOME`或系统全局路径。

软链使用`os.Symlink`、`os.Readlink`、`os.Lstat`、`os.Remove`、`filepath.Rel`等 Go 标准库实现。`force=1`只允许重建错误软链，绝不删除真实文件或目录。`agents`聚合入口在 TTY 中提供资源、动作、Agent 三层菜单，日常显示人类可读名称和简化汇总，详细路径状态留给资源子命令。

`agent-skills-link-cli`旧命令的完整行为已并入`agents-multi-resource`，当前主规范保留同等状态机约束；本分组不再保存与主规范完全相同的历史`spec.md`副本。

## OpenSpec Automation

月度 OpenSpec workflow 先确定是否存在已完成活跃变更，完成候选必须通过确定性 preflight；没有完成变更时跳过 AI 工具执行。自动归档和归档聚合都在所选 AI Coding 工具运行时内执行，工具由`AI_CODING_TOOL`变量选择，支持`codex`、`cc`和`copilot`，默认`codex`。

Codex、Claude Code 和 GitHub Copilot CLI 的运行环境隔离在各自 reusable workflow 中，主 workflow 只负责候选检测和路由。所有工具共享`.github/prompts/`下的自动归档和聚合提示词；凭据、模型和 base URL 仅在运行时由 secrets/variables 注入临时 AI home，真实密钥和 endpoint 不进入仓库工作区。

流程按阶段 fail-fast：preflight、auto-archive、completed-change assertion、OpenSpec 校验、archive consolidation、临时变更清理、变更范围保护、提交、推送和 PR 创建任一步失败都停止后续动作。归档结果通过固定维护分支创建或更新 PR，不直接写默认分支，且只允许`openspec/**`变更进入自动 PR。

## Release And Image Governance

release 发布链路以`framework.version`为唯一 tag 基线。`linactl release.tag.check`读取`apps/lina-core/manifest/config/metadata.yaml`并校验 tag 完全一致、版本非空、格式符合 Docker tag 兼容的 SemVer 子集。tag push workflow 在测试和镜像发布前执行同一命令，受控 release tag workflow 也在创建 tag 前复用同一校验，并通过 GitHub App installation token 配合 tag ruleset actor bypass。

`Release Test and Build`复用共享测试验证套件，使用 Main CI 的 brief scope，不运行完整 E2E。release 镜像发布依赖 tag 校验和共享测试全部成功，成功后发布 GHCR 多架构镜像和`latest`，再创建 GitHub Release。manual nightly image build 是明确的维护重发入口，直接调用镜像发布 workflow，不等待测试门禁；定时 nightly 继续保留完整测试门禁。

镜像构建使用每个平台对应的宿主二进制，多平台 buildx 构建必须`push=1`。demo Compose 使用 PostgreSQL 和 tmpfs 提供内存态体验环境，运行时配置独立放在`hack/deploy/config.yaml`并只读注入；测试 Compose 只提供手工开发容器，不自动初始化或启动业务服务。

## Upgrade And Installation

`make upgrade`是开发工具治理中的历史升级入口，用于区分框架升级与源码插件升级的文件发现、命令确认和开发期操作边界。后续插件有效版本、发现版本、运行时升级、发布切换、依赖校验、治理资源同步和失败诊断统一收敛到插件运行时升级模型；本分组只保留命令入口和开发工具侧的演进原因。

安装脚本位于`hack/scripts/install/`，Unix 使用`install.sh`，Windows PowerShell 使用`install.ps1`。安装默认下载源码 archive，不要求本地 Git；部署先解压到临时目录再移动到目标目录，非空目标默认拒绝覆盖。安装完成后输出 Go、Node.js、pnpm、MySQL、make 等环境健康检查和后续命令建议，不自动安装系统依赖。

## Performance Audit Skill

`lina-perf-audit`是手动触发技能，禁止由 CI、定时任务、git hook 或其他技能自动调用。完整审计分三阶段：Stage 0 重置本地环境、安装并启用内建插件、写入 audit-only stress fixture、扫描 endpoints 并准备日志；Stage 1 用 sub-agent 并发审计 endpoint；Stage 2 聚合`SUMMARY.md`、`meta.json`和持久 issue cards。

审计通过 GoFrame 默认`Trace-ID`关联请求与 SQL 日志，不添加生产中间件。破坏性 endpoint 使用自主 create-call-delete fixture，无匹配 create 时跳过并要求人工跟进。读取请求出现非预期写 SQL、N+1、缺索引、无分页、重复查询等问题按严重度写入`perf-issues/`，通过 fingerprint 跨运行去重；issue-card 正文为中文，机器字段保持原文。

## Cross-Domain Impacts

- `project-setup`承载项目初始化、运行时数据库、开发工作台入口和 E2E base path 当前契约；当前契约由`openspec/specs/project-setup/spec.md`承载，历史 owner 为`archive/foundation`，本分组只保留开发命令和环境入口影响。
- `database-bootstrap-commands`承载 init/mock 确认、SQL asset source 和 statement-by-statement 执行；当前契约由`openspec/specs/database-bootstrap-commands/spec.md`承载，历史 owner 为`archive/database-engine`。
- `cron-job-management`承载内置日志清理任务投射、任务权限和调度业务规则；当前契约由`openspec/specs/cron-job-management/spec.md`承载，历史 owner 为`archive/scheduled-jobs`。
- `e2e-suite-organization`承载 Playwright 测试目录、TC 编号、host-only/plugin-full 和 nightly E2E 治理；当前契约由`openspec/specs/e2e-suite-organization/spec.md`承载，历史 owner 为`archive/e2e-testing`。
- `release-image-build`同时影响 CI、workflow、image builder 和 demo Compose；本分组保留发布工具链历史 owner，但运行时容器行为以当前主规范为准。
- `plugin-upgrade-governance`承载源码插件与动态插件的有效版本隔离、运行时升级、发布切换和失败诊断；当前契约由`openspec/specs/plugin-upgrade-governance/spec.md`承载，历史 owner 为`archive/plugin-framework`。本分组只保留开发期命令入口对该能力的历史影响。
- `spec-governance`同时涉及 OpenSpec 流程与项目规范规则加载；本分组保留月度归档自动化影响，主规范仍由`openspec/specs/spec-governance/spec.md`承载。

## Governance Notes

本分组历史实现修改过开发工具、脚本、workflow、README、Agent skill 和 OpenSpec 文档，按开发工具和文档规则完成过跨平台影响记录。运行时影响均应在对应 owner 变更中审查：本归档压缩不新增 Go、API、SQL、前端、插件或运行时 i18n 变更。测试策略以工具单测、命令 smoke、workflow 语法、shell 语法、OpenSpec 校验、diff 检查和手工触发 dry-run 为主，E2E 仅在具体用户可观察行为变更的 owner 任务中维护。

反馈闭环处理过 FB-1（`make dao`/`make ctrl`只能在宿主或插件后端目录下正确生成）、FB-2（插件根目录 Makefile 硬编码插件后端路径）和 FB-3（Redis cluster smoke 脚本仍调用已移除的`make init`目标）。FB-1 根因为`linactl`的 GoFrame CLI 入口固定使用`apps/lina-core`作为项目目录，修复为支持显式目标目录参数；FB-2 根因为插件 Makefile 硬编码`dir=apps/lina-plugins/<plugin-id>/backend`，修复为引入共享`plugin.codegen.mk`片段；FB-3 根因为 smoke 脚本未同步`db.init`目标重命名，修复为更新命令名称。三项反馈均不修改运行时代码、HTTP API、SQL、前端 UI 或插件运行时契约。
