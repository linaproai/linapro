## Why

`LinaPro` 当前安装入口围绕"克隆仓库 + 跑 `make init` / `make mock`"收敛为单条 `curl | bash` 命令,并提供了 `hack/scripts/install/checks/prereq.sh` 做工具探测。但当前路径在职责划分与覆盖范围上仍有三个明显问题:

- **install 脚本职责过重**:`bootstrap.sh` 不仅克隆仓库,还分发到 `install-macos.sh` / `install-linux.sh` / `install-windows.sh` 三个平台脚本,后者再触发 `prereq.sh` 环境探测、`go mod download`、`pnpm install`、`make init`、`make mock`、端口探测等动作,把"安装框架代码"与"准备开发环境"两件本应解耦的事情绑在一起。一旦 `make init` 因 MySQL 未配置而失败,用户连仓库克隆是否成功都无法快速判断。
- **现有 `prereq.sh` 只检查 `go` / `node` / `pnpm` / `git` / `make` / `mysql` 是否存在,缺失时只打印安装提示文本,完全不动手安装**——用户仍需自己跑 `brew install go` / `apt-get install make` / `winget install OpenJS.NodeJS` 等命令,跨平台、跨包管理器的踩坑全数留给用户。
- **现有探测项不覆盖"AI 原生开发流程"硬依赖**:`openspec` CLI(SDD 工作流)、`gf` CLI(`make dao` / `make ctrl` 必须)、`Playwright` 浏览器驱动(`make test` 必须)、`goframe-v2` Claude 技能(后端编码规范权威源,体量 6.4MB / 962 文件)四件套都没纳入,用户安装完仓库后第一次跑 `make dao` / `make test` / 编写后端代码仍会立刻撞墙。

本次变更做两件事:

1. 把 `hack/scripts/install/` 的脚本职责**收敛为"git clone 即结束"**——`bootstrap.sh` 完成克隆后直接打印"Next steps: 调用 `lina-doctor` 检查环境 → 跑 `make init`",**移除**所有平台分发脚本、`prereq.sh` 环境探测、`lib/_common.sh` 公共库与项目初始化动作。
2. 新增 `lina-doctor` 技能,作为 `LinaPro` 的"环境医生":AI 主导诊断 → 列出待装清单 → 用户单步确认 → 拓扑序逐项安装 → 装完立刻复查 → 给出可复制的 `PATH` 修复命令。把"装好框架运行环境"这件事整体纳入 AI 原生工作流,与 `lina-upgrade` 形成完整的开发态生命周期闭环(诊断 → 升级)。

`goframe-v2` 技能因为体量过大,不再作为 LinaPro 仓库自带,改为通过 `lina-doctor` 用 `npx skills add github.com/gogf/skills -g` 命令按需安装到用户全局技能目录。

## What Changes

### 简化 `hack/scripts/install/` 脚本职责

- **BREAKING** 移除 `hack/scripts/install/install-macos.sh` / `install-linux.sh` / `install-windows.sh` 三个平台分发脚本;`bootstrap.sh` 不再 `exec` 任何二级脚本
- **BREAKING** 移除 `hack/scripts/install/checks/prereq.sh`,环境探测职责整体迁移到 `lina-doctor` 技能
- **BREAKING** 移除 `hack/scripts/install/lib/_common.sh` 与 `checks/` 目录;若有公共 bash 工具函数被 `lina-doctor` 复用,迁移到 `.claude/skills/lina-doctor/lib/_common.sh`
- **修改** `bootstrap.sh` 收敛为"探测 OS + 解析版本 + git clone + 打印 next-steps"四步,**完全不再触发** `go mod download` / `pnpm install` / `make init` / `make mock` / 端口探测等项目初始化动作
- **修改** 克隆完成后的 next-steps 文本统一为三行指引:
  1. `cd <project-dir>`
  2. 调用 AI 工具加载 `lina-doctor` 技能完成环境配置(英文示例:`ask Claude Code "run lina-doctor to set up my LinaPro environment"`)
  3. `make init`(数据库初始化) → `make dev`(启动开发环境)

### 新增 `lina-doctor` 技能

- **新增** `.claude/skills/lina-doctor/` 目录,作为 `Claude Code` / `Codex` 等 AI 工具的可加载技能
- **新增** `SKILL.md`(中文 frontmatter 描述 + 中文工作流说明,保留必要英文触发词),触发关键词覆盖"环境检查 / 安装依赖 / 缺少 go/node/openspec / 工具链 / lina-doctor / install dependencies / fix environment / setup linapro environment"
- **新增** 9 个工具的诊断 + 自动安装能力,覆盖三平台:
  - `go ≥ 1.22`、`node ≥ 20.19`、`pnpm ≥ 8`、`git`、`make`
  - `openspec`(SDD 工作流硬依赖,`brew install openspec` 或 `npm i -g @fission-ai/openspec@latest`)
  - `gf` GoFrame CLI(`go install github.com/gogf/gf/v2/cmd/gf@latest`,需要 Go 已就绪)
  - `Playwright` 浏览器(`pnpm exec playwright install`,需要 Node + pnpm + 仓库已克隆)
  - `goframe-v2` Claude 技能(`npx skills add github.com/gogf/skills -g`,需要 Node 已就绪;装到用户全局 `~/.claude/skills/` 而非项目仓库)
- **新增** 拓扑序渐进安装(`git → go → gf`、`node → pnpm`、`node → goframe-v2`、`pnpm + repo → playwright`;`openspec` 在 macOS 优先走独立的 brew 通道,否则走依赖 Node 的 npm 通道),禁止平铺式一把梭
- **新增** AI 引导的逐项确认安装(L2 模式):每条命令执行前都呈现给用户、单独确认、单独验证;`LINAPRO_DOCTOR_NON_INTERACTIVE=1` 时跳过确认
- **新增** 包管理器优先级探测(macOS 仅 `brew`;Linux 按 `apt > dnf > yum > pacman`;Windows 按 `winget > scoop > choco`),实际探测后自动选定,不强制让用户回答
- **新增** `PATH` 自检:装完 `gf` 后探测 `$GOBIN` / `$HOME/go/bin` 是否在 `PATH`,缺失时打印当前 shell 的 `export` 一行 + 持久化所需的 `.zshrc` / `.bashrc` / PowerShell `$PROFILE` 追加片段;**默认仅在当前 shell 内 export,不自动写入用户 rc 文件**
- **新增** 镜像探测:检测 `GOPROXY` / npm registry / `PLAYWRIGHT_DOWNLOAD_HOST` 是否已配置;未配置时在 plan 中提示常见镜像选项(`https://goproxy.cn`、`https://registry.npmmirror.com`、`https://npmmirror.com/mirrors/playwright/`),由用户决定是否启用,**技能不擅自切换默认值**
- **新增** `Playwright` 浏览器与 `goframe-v2` 技能作为可选项:默认安装,分别可通过 `LINAPRO_DOCTOR_SKIP_PLAYWRIGHT=1` / `LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL=1` 跳过
- **新增** `--check-only` 模式:仅诊断不安装,输出统一的 JSON 状态供 AI 消费;由 `lina-doctor` 完整覆盖原 `prereq.sh` 的探测语义并补全工具范围

### 移除 `goframe-v2` 仓库自带形态

- **BREAKING** 移除仓库内 `.claude/skills/goframe-v2/` 目录(6.4MB / 962 文件),从 git 历史中删除
- **新增** `lina-doctor` 中的 `goframe-v2` 安装项,通过 `npx skills add github.com/gogf/skills -g` 装到用户全局 `~/.claude/skills/goframe-v2/`
- **修改** `CLAUDE.md` 中"后端代码规范"章节对 `goframe-v2` 技能的引用文本,从"项目仓库自带"改为"通过 `lina-doctor` 安装到用户全局技能目录后即可触发"

### 文档与索引

- **修改** `README.md` / `README.zh_CN.md`:
  - "安装"章节末尾追加一句"如缺工具请通过 AI 工具调用 `lina-doctor` 技能",不展开命令细节
  - "Next steps"段补充三行指引(同上 `bootstrap.sh` 输出),与脚本输出保持一致
- **修改** `CLAUDE.md`:在 skill 列表 / 常用命令章节登记 `lina-doctor`,与 `lina-upgrade` 平级;同步更新 `goframe-v2` 引用文本
- **修改** `.claude/skills/lina-upgrade/SKILL.md`:在 prerequisites 一节追加"如缺 `gf` / `openspec`,先调用 `lina-doctor`"
- **新增** `.claude/skills/lina-doctor/references/tool-matrix.md`:9 个工具 × 3 平台的安装命令权威源
- **新增** `.claude/skills/lina-doctor/references/install-strategy.md`:拓扑序、包管理器选择、版本约束规则
- **新增** `.claude/skills/lina-doctor/references/path-and-shell.md`:`$GOBIN` / npm prefix / Windows `$env:Path` 的处理指南
- **新增** `.claude/skills/lina-doctor/references/troubleshooting.md`:nvm 撞车、winget 静默失败、网络代理、`npx skills add` 网络失败常见错误清单

## Capabilities

### New Capabilities

- `framework-environment-doctor`: AI 主导的开发态环境诊断与自动安装能力,覆盖 9 个工具 × 3 平台的拓扑序安装、单步确认、`PATH` 自检与镜像建议

### Modified Capabilities

- `framework-bootstrap-installer`: 安装入口的职责边界**显著收敛**——`bootstrap.sh` 仅负责 git clone 与 next-steps 提示,移除平台分发脚本、`prereq.sh` 环境探测、`lib/_common.sh` 公共库,以及 `go mod download` / `pnpm install` / `make init` / `make mock` / 端口探测等所有项目初始化动作

## Impact

### 受影响代码

- `.claude/skills/lina-doctor/`(全新目录,首次创建)
- `.claude/skills/lina-upgrade/SKILL.md`(在 prerequisites 段追加一句对 `lina-doctor` 的指引)
- `.claude/skills/goframe-v2/`(整个目录从仓库删除,改由 `lina-doctor` 调用 `npx skills add github.com/gogf/skills -g` 安装到用户全局 `~/.claude/skills/`)
- `hack/scripts/install/bootstrap.sh`(职责收敛:仅 git clone + 打印 next-steps;移除 `exec bash hack/scripts/install/install-${os_name}.sh` 分发逻辑)
- `hack/scripts/install/install-macos.sh`(整个文件删除)
- `hack/scripts/install/install-linux.sh`(整个文件删除)
- `hack/scripts/install/install-windows.sh`(整个文件删除)
- `hack/scripts/install/checks/prereq.sh`(整个文件删除,目录连同 `.gitkeep` 一并清理)
- `hack/scripts/install/lib/_common.sh`(整个文件删除;若 `lina-doctor` 需要复用 `version_ge` / `detect_os` / `confirm` / `retry` 等纯工具函数,迁移到 `.claude/skills/lina-doctor/lib/_common.sh`)
- `hack/tests/scripts/install-bootstrap.sh`(若该测试 fixture 包含对 `prereq.sh` / `install-*.sh` 的断言,同步精简为仅校验 `bootstrap.sh` 的克隆行为)

### 受影响文档

- `README.md` / `README.zh_CN.md`(根目录,更新"安装"章节与"Next steps"段)
- `CLAUDE.md`(常用命令、研发流程章节,同步登记 `lina-doctor` 与 `goframe-v2` 引用)
- `hack/scripts/install/README.md` / `README.zh_CN.md`(整体瘦身,只保留 `bootstrap.sh` 的语义说明,移除平台脚本与 prereq.sh 章节)
- `.agents/instructions/` / `.agents/prompts/`(若有提及"环境准备"段落同步更新)

### 受影响测试

- `hack/tests/e2e/install/TC0155-install-default.ts` / `TC0156-install-version-override.ts` / `TC0157-install-skip-mock.ts`:由于 `LINAPRO_SKIP_MOCK` 不再存在,且 `make mock` 不再由 install 链路触发,该用例需要重写为只验证 `bootstrap.sh` 的克隆行为;`LINAPRO_SKIP_MOCK` 环境变量整体移除

### 受影响外部资源

- `https://linapro.ai/install.sh`(CDN 静态托管):`bootstrap.sh` 内容收敛后必须重新部署一次,确保 CDN 与仓库同步

### i18n 评估

本变更**不影响**前端运行时语言包、宿主/插件运行时 `manifest/i18n` 资源以及 `apidoc i18n JSON`。`lina-doctor` 的 `SKILL.md` 与 `references/**` 文档使用简体中文编写,便于维护者 review;命令名、环境变量名、JSON 字段名和外部工具标识保持原始英文拼写,保证脚本解析与 AI 工具消费稳定。仓库根 `README.md` / `README.zh_CN.md` 与 `CLAUDE.md` 按"目录级主说明文档统一英文 + 中文镜像"规范同步维护即可。技能内部不引入任何 UI 元素、API DTO、菜单/路由/按钮/表单/表格、消息键或 apidoc 翻译资源,因此无需新增、修改或删除任何 `i18n JSON` 文件。

### 缓存一致性评估

本变更**不涉及**任何运行时缓存层(权限、配置、插件状态、租户隔离、字典、路由、国际化资源、apidoc catalog 等)。`lina-doctor` 是一次性的开发态环境治理工具,执行链路与宿主进程内缓存、分布式协调组件、`cluster.Service` 拓扑均无交集,因此无需评估单机/分布式部署下的一致性、失效作用域或最大可接受陈旧时间。

### 兼容性与升级路径

- 项目处于全新阶段,无历史兼容包袱,允许直接断更 `prereq.sh` 与三个平台 install 脚本,不保留向后兼容别名
- 安装路径 BREAKING:本变更落地后,`curl -fsSL https://linapro.ai/install.sh | bash` 仅完成 git clone,不再自动跑 `make init` / `make mock`;用户必须显式 `make init` 才能启动后端
- `LINAPRO_SKIP_MOCK` 环境变量被一并移除(不再有 install 链路触发 `make mock`);`make mock` 仍可独立调用
- `lina-upgrade` 技能调用 `lina-doctor` 是**软依赖**(文档级指引),不在脚本中硬连接,避免技能间循环引用
- `goframe-v2` 技能从仓库自带改为按需全局安装:已克隆仓库的现有用户在拉取本变更后会发现 `.claude/skills/goframe-v2/` 消失,需要执行一次 `lina-doctor` 把它装回到 `~/.claude/skills/`(用户全局)

### 风险

- **R1** `nvm` / `fnm` / `volta` 用户的 Node 安装冲突:包管理器装的 `node` 与 `nvm shim` 可能在 PATH 中互相覆盖。缓解:`doctor-check.sh` 探测 `$NVM_DIR` / `$FNM_DIR` / `$VOLTA_HOME`,存在则在 plan 中改用 `nvm install 20` / `fnm install 20` 命令,不调用 `brew install node` / `winget install OpenJS.NodeJS`
- **R2** `gf` 装在 `$GOBIN` 但 `$GOBIN` 未导出:`go install` 默认输出到 `$HOME/go/bin`,但用户 shell 配置可能未把它加入 `PATH`。缓解:装完立刻 `command -v gf` 验证,不在 `PATH` 时打印 export 一行 + rc-append 一行,失败不阻塞但醒目提示
- **R3** Windows Git Bash 调用 `winget`:`winget` 是 PowerShell 上下文工具,从 Git Bash 直接 `winget install` 在某些 Windows 版本上行为不稳定。缓解:Windows 分支统一通过 `powershell.exe -NoProfile -Command "winget install ..."` 调用,显式指定 PowerShell 进程
- **R4** `openspec` 的 brew formula 维护稳定性:用户机上 `openspec` 可来自 Homebrew(`/opt/homebrew/bin/openspec`),官方 npm 通道为 `@fission-ai/openspec`;若 brew 通道不稳定,统一回落 `npm i -g @fission-ai/openspec@latest`(此时依赖 Node ≥ 20.19 已就绪)
- **R5** 网络代理/镜像踩坑:中国大陆用户访问 `proxy.golang.org` / `registry.npmjs.org` / `playwright-cdn` 可能超时。缓解:plan 阶段探测并提示,但不擅自启用镜像;具体镜像 URL 在 `references/troubleshooting.md` 中维护
- **R6** `sudo` 权限要求:Linux `apt-get install` 通常需要 `sudo`,`npm i -g` 在某些 prefix 下也需要 `sudo`。缓解:plan 中显式标注哪些命令需要 `sudo`,让用户在确认阶段就知情;不擅自包装 `sudo`,由用户决定执行身份
