## 1. Workflow 基线与命名

- [x] 1.1 记录当前 `.github/workflows/nightly-test-and-build.yml` 与 `.github/workflows/release-build.yml` 的 job、权限、触发条件和 artifact 差异
- [x] 1.2 将 release workflow 调整为 `.github/workflows/release-test-and-build.yml`，并将 workflow name 改为 `Release Test and Build`
- [x] 1.3 移除或替换旧 `.github/workflows/release-build.yml`，确保仓库中不存在职责重叠的 release 镜像发布 workflow

## 2. 发布前测试门禁

- [x] 2.1 在 release workflow 中保留 Windows 命令冒烟 job
- [x] 2.2 在 release workflow 中加入 Go 单元测试 job，并优先复用 `.github/workflows/reusable-backend-unit-tests.yml`
- [x] 2.3 在 release workflow 中加入前端单元测试 job，并优先复用 `.github/workflows/reusable-frontend-unit-tests.yml`
- [x] 2.4 在 release workflow 中加入完整 E2E job，覆盖宿主和官方插件测试范围
- [x] 2.5 在 release workflow 中加入 Redis cluster smoke job，覆盖集群协调基础能力
- [x] 2.6 为 release E2E 和 Redis smoke 上传独立命名的日志与测试报告 artifact，避免与 nightly artifact 混淆

## 3. 官方插件完整验证

- [x] 3.1 为 release workflow 的 checkout 配置递归 submodule 初始化，兼容官方插件工作区迁移为 submodule 的形态
- [x] 3.2 增加官方插件工作区 preflight，检查 `apps/lina-plugins` 存在且至少包含官方插件 `plugin.yaml`
- [x] 3.3 增加插件 E2E 范围 preflight，确认 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/TC*.ts` 可被发现
- [x] 3.4 确保 preflight 失败时在镜像构建前快速失败，并输出 `git submodule update --init --recursive` 初始化提示
- [x] 3.5 确认 release 完整 E2E 使用 `pnpm test` 或等价 plugin-full 入口，而不是 host-only 入口

## 4. 镜像发布依赖与行为保持

- [x] 4.1 为 `release-image` 增加 `needs`，依赖 Windows 命令冒烟、Go 单测、前端单测、完整 E2E 和 Redis cluster smoke 全部成功
- [x] 4.2 保留 release tag Docker tag 合法性校验，确保非法 Git tag 不会进入镜像构建
- [x] 4.3 保留 `make image config=.github/image/config.yaml platforms=linux/amd64,linux/arm64 push=1` 多架构镜像构建与推送入口
- [x] 4.4 保留 release tag 和 `latest` 浮动标签发布，并在发布后执行远端 manifest inspect
- [x] 4.5 确认任一测试 job 失败、取消或超时时，GHCR login、镜像 push 和 `latest` 更新均不会执行

## 5. 验证与审查

- [x] 5.1 运行 workflow YAML 静态检查或等价解析验证，确认 release workflow 语法有效
- [x] 5.2 运行 `openspec validate release-test-and-build --strict`
- [x] 5.3 运行 `git diff --check -- .github/workflows openspec/changes/release-test-and-build`
- [x] 5.4 记录本变更不新增业务 API、数据库 schema、前端运行时文案、运行时缓存或数据权限逻辑，因此无需新增 i18n、缓存一致性或数据权限实现变更
- [x] 5.5 调用 `lina-review` 完成代码和规范审查，并修正审查发现

## 审查记录

- [x] 2026-05-13: `lina-review` 审查完成。审查范围来源：`git status --short -- .github/workflows/release-build.yml .github/workflows/release-test-and-build.yml .github/workflows/reusable-backend-unit-tests.yml .github/workflows/reusable-frontend-unit-tests.yml .github/workflows/reusable-windows-command-smoke.yml openspec/changes/release-test-and-build`、`git diff`、未跟踪文件展开、`openspec status --change release-test-and-build --json`。确认 release workflow 已替换为 `Release Test and Build`，镜像发布 job 通过 `needs` 等待 Windows 命令冒烟、Go 单测、前端单测、完整 E2E 和 Redis cluster smoke；release checkout 使用递归 submodule；E2E 与镜像发布前包含官方插件工作区 preflight；完整 E2E 使用 `pnpm test` 覆盖宿主与插件范围；测试失败、取消或超时会阻止 GHCR 登录、镜像 push 和 `latest` 更新。`actionlint`、OpenSpec 严格校验、YAML 解析和空白检查均通过。本变更只修改 GitHub Actions 和 OpenSpec 文档，不新增业务 API、数据库 schema、前端运行时文案、运行时缓存或数据权限逻辑，因此无 i18n、缓存一致性或数据权限实现变更。严重问题 0；警告 0。

## Feedback

- [x] **FB-1**: `hack/tools/image-builder/main.go` 单文件内容过大，CLI、配置归一化、平台解析、镜像构建、文件 staging 和命令执行职责集中，影响维护和审查；已拆分为同一 `package main` 下的入口编排、配置、参数、平台、环境输出、镜像、产物 staging 和命令执行源码文件，未新增可被外部 import 的包边界

## Feedback Verification

- [x] 2026-05-14: FB-1 验证通过：`cd hack/tools/image-builder && go test ./... -count=1`、`go run ./hack/tools/image-builder --preflight --tag=test-preflight`、`go run ./hack/tools/image-builder --print-build-env --tag=test-preflight`。确认 `main.go` 从 745 行降至约 100 行，最大拆分文件约 265 行，工具仍通过同一 `go run ./hack/tools/image-builder` 入口使用。i18n 影响：仅拆分开发工具源码文件，不新增或修改前端运行时文案、接口文档、manifest i18n 或 apidoc i18n 资源。缓存一致性影响：不涉及运行时缓存、失效或跨实例一致性策略。数据权限影响：不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。E2E 影响：无前端可见行为变化，使用工具单元测试、preflight 和 build env 输出验证覆盖。
- [x] 2026-05-14: FB-1 `lina-review` 审查完成。审查范围来源：`git status --short -- hack/tools/image-builder openspec/changes/release-test-and-build/tasks.md`、`git ls-files --others --exclude-standard hack/tools/image-builder`、`git diff -- hack/tools/image-builder/main.go openspec/changes/release-test-and-build/tasks.md`、`openspec status --change release-test-and-build --json`。确认拆分仅改变 `image-builder` 源码组织方式，`main.go` 保留入口和流程编排，其余职责保持同一 `package main` 下的不可 import 文件；未新增后端运行期依赖、HTTP/API、数据库 schema、前端运行时文案、缓存控制或数据权限逻辑。新增源码文件具备文件职责注释，函数/类型注释完整，未发现直接 `g.Log()`、隐式关键服务构造、未处理错误或用户可见接口错误风险。严重问题 0；警告 0。
