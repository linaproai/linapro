## 1. 实现

- [x] 1.1 调整普通文件本地存储路径生成，去掉新上传路径中的 `t` 目录层。
- [x] 1.2 同步更新上传路径注释和 API 示例，避免继续展示旧路径格式。

## 2. 测试

- [x] 2.1 补充或更新文件服务单元测试，覆盖新上传路径格式。
- [x] 2.2 补充或确认旧 `t/...` 路径仍按数据库记录访问。

## 3. 验证与审查

- [x] 3.1 运行 `openspec validate simplify-upload-storage-path --strict`。
- [x] 3.2 运行覆盖变更包的 `cd apps/lina-core && go test ./internal/service/file -count=1`。
- [x] 3.3 运行后端启动绑定编译烟测 `cd apps/lina-core && go test ./internal/cmd -count=1`。
- [x] 3.4 记录 i18n、缓存一致性、数据权限、开发工具跨平台影响评估，并执行 `lina-review` 审查。

## 4. 完成记录

- 实现：新上传文件的本地相对存储路径由 `t/<tenantId>/<yyyy>/<MM>/<filename>` 调整为 `<tenantId>/<yyyy>/<MM>/<filename>`；历史 `sys_file.path` 记录不迁移，读取时继续按数据库记录路径访问。
- 测试：新增 `TestLocalStoragePutUsesTenantIDWithoutTenantPrefix` 覆盖新路径格式；新增 `TestOpenByPathPreservesLegacyTenantPrefixPath` 覆盖旧 `t/...` 路径兼容读取。
- i18n：未新增、修改或删除用户可见运行时文案、前端语言包、插件 manifest/i18n 或 apidoc i18n 资源；仅同步 Go API DTO 的英文示例值。
- 缓存一致性：不新增或修改缓存；文件访问仍以 `sys_file.path` 数据库记录为权威路径，分布式环境下没有新增缓存一致性风险。
- 数据权限：不新增或修改数据操作接口；上传记录仍写入当前租户，读取仍依赖现有元数据查询、租户过滤和数据权限校验。
- 开发工具跨平台：不新增或修改开发工具、脚本、CI 入口或默认开发命令。
- 验证：`openspec validate simplify-upload-storage-path --strict` 通过；`cd apps/lina-core && go test ./internal/service/file -count=1` 通过；`cd apps/lina-core && go test ./api/file/v1 ./api/user/v1 -count=1` 通过；`cd apps/lina-core && go test ./internal/cmd -count=1` 通过。
- 审查：已按 `lina-review` 范围检查本变更的 OpenSpec 记录、Go 后端路径生成、API 示例、旧路径兼容测试、Go 编译门禁、i18n、缓存一致性、数据权限和开发工具跨平台影响，未发现阻断问题。
