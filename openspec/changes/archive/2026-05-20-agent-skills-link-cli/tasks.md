## 1. 子组件实现

- [x] 1.1 新增`hack/tools/linactl/internal/skilllink/`子组件，定义`AgentSpec`/`Category`/内置`vercel-labs/skills`项目路径映射表，覆盖`native`/`link`/`rootCollision`三类
- [x] 1.2 实现`Plan(root, selectors)`/`Apply(root, plan, force)`/`Unlink(root, selectors)`核心函数，使用`os.Symlink`/`os.Readlink`/`os.Lstat`/`os.Remove`/`os.MkdirAll`和`filepath.Rel`生成相对软链
- [x] 1.3 增加单元测试`skilllink_test.go`覆盖：`native`跳过、`link`首次创建、`link`再次执行幂等、`mismatch`无`force`保护、`mismatch`+`force`重建、普通目录冲突拒绝、`rootCollision`默认跳过、`unlink`仅移除受管软链

## 2. `linactl`命令接入

- [x] 2.1 新增`hack/tools/linactl/command_skills.link.go`，注册`skills.link`命令，解析`agent`/`force`参数并调用`skilllink.ApplyLink`
- [x] 2.2 新增`hack/tools/linactl/command_skills.unlink.go`，注册`skills.unlink`命令，解析`agent`参数并调用`skilllink.ApplyUnlink`
- [x] 2.3 更新`hack/tools/linactl/command.go`注册新命令并补齐`Description`/`Usage`
- [x] 2.4 实现列表化输出渲染（含`AGENT`/`PROJECT PATH`/`CATEGORY`/`STATUS`/`DETAIL`列与末尾 hint），并保证`error`类状态导致命令非零退出
- [x] 2.5 `Windows`下软链失败时输出开发者模式或管理员提示文案

## 3. `Make`包装与仓库引用

- [x] 3.1 新增`hack/makefiles/skills.mk`，提供`skills.link`/`skills.unlink`目标并转发`AGENT`/`FORCE`参数
- [x] 3.2 在根`Makefile`中`include hack/makefiles/skills.mk`
- [x] 3.3 在`.gitignore`中按内置映射表追加常见`AGENT`目录`/skills`忽略规则，避免软链入库

## 4. 文档更新

- [x] 4.1 更新`hack/tools/linactl/README.md`和`hack/tools/linactl/README.zh-CN.md`，新增`skills.link`/`skills.unlink`使用说明、`AGENT`/`FORCE`参数表与示例
- [x] 4.2 在文档中明确仅管理仓库内项目路径软链、不操作`HOME`目录、不下载远端`Skill`源码

## 5. 验证与审查

- [x] 5.1 运行`cd hack/tools/linactl && go test ./... -count=1`
- [x] 5.2 运行`cd hack/tools/linactl && go run . skills.link`、`cd hack/tools/linactl && go run . skills.link agent=codebuddy,qoder`、`cd hack/tools/linactl && go run . skills.unlink agent=codebuddy,qoder,windsurf,roo`等公开入口烟测
- [x] 5.3 运行`make skills.link AGENT=claude-code`验证`Make`包装层；通过 mismatch + `force=1` 重建场景和 conflict（含 sentinel.md 保护）场景验证`rootCollision`/`force`分支逻辑
- [x] 5.4 运行`git diff --check`、`go build ./...`和`go test ./... -count=1`确保无白空格问题与编译错误；`openspec validate`工具未在本机安装，通过文档结构静态校验和已归档变更模板对齐验证
- [x] 5.5 评估`i18n`、缓存一致性、数据权限、开发工具脚本影响并在`Implementation Notes`中记录，待`/lina-review`触发后补充审查结论

## 6. 交互式选择模式

- [x] 6.1 在`internal/skilllink`新增`IsInteractiveTerminal`/`LinkCandidates`/`UnlinkCandidates`/`PromptSelection`/`PromptYesNo`，仅依赖`Go`标准库（`os.File.Stat`+`ModeCharDevice`、`bufio.Reader`），不引入第三方交互库
- [x] 6.2 修改`command_skills.link.go`：未传`agent`且 stdin 是真实终端时进入交互式选择；mismatch 候选询问是否启用 force；非终端保持只读列表
- [x] 6.3 修改`command_skills.unlink.go`：未传`agent`且 stdin 是真实终端时进入交互式选择，仅列出受管软链；非终端保持必须显式`AGENT=`的非零退出
- [x] 6.4 新增`skilllink_interactive_test.go`覆盖交互式用例：逗号选择、`all`、取消（空/`q`/空白）、空候选、越界/非数字、去重、`yes/no`默认值、`yes/no`显式输入与非法输入、3 列网格渲染、所有`Status`到状态符号的映射
- [x] 6.5 通过`script -q -c "make skills.link" /dev/null < <(printf '...')` 模拟 TTY，验证编号渲染、选择执行与`q`取消行为
- [x] 6.6 同步更新`hack/tools/linactl/README.md`、`README.zh-CN.md`和`OpenSpec`提案/设计/规范文档，明确交互式与非交互回退策略
- [x] 6.7 将候选清单改为 3 列网格 + 单字符状态符号 + 图例排版，将 39 个`link`类 Agent 压缩到 17 行内一屏显示，使用 ASCII 符号避免`Unicode`字符宽度问题
- [x] 6.8 新增`linactl skills`聚合命令与`make skills`包装目标，TTY 下展示`link`/`unlink`/`quit`操作菜单并分发到对应子命令交互式；非 TTY 打印用法指引

## Implementation Notes

- 2026-05-20：完成`agent-skills-link-cli`首版实现。新增`hack/tools/linactl/internal/skilllink`子组件，集中维护`vercel-labs/skills`官方项目路径映射表（55 个 Agent，分类为 15 个`native` / 39 个`link` / 1 个`rootCollision`），并提供`PlanList`/`ApplyLink`/`ApplyUnlink`核心 API。新增`linactl skills.link`/`skills.unlink`命令以及`hack/makefiles/skills.mk`包装目标，根`Makefile`通过`include`接入。`.gitignore`新增 38 条 Agent 项目路径软链忽略规则，避免本地创建后污染仓库。`hack/tools/linactl/README.md`、`README.zh-CN.md`同步增加“Agent 技能软链管理”章节，明确仅管理仓库内项目路径软链、不操作`HOME`目录、不下载远端`Skill`源码。验证通过：`cd hack/tools/linactl && go test ./... -count=1`、`cd hack/tools/linactl && go build ./...`、`cd hack/tools/linactl && go run . skills.link`默认列表、`go run . skills.link agent=codebuddy,qoder`创建+幂等、`go run . skills.link agent=windsurf`mismatch + `force=1`重建、真实目录冲突（含 sentinel.md 保护）、`go run . skills.unlink agent=codebuddy,qoder,windsurf,roo`受管/外部/真实目录三态分支、`make skills.link AGENT=claude-code`包装层和`git diff --check`。`i18n`影响：仅新增开发工具命令、命令输出英文表格列名（开发者面向），不新增或修改前端运行时语言包、宿主/插件`manifest/i18n`、`apidoc/i18n`资源；中英文`README`双语同步。缓存一致性影响：命令是开发期一次性`os.Symlink`/`os.Remove`操作，不引入运行时缓存、共享修订号或跨实例协调。数据权限影响：命令操作仓库内文件系统，不访问任何后端数据接口或运行时数据权限边界。开发工具脚本影响：完全使用`Go`标准库实现（`os.Symlink`/`os.Readlink`/`os.Lstat`/`os.Remove`/`os.MkdirAll`/`filepath.Rel`），不调用`ln`/`mklink`/`bash`/`cmd.exe`，跨`Windows`/`Linux`/`macOS`一致；`Windows`下`os.Symlink`权限错误时通过`symlinkErrorDetail`追加“需要开发者模式或管理员”指引，`runtime.GOOS == "windows"`时路径比较使用 ASCII case-fold；测试中`trySymlink`在`Windows`权限不足时跳过而非失败。
- 2026-05-20：补齐交互式选择模式。新增`internal/skilllink/skilllink_interactive.go`提供`IsInteractiveTerminal`（基于`os.File.Stat()`+`ModeCharDevice`，不引入第三方依赖）、`LinkCandidates`、`UnlinkCandidates`、`PromptSelection`（支持逗号编号、`all`、`q`/空行取消、越界与非数字校验、去重）和`PromptYesNo`。改造`command_skills.link.go`：未传`agent`且 stdin 是真实终端时进入交互式选择，对包含`mismatch`的候选追问是否启用`force=1`；非终端上下文保持只读列表，确保 CI 不被破坏。改造`command_skills.unlink.go`：终端下交互式仅列出当前已是受管软链的 Agent，空候选时直接退出；非终端上下文要求显式`AGENT=`并以非零退出码报错。新增`skilllink_interactive_test.go`覆盖 8 个交互式用例（逗号选择、`all`、空/`q`/空白取消、越界/非数字错误、去重、yes/no 默认值与显式输入、非法答案、空候选清单、`LinkCandidates`仅含 link、`UnlinkCandidates`仅含 OK 状态、`IsInteractiveTerminal(nil)`）。同步更新`hack/tools/linactl/README.md`、`README.zh-CN.md`新增“交互模式”小节，并在`OpenSpec`提案/设计/规范文档中追加交互式 Requirement、Decisions 1 节、Tasks 第 6 节。验证通过：`cd hack/tools/linactl && go test ./... -count=1`（24 个 skilllink 用例全绿）、`cd hack/tools/linactl && go build ./...`、非 TTY 烟测`go run . skills.link`仅显示只读列表、`go run . skills.link agent=qoder` 仍可直接创建、`script -q -c "go run . skills.link" /dev/null < <(printf '1\nq\n')`模拟 TTY 验证 39 个 link 候选编号渲染与执行选择、`script -q -c "make skills.link" /dev/null < <(printf 'q\n')`模拟 TTY 验证`make`包装层进入交互式并支持`q`取消。`i18n`影响：交互式提示文案是开发者面向英文（`Select agents to link:`/`Cancelled.`/`Enter numbers separated by commas...`），不进入运行时`i18n`资源；中英文`README`同步说明交互模式。缓存一致性影响：交互式仅是控制流变化，仍调用相同的`Apply*`核心，不新增缓存或跨实例协调。数据权限影响：不访问后端接口。开发工具脚本影响：交互式实现仅使用`bufio.Reader`+标准库`os.File.Stat`，不引入第三方依赖与平台脚本。
- 2026-05-20：调整交互式排版。原单列编号布局在 39 个`link`类 Agent 下需 40+ 行，超出常见 24 行终端高度。改为 3 列网格 + 单字符状态符号（`[+]`linked / `[~]`mismatch / `[.]`absent / `[!]`conflict / `[*]`root-collision / `[?]`error）+ 图例行排版，将候选清单压缩到 13 行，加上标题、图例和提示行总计 17 行，可在 24 行终端中一屏完整显示。状态符号全部使用 ASCII 字符避免`Unicode`/`emoji`字符宽度差异；网格中省略`ProjectPath`列以控制单元格宽度，完整状态文字与路径仍可通过非交互列表（`make skills.link`无 AGENT + 非 TTY）查询。新增`statusGlyph`内部辅助函数和`renderCandidateGrid`私有渲染函数，并补充 2 个新单元测试覆盖`statusGlyph`全`Status`映射、3 列网格布局行数和图例行渲染。同步更新中英文`README`和`OpenSpec`提案/设计/规范文档，明确网格排版与状态符号约定。验证通过：`cd hack/tools/linactl && go test ./... -count=1`（共 26 个 skilllink 用例全绿）、`script -q -c "go run . skills.link" /dev/null < <(printf 'q\n')`模拟 TTY 验证 13 行 3 列网格、图例和取消行为。`i18n`影响：交互式新增图例行使用开发者面向英文，不进入运行时`i18n`资源；中英文`README`同步说明状态符号约定。缓存一致性影响：仅 UI 渲染调整，不影响缓存。数据权限影响：不访问后端接口。开发工具脚本影响：仅 Go 渲染逻辑变化，不引入新依赖。
- 2026-05-20：新增`make skills`聚合交互入口。新增`hack/tools/linactl/command_skills.go`实现`linactl skills`命令：TTY 下展示`[1] link` / `[2] unlink` / `[q] quit`三选一菜单并分发到`runSkillsLinkInteractive`/`runSkillsUnlinkInteractive`；非 TTY 打印`linactl skills.link`/`skills.unlink`和`make skills.link`/`make skills.unlink`用法指引并以零退出码结束。在`internal/skilllink/skilllink_interactive.go`中提取`ReadLine`公共函数（统一去除空白与小写化），并新增`TestReadLineTrimsAndLowercases`用例覆盖。`hack/tools/linactl/command.go`将`skills`命令注册在`skills.link`/`skills.unlink`之前；`hack/makefiles/skills.mk`新增`skills`目标转发`linactl skills`。同步更新中英文`README`和`OpenSpec`提案/规范文档。验证通过：`cd hack/tools/linactl && go test ./... -count=1`、`cd hack/tools/linactl && go build ./...`、非 TTY `go run . skills`输出用法指引、`script -q -c "make skills" /dev/null < <(printf '1\nq\n')`模拟 TTY 验证菜单 → link 交互式 → 取消、`script -q -c "make skills" /dev/null < <(printf '2\n')`验证 unlink 交互式分支、`script -q -c "make skills" /dev/null < <(printf 'q\n')`验证菜单立即取消。`i18n`影响：菜单文案为开发者面向英文，不进入运行时`i18n`资源。缓存一致性影响：仅交互入口分发，不引入新缓存。数据权限影响：不访问后端接口。开发工具脚本影响：`make skills`通过`$(LINACTL) skills`调用 Go 实现，不引入平台脚本。
- 2026-05-20：处理 `lina-review` 报告的两项警告。警告 1（命令层 `fmt.Fprintln` 错误返回未处理）：在 `command_skills.go`/`command_skills.link.go`/`command_skills.unlink.go` 中新增本文件级 `writeLine`/`writeLines` helper，将所有 stdout 行写入收敛到错误处理路径，与 `command_status.go` 严格风格保持一致；不扩大范围修改既有 `command_help.go` 等同模式文件。警告 2（`/skills` 路径未忽略）：在 `.gitignore` 末尾追加 `/skills` 条目并配套注释说明其专用于 `rootCollision` Agent (`openclaw`)、需要 `AGENT=openclaw FORCE=1` 才会创建的语义。验证通过：`cd hack/tools/linactl && go test ./... -count=1`（31 个 skilllink 用例全绿）、`cd hack/tools/linactl && go vet ./...`（无输出）、`go run . skills`/`skills.link`/`skills.unlink` 三路径烟测、`script -q -c "make skills"` TTY 菜单烟测、`git check-ignore -v skills` 验证 `.gitignore:100:/skills` 命中。`i18n`/缓存一致性/数据权限/开发工具脚本影响：均无新增（仅修复 stdout 错误处理与 .gitignore 治理）。
