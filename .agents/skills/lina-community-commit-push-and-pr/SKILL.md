---
name: lina-community-commit-push-and-pr
description: >-
  手动触发：为 LinaPro 主仓库及 apps/lina-plugins 子模块完成提交、PR 前 rebase、推送、创建 PR，
  主仓 CI 修复回路，以及 PR 合并后恢复原始分支并同步 main。禁止自动触发。
---

# 技能介绍

串联主仓库与`apps/lina-plugins`的提交、推送、`PR`与收尾。子模块有变更时先合子模块`PR`，再更新主仓库子模块指针并开主仓库`PR`。主仓库`PR`存活期间子模块修复走 fix 分支，禁止在子模块`main`上改代码。

## 核心规则

1. 主仓库默认当前工作目录，远端默认`linaproai/linapro`。
2. 子模块远端从`.gitmodules`和`git -C apps/lina-plugins remote -v`读取，通常是`linaproai/official-plugins`。
3. 子模块与主仓库`PR`目标分支均固定为各自`main`。
4. 创建任何`PR`前：提交当前变更 → `PR`前`rebase`门禁 → 推送 → 创建`PR`。
5. 不静默丢弃工作区变更；无关变更先报告，仅能限定到目标路径时才继续。
6. 除本技能要求的`PR`前`rebase`外，不使用`--force`、`--force-with-lease`、`git reset --hard`等历史重写，除非用户明确要求。
7. **子模块`main`只读对齐，禁止在其上`commit`/`push`修复。** 若当前在`main`且需改子模块，必须先`switch -c fix/<main-branch>`（或等价任务分支）。
8. 子模块`PR`创建后停止，等用户`已合并`/`继续`；主仓库`PR`创建后停止。用户说继续后用`gh pr view`或远端`main`确认已合并，未合并则停止。
9. 主仓库`PR`未合并前若需改子模块，走阶段 2.5，不得进入阶段三。
10. 两个`PR`均`MERGED`后，切回流程开始时的原始分支并包含最新`origin/main`。
11. 合并/`rebase`/子模块指针冲突时立即停止交用户；禁止自行解冲突或继续`rebase`。

## 前置检查

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

记录原始分支（收尾必用）：

```bash
original_main_branch=$(git branch --show-current)
original_sub_branch=$(git -C apps/lina-plugins branch --show-current)
```

主仓库或子模块处于分离`HEAD`，或任一原始分支为空时停止，除非用户明确要求继续。

## PR 前 rebase 门禁

顺序：提交 → `fetch origin main` → `rebase origin/main` → 检查`origin/main..HEAD` → 推送 → 创建`PR`。不得用`merge origin/main`代替。

子模块：

```bash
git -C apps/lina-plugins fetch origin main
git -C apps/lina-plugins rebase origin/main
git -C apps/lina-plugins merge-base --is-ancestor origin/main HEAD
git -C apps/lina-plugins log --oneline --decorate origin/main..HEAD
```

主仓库：

```bash
git fetch origin main
git rebase origin/main
git merge-base --is-ancestor origin/main HEAD
git log --oneline --decorate origin/main..HEAD
```

- `origin/main..HEAD`含无关/异常提交时不得开`PR`，报告后等用户决定。
- 普通`push`非快进被拒时不得自动强推；需用户授权或改用新分支名。
- `rebase`冲突：停止，不推送、不开`PR`；禁止`add`/`commit`/`rebase --continue|--skip|--abort`，除非用户明确要求。

## 冲突处理

可合并性先只读检查：

```bash
git merge-tree --write-tree HEAD origin/main
```

可自动合并时：

```bash
git merge --no-edit origin/main
git push origin "$(git branch --show-current)"
```

冲突则停止并报告；工作区已冲突时保留现场，不自动`merge --abort`，除非用户要求。

## 阶段一：子模块 PR

```bash
git -C apps/lina-plugins status --short --branch
git -C apps/lina-plugins diff --stat
git -C apps/lina-plugins diff --cached --stat
```

无内容更新则跳过，进入阶段二。

有更新时：

1. 若在`main`，先切任务分支：

```bash
sub_branch=$(git -C apps/lina-plugins branch --show-current)
if [ "$sub_branch" = "main" ]; then
  git -C apps/lina-plugins switch -c "<task-branch>"
  sub_branch=$(git -C apps/lina-plugins branch --show-current)
fi
```

2. 提交（`<类型>[可选作用域]: <描述>`）：

```bash
git -C apps/lina-plugins add -A
git -C apps/lina-plugins commit -m "<submodule-subject>"
```

3. `PR`前`rebase`门禁。
4. 推送并创建`PR`；记录`sub_pr_url`/`sub_branch`。

```bash
git -C apps/lina-plugins push origin "$sub_branch"
gh pr create \
  --repo "<submodule-owner>/<submodule-repo>" \
  --base main \
  --head "$sub_branch" \
  --title "<title>" \
  --body-file -
```

5. 输出`PR`地址后停止：

```text
请先 review 并合并该 submodule PR。合并完成后回复“已合并”或“继续”，我再更新主仓库 submodule 指针并创建主仓库 PR。
```

创建后不得进入阶段二。

## 阶段二：主仓库 PR

条件：子模块无更新，或阶段一`PR`已确认`MERGED`。

1. 拉取远端`main`。

```bash
git fetch origin main
git -C apps/lina-plugins fetch origin main
```

2. 若有子模块`PR`，确认已合并；未合并则停止。

```bash
gh pr view "<submodule-pr-number-or-url>" \
  --repo "<submodule-owner>/<submodule-repo>" \
  --json number,state,mergeCommit,url,baseRefName,headRefName
```

3. 子模块对齐`origin/main`（仅取指针，非修复工作区）：

```bash
git -C apps/lina-plugins switch main
git -C apps/lina-plugins merge --ff-only origin/main
```

4. 确认主仓库仅含预期变更；无关差异先报告。

```bash
git status --short --branch
git diff --submodule=log -- apps/lina-plugins
git diff --stat
```

5. 提交主仓库（指针与其它任务变更分开语义清晰即可）：

```bash
main_branch=$(git branch --show-current)
git add apps/lina-plugins
# 已确认属于本次任务的其它路径再 git add
git commit -m "chore(plugins): update lina-plugins submodule"
```

6. `PR`前`rebase`门禁。
7. 推送并创建主仓库`PR`；记录`main_pr_url`/`main_branch`。

```bash
git push origin "$main_branch"
gh pr create \
  --repo linaproai/linapro \
  --base main \
  --head "$main_branch" \
  --title "<title>" \
  --body-file -
```

8. 若`PR`状态为`CONFLICTING`，按冲突处理；能自动合则合入`origin/main`并推送，否则停止。

9. **主仓 PR 创建后：子模块切到 fix 分支，禁止停在`main`等待。**

```bash
git -C apps/lina-plugins fetch origin main
git -C apps/lina-plugins switch main
git -C apps/lina-plugins merge --ff-only origin/main
sub_fix_branch="fix/${main_branch}"
git -C apps/lina-plugins switch -c "$sub_fix_branch"
```

若`fix/${main_branch}`已存在，则`switch`该分支并`rebase origin/main`（冲突则停）。

阶段二完成后停止：

```text
请 review 主仓库 PR。若 CI/评审需改子模块：在 fix/<主仓分支> 上修改并走阶段 2.5，禁止在子模块 main 上提交。
主仓库 PR 合并后回复“已合并”或“继续收尾”。
```

## 阶段 2.5：主仓 PR 存活期修复（可选）

触发：主仓库`PR`已开未合，且 CI/评审需要改代码。

**门禁：**

- 子模块当前分支为`main`且要改内容 → 先切/建`fix/<main_branch>`，禁止在`main`上`commit`/`push`。
- 主仓库修复提交到已有`main_branch`并`push`更新同一`PR`（用户要求新开`PR`除外）。

子模块需改时：

1. 确保在`fix/<main_branch>`（基于最新`origin/main`）。
2. 提交 → `PR`前`rebase` → 推送 → 创建子模块 follow-up`PR`。
3. 停止，等用户合并该子模块`PR`。
4. 确认`MERGED`后：

```bash
git -C apps/lina-plugins fetch origin main
git -C apps/lina-plugins switch main
git -C apps/lina-plugins merge --ff-only origin/main
# 主仓库任务分支上更新指针
git switch "$main_branch"
git add apps/lina-plugins
git commit -m "chore(plugins): update lina-plugins submodule"
# PR 前 rebase 门禁后
git push origin "$main_branch"
# 子模块回到 fix 分支待命
git -C apps/lina-plugins switch -c "fix/${main_branch}" 2>/dev/null \
  || git -C apps/lina-plugins switch "fix/${main_branch}"
git -C apps/lina-plugins merge --ff-only origin/main
```

5. 仅改主仓库时：在`main_branch`提交、`rebase`、`push`即可。
6. 仍失败则重复本阶段；通过后仍等用户合并主仓库`PR`，不得提前阶段三。

## 阶段三：合并后恢复分支

条件：本流程创建过的子模块/`follow-up`与主仓库`PR`均已`MERGED`；`original_*`非空；无未提交修复。

1. 再确认`MERGED`。

```bash
gh pr view "<submodule-pr-or-followup>" \
  --repo "<submodule-owner>/<submodule-repo>" \
  --json number,state,mergeCommit,url
gh pr view "<framework-pr-number-or-url>" \
  --repo linaproai/linapro \
  --json number,state,mergeCommit,url
```

2. 刷新`main`。

```bash
git fetch origin main
git -C apps/lina-plugins fetch origin main
```

3. 子模块回原始分支并包含`origin/main`。

```bash
git -C apps/lina-plugins switch "$original_sub_branch"
git -C apps/lina-plugins merge --ff-only origin/main
```

`--ff-only`失败则按冲突处理；需人工选择时停止。

4. 主仓库同理。

```bash
git switch "$original_main_branch"
git merge --ff-only origin/main
```

5. 校验。

```bash
git status --short --branch
git merge-base --is-ancestor origin/main HEAD
git -C apps/lina-plugins status --short --branch
git -C apps/lina-plugins merge-base --is-ancestor origin/main HEAD
```

6. 相对远端领先则推送；无上游用`-u`；落后或分叉则停止。

```bash
git rev-list --left-right --count "origin/$original_main_branch"...HEAD
git -C apps/lina-plugins rev-list --left-right --count "origin/$original_sub_branch"...HEAD
git push origin "$original_main_branch"
git -C apps/lina-plugins push origin "$original_sub_branch"
```

## PR 正文

通用：`Summary`、`Tests`（未跑则写`Not run by this PR creation step.`）、`Related Issue`（识别到用`Resolves ...`，否则说明无关联且不阻塞）。

主仓库另加`Submodule`：`apps/lina-plugins`从何 SHA 更新到何 SHA。

## 输出

**阶段一：** 子模块分支/主题/推送目标、`rebase`与`origin/main..HEAD`、`PR`地址、等待合并提醒。

**阶段二：** 子模块对齐的`main` SHA、主仓库分支/主题、`rebase`检查、`PR`地址与可合并性、子模块当前 fix 分支、工作区是否干净、测试是否运行。

**阶段 2.5：** 修复落在哪一仓库/分支、子模块 follow-up`PR`（如有）、主仓库是否已 bump 指针并 push、下一步等待项。

**阶段三：** 已合并`PR`与 merge commit、两侧当前分支与是否含最新`main`/是否已推送、工作区是否干净、测试是否运行。
