# 后端中文硬编码审查清单

## 扫描基线

- 扫描时间：2026-04-28
- 扫描范围：`apps/lina-core`、`apps/lina-plugins`
- 目标：识别当前后端源码中仍存在的中文硬编码，并按调用端可见风险、用户可见投影、交付物文案、开发者诊断、生成文件、测试 fixture 分类治理。

| 范围 | 命中行数 | 命中文件数 | 处理策略 |
| ---- | -------- | ---------- | -------- |
| 全部 Go 源码 | 2282 | 229 | 仅作为总体规模参考，不直接作为阻断口径 |
| 手写、非测试、非生成 Go 字符串字面量 | 780 | 71 | 本变更主要治理对象 |
| 生成文件 `dao/do/entity` | 1218 | 111 | 不手工修改生成物；需要治理时改 SQL 注释或 codegen 输入源后重新生成 |
| 测试文件 `_test.go` | 280 | 47 | 允许保留样例和断言数据；后续 allowlist 说明用途 |
| Entity `description` 中文 | 406 | 37 | 评估是否进入 OpenAPI schema 或用户文档，必要时修改生成源 |
| API DTO `summary/dc/description/eg` 中文 | 0 | 0 | 当前合规；继续保持 API DTO 英文源文本规则 |
| 高风险错误构造中文 | 710 | 60 | 优先改为 `bizerr` 或英文开发者诊断 |

## 关键扫描命令

```bash
rg -n --glob '*.go' '[\p{Han}]' apps/lina-core apps/lina-plugins | wc -l
rg -l --glob '*.go' '[\p{Han}]' apps/lina-core apps/lina-plugins | wc -l
```

```bash
rg -n -P --glob '*.go' --glob '!**/*_test.go' --glob '!**/internal/dao/**' --glob '!**/internal/model/do/**' --glob '!**/internal/model/entity/**' '"(?:[^"\\]|\\.)*\p{Han}(?:[^"\\]|\\.)*"|`[^`]*\p{Han}[^`]*`' apps/lina-core apps/lina-plugins | wc -l
rg -l -P --glob '*.go' --glob '!**/*_test.go' --glob '!**/internal/dao/**' --glob '!**/internal/model/do/**' --glob '!**/internal/model/entity/**' '"(?:[^"\\]|\\.)*\p{Han}(?:[^"\\]|\\.)*"|`[^`]*\p{Han}[^`]*`' apps/lina-core apps/lina-plugins | wc -l
```

```bash
rg -n -P --glob '*.go' 'summary:"[^"]*\p{Han}[^"]*"|dc:"[^"]*\p{Han}[^"]*"|description:"[^"]*\p{Han}[^"]*"|eg:"[^"]*\p{Han}[^"]*"' apps/lina-core/api apps/lina-plugins/*/backend/api | wc -l
```

```bash
rg -n -P --glob '*.go' --glob '!**/*_test.go' --glob '!**/internal/dao/**' --glob '!**/internal/model/do/**' --glob '!**/internal/model/entity/**' 'gerror\.(?:New|Newf|Wrap|Wrapf|NewCode|NewCodef|WrapCode|WrapCodef)\([^\n]*\p{Han}|(?:errors\.New|fmt\.Errorf)\([^\n]*\p{Han}' apps/lina-core apps/lina-plugins | wc -l
```

## 命中集中目录

| 目录 | 命中行数 | 分类判断 |
| ---- | -------- | -------- |
| `apps/lina-core/pkg/pluginbridge` | 340 | 插件 bridge/host service 开发者诊断，优先改英文稳定文案 |
| `apps/lina-core/internal/service/plugin/internal/catalog` | 122 | manifest/spec/authorization 校验诊断，用户边界需要结构化包装 |
| `apps/lina-core/internal/service/plugin/internal/runtime` | 52 | 上传、归档、路由、定时任务诊断，按暴露边界区分治理 |
| `apps/lina-core/internal/service/plugin/internal/frontend` | 31 | 前端资源包和菜单装配诊断，默认英文开发者诊断 |
| `apps/lina-core/internal/service/plugin/internal/datahost` | 26 | plugin data service 诊断，管理 API 边界需要包装 |
| `apps/lina-plugins/plugin-demo-source/backend/internal/service/demo` | 25 | 源码插件调用端可见错误和示例摘要，需插件运行时 i18n |
| `apps/lina-core/internal/service/plugin/internal/wasm` | 24 | WASM 初始化和调用诊断，默认英文开发者诊断 |
| `apps/lina-plugins/monitor-operlog/backend/internal/service/operlog` | 22 | 操作日志错误、状态 fallback、导出表头，需结构化错误和 i18n |
| `apps/lina-core/internal/service/plugin/internal/integration` | 20 | 插件集成诊断，默认英文开发者诊断 |
| `apps/lina-core/internal/service/config` | 19 | 启动期配置诊断和用户可见 disabled reason，按边界治理 |

## 高优先级文件

| 文件 | 命中行数 | 优先处理原因 |
| ---- | -------- | ------------ |
| `apps/lina-core/pkg/pluginbridge/pluginbridge_codec.go` | 77 | bridge codec 错误集中，插件调用链可能透出 |
| `apps/lina-core/internal/service/plugin/internal/catalog/spec.go` | 43 | 插件 spec 校验错误，安装/发布流程可见 |
| `apps/lina-core/internal/service/plugin/internal/catalog/manifest_validate.go` | 41 | manifest 校验错误，插件管理流程可见 |
| `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_storage_codec.go` | 41 | host service storage codec 诊断集中 |
| `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice.go` | 40 | host service 编排错误集中 |
| `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_cache_codec.go` | 40 | host service cache codec 诊断集中 |
| `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_data_codec.go` | 39 | host service data codec 诊断集中 |
| `apps/lina-core/pkg/pluginbridge/pluginbridge_hostcall_codec.go` | 23 | host call codec 诊断集中 |
| `apps/lina-plugins/monitor-operlog/backend/internal/service/operlog/operlog.go` | 22 | 用户可见错误、字典 fallback、导出文案混合 |
| `apps/lina-core/internal/service/plugin/internal/runtime/upload.go` | 18 | 插件上传错误可能进入管理 API 响应 |

## 已知调用端可见错误

| 模块 | 代表文件 | 中文文案示例 | 处理策略 |
| ---- | -------- | ------------ | -------- |
| 插件资源权限 | `apps/lina-core/internal/controller/plugin/plugin_v1_resource_list.go` | `插件不存在或已禁用`、`无权限访问该插件资源` | 复用或新增 plugin/middleware `bizerr.Code` |
| 内容公告 | `apps/lina-plugins/content-notice/backend/internal/service/notice/notice.go` | `通知公告不存在`、`请选择要删除的记录` | 新增 notice 模块错误码和插件运行时翻译 |
| 组织部门 | `apps/lina-plugins/org-center/backend/internal/service/dept/dept.go` | `部门不存在`、`存在子部门，不允许删除`、`部门编码已存在` | 新增 dept 模块错误码和插件运行时翻译 |
| 组织岗位 | `apps/lina-plugins/org-center/backend/internal/service/post/post.go` | `岗位不存在`、`岗位ID %d 已分配给用户，不能删除` | 新增 post 模块错误码和命名参数 |
| 登录日志 | `apps/lina-plugins/monitor-loginlog/backend/internal/service/loginlog/loginlog.go` | `登录日志不存在` | 新增 loginlog 模块错误码 |
| 操作日志 | `apps/lina-plugins/monitor-operlog/backend/internal/service/operlog/operlog.go` | `操作日志不存在` | 新增 operlog 模块错误码 |
| 源码示例插件 | `apps/lina-plugins/plugin-demo-source/backend/internal/service/demo/*.go` | 记录、附件、上传大小、数据表未安装等错误 | 新增 demo 模块错误码和插件 i18n |
| 动态示例插件 | `apps/lina-plugins/plugin-demo-dynamic/backend/internal/service/dynamic/dynamic_demo_record.go` | 请求体、标题、内容、附件 Base64 和附件大小等错误 | 新增 dynamic demo 错误码和插件 i18n |

## 已知用户可见投影和交付物

| 模块 | 中文文案来源 | 处理策略 |
| ---- | ------------ | -------- |
| `org-center` 岗位 | `未分配部门`、导出表头、`正常/停用` | 使用插件运行时 i18n 或返回结构化状态后由导出边界本地化 |
| `monitor-loginlog` | `成功/失败` 状态 fallback、导出表头 | 优先使用字典或插件运行时 i18n 单一来源 |
| `monitor-operlog` | 操作类型、操作结果 fallback、导出表头 `Fallback` | 优先使用字典或插件运行时 i18n 单一来源 |
| `demo-control` | 演示模式写操作拦截消息 | 改为结构化错误或插件运行时 i18n |
| `plugin-demo-source` | summary 接口中文摘要 | 按请求语言返回或返回稳定 key 由前端渲染 |
| `config` | shell 模式不支持原因 | 改为运行时 i18n key 或结构化 reason code |
| `sysinfo` | 运行时长 `小时/分钟/秒` | 返回结构化时长或按请求语言格式化 |
| `sysconfig` | 中文 fallback label 和 JWT 配置 fallback | 确认可见边界，必要时改英文 fallback 并补齐翻译 |

## 已清理进展

| 范围 | 清理结果 | 验证 |
| ---- | -------- | ---- |
| `config/config_cron.go` | shell 模式不支持原因改为英文 fallback，并通过 `disabledReasonKey` 暴露运行时 i18n key | `go test ./internal/service/config -run 'Test.*Cron\|Test.*PublicFrontend'` |
| `sysinfo/sysinfo.go` | 运行时长新增 `runDurationSeconds` 结构化字段，`runDuration` 由控制器按请求语言格式化 | `go test ./internal/controller/sysinfo` |
| `sysconfig_i18n.go`、`sysconfig_import.go` | 导入模板、导出表头、JWT 示例名称/备注 fallback 改为英文，中文展示依赖运行时语言包 | `go test ./internal/service/sysconfig -run 'TestGenerateImportTemplateLocalizesHeaders\|TestExportLocalizesHeadersButKeepsRawRows\|TestListLocalizesConfigMetadata'` |
| 投影和导出运行时语言资源 | 已补齐 `zh-CN`、`en-US`、`zh-TW` 的 host/plugin 运行时资源 | `make check-runtime-i18n-messages` |
| `pkg/pluginbridge/*.go` | bridge、host call、host service codec/contract 中文诊断统一改为英文稳定文案；非测试 Go 中文扫描为 0 | `go test ./pkg/pluginbridge` |
| `pkg/pluginfs/pluginfs.go` | 插件资源路径、SQL 路径和文件类型校验诊断改为英文稳定文案；非测试 Go 中文扫描为 0 | `go test ./pkg/pluginfs` |
| `pkg/plugindb/**` | plugin data service 审计上下文、授权表和驱动类型诊断改为英文稳定文案；非测试 Go 中文扫描为 0 | `go test ./pkg/plugindb/...` |
| `pkg/excelutil/excelutil.go` | Excel 工具层 close、sheet 和 cell 操作错误改为英文低层诊断；非测试 Go 中文扫描为 0 | `go test ./pkg/excelutil` |
| `plugin/internal/catalog/**` | manifest、spec、authorization、release 校验诊断改为英文稳定文案；非测试 Go 中文扫描为 0 | `go test ./internal/service/plugin/internal/catalog` |
| `plugin/internal/runtime/**` | 上传、归档、reconciler、route、runtime cron 诊断改为英文稳定文案；非测试 Go 中文扫描为 0 | `go test ./internal/service/plugin/internal/runtime` |
| `plugin/internal/{frontend,integration,lifecycle,datahost,wasm}/**` | 前端资源、菜单/Hook/cron 集成、SQL 生命周期、datahost 和 Wasm host service 诊断改为英文稳定文案；非测试 Go 中文扫描为 0 | `go test ./internal/service/plugin/internal/frontend ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/lifecycle ./internal/service/plugin/internal/datahost ./internal/service/plugin/internal/wasm` |
| 插件平台单元测试断言 | catalog、runtime、frontend、datahost、wasm 相关测试断言已同步到英文稳定诊断 | `go test ./internal/service/plugin -run 'Test.*Frontend\|Test.*Integration\|Test.*Lifecycle\|Test.*Data\|Test.*Wasm\|Test.*HostService\|Test.*Menu\|Test.*Hook\|Test.*Cron\|Test.*Dynamic\|Test.*Route'` |

## 与活跃变更的边界

本变更与 `runtime-message-i18n-governance` 的关系如下：

- `runtime-message-i18n-governance` 负责运行时消息治理基础设施、统一响应本地化能力和部分模块落地。
- `backend-hardcoded-chinese-i18n-governance` 负责本次审查发现的后端中文硬编码清理闭环，以扫描清单、allowlist 和任务清单作为验收依据。
- 基础设施能力应复用既有 `bizerr`、运行时 i18n manifest、插件资源加载、统一响应中间件和扫描工具；本变更不重复设计新的消息系统。
- 如果两个活跃变更同时触达同一文件，本变更只处理当前硬编码清理所需内容，不回滚或覆盖其他变更。

## 初始 allowlist 分类

| 分类 | 范围 | 当前处理 |
| ---- | ---- | -------- |
| 生成文件 | `internal/dao`、`internal/model/do`、`internal/model/entity` | 暂不作为违规；后续通过 SQL 注释或生成源治理 |
| 测试 fixture | `_test.go`、测试 JSON、断言样例 | 暂不作为违规；测试断言需要随业务错误结构化同步更新 |
| 用户数据示例 | mock 数据、演示数据、用户输入样例 | 不翻译真实业务数据；仅翻译系统文案 |
| 协议和技术值 | 插件 ID、权限标识、路径、SQL、字段名、字典值 code | 不作为自然语言文案处理 |
| 英文开发者诊断 | 插件 bridge、WASM、codec、启动期 panic | 改为英文稳定文案后允许保留，不进入运行时 i18n |

## API DTO 检查结论

API DTO 文档源文本检查结果为 0 处中文。当前 `apps/lina-core/api` 与 `apps/lina-plugins/*/backend/api` 中 `summary`、`dc`、`description`、`eg` 未发现中文源文本。

后续新增或修改 API DTO 时仍需遵循以下边界：

- OpenAPI/API 文档源文本直接使用英文。
- 非英文文档翻译使用专用 `manifest/i18n/<locale>/apidoc/**/*.json`。
- `en-US/apidoc` 仅保留空对象占位，不写英文映射。
- 不为生成的 Entity schema 在 apidoc service 层维护中文到英文临时转换表。
