# Design

## Go Static Lint Gate

仓库通过根目录`.golangci.yml`、`.golangci-lint-version`和`.staticcheck-version`固定静态检查规则与工具版本。`linactl lint.go`是唯一跨平台扫描入口：省略`plugins`时按官方插件工作区自动探测；`plugins=0`扫描宿主与`linactl`；`plugins=1`生成临时完整`go.work`覆盖官方插件模块。根`Makefile`与`make.cmd`只做薄转发。

死代码门禁不启用独立`unused` linter，而由固定版本`staticcheck U1000`承担；含`wasip1`/`!wasip1`构建约束的包按宿主目标与`GOOS=wasip1 GOARCH=wasm` guest 目标归并结果，避免 guest 专属符号被宿主目标误判。`lint.go`与`env.setup`在扫描或初始化前按锁定版本自动检测/安装`golangci-lint`与`staticcheck`，安装使用`GOWORK=off`并剥离可能污染的构建变量。`CI`复用`make lint.go`/`linactl lint.go`路径，主验证与发布验证均阻断宿主与插件完整模式失败，不使用`only-new-issues`作为长期豁免。

**可选`dir=`定向粒度 = 单个 Go module，而非包。** 本地与 Agent 迭代修改单个后端组件时，全量工作区扫描反馈周期过长；`build`/`ctrl`/`dao`已支持`dir=`定向，lint 对齐同一语义。实现要点：

1. 先按`plugins`准备宿主或官方插件完整工作区，再在`go list -m`结果中过滤`dir`对应 module；插件 module 依赖`temp/go.work.plugins`，跳过环境准备会导致插件定向解析失败。
2. 路径解析：相对仓库根转绝对路径 → 必须存在且为目录 → 含`plugin.yaml`且存在`backend/go.mod`时优先`backend` → 否则向上查找最近`go.mod`（止于仓库根）→ 与 workspace module 列表做路径等价匹配（`EvalSymlinks`）→ 匹配失败则明确报错，不静默回退全量。
3. 日志 plan/summary 标识`scope=workspace|dir`与目标相对路径，避免审查误读为已跑全量。
4. 未传`dir`时行为与`CI`兼容、完全不变；首期仅单路径，不允许多个`dir`一次传入。
5. 明确不做包级`packages=`筛选（`U1000`跨包语义不可靠）、不做基于 git diff 的自动范围推断。

公开入口示例：`make lint dir=apps/lina-core`、`make lint.go dir=apps/lina-plugins/<id>/backend plugins=1`、`linactl lint.go dir=hack/tools/linactl plugins=0`。`lint.mk`只透传`dir`，业务逻辑仅在`linactl`。

风险：定向 lint 通过不代表其他 module 无问题 → 文档与规则写明`CI`/PR 仍跑全量，Agent 约定变更涉及的全部 module 至少各跑一次；插件根误解析到 monorepo 上级`go.mod` → 插件根优先`backend`且查找止于 repo root。

动态插件 builder 配置与静态检查治理同期收敛：`wasm.hooks`、`wasm.resources`与`wasm.lifecycle.timeouts`统一放在插件根`hack/config.yaml`，时长必须使用带单位字符串；构建工具与宿主本地目录加载均不得再扫描`backend/*/*.yaml`。

公开参数只接受标准契约：`plugins`显式值仅标准布尔，省略时按工作区探测；参数 key 仅 kebab-case；布尔仅`true/false/1/0`；`wasm`单插件仅`dir=`，拒绝`p`/`plugin-dir`；发布标签仅显式`tag=`；镜像 registry 仅配置或`registry=`，不读`LINAPRO_IMAGE_REGISTRY`；构建变量展开仅`TARGET_DIR`/`BUILD_DIR`/`REPO_ROOT`，定向构建缺`hack/config.yaml`时拒绝且不回退`package.json`。

## Command Entry And Tool Consolidation

`linactl`是仓库默认跨平台开发命令承载者。所有常用命令使用 Go 标准库处理路径、进程、HTTP readiness、端口、文件复制、PID、日志和参数解析。`Makefile`继续服务 Linux/macOS 与既有 CI，Windows`make.cmd`服务`cmd.exe`和 PowerShell，但两者只转发到同一套`linactl`实现。命令保留 make-style`key=value`参数，避免迁移破坏既有开发习惯。

工具实现从独立 Go 模块收敛到`hack/tools/linactl/internal/`。镜像构建、动态插件 Wasm 打包、运行时 i18n 扫描、GoFrame controller/DAO 生成、框架源码升级分别由内部子组件或命令实现承载；公开命令仍保持`make image`、`make image.build`、`make wasm`、`make i18n.check`、`make ctrl`、`make dao`、`make upgrade`等稳定入口。GoFrame CLI 通过仓库锁定的 module 版本嵌入到隐藏隔离入口执行，避免开发者本地`gf`版本或`latest`二进制漂移。

`env.check`只做轻量工具级健康检查，不启动服务、不修改依赖；`env.setup`承接原`dev.setup`语义，用于前端依赖和 Playwright Chromium headless shell。`linactl dev`只等待自己管理的旧进程端口释放，仍拒绝杀死未知外部端口占用。

## Runtime i18n Check For Config Display Keys

参数设置页对`sys_config`行通过`config.<config_key>.name`与`config.<config_key>.remark`做运行时投影；插件首次`SetValue`落库时`name`常等于技术 key，缺译时管理页会裸显类 i18n key 字符串。既有`i18n.check`覆盖 locale 对等、`bizerr` messageKey、插件管理展示键与前端静态`$t`，但未约束 config 展示键。

**期望键静态收集（不读运行时库）**：

| 范围 | 收集方式 |
|------|----------|
| 宿主 | `apps/lina-core/manifest/sql/**/*.sql`中`sys_config` seed 写入的 key 字面量；宿主 Go 中`SysConfigKey`及明确的`sys.*`/`demo.*`受保护常量 |
| 插件 | 仅`plugin.yaml`中`i18n.enabled: true`的插件；扫描模块内`hostconfigcap.SysConfigKey = "..."`（及等价赋值）字面量 |

**强制键**：各运行时 locale 必须齐全`config.<key>.name`与`config.<key>.remark`；宿主键落在宿主`manifest/i18n/<locale>/`，插件键落在该插件`manifest/i18n/<locale>/`（运行时由宿主合并）。未启用 i18n 的插件跳过该项。

**集成**：在`runtimei18n`的`messages`子检查中追加 config 展示元数据校验，与 bizerr / plugin display metadata 并列，由既有`RunCheck`/`make i18n.check`一并执行。历史错误键名若仍存在于资源中，locale 对等仍要求各语言一致，但不会替代正确键；正确键缺失即失败。动态拼接的 key 无法静态收集，约定插件必须以常量声明`SysConfigKey`。规则文件`.agents/rules/i18n.md`与`linactl`中英文 README 同步该契约。

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

## Framework Source Upgrade

二次开发或 fork 后的业务仓库需要持续合并上游框架。最终开发期入口为`linactl upgrade`（根`Makefile`/`make.cmd`薄包装为`make upgrade`），业务逻辑全部在 Go 中实现，禁止平台专属 shell 承载合并逻辑。handler 使用`runFrameworkUpgrade`，与`db.upgrade`的`runUpgrade`隔离；公开命令名`upgrade`与`db.upgrade`并存。

**升级源**：硬编码官方仓库`https://github.com/linaproai/linapro.git`，工具托管 remote 名`linapro`（不存在则创建，URL 不正确则纠正）。不使用本地`origin`/fork，不接受`remote=`覆盖；测试可通过可覆盖的官方 URL 变量指向临时 bare 仓。

**目标 ref**：

| `v` 输入 | 行为 |
| --- | --- |
| 省略 / 空 | 从官方 tags 取最新稳定版（匹配`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`，规范化为`vMAJOR.MINOR.PATCH`；含`-`预发布后缀不参与默认选择） |
| 稳定版本模式 | 合并对应 tag（缺`v`前缀则补齐）；用户显式指定时可尝试含预发布的 tag |
| 其它非空（如`main`） | 合并`linapro/<name>`分支 |

**安全门禁**：detached HEAD 始终拒绝；默认要求工作区干净（`git status --porcelain`为空），`force=1`仅跳过脏检查。门禁失败不执行 merge。

**Git 操作序列**：确认`git`可用与仓库根 → 解析当前分支 → 安全门禁 → 确保官方 remote 并`git fetch linapro --tags --prune` → 解析目标 ref → 记录 pre-merge`HEAD` →`git merge --no-edit --no-commit <ref>` → 将`apps/lina-plugins`恢复为升级前状态（官方插件变更不进入结果；升级前不存在则丢弃官方引入的该路径）→ 若剔除插件后无差异则`merge --abort`并报告无宿主变更 → 否则`git commit --no-edit`完成合并并提示插件需`plugins.update`单独更新。不自动处理冲突、不 hard reset、不 rebase、不 force-push；不自动执行`db.upgrade`、依赖 tidy 或服务重启。

历史`upgrade-governance`中带`scope=framework|source-plugin`、基于`hack/config.yaml`的`frameworkUpgrade`元数据与宿主 SQL 全量重放的源码升级路径，已由上述官方 remote + 稳定 tag 合并模型取代为当前开发期框架升级入口；插件运行时升级、有效版本隔离与发布切换仍归`plugin-upgrade-governance`。

## Installation

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
- `plugin-upgrade-governance`承载源码插件与动态插件的有效版本隔离、运行时升级、发布切换和失败诊断；当前契约由`openspec/specs/plugin-upgrade-governance/spec.md`承载，历史 owner 为`archive/plugin-framework`。本分组只保留开发期`upgrade`命令与插件路径隔离边界。
- `config-management`承载参数设置投影与 config 展示键本地化；本分组保留`i18n.check`门禁与展示键契约交叉影响，完整配置管理行为以主规范为准。
- `spec-governance`同时涉及 OpenSpec 流程与项目规范规则加载；本分组保留月度归档自动化影响，主规范仍由`openspec/specs/spec-governance/spec.md`承载。

## Governance Notes

本分组覆盖开发工具、脚本、workflow、README、Agent skill、OpenSpec 文档、框架源码升级命令与运行时 i18n 扫描门禁。跨平台入口以`linactl`/Go 标准库为准，`Makefile`/`make.cmd`仅透传。测试策略以工具单测、临时 Git 仓库集成路径、命令 smoke、`make i18n.check`、workflow/shell 语法、OpenSpec 校验和手工 dry-run 为主；E2E 仅在用户可观察页面行为变更的 owner 中维护。

关键反馈闭环：代码生成目标目录与插件 Makefile 硬编码路径、Redis cluster smoke 旧`init`目标、静态检查首批噪声与 wasip1 死代码归并、builder 配置迁入`hack/config.yaml`、`linactl`兼容面删除、升级源固定官方 remote 且合并不更新`lina-plugins`、参数设置页 config 展示键缺译导致类 i18n key 裸显（扩展`i18n.check`并补齐宿主/插件资源）。
