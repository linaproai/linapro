## Summary

- [x] 1. 在 `apps/lina-core/pkg/plugin/capability/ai/` 新增 `ai.Service` 聚合包，迁移 `aitext` 到 `capability/ai/aitext`，根 `capability.Services` 改为 `AI() ai.Service`，拆分消费请求与 provider 内部请求使 `SourcePluginID` 由 scoped service 注入
- [x] 2. 新增 `Image()`、`Embedding()`、`Audio()`、`Vision()`、`Document()`、`Safety()` 和 `Video()` 类型化子能力包、公共值对象（`AssetRef`、`ProviderOperationRef`）和 fallback service
- [x] 3. 扩展 `pluginbridge` host service 常量、能力分类映射和清单校验，支持多模态 `ai.*` 方法并拒绝 `computer.act`；扩展 WASM host service dispatcher 和 guest SDK
- [x] 4. 新建 `linapro-ai-core` 源码插件目录、`plugin.yaml`、后端 API/controller/service、前端页面、SQL、`i18n` 和源码插件注册
- [x] 5. 插件 SQL 创建 provider endpoint、model、model capability、tier、tier binding、invocation 和 provider operation 表，seed 多模态能力方法档位，幂等 DDL/Seed 并建立查询索引
- [x] 6. 实现 OpenAI-compatible、Anthropic-compatible 和 Voyage-compatible provider adapter，覆盖 base URL 规范化、usage 解析、错误脱敏、`thinkingEffort` 映射、`/v1` 重试缓存和模型名工具后缀去除
- [x] 7. 实现档位解析缓存，权威源为插件数据库，写后通过共享修订号失效，集群模式同步，30 秒兜底 TTL，缓存不可用降级读数据库
- [x] 8. 实现智能中心前端页面：渠道管理（provider endpoint 管理、密钥脱敏、模型摘要）、模型管理（独立页面、身份维护）、档位管理（能力类型 Tab、三档配置、草稿测试）、调用日志（多模态筛选、清理、脱敏详情）
- [x] 9. 动态插件 `ai` host service 多模态方法授权、资源策略、payload 上限、资产引用可见性、provider operation 查询和脱敏审计
- [x] 10. 为宿主多模态能力、动态 host service、插件后端 service/adapter、缓存和 provider endpoint 补充 Go 单元测试
- [x] 11. 创建插件 E2E：TC001 渠道管理、TC002 档位管理、TC003 调用日志、TC004 provider endpoint、TC005 多模态档位、TC006 多模态日志
- [x] 12. 修复 FB-1 至 FB-58（i18n key 未翻译、lina-review 子仓库展开、技能描述中文化、业务错误翻译资源、档位 seed 源内容、集群缓存共享修订号、调用日志 E2E 确定性数据、模块级 i18n 求值、模型列表分页、maxOutputTokens 文案、遗留进程占用、渠道页面交互迭代、同步模型按钮颜色、模型名工具后缀去除、渠道端点展示、档位管理 Tab 样式、默认参数移除、能力方法筛选、provider endpoint 单一事实来源、模型能力表单一事实来源、SQL 合并简化、调用日志清理、来源插件列、协议展示名、英文菜单术语统一、调用日志搜索区布局对齐、模型管理 identity-only 改造等）
- [x] 13. 运行 Go 编译门禁、前端类型检查、`i18n` 治理检查、E2E 全量通过、OpenSpec 严格校验和 `lina-review` 审查
