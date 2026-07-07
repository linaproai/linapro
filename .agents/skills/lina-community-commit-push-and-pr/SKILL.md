---
name: lina-community-commit-push-and-pr
description: >-
  手动触发：为 LinaPro 主仓库及 apps/lina-plugins 子模块完成提交、PR 前 rebase、推送、创建 PR，
  并在 PR 合并后恢复原始分支和同步 main。禁止自动触发。
---

# 技能介绍

用于串联`LinaPro`主仓库与`apps/lina-plugins`子模块的社区提交、推送、`PR`创建和合并后收尾流程。子模块有变更时先创建子模块`PR`，确认合并后再更新主仓库子模块指针并创建主仓库`PR`。

## 核心规则

1. 主仓库默认是当前工作目录，远端仓库默认是`linaproai/linapro`。
2. 子模块远端从`.gitmodules`和`git -C apps/lina-plugins remote -v`读取，通常是`linaproai/official-plugins`。
3. 子模块和主仓库`PR`目标分支均固定为各自仓库的`main`。
4. 创建任何`PR`前，必须先提交当前任务变更，执行`PR`前`rebase`门禁，再推送对应分支。
5. 不静默丢弃工作区变更；发现无关变更时先报告风险，只有能明确限定到目标路径时才继续。
6. 除本技能明确要求的`PR`前`rebase`外，不使用`--force`、`--force-with-lease`、`git reset --hard`或其他历史重写命令，除非用户明确要求。
7. 子模块`PR`创建后必须停止，等待用户合并后回复`已合并`或`继续`；主仓库`PR`创建后也必须停止，等待用户合并后回复`已合并`或`继续收尾`。
8. 用户回复继续后，必须用`gh pr view`或远端`main`提交确认对应`PR`已合并；未合并时停止并提示继续 review。
9. 两个`PR`均确认合并后，必须切回流程开始时记录的原始分支，并让原始分支包含各自最新`origin/main`。
10. 遇到合并、同步、`rebase`或子模块指针冲突时立即停止并交给用户处理；禁止自行编辑冲突文件、删除冲突标记、选择 ours/theirs、暂存冲突结果、提交冲突解决或继续`rebase`。

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

记录流程开始时的原始分支，收尾必须使用这些值恢复工作树：

```bash
original_main_branch=$(git branch --show-current)
original_sub_branch=$(git -C apps/lina-plugins branch --show-current)
```

如果主仓库或子模块处于分离`HEAD`状态，或任一原始分支为空，停止并说明原因，除非用户明确要求继续。

## PR 前 rebase 门禁

该门禁用于避免历史提交被带入`PR`。执行顺序固定为：提交当前任务变更 → `fetch origin main` → `rebase origin/main` → 检查`origin/main..HEAD` → 推送当前分支 → 创建`PR`。不得用`merge origin/main`代替。

子模块分支：

```bash
git -C apps/lina-plugins fetch origin main
git -C apps/lina-plugins rebase origin/main
git -C apps/lina-plugins merge-base --is-ancestor origin/main HEAD
git -C apps/lina-plugins log --oneline --decorate origin/main..HEAD
```

主仓库分支：

```bash
git fetch origin main
git rebase origin/main
git merge-base --is-ancestor origin/main HEAD
git log --oneline --decorate origin/main..HEAD
```

门禁处理要求：

- 如果`origin/main..HEAD`包含与本次任务无关、数量异常或无法解释的提交，不得创建`PR`；报告提交列表和风险，等待用户决定是否改用新分支、拆分提交或手动整理历史。
- 如果普通`git push origin "<branch>"`因非快进被拒绝，不得自动使用`--force`或`--force-with-lease`；说明需要用户明确授权强推，或让用户决定是否改用新分支名。
- 如果`rebase`发生冲突，立即停止，不推送、不创建`PR`，报告冲突仓库、分支和文件；禁止执行`git add`、`git commit`、`git rebase --continue`、`git rebase --skip`或`git rebase --abort`，除非用户明确要求。

## 冲突处理

执行可能产生合并结果的操作前，尽量先用只读方式判断是否可自动合并：

```bash
git merge-tree --write-tree HEAD origin/main
```

如果只读检查确认可以自动合并，可以执行正常合并并推送：

```bash
git merge --no-edit origin/main
git push origin "$(git branch --show-current)"
```

如果只读检查或实际合并命令报告冲突，立即停止并报告冲突仓库、分支和文件。若实际合并命令已让工作区进入冲突状态，保留现场让用户处理；不要自动执行`git merge --abort`，除非用户明确要求放弃该次合并。

## 阶段一：处理子模块

检查子模块是否有未提交、已暂存或未跟踪变更：

```bash
git -C apps/lina-plugins status --short --branch
git -C apps/lina-plugins diff --stat
git -C apps/lina-plugins diff --cached --stat
```

如果子模块没有内容更新，跳过子模块提交和子模块`PR`创建，直接进入阶段二。

如果子模块有内容更新：

1. 读取当前分支；若是`main`，先创建或要求切换到任务分支，并刷新`sub_branch`。

```bash
sub_branch=$(git -C apps/lina-plugins branch --show-current)
if [ "$sub_branch" = "main" ]; then
  git -C apps/lina-plugins switch -c "<task-branch>"
  sub_branch=$(git -C apps/lina-plugins branch --show-current)
fi
```

2. 根据实际差异生成`<类型>[可选作用域]: <描述>`格式的提交信息，并提交当前任务变更。

```bash
git -C apps/lina-plugins add -A
git -C apps/lina-plugins commit -m "<submodule-subject>"
```

3. 执行`PR`前`rebase`门禁。
4. 推送子模块分支并按`PR`正文要求创建子模块`PR`。

```bash
git -C apps/lina-plugins push origin "$sub_branch"
gh pr create \
  --repo "<submodule-owner>/<submodule-repo>" \
  --base main \
  --head "$sub_branch" \
  --title "<title>" \
  --body-file -
```

5. 输出子模块`PR`地址并停止，提醒用户：

```text
请先 review 并合并该 submodule PR。合并完成后回复“已合并”或“继续”，我再更新主仓库 submodule 指针并创建主仓库 PR。
```

阶段一创建`PR`后不得继续执行阶段二，即使本地能看到子模块分支已经推送。

## 阶段二：更新主仓库并创建 PR

仅当子模块无内容更新，或阶段一创建的子模块`PR`已确认合并时，才能进入本阶段。

1. 拉取主仓库和子模块远端`main`。

```bash
git fetch origin main
git -C apps/lina-plugins fetch origin main
```

2. 如果阶段一创建过子模块`PR`，先确认它已合并；未合并则停止。

```bash
gh pr view "<submodule-pr-number-or-url>" \
  --repo "<submodule-owner>/<submodule-repo>" \
  --json number,state,mergeCommit,url,baseRefName,headRefName
```

3. 将子模块工作树切到并快进到子模块`main`。

```bash
git -C apps/lina-plugins switch main
git -C apps/lina-plugins merge --ff-only origin/main
```

4. 确认主仓库只包含预期变更。若存在其他变更，先报告差异；只有确认属于当前任务时才一并提交。

```bash
git status --short --branch
git diff --submodule=log -- apps/lina-plugins
git diff --stat
```

5. 提交主仓库当前任务变更。

```bash
main_branch=$(git branch --show-current)
git add apps/lina-plugins
# 如果还有已确认属于本次任务的主仓库变更，按具体路径追加 git add。
git commit -m "chore(plugins): update lina-plugins submodule"
```

如果主仓库还有与本次任务相关且尚未提交的变更，先检查差异并生成符合规范的提交信息；不得把无关变更混入子模块指针提交。

6. 执行`PR`前`rebase`门禁。
7. 推送主仓库分支并按`PR`正文要求创建主仓库`PR`。

```bash
git push origin "$main_branch"
gh pr create \
  --repo linaproai/linapro \
  --base main \
  --head "$main_branch" \
  --title "<title>" \
  --body-file -
```

8. 如果`GitHub`显示主仓库`PR`为`CONFLICTING`，按“冲突处理”检查；能自动合并时正常合入`origin/main`并推送，出现冲突时停止。

阶段二完成后停止，提醒用户：

```text
请先 review 并合并主仓库 PR。合并完成后回复“已合并”或“继续收尾”，我再恢复主仓库和 submodule 的本地分支并同步 main。
```

## 阶段三：PR 合并后恢复分支并同步 main

进入本阶段前必须满足：

- 阶段一创建过子模块`PR`时，该`PR`已确认`MERGED`。
- 阶段二创建过主仓库`PR`时，该`PR`已确认`MERGED`。
- 已记录`original_main_branch`和`original_sub_branch`，且二者都不是空值。

1. 再次确认所有已创建的`PR`状态为`MERGED`；否则停止。

```bash
gh pr view "<submodule-pr-number-or-url>" \
  --repo "<submodule-owner>/<submodule-repo>" \
  --json number,state,mergeCommit,url,baseRefName,headRefName
gh pr view "<framework-pr-number-or-url>" \
  --repo linaproai/linapro \
  --json number,state,mergeCommit,url,baseRefName,headRefName
```

2. 刷新两个仓库的远端`main`。

```bash
git fetch origin main
git -C apps/lina-plugins fetch origin main
```

3. 切回子模块原始分支，并让该分支包含最新子模块`origin/main`。

```bash
git -C apps/lina-plugins switch "$original_sub_branch"
git -C apps/lina-plugins fetch origin main
git -C apps/lina-plugins merge --ff-only origin/main
```

如果`--ff-only`失败，按“冲突处理”判断是否可自动合并；冲突或需要人工选择时立即停止。

4. 切回主仓库原始分支，并让该分支包含最新主仓库`origin/main`。

```bash
git switch "$original_main_branch"
git fetch origin main
git merge --ff-only origin/main
```

如果`--ff-only`失败，按“冲突处理”判断是否可自动合并；冲突或需要人工选择时立即停止。

5. 验证两个原始分支均已包含最新`origin/main`，并检查工作区状态。

```bash
git status --short --branch
git log --oneline -1
git merge-base --is-ancestor origin/main HEAD
git -C apps/lina-plugins status --short --branch
git -C apps/lina-plugins log --oneline -1
git -C apps/lina-plugins merge-base --is-ancestor origin/main HEAD
```

6. 如果同步后当前分支相对远端同名分支存在本地领先提交，推送对应原始分支；若远端同名分支不存在，使用`-u`建立上游；若存在落后或分叉，停止并报告风险。

```bash
git rev-list --left-right --count "origin/$original_main_branch"...HEAD
git -C apps/lina-plugins rev-list --left-right --count "origin/$original_sub_branch"...HEAD
```

需要推送时执行：

```bash
git push origin "$original_main_branch"
git -C apps/lina-plugins push origin "$original_sub_branch"
```

## PR 正文要求

所有`PR`正文都应包含：

- `Summary`：详细概述本次变更内容。
- `Tests`：说明本次流程是否运行过测试；没有运行时写明`Not run by this PR creation step.`。
- `Related Issue`：从用户请求、任务记录、上下文、分支名、提交信息和相关链接中自动识别；识别到时使用`Resolves <issue-id-or-url>`，未识别到时说明无关联`Issue`且不得阻塞`PR`创建。

主仓库`PR`还必须包含：

- `Submodule`：记录`apps/lina-plugins`从哪个提交更新到哪个提交。

## 输出要求

阶段一如果创建了子模块`PR`，最终只输出：

- 子模块提交分支、提交主题和推送目标。
- 子模块`PR`前`rebase`目标与`origin/main..HEAD`检查结果。
- 子模块`PR`地址。
- 等待用户 review 并合并子模块`PR`的提醒。

阶段二完成后最终只输出：

- 子模块最终`main`提交。
- 主仓库提交分支、提交主题和推送目标。
- 主仓库`PR`前`rebase`目标与`origin/main..HEAD`检查结果。
- 主仓库`PR`地址和当前是否可合并。
- 本地工作区是否干净。
- 测试是否运行；未运行时如实说明。

阶段三完成后最终只输出：

- 已确认合并的子模块`PR`地址和合并提交。
- 已确认合并的主仓库`PR`地址和合并提交。
- 子模块当前分支、同步后的提交、是否已包含最新远端`main`和是否已推送。
- 主仓库当前分支、同步后的提交、是否已包含最新远端`main`和是否已推送。
- 本地工作区是否干净。
- 测试是否运行；未运行时如实说明。
