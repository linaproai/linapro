## 1. Deterministic Archive Workflow

- [x] 1.1 Add shared monthly OpenSpec auto-archive prompt execution for each AI tool runtime.
- [x] 1.2 Wire Codex, Claude Code, and GitHub Copilot reusable workflows to run auto archive before archive consolidation.
- [x] 1.3 Stop immediately when auto archive leaves completed active changes unarchived.
- [x] 1.4 Upgrade artifact upload workflow actions away from the Node 20 runtime generation.

## 2. OpenSpec Archive Blocker Fix

- [x] 2.1 Fix `remove-sqlite-support` so `openspec archive -y remove-sqlite-support` can apply against the current baseline.

## 3. Verification

- [x] 3.1 Run OpenSpec validation for this change and `remove-sqlite-support`.
- [x] 3.2 Run workflow YAML/action validation and shell syntax checks for modified CI files.
- [x] 3.3 Run monthly archive workflow validation for tool-runtime auto archive behavior and fail-fast blockers.
- [x] 3.4 Record i18n, cache, data permission, REST API, E2E, and Go production code impact.
- [x] 3.5 Run `lina-review` for the CI/OpenSpec governance change.

## Feedback

- [x] **FB-1**: Monthly OpenSpec archive failed because AI auto archive returned success while completed active changes remained unarchived.
- [x] **FB-2**: `remove-sqlite-support` cannot be archived because its OpenSpec delta removes a requirement header that no longer exists in the baseline spec.
- [x] **FB-3**: Manual monthly OpenSpec archive dispatch is skipped when triggered from a non-default branch.
- [x] **FB-4**: Manual monthly OpenSpec archive dispatch should run against the selected source branch and create a PR back to that same branch.
- [x] **FB-5**: Copilot archive consolidation can produce invalid OpenSpec and block the already validated deterministic archive PR.
- [x] **FB-6**: Archive branch push succeeds but repository policy blocks GitHub Actions from creating the pull request.
- [x] **FB-7**: Monthly OpenSpec archive workflow continues after archive consolidation or pull request steps fail.
- [x] **FB-8**: Monthly OpenSpec auto archive must use the selected AI tool runtime like archive consolidation instead of a standalone composite action.

## Verification Notes

- FB-1 修复：最初新增 `.github/actions/monthly-openspec-auto-archive`，使用固定 OpenSpec CLI 版本执行自动归档。FB-8 已按最新要求移除该 standalone composite action，改为通过所选 AI tool runtime 执行共享 auto-archive prompt。
- FB-1 workflow 接入：`.github/workflows/monthly-openspec-archive-{codex,cc,copilot}.yml` 先执行 auto archive，再检测 OpenSpec diff；只有存在 diff 时才执行 archive consolidation。FB-7 已进一步收紧失败处理：auto archive 报告失败时 workflow 直接停止，不再延迟到 PR finalization 后失败。
- FB-1 prompt 口径：FB-8 已恢复 `.github/prompts/monthly-openspec-auto-archive.zh-CN.md`，基础自动归档和 archive consolidation 都通过所选工具运行时执行共享 prompt。
- FB-1 artifact 升级：所有 `actions/upload-artifact@v4` 已升级为 `actions/upload-artifact@v7`，静态扫描确认 `.github` 中不再存在 `upload-artifact@v4`。
- FB-2 修复：将 `remove-sqlite-support` 中与当前主规范不匹配的 REMOVED/MODIFIED header 调整为现有 baseline requirement，删除已经被当前 baseline 吸收且不存在的 SQLite 专属 REMOVED delta；新增 header mismatch 检查确认该变更所有 MODIFIED/REMOVED requirement 标题均存在于当前主规范。
- 验证通过：`openspec validate remove-sqlite-support --strict`。
- 验证通过：`openspec validate deterministic-monthly-openspec-archive --strict`。
- 验证通过：`ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f); puts "ok #{f}" }' .github/workflows/*.yml .github/actions/*/action.yml`。
- 验证通过：`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/monthly-openspec-archive.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml .github/workflows/reusable-e2e-tests.yml .github/workflows/reusable-host-only-build-smoke.yml .github/workflows/reusable-redis-cluster-smoke.yml`。
- 验证通过：对 `remove-sqlite-support` 执行 Node header mismatch 扫描，所有 MODIFIED/REMOVED requirement 标题均匹配当前 `openspec/specs/<capability>/spec.md`。
- 验证说明：早期 standalone composite action 曾通过临时 action smoke；FB-8 已删除该 action，当前验证以 workflow YAML/actionlint、OpenSpec strict 和静态扫描为准。
- 验证通过：`git diff --check -- .github openspec/changes/deterministic-monthly-openspec-archive openspec/changes/remove-sqlite-support/specs`。
- i18n 影响：本次仅修改 GitHub Actions workflow、CI prompt、OpenSpec 变更文档和 OpenSpec delta，不新增、修改或删除用户运行时可见文案、前端语言包、宿主/插件 `manifest/i18n` 或 apidoc i18n JSON。
- 缓存一致性影响：本次不修改运行时业务缓存、缓存键、失效触发、分布式协调或跨实例一致性逻辑；新增的 CI action 只在 GitHub runner 工作区内执行 OpenSpec 归档，不涉及生产缓存。
- 数据权限影响：本次不新增或修改 HTTP/API 数据操作接口、服务数据访问路径、插件宿主服务适配器或聚合统计，不影响角色数据权限边界。
- REST API 影响：本次不新增或修改 REST API。
- E2E 影响：本次为 CI/OpenSpec 治理修复，不涉及用户可观察页面、路由、表单、表格或端到端业务流程；使用 OpenSpec、workflow/actionlint、header mismatch 扫描和临时归档 smoke 作为治理验证。
- Go 生产代码影响：本次不新增或修改 Go 生产代码，不触发后端 Go 编译门禁。`actionlint` 通过 `go run` 执行属于外部验证工具，不改变仓库 Go 源码。
- Review：已按 `lina-review` 口径完成审查。审查范围来源包括 `git status --short`、`git ls-files --others --exclude-standard`、`openspec status --change deterministic-monthly-openspec-archive --json`、`openspec status --change remove-sqlite-support --json`、`.github` 与目标 OpenSpec 文件 diff、OpenSpec strict 校验、workflow/action YAML 解析、actionlint、`remove-sqlite-support` header mismatch 扫描和临时 action smoke。FB-8 已将基础归档改为与归档聚合相同的工具运行时执行模式；三条 monthly reusable workflow 均先通过所选 AI runtime 执行 auto archive，再按需用同一 runtime 执行 archive consolidation；`upload-artifact@v4` 已清理。FB-7 后续已将失败处理收紧为任何归档、聚合、校验、推送或 PR 步骤失败均不继续。严重问题 0；警告 0。当前工作区仍存在与本次无关的 Go、前端、测试与其他 OpenSpec 改动，本次未修改或回退。
- FB-3 修复：`.github/workflows/monthly-openspec-archive.yml` 的 router `detect` job 现在允许 `workflow_dispatch` 从任意 branch ref 进入，同时继续只让定时触发在默认分支运行。
- FB-3 规范更新：`Manual archive run` 场景明确手动触发可来自任意暴露该 workflow 的 branch ref。
- FB-3 验证通过：`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/monthly-openspec-archive.yml`。
- FB-3 验证通过：`ruby -e 'require "yaml"; YAML.load_file(ARGV[0]); puts "ok #{ARGV[0]}"' .github/workflows/monthly-openspec-archive.yml`。
- FB-3 验证通过：`openspec validate deterministic-monthly-openspec-archive --strict`。
- FB-3 验证通过：`git diff --check -- .github/workflows/monthly-openspec-archive.yml openspec/changes/deterministic-monthly-openspec-archive`。
- FB-3 影响评估：本次仅修改 GitHub Actions router 条件和 OpenSpec 文档；不新增或修改 Go 生产代码、前端页面、REST API、运行时 i18n 资源、业务缓存、数据权限逻辑或用户可观察应用流程，因此不触发后端 Go 编译门禁和 E2E 测试。验证方式采用 workflow/actionlint、YAML 解析、OpenSpec strict 校验和 diff whitespace 检查。
- FB-3 Review：已按 `lina-review` 口径完成审查。确认手动触发入口不再受 `github.ref == default_branch` 限制；定时触发仍保留默认分支限制；手动触发限定为 branch ref，避免 tag ref 进入 PR base 语义。严重问题 0；警告 0。
- FB-4 修复：router 新增 `Resolve Target Branch` 步骤，从 `github.ref_name` 解析本次触发分支，输出 `target_branch` 和带安全化源分支标识的 `pr_branch`（格式为 `automation/monthly-openspec-archive-<branch-slug>`）。router 的检测 checkout 改为 `target_branch`，并将 `target_branch`、`pr_branch` 传入 Codex、Claude Code 和 Copilot reusable workflow。
- FB-4 工具 workflow 接入：`.github/workflows/monthly-openspec-archive-{codex,cc,copilot}.yml` 新增 required inputs `target_branch` 和 `pr_branch`；三个 workflow 均 checkout `target_branch` 执行确定性归档和按需聚合，并在 `Finalize Archive Pull Request` 中使用 `base-branch: target_branch`、`pr-branch: pr_branch`。
- FB-4 规范更新：`Manual archive run` 场景明确手动触发分支就是检测、执行和 PR base 分支，PR 来源分支必须包含该触发分支的安全化标识。
- FB-4 影响评估：本次仅修改 GitHub Actions workflow 和 OpenSpec 文档；不新增或修改 Go 生产代码、前端页面、REST API、运行时 i18n 资源、业务缓存、数据权限逻辑或用户可观察应用流程，因此不触发后端 Go 编译门禁和 E2E 测试。验证方式采用 workflow/actionlint、YAML 解析、OpenSpec strict 校验和 diff whitespace 检查。
- FB-4 验证通过：`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/monthly-openspec-archive.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml`。
- FB-4 验证通过：`ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f); puts "ok #{f}" }' .github/workflows/monthly-openspec-archive.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml`。
- FB-4 验证通过：`openspec validate deterministic-monthly-openspec-archive --strict`。
- FB-4 验证通过：`git diff --check -- .github/workflows/monthly-openspec-archive.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml openspec/changes/deterministic-monthly-openspec-archive`。
- FB-4 验证通过：Node slug smoke 确认 `john-e2e-enhance -> automation/monthly-openspec-archive-john-e2e-enhance`、`feature/archive test -> automation/monthly-openspec-archive-feature-archive-test`、`release/2026.05 -> automation/monthly-openspec-archive-release-2026.05`。
- FB-4 Review：已按 `lina-review` 口径完成审查。确认手动触发时 workflow 使用触发 branch ref 作为检测 checkout、归档 checkout 和归档 PR base；PR head 分支使用 `automation/monthly-openspec-archive-<branch-slug>`，包含源分支标识并替换非法字符；定时触发仍仅允许默认分支，并自然生成默认分支对应 PR head。严重问题 0；警告 0。
- FB-5 失败分析：GitHub Actions run `26200881048` 在 `john-e2e-enhance` 手动触发，auto archive 与 `Validate OpenSpec After Auto Archive` 均成功，`Run Lina Archive Consolidate` 返回成功，但 `Validate OpenSpec After Archive Consolidate` 失败，导致 `Finalize Archive Pull Request` 被跳过。GitHub job log 和 artifact API 当前返回 403，无法读取私有详细日志；从 job step 结论可确认失败范围在 AI 聚合后的 OpenSpec 校验。
- FB-5 修复：`.github/workflows/monthly-openspec-archive-{codex,cc,copilot}.yml` 在确定性归档校验通过后创建 `openspec` 快照；`Run Lina Archive Consolidate` 和聚合后校验改为 `continue-on-error: true`；当 AI 聚合命令失败或聚合后 OpenSpec 校验失败时，workflow 记录 `git status`、`git diff -- openspec` 和 `openspec validate --all` 日志到对应 AI 日志目录，随后恢复确定性归档快照并重新校验，通过后继续执行 `Finalize Archive Pull Request`。
- FB-5 日志复核：用户提供的 `/Users/john/Downloads/job-logs.txt` 确认 Copilot 在聚合阶段创建了临时活跃变更 `openspec/changes/archive-consolidation`，但最终没有清理；`openspec validate --all` 输出 `✗ change/archive-consolidation`，总计 `92 passed, 1 failed (93 items)`。已新增 `Check Archive Consolidate Temporary Change Cleanup` 步骤，显式检测该临时变更残留并触发诊断与回退。
- FB-5 规范更新：AI archive consolidation 被定义为可选增强阶段；失败或产生无效 OpenSpec 时不得阻塞已经通过校验的 deterministic archive PR，必须记录诊断、回滚聚合结果并继续归档。
- FB-5 影响评估：本次仅修改 GitHub Actions workflow 和 OpenSpec 文档；不新增或修改 Go 生产代码、前端页面、REST API、运行时 i18n 资源、业务缓存、数据权限逻辑或用户可观察应用流程，因此不触发后端 Go 编译门禁和 E2E 测试。验证方式采用 workflow/actionlint、YAML 解析、OpenSpec strict 校验和 diff whitespace 检查。
- FB-6 失败分析：GitHub Actions run `26203378371` 已成功 push `automation/monthly-openspec-archive-john-e2e-enhance`，但 `gh pr create` 返回 `GraphQL: GitHub Actions is not permitted to create or approve pull requests (createPullRequest)`，说明仓库 Actions 设置不允许 `GITHUB_TOKEN` 创建 PR。
- FB-6 修复：`.github/actions/monthly-openspec-finalize-pr/action.yml` 在 `gh pr create` 或 `gh pr edit` 失败时检查该仓库策略错误；若归档分支已经成功 push，则输出 base/head 分支和手动 PR URL 到日志与 step summary，并让 workflow 成功结束。其他 PR 命令错误仍按失败处理。
- FB-6 规范更新：新增仓库策略阻止 PR 创建场景，明确已推送有效归档分支时 workflow 应输出手动 PR 链接并成功结束。
- FB-6 影响评估：本次仅修改 GitHub Actions composite action 和 OpenSpec 文档；不新增或修改 Go 生产代码、前端页面、REST API、运行时 i18n 资源、业务缓存、数据权限逻辑或用户可观察应用流程，因此不触发后端 Go 编译门禁和 E2E 测试。验证方式采用 actionlint、YAML 解析、OpenSpec strict 校验、diff whitespace 检查和 shell 语法检查。
- FB-6 验证通过：`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/monthly-openspec-archive.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml`。
- FB-6 验证通过：`ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f); puts "ok #{f}" }' .github/actions/monthly-openspec-finalize-pr/action.yml .github/workflows/monthly-openspec-archive.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml`。
- FB-6 验证通过：`openspec validate deterministic-monthly-openspec-archive --strict`。
- FB-6 验证通过：`git diff --check -- .github/actions/monthly-openspec-finalize-pr/action.yml .github/workflows/monthly-openspec-archive.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml openspec/changes/deterministic-monthly-openspec-archive`。
- FB-6 验证通过：从 `.github/actions/monthly-openspec-finalize-pr/action.yml` 提取 `Create or Update Archive Pull Request` 的 bash 脚本并执行 `bash -n`。
- FB-6 验证通过：模拟 `GraphQL: GitHub Actions is not permitted to create or approve pull requests (createPullRequest)` 输出，确认策略错误匹配分支可识别该错误。
- FB-6 Review：已按 `lina-review` 口径完成审查。确认归档分支 push 仍是硬失败边界；只有 push 成功后的 PR 创建/更新被仓库策略拒绝时才降级为手动 PR 链接；其他 `gh pr` 错误仍会 `exit 1`。严重问题 0；警告 0。
- FB-7 失败分析：GitHub Actions run `26204364117` 的公开 job 页面显示 `Run Lina Archive Consolidate` step 报错 `Process completed with exit code 1`，但 job 仍显示 success。GitHub job logs API 返回 403 `Must have admin rights to Repository`，无法下载完整私有日志；结合本地 workflow 配置可确认根因是三套工具 workflow 对 `Run Lina Archive Consolidate`、临时变更清理和聚合后 OpenSpec 校验使用 `continue-on-error: true`，并在失败后恢复 deterministic archive snapshot 继续执行 `Finalize Archive Pull Request`。`monthly-openspec-finalize-pr` 还会把仓库策略导致的 PR 创建/更新失败降级为手动 PR 链接并成功结束。
- FB-7 修复：`.github/workflows/monthly-openspec-archive-{codex,cc,copilot}.yml` 已移除 archive consolidation、临时变更清理和聚合后校验的 `continue-on-error`，删除失败诊断后恢复 deterministic snapshot 的回滚步骤，删除 deterministic archive 失败后 PR finalization 再失败的延迟失败门禁，并移除日志上传步骤的 `if: always()`，确保前序步骤失败后不再继续执行后续归档步骤。
- FB-7 修复：auto archive 在单个候选归档失败、任务计数不一致、归档后仍保持活跃或归档后列表复查失败时必须以非 0 退出，不再继续处理后续候选；FB-8 已将该规则下沉到 `.github/prompts/monthly-openspec-auto-archive.zh-CN.md` 和工具运行时执行。
- FB-7 修复：`.github/actions/monthly-openspec-finalize-pr/action.yml` 不再吞掉 `gh pr create` 或 `gh pr edit` 的仓库策略错误；PR 创建或更新失败会按 `set -euo pipefail` 直接让步骤失败。
- FB-7 规范更新：`monthly-openspec-archive` 增量规范明确任一自动归档、归档聚合、临时变更清理、OpenSpec 校验、范围保护、提交、推送、PR 创建或更新步骤失败时必须立即停止；禁止恢复早期归档快照后继续 PR 收尾。
- FB-7 影响评估：本次仅修改 GitHub Actions workflow/composite action 和 OpenSpec 文档；不新增或修改 Go 生产代码、前端页面、REST API、运行时 i18n 资源、业务缓存、数据权限逻辑或用户可观察应用流程，因此不触发后端 Go 编译门禁和 E2E 测试。验证方式采用 actionlint、YAML 解析、OpenSpec strict 校验、diff whitespace 检查和静态扫描确认无降级路径残留。
- FB-7 验证通过：`ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f); puts "ok #{f}" }' .github/actions/monthly-openspec-finalize-pr/action.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml`。
- FB-7 验证通过：`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/monthly-openspec-archive.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml`。
- FB-7 验证通过：`openspec validate deterministic-monthly-openspec-archive --strict`。
- FB-7 验证通过：`git diff --check -- .github/actions/monthly-openspec-finalize-pr/action.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml openspec/changes/deterministic-monthly-openspec-archive`。
- FB-7 验证通过：静态扫描 `.github/workflows/monthly-openspec-archive-{codex,cc,copilot}.yml` 与相关 composite actions，确认 workflow/action 代码中不再存在 `continue-on-error`、`if: always()`、archive consolidation 失败后恢复 deterministic snapshot、PR 创建/更新失败降级为手动 PR 链接，或基于 `steps.auto-archive.outputs.had-failures` 的延迟失败门禁。
- FB-7 Review：已按 `lina-review` 口径完成审查。审查范围来源包括 GitHub run `26204364117` 公开注解、GitHub job logs API 403 结果、本地 workflow/action diff、OpenSpec strict 校验、YAML 解析、actionlint、diff whitespace 检查和降级路径静态扫描。确认此次失败根因是 `Run Lina Archive Consolidate` 返回 exit 1 后被 `continue-on-error` 吞掉，随后 workflow 通过 snapshot restore/finalize 路径继续；修复后三套 AI workflow 均恢复 GitHub Actions 默认失败即停语义，聚合失败、临时变更未清理、OpenSpec 校验失败、PR 创建/更新失败都会阻断后续步骤。未修改生产 Go、前端运行时、REST API、数据权限、缓存或 i18n 资源。严重问题 0；警告 0。
- FB-8 修复：新增 `.github/prompts/monthly-openspec-auto-archive.zh-CN.md`，明确 CI 自动归档必须使用 `.agents/skills/lina-auto-archive/SKILL.md` 规则、只处理活跃变更、只归档安全完成项、通过 OpenSpec CLI 执行 `openspec archive -y "<change-name>"`、归档后复查活跃列表并保持变更范围在 `openspec/**` 内。
- FB-8 修复：删除 `.github/actions/monthly-openspec-auto-archive/action.yml`；`.github/workflows/monthly-openspec-archive-{codex,cc,copilot}.yml` 不再调用 standalone composite auto archive action。三套 workflow 均先准备对应工具运行时，再通过同一个运行时执行 `Run Lina Auto Archive`，随后检测 OpenSpec diff、校验 auto archive 结果，并仅在产生归档变更时继续执行 `Run Lina Archive Consolidate`。
- FB-8 工具运行时：Codex 使用 `loads/codex:latest` 和 `codex exec` 执行 auto archive 与 archive consolidate；Claude Code 使用 `loads/cc:latest` 和 `run-cc-task` 执行 auto archive 与 archive consolidate；GitHub Copilot CLI 使用 `@github/copilot` 和 `copilot -p` 执行 auto archive 与 archive consolidate。
- FB-8 规范更新：`monthly-openspec-archive` 增量规范将基础归档从 shared deterministic archive action 改为 selected tool runtime auto archive，并要求 auto archive 和 archive consolidation 都通过对应工具运行时执行共享 prompt。
- FB-8 影响评估：本次仅修改 GitHub Actions workflow、CI prompt 和 OpenSpec 文档；不新增或修改 Go 生产代码、前端页面、REST API、运行时 i18n 资源、业务缓存、数据权限逻辑或用户可观察应用流程，因此不触发后端 Go 编译门禁和 E2E 测试。验证方式采用 workflow/actionlint、YAML 解析、OpenSpec strict 校验、diff whitespace 检查和结构化静态扫描。
- FB-8 验证通过：`ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f); puts "ok #{f}" }' .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml .github/actions/monthly-openspec-finalize-pr/action.yml`。
- FB-8 验证通过：`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/monthly-openspec-archive.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml`。
- FB-8 验证通过：`openspec validate deterministic-monthly-openspec-archive --strict`。
- FB-8 验证通过：`git diff --check -- .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml .github/prompts/monthly-openspec-auto-archive.zh-CN.md .github/actions/monthly-openspec-auto-archive/action.yml openspec/changes/deterministic-monthly-openspec-archive`。
- FB-8 验证通过：Ruby 结构扫描确认三套 reusable workflow 均包含 `Run Lina Auto Archive`、`Detect Archive Changes`、`Validate OpenSpec After Auto Archive`、`Run Lina Archive Consolidate`，且 `Run Lina Auto Archive` 均读取 `.github/prompts/monthly-openspec-auto-archive.zh-CN.md`。
- FB-8 验证通过：静态扫描确认 `.github` 中不再存在对 `.github/actions/monthly-openspec-auto-archive` 的引用，也不再存在 `Run Deterministic OpenSpec Auto Archive`、`continue-on-error` 或 `if: always()` 的失败继续路径。
- FB-8 Review：已按 `lina-review` 口径完成审查。审查范围来源包括 `git status --short`、`.github` 与 OpenSpec diff、OpenSpec strict 校验、YAML 解析、actionlint、diff whitespace 检查和 workflow 结构扫描。确认基础 auto archive 和 archive consolidation 已统一为 selected AI tool runtime 执行模式：Codex 走 `loads/codex:latest`，Claude Code 走 `loads/cc:latest`，Copilot 走 `@github/copilot`；standalone auto archive composite action 已删除且无引用残留。未修改生产 Go、前端运行时、REST API、数据权限、缓存或 i18n 资源。严重问题 0；警告 0。
