## 1. Workflow 实现

- [x] 1.1 新增仅支持`workflow_dispatch`的手动`nightly`镜像发布 workflow
- [x] 1.2 让手动 workflow 直接调用`.github/workflows/reusable-image-publish.yml`，并确认不存在测试 job 或`needs`测试依赖
- [x] 1.3 配置与现有 nightly 一致的镜像名称、平台、`tag-mode`、浮动标签和官方插件完整构建模式

## 2. 验证与审查

- [x] 2.1 运行 workflow YAML 解析和`actionlint`静态检查
- [x] 2.2 运行`openspec validate manual-nightly-image-build --strict`
- [x] 2.3 运行`git diff --check -- .github/workflows openspec/changes/manual-nightly-image-build`
- [x] 2.4 记录本变更不新增后端生产代码、业务 API、数据库 schema、前端运行时文案、运行时缓存、数据权限逻辑或默认开发工具脚本，因此无需新增 Go 编译门禁、i18n、缓存一致性或数据权限实现变更
- [x] 2.5 调用`lina-review`完成代码和规范审查

## 验证记录

- [x] 2026-05-17: 验证通过：`ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f) }' .github/workflows/manual-nightly-image-build.yml`、`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/manual-nightly-image-build.yml`、`openspec validate manual-nightly-image-build --strict`、`git diff --check -- .github/workflows openspec/changes/manual-nightly-image-build`，并对新增未跟踪文件补充执行`git diff --no-index --check`空白检查。确认新增 workflow 仅包含`workflow_dispatch`触发和直接调用`.github/workflows/reusable-image-publish.yml`的镜像发布 job，不包含测试 job 或测试`needs`依赖。i18n 影响：不新增或修改前端运行时文案、manifest i18n 或 apidoc i18n 资源。缓存一致性影响：不修改运行时缓存、缓存失效或跨实例一致性策略。数据权限影响：不新增或修改业务 API、数据库查询或角色数据权限边界。开发工具与脚本影响：仅新增`GitHub Actions`编排文件，不新增仓库默认开发工具、跨平台脚本或长期维护脚本入口。

## 审查记录

- [x] 2026-05-17: `lina-review`审查完成。审查范围来源：`git status --short`、`git ls-files --others --exclude-standard`、`openspec status --change manual-nightly-image-build --json`、新增 workflow 与 OpenSpec 文件内容检查。确认`.github/workflows/manual-nightly-image-build.yml`仅支持`workflow_dispatch`触发，只有一个直接调用`.github/workflows/reusable-image-publish.yml`的`manual-nightly-image` job，未声明`needs`、测试 job、共享测试验证套件、单元测试、`E2E`或 smoke 测试依赖；镜像名称、平台、`tag-mode: nightly`、`floating-tag: nightly`和`plugins: "1"`与现有 nightly 发布语义一致。OpenSpec 提案、设计、增量规范和任务记录均使用中文，变更范围与`release-image-build`能力匹配。本次不新增或修改 Go 生产代码、业务 API、数据库 schema、前端运行时文案、manifest i18n、apidoc i18n、运行时缓存、数据权限逻辑或默认开发工具脚本，因此无需 Go 编译门禁、功能 E2E、i18n 资源、缓存一致性实现或数据权限测试。严重问题 0；警告 0。
