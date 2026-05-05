## Context

### 当前状态

`LinaPro` 仓库当前提供以下"开发态环境治理"实现:

- `hack/scripts/install/checks/prereq.sh` 在 `bootstrap.sh` 安装链路末尾被调用,检查 `go ≥ 1.22` / `node ≥ 20` / `pnpm ≥ 8` / `git` / `make` / `mysql client` 的存在性与版本,占用 5666 / 8080 端口给 warning,缺失工具时只打印平台特定的安装提示文本(如 `brew install go`),**不动手安装**
- `hack/scripts/install/lib/_common.sh` 提供共享 bash 工具函数:`log_info` / `log_warn` / `log_error` / `version_ge` / `detect_os` / `is_port_in_use` / `confirm` / `retry` 等,可被新脚本复用
- `.claude/skills/lina-upgrade/` 是当前唯一的 LinaPro 专属技能,采用 "SKILL.md 主入口 + scripts/ 脚本族 + references/ 参考文档" 的目录结构,可作为 `lina-doctor` 的形式参考
- `openspec` CLI 1.3.1 可通过 Homebrew(`/opt/homebrew/bin/openspec`)安装;官方 npm 通道为 `@fission-ai/openspec`,而 `openspec@0.0.0` 是无效旧包,不得作为安装通道;`gf` v2.10.0 可通过 Homebrew 或 `go install github.com/gogf/gf/v2/cmd/gf@latest` 安装到 `$HOME/go/bin/gf`;`Playwright` 浏览器通过 `cd hack/tests && pnpm exec playwright install` 安装

### 约束

- `LinaPro` 项目处于全新阶段,**无历史兼容包袱**,允许直接增量加入新技能而不考虑旧路径
- 安装入口收敛为 `bootstrap.sh` 单文件下载入口:只负责把仓库源码下载到本地,不再执行 `prereq.sh` 探测、平台脚本分发、运行环境安装或项目初始化动作
- `lina-doctor` 仅服务"开发态环境治理"场景,不替代 `bootstrap.sh` 的"下载仓库源码到本地"职责
- 不引入新的二进制 CLI、不引入新的项目级配置文件,所有动作通过 `bash` + 系统包管理器完成
- 技能必须能被 `Claude Code` / `Codex` 等 AI 工具加载执行,工作流必须支持 AI 代替用户决策与用户单步确认两种模式(由环境变量 `LINAPRO_DOCTOR_NON_INTERACTIVE` 切换)
- 跨平台覆盖三套环境:`macOS` / `Linux` / `Windows-via-Git-Bash-or-WSL`,与 `bootstrap.sh` 一致

### 利益相关方

- **框架终端用户**: 在自己机器上准备 LinaPro 开发环境,期望"AI 一句话把缺的工具装齐",并对每条命令保持知情权
- **框架维护者**: 负责保持 `tool-matrix.md` 中各平台命令时效性,在新工具版本发布时同步更新
- **AI 工具(`Claude Code` / `Codex`)**: 加载 `lina-doctor` 技能,根据 `doctor-check.sh` 输出的 JSON 状态生成结构化安装计划,代替用户执行命令并捕获失败

## Goals / Non-Goals

### Goals

- **G1** 把"开发态环境治理"收敛为单条 AI 指令(如"运行 `lina-doctor`"或"检查我的 LinaPro 环境是否完整"),用户不需要记忆任何具体安装命令
- **G2** 覆盖 `LinaPro` 编码 + 测试链路必需的 9 个工具:`go` / `node` / `pnpm` / `git` / `make` / `openspec` / `gf` / `Playwright browsers` / `goframe-v2` Claude 技能
- **G3** 跨 `macOS` / `Linux` / `Windows-via-Git-Bash-or-WSL` 三平台行为一致;具体安装命令通过包管理器优先级探测自动选定
- **G4** 自动安装强度统一为 L2:每条命令执行前向用户呈现并请求单步确认(可通过环境变量整体跳过),命令执行后立刻复检该工具
- **G5** 拓扑序安装,前置依赖未满足时禁止启动后续步骤(避免"装 `gf` 时报 `go: command not found`")
- **G6** 不引入新的二进制 CLI、不引入新的项目级配置文件,但**承担**`bootstrap.sh` 卸下来的环境探测职责(原 `prereq.sh` 范围 + 新增 `openspec` / `gf` / `Playwright` / `goframe-v2`)
- **G6.5** 把 `hack/scripts/install/` 的脚本职责整体收敛到"git clone 即结束",移除平台分发脚本、`prereq.sh` 与公共库,让 install 与 doctor 两条路径职责完全解耦
- **G7** `SKILL.md` 与 `references/**` 文档使用简体中文编写,便于维护者 review;命令名、环境变量名、JSON 字段名和外部工具标识保持原始英文拼写;仓库根文档双语(中英镜像)
- **G8** 文档详细程度必须支持后续 `/opsx:apply` 时各任务直接落地,不再补设计决策

### Non-Goals

- **NG1** 不在本变更中安装 `MySQL` 服务器或 `MySQL` 客户端工具——用户明确排除该范围,本技能只负责"编码所需的依赖",数据库服务由用户自行部署
- **NG2** 不在本变更中处理 "用户尚未克隆仓库" 的环境(那是 `bootstrap.sh` 的职责);`lina-doctor` 默认假设运行在已克隆的仓库根目录下
- **NG3** 不实现"自动写入 `.zshrc` / `.bashrc` / PowerShell `$PROFILE`",PATH 持久化由用户决定执行;技能只输出可复制的命令片段
- **NG4** 不实现"代理/镜像自动启用",镜像建议仅在 plan 阶段呈现,由用户决定;具体设置仍由用户在自己 shell 中执行
- **NG5** 不为 `Docker` / `VS Code` / `Claude Code` 自身 / `Codex` 等"开发协作工具"提供安装能力;只覆盖 LinaPro 编码 + 测试链路必需工具
- **NG6** 不实现"卸载 / 降级"动作;只做"缺失或版本过低则安装/升级",已满足版本时跳过(幂等)
- **NG7** 不实现 GUI 或 Web 形式的诊断面板;`lina-doctor` 是纯 CLI + AI 编排技能
- **NG8** 不在 `bootstrap.sh` 链路中硬调用 `lina-doctor`;两条路径解耦,避免循环依赖

## Decisions

### D1 技能运行范围: 仅"克隆后"环境治理,不替代 bootstrap.sh

**决策**: `lina-doctor` 假设运行在已克隆的 `LinaPro` 仓库根目录(检测 `apps/lina-core` 与 `apps/lina-vben` 是否存在),不承担"克隆仓库"职责。仅"克隆后"场景纳入;"克隆前/源码下载"由 `bootstrap.sh` 的单文件下载入口负责,但 `bootstrap.sh` 不做运行环境探测和安装。

**为什么**:

- `bootstrap.sh` 单文件设计已承诺"零二级 curl",将 `lina-doctor` 拉进克隆前流程会破坏该约束
- `Playwright` 浏览器安装必须在 `hack/tests/` 内执行 `pnpm exec`,本质上要求仓库已存在
- 用户原始诉求是"安装了框架后,可能缺少运行环境",已显式指向克隆后场景
- 解耦后两个工具职责清晰:`bootstrap.sh` = 下载仓库源码;`lina-doctor` = 工具链治理

**替代方案与拒绝理由**:

- *让 `lina-doctor` 也能克隆仓库*: 与 `bootstrap.sh` 职责重叠,引入双入口分歧;拒绝
- *让 `bootstrap.sh` 在环境缺失时自动拉起 `lina-doctor`*: 破坏单文件下载入口、引入 skill 路径反向依赖;拒绝

### D2 自动安装强度: L2(AI 引导逐项确认)

**决策**: 每条安装命令执行前必须向用户呈现完整命令、所选包管理器与是否需要 sudo,逐项请求 y/N 确认。仅当 `LINAPRO_DOCTOR_NON_INTERACTIVE=1` 时整体跳过确认。

**为什么**:

- 跨 OS 自动装系统级工具是雷区,sudo 失败、网络断、Homebrew 锁、winget 静默失败、apt 包名差异等场景都需要用户视觉确认
- AI 原生定位要求"AI 主导决策但保留用户最终控制权",L2 是这一定位的精准表达
- 与 `lina-upgrade` 的"无冲突全自动,有冲突转人工"形成对偶语义,用户认知负担小

**替代方案与拒绝理由**:

- *L1 仅诊断不安装*: 与现有 `prereq.sh` 价值差异有限,只多了 3 个工具的覆盖范围;拒绝
- *L3 全静默自动安装*: 失败时用户难以介入,装到一半留下半残环境;拒绝

### D3 拓扑序渐进安装,违反顺序立刻中止

**决策**: 安装计划按以下偏序生成,前置工具未满足时拒绝启动后续步骤:

```
git ── (必须最先,否则 git clone 都做不到,但本技能默认仓库已 clone,所以仅校验存在)
│
├── go ──→ gf
├── node ──→ pnpm
│        │
│        ├──→ goframe-v2 skill (npx skills add github.com/gogf/skills -g)
│        └──→ playwright browsers (依赖 pnpm + repo)
├── openspec (macOS brew 通道由 Homebrew 管理 Node 依赖; npm 通道依赖 node)
└── make
```

**为什么**:

- `gf` 通过 `go install` 安装,Go 不在则失败
- `pnpm` 通过 `npm i -g` 安装,Node 不在则失败
- `openspec` 在 macOS 优先通过 `brew install openspec` 安装;Homebrew formula 自身依赖 Node 但由 Homebrew 管理。若 brew 通道不可用则回落 `npm i -g @fission-ai/openspec@latest`,此时依赖 Node ≥ 20.19。Linux/Windows 默认通过 `npm i -g @fission-ai/openspec@latest` 安装,因此依赖 Node ≥ 20.19
- `goframe-v2` 技能通过 `npx skills add ...` 安装,`npx` 来自 Node,因此 Node 不在则失败;无需 pnpm 或仓库
- `Playwright browsers` 通过 `pnpm exec playwright install` 在 `hack/tests/` 内执行,Node + pnpm + 仓库都需就位
- 平铺式一把梭会让单次失败的影响放大到全部步骤

**替代方案与拒绝理由**:

- *并行安装无依赖项以加速*: 实际收益有限(网络下载是瓶颈),代码复杂度显著增加;拒绝
- *把所有依赖关系放到 SKILL.md 的纯文本里让 AI 推断*: 拓扑顺序是确定性逻辑,放到代码里更可靠;拒绝

### D4 包管理器优先级探测,自动选定

**决策**: 各平台按以下优先级探测可用包管理器,选第一个存在的;均不存在则升级到 escalation:

| 平台 | 优先级 |
| --- | --- |
| `macOS` | `brew` |
| `Linux` | `apt-get` > `dnf` > `yum` > `pacman` |
| `Windows` (Git Bash) | `winget` > `scoop` > `choco` |
| `WSL` | 默认按 Linux 分支处理;仅当用户明确希望把工具装到 Windows 主机时,才通过 PowerShell 包装调用 `winget` / `scoop` / `choco` |

**为什么**:

- `apt > dnf > yum` 反映 Ubuntu/Debian → Fedora → CentOS 的市占率,`pacman` 兜底 Arch
- `winget` 是 Windows 11 自带、最稳定;`scoop` / `choco` 兜底
- 不让用户在 plan 阶段回答"用哪个包管理器",减少一轮交互;探测结果在 plan 中明示,用户可在确认阶段否决

**替代方案与拒绝理由**:

- *让 AI 在 plan 阶段交互式询问*: 多一轮问答,常见情形已被默认覆盖;拒绝
- *只支持每个平台一个包管理器*: 装了 `scoop` 没装 `winget` 的 Windows 用户被卡死;拒绝

### D5 PATH 处理: 当前 shell export + 打印 rc-append 命令

**决策**: 装完每个工具后立刻 `command -v <tool>` 验证;若不在 PATH 中:

1. 在当前 `lina-doctor` shell 中 `export PATH="$HOME/go/bin:$PATH"`(仅本进程生效,供后续 `gf` 调用)
2. 向用户打印一行可复制到自己 shell 的 `export` 命令
3. 向用户打印一行可追加到 `.zshrc` / `.bashrc` / PowerShell `$PROFILE` 的命令片段(具体路径按检测到的 shell 决定)
4. **不自动写入用户的 rc 文件**

**为什么**:

- 自动写 rc 文件需要二次确认 + 备份,引入显著复杂度;且 rc 文件是用户私有领域,LinaPro 不应擅自修改
- 仅当前 session export 让 `lina-doctor` 自身后续步骤能继续(`gf` 装完后 `gf version` 验证不会失败)
- 显式打印命令片段保证用户可以"复制 → 粘贴 → 一行搞定",体验透明

**替代方案与拒绝理由**:

- *自动写入 rc 文件 + 备份原文件*: 引入"用 sed 编辑用户 rc"的责任,边界场景多;拒绝
- *不动 PATH,完全靠用户*: `lina-doctor` 自身后续步骤会立刻失败(装 `gf` 后立刻 `gf version` 验证报错);拒绝

### D6 镜像探测: 在 plan 中提示,不擅自启用

**决策**: 在 plan 阶段探测以下环境变量是否已设置:

- `GOPROXY`(空或 `off` 则提示可设 `https://goproxy.cn` / `https://goproxy.io`)
- npm registry(通过 `npm config get registry` 获取,默认 `https://registry.npmjs.org` 时提示可设 `https://registry.npmmirror.com`)
- `PLAYWRIGHT_DOWNLOAD_HOST`(未设时提示可设 `https://npmmirror.com/mirrors/playwright/`)

提示内容仅在 plan 中作为可选项展示,不擅自修改任何环境变量,不要求用户回答"是否启用镜像";用户若需启用,需自行 `export` 后重新运行技能。

**为什么**:

- 镜像选择高度依赖用户网络环境与公司合规要求,默认值很容易踩雷
- LinaPro 不应替用户决定连接哪个公网域名
- 把建议明示但执行权留给用户,既照顾国内用户体验也不冒犯海外用户

**替代方案与拒绝理由**:

- *中国大陆 IP 自动启用国内镜像*: IP 探测不可靠 + 会冒犯一部分用户;拒绝
- *完全不提示*: 国内用户首次使用时网络超时无从下手;拒绝

### D7 doctor-check.sh 输出 JSON, AI 消费

**决策**: `scripts/doctor-check.sh` 输出严格 JSON 给 AI 消费,模式如下:

```json
{
  "os": "macos",
  "package_manager": "brew",
  "shell": "zsh",
  "repo_root_detected": true,
  "tools": {
    "go":              { "present": true,  "version": "1.22.0", "min_version": "1.22.0", "ok": true },
    "node":            { "present": true,  "version": "20.19.0", "min_version": "20.19.0", "ok": true },
    "pnpm":            { "present": false, "version": null, "min_version": "8.0.0", "ok": false },
    "git":             { "present": true,  "version": "2.45.0", "min_version": null, "ok": true },
    "make":            { "present": true,  "version": "3.81",   "min_version": null, "ok": true },
    "openspec":        { "present": true,  "version": "1.3.1",  "min_version": null, "ok": true },
    "gf":              { "present": false, "version": null,     "min_version": null, "ok": false },
    "playwright":      { "present": false, "version": null,     "min_version": null, "ok": false },
    "goframe-v2":      { "present": false, "version": null,     "min_version": null, "ok": false }
  },
  "path_issues": [
    { "tool": "gf", "expected_in": "$HOME/go/bin", "in_path": false }
  ],
  "mirror_hints": [
    { "var": "GOPROXY", "current": "(unset)", "suggested": "https://goproxy.cn" }
  ]
}
```

**为什么**:

- AI 消费结构化数据比解析自由文本可靠得多
- JSON 模式版本化后,后续追加工具不破坏已有消费方
- 同一份 JSON 可同时供 `--check-only` 模式输出与 plan 计算复用

**替代方案与拒绝理由**:

- *输出人类可读文本由 AI 解析*: 文本格式漂移成本大,且诊断与 plan 阶段需要重复解析;拒绝
- *用 yaml / toml 输出*: bash 生成 yaml/toml 比生成 JSON 更易出错;拒绝

### D8 Playwright 与 goframe-v2 默认安装,可显式跳过

**决策**: `Playwright browsers` 与 `goframe-v2` 技能默认纳入安装计划,分别通过 `LINAPRO_DOCTOR_SKIP_PLAYWRIGHT=1` / `LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL=1` 跳过;`--check-only` 模式仍报告其缺失但不计为关键失败。

**为什么**:

- `make test` 是 LinaPro 的主流验证手段,新用户跑 E2E 时会立刻撞上"浏览器未安装";默认装符合期望
- `goframe-v2` 是后端编码规范的权威源(CLAUDE.md 显式要求 Go 编码必须使用),不装会让 AI 工具在写后端代码时缺少必要约束
- Chromium 下载约 150MB,goframe-v2 装到 `~/.claude/skills/` 占用 6.4MB / 962 文件,纯前端开发者或不写 Go 的协作者可能希望跳过
- 标记为"可选项"而非"关键项"避免在 `--check-only` 模式中污染整体诊断结论

**替代方案与拒绝理由**:

- *默认跳过 Playwright / goframe-v2,要求用户显式加 flag*: 新用户首次使用立刻撞墙,体验差;拒绝
- *把 Playwright / goframe-v2 标记为关键项,缺失即认为环境不完整*: 纯前端开发者被强制下载浏览器,纯前端协作者被强制装 Go 编码技能;拒绝

### D8.5 lina-doctor 10 步工作流

**决策**: `SKILL.md` 必须把 `lina-doctor` 的 AI 编排流程写成以下 10 步,确保 Claude Code / Codex 等工具加载技能后行为一致:

1. 确认当前目录是 LinaPro 仓库根目录;若不是,提示用户先进入克隆后的项目目录
2. 读取 `LINAPRO_DOCTOR_*` 环境变量,识别 `--check-only`、跳过项、非交互模式和单步超时
3. 执行 `scripts/doctor-detect.sh`,采集 OS、包管理器、shell、Node 版本管理器、镜像变量和仓库根状态
4. 执行 `scripts/doctor-check.sh`,输出严格 JSON 诊断结果
5. 若 `--check-only` 被启用,直接返回诊断 JSON 和人工建议,不生成安装计划、不执行安装
6. 基于诊断 JSON 执行 `scripts/doctor-plan.sh`,按拓扑序生成结构化安装计划,并展示镜像建议
7. 向用户展示每个计划项的命令、包管理器、sudo/PowerShell 需求、关键/可选属性和跳过原因
8. 执行 `scripts/doctor-install.sh`,按计划逐项确认、安装、复检;关键工具失败即停止,可选目标失败记录非阻塞 escalation
9. 执行 `scripts/doctor-verify.sh`,汇总关键工具 smoke 结果、可选目标状态、PATH 修复提示和镜像建议
10. 输出最终报告:已满足工具、安装成功项、跳过项、非阻塞失败项、下一步 `make init` / `make dev` / `openspec list` / `pnpm test`

**为什么**:

- 技能由 AI 工具加载执行,必须把"AI 何时诊断、何时安装、何时停下"写成确定流程,避免不同工具自由发挥
- 10 步流程把只读诊断、计划生成、用户确认、安装复检、最终报告拆开,便于测试和故障定位

### D9 nvm / fnm / volta 用户的 Node 安装路径

**决策**: 在 `doctor-check.sh` 阶段探测 `$NVM_DIR` / `$FNM_DIR` / `$VOLTA_HOME`(以及 `command -v nvm` 等);若任一存在,Node 的安装命令切换为对应工具的 `install 20`(如 `nvm install 20 && nvm use 20`),不调用 `brew install node` 或 `winget install OpenJS.NodeJS`。

**为什么**:

- `nvm` / `fnm` 用户的 PATH 中通常优先解析它们的 shim,系统包管理器装的 node 反而被遮蔽
- 让用户的现有 Node 版本管理工具继续负责,避免 shim 与系统二进制双轨混乱
- 同样的逻辑不适用于 `gvm` / `g`(Go 的版本管理工具市占率很低,默认仍走 `brew` / `apt` / `winget`)

**替代方案与拒绝理由**:

- *无视 nvm / fnm,统一走包管理器*: 撞 shim 后 `node --version` 仍是旧版,体验混乱;拒绝
- *主动建议用户卸载 nvm / fnm 改用包管理器*: 干涉用户技术栈,不应是 LinaPro 的职责;拒绝

### D10 Windows / WSL 平台调用包管理器的方式

**决策**: Git Bash 中调用 Windows 包管理器统一通过 `powershell.exe -NoProfile -Command "winget install <package>"` 包装,不直接 `winget install`。WSL 默认按 Linux 分支处理,在 WSL 内安装 Linux 工具链;仅当用户明确表达 "我希望工具装到 Windows 主机" 时,才透出 PowerShell 调用 Windows 包管理器。

**为什么**:

- 直接 `winget install` 在 Git Bash 中行为不稳定(参数转义、stdin 处理、Unicode 输出 codec 等)
- `powershell.exe -NoProfile` 显式指定 PowerShell 进程,绕过用户 PowerShell `$PROFILE` 副作用
- WSL 用户的诉求二义性较高,默认选择 WSL 内 Linux 工具链,并在 plan 阶段明示;如用户要装到 Windows 主机,必须显式确认

**替代方案与拒绝理由**:

- *直接 `winget install ...` 让 Git Bash 兜底*: 兼容性问题难以追踪;拒绝
- *只支持 PowerShell 终端运行 lina-doctor*: 与 `bootstrap.sh` 的"Git Bash / WSL"约定冲突;拒绝

### D11 关键工具失败立刻停 + 可选目标非阻塞 escalation

**决策**: 关键工具单步执行失败(包管理器命令非零退出、复检 `command -v` 仍缺失、版本仍低于下限),立刻停止后续步骤,输出结构化 escalation。`Playwright browsers` 与 `goframe-v2` 是可选目标,失败时输出非阻塞 escalation 并继续后续步骤,最终报告中标记为"skipped or failed"。

- 失败工具名
- 选用的包管理器与执行命令
- 命令 stdout / stderr 末 50 行
- 推断的根因类别(`network` / `permission` / `package_not_found` / `shim_conflict` / `unknown`)
- 推荐的人工动作(具体到平台与工具)

**为什么**:

- 关键工具半安装状态比完全未装更难诊断,立刻停止保证用户清楚问题边界
- 可选目标失败不应阻断关键编码环境就绪,否则纯后端/纯前端协作者会被浏览器或 AI 技能下载问题卡住
- 结构化 escalation 让 AI 工具能继续协作(比如把 stderr 喂给 AI 让它建议修复)

**替代方案与拒绝理由**:

- *关键工具失败时仍然继续后续工具*: 后续工具失败级联报错,真正问题被淹没;拒绝
- *只输出文本错误信息*: AI 无法精准提取根因;拒绝

### D12 幂等执行: 已满足版本立刻跳过

**决策**: `doctor-check.sh` 标记为 `ok: true` 的工具一律跳过 install 阶段;不重装、不升级、不降级。版本下限以 `tool-matrix.md` 为权威源,具体工具:

| 工具 | 版本下限 |
| --- | --- |
| `go` | `1.22` |
| `node` | `20.19` |
| `pnpm` | `8` |
| `git` / `make` / `openspec` / `gf` | (仅检查存在性,不卡版本) |
| `Playwright browsers` | (仅检查 `chromium` 是否已下载) |
| `goframe-v2` | (仅检查 `~/.claude/skills/goframe-v2/SKILL.md` 是否存在) |

**为什么**:

- 重复运行 `lina-doctor` 必须是无副作用操作,符合"诊断 + 修复"语义
- `git` / `make` / `openspec` / `gf` / `goframe-v2` 当前没有明显需要卡下限的场景,过度约束会让用户反复升级

**替代方案与拒绝理由**:

- *统一卡所有工具的最新版本下限*: 用户被迫频繁升级,体验差;拒绝
- *允许 `--force` 重装*: 增加复杂度,LinaPro 不需要"重装"语义;拒绝

### D13 install 脚本职责收敛: bootstrap.sh 仅负责下载仓库源码

**决策**: `hack/scripts/install/bootstrap.sh` 收敛为"解析目标版本 + git clone 下载仓库源码 + 打印 next-steps"的单文件入口;平台判断仅用于给出不支持环境的快速失败提示,**完全不再触发**任何运行环境探测、依赖安装、项目初始化、端口探测。三个平台分发脚本(`install-macos.sh` / `install-linux.sh` / `install-windows.sh`)、`prereq.sh`、`lib/_common.sh` 整体从仓库删除。次发布的 `https://linapro.ai/install.sh` 内容同步更新。

**为什么**:

- 把"安装框架代码"与"准备开发环境"两件本应解耦的事情绑在一起,导致 `make init` 在 MySQL 未配置时失败,让用户连仓库克隆是否成功都不易判断
- `lina-doctor` 既然已经承担环境治理职责,`prereq.sh` 与平台脚本的环境探测逻辑就是双份维护
- `make init` / `make mock` / `go mod download` / `pnpm install` 都是已有 Make 目标或独立命令,用户在 `lina-doctor` 检查环境通过后显式调用即可,**不需要 install 链路代劳**
- 单一脚本(`bootstrap.sh`)更便于 CDN 部署与 review;减少对 `lib/_common.sh` 的依赖也让 `bootstrap.sh` 自包含语义更彻底

**替代方案与拒绝理由**:

- *保留 `prereq.sh` 但仅作为"提醒用户调用 lina-doctor"的薄壳*: 双份维护成本仍在,且 install 路径仍隐含 skill 路径依赖;拒绝
- *保留平台脚本但只跑 `make init`*: 数据库初始化对很多本机部署场景(用户尚未配 MySQL)是阻断点,默认拉起会让首次安装失败率显著上升;拒绝
- *把 `make init` 的触发权交给 `lina-doctor` 而非用户*: doctor 是环境治理工具,不应负责数据库初始化语义;拒绝

### D14 goframe-v2 技能外部化: 通过 npx skills add 全局安装

**决策**: 仓库内 `.claude/skills/goframe-v2/` 整个目录从 git 历史中移除;`lina-doctor` 在拓扑序中的"goframe-v2"环节通过 `npx skills add github.com/gogf/skills -g` 安装到用户全局 `~/.claude/skills/goframe-v2/`。该步骤依赖 Node + npx 已就绪。

**为什么**:

- `goframe-v2` 体量 6.4MB / 962 文件,`git clone` 与 `git pull` 时显著增加 IO,且对不写 Go 的协作者完全冗余
- 该技能的源头来自 `https://github.com/gogf/skills`,LinaPro 仓库内的副本只是镜像,无修改;镜像反而增加同步上游的成本
- 装到用户全局 `~/.claude/skills/` 是 Claude Code skill 生态的标准做法,跨项目共享一份
- `npx skills add` 是上游推荐的安装通道,统一通过它可以保证跟随 `gogf/skills` 仓库的最新 release 节奏

**替代方案与拒绝理由**:

- *继续保留仓库自带,加 `git lfs` 优化 IO*: 仍然是仓库膨胀,且 `git lfs` 引入额外配置;拒绝
- *从 `gogf/skills` 仓库通过 `git clone` 直接装到项目本地 `.claude/skills/`*: 与 `npx skills add -g` 全局装相比缺少版本管理,且占用项目目录;拒绝
- *作为 LinaPro 自建 npm 包发布*: 维护责任错位,`goframe-v2` 是 GoFrame 上游的产物;拒绝

## Risks / Trade-offs

### Risk 1: nvm / fnm / volta 探测可能漏网

**风险**: 用户使用其他 Node 版本管理工具(如 `n`、`asdf`)未被探测到,仍走包管理器路径,导致 shim 冲突。

**缓解**:
- `references/troubleshooting.md` 中列出 5 种常见 Node 版本管理工具及其特征
- 探测优先级覆盖 `nvm` / `fnm` / `volta`,其他工具兜底告知用户"检测到 PATH 中已有 node 但版本过低,建议手动升级"
- 不擅自覆盖 PATH 中已有的 node 二进制

### Risk 2: 网络环境差导致单步超时

**风险**: 国内用户首次使用,GOPROXY 未设、`go install gf` 卡 `proxy.golang.org` 超时,但 `lina-doctor` 已经在 install 阶段。

**缓解**:
- D6 决策在 plan 阶段就探测并提示镜像
- 单步超时(默认 5 分钟)后归类到 `network` 根因,escalation 中明示镜像建议
- `LINAPRO_DOCTOR_TIMEOUT` 环境变量允许用户调整单步超时

### Risk 3: 不同 Linux 发行版的包名差异

**风险**: `apt-get install` 在 Ubuntu 装 `golang-go` 但 Debian 较老版本是 `golang`;`mysql-client` 在 Debian 12 改名 `default-mysql-client`。

**缓解**:
- 本变更不安装 `mysql client`,缩小该风险面
- `tool-matrix.md` 按 `apt-get` / `dnf` / `yum` / `pacman` 维护具体包名,Debian/Ubuntu 不区分版本统一用 `golang-go` 与 `nodejs`,不行则告知用户回落到官方二进制
- 未识别的发行版统一回落"打印官方下载链接 + 让用户手动",不强行 install

### Risk 4: 与 lina-upgrade 技能的隐性耦合

**风险**: `lina-upgrade` 的工作流默认 `gf` / `openspec` 已就绪,若用户跳过 `lina-doctor` 直接调用 `lina-upgrade`,失败信息不够友好。

**缓解**:
- `lina-upgrade/SKILL.md` 在 prerequisites 一节追加"如缺 `gf` / `openspec`,先调用 `lina-doctor`"
- 仅文档级软依赖,不在脚本层强引用,避免循环依赖

### Risk 5: brew openspec formula 的稳定性

**风险**: `brew install openspec` 当前在维护者本机可用,但如果 formula 失效或本机 Homebrew 无法解析,macOS 用户会被卡在 OpenSpec 安装环节。

**缓解**:
- 在实施阶段通过 `brew info openspec` 确认 formula 来源(是 homebrew/core 还是 third-party tap)
- 如果 Homebrew 通道失败,统一回落官方 npm 包 `@fission-ai/openspec`(此时 Node ≥ 20.19 已就绪)
- 在 `tool-matrix.md` 的 macOS 列中维护"brew 通道为主,失败则 npm 兜底"的语义,并明确 `openspec@0.0.0` 不得使用

### Risk 6: install 脚本职责收敛后的用户首次安装体验降级

**风险**: 本变更前 `curl | bash` 一条命令完成"克隆 + `make init` + `make mock`"全套首次安装;本变更后只克隆,用户必须再显式执行 `make init`。习惯了"一条命令搞定"的用户可能感到体验降级。

**缓解**:
- `bootstrap.sh` 克隆完成后输出的 next-steps 提示必须显著、可复制,清晰说明"clone 已成功 → 调用 lina-doctor → 跑 make init"三步
- `README.md` / `README.zh_CN.md` 把"首次安装"章节重写为三步指引,而非单行 curl
- `lina-doctor` 验收阶段输出"All tools satisfied. Next: cd <project>; make init; make dev"作为闭环引导,降低用户认知断点
- 长远视角:解耦后用户对 `make init` 失败有更清晰的归因(MySQL 未配 / DDL 异常),反而改善了首次安装的可调试性

### Risk 7: `gogf/skills` 上游不稳定或 npx skills add 命令不可用

**风险**: `npx skills add github.com/gogf/skills -g` 是上游 `gogf/skills` 仓库的安装通道,若 `skills` npm 包名变更、命令签名改动或仓库下线,`lina-doctor` 的 goframe-v2 安装步骤会立刻失败。

**缓解**:
- `references/troubleshooting.md` 中维护"上游通道失效"的人工兜底:`git clone https://github.com/gogf/skills ~/.claude/skills/skills && link goframe-v2`
- `doctor-check.sh` 检查 `~/.claude/skills/goframe-v2/SKILL.md` 是否存在作为最终验证标准,与具体安装通道解耦
- escalation 输出中明示"goframe-v2 安装失败时不影响 LinaPro 编码,只影响 AI 工具触发 GoFrame 编码规范"——降低该步骤失败的阻塞感

### Risk 8: 安装入口变更 baseline 需要以 lina-doctor 为准

**风险**: `linapro-install-upgrade` 变更已被维护者手动删除,本迭代成为唯一活跃的安装入口变更。若文档或任务继续保留对旧变更的等待约束,会导致实施流程被已不存在的前置条件阻塞。

**缓解**:
- 以 `lina-doctor` 作为 `framework-bootstrap-installer` 能力调整的权威活跃变更
- `lina-doctor` 的 `framework-bootstrap-installer` delta 直接覆盖安装入口职责收敛、`prereq.sh`、平台脚本、`lib/_common.sh`、`LINAPRO_SKIP_MOCK` 等条目的最终目标状态
- `/opsx:apply lina-doctor` 启动后必须跑 `openspec validate lina-doctor --strict`,确认当前工作树 baseline 下无 spec 偏差

## Migration Plan

### 阶段 1: 确认当前 baseline

- 确认 `linapro-install-upgrade` 已不再是活跃 OpenSpec 变更
- 以当前工作树中的 `lina-doctor` 作为安装入口职责收敛的权威变更
- 重新 `openspec validate lina-doctor --strict`,处理任何当前 baseline 下的 spec 偏差

### 阶段 2: install 脚本职责收敛

- 删除 `hack/scripts/install/install-macos.sh` / `install-linux.sh` / `install-windows.sh` / `checks/prereq.sh` / `lib/_common.sh`
- 重写 `hack/scripts/install/bootstrap.sh` 为"探测 OS + 解析版本 + git clone + 打印 next-steps"四步
- 同步精简 `hack/scripts/install/README.md` / `README.zh_CN.md`
- 重写 `hack/tests/e2e/install/TC0155-TC0157` 用例,只校验克隆行为,移除对 `make init` / `make mock` / `LINAPRO_SKIP_MOCK` 的断言
- 部署一次更新版 `https://linapro.ai/install.sh`(运维任务,非 PR 范围)

### 阶段 3: lina-doctor 技能落地

- 创建 `.claude/skills/lina-doctor/` 完整目录结构
- 把原 `lib/_common.sh` 中的纯工具函数(`version_ge` / `confirm` / `retry` 等)迁移到 `.claude/skills/lina-doctor/lib/_common.sh`(若需要)
- 完成 `doctor-detect.sh` / `doctor-check.sh` / `doctor-plan.sh` / `doctor-install.sh` / `doctor-verify.sh` 五个核心脚本与 `lib/doctor-escalate.sh`
- 完成 `references/` 四份文档(包含 `goframe-v2` 安装说明)
- 通过 lina-e2e 技能补齐三平台的诊断与安装 E2E 测试用例(macOS smoke 必跑,Linux/Windows 在 CI 自动化范围外作为手动验收)

### 阶段 4: goframe-v2 仓库清理

- 从 git 历史中删除 `.claude/skills/goframe-v2/` 整个目录(`git rm -r`)
- 更新 `CLAUDE.md` 对 `goframe-v2` 的引用文本,从"项目仓库自带"改为"通过 `lina-doctor` 安装到用户全局技能目录后即可触发"
- 验证拉取本变更后的 fresh clone 不再包含 `.claude/skills/goframe-v2/`

### 阶段 5: 文档与索引同步

- `README.md` / `README.zh_CN.md` / `CLAUDE.md` 同步更新
- `lina-upgrade/SKILL.md` 追加 prerequisites 指引

### 阶段 6: 用户验收

- 在维护者本机重置环境(卸载 `gf` / `openspec` / `Playwright browsers` / `~/.claude/skills/goframe-v2`)→ 调用 `lina-doctor` → 验证全部装回
- 在干净 macOS / Ubuntu / Windows 11 Git Bash 环境上各跑一次"全新克隆 → lina-doctor → make init → make dev"完整流程,记录人工验收清单
