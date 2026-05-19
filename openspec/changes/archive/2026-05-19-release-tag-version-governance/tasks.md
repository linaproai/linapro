## 1. OpenSpec 与设计

- [x] 1.1 创建 `release-tag-version-governance` 变更提案、设计和 `release-image-build` 增量规范
- [x] 1.2 运行 OpenSpec 严格校验，确认变更文档可归档

## 2. 跨平台校验工具

- [x] 2.1 在 `hack/tools/linactl` 中新增 `release.tag.check` 命令，读取 `metadata.yaml` 并校验 tag 与 `framework.version`
- [x] 2.2 为 tag 相等、tag 不匹配、版本格式非法、缺少版本字段和环境变量 fallback 增加单元测试
- [x] 2.3 在 README 和命令帮助中补充本地校验用法

## 3. GitHub 发布链路

- [x] 3.1 在 `Release Test and Build` workflow 中新增最早执行的 release tag 版本一致性 job
- [x] 3.2 将所有 release 测试和镜像发布 job 依赖版本一致性 job，确保校验失败时不会执行发布制品步骤
- [x] 3.3 新增受控 `Create Release Tag` 手动 workflow，读取 `framework.version` 并在校验通过后创建同名 tag
- [x] 3.4 在发布文档中说明 GitHub tag ruleset 配置建议和 `GITHUB_TOKEN` 二次触发限制

## 4. 验证与审查

- [x] 4.1 运行 `cd hack/tools/linactl && go test ./... -count=1`
- [x] 4.2 运行 `go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/*.yml`
- [x] 4.3 运行 YAML 解析检查和 `git diff --check`
- [x] 4.4 运行 `openspec validate release-tag-version-governance --strict`
- [x] 4.5 记录 i18n、缓存一致性、数据权限和开发工具跨平台影响判断，并执行 `lina-review` 审查

## 审查记录

- [x] 2026-05-14: 验证通过：`cd hack/tools/linactl && go test ./... -count=1`、`make release.tag.check tag=v0.2.0`、`make release.tag.check print-version=1`、`ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }' .github/workflows/*.yml .github/actions/*/action.yml`、`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/*.yml`、`openspec validate release-tag-version-governance --strict`、`git diff --check -- .github/workflows/create-release-tag.yml .github/workflows/release-test-and-build.yml hack/tools/linactl/command.go hack/tools/linactl/command_release.go hack/tools/linactl/main_test.go hack/tools/linactl/README.md hack/tools/linactl/README.zh-CN.md hack/makefiles/release.mk Makefile README.md README.zh-CN.md AGENTS.md openspec/changes/release-tag-version-governance`。
- [x] 2026-05-14: `lina-review` 审查完成。审查范围来源：`git status --short -- .github/workflows/create-release-tag.yml .github/workflows/release-test-and-build.yml Makefile README.md README.zh-CN.md AGENTS.md hack/makefiles/release.mk hack/tools/linactl openspec/changes/release-tag-version-governance`、`git ls-files --others --exclude-standard -- .github/workflows/create-release-tag.yml hack/makefiles/release.mk hack/tools/linactl/command_release.go openspec/changes/release-tag-version-governance`、`git diff` 和 `openspec status --change release-tag-version-governance --json`。确认 `Release Test and Build` 已在所有 release 验证与镜像发布前执行 `release-tag-version-check`；所有测试和镜像发布 job 均依赖该前置 job；`Create Release Tag` 手动 workflow 通过 `linactl release.tag.check` 读取并校验 `framework.version`，拒绝移动既有 tag，并文档化 `RELEASE_TAG_TOKEN` 与 tag ruleset 要求；`linactl release.tag.check` 使用 Go/YAML 跨平台实现，支持 `tag`、`GITHUB_REF_NAME` 和 `print-version`，测试覆盖匹配、环境变量 fallback、不匹配、非法版本和缺少版本字段。当前工作区存在其他活跃变更和既有未跟踪文件，本次审查仅覆盖 release tag 治理相关文件和行，未回退或重写其他改动。i18n 影响：仅新增开发工具输出、CI 文档和发布说明，不新增前端运行时文案、接口文档、manifest i18n 或 apidoc i18n 资源。缓存一致性影响：不涉及运行时缓存、失效或跨实例一致性策略。数据权限影响：不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。开发工具与脚本影响：新增默认入口实现于 `hack/tools/linactl` Go 工具，`Makefile` 仅作为薄包装；新增 GitHub Actions `run` 脚本仅在 Linux runner 内部使用，并有 `linactl` 跨平台入口覆盖本地校验。严重问题 0；警告 0。

## Feedback

- [x] **FB-1**: GitHub tag ruleset bypass 说明需要明确为 actor 级别，不应表述为可配置特定 token 绕过规则
- [x] **FB-2**: 受控打标 workflow 需要使用已配置的 GitHub App 凭证生成 installation token，而不是继续依赖 `RELEASE_TAG_TOKEN`
- [x] **FB-3**: README 中发布标签说明不应承载维护者级 tag ruleset 与受控打标细节，应迁移到贡献文档
- [x] **FB-4**: 根 README 不应保留独立发布标签章节，发布标签说明应统一由贡献文档承载

## Feedback Review

- [x] 2026-05-14: FB-1 验证通过：`ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }' .github/workflows/*.yml .github/actions/*/action.yml`、`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/*.yml`、`openspec validate release-tag-version-governance --strict`、`git diff --check -- .github/workflows/create-release-tag.yml README.md README.zh-CN.md AGENTS.md openspec/changes/release-tag-version-governance`。已将 README、AGENTS、workflow 输出、OpenSpec 设计与增量规范中的表述统一修正为 ruleset bypass 配置的是 token 对应的 actor：GitHub App token 对应 GitHub App，PAT 对应用户或机器用户账号；不再表述为可在 bypass list 中配置某个 token 字符串。i18n 影响：仅修正文档和 CI 日志说明，不涉及前端运行时文案、manifest i18n 或 apidoc i18n。缓存一致性影响：无。数据权限影响：无。开发工具与脚本影响：仅调整 GitHub Actions 提示文本，未改变执行逻辑。`lina-review` 审查完成，严重问题 0；警告 0。
- [x] 2026-05-14: FB-2 验证通过：`ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }' .github/*.yaml .github/workflows/*.yml .github/actions/*/action.yml`、`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/*.yml`、`openspec validate release-tag-version-governance --strict`、`git diff --check -- .github/actionlint.yaml .github/workflows/create-release-tag.yml README.md README.zh-CN.md AGENTS.md openspec/changes/release-tag-version-governance`。已将 `Create Release Tag` workflow 改为通过 `actions/create-github-app-token@v3` 使用仓库变量 `RELEASE_APP_CLIENT_ID` 和仓库密钥 `RELEASE_APP_PRIVATE_KEY` 生成 GitHub App installation token；checkout、远端 tag 查询和 tag push 均使用该 token，且 checkout 不持久化凭证。`actionlint` `v1.7.12` 的内置 popular actions 元数据仍把 `actions/create-github-app-token@v3` 的旧 `app-id` 输入视为必填；官方 `v3` `action.yml` 与 README 已支持 `client-id` 并将 `app-id` 标记为弃用，因此新增 `.github/actionlint.yaml` 仅针对 `create-release-tag.yml` 忽略该 action 的两个过期元数据误报。i18n 影响：仅修改发布治理文档和 CI 日志，不涉及前端运行时文案、manifest i18n 或 apidoc i18n。缓存一致性影响：无。数据权限影响：无。开发工具与脚本影响：新增 `actionlint` 配置用于静态检查误报治理，不改变本地跨平台工具入口；GitHub Actions `run` 脚本仍限定在 Linux runner 内部执行。
- [x] 2026-05-14: FB-3 验证通过：`openspec validate release-tag-version-governance --strict`、`git diff --check -- README.md README.zh-CN.md CONTRIBUTING.md openspec/changes/release-tag-version-governance/tasks.md`、`rg -n "RELEASE_APP|GITHUB_TOKEN|Ruleset bypass|tag ruleset|Tag ruleset|发布标签|Release Tags|contributing guide|贡献指南|CONTRIBUTING" README.md README.zh-CN.md CONTRIBUTING.md openspec/changes/release-tag-version-governance/tasks.md`。已将维护者级 `GitHub tag ruleset`、`GitHub App installation token`、`RELEASE_APP_CLIENT_ID`、`RELEASE_APP_PRIVATE_KEY` 和 `GITHUB_TOKEN` 触发限制说明迁移到 `CONTRIBUTING.md` 的 `Release Tags` 章节。i18n 影响：仅调整仓库文档，不涉及前端运行时文案、manifest i18n 或 apidoc i18n。缓存一致性影响：无。数据权限影响：无。开发工具与脚本影响：无新增或修改默认开发工具入口。`lina-review` 审查完成，严重问题 0；警告 0。
- [x] 2026-05-14: FB-2/FB-3 复核完成：`ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }' .github/*.yaml .github/workflows/*.yml .github/actions/*/action.yml`、`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/*.yml`、`openspec validate release-tag-version-governance --strict`、`git diff --check -- .github/actionlint.yaml .github/workflows/create-release-tag.yml README.md README.zh-CN.md CONTRIBUTING.md AGENTS.md openspec/changes/release-tag-version-governance` 均通过。`lina-review` 审查范围来自 `git status --short`、`git ls-files --others --exclude-standard`、`git diff` 和 `openspec status --change release-tag-version-governance --json`；确认受控打标 workflow 已以 GitHub App installation token 作为唯一写入凭证，tag 存在性检查不会把鉴权或网络失败误判为不存在，文档已将维护者级配置集中到贡献指南。严重问题 0；警告 0。
- [x] 2026-05-14: FB-4 验证通过：`openspec validate release-tag-version-governance --strict`、`git diff --check -- README.md README.zh-CN.md CONTRIBUTING.md openspec/changes/release-tag-version-governance/tasks.md`、`rg -n "发布标签|Release Tags|release.tag.check|Create Release Tag|tag ruleset|RELEASE_APP|GITHUB_TOKEN" README.md README.zh-CN.md CONTRIBUTING.md openspec/changes/release-tag-version-governance/tasks.md`。已从根 `README.md` 和 `README.zh-CN.md` 删除独立发布标签章节，发布标签说明统一保留在 `CONTRIBUTING.md` 的 `Release Tags` 章节。i18n 影响：仅调整仓库文档，不涉及前端运行时文案、manifest i18n 或 apidoc i18n。缓存一致性影响：无。数据权限影响：无。开发工具与脚本影响：无新增或修改默认开发工具入口。`lina-review` 审查完成，严重问题 0；警告 0。
