---
name: git-commit-push
description: Review current Git workspace changes, generate a commit message from actual diffs following the repo's commit or PR naming conventions, commit all changes on the current branch, and push to `origin`. Use when the user asks to "commit", "push", "commit and push", "generate a commit message", "commit current changes", or push the current branch changes to the remote.
---

# Git 提交与推送

审查当前仓库变更，生成符合仓库规范的简洁提交主题，在当前分支上提交所有修改，并将分支推送到 `origin`。

此技能用于执行而非仅提供建议。触发时应直接运行 Git 工作流，除非仓库状态导致操作不安全或无法执行。

**交互语言**：与用户交互的内容语言以用户上下文使用的语言为准，用户使用英文则使用英文，用户使用中文则使用中文。

## 适用场景

- 用户要求提交当前变更，无论是否需要推送
- 用户希望从差异内容生成提交信息，而非预先编写
- 用户提及仓库的 PR 或提交命名规范并希望遵循
- 用户说"提交当前分支"、"帮我提交"、"提交并推送"、"生成提交信息并推送"，或"将变更推送到 origin"

## 核心行为

1. 确认当前处于 Git 仓库中，通过 `git branch --show-current` 检测当前分支。
2. 提交前检查工作区状态：
   - `git status --short --branch`
   - `git diff --stat`
   - `git diff --cached --stat`
   - `git diff -- . ':(exclude)package-lock.json'` 或按需使用更精确的路径过滤以提高可读性
3. **检测子模块变更**：检查 `.gitmodules` 中声明的子模块（如 `apps/lina-plugins`）是否已初始化且存在未提交变更。若存在，先在子模块内执行提交与推送（详见「子模块处理」章节）。
4. 根据实际变更文件和差异内容生成提交主题，而非仅依赖用户描述。
5. 使用 `git add -A` 暂存当前分支上的所有修改（此时子模块指针已更新）。
6. 使用生成的信息执行一次提交。
7. 使用 `git push origin <当前分支>` 推送当前分支到 `origin`。

## 子模块处理

当仓库包含 Git 子模块时，必须在主仓库提交**之前**先处理子模块内的变更，确保主仓库记录的是子模块最新的提交指针。

### 检测逻辑

1. 读取 `.gitmodules` 获取所有声明的子模块路径。
2. 对每个子模块路径，检查目录是否存在且已初始化（`git -C <submodule-path> rev-parse --git-dir` 能成功执行）。
3. 对已初始化的子模块，检查其工作区是否有变更：
   - `git -C <submodule-path> status --short`
   - `git -C <submodule-path> diff --stat`
   - `git -C <submodule-path> diff --cached --stat`
4. 若子模块工作区干净（无变更、无未跟踪文件），跳过该子模块。

### 子模块提交与推送流程

对存在变更的子模块，按以下顺序操作：

```bash
# 1. 进入子模块目录，检测当前分支
sub_branch=$(git -C <submodule-path> branch --show-current)

# 2. 查看子模块变更内容
git -C <submodule-path> status --short --branch
git -C <submodule-path> diff --stat
git -C <submodule-path> diff --cached --stat

# 3. 根据子模块内的变更生成独立的提交信息（遵循相同的提交信息规范）
#    作用域应使用子模块的组件名，如 feat(plugin): add xxx

# 4. 暂存并提交子模块变更
git -C <submodule-path> add -A
git -C <submodule-path> commit -m "<generated-subject-for-submodule>"

# 5. 推送子模块到其 origin
git -C <submodule-path> push origin "$sub_branch"
```

子模块提交完成后，再回到主仓库执行正常的提交流程。此时 `git add -A` 会将子模块的新提交指针一并暂存。

### 子模块提交信息规范

子模块的提交信息遵循与主仓库相同的 `<类型>[可选作用域]: <描述>` 规范，但作用域应反映子模块内部的组件结构而非主仓库的作用域。例如：

- 主仓库作用域示例：`feat(core): add plugin sync`
- 子模块作用域示例：`feat(plugin): add new source plugin`

### 执行约束

- **子模块先于主仓库**：子模块的提交和推送必须在主仓库 `git add -A` 之前完成，否则主仓库无法记录子模块的最新指针。
- **独立提交信息**：子模块使用独立生成的提交信息，不与主仓库共享同一条提交信息。
- **独立分支检测**：子模块可能处于与主仓库不同的分支上，以子模块实际检测到的分支为准。
- **推送失败处理**：若子模块推送失败，报告具体错误并停止，不继续执行主仓库的提交。用户需先解决子模块的推送问题。
- **子模块无变更则跳过**：若子模块工作区干净，直接跳过，不产生额外操作。

## 提交信息规范

提交信息格式为：`<类型>[可选作用域]: <描述>`，例如 `fix(os/gtime): fix time zone issue`
  + `<类型>` 为必填项，可选值包括 `fix`、`feat`、`build`、`ci`、`docs`、`style`、`refactor`、`perf`、`test`、`chore`
    + `fix`：修复 Bug
    + `feat`：新增功能
    + `build`：构建系统或外部依赖变更，如依赖升级、Node 版本变更等
    + `ci`：持续集成配置变更，如 Travis、Jenkins 工作流调整
    + `docs`：文档变更，如 README、API 文档更新等
    + `style`：代码风格调整，如缩进、空格、空行等
    + `refactor`：代码重构，如结构调整、变量/函数重命名等，不涉及功能变更
    + `perf`：性能优化，如提升代码性能、降低内存占用等
    + `test`：测试用例变更，如新增、删除或修改测试用例
    + `chore`：非业务代码变更，如构建流程或工具配置调整
  + `<类型>` 后用括号指定受影响的包名或作用域，例如 `(os/gtime)`
  + 冒号后使用动词短语，以动词原形开头
  + 冒号后首字母小写
  + 末尾不加句号
  + 标题尽量简短，理想情况下不超过 76 个字符
+ 如有对应的 Issue，在提交信息末尾添加 `fixes #1234`（若未完全修复则使用 `refs #1234`）

### 示例

#### 包含描述和破坏性变更脚注的提交信息
```
feat: allow provided config object to extend other configs
BREAKING CHANGE: `extends` key in config file is now used for extending other config files
```

#### 使用 ! 标记破坏性变更
```
feat!: send an email to the customer when a product is shipped
```

#### 包含作用域和 ! 标记的提交信息
```
feat(api)!: send an email to the customer when a product is shipped
```

#### 同时使用 ! 和 BREAKING CHANGE 脚注
```
feat!: drop support for Node 6
BREAKING CHANGE: use JavaScript features not available in Node 6.
```

#### 无正文的提交信息
```
docs: correct spelling of CHANGELOG
```

#### 包含作用域的提交信息
```
feat(lang): add Polish language
```

#### 包含多段正文和多个脚注的提交信息
```
fix: prevent racing of requests

Introduce a request id and a reference to latest request. Dismiss
incoming responses other than from latest request.

Remove timeouts which were used to mitigate the racing issue but are
obsolete now.

Reviewed-by: Z
Refs: #123
```

## 执行规则

- 提交工作区中所有已跟踪和未跟踪的变更，因为此技能用于"提交当前状态"的请求
- 如果没有变更，明确告知用户并在提交或推送前停止
- 如果 `git branch --show-current` 为空，说明当前处于分离 HEAD 状态并停止，除非用户明确要求从分离 HEAD 提交
- 除非用户明确要求，否则禁止使用 `--force`、`--force-with-lease` 或历史重写命令
- 如果推送因远程分支已更新而失败，报告具体错误并停止，不要自动变基或合并
- 除非用户要求排除，否则不要静默丢弃文件

## 建议的命令流程

### 步骤 1：检查主仓库状态

```bash
git status --short --branch
git diff --stat
git diff --cached --stat
branch_name=$(git branch --show-current)
```

### 步骤 2：处理子模块（若有变更）

```bash
# 检测子模块是否已初始化且有变更
git -C apps/lina-plugins status --short

# 若有变更，进入子模块执行提交与推送
sub_branch=$(git -C apps/lina-plugins branch --show-current)
git -C apps/lina-plugins diff --stat
git -C apps/lina-plugins add -A
git -C apps/lina-plugins commit -m "<submodule-commit-subject>"
git -C apps/lina-plugins push origin "$sub_branch"
```

### 步骤 3：提交并推送主仓库

```bash
git add -A
git commit -m "<generated-subject>"
git push origin "$branch_name"
```

如果暂存前的差异内容嘈杂，或未跟踪文件显著改变了变更范围，暂存后应再次检查 `git diff --cached`。

## 输出约定

使用此技能时：

- 告知用户提交到了哪个分支
- 提供最终使用的提交主题
- 说明已暂存所有当前变更
- 报告推送目标为 `origin/<分支名>`
- **若处理了子模块变更，单独报告子模块的提交分支、提交主题和推送目标**
- 如果提交或推送未执行，说明具体原因

## 示例

用户请求：

```text
根据此仓库的规范生成提交信息，然后提交并推送当前分支
```

预期行为：

- 检查仓库状态和差异
- 根据实际变更生成符合规范的主题
- 对当前工作区执行一次提交
- 将当前分支推送到 `origin`
