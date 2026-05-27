## 1. 修复任务

- [x] 1.1 记录`make dev`端口释放延迟反馈，并更新`project-setup`增量规范。
- [x] 1.2 调整`linactl dev`停止和端口检查编排，等待已管理服务释放端口且不强杀未知占用者。
- [x] 1.3 增加或更新`linactl`单元测试，覆盖已管理服务端口延迟释放和未知端口占用拒绝。
- [x] 1.4 运行`linactl`包测试、OpenSpec 严格校验和空白检查，并执行`lina-review`审查。

## 2. 执行记录

- i18n 影响评估：本变更只修改开发工具重启行为和 OpenSpec 文档，不新增、修改或删除运行时前端语言包、宿主`manifest/i18n`、`apidoc i18n`、菜单、路由或插件清单资源。
- 缓存一致性影响评估：本变更不涉及运行时缓存、分布式缓存、跨实例失效、共享修订号或业务状态缓存。
- 数据权限影响评估：本变更不新增或修改数据操作接口、查询过滤、详情读取、写操作、下载、聚合统计或插件数据访问路径。
- 开发工具跨平台影响评估：本变更修改`hack/tools/linactl`，必须保持跨平台，不新增平台专属命令或脚本；验证使用`cd hack/tools/linactl && go test ./... -count=1`。
- 影响分析：修改`linactl dev`和`linactl stop`的内部停止编排、`devservice`端口检测辅助函数和`linactl`单元测试；用户可见行为仅限`make dev`在停止已管理服务后短暂等待端口释放。未知外部端口占用仍返回原端口占用错误，不自动终止未知进程。
- 验证：已运行`go test ./hack/tools/linactl -run 'TestRunDev(RejectsOccupiedPort|WaitsForManagedServicePortsToRelease|OnlyWaitsForStoppedManagedPorts|StartsServicesAsAsyncProcessesAndPrintsFinalStatus)' -count=1`，通过。
- 验证：已运行`cd hack/tools/linactl && go test ./... -count=1`，通过。
- 验证：已运行`openspec validate fix-linactl-dev-port-release --strict`，通过。
- 验证：已运行`git diff --check -- hack/tools/linactl/constants_types.go hack/tools/linactl/command_stop.go hack/tools/linactl/command_dev.go hack/tools/linactl/internal/devservice/devservice.go hack/tools/linactl/main_test.go openspec/changes/fix-linactl-dev-port-release`，通过。
- Review：已按`lina-review`口径完成审查。审查范围限定为本变更新增 OpenSpec 文档、`linactl`命令编排、`devservice`端口释放等待和相关测试；未发现阻塞问题。该变更不涉及生产后端 API、Controller、路由绑定、数据库访问、运行时缓存、数据权限、前端 UI 或 i18n 资源；Go 编译门禁由`linactl`完整包测试覆盖。
