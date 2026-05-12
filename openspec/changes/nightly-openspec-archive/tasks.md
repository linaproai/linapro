## 1. OpenSpec 与 Codex 配置治理

- [x] 1.1 调整 `.gitignore`，允许提交 `.github/codex/config.toml` 和无密钥模板，同时继续忽略真实认证文件
- [x] 1.2 新增或整理 `.github/codex` 配置模板，确保 workflow 可复制配置但不会提交真实 `OPENAI_API_KEY`

## 2. Nightly Workflow 实现

- [x] 2.1 新增 `Nightly OpenSpec Archive` GitHub Actions workflow，支持 schedule 和 `workflow_dispatch`
- [x] 2.2 在 workflow 中通过临时 `CODEX_HOME` 注入 Codex 配置和 `OPENAI_API_KEY` secret
- [x] 2.3 实现 OpenSpec 完成状态预检查，无可归档变更时跳过 Codex 执行
- [x] 2.4 通过 `loads/codex:latest` 依次执行 `lina-auto-archive` 和条件性的 `lina-archive-consolidate`
- [x] 2.5 实现 OpenSpec 校验、变更范围保护和自动提交逻辑

## 3. 验证与审查

- [x] 3.1 运行 `openspec validate nightly-openspec-archive --strict`
- [x] 3.2 对新增 workflow 和 Codex 模板执行静态检查，确认 YAML、JSON 和 TOML 格式有效
- [x] 3.3 记录 i18n、缓存一致性、数据权限和 E2E 影响判断
- [x] 3.4 完成实现后调用 `lina-review` 审查本次变更

## Verification Notes

- [x] 2026-05-12: OpenSpec 与静态格式验证通过：`openspec validate nightly-openspec-archive --strict`；`python3 -m json.tool .github/codex/auth.template.json >/dev/null`；`python3` 使用 `tomllib` 解析 `.github/codex/config.toml`；`ruby -e 'require "yaml"; YAML.load_file(".github/workflows/nightly-openspec-archive.yml")'`；`git diff --check -- .github/workflows/nightly-openspec-archive.yml .github/codex .gitignore openspec/changes/nightly-openspec-archive`。确认 `.github/codex/auth.json` 仍被 `.gitignore` 忽略，`.github/codex/config.toml` 与 `.github/codex/auth.template.json` 可纳入版本控制。
- [x] 2026-05-12: `lina-review` 审查完成。审查范围来源：`git status --short`、`git ls-files --others --exclude-standard`、`openspec status --change nightly-openspec-archive --json`、`git status --ignored --short -- .github/codex`。发现并修正 1 个 workflow 审查问题：原先变更范围保护仅在 `openspec/` 有变化时运行，若 Codex 产生非 OpenSpec 意外变更但未产生归档变化会被静默跳过；已调整为只要存在完成候选并执行过 Codex，就运行 `Guard Generated Changes`。复验通过：`openspec validate nightly-openspec-archive --strict`、Codex JSON/TOML 解析、workflow YAML 解析、`git diff --check`。i18n 影响：本轮仅新增 GitHub Actions workflow、Codex 配置模板、`.gitignore` 规则和 OpenSpec 文档，不新增或修改前端运行时语言包、宿主/插件 manifest i18n、apidoc i18n、菜单、按钮、表单、接口文档或用户可见运行时文案。缓存一致性影响：本轮不修改运行时代码、不新增缓存、不改变缓存权威数据源、失效触发、跨实例同步或故障降级策略。数据权限影响：本轮不新增或修改后端数据操作接口、服务数据访问路径、插件宿主服务适配器或聚合统计，不影响角色数据权限边界。E2E 影响：本轮不涉及用户可观察页面、路由、表单、表格或端到端业务流程，使用 OpenSpec 与 workflow 静态验证即可，不新增 E2E。
- [x] 2026-05-12: 二次安全复核后收紧 nightly 自动写回范围：当前迭代可提交 `.github/workflows/nightly-openspec-archive.yml`、`.github/codex/config.toml`、`.github/codex/auth.template.json` 和 `.gitignore` 作为人工治理变更，但 workflow 运行时的 `Guard Generated Changes` 与 `Commit Archive Changes` 仅允许自动提交 `openspec/**`，防止 Codex 在无人值守环境中修改 workflow、Codex 配置或密钥忽略规则。复验命令：`openspec validate nightly-openspec-archive --strict`、Codex JSON/TOML 解析、workflow YAML 解析、`git diff --check -- .github/workflows/nightly-openspec-archive.yml .github/codex .gitignore openspec/changes/nightly-openspec-archive`。

## Feedback

- [x] **FB-1**: `.github/codex/config.toml` 中的 provider `base_url` 不应写入真实 endpoint，必须改为通过 GitHub Secret 在 nightly workflow 运行时注入
- [x] **FB-2**: `base_url` 模板占位符应与 `auth.template.json` 的占位符风格保持一致，使用 `${OPENAI_BASE_URL}`

### Feedback Verification

- [x] 2026-05-12: FB-1 验证通过。`.github/codex/config.toml` 已将 `base_url` 改为占位符；`Nightly OpenSpec Archive` workflow 在 `Prepare Codex Runtime` 中要求 `secrets.OPENAI_BASE_URL`，并仅在临时 `CODEX_HOME/config.toml` 中用 JSON 字符串编码后的 secret 替换占位符，真实 endpoint 不进入仓库工作区。验证命令：`python3 -m json.tool .github/codex/auth.template.json >/dev/null`；`python3` 使用 `tomllib` 解析 `.github/codex/config.toml`；`ruby -e 'require "yaml"; YAML.load_file(".github/workflows/nightly-openspec-archive.yml")'`；`openspec validate nightly-openspec-archive --strict`；`rg -n 'co\\.yes\\.vg|https://co' .github/codex .github/workflows/nightly-openspec-archive.yml openspec/changes/nightly-openspec-archive`；临时目录模拟替换并用 `tomllib` 解析包含特殊字符的 `OPENAI_BASE_URL`；`git diff --check -- .github/workflows/nightly-openspec-archive.yml .github/codex .gitignore openspec/changes/nightly-openspec-archive`。i18n 影响：本反馈仅修改 GitHub Actions 配置、Codex 模板和 OpenSpec 文档，不涉及运行时用户可见文案或翻译资源。缓存一致性影响：不修改运行时代码或缓存逻辑。数据权限影响：不涉及数据操作接口或权限边界。E2E 影响：不涉及前端页面或端到端用户流程，采用治理验证。
- [x] 2026-05-12: FB-1 `lina-review` 审查完成。审查范围来源：`sed -n '1,260p' .github/workflows/nightly-openspec-archive.yml`、`sed -n '1,80p' .github/codex/config.toml`、`openspec status --change nightly-openspec-archive --json`、`git status --short -- .github/codex .github/workflows/nightly-openspec-archive.yml openspec/changes/nightly-openspec-archive .gitignore`。确认真实 `base_url` 已从仓库模板中移除，workflow 同时检查 `OPENAI_API_KEY` 和 `OPENAI_BASE_URL`，并只在 runner 临时 `CODEX_HOME` 中写入带真实 endpoint 的配置；模板占位符替换使用 `json.dumps(base_url)` 生成合法 TOML 字符串，避免特殊字符破坏配置。严重问题 0；警告 0。
- [x] 2026-05-12: FB-2 验证通过。`base_url` 模板占位符已统一为 `"${OPENAI_BASE_URL}"`，workflow 仍通过 Python 精确替换整个 TOML 字符串值，并使用 `json.dumps(base_url)` 保证特殊字符会被编码为合法 TOML 字符串；不依赖 shell 变量展开。验证命令：`python3 -m json.tool .github/codex/auth.template.json >/dev/null`；`python3` 使用 `tomllib` 解析 `.github/codex/config.toml`；`ruby -e 'require "yaml"; YAML.load_file(".github/workflows/nightly-openspec-archive.yml")'`；`openspec validate nightly-openspec-archive --strict`；临时目录模拟替换并用 `tomllib` 解析包含特殊字符的 `OPENAI_BASE_URL`；`rg -n 'co\\.yes\\.vg|https://co' .github/codex .github/workflows/nightly-openspec-archive.yml openspec/changes/nightly-openspec-archive`；`git diff --check -- .github/workflows/nightly-openspec-archive.yml .github/codex .gitignore openspec/changes/nightly-openspec-archive`。
