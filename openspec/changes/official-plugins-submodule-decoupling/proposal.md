## Why

官方源码插件数量持续增加，继续把完整插件代码直接维护在主仓库会提高主仓库体积、上下文复杂度和版本协同成本。将官方插件独立为仓库并通过 `git submodule` 挂载回 `apps/lina-plugins` 前，需要先证明宿主服务、前端工作台、开发工具和测试套件在源码插件工作区缺失或为空时仍能独立运行。

## What Changes

- 引入“官方插件工作区可选化”治理：`apps/lina-plugins` 可不存在、为空目录或作为 submodule 存在。
- 增加 host-only 验证路径：临时移除或置空 `apps/lina-plugins` 后，宿主后端单元测试、宿主前端构建/类型检查和宿主 E2E 必须可执行。
- 修正因 `go.work`、宿主编译期 blank import、插件扫描、前端插件页面发现、Playwright 测试发现或开发工具硬编码路径引发的强耦合失败。
- 增加 submodule 恢复验证路径：官方插件作为 submodule 挂载回 `apps/lina-plugins` 后，所有官方插件后端单元测试和插件自有 E2E 必须可执行。
- 更新文档与开发入口，明确 clone/submodule 初始化、host-only 验证、插件完整验证三类工作流。

## Capabilities

### New Capabilities
- `official-plugin-workspace-decoupling`: 定义官方源码插件工作区作为可选 submodule 的宿主解耦、验证和回归要求。

### Modified Capabilities
- `project-setup`: 宿主后端与前端基础命令在官方源码插件工作区缺失或为空时仍应可执行。
- `plugin-manifest-lifecycle`: 源码插件清单扫描必须容忍插件工作区缺失或为空，并将显式同步与运行时发现行为区分清楚。
- `plugin-ui-integration`: 前端构建和运行时插件页面注册必须容忍源码插件前端目录缺失或为空。
- `e2e-suite-organization`: E2E 测试套件必须支持宿主测试与插件测试分离执行，插件工作区缺失时宿主测试仍可发现和执行。
- `release-image-build`: 标准构建入口必须区分 host-only 构建与包含官方插件 submodule 的完整构建。

## Impact

- 影响根目录 `go.work`、Go 构建/测试入口、宿主 `main.go` 或插件注册方式。
- 影响 `apps/lina-core/internal/service/plugin/internal/catalog` 的源码插件发现、同步和 fallback 行为。
- 影响 `apps/lina-vben/apps/web-antd` 的 Vite 插件页面扫描、路由注册和访问过滤逻辑。
- 影响 `hack/tests` 的 Playwright 配置、执行治理脚本、测试清单和插件测试范围解析。
- 影响 `hack/tools/linactl`、`make build`、`make test`、`make wasm` 等开发/发布工具入口。
- 影响 README/CONTRIBUTING/AGENTS 等文档中关于官方源码插件目录、submodule 初始化和测试命令的说明。
