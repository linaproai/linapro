## 1. 准备工作

- [x] 1.1 阅读 `proposal.md` / `design.md` / `specs/**` 确认对全部决策达成一致
- [x] 1.2 在仓库根创建工作分支(若使用 worktree 则按 `git-worktree` 技能流程执行)
- [x] 1.3 通过 `grep -rn "make upgrade"` / `grep -rn "frameworkUpgrade"` / `grep -rn "upgrade-source"` 三条命令快照当前所有引用,作为后续清理的核对清单
- [ ] 1.4 确认当前 `apps/lina-core/manifest/config/metadata.yaml.framework.version` 与最近一次 git tag 是否一致;若不一致先与用户对齐期望基线再开始
  - 2026-04-30 执行记录:本地、`origin` 与 `https://github.com/linaproai/linapro.git` 均未发现 `v*` tag,当前 `metadata.yaml.framework.version` 为 `v0.5.0`;实现继续推进,发布前仍需由维护者确认并发布对应 tag。

## 2. .gitattributes 与项目根配置

- [x] 2.1 检查仓库根是否存在 `.gitattributes`;不存在则新建
- [x] 2.2 在 `.gitattributes` 中追加以下规则,并保证已有规则未被破坏:
  ```
  *.sh text eol=lf
  *.bash text eol=lf
  hack/scripts/install/bootstrap.sh text eol=lf
  ```
- [x] 2.3 在新工作树上跑 `git check-attr eol -- hack/scripts/install/bootstrap.sh` 验证规则生效

## 3. 安装脚本目录结构搭建

- [x] 3.1 创建目录 `hack/scripts/install/`、`hack/scripts/install/lib/`、`hack/scripts/install/checks/`(空目录用 `.gitkeep` 占位避免被 git 忽略)
- [x] 3.2 创建 `hack/scripts/install/README.md`(英文),内容包括:
  - 安装入口命令 `curl -fsSL https://linapro.ai/install.sh | bash`
  - 三个支持平台说明(macOS / Linux / Windows-via-Git-Bash-or-WSL)
  - 五个环境变量 `LINAPRO_VERSION` / `LINAPRO_DIR` / `LINAPRO_NON_INTERACTIVE` / `LINAPRO_SKIP_MOCK` / `LINAPRO_SHALLOW` 的语义、默认值、用例
  - Windows 用户必须使用 Git Bash 或 WSL 的明确说明
  - 本地执行等价命令 `bash hack/scripts/install/bootstrap.sh`
  - 失败时的诊断与重试建议
- [x] 3.3 创建 `hack/scripts/install/README.zh_CN.md`(中文镜像),内容与 3.2 完全对应

## 4. 公共库 `lib/_common.sh`

- [x] 4.1 创建 `hack/scripts/install/lib/_common.sh`,包含以下函数(均要有英文注释):
  - `log_info` / `log_warn` / `log_error` / `log_debug`(统一颜色与前缀,用 `tput` 或 ANSI escape)
  - `die "<message>"`(打印错误到 stderr 并 `exit 1`)
  - `version_ge "<a>" "<b>"`(语义化版本比较,返回 0 表示 a >= b)
  - `require_command "<cmd>" "<install-hint>"`(检测命令存在,缺失时打印安装提示并 die)
  - `retry "<times>" "<delay>" -- <cmd...>`(简单重试逻辑,用于网络抖动场景)
  - `confirm "<message>"`(交互式 y/N 确认;`LINAPRO_NON_INTERACTIVE=1` 时直接返回 0)
  - `detect_os`(返回 `macos` / `linux` / `windows` / `unsupported`)
  - `is_port_in_use "<port>"`(检测端口占用,跨平台兼容)
- [x] 4.2 在文件顶部添加文件用途注释(描述该库被三个平台脚本共享)
- [x] 4.3 用 shellcheck 校验 `lib/_common.sh` 没有 SC2086 / SC2155 等高危告警(要求作为开发态约定,可在 README 中说明)
  - 2026-04-30 执行记录:本机未安装 `shellcheck`;已通过 Docker 镜像 `koalaman/shellcheck:stable` 运行 `shellcheck -x`,并通过 `bash -n hack/scripts/install/lib/_common.sh`。

## 5. 前置探测脚本 `checks/prereq.sh`

- [x] 5.1 创建 `hack/scripts/install/checks/prereq.sh`,文件顶部加用途注释
- [x] 5.2 source 公共库 `_common.sh`,用 `log_*` 输出
- [x] 5.3 实现以下检测,每项失败打印缺失原因与平台对应安装提示链接:
  - `go` 版本 ≥ 1.22(用 `version_ge`)
  - `node` 版本 ≥ 20
  - `pnpm` 版本 ≥ 8(若缺失提示 `npm i -g pnpm`)
  - `git` 在 PATH 中
  - `make` 在 PATH 中(Git Bash 缺失时给出 mingw32-make 或 chocolatey 提示)
  - MySQL 客户端可达(`mysql --version` 通过 即可,深入连接放到 `make init`)
- [x] 5.4 探测 TCP 端口 5666 / 8080 是否占用,占用时打印 warning 但不 die
- [x] 5.5 退出码 0 = 全部通过、1 = 任一关键工具缺失、2 = 仅警告(端口占用 / MySQL 未连接)

## 6. 单文件 bootstrap `bootstrap.sh`(curl|bash 入口)

- [x] 6.1 创建 `hack/scripts/install/bootstrap.sh`,文件顶部加文件用途说明并标注"本文件 = `linapro.ai/install.sh` 部署内容"
- [x] 6.2 文件首行用 `#!/usr/bin/env bash`,紧接 `set -euo pipefail`
- [x] 6.3 注册 `trap 'log_error "..." ' ERR EXIT` 错误捕获(注意:bootstrap 因为可能通过 curl|bash 调用,不能 source `_common.sh`(尚未克隆),需自己内联最简日志函数)
- [x] 6.4 实现 `detect_os` 内联版本(只用 `uname -s`)
- [x] 6.5 实现版本解析逻辑:
  - 优先读 `LINAPRO_VERSION` 环境变量
  - 否则 `curl -sIL https://github.com/linaproai/linapro/releases/latest`,从 `Location:` 头提取 `tag/<version>`
  - 解析失败时 die,错误消息包含 `LINAPRO_VERSION=v0.x.y` 显式覆盖示例
- [x] 6.6 实现目标目录解析:
  - 读 `LINAPRO_DIR`,默认 `./linapro`
  - 若目录已存在且非空,die 并提示"set LINAPRO_FORCE=1 to overwrite"(`LINAPRO_FORCE` 作为隐藏开关,文档不强调)
- [x] 6.7 实现 `git clone`:
  - `LINAPRO_SHALLOW=1` 时附加 `--depth 1`,否则 full clone
  - `git clone --branch <tag> https://github.com/linaproai/linapro.git "$LINAPRO_DIR"`
  - clone 失败时 die,提示用户检查网络与 tag 是否存在
- [x] 6.8 实现 dispatch:
  - `cd "$LINAPRO_DIR"`
  - `exec bash hack/scripts/install/install-${OS}.sh "$@"`(把环境变量与位置参数都透传)
- [x] 6.9 在 bootstrap 顶部输出一个 banner,包含解析后的版本号、目标目录、平台,便于 curl|bash 用户快速看到当前进度
- [x] 6.10 用 shellcheck 校验
  - 2026-04-30 执行记录:本机未安装 `shellcheck`;已通过 Docker 镜像 `koalaman/shellcheck:stable` 运行 `shellcheck -x`,并通过 `bash -n hack/scripts/install/bootstrap.sh`。另外补充了 latest release 非版本 tag 拒绝逻辑、`LINAPRO_FORCE=1` 危险目标目录拒绝逻辑,并通过 `make test-install` 覆盖。

## 7. 平台脚本 `install-macos.sh`

- [x] 7.1 创建 `hack/scripts/install/install-macos.sh`,文件顶部加用途注释
- [x] 7.2 source `lib/_common.sh`(此时已在 git clone 后,允许相对路径)
- [x] 7.3 调用 `checks/prereq.sh`,根据退出码决定是否继续
- [x] 7.4 缺失工具时给出 macOS 专用提示(`brew install go node pnpm git make mysql`)
- [x] 7.5 `cd apps/lina-core && go mod download`
- [x] 7.6 `cd ../../apps/lina-vben && pnpm install`
- [x] 7.7 配置文件初始化:若 `apps/lina-core/manifest/config/config.yaml` 不存在,从 `config.template.yaml` 复制
- [x] 7.8 数据库初始化:`make init`(在仓库根目录执行)
- [x] 7.9 mock 数据加载:若 `LINAPRO_SKIP_MOCK` 未设置,执行 `make mock`
- [x] 7.10 端口探测:用 `_common.sh` 的 `is_port_in_use` 检查 5666 / 8080,占用时输出 warning
- [x] 7.11 输出成功 banner:项目目录、admin/admin123、`make dev` 启动命令
- [x] 7.12 用 shellcheck 校验
  - 2026-04-30 执行记录:本机未安装 `shellcheck`;已通过 Docker 镜像 `koalaman/shellcheck:stable` 运行 `shellcheck -x`,并通过 `bash -n hack/scripts/install/install-macos.sh`。

## 8. 平台脚本 `install-linux.sh`

- [x] 8.1 创建 `hack/scripts/install/install-linux.sh`,顶部加用途注释
- [x] 8.2 source `lib/_common.sh`,调用 `checks/prereq.sh`
- [x] 8.3 缺失工具时给出 Linux 提示(分发版常见命令,如 `apt-get install -y golang nodejs`,并提示对于不支持的发行版查看 README)
- [x] 8.4 复用 7.5 - 7.12 的整体流程,差异只在工具安装提示

## 9. 平台脚本 `install-windows.sh`(Git Bash / WSL)

- [x] 9.1 创建 `hack/scripts/install/install-windows.sh`,顶部加用途注释 + 明示"必须在 Git Bash 或 WSL 中运行"
- [x] 9.2 在脚本开头探测当前 shell 是否为 MSYS / MINGW(`uname -s` 二次检查),否则 die 并提示切到 Git Bash
- [x] 9.3 source `lib/_common.sh`,调用 `checks/prereq.sh`
- [x] 9.4 缺失工具时给出 Windows 提示(`winget install GoLang.Go`、`winget install OpenJS.NodeJS`、`scoop install pnpm`、`scoop install make` 等;同时提供 chocolatey 备选)
- [x] 9.5 处理路径风格:在调用 Windows 原生工具前用 `cygpath -w` 转换路径
- [x] 9.6 注意:Git Bash 下没有 `make` 时,允许跳过 `make init`,改为直接执行底层命令(在脚本中实现 `make_init_fallback` 函数,内部调 `go run` + `mysql -e` 等价操作);**默认仍尝试 `make init`,fallback 仅当 make 不存在时启用**
- [x] 9.7 复用 7.6 - 7.11 的总体流程
- [x] 9.8 用 shellcheck 校验(注意 MSYS 兼容性 lint 项)
  - 2026-04-30 执行记录:本机未安装 `shellcheck`;已通过 Docker 镜像 `koalaman/shellcheck:stable` 运行 `shellcheck -x`,并通过 `bash -n hack/scripts/install/install-windows.sh`。

## 10. lina-upgrade 技能目录结构

- [x] 10.1 创建目录 `.claude/skills/lina-upgrade/`、`.claude/skills/lina-upgrade/references/`、`.claude/skills/lina-upgrade/scripts/`
- [x] 10.2 创建 `.claude/skills/lina-upgrade/SKILL.md`,frontmatter 格式参考 Claude Code 官方文档:
  ```
  ---
  name: lina-upgrade
  description: <英文一句话描述,触发场景关键词必须含 "upgrade", "linapro", "framework", "plugin">
  ---
  ```
- [x] 10.3 SKILL.md 正文必须按以下结构组织(英文撰写):
  - **Purpose**:技能的目标与边界
  - **When to invoke**:用户触发关键词举例(中英文都覆盖)
  - **Inputs the AI must collect from the user**:目标版本、scope(framework / source-plugin)、源码插件 ID(若 scope=source-plugin)
  - **Workflow (10 steps)**:严格按 design.md D9 列出 10 步,每步注明调用的脚本、可能的失败模式、转人工的判定
  - **Outputs the AI must produce**:升级计划、最终报告、人工动作清单
  - **References**:链接到 references/ 下的 4 份文档
  - **Allowed tools**:Bash / Read / Edit / Grep / Glob

## 11. 升级技能 references 文档族

- [x] 11.1 创建 `.claude/skills/lina-upgrade/references/tier-classification.md`(英文),内容包括:
  - Tier 1 / 2 / 3 路径模式表(同 specs/source-upgrade-governance 中的表格)
  - 每条路径的归类理由
  - 未匹配路径的默认 fallback(默认归 Tier 2)
  - `upgrade-classify.sh` 与本文档的对应关系
- [x] 11.2 创建 `.claude/skills/lina-upgrade/references/conflict-resolution.md`(英文),内容包括:
  - Tier 1 冲突的处理范式(立刻转人工 + 报告路径与 changelog 链接)
  - Tier 2 三路合并的判断:何时 AI 信心高(语义清晰、变更孤立)、何时低(跨多文件、涉及业务规则)
  - Tier 3 重生成的步骤(`git checkout --theirs <path>` + `make dao` / `make ctrl`)
  - 常见冲突示例与处理(import 重排、相邻方法新增、签名变更等)
- [x] 11.3 创建 `.claude/skills/lina-upgrade/references/escalation-rules.md`(英文),严格列出 5 条转人工硬规则(同 specs):
  1. Tier 1 区域出现冲突
  2. Tier 2 三路合并 AI 信心不足或语义有歧义
  3. SQL 迁移可能破坏用户数据(DROP COLUMN 命中已有数据等)
  4. e2e smoke 失败且自动回滚未恢复
  5. 用户主动声明改过的核心文件被上游也改了
- [x] 11.4 创建 `.claude/skills/lina-upgrade/references/changelog-conventions.md`(英文),内容包括:
  - 如何从 `CHANGELOG.md` 提取版本区间内的条目
  - 如何遍历 `openspec/changes/archive/` 中归档变更的 `proposal.md` 提取"Why" / "What Changes"
  - 如何识别 `**BREAKING**` 标记
  - 如何把破坏性变更与 Tier 1 路径关联,在升级计划中显著展示

## 12. 升级技能 scripts 脚本族

- [x] 12.1 创建 `.claude/skills/lina-upgrade/scripts/upgrade-baseline-check.sh`(英文注释),实现 design.md D7 的四层校验:
  - 读 `apps/lina-core/manifest/config/metadata.yaml.framework.version`(用 awk 解析,不依赖 yq)
  - 解析或新增 `upstream` remote(若 `origin` 指向官方仓库则复用 origin)
  - `git fetch --tags --quiet upstream`
  - 层 1 存在性 → `ERR_TAG_NOT_FOUND` + 列出最近 3 个 stable tag
  - 层 2 可达性 → `ERR_HEAD_NOT_DESCENDANT` + 输出 tag commit / HEAD commit
  - 层 3 身份对照 → 软警告
  - 层 4 汇总 → `OK_BASELINE_CONFIRMED` + 4 项指标
  - 必须保证脚本无副作用(只读)
- [x] 12.2 创建 `.claude/skills/lina-upgrade/scripts/upgrade-classify.sh`,接收单个路径参数,输出 `tier1` / `tier2` / `tier3` / `unknown`,实现路径模式匹配(可读取 `references/tier-classification.md` 中的模式定义,但更稳妥的实现是把模式列表内联到脚本本身,文档作为权威源)
- [x] 12.3 创建 `.claude/skills/lina-upgrade/scripts/upgrade-plan.sh`,输出结构化升级计划:
  - 目标版本、基线版本、commits_ahead
  - changelog 摘要(从 CHANGELOG.md 与 OpenSpec 归档抽取)
  - 改动文件清单(`git diff baseline...HEAD -- apps/lina-core apps/lina-vben`)按 Tier 分组
  - 新 SQL 文件清单(只列编号大于当前最大值的)
  - 风险标识(标 `**BREAKING**` 的归档变更突出展示)
- [x] 12.4 创建 `.claude/skills/lina-upgrade/scripts/upgrade-regenerate.sh`,封装 `make dao` + `make ctrl`,捕获失败并打印日志路径
- [x] 12.5 创建 `.claude/skills/lina-upgrade/scripts/upgrade-verify.sh`,封装:
  - `cd apps/lina-core && go build ./...`
  - `cd apps/lina-vben && pnpm typecheck`
  - `cd apps/lina-vben && pnpm lint`
  - `cd hack/tests && pnpm playwright test e2e/login/TC0001-login.ts`(以及插件加载 smoke、CRUD smoke,具体用例编号待 lina-e2e 技能确定)
- [x] 12.6 所有脚本必须有 `set -euo pipefail`、文件用途注释、英文 inline 注释
- [x] 12.7 用 shellcheck 校验所有 5 个脚本
  - 2026-04-30 执行记录:本机未安装 `shellcheck`;已通过 Docker 镜像 `koalaman/shellcheck:stable` 对 5 个升级脚本及安装/test 脚本运行 `shellcheck -x`,并通过 `bash -n` 校验 5 个脚本。

## 13. 移除旧的 make upgrade 实现

- [x] 13.1 删除整个 `hack/tools/upgrade-source/` 目录(独立 Go module)
- [x] 13.2 删除 `Makefile` 中的 `upgrade` target 与 `.PHONY: upgrade` 声明(约第 44-58 行)
- [x] 13.3 删除 `apps/lina-core/Makefile` 中的 `upgrade` 代理 target(约第 41-52 行)
- [x] 13.4 删除 `apps/lina-core/hack/config.yaml` 中的整个 `frameworkUpgrade:` 区块(若 `repositoryUrl` 在其他地方有引用,先 grep 确认无依赖再删)
- [x] 13.5 全仓 `grep -rn "make upgrade"` 验证除文档外没有代码残留
  - 2026-04-30 执行记录:代码和运行文档已清理;仅当前变更文档与未归档的 `openspec/specs/**` 基线规范仍包含旧条目,后者按 OpenSpec 流程待归档时由本变更同步更新。

## 14. 调整宿主代码引用

- [x] 14.1 检查 `apps/lina-core/internal/service/plugin/internal/sourceupgrade/sourceupgrade.go` 第 495 行附近的错误消息,把 `command=make upgrade confirm=upgrade scope=source-plugin plugin=...` 改为 lina-upgrade 技能引导文本(英文示例:`use the lina-upgrade skill via your AI tooling, e.g. ask "upgrade source plugin <id>"`)
- [x] 14.2 同步更新 `bulkCommand` 提示文本(行 509)
- [x] 14.3 更新 `apps/lina-core/internal/service/plugin/plugin_source_upgrade_test.go` 测试期望(行 168)以匹配新提示文本
- [x] 14.4 更新 `apps/lina-core/internal/service/plugin/internal/sourceupgrade/sourceupgrade_status_test.go` 测试期望(行 128-130)
- [x] 14.5 全仓 `grep -rn "frameworkUpgrade"` 应返回零结果(Go 代码不再引用)
  - 2026-04-30 执行记录:Go 代码与运行配置已清理;仅当前变更文档与未归档的 `openspec/specs/**` 基线规范仍包含旧条目,后者按 OpenSpec 流程待归档时由本变更同步更新。
- [x] 14.6 在 `apps/lina-core/manifest/config/metadata.yaml` 的 `framework.version` 字段处增加注释红线:
  ```yaml
  # IMPORTANT: do not edit this field manually.
  # The lina-upgrade skill validates this value as the upgrade baseline.
  # 重要:请勿手工修改此字段。lina-upgrade 技能会以此值作为升级基线进行校验。
  ```

## 15. 更新现有 README/CLAUDE.md/项目文档

- [x] 15.1 更新仓库根 `README.md`(英文):
  - 安装章节改为 `curl -fsSL https://linapro.ai/install.sh | bash`
  - Windows 用户使用 Git Bash / WSL 的提示
  - 升级章节改为"通过 AI 工具(Claude Code 等)调用 lina-upgrade 技能"
  - 删除所有 `make upgrade` 引用
- [x] 15.2 更新仓库根 `README.zh_CN.md`(中文镜像)与 15.1 对应
- [x] 15.3 更新 `CLAUDE.md` 的"常用命令"章节,删除 `make upgrade` 引用,增加 `lina-upgrade` 技能调用说明
- [x] 15.4 更新 `apps/lina-core/README.md`(英文):
  - 删除"开发态升级"章节中的 `make upgrade` 命令
  - 改为引导到 `lina-upgrade` 技能,保留对源码插件升级流程的说明(说明触发方式)
- [x] 15.5 更新 `apps/lina-core/README.zh_CN.md`(中文镜像)与 15.4 对应
- [x] 15.6 更新 `apps/lina-plugins/README.md`(英文)与 `apps/lina-plugins/README.zh_CN.md`(中文):
  - 把 `make upgrade confirm=upgrade scope=source-plugin plugin=<id>` 改为 lina-upgrade 技能引导文本
- [x] 15.7 更新 `apps/lina-plugins/OPERATIONS.md`(若文件类似双语处理)
- [x] 15.8 检查 `.agents/instructions/` 与 `.agents/prompts/` 中是否有 `make upgrade` 引用,若有同步更新
- [ ] 15.9 全仓最终核对:`grep -rn "make upgrade"` 应只在 OpenSpec 归档目录或本变更文档中出现(因为归档变更不允许修改)
  - 2026-04-30 执行记录:运行时代码、README、工具文档与脚本入口均已清理;当前仍有 `openspec/specs/source-upgrade-governance/spec.md`、`openspec/specs/plugin-upgrade-governance/spec.md` 旧基线规范命中,需在 `/opsx:archive` 同步主规范时消除。

## 16. E2E 与单元测试

- [x] 16.1 调用 `lina-e2e` 技能,生成新的 E2E 测试用例:
  - `hack/tests/e2e/install/TC{NNNN}-install-default.ts` 验证 `bash hack/scripts/install/bootstrap.sh` 在 fixture 环境中能成功 clone + init(用 mock GitHub 或真实小型 fixture 仓库)
  - `hack/tests/e2e/install/TC{NNNN}-install-version-override.ts` 验证 `LINAPRO_VERSION=v0.x.y` 覆盖逻辑
  - `hack/tests/e2e/install/TC{NNNN}-install-skip-mock.ts` 验证 `LINAPRO_SKIP_MOCK=1` 跳过 mock 加载
  - 实际 TC 编号由 lina-e2e 技能分配
  - 2026-04-30 执行记录:已分配 `TC0155` / `TC0156` / `TC0157`,并复用 `hack/tests/scripts/install-bootstrap.sh` 的本地 git fixture。
- [x] 16.2 为 `upgrade-baseline-check.sh` 编写 bash 单元测试(用 `bats` 或自定义 fixture):
  - 通过场景:declared 等于 tag 且 HEAD 是后代
  - `ERR_TAG_NOT_FOUND` 场景
  - `ERR_HEAD_NOT_DESCENDANT` 场景
  - 测试文件放 `hack/tests/scripts/upgrade-baseline-check.bats`(若决定不引入 bats,则用纯 bash assertion 风格,放同路径)
  - 2026-04-30 执行记录:采用纯 bash assertion 风格,文件为 `hack/tests/scripts/upgrade-baseline-check.sh`,已运行通过。
- [x] 16.3 为 `upgrade-classify.sh` 编写 bash 单元测试,覆盖每个 Tier 的代表性路径
  - 2026-04-30 执行记录:`hack/tests/scripts/upgrade-classify.sh` 已运行通过。
- [x] 16.4 为 14.3 / 14.4 修改的 Go 测试运行 `go test ./...`,确保通过
  - 2026-04-30 执行记录:已运行定向测试 `go test ./internal/service/plugin/internal/sourceupgrade ./internal/service/plugin -run 'TestBuildSourcePluginUpgradePendingErrorIncludesBulkCommand|TestValidateSourcePluginUpgradeReadinessBlocksPendingUpgrade'`;已运行 `cd apps/lina-core && go test -p 1 ./... -count=1` 通过;已运行 `go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/integration ./internal/service/plugin ./internal/service/plugin/internal/sourceupgrade -count=1` 通过;最新默认并行 `cd apps/lina-core && go test ./... -count=1` 通过。
- [ ] 16.5 跑现有完整 E2E 套件 `make test`,确保没有回归
  - 2026-04-30 执行记录:完整 `make test` 已执行,结果为 `333 passed`, `3 failed`, `6 skipped`, `3 did not run`;失败项 `TC0098c`、`TC0054h`、`TC0154a` 已定向重跑通过。`TC0154a` 的测试污染根因已通过恢复 `IsBuiltin` 清理修复;最新 bootstrap 变更后已重跑安装 E2E `TC0155` / `TC0156` / `TC0157` 通过,未再完整重跑 `make test`。
  - 2026-04-30 复跑记录:最新完整 `make test` 结果为 `337 passed`, `2 failed`, `6 skipped`;失败项为 `TC0098c monitor-server` 未等到 `服务信息` 与 `TC0054h` 60s 超时。随后定向重跑 `TC0098c|TC0054h` 通过,判断仍属于长套件波动项,未发现本变更直接引入的稳定回归。
  - 2026-04-30 再复跑记录:完整 `make test` 结果为 `338 passed`, `1 failed`, `6 skipped`;唯一失败项仍为 `TC0054h` 60s 超时。随后定向重跑 `cd hack/tests && pnpm playwright test e2e/settings/dict/TC0054-dict-type-import-upload.ts -g 'TC0054h'` 通过(`1 passed`),判断仍属于长套件波动项,未发现本变更直接引入的稳定回归。
- [x] 16.6 在 fixture 仓库上手动跑一次完整升级流程(框架升级 + 源码插件升级各一次),记录人工验证清单
  - 2026-04-30 执行记录:已在临时 fixture 仓库上验证框架升级路径:创建 `v0.5.0` baseline 与 `v0.6.0` target tag,运行 `upgrade-baseline-check.sh` 通过,运行 `upgrade-plan.sh v0.6.0` 并确认 Tier 分类、新 SQL 列表和源码插件路径识别,随后执行 `git merge --no-commit --no-ff refs/tags/v0.6.0` 通过且 metadata/plugin manifest 均更新到 `v0.6.0`;源码插件升级路径已运行 `go test ./internal/service/plugin -run 'TestUpgradeSourcePluginAppliesPreparedRelease|TestValidateSourcePluginUpgradeReadinessFailsForPendingUpgrade|TestListSourceUpgradeStatusesSkipsDynamicPlugins' -count=1` 通过。

## 17. CDN 部署文档化(附属人工任务)

- [x] 17.1 在 `hack/scripts/install/README.md` / `README.zh_CN.md` 末尾增加"Deployment to linapro.ai"章节,描述以下流程:
  - `bootstrap.sh` 通过 CI/CD 推送到 `linapro.ai` CDN 路径 `/install.sh`
  - 同步推送 `/install.ps1`(如果未来需要,目前留为占位说明)
  - CDN 缓存失效策略(每次新 stable tag 发布同步更新)
  - 验证步骤(在干净环境中跑 `curl -fsSL https://linapro.ai/install.sh | bash`)
- [ ] 17.2 在 PR 描述中明确标注"本 PR 合并后需运维侧执行 17.1 流程,远程入口才会生效"
  - 2026-04-30 执行记录:PR 描述草稿已在 18.7 准备;实际 PR 创建时仍需复制该运维提示。

## 18. 自检与发布前检查

- [x] 18.1 跑 `openspec validate linapro-install-upgrade --strict`,所有 spec 验证通过
- [x] 18.2 跑 `make build`(后端) + `pnpm run build`(前端),确认编译通过
  - 2026-04-30 执行记录:`make build` 通过(包含前端 build、manifest pack、Wasm build、宿主二进制 build);`cd apps/lina-vben && pnpm run check:type` 通过;`pnpm lint` 因既有无关 stylelint / prettier 问题失败,未作为本变更阻断项。
- [ ] 18.3 跑 `make test`,确保 E2E 全部通过
  - 2026-04-30 执行记录:见 16.5,完整套件最新仍有 1 个波动失败且定向重跑通过;完整全绿记录仍需后续在隔离这些长套件波动后取得。
- [x] 18.4 跑 `go test ./...` 在 `apps/lina-core/`,确保单元测试通过
  - 2026-04-30 执行记录:`go test -p 1 ./... -count=1` 通过;最新默认并行 `go test ./... -count=1` 通过。
- [x] 18.5 调用 `/lina-review` 技能进行代码与规范审查
  - 2026-04-30 执行记录:已按 `lina-review` 对 `git status --short`、`git ls-files --others --exclude-standard` 与展开后的未跟踪目录进行全量审查;后端 Go 变更、安装脚本、升级技能脚本、E2E 命名与 i18n 影响均已覆盖。审查未发现 high / medium 阻断问题。
- [x] 18.6 处理 `/lina-review` 反馈中的所有 high / medium 严重度问题
  - 2026-04-30 执行记录:审查中发现的 bootstrap 覆盖目录安全边界、latest release 非稳定 tag 识别、`confirm` EOF 处理、MySQL 版本探测 warning 处理、脚本可执行位、baseline-check fixture 误访问真实 GitHub 的问题均已修复并通过 `make test-install`、`shellcheck -x`、脚本语法校验、安装 E2E 与升级脚本单测复验。剩余事项为环境或流程残留:官方 `v0.5.0` tag 待发布确认、完整 E2E 需后续取得一次全绿记录、`pnpm lint` 仍有无关既有样式问题。
- [x] 18.7 准备 PR 描述,包含:
  - 变更摘要
  - 已知影响(make upgrade 删除、frameworkUpgrade 字段移除)
  - 运维侧需要执行的 CDN 部署动作(指向 17.1)
  - 测试覆盖情况
  - 2026-04-30 PR 描述草稿:
    - Summary: replace legacy installer with `hack/scripts/install/bootstrap.sh` + platform scripts, add `lina-upgrade` skill workflow, remove `make upgrade` / `hack/tools/upgrade-source`, update source-plugin pending-upgrade hints and docs.
    - Breaking impact: `make upgrade` is removed; `apps/lina-core/hack/config.yaml.frameworkUpgrade` is removed; upgrade baseline is `apps/lina-core/manifest/config/metadata.yaml.framework.version`; Windows install uses Git Bash or WSL.
    - Ops action after merge: deploy `hack/scripts/install/bootstrap.sh` byte-identically to `https://linapro.ai/install.sh`, refresh CDN cache, then verify in a clean environment with `curl -fsSL https://linapro.ai/install.sh | bash`.
    - Tests: `openspec validate linapro-install-upgrade --strict`; `git diff --check`; Docker `koalaman/shellcheck:stable shellcheck -x`; `bash -n` for installer/upgrade/test shell scripts; `make test-install`; `bash hack/tests/scripts/upgrade-classify.sh`; `bash hack/tests/scripts/upgrade-baseline-check.sh`; install E2E `TC0155` / `TC0156` / `TC0157`; `make build`; `pnpm run check:type`; `go test -p 1 ./... -count=1`; `go test ./... -count=1`. Known residuals: `pnpm lint` fails on unrelated existing style/prettier files; latest full `make test` still has the recorded `TC0054h` long-suite timeout, while targeted rerun passes.
  - i18n 影响评估:本变更修改了源码插件升级待处理错误提示,已同步 `apps/lina-core/manifest/i18n/{en-US,zh-CN,zh-TW}/error.json`;未新增或修改 API DTO,因此无需新增 apidoc i18n 资源;前端运行时语言包、菜单、路由、按钮、表单、表格文案无新增影响;OpenAPI title 源文本从 Service 调整为 Framework,中文文档站已有源数据展示路径,无需新增 apidoc 翻译键。

## 19. 归档前确认(由用户决定是否走 /opsx:archive)

- [ ] 19.1 用户确认本次迭代功能完整、无遗漏反馈
- [ ] 19.2 调用 `/lina-review` 进行归档前最终审查
- [ ] 19.3 准备归档时把所有 OpenSpec 文档(`proposal.md` / `design.md` / `tasks.md` / 增量规范)翻译为英文(项目规范要求)
- [ ] 19.4 执行 `/opsx:archive linapro-install-upgrade`
