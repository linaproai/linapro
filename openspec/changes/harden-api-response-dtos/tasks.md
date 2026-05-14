## 1. API 响应 DTO 加固

- [x] 1.1 审查宿主 API 定义中直接嵌入或返回 `entity.*` 的响应类型，确认受影响模块和风险字段
- [x] 1.2 为用户、文件、系统配置、字典、定时任务、任务日志和任务分组响应定义独立 DTO，只暴露必要字段；DTO 已按 API 主源码文件规则放入 `api/*/v1/<api-name>.go`
- [x] 1.3 调整对应控制器响应映射，禁止直接把实体指针塞入 API 响应
- [x] 1.4 增加自动化测试或静态验证，覆盖 API 层不再依赖 `internal/model/entity` 且用户响应不暴露 `password`

## 2. 验证与审查

- [x] 2.1 运行后端相关 Go 测试，确认 DTO 映射和响应序列化行为正确
- [x] 2.2 运行静态扫描，确认 `apps/lina-core/api` 不再引用 `internal/model/entity`
- [x] 2.3 运行 `openspec validate harden-api-response-dtos --strict`
- [x] 2.4 记录 i18n、缓存一致性、数据权限和开发工具影响评估
- [x] 2.5 调用 `lina-review` 完成代码和规范审查，并修正审查发现

## Feedback

- [x] **FB-1**: 部分 API 响应直接嵌入数据库实体，响应合同被数据库结构驱动，并可能暴露密码、软删除字段、文件存储路径、哈希、租户内部字段等实现细节；应改为独立响应 DTO 并只映射允许暴露的字段
- [x] **FB-2**: 源码插件 API DTO 仍存在 `*Entity` 命名、`*_entity.go` 存放方式、软删除字段暴露和操作日志列表返回完整请求/响应载荷等合同治理缺口；应统一改为插件 API 主源码文件中的独立响应 DTO，并只暴露必要字段

## Impact Assessment

- i18n：已删除文件详情页 `engine`、`hash` 运行时翻译键，并同步清理用户/文件 apidoc i18n 中响应侧不再暴露的 `password`、`path`、`engine`、`hash` 字段翻译；请求侧密码、文件访问路径翻译保留。源码插件侧同步清理 `deletedAt` 和旧 `*Entity` 响应 schema 的 apidoc 翻译键，并将多租户插件 apidoc 键同步改为 `*Item` DTO 名；操作日志详情仍保留 `operParam/jsonResult` 文档，列表响应不再暴露这两个大字段。
- 缓存一致性：不新增或修改缓存、缓存键、缓存失效、跨实例广播或分布式协调逻辑。
- 数据权限：不改变查询、详情、下载、写操作或数据权限过滤逻辑；本次只调整 API 合同与响应边界字段映射。
- 开发工具与脚本：不新增或修改开发工具、构建脚本、测试脚本或跨平台入口。

## Verification

- `cd apps/lina-core && go test ./api/... ./internal/controller/...`
- `cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`
- `cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`
- `rg -n "internal/model/entity|entity\\." apps/lina-core/api -g '*.go'`
- `rg -n "internal/model/entity|entity\\." apps/lina-plugins/*/backend/api apps/lina-core/api -g '*.go'`
- `cd apps/lina-plugins && GOWORK=/Users/john/Workspace/github/linaproai/linapro/temp/go.work.plugins GOFLAGS=-mod=readonly go test -run 'TestPluginAPI|TestPluginResponseDTOs|TestPluginOperLog' -count=1 .`
- `cd apps/lina-plugins/content-notice && GOWORK=/Users/john/Workspace/github/linaproai/linapro/temp/go.work.plugins GOFLAGS=-mod=readonly go test ./backend/api/... ./backend/internal/controller/...`
- `cd apps/lina-plugins/org-center && GOWORK=/Users/john/Workspace/github/linaproai/linapro/temp/go.work.plugins GOFLAGS=-mod=readonly go test ./backend/api/... ./backend/internal/controller/...`
- `cd apps/lina-plugins/monitor-loginlog && GOWORK=/Users/john/Workspace/github/linaproai/linapro/temp/go.work.plugins GOFLAGS=-mod=readonly go test ./backend/api/... ./backend/internal/controller/...`
- `cd apps/lina-plugins/monitor-operlog && GOWORK=/Users/john/Workspace/github/linaproai/linapro/temp/go.work.plugins GOFLAGS=-mod=readonly go test ./backend/api/... ./backend/internal/controller/...`
- `cd apps/lina-plugins/multi-tenant && GOWORK=/Users/john/Workspace/github/linaproai/linapro/temp/go.work.plugins GOFLAGS=-mod=readonly go test ./backend/api/... ./backend/internal/controller/...`
- `rg -n "deletedAt|NoticeEntity|DeptEntity|PostEntity|LoginLogEntity|OperLogEntity|TenantEntity|LoginTenantEntity|TenantPluginEntity|internal/model/entity|type\\s+\\w*Entity\\b" apps/lina-plugins/*/backend/api apps/lina-plugins/*/manifest/i18n -g '*.go' -g '*.json'`
- `openspec validate harden-api-response-dtos --strict`
- `git diff --check`
