## 1. OpenSpec 契约

- [x] 1.1 创建动态插件 Storage 分片上传 proposal、design 和增量规范
- [x] 1.2 运行`openspec validate stream-dynamic-plugin-storage-upload --strict`

## 2. Host Service 协议与实现

- [x] 2.1 扩展 storage host service 方法目录和协议编解码，覆盖`put.init`、`put.chunk`、`put.commit`和`put.abort`
- [x] 2.2 在 WASM storage host service 中实现上传会话、chunk 顺序校验、commit 流式写入和 abort 清理
- [x] 2.3 更新动态插件 guest storage adapter，使`Storage().Put`按输入大小自动选择单次或分片上传

## 3. 测试与审查

- [x] 3.1 补充协议、WASM host service 和 guest SDK 单元测试，覆盖小文件直传、大文件分片、未知大小 reader、offset 错误和 abort 清理
- [x] 3.2 运行 Go 编译门禁、相关单元测试、`git diff --check`和 OpenSpec 严格校验
- [x] 3.3 执行`lina-review`并修复阻塞问题后记录影响分析

## 4. 影响分析与审查记录

- [x] `i18n`影响：本变更不新增运行时用户可见文案、菜单、路由、按钮、表单、API 文档源文本或语言包资源；动态样例仅新增`plugin.yaml`方法声明注释和 README 示例，未改变插件运行时翻译资源。
- [x] 缓存一致性影响：本变更不新增缓存或关键运行时快照；上传会话是一次`Storage().Put`调用内的进程内临时传输状态，使用临时文件和 TTL 清理，不提供跨实例断点续传语义。
- [x] 数据权限影响：未新增 HTTP API、数据库查询、列表、详情、下载或文件中心数据访问；动态插件 storage 路径授权仍通过`storage.resources.paths`校验最终 logical path，未绕过宿主文件可见性边界。
- [x] 开发工具跨平台影响：未修改`Makefile`、脚本、CI、代码生成或 linactl 工具；无需额外跨平台工具验证。
- [x] DI 来源检查：新增的`storageUploads`为`hostServiceRuntime`内部临时上传会话状态，在`NewRuntime`创建并随 runtime 传入 host call 上下文；没有新增外部运行期服务依赖，也没有在请求路径临时`New()`关键服务。
- [x] API 契约影响：未修改 HTTP API；仅扩展`pluginbridge`host service 协议和动态插件可声明方法目录，协议 DTO 与方法目录已通过单元测试和 catalog/descriptor 测试覆盖。
- [x] E2E 影响：无前端页面、路由或用户可观察工作流变化，未触发 E2E 新增要求；内部可执行行为由协议、WASM host service 和 guest SDK 单元测试覆盖。
- [x] 子代理协作判断：本次变更需要在已有脏工作区和同一批文件中精确合并，使用子代理会增加覆盖前序无关变更的风险，因此采用本地串行实现和审查。
- [x] 验证证据：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/protocol ./pkg/plugin/pluginbridge/protocol/hostservices ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`git diff --check`通过；`git -C apps/lina-plugins diff --check`通过；`openspec validate stream-dynamic-plugin-storage-upload --strict`通过。
