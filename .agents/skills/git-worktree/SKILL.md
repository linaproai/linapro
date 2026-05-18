---
name: git-worktree
description: Create and switch to an isolated Git worktree for the current task, continuing work in a new directory instead of the original checkout. Use when the user wants an isolated worktree, a clean branch directory, safer parallel changes, or a fresh workspace that won't interfere with existing local modifications.
---

# Git 工作树

为当前任务创建专用的 `git worktree`，然后在新目录中继续工作，而非使用原始检出。

此技能用于执行而非仅提供建议。触发时应直接创建工作树，除非仓库状态导致无法执行。

此技能不引入辅助脚本，直接使用内联的 `git` 和 shell 命令。

**交互语言**：与用户交互的内容语言以用户上下文使用的语言为准，用户使用英文则使用英文，用户使用中文则使用中文。

## 适用场景

- 用户明确要求创建新的 `git worktree`、独立分支目录或隔离工作空间
- 当前检出包含无关的本地变更，隔离是最安全的方式
- 用户希望在多个任务上并行工作，无需暂存或干扰原始工作树
- 用户说"创建独立的分支目录"、"打开新的工作树"、"使用干净的检出"，或"在隔离空间中工作"

## 核心规则

创建工作树后，将新路径视为任务执行期间的活跃工作目录。

在任何代理环境中，"进入目录"意味着：

- 后续命令以新工作树路径为执行目标
- 所有编辑操作都在该工作树路径下进行
- 不要意外继续使用原始检出
- 通过在新工作树中至少执行一条后续命令来确认切换完成

除非后续操作确实指向新的 `worktree_path`，否则不要声称已"切换"。

如果环境支持按命令设置工作目录，则为后续每个命令使用该设置。如果不支持，则在后续命令前添加显式的 `cd <worktree_path> && ...`。

## 名称派生

- 根据用户的实际任务派生简短的 ASCII kebab-case 任务标识，如 `login-timeout-fix` 或 `user-export`
- 除非请求过于模糊，否则不要使用 `git-worktree`、`new-worktree` 或 `task` 等通用名称
- 如果请求主要为非 ASCII 内容或没有明显的标识可用，回退到 `task-$(date +%Y%m%d-%H%M%S)`
- 默认分支前缀为 `worktree/`
- 默认工作树目录为仓库根目录的同级目录，命名为 `<仓库名>-<标识>`

## 默认工作流

1. 从当前检出检查仓库上下文：
   - `git rev-parse --show-toplevel`
   - `git branch --show-current`
   - `git status --short`
   - `git worktree list --porcelain`
2. 根据上述规则自行确定任务标识
3. 内联构建分支和路径名称，然后使用直接的 shell 命令创建工作树：

   ```bash
   repo_root=$(git rev-parse --show-toplevel)
   repo_name=$(basename "$repo_root")
   parent_dir=$(dirname "$repo_root")
   source_branch=$(git -C "$repo_root" branch --show-current)
   if [ -n "$source_branch" ]; then
     source_ref="$source_branch"
   else
     source_ref="HEAD@$(git -C "$repo_root" rev-parse --short HEAD)"
   fi

   slug="<task-slug>"
   base_branch="worktree/$slug"
   branch_name="$base_branch"
   base_path="$parent_dir/$repo_name-$slug"
   worktree_path="$base_path"
   index=2

   while git -C "$repo_root" show-ref --verify --quiet "refs/heads/$branch_name" || [ -e "$worktree_path" ]; do
     branch_name="${base_branch}-$index"
     worktree_path="${base_path}-$index"
     index=$((index + 1))
   done

   git -C "$repo_root" worktree add -b "$branch_name" "$worktree_path" HEAD
   ```

4. 立即在新工作树中验证切换，例如：

   ```bash
   pwd
   git status --short --branch
   ```

   这些验证命令必须以 `worktree_path` 为执行目标。

5. 简要告知新活跃路径，然后在新目录中继续执行主任务
6. 任务剩余部分中，所有相关命令或编辑操作均以 `worktree_path` 为工作目录

## 行为规则

- 默认基准引用为当前检出的 `HEAD`，以避免将未提交的本地变更带入新工作树
- 如果分支名或路径已存在，自动递增后缀而非报错
- 如果已在非默认工作树中且用户仍需要另一个隔离空间，从当前 `HEAD` 创建新工作树
- 如果目录不是 Git 仓库，明确说明并不要假装已创建工作树
- 如果工作树创建成功，继续执行用户的实际任务，而非停留在设置阶段
- 如果工作树因文件系统权限创建失败，请求最小必要权限后重试

## 未提交变更策略

安全默认策略是与未提交变更隔离。

- 如果源检出有未提交变更，仍从 `HEAD` 创建新工作树，除非用户明确要求携带本地修改
- 不要静默暂存、重置或移动用户现有的变更
- 如果用户希望将本地修改复制到新工作树中，使用显式流程（如临时提交、补丁或拣选），并说明正在执行的操作

## 输出约定

使用此技能时：

- 告知用户创建了哪个分支和目录
- 明确后续工作将在该路径下进行
- 在上下文相关时说明源引用和原始检出是否有未提交变更
- 如果用户要求额外工作，不要在设置完成后停止，继续在新工作树中执行任务

## 示例

用户请求：

```text
为这个任务创建一个独立的工作树，然后开始实现。
```

预期行为：

- 检查当前仓库状态
- 使用直接的 `git worktree` 命令创建新的 `worktree/...` 分支和同级目录
- 将后续所有命令切换到该目录
- 在新目录中继续执行请求的实现工作

## 清理

仅在用户要求或清理明显是任务一部分时才移除工作树。

清理前：

- 检查所创建工作树的状态
- 确认正在移除的是正确的路径
- 永远不要移除用户的原始检出
