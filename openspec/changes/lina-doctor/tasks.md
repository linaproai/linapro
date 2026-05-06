## 1. 准备工作

- [x] 1.1 阅读 `proposal.md` / `design.md` / `specs/framework-environment-doctor/spec.md` / `specs/framework-bootstrap-installer/spec.md` 确认对全部决策达成一致
- [x] 1.2 确认 `linapro-install-upgrade` 变更已删除且 `lina-doctor` 是唯一活跃安装入口变更;以当前工作树为 baseline 跑 `openspec validate lina-doctor --strict` 确认无 spec 偏差
- [x] 1.3 在仓库根创建工作分支(若使用 worktree 则按 `git-worktree` 技能流程执行)
- [x] 1.4 通过 `brew info openspec` / `brew info gf` / `npm view openspec` / `npm view @fission-ai/openspec` 快照各通道当前可用性,作为 `tool-matrix.md` 的实施依据;确认 `openspec@0.0.0` 不作为安装通道
- [x] 1.5 通过 `npm view skills` 与对 `https://github.com/gogf/skills` 的探查确认 `npx skills add github.com/gogf/skills -g` 安装通道当前可用性,记录到 `references/troubleshooting.md`
- [x] 1.6 把当前 `hack/scripts/install/lib/_common.sh` 中可能被 `lina-doctor` 复用的纯工具函数(`version_ge` / `confirm` / `retry` / `detect_os` / `is_port_in_use`)清单化,作为后续迁移到 `.claude/skills/lina-doctor/lib/_common.sh` 的依据

## 1.5 install 脚本职责收敛(在 lina-doctor 技能落地前完成,避免出现"环境探测真空期")

- [x] 1.5.1 在仓库根创建 `lina-doctor` 工作分支后,先把 `.claude/skills/lina-doctor/lib/_common.sh` 写好(从原 `hack/scripts/install/lib/_common.sh` 迁出 `version_ge` / `confirm` / `retry` / `detect_os` 等纯工具函数;不迁与 install 强相关的 `run_standard_install` / `run_prereq_check` / `download_go_modules` / `install_frontend_deps` / `copy_default_config` / `run_make_or_fallback` / `run_core_fallback`)
- [x] 1.5.2 删除 `hack/scripts/install/install-macos.sh` / `install-linux.sh` / `install-windows.sh`
- [x] 1.5.3 删除 `hack/scripts/install/checks/prereq.sh`,并清理空的 `checks/` 目录(连同 `.gitkeep`)
- [x] 1.5.4 删除 `hack/scripts/install/lib/_common.sh`,并清理空的 `lib/` 目录(连同 `.gitkeep`)
- [x] 1.5.5 重写 `hack/scripts/install/bootstrap.sh`,只保留以下逻辑:
  - 文件顶部注释保留"本文件 = `linapro.ai/install.sh` 部署内容"声明
  - `set -Eeuo pipefail` + `trap on_error / on_exit`
  - `detect_os`(失败立即 die)
  - `resolve_version`(读 `LINAPRO_VERSION` 或解析 GitHub `releases/latest` redirect)
  - `prepare_target`(读 `LINAPRO_DIR`,默认 `./linapro`;非空检查 + `LINAPRO_FORCE` 安全门)
  - `git clone --branch <tag> <url> <dir>`(`LINAPRO_SHALLOW=1` 时附加 `--depth 1`)
  - 打印 next-steps banner(三行指引:`cd <dir>` / `ask Claude Code "run lina-doctor ..."` / `make init && make dev`)
  - **完全移除** `exec bash hack/scripts/install/install-${os_name}.sh "$@"` 行
  - **完全移除** 任何对 `prereq.sh` / `make init` / `make mock` / `go mod download` / `pnpm install` / `is_port_in_use` 的调用
- [x] 1.5.6 移除 `bootstrap.sh` 中对 `LINAPRO_NON_INTERACTIVE` 与 `LINAPRO_SKIP_MOCK` 环境变量的处理逻辑(它们不再有意义);保留 `LINAPRO_VERSION` / `LINAPRO_DIR` / `LINAPRO_SHALLOW` / `LINAPRO_FORCE`
- [x] 1.5.7 跑 Docker `koalaman/shellcheck:stable` 对新版 `bootstrap.sh` 校验,并用 `bash -n` 做语法校验
- [x] 1.5.8 重写 `hack/tests/e2e/install/TC0155-install-default.ts` 用例:断言 clone 完成 + next-steps 文本包含 `lina-doctor` 关键字 + 不再 `make init` / `make mock`
- [x] 1.5.9 重写 `hack/tests/e2e/install/TC0156-install-version-override.ts` 用例:仅断言 `LINAPRO_VERSION` 覆盖逻辑生效,不验证后续 init 行为
- [x] 1.5.10 删除 `hack/tests/e2e/install/TC0157-install-skip-mock.ts` 用例;`LINAPRO_SKIP_MOCK` 不再存在,该用例失去语义
- [x] 1.5.11 同步精简 `hack/scripts/install/README.md` / `README.zh_CN.md`,移除"prereq.sh"、"install-*.sh"、"环境探测"、"`make init`"等章节;只保留 `bootstrap.sh` 单文件入口的语义说明 + Next steps 指引
- [x] 1.5.12 在 PR 描述中显式列出"运维侧需要把更新版 `bootstrap.sh` 重新部署到 `https://linapro.ai/install.sh`"

## 2. 技能目录结构搭建

- [x] 2.1 创建目录 `.claude/skills/lina-doctor/` 与 `scripts/`、`references/` 子目录
- [x] 2.2 在每个新建目录下创建 `.gitkeep`(若为空目录占位)避免 git 忽略
- [x] 2.3 创建 `.claude/skills/lina-doctor/SKILL.md`,frontmatter 格式与 `lina-upgrade/SKILL.md` 一致:
  ```
  ---
  name: lina-doctor
  description: <中文一句话,同时保留 environment, install, fix, lina-doctor, dependencies, toolchain 等英文触发词>
  ---
  ```
- [x] 2.4 SKILL.md 正文按以下结构组织(中文撰写,命令名/环境变量/JSON 字段保持英文原文):
  - **Purpose**: 技能的目标与边界(诊断 + AI 引导安装,不替代 bootstrap.sh)
  - **When to invoke**: 触发关键词举例(中英文都覆盖)
  - **Inputs the AI must collect from the user**: 是否启用 `--check-only`、是否启用 `LINAPRO_DOCTOR_SKIP_PLAYWRIGHT` / `LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL` / `LINAPRO_DOCTOR_NON_INTERACTIVE`
  - **Workflow (10 steps)**: 严格按 design.md "10 步工作流"列出每一步,标注调用脚本、可能失败模式、转人工的判定条件
  - **Outputs the AI must produce**: 诊断 JSON、安装计划、最终报告、escalation 报告
  - **References**: 链接到 `references/` 下的 4 份文档
  - **Allowed tools**: `Bash`、`Read`、`Edit`、`Grep`、`Glob`

## 3. 公共依赖与脚本骨架

- [x] 3.1 在 `.claude/skills/lina-doctor/lib/_common.sh`(任务 1.5.1 已建立)的基础上补齐 `lina-doctor` 专属函数:`detect_shell` / `detect_package_manager` / `detect_node_version_manager` / `detect_npm_global_prefix` / `print_path_fix_hint`
- [x] 3.2 不得在多个脚本文件中重复实现相同逻辑;所有共享函数集中在 `lib/_common.sh`,`scripts/*.sh` 通过 `source` 引入
- [x] 3.3 在每个 `scripts/*.sh` 文件顶部加用途注释 + `set -euo pipefail` + `IFS` 标准化
- [x] 3.4 所有脚本在执行前必须 `cd` 到仓库根(用 `git rev-parse --show-toplevel` 获取),以保证相对路径稳定
- [x] 3.5 `LINAPRO_DOCTOR_DEBUG=1` 时启用 `set -x` 与 `log_debug` 输出

## 4. 诊断脚本 `scripts/doctor-detect.sh`

- [x] 4.1 创建 `scripts/doctor-detect.sh`,职责仅限"探测元数据"——OS、包管理器、shell、版本管理工具、镜像变量、仓库根
- [x] 4.2 探测项:
  - OS: 复用 `_common.sh` 的 `detect_os`(返回 `macos` / `linux` / `windows`)
	  - 包管理器优先级:macOS = `brew`;Linux/WSL 默认 = `apt-get` > `dnf` > `yum` > `pacman`;Windows Git Bash = `winget` > `scoop` > `choco`;WSL 仅在用户明确要求 Windows 主机安装时才走 PowerShell 包装的 Windows 包管理器
  - Shell:从 `$SHELL` + `$BASH_VERSION` + `$ZSH_VERSION` + `$PSModulePath` 推断,返回 `bash` / `zsh` / `fish` / `powershell`
  - Node 版本管理工具:`$NVM_DIR` / `$FNM_DIR` / `$VOLTA_HOME` 与对应 `command -v`,返回首个命中的工具或 `none`
  - 镜像变量:`GOPROXY`、`npm config get registry`(若 npm 已装)、`PLAYWRIGHT_DOWNLOAD_HOST`
  - 仓库根:存在 `apps/lina-core` 与 `apps/lina-vben` 则视为 `repo_root_detected: true`
- [x] 4.3 输出结果作为单一 JSON 对象写入 stdout,stderr 仅放调试信息

## 5. 诊断脚本 `scripts/doctor-check.sh`

- [x] 5.1 创建 `scripts/doctor-check.sh`,职责为"基于 `doctor-detect.sh` 的元数据,逐工具检查存在性与版本"
- [x] 5.2 调用 `doctor-detect.sh` 获取元数据后,按以下顺序检查 9 个工具:
  - `git`:仅 `command -v git`
  - `make`:仅 `command -v make`
  - `go`:`command -v go` + `go version` 输出版本对比 `1.22`
  - `node`:`command -v node` + `node --version` 对比 `20.19`
  - `pnpm`:`command -v pnpm` + `pnpm --version` 对比 `8`
  - `openspec`:`command -v openspec` + `openspec --version`(无最低版本)
  - `gf`:`command -v gf` + `gf version`(无最低版本)
  - `playwright`:在 `hack/tests/` 内 `pnpm exec playwright --version` 是否成功 + `node_modules/.cache/ms-playwright/` 或等价路径是否有 `chromium`
  - `goframe-v2`:检查 `~/.claude/skills/goframe-v2/SKILL.md` 是否存在
- [x] 5.3 PATH 自检:对 `gf`(`$HOME/go/bin`)、`pnpm`(npm prefix `bin`)、`openspec`(npm prefix `bin`)三项检查是否在 `$PATH` 中,把缺失项放入 `path_issues` 数组
- [x] 5.4 镜像探测:对 `GOPROXY` / npm registry / `PLAYWRIGHT_DOWNLOAD_HOST` 三项检查是否需要建议,放入 `mirror_hints` 数组
- [x] 5.5 拼装最终 JSON,保证结构与 design.md D7 中的 schema 完全一致,使用 `printf` 或 here-doc 输出避免 jq 依赖
- [x] 5.6 退出码:全部 `ok` = 0;任一关键工具(go/node/pnpm/git/make/openspec/gf)缺失或版本过低 = 1;仅有 PATH / 镜像 / Playwright / goframe-v2 警告 = 2
- [x] 5.7 支持 `--check-only` 标志:行为与默认完全相同,只是显式声明语义,便于 AI 工具区分调用意图

## 6. 计划脚本 `scripts/doctor-plan.sh`

- [x] 6.1 创建 `scripts/doctor-plan.sh`,接收 `doctor-check.sh` 的 JSON(从 stdin 或 `--input` 文件路径)
- [x] 6.2 按拓扑序生成安装清单(go/node/git/make 同优先级,gf 依赖 go,pnpm 依赖 node,openspec 在 macOS brew 通道不依赖 node、npm 通道依赖 node,goframe-v2 依赖 node,playwright 依赖 pnpm + repo_root)
- [x] 6.3 nvm/fnm/volta 命中时,Node 安装命令切换为对应工具的 `install 20`
- [x] 6.4 macOS openspec 优先 `brew install openspec`,失败回落 `npm i -g @fission-ai/openspec@latest`(在 install 阶段实现回落逻辑,plan 阶段先用 brew 命令)
- [x] 6.5 Windows 包装:winget/scoop/choco 命令统一通过 `powershell.exe -NoProfile -Command "..."` 包装
- [x] 6.6 sudo 标注:Linux 下 `apt-get` / `dnf` / `yum` / `pacman` 命令前缀 `sudo`,但要在 plan 中显式标注 "needs sudo"
- [x] 6.7 输出结构化计划(JSON 数组),每项包含 `tool` / `command` / `package_manager` / `requires_sudo` / `depends_on` / `optional`(布尔,标记 `playwright` 与 `goframe-v2` 为 true)字段
- [x] 6.8 镜像建议:plan 头部输出 `mirror_hints` 段,内容来自 `doctor-check.sh` JSON 的 `mirror_hints` 字段,但不在 `command` 字段中插入 `export GOPROXY=...`
- [x] 6.9 跳过逻辑:`LINAPRO_DOCTOR_SKIP_PLAYWRIGHT=1` 命中时不输出 playwright 项;`LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL=1` 命中时不输出 goframe-v2 项;两者皆未设时默认包含两项

## 7. 安装调度脚本 `scripts/doctor-install.sh`

- [x] 7.1 创建 `scripts/doctor-install.sh`,接收 `doctor-plan.sh` 的 JSON(从 stdin 或 `--plan` 文件路径)
- [x] 7.2 按拓扑序逐项执行:
  - 显示当前命令、所选包管理器、是否需要 sudo
  - 单项确认(除非 `LINAPRO_DOCTOR_NON_INTERACTIVE=1`)
	  - 执行命令,捕获 stdout / stderr 到 `/tmp/lina-doctor-<tool>.log`
	  - 单项超时:`LINAPRO_DOCTOR_TIMEOUT`(默认 300s),超时杀进程并归类 `network`
	  - 立刻调用 `doctor-check.sh` 中对应工具的检查函数(把检查函数抽到 `lib/doctor-checks.sh` 共享)
	  - 检查通过则进入下一项;关键工具检查失败则停止并 emit escalation;`optional: true` 目标检查失败则记录非阻塞 escalation 并继续后续步骤
- [x] 7.3 PATH 处理:
  - 安装 `gf` 后若 `$HOME/go/bin` 不在 `$PATH` 中,本进程内 `export PATH="$HOME/go/bin:$PATH"`
  - 安装 `pnpm` / `openspec` 后若 npm 全局 prefix `bin` 不在 `$PATH` 中,同样 export
  - 单独打印两行:当前 shell 用的 `export` 命令、可追加到 rc 的命令片段
  - 不写入用户 rc 文件
- [x] 7.4 macOS openspec brew 失败回落:第一次 `brew install openspec` 失败时归类 `package_not_found`,自动重试 `npm i -g @fission-ai/openspec@latest`(此时 Node ≥ 20.19 已就绪),再失败才 emit escalation
- [x] 7.5 可选目标失败非阻塞:当 `playwright` 或 `goframe-v2` 这类 `optional: true` 的步骤失败时,不停止后续步骤,只记录到"非阻塞失败"列表,继续执行;最终报告中分两段呈现"已成功"与"非阻塞失败"
- [x] 7.6 全部安装完成后调用 `doctor-verify.sh`(见 8)做最终复检

## 8. 验证脚本 `scripts/doctor-verify.sh`

- [x] 8.1 创建 `scripts/doctor-verify.sh`,职责为"全部 install 完成后,跑一次完整 `doctor-check.sh` + 关键工具的 smoke 输出"
- [x] 8.2 smoke 检查:
  - `go version` 输出版本号
  - `node --version` 输出版本号
  - `pnpm --version` 输出版本号
  - `gf version` 输出 GoFrame 版本
  - `openspec --version` 输出 OpenSpec 版本
  - 在 `hack/tests/` 内 `pnpm exec playwright --version` 输出 Playwright 版本
  - `~/.claude/skills/goframe-v2/SKILL.md` 文件存在性 + `head -5` 验证内容像合法 frontmatter
- [x] 8.3 如果任一 smoke 失败,把失败工具加入 escalation 列表,但不阻塞已成功的输出
- [x] 8.4 输出最终报告,包含:
  - 安装成功的工具及其版本
  - 跳过的工具(已满足或被 LINAPRO_DOCTOR_SKIP_PLAYWRIGHT 跳过)
  - PATH 修复提示(如有)
  - 镜像建议(如有)
	  - 下一步推荐命令:`make init` / `make dev` / `openspec list` / `pnpm test`

## 9. Escalation 输出 `lib/doctor-escalate.sh`

- [x] 9.1 创建 `lib/doctor-escalate.sh`,封装结构化 escalation 输出
- [x] 9.2 输入:失败工具名、命令、log 文件路径、根因类别
- [x] 9.3 根因推断逻辑(基于 stderr 关键词):
  - `network`:`timeout` / `Could not resolve host` / `connection refused` / `proxy` / `i/o timeout`
  - `permission`:`EACCES` / `permission denied` / `operation not permitted` / `must be run as root`
  - `package_not_found`:`No such package` / `Unable to locate package` / `formula was not found` / `not in registry`
  - `shim_conflict`:`already installed` / `conflicts with` / `node command in PATH points to`
  - 兜底:`unknown`
- [x] 9.4 输出格式:
  - 顶部一行 `ERROR: failed to install <tool>`
  - 紧跟"Command:"、"Package manager:"、"Root cause:"、"Last 50 lines of output:"四段
  - 末尾"Recommended action:"按平台 + 根因给出具体建议(从 `references/troubleshooting.md` 中索引)
- [x] 9.5 同时把上述结构序列化为 JSON 写入 `/tmp/lina-doctor-escalation.json`,供 AI 工具进一步消费

## 10. References 文档族

- [x] 10.1 创建 `references/tool-matrix.md`(中文),内容包括:
  - 9 个工具 × 3 平台 × 各包管理器命令的二维矩阵
  - 每个命令是否需要 sudo / PowerShell wrapper
  - macOS openspec 的 brew + npm 双通道说明
  - 版本下限的权威值(与 spec.md 表格保持一致,本文档为权威源)
- [x] 10.2 创建 `references/install-strategy.md`(中文),内容包括:
  - 拓扑顺序(图示)
  - 包管理器优先级表
  - nvm / fnm / volta 探测策略
  - Windows PowerShell wrapper 调用模板
  - 单步超时与重试策略
- [x] 10.3 创建 `references/path-and-shell.md`(中文),内容包括:
  - `$GOBIN` 默认值与各 shell rc 写入片段(zsh / bash / fish / PowerShell)
  - npm 全局 prefix 路径(`npm config get prefix`)
  - 用户已有 rc 文件冲突时的安全处理(只追加注释行 + 命令,不删改既有内容)
  - PowerShell `$PROFILE` 路径与不同 PowerShell 主机版本差异
- [x] 10.4 创建 `references/troubleshooting.md`(中文),内容包括:
  - 常见 escalation 根因类别 × 推荐人工动作
  - nvm / fnm / volta 撞 PATH shim 的诊断与修复
  - 国内镜像列表(GOPROXY / npm registry / PLAYWRIGHT_DOWNLOAD_HOST)
  - winget 静默失败、scoop 缓存损坏、apt sources.list 失效的常见错误
  - `gf` 装到 GOBIN 但未在 PATH 的固定修复脚本
  - `Playwright browsers` 下载卡住时的镜像切换步骤

## 11. goframe-v2 仓库自带形态移除

- [x] 11.1 通过 `git rm -r .claude/skills/goframe-v2/` 删除整个目录
- [ ] 11.2 提交一次专门的 commit(message 形如 "chore(skills): externalize goframe-v2 skill, install via lina-doctor"),便于 review 时清晰看到该删除动作的边界
- [x] 11.3 验证拉取本变更后的 fresh clone 中 `.claude/skills/` 不再含 `goframe-v2`
- [x] 11.4 检查仓库内是否有任何对 `.claude/skills/goframe-v2/` 的硬路径引用(`grep -rn ".claude/skills/goframe-v2"`),除当前变更文档与 archive 文档外应为零
- [x] 11.5 在 `lina-doctor` 的本地验收中,把"`.claude/skills/goframe-v2/` 不应存在,但 `~/.claude/skills/goframe-v2/SKILL.md` 应存在"作为强断言

## 12. 文档与索引同步

- [x] 12.1 更新 `README.md`(英文):
  - 重写 "Installation" 章节为三步指引:`curl | bash` 克隆 → invoke `lina-doctor` → `make init && make dev`
  - 在末尾追加一行 "If your environment is missing development tools (Go, Node, openspec, gf, goframe-v2 skill, etc.), invoke the `lina-doctor` skill via your AI tool"
  - 不展开 lina-doctor 命令细节
- [x] 12.2 更新 `README.zh_CN.md`(中文镜像)与 12.1 对应
- [x] 12.3 更新 `CLAUDE.md`:
  - 在"研发流程规范" / "常用命令"章节追加 `lina-doctor` 的简要说明,与 `lina-upgrade` 平级
  - 在 skill 列表中(若有)登记
  - 修改对 `goframe-v2` 技能的引用文本,从"项目仓库自带"改为"通过 `lina-doctor` 安装到用户全局技能目录后即可触发"
- [x] 12.4 更新 `.claude/skills/lina-upgrade/SKILL.md`:
  - 在 "Workflow" 第 1 步前 / "Inputs" 段后追加一句 "Prerequisite: ensure `gf` and `openspec` are installed; invoke `lina-doctor` first if any tool is missing"
  - 不在脚本中硬连接 lina-doctor,仅文档级软引用

## 13. E2E 与单元测试

- [x] 13.1 调用 `lina-e2e` 技能,生成 doctor 流程的 E2E 测试用例(基于 fixture 环境):
  - `hack/tests/e2e/doctor/TC{NNNN}-doctor-check-only.ts` 验证 `--check-only` 输出符合 schema 且不触发安装
  - `hack/tests/e2e/doctor/TC{NNNN}-doctor-skip-satisfied.ts` 验证已满足工具被正确跳过
  - `hack/tests/e2e/doctor/TC{NNNN}-doctor-skip-playwright.ts` 验证 `LINAPRO_DOCTOR_SKIP_PLAYWRIGHT=1` 跳过逻辑
  - `hack/tests/e2e/doctor/TC{NNNN}-doctor-skip-goframe-skill.ts` 验证 `LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL=1` 跳过逻辑
  - `hack/tests/e2e/doctor/TC{NNNN}-doctor-goframe-skill-nonblocking.ts` 验证 goframe-v2 安装失败不阻塞后续工具
  - 实际 TC 编号由 lina-e2e 技能分配
- [x] 13.2 为 `doctor-check.sh` 编写 bash 断言式单元测试,放在 `hack/tests/scripts/doctor-check.sh`:
  - 全工具就绪场景 → exit 0 + JSON `tools.*.ok` 全 true
  - 缺 `pnpm` 场景 → exit 1 + `tools.pnpm.ok = false`
  - 仅 PATH 警告场景 → exit 2
  - `~/.claude/skills/goframe-v2/SKILL.md` 缺失时 `tools.goframe-v2.ok = false` 但 exit 仍为 2
- [x] 13.3 为 `doctor-plan.sh` 编写 bash 断言式单元测试,放在 `hack/tests/scripts/doctor-plan.sh`:
	  - 验证拓扑顺序(go 在 gf 前、node 在 pnpm 前、node 在 goframe-v2 前)
	  - 验证 nvm 命中时 Node 命令变为 nvm install
	  - 验证 Windows 命令统一被 PowerShell 包装
	  - 验证 WSL 默认按 Linux 包管理器生成命令,不自动走 Windows PowerShell 包装
  - 验证 LINAPRO_DOCTOR_SKIP_PLAYWRIGHT 命中时 plan 不含 playwright
  - 验证 LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL 命中时 plan 不含 goframe-v2
  - 验证 playwright 与 goframe-v2 在 plan 中标记 `optional: true`
- [x] 13.4 为 `lib/doctor-escalate.sh` 编写 bash 断言式单元测试,放在 `hack/tests/scripts/doctor-escalate.sh`:
  - 验证根因推断:`EACCES` → permission;`timeout` → network;`Unable to locate package` → package_not_found
- [x] 13.5 跑现有完整 E2E 套件 `make test`,确保没有回归(尤其确认重写后的 `TC0155` / `TC0156` 通过,删除后的 `TC0157` 不再被执行)
- [ ] 13.6 在维护者本机做手动验收(macOS):卸载 `gf` / `openspec` / `Playwright browsers` / `~/.claude/skills/goframe-v2/` → 调用 `lina-doctor` → 验证全部装回,记录人工验证清单

## 14. 自检与发布前检查

- [x] 14.1 跑 `openspec validate lina-doctor --strict`,所有 spec 验证通过
- [x] 14.2 用 `koalaman/shellcheck:stable` Docker 镜像跑 `shellcheck -x` 校验所有新建脚本以及精简后的 `bootstrap.sh`
- [x] 14.3 跑 `bash -n` 校验所有新建脚本的语法
- [x] 14.4 跑 `make build`(后端) + `pnpm run build`(前端),确认无回归
- [x] 14.5 跑 `make test`,确保 E2E 全部通过
- [x] 14.6 跑 `git ls-files .claude/skills/goframe-v2/ 2>&1 | wc -l` 确认结果为 0,即仓库已不再追踪该目录
- [x] 14.7 调用 `/lina-review` 技能进行代码与规范审查
- [x] 14.8 处理 `/lina-review` 反馈中的所有 high / medium 严重度问题
- [x] 14.9 准备 PR 描述,包含:
  - 变更摘要(install 脚本收敛 + lina-doctor 新增 + goframe-v2 外部化)
  - 已知 BREAKING:`make init` / `make mock` 不再由 install 链路触发;`LINAPRO_SKIP_MOCK` / `LINAPRO_NON_INTERACTIVE` 移除;`.claude/skills/goframe-v2/` 仓库自带形态移除;平台 install 脚本与 `prereq.sh` 删除
  - 运维侧动作:更新版 `bootstrap.sh` 重新部署到 `https://linapro.ai/install.sh`
  - 测试覆盖情况
  - i18n 评估结论(无影响)
  - 缓存一致性评估结论(不涉及)

## 15. 归档前确认(由用户决定是否走 /opsx:archive)

- [ ] 15.1 用户确认本次迭代功能完整、无遗漏反馈
- [ ] 15.2 调用 `/lina-review` 进行归档前最终审查
- [ ] 15.3 准备归档时把所有 OpenSpec 文档(`proposal.md` / `design.md` / `tasks.md` / 增量规范)翻译为英文(项目规范要求)
- [ ] 15.4 执行 `/opsx:archive lina-doctor`

## Feedback

- [x] **FB-1**: Clarify that `hack/scripts/install` only downloads the repository source and no longer performs environment detection or installation
- [x] **FB-2**: Clarify that `goframe-v2` is not bundled with the repository source and is installed globally via `npx skills add github.com/gogf/skills -g`
- [x] **FB-3**: Clarify Linux `sudo` handling so elevated commands are displayed and confirmed before execution
- [x] **FB-4**: Clarify optional target failure handling so `Playwright browsers` and `goframe-v2` do not contradict the critical-tool stop policy
- [x] **FB-5**: Align design failure policy so only critical tool failures stop the workflow and optional target failures remain non-blocking
- [x] **FB-6**: Clarify WSL package-manager handling and avoid treating every WSL run as a Windows host package-manager install
- [x] **FB-7**: Align `openspec` install topology for macOS Homebrew primary channel and npm fallback
- [x] **FB-8**: Add `goframe-v2` to the documented `doctor-check.sh` JSON example
- [x] **FB-9**: Document `LINAPRO_FORCE` as the explicit safety override for non-empty install target directories
- [x] **FB-10**: Align doctor final next-step guidance to recommend `make init` before `make dev`
- [x] **FB-11**: Add the explicit 10-step doctor workflow referenced by the SKILL.md implementation task
- [x] **FB-12**: Treat `lina-doctor` as the authoritative active change after `linapro-install-upgrade` was manually removed
- [x] **FB-13**: Write `lina-doctor` SKILL.md and references in Simplified Chinese for maintainer review
- [x] **FB-14**: Use the official `@fission-ai/openspec` npm package and raise Node minimum to 20.19 for OpenSpec compatibility
- [x] **FB-15**: Align active OpenSpec delta specs to Chinese, keep parser-required `SHALL`/`MUST` markers, and complete PR description i18n/cache conclusions
- [x] **FB-16**: Ensure `doctor-install.sh` timeout failures append a timeout marker so escalation can classify them as `network`
- [x] **FB-17**: Add automated coverage for the `lina-upgrade` skill workflow and failure handling
- [x] **FB-18**: `make build` 应从 `hack/config.yaml` 读取目标平台/架构等构建配置，`image` 配置段不再维护二进制构建参数并改为复用 `make build` 产物
