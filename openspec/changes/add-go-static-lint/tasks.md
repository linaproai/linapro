## 1. 静态检查配置

- [x] 1.1 新增仓库根目录`.golangci.yml`，配置首批`linters`、`formatters`、生成代码排除和`nolint`治理策略。
- [x] 1.2 新增仓库根目录`.golangci-lint-version`，固定`golangci-lint`版本，并确保`CI`和文档引用同一版本来源。
- [x] 1.3 本地运行配置语法检查或等价`golangci-lint`空跑验证，确认配置可被当前固定版本解析。

## 2. `linactl`和`Makefile`入口

- [x] 2.1 在`hack/tools/linactl`新增`lint.go`命令注册和命令实现，支持`plugins=auto|0|1`与显式`fix=true`参数。
- [x] 2.2 复用现有官方插件工作区和临时`go.work`准备逻辑，使`plugins=0`覆盖宿主工作区，`plugins=1`覆盖官方插件完整工作区。
- [x] 2.3 在根`Makefile`拆分片段中新增`lint.go`薄转发目标，并按最终决策新增或保留聚合`lint`目标。
- [x] 2.4 为`linactl lint.go`补充单元测试，覆盖命令注册、参数解析、宿主模式分发、插件完整模式分发和插件工作区缺失错误。

## 3. `CI`集成

- [x] 3.1 新增可复用`GitHub Actions`工作流，用固定版本安装`golangci-lint`并通过仓库`make`或`linactl`入口执行检查。
- [x] 3.2 将`Go`静态检查接入`reusable-test-verification-suite.yml`，并在主`CI`和发布验证中默认启用宿主模式与插件完整模式。
- [x] 3.3 确认`CI`不以`only-new-issues`作为长期默认豁免策略，静态检查失败必须阻断主验证和发布验证。

## 4. 文档和规范

- [x] 4.1 更新`hack/tools/linactl/README.md`和`README.zh-CN.md`，说明`lint.go`命令、参数、插件模式、版本锁定、自动修复入口和验证方式。
- [x] 4.2 更新相关开发工具或后端治理规则，记录`Go`静态检查门禁、跨平台要求、`CI`执行策略和审查要求。
- [x] 4.3 记录本变更影响判断：无运行时`i18n`资源影响、无缓存一致性影响、无数据权限影响、无运行期服务依赖影响、无需`E2E`测试。

## 影响判断记录

- 跨平台影响：默认本地入口收敛到`linactl lint.go`，根`Makefile`和`make.cmd`保持薄包装；命令实现使用`Go`路径、环境变量和子进程能力，支持`Windows`、`Linux`和`macOS`。新增`GitHub Actions`安装步骤仅运行在`ubuntu-latest`可复用工作流中，不作为本地开发默认路径。
- `i18n`影响：本变更只新增开发工具命令、`CI`门禁和技术文档，不新增运行时用户可见文案、菜单、路由、接口文档源文本或翻译资源，确认无运行时`i18n`资源影响。
- 缓存一致性影响：不新增运行时缓存、缓存失效、订阅状态、权限快照或分布式协调路径，确认无缓存一致性影响。
- 数据权限影响：不新增或修改列表、详情、导出、下拉、批量接口或数据可见性逻辑，确认无数据权限影响。
- 运行期服务依赖影响：不新增`Controller`、`Middleware`、`Service`、插件宿主服务适配器或`WASM host service`运行期依赖，确认无需`DI`来源检查。
- 测试策略影响：本变更属于开发工具与`CI`治理，不改变前端页面、用户交互或端到端业务流程，确认无需新增`E2E`测试；验证聚焦`linactl`单元测试、`golangci-lint`配置检查、宿主/插件模式静态扫描和`OpenSpec`严格校验。

## 5. 验证和审查

- [x] 5.1 运行`go test`覆盖`hack/tools/linactl`新增或修改的命令实现和测试。
- [x] 5.2 运行`make lint.go plugins=0`并修复宿主模式静态检查问题。
- [x] 5.3 在官方插件工作区可用时运行`make lint.go plugins=1`并修复插件完整模式静态检查问题；如环境不可用，记录阻断原因和替代验证。
- [x] 5.4 运行`openspec validate add-go-static-lint --strict`。
- [x] 5.5 完成实现后调用`lina-review`进行代码和规范审查，审查通过后再标记全部任务完成。
