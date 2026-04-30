## Why

`LinaPro` 当前的安装与升级体验是开发态遗留实现:安装层需要用户手工 `git clone` + `make init`,跨平台规范停留在 `install.sh` + `install.ps1` 双脚本;升级层依赖 `make upgrade` + `hack/tools/upgrade-source/` 这套独立 Go 工具,需要操作者手工传 `confirm=upgrade scope=framework target=...` 等参数,完全不具备"AI 主导执行,异常才转人工"的体验。这两条路径都不符合 `LinaPro` 作为 **AI 原生全栈框架** 的定位。

本次变更把安装路径收敛为单条 `curl -fsSL https://linapro.ai/install.sh | bash` 命令,把升级路径完全交给 `Claude Code` 的 `lina-upgrade` 技能驱动,让"安装"和"升级"成为可以被 AI 工具自动化执行、可被 LinaPro 用户在终端里一行触达的标准流程。

## What Changes

### 安装路径

- **新增** 单一安装入口 `curl -fsSL https://linapro.ai/install.sh | bash`,该 URL 对应仓库内的 `hack/scripts/install/bootstrap.sh`
- **新增** `bootstrap.sh` 作为完全自包含的单文件入口,只做"探测 OS + git clone + dispatch",不再 curl 二级脚本
- **新增** 三平台 `bash` 实现:`install-macos.sh` / `install-linux.sh` / `install-windows.sh`,共享 `lib/_common.sh` 公共函数
- **新增** `checks/prereq.sh` 统一前置工具探测(`go` / `node` / `pnpm` / `git` / `mysql`)
- **新增** 环境变量覆盖:`LINAPRO_VERSION` / `LINAPRO_DIR` / `LINAPRO_NON_INTERACTIVE` / `LINAPRO_SKIP_MOCK` / `LINAPRO_SHALLOW`
- **新增** `.gitattributes` 强制所有 `*.sh` 与 `bootstrap.sh` 使用 LF 行尾(防 Windows CRLF 污染)
- **BREAKING** 移除 `install.sh` + `install.ps1` 双脚本规范,Windows 用户必须在 `Git Bash` 或 `WSL` 终端中执行
- **BREAKING** 移除"基于 source archive 下载,不依赖 git clone"的旧约定,改为 `git clone --branch <tag>` 默认 full clone(保留历史以支持后续 `lina-upgrade` 技能)
- **BREAKING** 移除"无法解析稳定 tag 时回落 main 分支"的兜底策略,改为硬失败 + 提示用户显式设置 `LINAPRO_VERSION`

### 升级路径

- **新增** `.claude/skills/lina-upgrade/` 作为升级技能,由 `Claude Code` / `Codex` 等 AI 工具调用
- **新增** 四层 baseline 校验脚本(存在性 / 可达性 / 身份对照 / 汇总),失败由 AI 与用户对话修正,不强阻塞
- **新增** Tier 1/2/3 文件分类规则,自动按 Tier 选择冲突解决策略
- **新增** 升级转人工的硬规则文档(`escalation-rules.md`)
- **新增** 升级技能脚本族:`upgrade-baseline-check.sh` / `upgrade-plan.sh` / `upgrade-classify.sh` / `upgrade-regenerate.sh` / `upgrade-verify.sh`
- **BREAKING** 移除 `make upgrade` 整套实现及其 Go 工具 `hack/tools/upgrade-source/`
- **BREAKING** 移除 `apps/lina-core/hack/config.yaml` 中 `frameworkUpgrade.version` 作为升级基线的约定,统一改读 `apps/lina-core/manifest/config/metadata.yaml.framework.version`
- **BREAKING** 移除"框架升级时从第一个 SQL 文件全量重放"的语义,改为按编号递增执行新增 SQL(增量迁移)
- **BREAKING** 源码插件升级入口由 `make upgrade scope=source-plugin plugin=<id>` 改为通过 `lina-upgrade` 技能子流程触发(技能内部分发到框架升级 / 源码插件升级两条子工作流)

### 文档与红线

- **新增** `apps/lina-core/manifest/config/metadata.yaml.framework.version` 不允许用户手工编辑的红线(由升级技能在校验阶段守门)
- **修改** `README.md` / `README.zh_CN.md` / `CLAUDE.md` 同步更新安装与升级章节,移除 `make upgrade` 引用

## Capabilities

### New Capabilities

(本次变更不引入全新能力,所有新行为都通过修改既有能力的需求条目落地)

### Modified Capabilities

- `framework-bootstrap-installer`: 把双脚本入口收敛为单一 `bash` 入口,改用 `git clone`,新增 `curl|bash` 托管入口,新增环境变量覆盖与默认稳定 tag 解析策略
- `source-upgrade-governance`: 把 `make upgrade` 入口替换为 `.claude/skills/lina-upgrade/` 技能,把基线读取源由 `hack/config.yaml.frameworkUpgrade.version` 改为 `metadata.yaml.framework.version`,把"全量 SQL 重放"改为"按编号增量迁移",新增 baseline 四层校验、Tier 分类、转人工规则
- `plugin-upgrade-governance`: 把源码插件升级的命令引用从 `make upgrade scope=source-plugin` 改为 `lina-upgrade` 技能子流程,宿主启动失败提示同步更新

## Impact

### 受影响代码

- `hack/scripts/install/`(全新目录,首次创建)
- `hack/tools/upgrade-source/`(整个目录删除,独立 Go module)
- `Makefile`(根目录,删除 `upgrade` target 与相关依赖)
- `apps/lina-core/Makefile`(删除 `upgrade` 代理 target)
- `apps/lina-core/hack/config.yaml`(`frameworkUpgrade` 区块语义降级,仅保留 `repositoryUrl` 用于安装时的 origin 推断或彻底删除)
- `apps/lina-core/internal/cmd/`(若启动时引用 `frameworkUpgrade.version` 作为校验依据,改读 `metadata.yaml`)
- `.claude/skills/lina-upgrade/`(全新目录)
- `.gitattributes`(新增或更新,强制 `*.sh` 使用 LF)

### 受影响文档

- `README.md` / `README.zh_CN.md`(根目录与各子模块)
- `CLAUDE.md`(常用命令、开发流程章节)
- `.agents/instructions/` / `.agents/prompts/`(若有 `make upgrade` 引用)
- `apps/lina-core/manifest/config/metadata.yaml`(增加"该字段不允许手工编辑"的注释红线)

### 受影响外部资源

- `https://linapro.ai/install.sh`(CDN 静态托管,内容 = `hack/scripts/install/bootstrap.sh`,部署流程作为附属任务)
- `https://github.com/linaproai/linapro/releases/`(必须保证存在合规的 stable tag,以便 GitHub `releases/latest` 重定向能解析到具体版本)

### i18n 评估

本变更**不影响**前端运行时语言包、宿主/插件运行时 `manifest/i18n` 资源以及 `apidoc i18n JSON`。仅安装脚本与 `README` 涉及英文/中文文案,按项目"目录级主说明文档统一英文 + 中文镜像"规范同步维护即可。

### 兼容性与升级路径

- 当前项目处于全新阶段,无需保留 `make upgrade` 的向后兼容别名,直接断更
- 既有的 `openspec/specs/source-upgrade-governance/spec.md` 中关于 `make upgrade` 与 `frameworkUpgrade.version` 的需求条目通过本变更的 `MODIFIED Requirements` + `REMOVED Requirements` 显式淘汰
- 用户在升级到新版本后,既有的 `apps/lina-core/hack/config.yaml.frameworkUpgrade.version` 字段保留也无副作用,但 `lina-upgrade` 技能不再读它

### 风险

- `linapro.ai/install.sh` 的 CDN 部署流程不在本仓库 PR 范围内,需要在合并后执行运维侧动作,本变更通过附属任务文档化该流程,但 PR 自身不会自动生效远程 URL
- Windows `Git Bash` 行为与原生 `Linux` 仍有路径与权限位差异,本变更仅在 Git Bash / WSL 中标 stable,原生 PowerShell 用户必须切换终端
- 升级技能"无冲突全自动"的承诺取决于 `Tier 1` 公共契约的稳定性,若框架未来出现破坏性 API 变更,需通过单独的 OpenSpec 变更明确升级人工动作
