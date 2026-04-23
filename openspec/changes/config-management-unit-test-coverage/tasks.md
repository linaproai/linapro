## 1. 测试基线与夹具整理

- [x] 1.1 盘点 `apps/lina-core/internal/service/config` 当前覆盖率明细，确认本轮优先补测的低覆盖文件与分支。
- [x] 1.2 整理 `config` 包测试夹具，补充静态缓存、runtime snapshot、revision 状态与 plugin path override 的成对重置辅助方法。
- [x] 1.3 如现有实现对测试隔离不友好，完成最小范围的测试友好性整理，确保不改变生产语义。

## 2. 补齐低覆盖配置子模块单元测试

- [x] 2.1 为 `config_plugin.go` 补充单元测试，覆盖默认目录、`runtime.storagePath` 兼容回退、override 生效与清理逻辑。
- [x] 2.2 为 `config_public_frontend.go` 补充单元测试，覆盖 `PublicFrontendSettingSpecs` 拷贝语义、`IsProtectedConfigParam` 判定、`ValidateProtectedConfigValue` 分发与时区解析分支。
- [x] 2.3 为 `config_runtime_params_revision.go` 补充单元测试，覆盖 clustered revision 的读取、同步、递增与共享 KV 错误传播路径。
- [x] 2.4 为 `config_runtime_params_cache.go` 补充单元测试，覆盖缓存命中、revision 变化后的重建、无效缓存值移除、异常回退与本地 TTL 刷新路径。
- [x] 2.5 为 `config_jwt.go`、`config_session.go`、`config_upload.go`、`config_login.go`、`config_metadata.go` 等剩余低覆盖 getter/helper 补充默认值、空对象与异常分支测试。

## 3. 覆盖率验证与结果收口

- [x] 3.1 执行 `cd apps/lina-core && go test ./internal/service/config -cover`，确认包级覆盖率达到 `80%` 及以上。
- [x] 3.2 如首次未达标，继续补足缺口分支并重复验证，直到覆盖率门槛达成。
- [x] 3.3 在变更记录中补充最终覆盖率结果与本轮新增测试范围，作为评审输入。

## 4. 当前结果

- `2026-04-23`：新增 `config_plugin_test.go`、`config_protected_settings_test.go`、`config_runtime_params_revision_additional_test.go`，补齐插件配置、受保护配置辅助逻辑、cluster revision 与 runtime snapshot cache 的关键分支测试。
- `2026-04-23`：本轮直接复用并规范使用现有 `setTestConfigContent`、`resetRuntimeParamCacheTestState`、`SetPluginDynamicStoragePathOverride` 等测试夹具完成隔离控制，未引入额外生产代码重构。
- `2026-04-23`：执行 `cd apps/lina-core && go test ./internal/service/config -cover`，结果为 `coverage: 83.0% of statements`，已达到本次变更要求的 `80%+` 门槛。
