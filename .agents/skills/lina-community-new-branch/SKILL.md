---
name: lina-community-new-branch
description: >-
  手动触发：将主仓库与 apps/lina-plugins 对齐最新 main 后，
  按提示在两侧各建独立分支；分支就绪后继续处理提示词中的后续请求。
  必须用户手动触发，禁止自动触发该技能。
---

# Lina Community New Branch

将主仓库与子模块切到最新 `main`，再按用户提示在两侧各创建独立任务分支。**分支创建（或切换）成功后，必须继续处理同一条提示词中除“建分支”以外的后续请求**，不得只建分支就结束。

## 核心规则

1. **只手动触发**：用户必须明确调用本技能或给出分支提示；禁止因看到脏工作区、新需求或对话上下文而自动启动。
2. 主仓库默认当前工作目录，远端默认 `linaproai/linapro`。
3. 子模块路径固定 `apps/lina-plugins`，远端从 `.gitmodules` / `git -C apps/lina-plugins remote -v` 读取，通常 `linaproai/official-plugins`。
4. 两侧目标基线均为各自 `origin/main`；创建分支前必须 `fetch` 且快进对齐 `main`。
5. 主仓库与子模块**各自**创建同名任务分支（除非用户明确要求不同名）。
6. **禁止**丢弃未提交变更、`--force`、`reset --hard`、`clean -fd`；脏工作区先报告并停止，除非用户明确授权处理方式。
7. **禁止**在本技能的“建分支阶段”内 `commit` / `push` / 创建 PR。后续请求若本身需要提交/PR，应改走对应技能（如 `lina-community-commit-push-and-pr`），而非在本技能建分支步骤中执行。
8. 分支名从用户提示推导；用户已给合法分支名则直接用。命名：小写、`/` 分层、数字字母与连字符；推荐 `feature/<slug>`、`fix/<slug>`、`chore/<slug>`。
9. 分支已存在：本地有则切过去并尝试包含最新 `origin/main`（优先 ff-only merge）；远端有本地无则 `switch --track`；冲突或分叉则停止交用户。
10. 对齐 `main` 失败（非快进、冲突、分离 HEAD 无法安全处理）立即停止，不硬切。
11. **分支就绪后继续后续请求**（见「阶段 B」）：建分支只是入口，不是完整工作的终点。

## 输入

识别用户给出的：

- 技能名 / `新建分支` / `new branch`
- 分支名（如 `feature/cloud-storage`）
- 或任务描述（据此生成 slug 分支名）
- **同一条提示中的后续请求**（如：实现某功能、修 bug、写 proposal、继续某 OpenSpec 变更等）

未给出分支线索时先问清再执行，不得自造无依据分支名。

### 提示词拆分

进入执行前，把用户消息拆成两部分并记下来：

| 部分 | 含义 | 示例 |
| --- | --- | --- |
| **A. 分支意图** | 要建/切的分支名或可推导 slug 的描述 | `新建分支 feature/foo`、`基于 main 开 fix/bar` |
| **B. 后续请求** | 去掉建分支语句后仍需完成的工作 | `实现插件卸载禁用`、`/opsx:propose ...`、`继续 tasks.md` |

规则：

1. 若整条消息**只有**建分支意图、没有可执行的后续请求 → 完成阶段 A 后按「仅建分支」输出并结束。
2. 若同时有后续请求 → **必须**在阶段 A 成功后立刻进入阶段 B，不得停在分支结果报告等待用户再说一次。
3. 后续请求语义不清时，先完成阶段 A，再简短确认后续范围；确认后继续执行，不要把分支结果当作任务完成。
4. 用户只说“新建分支做 X”且 X 本身就是任务描述时：用 X 生成分支名，**同时把 X 当作后续请求**（阶段 B 继续做 X），不要当成“只建分支”。

## 前置检查

```bash
git rev-parse --show-toplevel
git status --short --branch
git branch --show-current
git remote -v
cat .gitmodules
git submodule status apps/lina-plugins
git -C apps/lina-plugins status --short --branch
git -C apps/lina-plugins branch --show-current
git -C apps/lina-plugins remote -v
```

阻断条件（任一成立则停止并报告；**不得进入阶段 B**）：

- 主仓库或子模块工作区不干净（有未提交改动）
- 主仓库或子模块处于 rebase/merge/cherry-pick 进行中
- 无法解析子模块路径或远端
- 无法从提示得到合法分支名

## 分支名

1. 用户显式给出 → 规范化后使用（去首尾空白，禁止空格与非法字符）。
2. 仅有描述 → 提炼英文/拼音 slug，前缀按意图选 `feature/` / `fix/` / `chore/`，默认 `feature/`。
3. 最终确认将使用的 `branch_name`，再执行后续步骤。

## 阶段 A：对齐 main 并创建任务分支

### 1. 拉取最新 main

```bash
git fetch origin main
git -C apps/lina-plugins fetch origin main
```

### 2. 主仓库对齐 main

```bash
git switch main
git merge --ff-only origin/main
```

`--ff-only` 失败则停止。

### 3. 子模块对齐 main

```bash
git -C apps/lina-plugins switch main
git -C apps/lina-plugins merge --ff-only origin/main
```

`--ff-only` 失败则停止。此时主仓库可能显示子模块指针变化，**不要**为此提交。

### 4. 创建任务分支

主仓库：

```bash
if git show-ref --verify --quiet "refs/heads/$branch_name"; then
  git switch "$branch_name"
  git merge --ff-only origin/main   # 失败则停止
elif git ls-remote --exit-code --heads origin "$branch_name" >/dev/null 2>&1; then
  git fetch origin "$branch_name"
  git switch --track "origin/$branch_name"
else
  git switch -c "$branch_name"
fi
```

子模块：

```bash
if git -C apps/lina-plugins show-ref --verify --quiet "refs/heads/$branch_name"; then
  git -C apps/lina-plugins switch "$branch_name"
  git -C apps/lina-plugins merge --ff-only origin/main   # 失败则停止
elif git -C apps/lina-plugins ls-remote --exit-code --heads origin "$branch_name" >/dev/null 2>&1; then
  git -C apps/lina-plugins fetch origin "$branch_name"
  git -C apps/lina-plugins switch --track "origin/$branch_name"
else
  git -C apps/lina-plugins switch -c "$branch_name"
fi
```

### 5. 校验

```bash
git status --short --branch
git rev-parse --short HEAD
git merge-base --is-ancestor origin/main HEAD && echo "main: contains origin/main"
git -C apps/lina-plugins status --short --branch
git -C apps/lina-plugins rev-parse --short HEAD
git -C apps/lina-plugins merge-base --is-ancestor origin/main HEAD && echo "sub: contains origin/main"
```

阶段 A 失败 → 报告原因并停止，**不要**假装进入后续请求。

阶段 A 成功 → 先给出简短分支结果（可并入最终回复），**有后续请求则立即进入阶段 B**。

## 阶段 B：继续处理提示词后续请求

### 何时进入

- 阶段 A 已成功（两侧已在目标分支，且含最新 `origin/main` 或已按规则处理已有分支）。
- 提示词拆分得到非空的**后续请求**，或用户描述本身即“在新分支上完成 X”。

### 执行要求

1. **同一回合内继续**：不要只输出“分支已创建，请告诉我下一步”；应直接开始做后续请求。
2. **按后续请求类型路由**，例如：
   - 新功能 / 缺陷 / 改进 → 命中 OpenSpec 时走 `openspec-propose` / `openspec-apply-change` / `lina-feedback` 等对应技能与项目规则。
   - 用户已点名其它技能或 slash 命令 → 读取并执行该技能。
   - 纯调查 / 读代码 / 回答问题 → 在当前新分支上直接完成，无需强行开变更。
3. **遵守项目规范**：阶段 B 中的代码、文档、测试变更仍须按 `AGENTS.md` 与命中的 `.agents/rules/*` 执行；本技能不豁免任何规则域。
4. **工作区边界**：阶段 B 可以正常改代码与写文档；阶段 A 的“不 commit / 不 push / 不开 PR”约束仍适用于“仅为建分支而做的动作”。若后续请求明确要求提交或开 PR，再调用 `lina-community-commit-push-and-pr` 等专门技能，且仍须用户按该技能的手动触发规则授权。
5. **范围克制**：只做提示词后续请求所覆盖的工作；不要借机扩大范围或顺手重构无关模块。
6. **阶段 B 失败**：报告已完成的分支状态 + 后续请求失败点；分支保留，不回滚、不 fortce 清理。

### 无后续请求时

仅输出阶段 A 结果，明确说明“未检测到后续请求，分支已就绪”，然后结束。

## 输出

### 仅建分支

```markdown
## 新分支结果

- 分支名：`<branch_name>`
- 主仓库：当前分支 / HEAD / 是否含 `origin/main` / 工作区
- 子模块：当前分支 / HEAD / 是否含 `origin/main` / 工作区
- 动作摘要：对齐 main / 新建或切换
- 未执行：commit、push、PR
- 后续请求：无（本次仅建分支）
```

### 建分支 + 继续后续请求

先简要给出分支结果，再汇报后续请求进展（或最终完成情况）：

```markdown
## 新分支结果

- 分支名：`<branch_name>`
- 主仓库 / 子模块：…（同上，可从简）
- 后续请求：已继续处理 → `<一句话概括后续请求>`

## 后续请求进展

- …（按实际技能/实现结果填写：做了什么、验证了什么、还剩什么）
```

## 护栏

- 只手动触发
- 脏工作区先停
- 只快进对齐 `main`
- 两侧独立同名分支
- 建分支阶段不提交、不推送、不开 PR
- 不丢弃用户本地改动
- **有后续请求时，分支成功后必须继续处理，禁止只建分支就结束**
- 后续请求失败不撤销已创建的分支
