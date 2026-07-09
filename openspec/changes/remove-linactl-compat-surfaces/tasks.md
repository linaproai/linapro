## 1. 规范与文档

- [x] 1.1 更新`linactl`构建工具规范，删除旧参数、旧变量、旧回退路径和宽松输入的契约。
- [x] 1.2 更新`hack/tools/linactl/README.md`和`README.zh-CN.md`，保持中英文工具说明一致。
- [x] 1.3 同步修正活跃变更`add-go-static-lint`中关于`plugins=auto`的旧描述。

## 2. 实现

- [x] 2.1 删除旧构建变量兼容，只保留`$(TARGET_DIR)`、`$(BUILD_DIR)`和`$(REPO_ROOT)`标准变量展开。
- [x] 2.2 删除`dir=<path>`构建目标缺少`hack/config.yaml`时回退读取`package.json`的逻辑。
- [x] 2.3 删除`plugins=auto`显式参数值，未传`plugins`时仍按工作区自动判断。
- [x] 2.4 删除镜像构建的`LINAPRO_IMAGE_REGISTRY`环境变量覆盖和`print_build_env`调试兼容入口。
- [x] 2.5 删除参数 key 的连字符/下划线互通和`yes/y/on/no/n/off`布尔别名。
- [x] 2.6 删除`hack/makefiles/release.mk`对旧发布打印参数名的转发。

## 3. 验证与审查

- [x] 3.1 更新`hack/tools/linactl`单元测试，覆盖旧写法被拒绝和标准写法仍可用。
- [x] 3.2 运行`cd hack/tools/linactl && go test ./... -count=1`。
- [x] 3.3 运行`openspec validate remove-linactl-compat-surfaces --strict`和`openspec validate add-go-static-lint --strict`。
- [x] 3.4 运行`git diff --check`并完成`lina-review`审查。

## 影响判断记录

- `i18n`影响：本变更只修改开发工具契约、技术文档和 OpenSpec 文档，不新增或修改运行时用户可见文案、菜单、路由、接口文档源文本、插件清单或语言包资源，确认无运行时`i18n`资源影响。
- 缓存一致性影响：不新增运行时缓存、缓存失效、订阅状态、权限快照或分布式协调路径，确认无缓存一致性影响。
- 数据权限影响：不新增或修改列表、详情、导出、下拉、批量接口或数据可见性逻辑，确认无数据权限影响。
- 开发工具跨平台影响：`linactl`继续使用`Go`标准库完成参数解析、文件读取、路径处理和子进程编排，不引入`bash`、`sed`、`awk`或平台专属默认入口；验证以`linactl`工具单元测试和 OpenSpec 严格校验为准。
- 测试策略影响：该变更不涉及前端页面、用户交互或端到端业务流程，确认无需新增`E2E`测试；验证聚焦工具单元测试、文档/规范校验和静态差异检查。
- 运行期服务依赖影响：不新增`Controller`、`Middleware`、`Service`、插件宿主服务适配器或`WASM host service`运行期依赖，确认无需`DI`来源检查。
- 实际验证：已运行`cd hack/tools/linactl && go test ./... -count=1`、`openspec validate remove-linactl-compat-surfaces --strict`、`openspec validate add-go-static-lint --strict`、`git diff --check`、`make -n release.tag.check print-version=1`、下划线形式发布打印参数的`make -n`反向验证、`make -n wasm dir=apps/lina-plugins/linapro-demo-dynamic`、`make -n wasm p=linapro-demo-dynamic`，并完成`hack/tools`旧变量、旧环境变量、旧参数、旧布尔别名、旧调试入口和`wasm`旧插件选择参数残留检索，均通过。

## Feedback

- [x] **FB-1**: 删除`command_build.go`中对旧构建变量名的显式判断，避免工具继续认识旧变量契约。
- [x] **FB-2**: 审查`hack/tools`其它源码并删除动态`WASM`布尔别名、发布标签环境变量兜底、旧调试入口专门分支和旧交互 helper。
- [x] **FB-3**: 将`linactl`公开参数统一为`kebab-case`，删除`snake_case`公开参数形态。
- [x] **FB-4**: 将`wasm`单插件构建入口统一为`dir=<path>`，删除`wasm`命令的`p`和`plugin-dir`公开参数。
- [x] **FB-5**: 修复`GitHub Actions`中`linactl`静态扫描残留未使用函数导致的失败。
- [x] **FB-6**: 修复`GitHub Actions`中动态插件 runtime 单测仍使用旧`plugin_dir`参数导致的`WASM`构建失败。

## Feedback 修复记录

- FB-5 根因：`linactl`镜像构建参数收敛后，`defaultBuildBinaryPath`和`joinPlatformSpace`不再被调用，但源码中仍保留，触发`staticcheck U1000`。
- FB-5 修复：删除两个由本次收敛产生的未使用 helper，不改变镜像构建契约。
- FB-6 根因：动态插件 runtime 测试 helper 仍用旧参数`plugin_dir=<path>`调用`linactl wasm`；参数收敛后该参数不再被识别，命令退回到扫描官方插件工作区，host-only CI 因`apps/lina-plugins`未初始化失败，plugin-full CI 未构建测试临时插件产物。
- FB-6 修复：将测试 helper 的`linactl wasm`调用改为当前契约`dir=<path>`。
- `i18n`影响：本次反馈只修改开发工具内部死代码、测试 helper 和 OpenSpec 任务记录，不新增或修改运行时用户可见文案、菜单、路由、API 文档源文本、插件清单或语言包资源，确认无`i18n`资源影响。
- 缓存一致性影响：不新增或修改运行时缓存、失效广播、订阅状态、权限快照或分布式协调路径，确认无缓存一致性影响。
- 数据权限影响：不新增或修改业务数据读写、列表、详情、导出、批量、下拉候选或租户/组织边界逻辑，确认无数据权限影响。
- 开发工具跨平台影响：保留`linactl`基于`Go`标准库的跨平台入口，未新增`shell`、平台专属命令或路径语义；通过`make lint.go plugins=0`、`make lint.go plugins=1`和`hack/tools/linactl`全包单测验证。
- 测试策略影响：不涉及前端页面或用户可观察业务流程，确认无需新增`E2E`；复用并修复 CI 已失败的 Go 单测路径，补充`-race`定向验证。
- 插件影响：只修复动态插件 runtime artifact 测试构建入口参数，不修改插件目录结构、插件清单、插件资源或 host service 协议。
- 运行期服务依赖影响：不新增`Controller`、`Middleware`、`Service`、插件宿主服务适配器或`WASM host service`运行期依赖，确认无需`DI`来源检查。
- 已加载规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/backend-go.md`、`.agents/rules/testing.md`、`.agents/rules/plugin.md`、`.agents/rules/documentation.md`。
- 实际验证：已运行`cd hack/tools/linactl && go test ./... -count=1`、`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run 'TestBuildRuntimeWasmArtifactEmbedsBackendContracts|TestLoadRuntimePluginManifestFromArtifactHydratesBackendContracts' -count=1`、`cd apps/lina-core && go test -race ./internal/service/plugin/internal/runtime -run 'TestBuildRuntimeWasmArtifactEmbedsBackendContracts|TestLoadRuntimePluginManifestFromArtifactHydratesBackendContracts' -count=1`、`make lint.go plugins=0`、`make lint.go plugins=1`、`openspec validate remove-linactl-compat-surfaces --strict`、`openspec validate add-go-static-lint --strict`、`git diff --check`，均通过。
