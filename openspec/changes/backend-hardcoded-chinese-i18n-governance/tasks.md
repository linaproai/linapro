## 1. 基线扫描与范围固化

- [ ] 1.1 运行后端中文字符串基线扫描，分别统计全部 Go、手写非测试非生成 Go、生成文件、测试文件的命中数量
- [ ] 1.2 创建 `openspec/changes/backend-hardcoded-chinese-i18n-governance/backend-hardcoded-chinese-audit.md`，记录本次审查命中项、分类、处理策略和优先级
- [ ] 1.3 对照活跃 `runtime-message-i18n-governance` 任务，标注已由其他活跃变更覆盖的基础设施任务，避免重复实现
- [ ] 1.4 建立初始 allowlist，明确生成文件、测试 fixture、用户数据示例和暂不阻断的英文开发者诊断边界
- [ ] 1.5 确认 API DTO `summary/dc/description/eg` 无中文源文本，并把检查命令记录到审查清单

## 2. 调用端可见错误结构化

- [ ] 2.1 清理 `apps/lina-core/internal/controller/plugin/plugin_v1_resource_list.go` 中插件资源权限错误，复用或新增 middleware/plugin `bizerr` 错误定义
- [ ] 2.2 为 `apps/lina-plugins/content-notice/backend/internal/service/notice/` 增加模块错误码，替换“通知公告不存在”“请选择要删除的记录”等中文自由文本错误
- [ ] 2.3 为 `apps/lina-plugins/org-center/backend/internal/service/dept/` 增加模块错误码，替换部门不存在、存在子部门、部门存在用户、部门编码重复等错误
- [ ] 2.4 为 `apps/lina-plugins/org-center/backend/internal/service/post/` 增加模块错误码，替换岗位不存在、删除选择为空、岗位已分配、有效 ID 为空、岗位编码重复等错误
- [ ] 2.5 为 `apps/lina-plugins/monitor-loginlog/backend/internal/service/loginlog/` 增加模块错误码，替换登录日志不存在等调用端可见错误
- [ ] 2.6 为 `apps/lina-plugins/monitor-operlog/backend/internal/service/operlog/` 增加模块错误码，替换操作日志不存在等调用端可见错误
- [ ] 2.7 为 `apps/lina-plugins/plugin-demo-source/backend/internal/service/demo/` 增加模块错误码，替换记录、附件、上传大小、数据表未安装等调用端可见错误
- [ ] 2.8 为 `apps/lina-plugins/plugin-demo-dynamic/backend/internal/service/dynamic/` 增加模块错误码，替换示例记录、请求体、标题、内容、附件 Base64 和附件大小等调用端可见错误
- [ ] 2.9 为上述宿主和插件错误补齐 `zh-CN`、`en-US`、`zh-TW` 运行时 `error.json` 或对应插件运行时语言资源
- [ ] 2.10 更新相关 Go 单元测试，断言 `errorCode/messageKey/messageParams` 和本地化结果，不再断言中文自由文本

## 3. 用户可见投影与交付物文案

- [ ] 3.1 清理 `org-center` 部门/岗位树中的“未分配部门”标签，改为插件运行时 i18n 或结构化 label code
- [ ] 3.2 清理 `org-center` 岗位导出表头和状态“正常/停用”，导出时按请求语言解析表头和状态
- [ ] 3.3 清理 `monitor-loginlog` 登录状态 `成功/失败` fallback 和导出表头中文 fallback，优先使用字典或插件运行时 i18n 单一来源
- [ ] 3.4 清理 `monitor-operlog` 操作类型、操作结果 fallback 和导出表头中文 fallback，优先使用字典或插件运行时 i18n 单一来源
- [ ] 3.5 清理 `demo-control` 演示模式拦截消息，改为结构化错误或插件运行时 i18n 文案
- [ ] 3.6 清理 `plugin-demo-source` summary 接口返回的中文摘要，改为按请求语言返回或返回稳定 key 由前端渲染
- [ ] 3.7 清理 `config/config_cron.go` 的 shell 模式不支持原因，改为运行时 i18n key 或结构化 reason code
- [ ] 3.8 清理 `sysinfo/sysinfo.go` 的运行时长中文单位，改为结构化时长字段或按请求语言格式化
- [ ] 3.9 检查 `sysconfig_i18n.go`、`sysconfig_import.go` 中中文 fallback 是否仍会用户可见，必要时改为英文 fallback 并补齐运行时翻译资源
- [ ] 3.10 为投影和导出文案补齐 `zh-CN`、`en-US`、`zh-TW` 语言资源，并增加缺失翻译校验

## 4. 插件平台与开发者诊断治理

- [ ] 4.1 将 `apps/lina-core/pkg/pluginbridge/*.go` 中 bridge、host call、host service codec 的中文诊断统一改为英文稳定文案
- [ ] 4.2 将 `apps/lina-core/pkg/pluginfs/pluginfs.go` 中插件资源路径和 SQL 资源路径诊断改为英文，并确认用户边界包装策略
- [ ] 4.3 将 `apps/lina-core/pkg/plugindb/**` 中 plugin data service 诊断改为英文，并确认审计上下文错误不会直接裸露给管理端用户
- [ ] 4.4 将 `apps/lina-core/pkg/excelutil/excelutil.go` 中 Excel 工具层中文错误改为英文低层诊断，并确保业务边界负责本地化用户提示
- [ ] 4.5 清理 `apps/lina-core/internal/service/plugin/internal/catalog/**` 中 manifest、spec、authorization、release 校验中文诊断，用户边界需要结构化包装
- [ ] 4.6 清理 `apps/lina-core/internal/service/plugin/internal/runtime/**` 中上传、归档、reconciler、route、runtime cron 中文诊断，管理 API 可见错误必须结构化
- [ ] 4.7 清理 `apps/lina-core/internal/service/plugin/internal/frontend/**`、`integration/**`、`lifecycle/**`、`datahost/**`、`wasm/**` 中中文诊断，并区分英文开发者诊断和用户可见错误
- [ ] 4.8 更新插件平台相关单元测试，改为断言稳定英文诊断、错误码或 message key

## 5. 配置启动期与生成源治理

- [ ] 5.1 将 `apps/lina-core/internal/service/config/config_duration.go`、`config_i18n.go`、`config_metadata.go`、`config_plugin.go` 中启动期 panic 和配置诊断改为英文开发者诊断
- [ ] 5.2 扫描 Entity `description` 中文来源，列出会进入 OpenAPI schema 或用户可见文档的表和字段
- [ ] 5.3 对需要治理的 schema，修改对应 SQL 表/字段注释或 codegen 输入源为英文，不手工编辑生成文件
- [ ] 5.4 如修改 SQL 注释，执行项目约定的 `make init` 和 `make dao` 流程并确认生成结果稳定
- [ ] 5.5 确认 apidoc service 没有新增中文到英文临时映射，也没有把生成 schema 翻译项写入 `en-US/apidoc`

## 6. 扫描工具、文档与门禁

- [ ] 6.1 扩展 `hack/tools/runtime-i18n` 或新增等价 Go 工具，覆盖中文 `gerror.New*`、`gerror.Wrap*`、`errors.New`、`fmt.Errorf` 高风险模式
- [ ] 6.2 扩展扫描规则覆盖 `Message`、`Reason`、`Fallback`、`Label`、`Title`、`DisabledReason` 字段赋值和导出表头数组
- [ ] 6.3 扩展扫描规则覆盖状态/类型 label map、树节点 label 构造、插件 bridge/host service/manifest 校验错误构造
- [ ] 6.4 为扫描工具增加分类化报告输出，至少区分违规、生成文件统计、测试 fixture 统计和 allowlist 命中
- [ ] 6.5 更新或新增 allowlist 文件，要求每条记录包含文件、分类、保留原因和适用范围
- [ ] 6.6 将扫描命令接入本地验证入口或文档化为本变更验收命令
- [ ] 6.7 更新相关 README 或治理文档；如新增目录说明，按项目规范同步维护 `README.md` 和 `README.zh_CN.md`

## 7. 自动化测试与 E2E

- [ ] 7.1 增加后端单元测试，覆盖内容公告、组织中心、登录日志、操作日志、示例插件等业务错误在 `zh-CN`、`en-US`、`zh-TW` 下的本地化
- [ ] 7.2 增加插件平台单元测试，覆盖插件 bridge/host service/catalog/runtime 错误的英文开发者诊断和用户边界结构化包装
- [ ] 7.3 增加导出和投影单元测试，覆盖岗位导出、登录日志导出、操作日志导出、部门树未分配标签、系统信息运行时长等语言变化
- [ ] 7.4 创建 `hack/tests/e2e/i18n/TC0135-backend-error-localization.ts`，验证同一后端业务错误在 `zh-CN`、`en-US`、`zh-TW` 下展示不同语言且 `errorCode` 稳定
- [ ] 7.5 创建 `hack/tests/e2e/i18n/TC0136-backend-export-localization.ts`，验证岗位、登录日志或操作日志导出表头和状态值按当前语言输出
- [ ] 7.6 创建 `hack/tests/e2e/i18n/TC0137-backend-hardcoded-chinese-regression.ts`，验证关键后端投影和插件页面切换英文后无残留中文系统文案
- [ ] 7.7 运行相关 `go test`、扫描工具、必要前端检查和新增 E2E 用例，并记录命令与结果

## 8. 收口审查

- [ ] 8.1 重新运行后端中文字符串扫描，更新 `backend-hardcoded-chinese-audit.md` 的已清理项和剩余 allowlist
- [ ] 8.2 运行 `openspec status --change backend-hardcoded-chinese-i18n-governance --json`，确认 proposal、design、specs、tasks 状态一致
- [ ] 8.3 执行 `lina-review`，重点审查 i18n 资源归属、调用端 `bizerr`、生成文件边界、扫描 allowlist 和测试覆盖
- [ ] 8.4 根据审查结果修复阻断问题，并确认任务清单状态准确
