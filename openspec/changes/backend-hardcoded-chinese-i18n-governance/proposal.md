## Why

当前后端源码审查发现，手写、非测试、非生成 Go 文件中仍存在大量中文字符串字面量，其中一部分会直接进入 HTTP 响应、插件接口响应、导出文件、管理端展示字段或前端可见运行时配置。已有 `runtime-message-i18n-governance` 变更已经建立了运行时 i18n 与结构化错误治理方向，但本次审查暴露出仍需单独跟进的后端残留清理范围，避免继续依赖中文硬编码作为用户可见文案或业务错误语义。

## What Changes

- 基于本次审查结果建立后端中文硬编码清理清单，明确哪些字符串必须改为运行时 i18n、哪些必须改为 `bizerr` 结构化错误、哪些只需要改为英文开发者诊断、哪些可作为生成文件或测试 fixture 保留。
- 清理调用端可见错误：宿主插件资源权限、内容公告、组织中心、登录日志、操作日志、源码插件示例、动态插件示例等模块中的中文 `gerror.New/Newf/Wrap/Wrapf` 必须改为模块内 `*_code.go` 定义的 `bizerr` 错误，并补齐宿主或插件自有 `manifest/i18n/<locale>/error.json`。
- 清理用户可见投影和交付物文案：部门树“未分配部门”、岗位导出表头和状态、登录/操作日志导出 fallback、演示模式拦截消息、源码插件摘要、Cron shell 不支持原因、系统信息运行时长等必须改为运行时 i18n 或结构化字段。
- 整理插件平台和开发者诊断：`pkg/pluginbridge`、`pkg/pluginfs`、`pkg/plugindb`、插件 catalog/runtime/wasm/datahost 等技术诊断默认改为英文稳定文案；如果错误会穿透到 HTTP/API/插件调用边界，则在边界包装为 `bizerr` 或结构化插件错误。
- 处理生成文件与 OpenAPI schema 来源：DAO/DO/Entity 生成文件中的中文注释和 `description` tag 不手改生成物；若这些 schema 会进入接口文档，应通过 SQL 表/字段注释或生成源调整为英文，避免在 apidoc service 层做临时翻译映射。
- 补齐自动化门禁：扩展或新增运行时硬编码扫描规则，覆盖本次发现的后端高风险位置，并维护 allowlist 说明生成文件、测试 fixture、用户数据示例和英文开发者诊断的边界。
- 补齐测试和审查：增加后端单元测试、插件单元测试和必要 E2E，验证错误码稳定、错误消息按语言展示、导出/投影文案按语言变化，并在任务完成后执行 `lina-review`。

## Capabilities

### New Capabilities

- `backend-hardcoded-chinese-i18n-governance`: 后端源码中文硬编码的识别、分类、清理、i18n/bizerr 改造、生成文件边界和防回归门禁。

### Modified Capabilities

- 无。该变更主要把现有运行时 i18n 与后端规范落到具体残留清理，不改变已归档主规范的用户功能边界。

## Impact

- **宿主后端**：影响 `apps/lina-core/internal/controller/plugin/`、`apps/lina-core/internal/service/config/`、`apps/lina-core/internal/service/sysinfo/`、`apps/lina-core/internal/service/plugin/internal/**`、`apps/lina-core/pkg/pluginbridge/`、`apps/lina-core/pkg/pluginfs/`、`apps/lina-core/pkg/plugindb/`、`apps/lina-core/pkg/excelutil/` 等中文错误、诊断和投影文案。
- **源码插件后端**：影响 `apps/lina-plugins/content-notice/`、`demo-control/`、`monitor-loginlog/`、`monitor-operlog/`、`org-center/`、`plugin-demo-source/`、`plugin-demo-dynamic/` 等模块的业务错误、导出表头、状态 fallback、树节点标签和示例接口文案。
- **语言资源**：需要补齐宿主与插件自己的 `manifest/i18n/zh-CN`、`manifest/i18n/en-US`、`manifest/i18n/zh-TW` 运行时资源；插件运行时文案必须保留在插件目录中，不集中写入 lina-core。
- **接口文档与生成源**：API DTO 源文本保持英文；生成 Entity schema 若仍含中文，应通过 SQL 注释或生成源治理，而不是修改生成文件或 apidoc i18n 服务兜底。
- **测试与工具**：影响 `hack/tools/runtime-i18n` 或等价扫描工具、后端 Go 单元测试、插件单元测试和必要 E2E；新增 allowlist 需要带分类和原因。
