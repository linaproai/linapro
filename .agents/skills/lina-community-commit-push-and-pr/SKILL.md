---
name: lina-community-commit-push-and-pr
description: 为LinaPro以及SubModule仓库自动提交、推送并创建 PR。必须用户手动触发，禁止自动触发该技能。
---

# 技能介绍

`LinaPro`社区提交与`PR`串联技能。该技能用于把`apps/lina-plugins`子模块和主框架仓库按正确顺序分别提交、推送、创建`PR`，并强制在子模块`PR`合并后才推进主框架`PR`。

## 核心规则

1. 默认主框架仓库路径是当前工作目录，默认主框架远端仓库是`linaproai/linapro`。
2. 默认子模块远端仓库从`.gitmodules`和`git -C apps/lina-plugins remote -v`读取，通常是`linaproai/official-plugins`。
3. 子模块`PR`目标分支固定为子模块仓库的`main`。
4. 主框架`PR`目标分支固定为主框架仓库的`main`。
5. 不使用`--force`、`--force-with-lease`、`git reset --hard`或历史重写命令，除非用户明确要求。
6. 不静默丢弃任何工作区变更。发现无关变更时先报告风险；只有在能明确限定到目标路径时才继续。
7. 创建任何`PR`前必须确认对应分支已经推送到`origin`。
8. 创建子模块`PR`后必须停止，并提醒用户先 review 和合并该`PR`。用户回复`已合并`或`继续`之前，不得更新主框架的子模块指针、提交主框架或创建主框架`PR`。
9. 用户回复`已合并`或`继续`后，必须先通过`gh pr view`或远端`main`提交确认子模块`PR`已经合并；如果尚未合并，停止并提示用户继续 review。
10. 执行过程中遇到`git`合并、同步或更新分支时，只允许`Git`自动完成无冲突合并；一旦出现内容冲突、子模块指针冲突或需要人工选择的冲突，必须立即停止并交给用户处理，禁止自行编辑冲突文件、删除冲突标记、选择 ours/theirs、暂存冲突结果或提交冲突解决。

## 合并冲突处理

执行任何可能产生合并结果的操作前，先尽量使用只读方式判断是否可自动合并，例如：

```bash
git merge-tree --write-tree HEAD origin/main
```

如果只读检查确认可以自动合并，可以执行正常合并命令并推送结果，例如：

```bash
git merge --no-edit origin/main
git push origin "$(git branch --show-current)"
```

如果只读检查或实际合并命令报告冲突：

1. 立即停止当前流程，不继续提交、推送或创建后续`PR`。
2. 报告冲突仓库、冲突分支和冲突文件列表。
3. 明确提醒用户需要手动处理冲突后再回复继续。
4. 禁止修改冲突文件内容、禁止执行`git checkout --ours`、`git checkout --theirs`、`git add`、`git commit`或任何等价的冲突解决动作。

如果实际合并命令已经让工作区进入冲突状态，保留现场并停止，让用户完成冲突处理；不要自动执行`git merge --abort`，除非用户明确要求放弃该次合并。

## 前置检查

先执行只读检查：

```bash
git status --short --branch
git branch --show-current
cat .gitmodules
git submodule status apps/lina-plugins
git -C apps/lina-plugins status --short --branch
git remote -v
git -C apps/lina-plugins remote -v
gh auth status
```

如果主仓库或子模块处于分离`HEAD`状态，停止并说明原因，除非用户明确要求在分离`HEAD`状态继续。

## 阶段一：处理子模块

### 判断是否需要子模块 PR

检查`apps/lina-plugins`是否存在未提交、已暂存或未跟踪变更：

```bash
git -C apps/lina-plugins status --short --branch
git -C apps/lina-plugins diff --stat
git -C apps/lina-plugins diff --cached --stat
```

如果子模块没有内容更新，跳过子模块提交和子模块`PR`创建，直接进入阶段二。

如果子模块有内容更新：

1. 检查子模块当前分支：

```bash
sub_branch=$(git -C apps/lina-plugins branch --show-current)
```

2. 如果`sub_branch`是`main`，不要直接在`main`上提交。创建或要求切换到任务分支，例如：

```bash
git -C apps/lina-plugins switch -c "<task-branch>"
```

3. 根据实际差异生成子模块提交信息，使用`<类型>[可选作用域]: <描述>`格式。
4. 提交并推送子模块分支：

```bash
git -C apps/lina-plugins add -A
git -C apps/lina-plugins commit -m "<submodule-subject>"
git -C apps/lina-plugins push origin "$sub_branch"
```

5. 创建子模块`PR`：

```bash
gh pr create \
  --repo "<submodule-owner>/<submodule-repo>" \
  --base main \
  --head "$sub_branch" \
  --title "<title>" \
  --body-file -
```

6. 输出子模块`PR`地址，并停止当前流程。必须明确提醒用户：

```text
请先 review 并合并该 submodule PR。合并完成后回复“已合并”或“继续”，我再更新主框架 submodule 指针并创建主框架 PR。
```

阶段一完成后不得继续执行阶段二，即使本地能看到子模块分支已经推送。

## 阶段二：更新主框架依赖并创建主框架 PR

只有在以下任一条件满足时才能进入阶段二：

- 阶段一判断子模块没有内容更新；
- 用户在子模块`PR`创建后回复`已合并`或`继续`，且已确认该子模块`PR`状态为`MERGED`。

进入阶段二后执行：

1. 拉取主框架和子模块远端`main`：

```bash
git fetch origin main
git -C apps/lina-plugins fetch origin main
```

2. 如果阶段一创建过子模块`PR`，确认该`PR`已合并：

```bash
gh pr view "<submodule-pr-number-or-url>" \
  --repo "<submodule-owner>/<submodule-repo>" \
  --json number,state,mergeCommit,url,baseRefName,headRefName
```

如果`state`不是`MERGED`，停止并再次提醒用户先 review 和合并子模块`PR`。

3. 将子模块工作树切到并快进到子模块`main`：

```bash
git -C apps/lina-plugins switch main
git -C apps/lina-plugins merge --ff-only origin/main
```

4. 确认主框架只包含预期的子模块指针变更。如果还存在其他主框架变更，应报告差异；若这些变更是当前任务已有变更，可一并提交，否则先向用户说明风险。

```bash
git status --short --branch
git diff --submodule=log -- apps/lina-plugins
git diff --stat
```

5. 在主框架提交并推送当前分支：

```bash
main_branch=$(git branch --show-current)
git add apps/lina-plugins
git commit -m "chore(plugins): update lina-plugins submodule"
git push origin "$main_branch"
```

如果主框架当前还有与本次任务相关且尚未提交的变更，先检查差异并生成符合规范的提交信息；不得无说明地把无关变更混入子模块指针提交。

6. 创建主框架`PR`到`main`：

```bash
gh pr create \
  --repo linaproai/linapro \
  --base main \
  --head "$main_branch" \
  --title "<title>" \
  --body-file -
```

7. 如果`GitHub`显示主框架`PR`为`CONFLICTING`，先按“合并冲突处理”规则做只读检查：

```bash
git merge-tree --write-tree HEAD origin/main
gh pr view "<framework-pr-number-or-url>" \
  --repo linaproai/linapro \
  --json mergeable,mergeStateStatus,statusCheckRollup
```

若本地确认可自动合并，可以把`origin/main`正常合入当前分支并推送；若出现冲突，停止并报告冲突文件。

## PR 正文要求

子模块`PR`正文至少包含：

- `Summary`：概述子模块变更。
- `Tests`：说明本次流程是否运行过测试；没有运行时写明`Not run by this PR creation step.`。

主框架`PR`正文至少包含：

- `Summary`：概述主框架变更和子模块依赖更新。
- `Submodule`：记录`apps/lina-plugins`从哪个提交更新到哪个提交。
- `Tests`：说明本次流程是否运行过测试；没有运行时写明`Not run by this PR creation step.`。

## 输出要求

阶段一如果创建了子模块`PR`，最终输出只包含：

- 子模块提交分支和提交主题；
- 子模块推送目标；
- 子模块`PR`地址；
- 明确提醒用户 review 并合并该`PR`，合并后回复`已合并`或`继续`。

阶段二完成后最终输出：

- 子模块最终`main`提交；
- 主框架提交分支和提交主题；
- 主框架推送目标；
- 主框架`PR`地址；
- `PR`当前是否可合并；
- 本地工作区是否干净；
- 测试是否运行，未运行时如实说明。
